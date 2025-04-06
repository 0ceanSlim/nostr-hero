package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// GetSessionHandler returns the current user's session data as JSON
func GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Get current session using the session manager
	session := GetCurrentUser(r)
	if session == nil {
		http.Error(w, "No active session found", http.StatusUnauthorized)
		log.Println("❌ GetSessionHandler: No active session found")
		return
	}

	// Create response data structure
	sessionData := map[string]interface{}{
		"publicKey": session.PublicKey,
		"lastActive": session.LastActive,
	}

	// Log session data for debugging
	log.Printf("🔍 Returning session data for user: %s", session.PublicKey)

	// Set JSON response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode session data as JSON and send response
	if err := json.NewEncoder(w).Encode(sessionData); err != nil {
		log.Printf("❌ Failed to encode session data: %v", err)
		http.Error(w, "Failed to retrieve session data", http.StatusInternalServerError)
		return
	}
}