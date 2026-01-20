package api

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"sync"
	"time"

	"nostr-hero/db"
	"nostr-hero/utils"
)

// GameSession holds the in-memory game state for an active session
type GameSession struct {
	Npub      string   `json:"npub"`
	SaveID    string   `json:"save_id"`
	SaveData  SaveFile `json:"save_data"`
	LoadedAt  int64    `json:"loaded_at"`
	UpdatedAt int64    `json:"updated_at"`

	// Session-only data (not persisted to save files)
	BookedShows    []map[string]interface{} `json:"booked_shows,omitempty"`    // Current show bookings
	PerformedShows []string                 `json:"performed_shows,omitempty"` // Shows performed today (to prevent re-booking)
	RentedRooms    []map[string]interface{} `json:"rented_rooms,omitempty"`    // Current room rentals

	// Delta system: cached state for surgical updates
	LastSnapshot   *SessionSnapshot `json:"-"` // Previous state for delta calculation
	NPCsAtLocation []string         `json:"-"` // Cached NPCs at current location
	NPCsLastHour   int              `json:"-"` // Hour when NPCs were last fetched
	BuildingStates map[string]bool  `json:"-"` // Cached building open/close states
	BuildingsLastCheck int          `json:"-"` // Time when buildings were last checked
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
		return session, nil
	}

	// Load save file from disk
	saveData, err := LoadSaveByID(npub, saveID)
	if err != nil {
		return nil, fmt.Errorf("failed to load save file: %w", err)
	}

	// Initialize fatigue/hunger accumulation and penalty effects
	if err := initializeFatigueHungerEffects(saveData); err != nil {
		log.Printf("⚠️ Warning: Failed to initialize fatigue/hunger effects: %v", err)
	}

	// Create new session in memory
	session := &GameSession{
		Npub:           npub,
		SaveID:         saveID,
		SaveData:       *saveData,
		LoadedAt:       currentTimestamp(),
		UpdatedAt:      currentTimestamp(),
		BuildingStates: make(map[string]bool),
	}

	// Initialize building states and NPCs for current location
	database := db.GetDB()
	if database != nil {
		timeOfDay := saveData.TimeOfDay
		currentHour := timeOfDay / 60

		// Load initial building states
		buildingStates, err := utils.GetAllBuildingStatesForDistrict(
			database,
			saveData.Location,
			saveData.District,
			timeOfDay,
		)
		if err == nil && len(buildingStates) > 0 {
			session.BuildingStates = buildingStates
			session.BuildingsLastCheck = timeOfDay
		}

		// Load initial NPCs at location
		npcIDs := GetNPCIDsAtLocation(
			saveData.Location,
			saveData.District,
			saveData.Building,
			timeOfDay,
		)
		session.NPCsAtLocation = npcIDs
		session.NPCsLastHour = currentHour
	}

	// Initialize snapshot for delta system
	session.InitializeSnapshot()

	sm.sessions[key] = session

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

	// Initialize fatigue/hunger accumulation and penalty effects
	if err := initializeFatigueHungerEffects(saveData); err != nil {
		log.Printf("⚠️ Warning: Failed to initialize fatigue/hunger effects: %v", err)
	}

	// Create/overwrite session in memory
	session := &GameSession{
		Npub:           npub,
		SaveID:         saveID,
		SaveData:       *saveData,
		LoadedAt:       currentTimestamp(),
		UpdatedAt:      currentTimestamp(),
		BuildingStates: make(map[string]bool),
	}

	// Initialize building states and NPCs for current location
	database := db.GetDB()
	if database != nil {
		timeOfDay := saveData.TimeOfDay
		currentHour := timeOfDay / 60

		// Load initial building states
		buildingStates, err := utils.GetAllBuildingStatesForDistrict(
			database,
			saveData.Location,
			saveData.District,
			timeOfDay,
		)
		if err == nil && len(buildingStates) > 0 {
			session.BuildingStates = buildingStates
			session.BuildingsLastCheck = timeOfDay
		}

		// Load initial NPCs at location
		npcIDs := GetNPCIDsAtLocation(
			saveData.Location,
			saveData.District,
			saveData.Building,
			timeOfDay,
		)
		session.NPCsAtLocation = npcIDs
		session.NPCsLastHour = currentHour
	}

	// Initialize snapshot for delta system
	session.InitializeSnapshot()

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

	return nil // The actual write will happen in the handler
}

// UnloadSession removes a session from memory
func (sm *SessionManager) UnloadSession(npub, saveID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sessionKey(npub, saveID)
	delete(sm.sessions, key)
}

// GetAllSessions returns all active sessions (for debugging)
func (sm *SessionManager) GetAllSessions() map[string]*GameSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy to avoid race conditions
	sessionsCopy := make(map[string]*GameSession, len(sm.sessions))
	maps.Copy(sessionsCopy, sm.sessions)

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
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Session initialized successfully",
		"session": map[string]any{
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
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Session reloaded from disk successfully",
		"session": map[string]any{
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
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"session": session,
	})
}

// Helper function to get current timestamp
func currentTimestamp() int64 {
	return time.Now().Unix()
}

// UpdateSnapshotAndCalculateDelta updates the session's snapshot and returns the delta
func (s *GameSession) UpdateSnapshotAndCalculateDelta() *Delta {
	// Create new snapshot from current state
	newSnapshot := CreateSnapshot(&s.SaveData, s.NPCsAtLocation, s.BuildingStates)

	// Calculate show readiness based on booked shows and current time
	newSnapshot.ShowReady, newSnapshot.ShowReadyBuilding = s.calculateShowReadiness()

	// Log show_ready changes
	if s.LastSnapshot != nil && (s.LastSnapshot.ShowReady != newSnapshot.ShowReady || s.LastSnapshot.ShowReadyBuilding != newSnapshot.ShowReadyBuilding) {
		log.Printf("🎭 ShowReady changed: %v@%s -> %v@%s (time: %d, building: %s)",
			s.LastSnapshot.ShowReady, s.LastSnapshot.ShowReadyBuilding,
			newSnapshot.ShowReady, newSnapshot.ShowReadyBuilding,
			s.SaveData.TimeOfDay, s.SaveData.Building)
	}

	// Calculate delta from old to new
	var delta *Delta
	if s.LastSnapshot != nil {
		delta = CalculateDelta(s.LastSnapshot, newSnapshot)
	}

	// Update stored snapshot
	s.LastSnapshot = newSnapshot

	return delta
}

// calculateShowReadiness checks if a booked show is ready to perform
func (s *GameSession) calculateShowReadiness() (bool, string) {
	if s.BookedShows == nil || len(s.BookedShows) == 0 {
		return false, ""
	}

	currentTime := s.SaveData.TimeOfDay
	currentDay := s.SaveData.CurrentDay

	for _, booking := range s.BookedShows {
		// Skip already performed shows
		performed, _ := booking["performed"].(bool)
		if performed {
			continue
		}

		bookingDay := 0
		if day, ok := booking["day"].(float64); ok {
			bookingDay = int(day)
		} else if day, ok := booking["day"].(int); ok {
			bookingDay = day
		}

		showTime := 1260 // Default 9 PM
		if st, ok := booking["show_time"].(float64); ok {
			showTime = int(st)
		} else if st, ok := booking["show_time"].(int); ok {
			showTime = st
		}

		venueID := ""
		if vid, ok := booking["venue_id"].(string); ok {
			venueID = vid
		}

		// Check if within the 60-minute show window (same day, show_time to show_time+60)
		if bookingDay == currentDay {
			timeDiff := currentTime - showTime
			if timeDiff >= 0 && timeDiff <= 60 {
				return true, venueID
			}
		}
	}

	return false, ""
}

// InitializeSnapshot creates the initial snapshot (called when session is first loaded)
func (s *GameSession) InitializeSnapshot() {
	if s.BuildingStates == nil {
		s.BuildingStates = make(map[string]bool)
	}
	s.LastSnapshot = CreateSnapshot(&s.SaveData, s.NPCsAtLocation, s.BuildingStates)
	// Calculate initial show readiness
	s.LastSnapshot.ShowReady, s.LastSnapshot.ShowReadyBuilding = s.calculateShowReadiness()
}

// UpdateNPCsAtLocation updates the cached NPC list (called when hour changes)
func (s *GameSession) UpdateNPCsAtLocation(npcs []string, currentHour int) {
	s.NPCsAtLocation = npcs
	s.NPCsLastHour = currentHour
}

// UpdateBuildingStates updates the cached building states
func (s *GameSession) UpdateBuildingStates(buildings map[string]bool, currentTime int) {
	if s.BuildingStates == nil {
		s.BuildingStates = make(map[string]bool)
	}
	for id, isOpen := range buildings {
		s.BuildingStates[id] = isOpen
	}
	s.BuildingsLastCheck = currentTime
}

// ShouldRefreshNPCs returns true if NPCs should be refreshed (hour changed)
func (s *GameSession) ShouldRefreshNPCs(currentHour int) bool {
	return currentHour != s.NPCsLastHour || len(s.NPCsAtLocation) == 0
}

// ShouldRefreshBuildings returns true if building states should be refreshed
func (s *GameSession) ShouldRefreshBuildings(currentTime int) bool {
	// Refresh every 5 in-game minutes
	return s.BuildingsLastCheck == 0 || (currentTime-s.BuildingsLastCheck) >= 5
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
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Session unloaded from memory",
	})
}

