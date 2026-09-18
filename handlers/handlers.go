package handlers

import (
	"fmt"
	"net/http"
	"pokedex-go/database"
	"pokedex-go/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Home renders the main index page.
func Home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "PokeDB",
	})
}

// PokemonList renders the grid of all Pokémon (HTMX target) with pagination.
func PokemonList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage := 12

	offset := (page - 1) * perPage
	pokemons, err := database.GetPokemonPage(offset, perPage)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load Pokémon: %v", err)
		return
	}

	total, _ := database.CountPokemon()
	totalPages := (total + perPage - 1) / perPage

	if len(pokemons) == 0 && page == 1 {
		c.Header("HX-Trigger", `{"refreshStats":true}`)
		c.HTML(http.StatusOK, "list_empty.html", gin.H{})
		return
	}

	var html strings.Builder
	for _, p := range pokemons {
		html.WriteString(renderPokemonGridItem(p))
	}

	// Build pagination controls
	paginationHTML := buildPagination(page, totalPages, total)

	// If this is an HTMX request, return grid + pagination
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", `{"refreshStats":true}`)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.String()+paginationHTML))
		return
	}

	// Full page load
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":      "PokeDB",
		"pokemons":   pokemons,
		"page":       page,
		"totalPages": totalPages,
		"total":      total,
	})
}

// buildPagination generates pagination HTML controls.
func buildPagination(currentPage, totalPages, total int) string {
	if totalPages <= 1 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<div class="flex items-center justify-center gap-2 mt-8">`)

	// Previous button
	if currentPage > 1 {
		sb.WriteString(fmt.Sprintf(`<a href="/pokemon?page=%d" hx-get="/pokemon?page=%d" hx-target="#pokemon-grid" hx-swap="innerHTML" class="px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-gray-300 hover:text-white hover:bg-white/10 transition text-sm">← Prev</a>`, currentPage-1, currentPage-1))
	}

	// Page numbers
	start := currentPage - 2
	if start < 1 {
		start = 1
	}
	end := currentPage + 2
	if end > totalPages {
		end = totalPages
	}

	for i := start; i <= end; i++ {
		if i == currentPage {
			sb.WriteString(fmt.Sprintf(`<span class="px-3 py-2 rounded-lg bg-poke-red text-white font-medium text-sm">%d</span>`, i))
		} else {
			sb.WriteString(fmt.Sprintf(`<a href="/pokemon?page=%d" hx-get="/pokemon?page=%d" hx-target="#pokemon-grid" hx-swap="innerHTML" class="px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-gray-300 hover:text-white hover:bg-white/10 transition text-sm">%d</a>`, i, i, i))
		}
	}

	// Next button
	if currentPage < totalPages {
		sb.WriteString(fmt.Sprintf(`<a href="/pokemon?page=%d" hx-get="/pokemon?page=%d" hx-target="#pokemon-grid" hx-swap="innerHTML" class="px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-gray-300 hover:text-white hover:bg-white/10 transition text-sm">Next →</a>`, currentPage+1, currentPage+1))
	}

	sb.WriteString(`</div>`)
	return sb.String()
}

// PokemonDetail renders a single Pokémon detail card (also HTMX target).
func PokemonDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Pokémon ID")
		return
	}

	p, err := database.GetPokemonByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	if p == nil {
		c.String(http.StatusNotFound, "Pokémon not found")
		return
	}

	c.HTML(http.StatusOK, "pokemon_show.html", p)
}

var pokemonTypes = []string{
	"Normal", "Fire", "Water", "Electric", "Grass", "Ice",
	"Fighting", "Poison", "Ground", "Flying", "Psychic", "Bug",
	"Rock", "Ghost", "Dragon", "Dark", "Steel", "Fairy",
}

// PokemonCreate renders the create form.
func PokemonCreate(c *gin.Context) {
	c.HTML(http.StatusOK, "pokemon_form.html", gin.H{
		"mode":    "create",
		"pokemon": models.Pokemon{},
		"types":   pokemonTypes,
	})
}

// PokemonEdit renders the edit form.
func PokemonEdit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}

	p, err := database.GetPokemonByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	if p == nil {
		c.String(http.StatusNotFound, "Pokémon not found")
		return
	}

	c.HTML(http.StatusOK, "pokemon_form.html", gin.H{
		"mode":    "edit",
		"pokemon": *p,
		"id":      id,
		"types":   pokemonTypes,
	})
}

// PokemonStore handles the create form submission (both HTMX and full-page).
func PokemonStore(c *gin.Context) {
	p := models.Pokemon{
		Name:              c.PostForm("name"),
		Type1:             c.PostForm("type1"),
		Type2:             c.PostForm("type2"),
		Height:            c.PostForm("height"),
		Weight:            c.PostForm("weight"),
		BaseExperience:    parseInt(c.PostForm("base_experience")),
		ImageURL:          c.PostForm("image_url"),
		Description:       c.PostForm("description"),
		EvolvesFromStatus: c.PostForm("evolves_from"),
	}

	if strings.TrimSpace(p.Name) == "" {
		c.String(http.StatusBadRequest, "Name is required")
		return
	}
	if strings.TrimSpace(p.Type1) == "" {
		c.String(http.StatusBadRequest, "Type 1 is required")
		return
	}
	if strings.TrimSpace(p.Height) == "" {
		c.String(http.StatusBadRequest, "Height is required")
		return
	}
	if strings.TrimSpace(p.Weight) == "" {
		c.String(http.StatusBadRequest, "Weight is required")
		return
	}

	created, err := database.CreatePokemon(p)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "UNIQUE constraint failed") {
			c.String(http.StatusConflict, "A Pokémon with this name already exists")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to create: %v", err)
		return
	}

	// HTMX request: swap in the new card at the top of the grid
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", `{"refreshStats":true}`)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderPokemonGridItem(*created)))
		return
	}

	c.Redirect(http.StatusFound, "/pokemon/"+strconv.FormatInt(created.ID, 10))
}

// PokemonUpdate handles the edit form submission.
func PokemonUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}

	existing, err := database.GetPokemonByID(id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	if existing == nil {
		c.String(http.StatusNotFound, "Pokémon not found")
		return
	}

	p := models.Pokemon{
		ID:                id,
		Name:              c.PostForm("name"),
		Type1:             c.PostForm("type1"),
		Type2:             c.PostForm("type2"),
		Height:            c.PostForm("height"),
		Weight:            c.PostForm("weight"),
		BaseExperience:    parseInt(c.PostForm("base_experience")),
		ImageURL:          c.PostForm("image_url"),
		Description:       c.PostForm("description"),
		EvolvesFromStatus: c.PostForm("evolves_from"),
	}

	if strings.TrimSpace(p.Name) == "" {
		c.String(http.StatusBadRequest, "Name is required")
		return
	}
	if strings.TrimSpace(p.Type1) == "" {
		c.String(http.StatusBadRequest, "Type 1 is required")
		return
	}

	updated, err := database.UpdatePokemon(id, p)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "UNIQUE constraint failed") {
			c.String(http.StatusConflict, "Another Pokémon with this name already exists")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to update: %v", err)
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", `{"refreshStats":true}`)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderPokemonGridItem(*updated)))
		return
	}

	c.Redirect(http.StatusFound, "/pokemon/"+strconv.FormatInt(updated.ID, 10))
}

// PokemonDelete handles deletion (HTMX swap + redirect toggle).
func PokemonDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}

	err = database.DeletePokemon(id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete: %v", err)
		return
	}

	// Tell the parent page to refresh the grid after deletion
	c.Header("HX-Trigger", `{"refreshList":true}`)
	c.String(http.StatusOK, "deleted")
}

// PokemonSearch handles HTMX search by name.
func PokemonSearch(c *gin.Context) {
	query := strings.TrimSpace(c.PostForm("query"))
	if query == "" {
		PokemonList(c)
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, name, type1, type2, height, weight, base_experience,
		       image_url, description, evolves_from, created_at, updated_at
		FROM pokemon
		WHERE name LIKE ?
		ORDER BY id ASC`, "%"+query+"%")
	if err != nil {
		c.String(http.StatusInternalServerError, "Search error: %v", err)
		return
	}
	defer rows.Close()

	var pokemons []models.Pokemon
	for rows.Next() {
		var p models.Pokemon
		if err := rows.Scan(&p.ID, &p.Name, &p.Type1, &p.Type2, &p.Height, &p.Weight,
			&p.BaseExperience, &p.ImageURL, &p.Description, &p.EvolvesFromStatus, &p.CreatedAt, &p.UpdatedAt); err != nil {
			c.String(http.StatusInternalServerError, "Scan error: %v", err)
			return
		}
		pokemons = append(pokemons, p)
	}

	if len(pokemons) == 0 {
		c.HTML(http.StatusOK, "list_empty.html", gin.H{
			"message": fmt.Sprintf("No Pokémon matching \"%s\"", query),
		})
		return
	}

	var html strings.Builder
	for _, p := range pokemons {
		html.WriteString(renderPokemonGridItem(p))
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.String()))
}

// Stats returns the total count (HTMX target).
func Stats(c *gin.Context) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM pokemon`).Scan(&count)
	if err != nil {
		c.String(http.StatusInternalServerError, "0")
		return
	}
	c.HTML(http.StatusOK, "stats.html", gin.H{"Count": count})
}

// parseInt parses a form value to int, defaulting to 0 on error.
func parseInt(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

// renderPokemonGridItem renders a single Pokémon grid card as HTML.
func renderPokemonGridItem(p models.Pokemon) string {
	type1Color := "bg-poke-red/90"
	type2Color := "bg-poke-accent/90"
	if p.Type1 == "Water" {
		type1Color = "bg-blue-500/90"
	} else if p.Type1 == "Grass" {
		type1Color = "bg-green-500/90"
	} else if p.Type1 == "Electric" {
		type1Color = "bg-yellow-500/90"
	} else if p.Type1 == "Fire" {
		type1Color = "bg-red-500/90"
	} else if p.Type1 == "Psychic" {
		type1Color = "bg-purple-500/90"
	}

	imgTag := ""
	if p.ImageURL != "" {
		imgTag = fmt.Sprintf(`<img src="%s" alt="%s" class="absolute inset-0 w-full h-full object-cover opacity-70 hover:opacity-100 transition-opacity">`, p.ImageURL, p.Name)
	}

	return fmt.Sprintf(`
<div class="poke-card bg-poke-dark/70 backdrop-blur rounded-2xl border border-white/5 overflow-hidden">
  <div class="relative h-40 bg-gradient-to-br from-poke-dark to-poke-blue flex items-center justify-center overflow-hidden">
    %s
    <div class="absolute top-2 right-2 flex gap-1">
      <span class="type-badge px-1.5 py-0.5 rounded-full %s text-white text-[10px]">%s</span>
      %s
    </div>
    <div class="absolute bottom-2 left-3 text-xs text-gray-400 bg-poke-dark/80 px-1.5 py-0.5 rounded">
      No.%d
    </div>
  </div>
  <div class="p-4">
    <h3 class="text-base font-bold text-white mb-1">%s</h3>
    <p class="text-xs text-gray-400 mb-2 line-clamp-2">%s</p>
    <div class="flex items-center gap-3 text-xs text-gray-400 mb-3">
      <span>📏 %s</span>
      <span>⚖️ %s</span>
    </div>
    <a href="/pokemon/%d" class="block w-full bg-poke-blue/80 hover:bg-poke-blue text-white text-sm font-medium py-2 rounded-lg transition text-center">
      View Details
    </a>
  </div>
</div>`,
		imgTag,
		type1Color,
		p.Type1,
		func() string {
			if p.Type2 != "" {
				return fmt.Sprintf(`<span class="type-badge px-1.5 py-0.5 rounded-full %s text-white text-[10px] ml-0.5">%s</span>`, type2Color, p.Type2)
			}
			return ""
		}(),
		p.ID,
		p.Name,
		p.Description,
		p.Height,
		p.Weight,
		p.ID,
	)
}
