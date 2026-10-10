package seed

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestToTitle(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"bulbasaur", "Bulbasaur"},
		{"pikachu", "Pikachu"},
		{"mr-mime", "Mr-mime"},
		{"", ""},
		{"CHARIZARD", "CHARIZARD"},
	}
	for _, c := range cases {
		if got := toTitle(c.in); got != c.want {
			t.Errorf("toTitle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCleanDesc(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"A strange seed was planted\fon its back.", "A strange seed was plantedon its back."},
		{"  padded text  ", "padded text"},
		{"", ""},
	}
	for _, c := range cases {
		if got := cleanDesc(c.in); got != c.want {
			t.Errorf("cleanDesc(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCleanDescTruncatesTo500(t *testing.T) {
	long := strings.Repeat("a", 600)
	got := cleanDesc(long)
	if len(got) != 500 {
		t.Errorf("expected 500 chars, got %d", len(got))
	}
}

func TestExtractIDFromURL(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"https://pokeapi.co/api/v2/pokemon/25/", 25},
		{"https://pokeapi.co/api/v2/pokemon/1", 1},
		{"https://pokeapi.co/api/v2/pokemon/", 0},
		{"", 0},
		{"https://example.com/species/9999/", 9999},
	}
	for _, c := range cases {
		if got := extractIDFromURL(c.in); got != c.want {
			t.Errorf("extractIDFromURL(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// --- HTTP fetch tests with a mocked server ---

func TestFetchPokemonList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"count": 2,
			"next": null,
			"results": [
				{"name": "bulbasaur", "url": "https://pokeapi.co/api/v2/pokemon/1/"},
				{"name": "ivysaur", "url": "https://pokeapi.co/api/v2/pokemon/2/"}
			]
		}`))
	}))
	defer srv.Close()

	list, err := fetchPokemonList(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("fetchPokemonList returned error: %v", err)
	}
	if list.Count != 2 {
		t.Errorf("expected count 2, got %d", list.Count)
	}
	if len(list.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(list.Results))
	}
	if list.Results[0].Name != "bulbasaur" {
		t.Errorf("expected first result bulbasaur, got %q", list.Results[0].Name)
	}
}

func TestFetchPokemonListHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	if _, err := fetchPokemonList(srv.Client(), srv.URL); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestFetchPokemonDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 25,
			"name": "pikachu",
			"height": 4,
			"weight": 60,
			"base_experience": 112,
			"types": [
				{"type": {"name": "electric"}}
			],
			"sprites": {"front_default": "https://example.com/pikachu.png"},
			"species": {"url": "https://pokeapi.co/api/v2/pokemon-species/25/"}
		}`))
	}))
	defer srv.Close()

	detail, err := fetchPokemonDetail(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("fetchPokemonDetail returned error: %v", err)
	}
	if detail.ID != 25 {
		t.Errorf("expected ID 25, got %d", detail.ID)
	}
	if len(detail.Types) == 0 || detail.Types[0].Type.Name != "electric" {
		t.Errorf("expected electric type, got %+v", detail.Types)
	}
	if detail.Sprites.FrontDefault == "" {
		t.Error("expected a front_default sprite")
	}
}

func TestFetchDescriptionEnglish(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"flavor_text_entries": [
				{"language": {"name": "ja"}, "flavor_text": "でんきねずみ"},
				{"language": {"name": "en"}, "flavor_text": "When several of these Pokémon gather, their electricity could build and cause lightning storms."}
			]
		}`))
	}))
	defer srv.Close()

	desc, err := fetchDescription(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("fetchDescription returned error: %v", err)
	}
	if !strings.Contains(desc, "lightning storms") {
		t.Errorf("expected English flavor text, got %q", desc)
	}
}

func TestFetchDescriptionNoEnglish(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"flavor_text_entries": [
				{"language": {"name": "ja"}, "flavor_text": "でんきねずみ"}
			]
		}`))
	}))
	defer srv.Close()

	desc, err := fetchDescription(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("fetchDescription returned error: %v", err)
	}
	if desc != "" {
		t.Errorf("expected empty description, got %q", desc)
	}
}

func TestFetchDescriptionHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := fetchDescription(srv.Client(), srv.URL); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

// --- worker test with mocked API ---

func TestWorkerBuildsPokemon(t *testing.T) {
	var speciesURL string // populated once the test server is created

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/pokemon/25", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"id": 25,
			"name": "pikachu",
			"height": 4,
			"weight": 60,
			"base_experience": 112,
			"types": [
				{"type": {"name": "electric"}}
			],
			"sprites": {"front_default": "https://example.com/pikachu.png"},
			"species": {"url": %q}
		}`, speciesURL)
	})
	mux.HandleFunc("/api/v2/pokemon-species/25", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"flavor_text_entries": [
				{"language": {"name": "en"}, "flavor_text": "When several of these Pokémon gather, their electricity could build and cause lightning storms."}
			]
		}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	speciesURL = srv.URL + "/api/v2/pokemon-species/25/"

	jobs := make(chan pokemonJob, 1)
	results := make(chan pokemonResult, 1)
	jobs <- pokemonJob{url: srv.URL + "/api/v2/pokemon/25", id: 25}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(1)
	go worker(srv.Client(), jobs, results, &wg)
	wg.Wait()
	close(results)

	res, ok := <-results
	if !ok {
		t.Fatal("expected one result from worker")
	}
	if res.err != nil {
		t.Fatalf("worker returned error: %v", res.err)
	}
	if res.pokemon.Name != "Pikachu" {
		t.Errorf("expected Pikachu, got %q", res.pokemon.Name)
	}
	if res.pokemon.Type1 != "Electric" {
		t.Errorf("expected Electric type, got %q", res.pokemon.Type1)
	}
	if res.pokemon.Height != "0.4 m" {
		t.Errorf("expected height 0.4 m, got %q", res.pokemon.Height)
	}
	if res.pokemon.Weight != "6.0 kg" {
		t.Errorf("expected weight 6.0 kg, got %q", res.pokemon.Weight)
	}
	if res.pokemon.Description == "" {
		t.Error("expected a description fetched from the species endpoint")
	}
}
