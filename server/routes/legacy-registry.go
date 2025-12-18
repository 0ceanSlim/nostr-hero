package routes

import (
	"encoding/json"
	"net/http"
	"nostr-hero/utils"
	"os"
)

func LegacyRegistry(w http.ResponseWriter, r *http.Request) {
	// Read registry.json file
	file, err := os.ReadFile("wb/data/legacy-registry.json")
	if err != nil {
		http.Error(w, "Error reading registry", http.StatusInternalServerError)
		return
	}

	var legacyRegistry []map[string]interface{}
	if err := json.Unmarshal(file, &legacyRegistry); err != nil {
		http.Error(w, "Error parsing legacy Registry", http.StatusInternalServerError)
		return
	}

	data := utils.PageData{
		Title:      "Legacy Characters",
		CustomData: map[string]interface{}{"legacyRegistry": legacyRegistry},
	}

	utils.RenderTemplate(w, data, "legacy-registry.html", false)
}
