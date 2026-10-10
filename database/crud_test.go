package database

import (
	"errors"
	"strings"
	"testing"

	"pokedex-go/mocks"
	"pokedex-go/models"
)

// setupTestDB connects to a fresh in-memory database and seeds it with
// the shared mock Pokémon data. It returns a cleanup func.
func setupTestDB(t *testing.T) {
	t.Helper()
	if err := ConnectDB(":memory:"); err != nil {
		t.Fatalf("ConnectDB failed: %v", err)
	}
	// A single connection guarantees the in-memory DB persists across queries.
	DB.SetMaxOpenConns(1)

	if err := BulkCreatePokemon(mocks.MockPokemons()); err != nil {
		t.Fatalf("seeding mock data failed: %v", err)
	}

	t.Cleanup(func() {
		if err := CloseDB(); err != nil {
			t.Errorf("CloseDB failed: %v", err)
		}
	})
}

func TestCreatePokemon(t *testing.T) {
	setupTestDB(t)

	p := models.Pokemon{
		Name:              "Testmon",
		Type1:             "Normal",
		Type2:             "",
		Height:            "1.0 m",
		Weight:            "10.0 kg",
		BaseExperience:    50,
		ImageURL:          "https://example.com/testmon.png",
		Description:       "A test Pokémon.",
		EvolvesFromStatus: "",
	}

	created, err := CreatePokemon(p)
	if err != nil {
		t.Fatalf("CreatePokemon returned error: %v", err)
	}

	if created.ID == 0 {
		t.Error("expected a non-zero ID to be assigned")
	}
	if created.Name != p.Name {
		t.Errorf("expected name %q, got %q", p.Name, created.Name)
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Error("expected timestamps to be set")
	}
}

func TestCreatePokemonDuplicateName(t *testing.T) {
	setupTestDB(t)

	dup := mocks.MockPokemons()[0] // Bulbasaur already seeded
	_, err := CreatePokemon(dup)
	if err == nil {
		t.Fatal("expected error creating duplicate name, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Errorf("expected UNIQUE constraint error, got: %v", err)
	}
}

func TestGetAllPokemon(t *testing.T) {
	setupTestDB(t)

	all, err := GetAllPokemon()
	if err != nil {
		t.Fatalf("GetAllPokemon returned error: %v", err)
	}

	if len(all) != len(mocks.MockPokemons()) {
		t.Fatalf("expected %d pokemon, got %d", len(mocks.MockPokemons()), len(all))
	}

	// Ordered by ID ascending.
	if all[0].Name != "Bulbasaur" {
		t.Errorf("expected first pokemon Bulbasaur, got %q", all[0].Name)
	}
}

func TestGetPokemonPage(t *testing.T) {
	setupTestDB(t)

	page, err := GetPokemonPage(0, 2)
	if err != nil {
		t.Fatalf("GetPokemonPage returned error: %v", err)
	}
	if len(page) != 2 {
		t.Fatalf("expected 2 pokemon on first page, got %d", len(page))
	}
	if page[0].Name != "Bulbasaur" {
		t.Errorf("expected Bulbasaur first, got %q", page[0].Name)
	}

	page2, err := GetPokemonPage(2, 2)
	if err != nil {
		t.Fatalf("GetPokemonPage returned error: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 pokemon on second page, got %d", len(page2))
	}
	if page2[0].Name != "Squirtle" {
		t.Errorf("expected Squirtle on second page, got %q", page2[0].Name)
	}
}

func TestCountPokemon(t *testing.T) {
	setupTestDB(t)

	count, err := CountPokemon()
	if err != nil {
		t.Fatalf("CountPokemon returned error: %v", err)
	}
	if count != len(mocks.MockPokemons()) {
		t.Errorf("expected count %d, got %d", len(mocks.MockPokemons()), count)
	}
}

func TestGetPokemonByID(t *testing.T) {
	setupTestDB(t)

	p, err := GetPokemonByID(25)
	if err != nil {
		t.Fatalf("GetPokemonByID returned error: %v", err)
	}
	if p == nil {
		t.Fatal("expected pokemon with ID 25")
	}
	if p.Name != "Pikachu" {
		t.Errorf("expected Pikachu, got %q", p.Name)
	}
}

func TestGetPokemonByIDNotFound(t *testing.T) {
	setupTestDB(t)

	p, err := GetPokemonByID(9999)
	if err != nil {
		t.Fatalf("GetPokemonByID returned error: %v", err)
	}
	if p != nil {
		t.Errorf("expected nil for missing ID, got %+v", p)
	}
}

func TestUpdatePokemon(t *testing.T) {
	setupTestDB(t)

	update := models.Pokemon{
		Name:              "Pikachu",
		Type1:             "Electric",
		Type2:             "Steel",
		Height:            "0.4 m",
		Weight:            "6.0 kg",
		BaseExperience:    200,
		ImageURL:          "https://example.com/pikachu2.png",
		Description:       "Updated description.",
		EvolvesFromStatus: "",
	}

	updated, err := UpdatePokemon(25, update)
	if err != nil {
		t.Fatalf("UpdatePokemon returned error: %v", err)
	}

	if updated.Type2 != "Steel" {
		t.Errorf("expected Type2 Steel, got %q", updated.Type2)
	}
	if updated.BaseExperience != 200 {
		t.Errorf("expected BaseExperience 200, got %d", updated.BaseExperience)
	}
	if updated.Name != "Pikachu" {
		t.Errorf("expected name unchanged, got %q", updated.Name)
	}
}

func TestUpdatePokemonNotFound(t *testing.T) {
	setupTestDB(t)

	update := mocks.MockPokemons()[0]
	updated, err := UpdatePokemon(9999, update)
	if err != nil {
		t.Fatalf("UpdatePokemon returned error: %v", err)
	}
	if updated != nil {
		t.Errorf("expected nil for missing ID, got %+v", updated)
	}
}

func TestDeletePokemon(t *testing.T) {
	setupTestDB(t)

	if err := DeletePokemon(25); err != nil {
		t.Fatalf("DeletePokemon returned error: %v", err)
	}

	p, err := GetPokemonByID(25)
	if err != nil {
		t.Fatalf("GetPokemonByID returned error: %v", err)
	}
	if p != nil {
		t.Errorf("expected pokemon 25 to be deleted, got %+v", p)
	}
}

func TestDeletePokemonNotFound(t *testing.T) {
	setupTestDB(t)

	err := DeletePokemon(9999)
	if err == nil {
		t.Fatal("expected error deleting missing pokemon, got nil")
	}
	if !errors.Is(err, errors.New("pokemon not found")) && err.Error() != "pokemon not found" {
		t.Errorf("expected 'pokemon not found' error, got: %v", err)
	}
}

func TestBulkCreatePokemon(t *testing.T) {
	setupTestDB(t)

	extra := []models.Pokemon{
		{
			ID:             201,
			Name:           "Unown A",
			Type1:          "Psychic",
			Height:         "0.5 m",
			Weight:         "5.0 kg",
			BaseExperience: 118,
		},
		{
			ID:             202,
			Name:           "Unown B",
			Type1:          "Psychic",
			Height:         "0.5 m",
			Weight:         "5.0 kg",
			BaseExperience: 118,
		},
	}

	if err := BulkCreatePokemon(extra); err != nil {
		t.Fatalf("BulkCreatePokemon returned error: %v", err)
	}

	count, err := CountPokemon()
	if err != nil {
		t.Fatalf("CountPokemon returned error: %v", err)
	}
	if count != len(mocks.MockPokemons())+len(extra) {
		t.Errorf("expected count %d, got %d", len(mocks.MockPokemons())+len(extra), count)
	}
}

func TestBulkCreatePokemonEmpty(t *testing.T) {
	setupTestDB(t)

	if err := BulkCreatePokemon(nil); err != nil {
		t.Fatalf("BulkCreatePokemon with empty slice returned error: %v", err)
	}

	count, _ := CountPokemon()
	if count != len(mocks.MockPokemons()) {
		t.Errorf("expected count unchanged at %d, got %d", len(mocks.MockPokemons()), count)
	}
}
