package api

import (
	"pubkey-quest/cmd/server/session"
)

// Type aliases for backward compatibility
type GameSession = session.GameSession

// GetSessionManager returns the global session manager
// Re-exported from session package for backward compatibility
func GetSessionManager() *session.SessionManagerWrapper {
	return session.GetSessionManager()
}
