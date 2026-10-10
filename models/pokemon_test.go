package models

import (
	"encoding/json"
	"testing"
)

// TestPokemonMarshalJSON verifies that a Pokemon marshals to the expected
// JSON keys and that empty optional fields are omitted.
func TestPokemonMarshalJSON(t *testing.T) {
	p := Pokemon{
		ID:             25,
		Name:           "Pikachu",
		Type1:          "Electric",
		Type2:          "",
		Height:         "0.4 m",
		Weight:         "6.0 kg",
		BaseExperience: 112,
		ImageURL:       "https://example.com/pikachu.png",
		Description:    "Electric mouse.",
		CreatedAt:      "2026-01-01 00:00:00",
		UpdatedAt:      "2026-01-01 00:00:00",
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	got := string(data)
	wantFields := []string{
		`"id":25`,
		`"name":"Pikachu"`,
		`"type1":"Electric"`,
		`"height":"0.4 m"`,
		`"weight":"6.0 kg"`,
		`"base_experience":112`,
		`"image_url":"https://example.com/pikachu.png"`,
		`"description":"Electric mouse."`,
	}
	for _, field := range wantFields {
		if !contains(got, field) {
			t.Errorf("marshaled JSON missing %s; got %s", field, got)
		}
	}

	// Empty Type2 should be omitted due to omitempty.
	if contains(got, `"type2"`) {
		t.Errorf("empty Type2 should be omitted; got %s", got)
	}
	if contains(got, `"evolves_from_status"`) {
		t.Errorf("empty EvolvesFromStatus should be omitted; got %s", got)
	}
}

// TestPokemonMarshalJSONWithOptionalFields ensures populated optional
// fields (Type2, EvolvesFromStatus) are included.
func TestPokemonMarshalJSONWithOptionalFields(t *testing.T) {
	p := Pokemon{
		ID:                1,
		Name:              "Bulbasaur",
		Type1:             "Grass",
		Type2:             "Poison",
		EvolvesFromStatus: "Old Bulbasaur",
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	got := string(data)
	if !contains(got, `"type2":"Poison"`) {
		t.Errorf("expected type2 to be present; got %s", got)
	}
	if !contains(got, `"evolves_from_status":"Old Bulbasaur"`) {
		t.Errorf("expected evolves_from_status to be present; got %s", got)
	}
}

// TestPokemonUnmarshalJSON verifies that JSON decodes back into a Pokemon.
func TestPokemonUnmarshalJSON(t *testing.T) {
	input := `{
		"id": 4,
		"name": "Charmander",
		"type1": "Fire",
		"type2": "",
		"height": "0.6 m",
		"weight": "8.5 kg",
		"base_experience": 62,
		"image_url": "https://example.com/charmander.png",
		"description": "Lizard.",
		"evolves_from_status": "Old Charmander",
		"created_at": "2026-01-01 00:00:00",
		"updated_at": "2026-01-01 00:00:00"
	}`

	var p Pokemon
	if err := json.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if p.ID != 4 || p.Name != "Charmander" || p.Type1 != "Fire" {
		t.Errorf("unexpected pokemon: %+v", p)
	}
	if p.BaseExperience != 62 {
		t.Errorf("expected base_experience 62, got %d", p.BaseExperience)
	}
	if p.EvolvesFromStatus != "Old Charmander" {
		t.Errorf("expected evolves_from_status, got %q", p.EvolvesFromStatus)
	}
}

// TestPokemonJSONRoundTrip encodes a Pokemon, decodes it, and compares fields.
func TestPokemonJSONRoundTrip(t *testing.T) {
	original := Pokemon{
		ID:             39,
		Name:           "Jigglypuff",
		Type1:          "Normal",
		Type2:          "Fairy",
		Height:         "0.5 m",
		Weight:         "5.5 kg",
		BaseExperience: 95,
		ImageURL:       "https://example.com/jigglypuff.png",
		Description:    "Balloon.",
		CreatedAt:      "2026-01-01 00:00:00",
		UpdatedAt:      "2026-01-01 00:00:00",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var decoded Pokemon
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if original != decoded {
		t.Errorf("round trip mismatch:\noriginal: %+v\ndecoded:  %+v", original, decoded)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
