package main

import (
	"fmt"
	"log"
	"net/http"
	"nostr-hero/src/api"
	"nostr-hero/src/handlers"
	"nostr-hero/src/routes"
	"nostr-hero/src/utils"
	"time"
)

func main() {
	// Default config path
	configPath := "config.yml"

	// Load config
	if err := utils.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize session manager
	sessionMgr := handlers.SessionMgr
	
	// Set up periodic session cleanup
	go func() {
		for {
			time.Sleep(6 * time.Hour)
			sessionMgr.CleanupSessions(168 * time.Hour) // Clean sessions older than 24 hours
			log.Println("🧹 Session cleanup completed")
		}
	}()

	mux := http.NewServeMux()

	// Initialize Routes
	routes.InitializeRoutes(mux)

	mux.HandleFunc("/api/character", api.CharacterHandler)
	mux.HandleFunc("/init-user", handlers.InitUser)
	mux.HandleFunc("/get-session", handlers.GetSessionHandler)
	mux.HandleFunc("/get-cache", handlers.GetCacheHandler)

	port := utils.AppConfig.Server.Port
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
