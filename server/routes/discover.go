package routes

import (
	"net/http"
	"nostr-hero/utils"
)

func Discover(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "discover",
	}

	utils.RenderTemplate(w, data, "discover.html", false)
}
