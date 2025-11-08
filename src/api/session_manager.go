package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GameSession holds the in-memory game state for an active session
type GameSession struct {
	Npub      string   `json:"npub"`
	SaveID    string   `json:"save_id"`
	SaveData  SaveFile `json:"save_data"`
	LoadedAt  int64    `json:"loaded_at"`
	UpdatedAt int64    `json:"updated_at"`
}

// SessionManager manages all active game sessions in memory
type SessionManager struct {
	sessions map[string]*GameSession // Key: "{npub}:{saveID}"
	mu       sync.RWMutex
}

// Global session manager instance
var sessionManager = &SessionManager{
	sessions: make(map[string]*GameSession),
}

// Get the global session manager
func GetSessionManager() *SessionManager {
	return sessionManager
}

// Generate session key from npub and saveID
func sessionKey(npub, saveID string) string {
	return fmt.Sprintf("%s:%s", npub, saveID)
}

// LoadSession loads a save file into memory
func (sm *SessionManager) LoadSession(npub, saveID string) (*GameSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sessionKey(npub, saveID)

	// Check if already loaded
	if session, exists := sm.sessions[key]; exists {
		log.Printf("📍 Session already loaded: %s", key)
		return session, nil
	}

	// Load save file from disk
	saveData, err := LoadSaveByID(npub, saveID)
	if err != nil {
		return nil, fmt.Errorf("failed to load save file: %w", err)
	}

	// Create new session in memory
	session := &GameSession{
		Npub:      npub,
		SaveID:    saveID,
		SaveData:  *saveData,
		LoadedAt:  currentTimestamp(),
		UpdatedAt: currentTimestamp(),
	}

	sm.sessions[key] = session
	log.Printf("✅ Session loaded into memory: %s", key)

	return session, nil
}

// ReloadSession forces a reload from disk, discarding in-memory changes
func (sm *SessionManager) ReloadSession(npub, saveID string) (*GameSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sessionKey(npub, saveID)

	// Load save file from disk (even if session exists in memory)
	saveData, err := LoadSaveByID(npub, saveID)
	if err != nil {
		return nil, fmt.Errorf("failed to load save file: %w", err)
	}

	// Create/overwrite session in memory
	session := &GameSession{
		Npub:      npub,
		SaveID:    saveID,
		SaveData:  *saveData,
		LoadedAt:  currentTimestamp(),
		UpdatedAt: currentTimestamp(),
	}

	sm.sessions[key] = session
	log.Printf("🔄 Session reloaded from disk (discarded in-memory changes): %s", key)

	return session, nil
}

// GetSession retrieves an active session from memory
func (sm *SessionManager) GetSession(npub, saveID string) (*GameSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := sessionKey(npub, saveID)
	session, exists := sm.sessions[key]

	if !exists {
		return nil, fmt.Errorf("session not found in memory: %s", key)
	}

	return session, nil
}

// UpdateSession updates the in-memory game state
func (sm *SessionManager) UpdateSession(npub, saveID string, saveData SaveFile) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sessionKey(npub, saveID)
	session, exists := sm.sessions[key]

	if !exists {
		return fmt.Errorf("session not found in memory: %s", key)
	}

	// Update the save data
	session.SaveData = saveData
	session.UpdatedAt = currentTimestamp()

	log.Printf("✅ Session updated in memory: %s", key)
	return nil
}

// SaveSessionToDisk writes the in-memory state to disk
func (sm *SessionManager) SaveSessionToDisk(npub, saveID string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := sessionKey(npub, saveID)
	_, exists := sm.sessions[key]

	if !exists {
		return fmt.Errorf("session not found in memory: %s", key)
	}

	// Write to disk using existing save logic
	// (This will be called by the save endpoint)
	log.Printf("💾 Writing session to disk: %s", key)

	return nil // The actual write will happen in the handler
}

// UnloadSession removes a session from memory
func (sm *SessionManager) UnloadSession(npub, saveID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sessionKey(npub, saveID)
	delete(sm.sessions, key)

	log.Printf("🗑️ Session unloaded from memory: %s", key)
}

// GetAllSessions returns all active sessions (for debugging)
func (sm *SessionManager) GetAllSessions() map[string]*GameSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy to avoid race conditions
	sessionsCopy := make(map[string]*GameSession, len(sm.sessions))
	for key, session := range sm.sessions {
		sessionsCopy[key] = session
	}

	return sessionsCopy
}

// API Handlers

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
	session, err := sessionManager.LoadSession(request.Npub, request.SaveID)
	if err != nil {
		log.Printf("❌ Failed to initialize session: %v", err)
		http.Error(w, fmt.Sprintf("Failed to initialize session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Session initialized successfully",
		"session": map[string]interface{}{
			"npub":    session.Npub,
			"save_id": session.SaveID,
			"loaded_at": session.LoadedAt,
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
	session, err := sessionManager.ReloadSession(request.Npub, request.SaveID)
	if err != nil {
		log.Printf("❌ Failed to reload session: %v", err)
		http.Error(w, fmt.Sprintf("Failed to reload session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Session reloaded from disk successfully",
		"session": map[string]interface{}{
			"npub":       session.Npub,
			"save_id":    session.SaveID,
			"loaded_at":  session.LoadedAt,
			"updated_at": session.UpdatedAt,
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
	session, err := sessionManager.GetSession(npub, saveID)
	if err != nil {
		// If not in memory, try to load it
		session, err = sessionManager.LoadSession(npub, saveID)
		if err != nil {
			log.Printf("❌ Failed to get session: %v", err)
			http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.SaveData)
}

// UpdateSessionHandler updates the in-memory game state
// POST /api/session/update
func UpdateSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Npub     string                 `json:"npub"`
		SaveID   string                 `json:"save_id"`
		SaveData map[string]interface{} `json:"save_data"`
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

	var saveData SaveFile
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
	json.NewEncoder(w).Encode(map[string]interface{}{
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
	session, err := sessionManager.GetSession(request.Npub, request.SaveID)
	if err != nil {
		log.Printf("❌ Session not found in memory: %v", err)
		http.Error(w, "Session not found in memory", http.StatusNotFound)
		return
	}

	// Write to disk using existing save logic
	savePath := fmt.Sprintf("%s/%s/%s.json", SavesDirectory, request.Npub, request.SaveID)
	if err := writeSaveFile(savePath, &session.SaveData); err != nil {
		log.Printf("❌ Failed to write save file: %v", err)
		http.Error(w, "Failed to write save file", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Session saved to disk: %s", sessionKey(request.Npub, request.SaveID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
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
	json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"session_count": len(sessions),
			"sessions":      sessions,
		})
		return
	}

	// Get specific session
	session, err := sessionManager.GetSession(npub, saveID)
	if err != nil {
		// Try to load it if not in memory
		session, err = sessionManager.LoadSession(npub, saveID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Session not found: %v", err), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"session": session,
	})
}

// Helper function to get current timestamp
func currentTimestamp() int64 {
	return time.Now().Unix()
}

// CleanupHandler removes a session from memory
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Session unloaded from memory",
	})
}

// Extract npub from URL path
func extractNpubFromPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/api/session/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
