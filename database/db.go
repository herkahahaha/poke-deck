package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// ConnectDB opens the SQLite database and runs migrations.
func ConnectDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if err = migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

// migrate creates the pokemon table if it doesn't exist.
func migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS pokemon (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		name            TEXT    NOT NULL UNIQUE,
		type1           TEXT    NOT NULL,
		type2           TEXT,
		height          TEXT    NOT NULL,
		weight          TEXT    NOT NULL,
		base_experience INTEGER NOT NULL DEFAULT 0,
		image_url       TEXT,
		description     TEXT,
		evolves_from   TEXT,
		created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(schema)
	return err
}

// CloseDB closes the database connection.
func CloseDB() error {
	if DB == nil {
		return nil
	}
	return DB.Close()
}
