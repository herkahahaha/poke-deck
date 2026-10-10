// Package mocks provides reusable mock Pokémon data for unit tests
// across the database, handlers, and seed packages.
package mocks

import "pokedex-go/models"

// MockPokemons returns a slice of representative Pokémon used to seed
// tests and exercise CRUD/handler behaviour without touching the real
// database or PokéAPI.
func MockPokemons() []models.Pokemon {
	return []models.Pokemon{
		{
			ID:             1,
			Name:           "Bulbasaur",
			Type1:          "Grass",
			Type2:          "Poison",
			Height:         "0.7 m",
			Weight:         "6.9 kg",
			BaseExperience: 64,
			ImageURL:       "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/1.png",
			Description:    "A strange seed was planted on its back at birth. The plant sprouts and grows with this Pokémon.",
			CreatedAt:      "2026-01-01 00:00:00",
			UpdatedAt:      "2026-01-01 00:00:00",
		},
		{
			ID:             4,
			Name:           "Charmander",
			Type1:          "Fire",
			Type2:          "",
			Height:         "0.6 m",
			Weight:         "8.5 kg",
			BaseExperience: 62,
			ImageURL:       "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/4.png",
			Description:    "The flame on its tail shows the strength of its life force. If it is weak, the flame also burns weakly.",
			CreatedAt:      "2026-01-01 00:00:00",
			UpdatedAt:      "2026-01-01 00:00:00",
		},
		{
			ID:             7,
			Name:           "Squirtle",
			Type1:          "Water",
			Type2:          "",
			Height:         "0.5 m",
			Weight:         "9.0 kg",
			BaseExperience: 63,
			ImageURL:       "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/7.png",
			Description:    "After birth, its back swells and hardens into a shell. It powerfully sprays foam from its mouth.",
			CreatedAt:      "2026-01-01 00:00:00",
			UpdatedAt:      "2026-01-01 00:00:00",
		},
		{
			ID:             25,
			Name:           "Pikachu",
			Type1:          "Electric",
			Type2:          "",
			Height:         "0.4 m",
			Weight:         "6.0 kg",
			BaseExperience: 112,
			ImageURL:       "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/25.png",
			Description:    "When several of these Pokémon gather, their electricity could build and cause lightning storms.",
			CreatedAt:      "2026-01-01 00:00:00",
			UpdatedAt:      "2026-01-01 00:00:00",
		},
		{
			ID:             39,
			Name:           "Jigglypuff",
			Type1:          "Normal",
			Type2:          "Fairy",
			Height:         "0.5 m",
			Weight:         "5.5 kg",
			BaseExperience: 95,
			ImageURL:       "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/39.png",
			Description:    "When its huge eyes light up, it sings a mysteriously soothing melody that lulls its enemies to sleep.",
			CreatedAt:      "2026-01-01 00:00:00",
			UpdatedAt:      "2026-01-01 00:00:00",
		},
		{
			ID:             150,
			Name:           "Mewtwo",
			Type1:          "Psychic",
			Type2:          "",
			Height:         "2.0 m",
			Weight:         "122.0 kg",
			BaseExperience: 306,
			ImageURL:       "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/150.png",
			Description:    "It was created by a scientist after years of horrific gene-splicing and DNA-engineering experiments.",
			CreatedAt:      "2026-01-01 00:00:00",
			UpdatedAt:      "2026-01-01 00:00:00",
		},
	}
}
