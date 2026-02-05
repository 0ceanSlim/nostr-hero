package main

import (
	"log"
	"net/http"

	"pubkey-quest/cmd/server/api"
	"pubkey-quest/cmd/server/app"
	"pubkey-quest/cmd/server/routes"
	"pubkey-quest/cmd/server/utils"
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
