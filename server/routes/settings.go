package routes

import (
	"net/http"
	"nostr-hero/utils"
)

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "Settings",
	}

	utils.RenderTemplate(w, data, "settings.html", false)
}
