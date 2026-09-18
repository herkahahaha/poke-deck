package seed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"pokedex-go/database"
	"pokedex-go/models"
)

// PokemonList represents the PokéAPI list response.
type PokemonList struct {
	Count   int          `json:"count"`
	Next    *string      `json:"next"`
	Results []pokemonName `json:"results"`
}

// PokemonDetail represents the Pokémon detail from PokéAPI.
type PokemonDetail struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	BaseExperience int    `json:"base_experience"`
	Types          []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Sprites struct {
		FrontDefault string `json:"front_default"`
	} `json:"sprites"`
	Species struct {
		URL string `json:"url"`
	} `json:"species"`
}

// SpeciesDetail represents PokéAPI species endpoint.
type SpeciesDetail struct {
	FlavorTextEntries []struct {
		Language struct {
			Name string `json:"name"`
		} `json:"language"`
		FlavorText string `json:"flavor_text"`
	} `json:"flavor_text_entries"`
}

// SeedFromPokeAPI fetches Pokémon from PokéAPI concurrently and populates the database.
// limit <= 0 means fetch all. Uses worker pool for concurrent HTTP requests.
func SeedFromPokeAPI(limit int) error {
	fmt.Println("🧬 Seeding Pokémon from PokéAPI...")

	// Shared HTTP client with connection pooling
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Fetch all Pokémon names first
	allPokemon, err := fetchAllPokemonNames(client)
	if err != nil {
		return fmt.Errorf("failed to fetch Pokémon list: %w", err)
	}

	fmt.Printf("📊 Total Pokémon available: %d\n", len(allPokemon))

	// Apply limit
	if limit > 0 && len(allPokemon) > limit {
		allPokemon = allPokemon[:limit]
		fmt.Printf("⏱️  Limiting seed to first %d Pokémon\n", limit)
	}

	fmt.Printf("✅ Fetched %d Pokémon names. Now fetching details concurrently...\n", len(allPokemon))

	// Get existing IDs to skip duplicates
	existingIDs := getExistingIDs()
	fmt.Printf("📦 %d Pokémon already in database, skipping duplicates\n", len(existingIDs))

	// Concurrent detail fetching with worker pool
	const numWorkers = 20
	jobs := make(chan pokemonJob, len(allPokemon))
	results := make(chan pokemonResult, len(allPokemon))

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(client, jobs, results, &wg)
	}

	// Queue jobs - extract ID from URL
	go func() {
		for _, p := range allPokemon {
			if existingIDs[extractIDFromURL(p.URL)] {
				continue // Skip existing
			}
			id := extractIDFromURL(p.URL)
			jobs <- pokemonJob{url: p.URL, id: id}
		}
		close(jobs)
	}()

	// Wait for workers to finish, then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and batch insert
	var toInsert []models.Pokemon
	processed := 0
	errors := 0

	for res := range results {
		processed++
		if res.err != nil {
			errors++
			if errors <= 5 {
				fmt.Printf("    ⚠️  Error fetching ID %d: %v\n", res.id, res.err)
			}
			continue
		}
		toInsert = append(toInsert, res.pokemon)

		// Batch insert every 50
		if len(toInsert) >= 50 {
			if err := database.BulkCreatePokemon(toInsert); err != nil {
				fmt.Printf("    ⚠️  Batch insert error: %v\n", err)
			}
			toInsert = toInsert[:0]
		}

		if processed%100 == 0 {
			fmt.Printf("  ⏳ Processed %d/%d...\n", processed, len(allPokemon))
		}
	}

	// Insert remaining
	if len(toInsert) > 0 {
		if err := database.BulkCreatePokemon(toInsert); err != nil {
			fmt.Printf("    ⚠️  Final batch insert error: %v\n", err)
		}
	}

	fmt.Printf("✅ Seeding complete! Processed %d, errors %d\n", processed, errors)
	return nil
}

type pokemonJob struct {
	url string
	id  int
}

type pokemonResult struct {
	id      int
	pokemon models.Pokemon
	err     error
}

func worker(client *http.Client, jobs <-chan pokemonJob, results chan<- pokemonResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		p, err := fetchPokemonDetail(client, job.url)
		if err != nil {
			results <- pokemonResult{id: job.id, err: err}
			continue
		}

		pokemon := models.Pokemon{
			ID:             int64(p.ID),
			Name:           toTitle(p.Name),
			Type1:          toTitle(p.Types[0].Type.Name),
			Height:         fmt.Sprintf("%.1f m", float64(p.Height)/10.0),
			Weight:         fmt.Sprintf("%.1f kg", float64(p.Weight)/10.0),
			BaseExperience: p.BaseExperience,
			ImageURL:       p.Sprites.FrontDefault,
		}

		if len(p.Types) > 1 {
			pokemon.Type2 = toTitle(p.Types[1].Type.Name)
		}

		// Fetch species description
		if p.Species.URL != "" {
			if desc, err := fetchDescription(client, p.Species.URL); err == nil && desc != "" {
				pokemon.Description = cleanDesc(desc)
			}
		}

		results <- pokemonResult{id: p.ID, pokemon: pokemon}
	}
}

type pokemonName struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	ID   int    `json:"-"` // extracted from URL
}

func fetchAllPokemonNames(client *http.Client) ([]pokemonName, error) {
	var all []pokemonName

	nextURL := "https://pokeapi.co/api/v2/pokemon?limit=20"
	for nextURL != "" {
		page, err := fetchPokemonList(client, nextURL)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Results...)
		if page.Next != nil {
			nextURL = *page.Next
		} else {
			nextURL = ""
		}
	}
	return all, nil
}

func getExistingIDs() map[int]bool {
	ids := make(map[int]bool)
	rows, err := database.DB.Query(`SELECT id FROM pokemon`)
	if err != nil {
		return ids
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			ids[id] = true
		}
	}
	return ids
}

func fetchPokemonList(client *http.Client, url string) (*PokemonList, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var result PokemonList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	return &result, nil
}

func fetchPokemonDetail(client *http.Client, url string) (*PokemonDetail, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var detail PokemonDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	return &detail, nil
}

func fetchDescription(client *http.Client, speciesURL string) (string, error) {
	resp, err := client.Get(speciesURL)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var species SpeciesDetail
	if err := json.NewDecoder(resp.Body).Decode(&species); err != nil {
		return "", fmt.Errorf("json decode: %w", err)
	}

	for _, entry := range species.FlavorTextEntries {
		if entry.Language.Name == "en" {
			return entry.FlavorText, nil
		}
	}
	return "", nil
}

func cleanDesc(text string) string {
	text = strings.ReplaceAll(text, "\f", "")
	text = strings.TrimSpace(text)
	if len(text) > 500 {
		text = text[:500]
	}
	return text
}

func extractIDFromURL(url string) int {
	// URL format: https://pokeapi.co/api/v2/pokemon/25/
	parts := strings.Split(url, "/")
	for i := len(parts) - 2; i >= 0; i-- {
		if id, err := strconv.Atoi(parts[i]); err == nil {
			return id
		}
	}
	return 0
}

func toTitle(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}


