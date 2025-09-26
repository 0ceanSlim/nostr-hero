package routes

import (
	"net/http"
	"nostr-hero/src/utils"
)

// NewGameHandler serves the character generation page
func NewGameHandler(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "Nostr Hero - New Adventure",
		Theme: "dark",
		CustomData: map[string]interface{}{
			"NewGameMode": true,
		},
	}

	utils.RenderTemplate(w, data, "new-game.html", false)
}