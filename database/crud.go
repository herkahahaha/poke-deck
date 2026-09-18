package database

import (
	"database/sql"
	"fmt"
	"time"

	"pokedex-go/models"
)

// CRUD operations for Pokemon.

func CreatePokemon(p models.Pokemon) (*models.Pokemon, error) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	query := `
	INSERT INTO pokemon (name, type1, type2, height, weight, base_experience, image_url, description, evolves_from, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := DB.Exec(query, p.Name, p.Type1, p.Type2, p.Height, p.Weight,
		p.BaseExperience, p.ImageURL, p.Description, p.EvolvesFromStatus, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert pokemon: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("lastinsertid: %w", err)
	}

	return GetPokemonByID(id)
}

func GetAllPokemon() ([]models.Pokemon, error) {
	query := `
	SELECT id, name, type1, type2, height, weight, base_experience,
	       image_url, description, evolves_from, created_at, updated_at
	FROM pokemon
	ORDER BY id ASC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var pokemons []models.Pokemon
	for rows.Next() {
		var p models.Pokemon
		if err := rows.Scan(&p.ID, &p.Name, &p.Type1, &p.Type2, &p.Height, &p.Weight,
			&p.BaseExperience, &p.ImageURL, &p.Description, &p.EvolvesFromStatus, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		pokemons = append(pokemons, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return pokemons, nil
}

// BulkCreatePokemon inserts multiple Pokémon in a single transaction.
func BulkCreatePokemon(pokemons []models.Pokemon) error {
	if len(pokemons) == 0 {
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO pokemon (id, name, type1, type2, height, weight, base_experience, image_url, description, evolves_from, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	for _, p := range pokemons {
		_, err := stmt.Exec(p.ID, p.Name, p.Type1, p.Type2, p.Height, p.Weight,
			p.BaseExperience, p.ImageURL, p.Description, p.EvolvesFromStatus, now, now)
		if err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}

	return tx.Commit()
}

// GetPokemonPage returns a paginated list of Pokémon with LIMIT/OFFSET.
func GetPokemonPage(offset, limit int) ([]models.Pokemon, error) {
	query := `
	SELECT id, name, type1, type2, height, weight, base_experience,
	       image_url, description, evolves_from, created_at, updated_at
	FROM pokemon
	ORDER BY id ASC
	LIMIT ? OFFSET ?`

	rows, err := DB.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var pokemons []models.Pokemon
	for rows.Next() {
		var p models.Pokemon
		if err := rows.Scan(&p.ID, &p.Name, &p.Type1, &p.Type2, &p.Height, &p.Weight,
			&p.BaseExperience, &p.ImageURL, &p.Description, &p.EvolvesFromStatus, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		pokemons = append(pokemons, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return pokemons, nil
}

// CountPokemon returns the total number of Pokémon in the database.
func CountPokemon() (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM pokemon`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

func GetPokemonByID(id int64) (*models.Pokemon, error) {
	query := `
	SELECT id, name, type1, type2, height, weight, base_experience,
	       image_url, description, evolves_from, created_at, updated_at
	FROM pokemon
	WHERE id = ?`

	var p models.Pokemon
	err := DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Type1, &p.Type2, &p.Height, &p.Weight,
		&p.BaseExperience, &p.ImageURL, &p.Description, &p.EvolvesFromStatus, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query row: %w", err)
	}
	return &p, nil
}

func UpdatePokemon(id int64, p models.Pokemon) (*models.Pokemon, error) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	query := `
	UPDATE pokemon SET
		name            = ?,
		type1           = ?,
		type2           = ?,
		height          = ?,
		weight          = ?,
		base_experience = ?,
		image_url       = ?,
		description     = ?,
		evolves_from    = ?,
		updated_at      = ?
	WHERE id = ?`

	_, err := DB.Exec(query, p.Name, p.Type1, p.Type2, p.Height, p.Weight,
		p.BaseExperience, p.ImageURL, p.Description, p.EvolvesFromStatus, now, id)
	if err != nil {
		return nil, fmt.Errorf("update pokemon: %w", err)
	}

	return GetPokemonByID(id)
}

func DeletePokemon(id int64) error {
	query := `DELETE FROM pokemon WHERE id = ?`
	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete pokemon: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pokemon not found")
	}
	return nil
}
