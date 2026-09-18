package models

// Pokemon represents a Pokémon entry in the database.
type Pokemon struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Type1             string `json:"type1"`
	Type2             string `json:"type2,omitempty"`
	Height            string `json:"height"`
	Weight            string `json:"weight"`
	BaseExperience    int    `json:"base_experience"`
	ImageURL          string `json:"image_url"`
	Description       string `json:"description"`
	EvolvesFromStatus string `json:"evolves_from_status,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}
