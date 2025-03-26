package routes

import (
	"net/http"
	"nostr-hero/src/utils"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		fileServer := http.FileServer(http.Dir("web"))
		http.StripPrefix("/", fileServer).ServeHTTP(w, r)
		return
	}

	data := utils.PageData{
		Title: "discover",
	}

	utils.RenderTemplate(w, data, "index.html", false)
}
