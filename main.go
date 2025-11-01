package main

import (
	"fmt"
	"log"
	"net/http"
	"nostr-hero/src/api"
	"nostr-hero/src/auth"
	"nostr-hero/src/db"
	"nostr-hero/src/routes"
	"nostr-hero/src/utils"
)

func main() {
	// Default config path
	configPath := "config.yml"

	// Load config
	if err := utils.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize grain client for authentication
	if err := auth.InitializeGrainClient(&utils.AppConfig); err != nil {
		log.Fatalf("Failed to initialize Grain client: %v", err)
	}
	defer auth.ShutdownGrainClient()

	// Initialize database
	if err := db.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migration on startup (can be made conditional with a flag later)
	if err := db.MigrateFromJSON(); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("www/scripts/"))))
	mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("www/res/"))))

	// Serve data files
	mux.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("docs/data/"))))

	// Initialize Routes
	routes.InitializeRoutes(mux)

	// Game data API endpoints
	mux.HandleFunc("/api/game-data", api.GameDataHandler)
	mux.HandleFunc("/api/items", api.ItemsHandler)
	mux.HandleFunc("/api/spells/", api.SpellsHandler) // Note the trailing slash to match /api/spells/{id}
	mux.HandleFunc("/api/monsters", api.MonstersHandler)
	mux.HandleFunc("/api/locations", api.LocationsHandler)
	mux.HandleFunc("/api/npcs", api.NPCsHandler)

	// Character generation API endpoints
	mux.HandleFunc("/api/weights", api.WeightsHandler)
	mux.HandleFunc("/api/introductions", api.IntroductionsHandler)
	mux.HandleFunc("/api/starting-gear", api.StartingGearHandler)

	// Authentication endpoints (using grain client)
	authHandler := auth.NewAuthHandler(&utils.AppConfig)
	mux.HandleFunc("/api/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/api/auth/logout", authHandler.HandleLogout)
	mux.HandleFunc("/api/auth/session", authHandler.HandleSession)
	mux.HandleFunc("/api/auth/generate-keys", authHandler.HandleGenerateKeys)
	mux.HandleFunc("/api/auth/amber-callback", authHandler.HandleAmberCallback)

	// Save/load API endpoints
	mux.HandleFunc("/api/saves/", api.SavesHandler)

	// Inventory API endpoints
	mux.HandleFunc("/api/inventory/action", api.InventoryHandler)
	mux.HandleFunc("/api/inventory/actions", api.ItemActionsHandler)

	// Existing API endpoints
	mux.HandleFunc("/api/character", api.CharacterHandler)
	mux.HandleFunc("/api/profile", api.ProfileHandler)

	port := utils.AppConfig.Server.Port
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
