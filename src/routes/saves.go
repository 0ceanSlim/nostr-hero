package routes

import (
	"net/http"
	"nostr-hero/src/utils"
)

// SavesHandler serves the save selection interface
func SavesHandler(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "Nostr Hero - Select Save",
		Theme: "dark",
		CustomData: map[string]interface{}{
			"SavesMode": true,
		},
	}

	utils.RenderTemplate(w, data, "saves.html", false)
}