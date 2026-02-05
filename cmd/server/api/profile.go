package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/0ceanslim/grain/client/core/tools"

	"pubkey-quest/cmd/server/cache"
	"pubkey-quest/cmd/server/utils"
)

// ProfileMetadata represents a Nostr user profile
// swagger:model ProfileMetadata
type ProfileMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	About       string `json:"about"`
	Picture     string `json:"picture"`
	Banner      string `json:"banner"`
	Nip05       string `json:"nip05"`
	Lud16       string `json:"lud16"`
}

// ProfileResponse represents the response from the profile endpoint
// swagger:model ProfileResponse
type ProfileResponse struct {
	Npub    string          `json:"npub"`
	Pubkey  string          `json:"pubkey"`
	Profile ProfileMetadata `json:"profile"`
	Cached  bool            `json:"cached"`
}

// ProfileHandler godoc
// @Summary      Get Nostr profile
// @Description  Fetch Nostr profile metadata for a given npub
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Param        npub  query     string  true  "Nostr public key (npub format)"
// @Success      200   {object}  ProfileResponse
// @Failure      400   {string}  string  "Missing or invalid npub"
// @Failure      404   {string}  string  "Profile not found"
// @Failure      500   {string}  string  "Error fetching profile"
// @Router       /profile [get]
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	npub := r.URL.Query().Get("npub")
	if npub == "" {
		http.Error(w, "Missing npub parameter", http.StatusBadRequest)
		return
	}

	pubKey, err := tools.DecodeNpub(npub)
	if err != nil {
		http.Error(w, "Invalid npub", http.StatusBadRequest)
		return
	}

	// Check cache first
	if cachedProfile, found := cache.GlobalProfileCache.Get(pubKey); found {
		log.Printf("✅ Profile cache hit for %s", pubKey[:8]+"...")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"npub":    npub,
			"pubkey":  pubKey,
			"profile": ProfileMetadata{
				Name:        cachedProfile.Name,
				DisplayName: cachedProfile.DisplayName,
				About:       cachedProfile.About,
				Picture:     cachedProfile.Picture,
				Nip05:       cachedProfile.NIP05,
				Lud16:       cachedProfile.LUD16,
			},
			"cached": true,
		})
		return
	}

	log.Printf("⏳ Profile cache miss for %s, fetching from relays...", pubKey[:8]+"...")

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

	// Cache the profile data
	cache.GlobalProfileCache.Set(pubKey, cache.ProfileData{
		DisplayName: profile.DisplayName,
		Name:        profile.Name,
		Picture:     profile.Picture,
		About:       profile.About,
		NIP05:       profile.Nip05,
		LUD16:       profile.Lud16,
	})
	log.Printf("📝 Cached profile for %s (display_name: %s)", pubKey[:8]+"...", profile.DisplayName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"npub":    npub,
		"pubkey":  pubKey,
		"profile": profile,
		"cached":  false,
	})
}
