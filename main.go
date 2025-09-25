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

	// Serve resource files
	mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("www/res/"))))

	// Initialize Routes
	routes.InitializeRoutes(mux)

	// Game data API endpoints
	mux.HandleFunc("/api/game-data", api.GameDataHandler)
	mux.HandleFunc("/api/items", api.ItemsHandler)
	mux.HandleFunc("/api/spells", api.SpellsHandler)
	mux.HandleFunc("/api/monsters", api.MonstersHandler)
	mux.HandleFunc("/api/locations", api.LocationsHandler)

	// Authentication endpoints (using grain client)
	authHandler := auth.NewAuthHandler(&utils.AppConfig)
	mux.HandleFunc("/api/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/api/auth/logout", authHandler.HandleLogout)
	mux.HandleFunc("/api/auth/session", authHandler.HandleSession)
	mux.HandleFunc("/api/auth/generate-keys", authHandler.HandleGenerateKeys)
	mux.HandleFunc("/api/auth/amber-callback", authHandler.HandleAmberCallback)

	// Save/load API endpoints
	mux.HandleFunc("/api/save-game", api.SaveGameHandler)
	mux.HandleFunc("/api/load-save/", api.LoadSaveHandler)

	// Existing API endpoints
	mux.HandleFunc("/api/character", api.CharacterHandler)

	port := utils.AppConfig.Server.Port
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
