package api

import (
	"fmt"

	"nostr-hero/api/data"
	"nostr-hero/db"
	"nostr-hero/game/status"
	"nostr-hero/session"
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

// GetSessionManager returns the global session manager
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
		data.GetNPCIDsAtLocation,
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
		data.GetNPCIDsAtLocation,
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
