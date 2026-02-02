package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"nostr-hero/db"
	"nostr-hero/game/effects"
	"nostr-hero/game/status"
	"nostr-hero/session"
	"nostr-hero/types"
	"nostr-hero/utils"
)

// Type alias for backward compatibility
type GameSession = session.GameSession

// sessionManagerWrapper wraps session.SessionManager with api-specific dependencies
type sessionManagerWrapper struct {
	*session.SessionManager
}

// Global session manager instance
var sessionManager = &sessionManagerWrapper{
	SessionManager: session.NewSessionManager(),
}

// Get the global session manager
func GetSessionManager() *sessionManagerWrapper {
	return sessionManager
}

// LoadSession loads a save file into memory with api-specific initialization
func (sm *sessionManagerWrapper) LoadSession(npub, saveID string) (*GameSession, error) {
	return sm.SessionManager.LoadSession(
		npub,
		saveID,
		LoadSaveByID,
		status.InitializeFatigueHungerEffects,
		GetNPCIDsAtLocation,
		getBuildingStatesWrapper,
	)
}

// ReloadSession forces a reload from disk with api-specific initialization
func (sm *sessionManagerWrapper) ReloadSession(npub, saveID string) (*GameSession, error) {
	return sm.SessionManager.ReloadSession(
		npub,
		saveID,
		LoadSaveByID,
		status.InitializeFatigueHungerEffects,
		GetNPCIDsAtLocation,
		getBuildingStatesWrapper,
	)
}

// getBuildingStatesWrapper wraps utils.GetAllBuildingStatesForDistrict
func getBuildingStatesWrapper(location, district string, timeOfDay int) (map[string]bool, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}
	return utils.GetAllBuildingStatesForDistrict(database, location, district, timeOfDay)
}

// Helper to generate session key for logging
func sessionKey(npub, saveID string) string {
	return fmt.Sprintf("%s:%s", npub, saveID)
}

// ============================================================================
// HTTP HANDLERS
// ============================================================================

// InitSessionHandler initializes a game session from a save file
// POST /api/session/init
func InitSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Npub   string `json:"npub"`
		SaveID string `json:"save_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Npub == "" || request.SaveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	// Load session into memory
	sess, err := sessionManager.LoadSession(request.Npub, request.SaveID)
	if err != nil {
		log.Printf("❌ Failed to initialize session: %v", err)
		http.Error(w, fmt.Sprintf("Failed to initialize session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Session initialized successfully",
		"session": map[string]any{
			"npub":      sess.Npub,
			"save_id":   sess.SaveID,
			"loaded_at": sess.LoadedAt,
		},
	})
}

// ReloadSessionHandler forces a reload from disk, discarding in-memory changes
// POST /api/session/reload
func ReloadSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Npub   string `json:"npub"`
		SaveID string `json:"save_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Npub == "" || request.SaveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	// Force reload session from disk
	sess, err := sessionManager.ReloadSession(request.Npub, request.SaveID)
	if err != nil {
		log.Printf("❌ Failed to reload session: %v", err)
		http.Error(w, fmt.Sprintf("Failed to reload session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Session reloaded from disk successfully",
		"session": map[string]any{
			"npub":       sess.Npub,
			"save_id":    sess.SaveID,
			"loaded_at":  sess.LoadedAt,
			"updated_at": sess.UpdatedAt,
		},
	})
}

// GetSessionHandler retrieves the current in-memory session state
// GET /api/session/state?npub={npub}&save_id={saveID}
func GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	npub := r.URL.Query().Get("npub")
	saveID := r.URL.Query().Get("save_id")

	if npub == "" || saveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	// Get session from memory
	sess, err := sessionManager.GetSession(npub, saveID)
	if err != nil {
		// If not in memory, try to load it
		sess, err = sessionManager.LoadSession(npub, saveID)
		if err != nil {
			log.Printf("❌ Failed to get session: %v", err)
			http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
			return
		}
	}

	// Create response with enriched active effects
	response := map[string]interface{}{
		"d":                     sess.SaveData.D,
		"created_at":            sess.SaveData.CreatedAt,
		"race":                  sess.SaveData.Race,
		"class":                 sess.SaveData.Class,
		"background":            sess.SaveData.Background,
		"alignment":             sess.SaveData.Alignment,
		"experience":            sess.SaveData.Experience,
		"hp":                    sess.SaveData.HP,
		"max_hp":                sess.SaveData.MaxHP,
		"mana":                  sess.SaveData.Mana,
		"max_mana":              sess.SaveData.MaxMana,
		"fatigue":               sess.SaveData.Fatigue,
		"hunger":                sess.SaveData.Hunger,
		"stats":                 sess.SaveData.Stats,
		"location":              sess.SaveData.Location,
		"district":              sess.SaveData.District,
		"building":              sess.SaveData.Building,
		"current_day":           sess.SaveData.CurrentDay,
		"time_of_day":           sess.SaveData.TimeOfDay,
		"inventory":             sess.SaveData.Inventory,
		"vaults":                sess.SaveData.Vaults,
		"known_spells":          sess.SaveData.KnownSpells,
		"spell_slots":           sess.SaveData.SpellSlots,
		"locations_discovered":  sess.SaveData.LocationsDiscovered,
		"music_tracks_unlocked": sess.SaveData.MusicTracksUnlocked,
		"active_effects":        effects.EnrichActiveEffects(sess.SaveData.ActiveEffects),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateSessionHandler updates the in-memory game state
// POST /api/session/update
func UpdateSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Npub     string         `json:"npub"`
		SaveID   string         `json:"save_id"`
		SaveData map[string]any `json:"save_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Npub == "" || request.SaveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	// Convert map to SaveFile struct
	jsonData, err := json.Marshal(request.SaveData)
	if err != nil {
		http.Error(w, "Invalid save data", http.StatusBadRequest)
		return
	}

	var saveData types.SaveFile
	if err := json.Unmarshal(jsonData, &saveData); err != nil {
		http.Error(w, "Invalid save data format", http.StatusBadRequest)
		return
	}

	// Set internal metadata
	saveData.InternalNpub = request.Npub
	saveData.InternalID = request.SaveID

	// Update session in memory
	if err := sessionManager.UpdateSession(request.Npub, request.SaveID, saveData); err != nil {
		log.Printf("❌ Failed to update session: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Session updated successfully",
	})
}

// SaveSessionHandler writes the in-memory state to disk
// POST /api/session/save
func SaveSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Npub   string `json:"npub"`
		SaveID string `json:"save_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Npub == "" || request.SaveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	// Get session from memory
	sess, err := sessionManager.GetSession(request.Npub, request.SaveID)
	if err != nil {
		log.Printf("❌ Session not found in memory: %v", err)
		http.Error(w, "Session not found in memory", http.StatusNotFound)
		return
	}

	// Write to disk using existing save logic
	savePath := fmt.Sprintf("%s/%s/%s.json", SavesDirectory, request.Npub, request.SaveID)
	if err := writeSaveFile(savePath, &sess.SaveData); err != nil {
		log.Printf("❌ Failed to write save file: %v", err)
		http.Error(w, "Failed to write save file", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Session saved to disk: %s", sessionKey(request.Npub, request.SaveID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Game saved successfully",
		"save_id": request.SaveID,
	})
}

// DebugSessionsHandler returns all active sessions (debug only)
// GET /api/debug/sessions
func DebugSessionsHandler(w http.ResponseWriter, r *http.Request, debugMode bool) {
	if !debugMode {
		http.Error(w, "Debug mode disabled", http.StatusForbidden)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessions := sessionManager.GetAllSessions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":       true,
		"session_count": len(sessions),
		"sessions":      sessions,
	})
}

// DebugStateHandler returns the current game state for a specific session (debug only)
// GET /api/debug/state?npub={npub}&save_id={saveID}
func DebugStateHandler(w http.ResponseWriter, r *http.Request, debugMode bool) {
	if !debugMode {
		http.Error(w, "Debug mode disabled", http.StatusForbidden)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get npub and save_id from URL or from active sessions
	npub := r.URL.Query().Get("npub")
	saveID := r.URL.Query().Get("save_id")

	// If no parameters, return all sessions
	if npub == "" || saveID == "" {
		sessions := sessionManager.GetAllSessions()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success":       true,
			"session_count": len(sessions),
			"sessions":      sessions,
		})
		return
	}

	// Get specific session
	sess, err := sessionManager.GetSession(npub, saveID)
	if err != nil {
		// Try to load it if not in memory
		sess, err = sessionManager.LoadSession(npub, saveID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"session": sess,
	})
}

// CleanupSessionHandler removes a session from memory
// DELETE /api/session/cleanup?npub={npub}&save_id={saveID}
func CleanupSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	npub := r.URL.Query().Get("npub")
	saveID := r.URL.Query().Get("save_id")

	if npub == "" || saveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	sessionManager.UnloadSession(npub, saveID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Session unloaded from memory",
	})
}
