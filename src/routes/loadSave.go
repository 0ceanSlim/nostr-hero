package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nostr-hero/src/cache"
	"nostr-hero/src/handlers"
	"nostr-hero/src/utils"
)

func LoadSave(w http.ResponseWriter, r *http.Request) {
	// Get current user session
	session := handlers.GetCurrentUser(r)
	if session == nil {
		http.Error(w, "No active session found", http.StatusUnauthorized)
		log.Println("❌ LoadSave: No active session found")
		return
	}

	// Get user data from cache using public key
	userData, exists := cache.GetUserData(session.PublicKey)
	if !exists {
		http.Error(w, "User data not found in cache", http.StatusNotFound)
		log.Printf("❌ LoadSave: User data not found for %s", session.PublicKey)
		return
	}

	// Parse metadata
	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(userData.Metadata), &metadata)
	if err != nil {
		http.Error(w, "Invalid metadata format", http.StatusInternalServerError)
		log.Printf("❌ LoadSave: Failed to parse metadata: %v", err)
		return
	}

	// Parse content field from metadata
	contentJSON, ok := metadata["content"].(string)
	if !ok {
		http.Error(w, "Content not found in metadata", http.StatusInternalServerError)
		log.Printf("❌ LoadSave: Content field not found in metadata")
		return
	}

	var content map[string]interface{}
	err = json.Unmarshal([]byte(contentJSON), &content)
	if err != nil {
		http.Error(w, "Invalid content format", http.StatusInternalServerError)
		log.Printf("❌ LoadSave: Failed to parse content: %v", err)
		return
	}

	// Generate npub from public key using the same utility function as GetCacheHandler
    npub, err := utils.EncodeNpub(session.PublicKey)
    if err != nil {
        http.Error(w, "Failed to encode npub", http.StatusInternalServerError)
        log.Printf("❌ LoadSave: Failed to encode npub: %v", err)
        return
    }

	// Fetch character data using the public key
	characterData, err := fetchCharacterData(npub)
	if err != nil {
		log.Printf("⚠️ LoadSave: Failed to fetch character data: %v", err)
		// Continue with empty character data if error occurs
		characterData = map[string]interface{}{}
	}

	// Prepare placeholder save data - this would be replaced with actual save data in future
	saveData := []map[string]interface{}{
		{
			"id":        "save1",
			"name":      "Adventure in the Dark Forest",
			"timestamp": "2025-04-01 14:30:22",
			"level":     5,
		},
		{
			"id":        "save2",
			"name":      "The Mountain Quest",
			"timestamp": "2025-04-05 18:12:45",
			"level":     8,
		},
	}

	// Create custom data for template
	customData := map[string]interface{}{
		"profile":   content,
		"character": characterData,
		"saves":     saveData,
		"npub":    npub,
	}

	// Render the template
	data := utils.PageData{
		Title:      "Load Game",
		Theme:      "dark", // Or get from user preferences
		CustomData: customData,
	}

	utils.RenderTemplate(w, data, "load-save.html", false)
}

// Helper function to fetch character data from API
func fetchCharacterData(npub string) (map[string]interface{}, error) {

	// Create the full URL with protocol, host and port
	port := utils.AppConfig.Server.Port
	apiURL := fmt.Sprintf("http://localhost:%d/api/character?npub=%s", port, npub)
	
	log.Printf("🔍 Fetching character data from: %s", apiURL)
	
	// Make API call to character endpoint
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to character API: %w", err)
	}
	defer resp.Body.Close()
	
	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	var characterData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&characterData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode character data: %w", err)
	}

	log.Printf("✅ Successfully retrieved character data for npub: %s", npub)
	return characterData, nil
}