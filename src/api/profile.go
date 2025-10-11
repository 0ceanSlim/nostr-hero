package api

import (
	"encoding/json"
	"net/http"

	"nostr-hero/src/utils"
)

// ProfileMetadata represents a Nostr user profile
type ProfileMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	About       string `json:"about"`
	Picture     string `json:"picture"`
	Banner      string `json:"banner"`
	Nip05       string `json:"nip05"`
	Lud16       string `json:"lud16"`
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	npub := r.URL.Query().Get("npub")
	if npub == "" {
		http.Error(w, "Missing npub parameter", http.StatusBadRequest)
		return
	}

	pubKey, err := utils.DecodeNpub(npub)
	if err != nil {
		http.Error(w, "Invalid npub", http.StatusBadRequest)
		return
	}

	// Default relays to query
	relays := []string{
		"wss://relay.damus.io",
		"wss://nos.lol",
		"wss://relay.nostr.band",
		"wss://nostr.wine",
	}

	// Fetch metadata from relays
	event, err := utils.FetchUserMetadata(pubKey, relays)
	if err != nil {
		http.Error(w, "Error fetching profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if event == nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Parse the content JSON
	var profile ProfileMetadata
	if err := json.Unmarshal([]byte(event.Content), &profile); err != nil {
		http.Error(w, "Error parsing profile data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"npub":    npub,
		"pubkey":  pubKey,
		"profile": profile,
	})
}
