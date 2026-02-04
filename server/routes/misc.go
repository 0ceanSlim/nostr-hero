package routes

import (
	"net/http"
	"nostr-hero/utils"
)

// Discover serves the discover page
func Discover(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "discover",
	}

	utils.RenderTemplate(w, data, "discover.html", false)
}

// SettingsHandler serves the settings page
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "Settings",
	}

	utils.RenderTemplate(w, data, "settings.html", false)
}
