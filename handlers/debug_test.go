package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url"

	"github.com/gin-gonic/gin"
	"pokedex-go/database"
	"pokedex-go/mocks"
	"pokedex-go/models"
)

func TestPokemonSearchNoResultsDebug(t *testing.T) {
	r := setupTestRouter(t)

	form := url.Values{"query": {"zzzzzz"}}
	w := doRequest(r, http.MethodPost, "/pokemon/search", strings.NewReader(form.Encode()), nil)
	t.Logf("status: %d", w.Code)
	t.Logf("body length: %d", len(w.Body.String()))
	t.Logf("body: %s", w.Body.String())
	if strings.Contains(w.Body.String(), "No Pokémon matching") {
		t.Log("FOUND 'No Pokémon matching'")
	} else {
		t.Log("NOT FOUND 'No Pokémon matching'")
	}
	// Check what IS in the body
	for _, substr := range []string{"No Pokémon", "empty", "list_empty", "matching", "results"} {
		if strings.Contains(w.Body.String(), substr) {
			t.Logf("body contains '%s'", substr)
		}
	}
}
