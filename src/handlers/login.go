package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"nostr-hero/src/cache"
	"nostr-hero/src/types"
	"nostr-hero/src/utils"
)

// Create a global session manager instance
var SessionMgr = NewSessionManager()

// InitUser handles user login and initialization
func InitUser(w http.ResponseWriter, r *http.Request) {
	log.Println("🔑 InitUser called")

	// Check if user is already logged in
	token := SessionMgr.GetSessionToken(r)
	if token != "" {
		session := SessionMgr.GetUserSession(token)
		if session != nil {
			log.Printf("User already logged in with key:%s", session.PublicKey)
			return
		}
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Failed to parse form: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	publicKey := r.FormValue("publicKey")
	if publicKey == "" {
		log.Println("❌ Missing publicKey in form data")
		http.Error(w, "Missing publicKey", http.StatusBadRequest)
		return
	}
	log.Printf("✅ Received publicKey: %s", publicKey)

	// Try to get user data from cache first
	cachedUserData, exists := cache.GetUserData(publicKey)
	if exists {
		// Parse cached data
		var userMetadata types.NostrEvent
		if err := json.Unmarshal([]byte(cachedUserData.Metadata), &userMetadata); err != nil {
			log.Printf("⚠️ Failed to parse cached metadata: %v", err)
		} else {
			var mailboxes utils.Mailboxes
			if err := json.Unmarshal([]byte(cachedUserData.Mailboxes), &mailboxes); err != nil {
				log.Printf("⚠️ Failed to parse cached mailboxes: %v", err)
			} else {
				// Create session with cached data
				session, err := SessionMgr.CreateSession(w, publicKey)
				if err == nil {
					log.Printf("🔄 Using cached data of public key:%s", session.PublicKey)
					return
				}
			}
		}
	}

	// Fall back to fetching data from relays
	appRelays := []string{
		"wss://purplepag.es", "wss://relay.damus.io", "wss://nos.lol",
		"wss://relay.primal.net", "wss://relay.nostr.band", "wss://offchain.pub",
	}

	mailboxes, err := utils.FetchUserRelays(publicKey, appRelays)
	if err != nil {
		log.Printf("❌ Failed to fetch user relays: %v", err)
		http.Error(w, "Failed to fetch user relays", http.StatusInternalServerError)
		return
	}

	allRelays := append(mailboxes.Read, mailboxes.Write...)
	allRelays = append(allRelays, mailboxes.Both...)
	log.Printf("✅ Fetched user relays: %+v", mailboxes)

	userMetadata, err := utils.FetchUserMetadata(publicKey, allRelays)
	if err != nil || userMetadata == nil {
		log.Printf("❌ Failed to fetch user metadata: %v", err)
		http.Error(w, "Failed to fetch user metadata", http.StatusInternalServerError)
		return
	}

	// Cache the user data
	relaysJSON, _ := json.Marshal(mailboxes)
	eventJSON, _ := json.Marshal(userMetadata)
	cache.SetUserData(publicKey, string(eventJSON), string(relaysJSON))
	log.Println("✅ User data cached successfully")

	// Create new session
	session, err := SessionMgr.CreateSession(w, publicKey)
	if err != nil {
		log.Printf("❌ Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	log.Printf("public key:%s is logged in", session.PublicKey)
}

// GetCurrentUser retrieves the current user from the session
func GetCurrentUser(r *http.Request) *UserSession {
	token := SessionMgr.GetSessionToken(r)
	if token == "" {
		return nil
	}
	return SessionMgr.GetUserSession(token)
}

// LogoutUser handles user logout
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	SessionMgr.ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}