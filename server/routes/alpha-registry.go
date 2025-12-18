package routes

import (
	"encoding/json"
	"net/http"
	"nostr-hero/utils"
	"os"
)

func AlphaRegistry(w http.ResponseWriter, r *http.Request) {
	// Read registry.json file
	file, err := os.ReadFile("www/data/alpha-registry.json")
	if err != nil {
		http.Error(w, "Error reading registry", http.StatusInternalServerError)
		return
	}

	var alphaRegistry []map[string]interface{}
	if err := json.Unmarshal(file, &alphaRegistry); err != nil {
		http.Error(w, "Error parsing legacy Registry", http.StatusInternalServerError)
		return
	}

	data := utils.PageData{
		Title:      "Alpha Characters",
		CustomData: map[string]interface{}{"alphaRegistry": alphaRegistry},
	}

	utils.RenderTemplate(w, data, "alpha-registry.html", false)
}
