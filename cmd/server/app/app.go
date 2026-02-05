package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"pubkey-quest/cmd/server/auth"
	"pubkey-quest/cmd/server/cache"
	"pubkey-quest/cmd/server/db"
	"pubkey-quest/cmd/server/utils"
)

// Init initializes all application services
func Init() {
	if err := db.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := auth.InitializeGrainClient(&utils.AppConfig); err != nil {
		log.Fatalf("Failed to initialize Grain client: %v", err)
	}

	cache.InitProfileCache(24 * time.Hour)
	log.Println("✅ All services initialized")
}

// Shutdown cleans up all application services
func Shutdown() {
	db.Close()
	auth.ShutdownGrainClient()
	log.Println("✅ All services shut down")
}

// Start starts the HTTP server
func Start(mux *http.ServeMux) {
	port := utils.AppConfig.Server.Port
	fmt.Printf("Server running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
