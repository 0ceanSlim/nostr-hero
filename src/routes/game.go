package routes

import (
	"net/http"
	"nostr-hero/src/utils"
)

// GameHandler serves the main game interface
func GameHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in (you may want to implement proper session checking)
	// For now, we'll serve the game interface directly

	data := utils.PageData{
		Title: "Nostr Hero - Game",
		Theme: "dark",
		CustomData: map[string]interface{}{
			"GameMode": true,
		},
	}

	utils.RenderTemplate(w, data, "game.html", false)
}