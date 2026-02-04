package main

import (
	"log"
	"net/http"

	"nostr-hero/api"
	"nostr-hero/app"
	"nostr-hero/routes"
	"nostr-hero/utils"
)

func main() {
	if err := utils.LoadConfig("config.yml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app.Init()
	defer app.Shutdown()

	mux := http.NewServeMux()

	routes.RegisterStatic(mux)
	routes.RegisterPages(mux)
	api.RegisterRoutes(mux)

	app.Start(mux)
}
