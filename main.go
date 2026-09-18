package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"pokedex-go/database"
	"pokedex-go/handlers"
	"pokedex-go/seed"
)

func main() {
	// CLI flags
	seedFlag := flag.Bool("seed", false, "Seed database from PokéAPI")
	seedLimit := flag.Int("limit", 151, "Number of Pokémon to seed (0 = all, default 151 for Gen 1)")
	flag.Parse()

	// Determine project root (Documents/pokedex-go)
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	projectRoot := filepath.Dir(execPath)

	// Initialize SQLite database in project root
	dbPath := filepath.Join(projectRoot, "pokedex.db")
	if err := database.ConnectDB(dbPath); err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer database.CloseDB()

	// Seed from PokéAPI if requested
	if *seedFlag {
		if err := seed.SeedFromPokeAPI(*seedLimit); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		fmt.Println("✅ Database seeded successfully!")
		return
	}

	// Gin setup
	r := gin.Default()

	// CORS for HTMX cross-origin if needed
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "HX-Request", "HX-Target", "HX-Trigger", "HX-Current-Path", "HX-Active-To"},
	}))

	// Load HTML templates
	r.LoadHTMLGlob(filepath.Join(projectRoot, "templates", "*.html"))
	r.Static("/static", filepath.Join(projectRoot, "static"))

	// === Routes ===

	// Home page
	r.GET("/", handlers.Home)
	r.GET("/index", handlers.Home)

	// Pokémon list (HTMX target)
	r.GET("/pokemon", handlers.PokemonList)
	r.POST("/pokemon/search", handlers.PokemonSearch)

	// Pokémon CRUD
	r.GET("/pokemon/create", handlers.PokemonCreate)
	r.POST("/pokemon/create", handlers.PokemonStore)

	r.GET("/pokemon/:id/edit", handlers.PokemonEdit)
	r.POST("/pokemon/:id/update", handlers.PokemonUpdate)

	r.GET("/pokemon/:id", handlers.PokemonDetail)
	r.DELETE("/pokemon/:id", handlers.PokemonDelete)

	// Stats (HTMX target)
	r.GET("/stats", handlers.Stats)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 PokeDB running at http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
