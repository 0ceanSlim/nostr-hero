package routes

import (
	"net/http"
	"nostr-hero/utils"
)

// GameHandler serves the main game interface
func GameHandler(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	saveID := r.URL.Query().Get("save")
	newGame := r.URL.Query().Get("new")

	data := utils.PageData{
		Title: "Nostr Hero - Game",
		Theme: "dark",
		CustomData: map[string]interface{}{
			"GameMode":  true,
			"SaveID":    saveID,
			"NewGame":   newGame == "true",
			"DebugMode": utils.AppConfig.Server.DebugMode,
		},
	}

	utils.RenderTemplateWithLayout(w, data, "game.html", false, true)
}