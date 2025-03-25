package main

import (
	"fmt"
	"log"
	"net/http"
	"nostr-hero/src/api"
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

	mux := http.NewServeMux()

	// Initialize Routes
	routes.InitializeRoutes(mux)

	mux.HandleFunc("/api/character", api.CharacterHandler)

	port := utils.AppConfig.Server.Port
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}