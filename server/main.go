package main

import (
	"fmt"
	"log"
	"net/http"
	"nostr-hero/api"
	"nostr-hero/auth"
	"nostr-hero/cache"
	"nostr-hero/db"
	"nostr-hero/routes"
	"nostr-hero/utils"
	"time"
)

func main() {
	// Default config path
	configPath := "config.yml"

	// Load config
	if err := utils.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := db.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize grain client for authentication
	if err := auth.InitializeGrainClient(&utils.AppConfig); err != nil {
		log.Fatalf("Failed to initialize Grain client: %v", err)
	}
	defer auth.ShutdownGrainClient()

	// Initialize profile cache (24 hour TTL)
	cache.InitProfileCache(24 * time.Hour)
	log.Println("✅ Profile cache initialized (24h TTL)")

	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("www/res/"))))
	mux.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("www/dist/"))))

	// Serve data files
	mux.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("game-data/"))))

	// Initialize Routes
	routes.InitializeRoutes(mux)

	// Game data API endpoints
	mux.HandleFunc("/api/game-data", api.GameDataHandler)
	mux.HandleFunc("/api/items", api.ItemsHandler)
	mux.HandleFunc("/api/spells/", api.SpellsHandler) // Note the trailing slash to match /api/spells/{id}
	mux.HandleFunc("/api/monsters", api.MonstersHandler)
	mux.HandleFunc("/api/locations", api.LocationsHandler)
	mux.HandleFunc("/api/npcs", api.NPCsHandler)
	mux.HandleFunc("/api/npcs/at-location", api.GetNPCsAtLocationHandler)
	mux.HandleFunc("/api/abilities", api.AbilitiesHandler)

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

	// Session management API endpoints (in-memory state)
	mux.HandleFunc("/api/session/init", api.InitSessionHandler)
	mux.HandleFunc("/api/session/reload", api.ReloadSessionHandler)
	mux.HandleFunc("/api/session/state", api.GetSessionHandler)
	mux.HandleFunc("/api/session/update", api.UpdateSessionHandler)
	mux.HandleFunc("/api/session/save", api.SaveSessionHandler)
	mux.HandleFunc("/api/session/cleanup", api.CleanupSessionHandler)

	// Game action API endpoints (Go-first game logic)
	mux.HandleFunc("/api/game/action", api.GameActionHandler)
	mux.HandleFunc("/api/game/state", api.GetGameStateHandler)

	// Debug endpoints (only enabled in debug mode)
	if utils.AppConfig.Server.DebugMode {
		log.Println("🐛 Debug mode enabled")
		mux.HandleFunc("/api/debug/sessions", func(w http.ResponseWriter, r *http.Request) {
			api.DebugSessionsHandler(w, r, true)
		})
		mux.HandleFunc("/api/debug/state", func(w http.ResponseWriter, r *http.Request) {
			api.DebugStateHandler(w, r, true)
		})
	}

	// Shop API endpoints
	mux.HandleFunc("/api/shop/", api.ShopHandler)

	// Existing API endpoints
	mux.HandleFunc("/api/character", api.CharacterHandler)
	mux.HandleFunc("/api/character/create-save", api.CreateCharacterHandler)
	mux.HandleFunc("/api/profile", api.ProfileHandler)

	port := utils.AppConfig.Server.Port
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
