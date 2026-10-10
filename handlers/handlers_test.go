package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"pokedex-go/database"
	"pokedex-go/mocks"
	"pokedex-go/models"
)

// setupTestRouter creates a Gin router in test mode, connects an in-memory
// database seeded with mock data, loads templates, and registers routes.
func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	if err := database.ConnectDB(":memory:"); err != nil {
		t.Fatalf("ConnectDB failed: %v", err)
	}
	database.DB.SetMaxOpenConns(1)
	if err := database.BulkCreatePokemon(mocks.MockPokemons()); err != nil {
		t.Fatalf("seeding mock data failed: %v", err)
	}
	t.Cleanup(func() {
		_ = database.CloseDB()
	})

	r := gin.New()
	r.LoadHTMLGlob("../templates/*.html")

	r.GET("/", Home)
	r.GET("/index", Home)
	r.GET("/pokemon", PokemonList)
	r.POST("/pokemon/search", PokemonSearch)
	r.GET("/pokemon/create", PokemonCreate)
	r.POST("/pokemon/create", PokemonStore)
	r.GET("/pokemon/:id/edit", PokemonEdit)
	r.POST("/pokemon/:id/update", PokemonUpdate)
	r.GET("/pokemon/:id", PokemonDetail)
	r.DELETE("/pokemon/:id", PokemonDelete)
	r.GET("/stats", Stats)

	return r
}

// doRequest performs a request against the router and returns the recorder.
func doRequest(r *gin.Engine, method, target string, body *strings.Reader, headers map[string]string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHome(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "PokeDB") {
		t.Error("expected home page to contain PokeDB")
	}
}

func TestPokemonList(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	// Non-HTMX request renders the full index page.
	if !strings.Contains(body, "PokeDB") {
		t.Error("expected full page render")
	}
}

func TestPokemonListHTMX(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon?page=1", nil, map[string]string{"HX-Request": "true"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{"Bulbasaur", "Pikachu", "Mewtwo", "poke-card"} {
		if !strings.Contains(body, want) {
			t.Errorf("HTMX grid missing %q", want)
		}
	}
}

func TestPokemonListEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := database.ConnectDB(":memory:"); err != nil {
		t.Fatalf("ConnectDB failed: %v", err)
	}
	database.DB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.CloseDB() })

	r := gin.New()
	r.LoadHTMLGlob("../templates/*.html")
	r.GET("/pokemon", PokemonList)

	w := doRequest(r, http.MethodGet, "/pokemon", nil, map[string]string{"HX-Request": "true"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Empty DB on page 1 renders the empty state.
	if !strings.Contains(w.Body.String(), "No Pokémon") {
		t.Error("expected empty state message")
	}
}

func TestPokemonListInvalidPage(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon?page=abc", nil, map[string]string{"HX-Request": "true"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (invalid page falls back to 1), got %d", w.Code)
	}
}

func TestPokemonDetail(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon/25", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Pikachu") {
		t.Error("expected detail page to contain Pikachu")
	}
	if !strings.Contains(body, "Electric") {
		t.Error("expected detail page to contain Electric type")
	}
}

func TestPokemonDetailNotFound(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon/9999", nil, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPokemonDetailInvalidID(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon/notanumber", nil, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPokemonCreateForm(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon/create", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Add") {
		t.Error("expected create form")
	}
}

func TestPokemonEditForm(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon/25/edit", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Edit") {
		t.Error("expected edit form")
	}
	if !strings.Contains(body, "Pikachu") {
		t.Error("expected edit form to contain existing name")
	}
}

func TestPokemonEditNotFound(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/pokemon/9999/edit", nil, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPokemonStore(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{
		"name":            {"Togepi"},
		"type1":           {"Fairy"},
		"type2":           {""},
		"height":          {"0.3 m"},
		"weight":          {"1.5 kg"},
		"base_experience": {"49"},
		"image_url":       {"https://example.com/togepi.png"},
		"description":     {"Spiky ball."},
		"evolves_from":    {""},
	}

	w := doRequest(r, http.MethodPost, "/pokemon/create", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect 302, got %d", w.Code)
	}

	// Verify the pokemon was persisted under its auto-assigned ID.
	id, err := pokemonIDByName("Togepi")
	if err != nil {
		t.Fatalf("verification query failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected Togepi to be persisted")
	}
	p, err := database.GetPokemonByID(id)
	if err != nil {
		t.Fatalf("GetPokemonByID failed: %v", err)
	}
	if p == nil || p.Name != "Togepi" {
		t.Errorf("expected persisted Togepi, got %+v", p)
	}
}

func TestPokemonStoreHTMX(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{
		"name":      {"Togepi"},
		"type1":     {"Fairy"},
		"type2":     {""},
		"height":    {"0.3 m"},
		"weight":    {"1.5 kg"},
		"image_url": {"https://example.com/togepi.png"},
	}
	w := doRequest(r, http.MethodPost, "/pokemon/create", strings.NewReader(form.Encode()), map[string]string{"HX-Request": "true"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Togepi") {
		t.Error("expected HTMX response to contain new pokemon card")
	}
}

func TestPokemonStoreValidation(t *testing.T) {
	r := setupTestRouter(t)

	// Missing required name.
	form := url.Values{
		"type1":  {"Fairy"},
		"height": {"0.3 m"},
		"weight": {"1.5 kg"},
	}
	w := doRequest(r, http.MethodPost, "/pokemon/create", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestPokemonStoreDuplicateName(t *testing.T) {
	r := setupTestRouter(t)

	// Bulbasaur already exists in seeded mock data.
	form := url.Values{
		"name":   {"Bulbasaur"},
		"type1":  {"Grass"},
		"height": {"0.7 m"},
		"weight": {"6.9 kg"},
	}
	w := doRequest(r, http.MethodPost, "/pokemon/create", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate name, got %d", w.Code)
	}
}

func TestPokemonUpdate(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{
		"name":            {"Pikachu"},
		"type1":           {"Electric"},
		"type2":           {"Steel"},
		"height":          {"0.4 m"},
		"weight":          {"6.0 kg"},
		"base_experience": {"200"},
		"image_url":       {"https://example.com/pikachu.png"},
		"description":     {"Updated."},
		"evolves_from":    {""},
	}
	w := doRequest(r, http.MethodPost, "/pokemon/25/update", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect 302, got %d", w.Code)
	}

	p, err := database.GetPokemonByID(25)
	if err != nil {
		t.Fatalf("GetPokemonByID failed: %v", err)
	}
	if p.Type2 != "Steel" {
		t.Errorf("expected Type2 Steel, got %q", p.Type2)
	}
	if p.BaseExperience != 200 {
		t.Errorf("expected BaseExperience 200, got %d", p.BaseExperience)
	}
}

func TestPokemonUpdateNotFound(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{
		"name":   {"Nope"},
		"type1":  {"Normal"},
		"height": {"1.0 m"},
		"weight": {"1.0 kg"},
	}
	w := doRequest(r, http.MethodPost, "/pokemon/9999/update", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPokemonUpdateValidation(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{
		"type1":  {"Electric"},
		"height": {"0.4 m"},
		"weight": {"6.0 kg"},
	}
	w := doRequest(r, http.MethodPost, "/pokemon/25/update", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestPokemonDelete(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodDelete, "/pokemon/25", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "deleted") {
		t.Error("expected 'deleted' response body")
	}

	p, err := database.GetPokemonByID(25)
	if err != nil {
		t.Fatalf("GetPokemonByID failed: %v", err)
	}
	if p != nil {
		t.Errorf("expected pokemon 25 to be deleted, got %+v", p)
	}
}

func TestPokemonDeleteInvalidID(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodDelete, "/pokemon/notanumber", nil, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPokemonSearch(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{"query": {"pika"}}
	w := doRequest(r, http.MethodPost, "/pokemon/search", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pikachu") {
		t.Error("expected search result to contain Pikachu")
	}
}

func TestPokemonSearchEmptyQuery(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{"query": {"   "}}
	w := doRequest(r, http.MethodPost, "/pokemon/search", strings.NewReader(form.Encode()), map[string]string{"HX-Request": "true"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Empty query falls back to the full list.
	if !strings.Contains(w.Body.String(), "Bulbasaur") {
		t.Error("expected empty query to return full list")
	}
}

func TestPokemonSearchNoResults(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{"query": {"zzzzzz"}}
	w := doRequest(r, http.MethodPost, "/pokemon/search", strings.NewReader(form.Encode()), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No Pokémon matching") {
		t.Error("expected no-results message")
	}
}

func TestStats(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/stats", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := strings.TrimSpace(w.Body.String()); got != "6" {
		t.Errorf("expected stats count 6, got %q", got)
	}
}

// ---- Pure helper tests ----

func TestParseInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"112", 112},
		{" 64 ", 64},
		{"", 0},
		{"abc", 0},
		{"-5", -5},
	}
	for _, c := range cases {
		if got := parseInt(c.in); got != c.want {
			t.Errorf("parseInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestBuildPagination(t *testing.T) {
	// Single page -> empty.
	if got := buildPagination(1, 1, 6); got != "" {
		t.Errorf("single page should be empty, got %q", got)
	}

	// Middle page includes prev + next + page numbers.
	got := buildPagination(2, 3, 30)
	for _, want := range []string{"Prev", "Next", ">2<"} {
		if !strings.Contains(got, want) {
			t.Errorf("pagination missing %q: %s", want, got)
		}
	}

	// First page has no prev.
	first := buildPagination(1, 3, 30)
	if strings.Contains(first, "Prev") {
		t.Error("first page should not contain Prev")
	}

	// Last page has no next.
	last := buildPagination(3, 3, 30)
	if strings.Contains(last, "Next") {
		t.Error("last page should not contain Next")
	}
}

func TestRenderPokemonGridItem(t *testing.T) {
	p := models.Pokemon{
		ID:          25,
		Name:        "Pikachu",
		Type1:       "Electric",
		Type2:       "",
		Height:      "0.4 m",
		Weight:      "6.0 kg",
		Description: "Electric mouse.",
	}

	html := renderPokemonGridItem(p)
	for _, want := range []string{"Pikachu", "Electric", "0.4 m", "6.0 kg", "No.25", "View Details"} {
		if !strings.Contains(html, want) {
			t.Errorf("grid item missing %q", want)
		}
	}
	// Electric type maps to a yellow badge.
	if !strings.Contains(html, "bg-yellow-500/90") {
		t.Error("expected Electric type badge colour")
	}
}

func TestRenderPokemonGridItemNoImage(t *testing.T) {
	p := models.Pokemon{
		ID:     150,
		Name:   "Mewtwo",
		Type1:  "Psychic",
		Type2:  "",
		Height: "2.0 m",
		Weight: "122.0 kg",
	}

	html := renderPokemonGridItem(p)
	if strings.Contains(html, "<img") {
		t.Error("expected no img tag when ImageURL is empty")
	}
	if !strings.Contains(html, "bg-purple-500/90") {
		t.Error("expected Psychic type badge colour")
	}
}

func TestRenderPokemonGridItemDualType(t *testing.T) {
	p := models.Pokemon{
		ID:     1,
		Name:   "Bulbasaur",
		Type1:  "Grass",
		Type2:  "Poison",
		Height: "0.7 m",
		Weight: "6.9 kg",
	}

	html := renderPokemonGridItem(p)
	if !strings.Contains(html, "Poison") {
		t.Error("expected secondary type badge to render")
	}
	if !strings.Contains(html, "bg-green-500/90") {
		t.Error("expected Grass type badge colour")
	}
}

// pokemonIDByName returns the ID of the first pokemon with the given name,
// or 0 if none exists.
func pokemonIDByName(name string) (int64, error) {
	rows, err := database.DB.Query(`SELECT id FROM pokemon WHERE name = ?`, name)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, nil
	}
	var id int64
	if err := rows.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
