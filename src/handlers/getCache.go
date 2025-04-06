package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"nostr-hero/src/cache"
	"nostr-hero/src/utils"
)

// GetCacheHandler returns the cached user data as JSON
func GetCacheHandler(w http.ResponseWriter, r *http.Request) {
	// Get current session using the session manager
	session := GetCurrentUser(r)
	if session == nil {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		log.Println("❌ GetCacheHandler: No active session found")
		return
	}

	publicKey := session.PublicKey
	cachedData, found := cache.GetUserData(publicKey)
	if !found {
		http.Error(w, "No cached data found", http.StatusNotFound)
		log.Printf("❌ No cached data found for user: %s", session.PublicKey)
		return
	}

	// Parse metadata to extract useful fields (optional)
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(cachedData.Metadata), &metadata); err != nil {
		log.Printf("⚠️ Failed to parse metadata, returning raw data: %v", err)
	}

	npub, err := utils.EncodeNpub(publicKey)
	if err != nil {
		log.Printf("❌ Failed to encode npub: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create response structure
	response := map[string]interface{}{
		"publicKey": publicKey,
		"npub":      npub,
		"metadata":  cachedData.Metadata,
		"mailboxes": cachedData.Mailboxes,
		"cacheTime": cachedData.Timestamp,
	}

	// Log cache retrieval
	log.Printf("🔍 Returning cached data for user: %s", session.PublicKey)

	// Set JSON response headers
	w.Header().Set("Content-Type", "application/json")
	
	// Encode cached data as JSON and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Failed to encode cached data: %v", err)
		http.Error(w, "Failed to retrieve cached data", http.StatusInternalServerError)
		return
	}
}