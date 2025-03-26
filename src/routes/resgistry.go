package routes

import (
	"encoding/json"
	"net/http"
	"nostr-hero/src/utils"
	"os"
)

func Registry(w http.ResponseWriter, r *http.Request) {
	// Read registry.json file
	file, err := os.ReadFile("registry.json")
	if err != nil {
		http.Error(w, "Error reading registry", http.StatusInternalServerError)
		return
	}

	var registry []map[string]interface{}
	if err := json.Unmarshal(file, &registry); err != nil {
		http.Error(w, "Error parsing registry", http.StatusInternalServerError)
		return
	}

	data := utils.PageData{
		Title:      "📜 Character Registry",
		CustomData: map[string]interface{}{"registry": registry},
	}

	utils.RenderTemplate(w, data, "registry.html", false)
}
