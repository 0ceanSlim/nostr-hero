package staging

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ChangeType represents the type of change being made
type ChangeType string

const (
	ChangeCreate ChangeType = "create"
	ChangeUpdate ChangeType = "update"
	ChangeDelete ChangeType = "delete"
)

// Change represents a single file modification
type Change struct {
	Type       ChangeType `json:"type"`
	FilePath   string     `json:"file_path"`
	OldContent []byte     `json:"old_content,omitempty"`
	NewContent []byte     `json:"new_content,omitempty"`
	Timestamp  time.Time  `json:"timestamp"`
}

// Session represents a user's editing session
type Session struct {
	ID        string    `json:"id"`
	Npub      string    `json:"npub"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Changes   []Change  `json:"changes"`
}

// SessionManager manages all active sessions
type SessionManager struct {
	sessions          map[string]*Session
	lastSubmissions   map[string]time.Time // Key: sessionID or IP, Value: last submission time
	mu                sync.RWMutex
	submissionCooldown time.Duration
}

// Manager is the global session manager instance
var Manager = &SessionManager{
	sessions:          make(map[string]*Session),
	lastSubmissions:   make(map[string]time.Time),
	submissionCooldown: 12 * time.Hour,
}

// CreateSession creates a new session with a unique ID
func (sm *SessionManager) CreateSession(npub string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		ID:        uuid.New().String(),
		Npub:      npub,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Changes:   []Change{},
	}

	sm.sessions[session.ID] = session
	return session
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.sessions[id]
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, id)
}

// AddChange adds a change to a session, consolidating if the file already has changes
func (s *Session) AddChange(change Change) {
	// Find existing change for this file path
	for i, existing := range s.Changes {
		if existing.FilePath == change.FilePath {
			// Consolidate: preserve original OldContent, update NewContent
			if existing.Type == ChangeCreate {
				// If the original was a CREATE, keep it as CREATE with new content
				change.Type = ChangeCreate
				change.OldContent = nil
			} else if existing.Type == ChangeDelete && change.Type == ChangeCreate {
				// If we deleted then created, it's an UPDATE with original old content
				change.Type = ChangeUpdate
				change.OldContent = existing.OldContent
			} else {
				// Otherwise preserve the original old content
				change.OldContent = existing.OldContent
			}

			// Replace the existing change with consolidated one
			s.Changes[i] = change
			return
		}
	}

	// New file, append
	s.Changes = append(s.Changes, change)
}

// UpdateNpub updates the npub for a session
func (s *Session) UpdateNpub(npub string) {
	s.Npub = npub
}

// CleanupExpiredSessions removes expired sessions (run periodically)
func (sm *SessionManager) CleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
		}
	}
}

// CanSubmit checks if a session/IP is allowed to submit based on rate limiting
func (sm *SessionManager) CanSubmit(key string) (bool, time.Duration) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	lastSubmission, exists := sm.lastSubmissions[key]
	if !exists {
		return true, 0
	}

	timeSinceLastSubmission := time.Since(lastSubmission)
	if timeSinceLastSubmission < sm.submissionCooldown {
		remainingCooldown := sm.submissionCooldown - timeSinceLastSubmission
		return false, remainingCooldown
	}

	return true, 0
}

// RecordSubmission records a successful submission
func (sm *SessionManager) RecordSubmission(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.lastSubmissions[key] = time.Now()
}

// CleanupOldSubmissions removes submission records older than the cooldown period
func (sm *SessionManager) CleanupOldSubmissions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-sm.submissionCooldown)
	for key, timestamp := range sm.lastSubmissions {
		if timestamp.Before(cutoff) {
			delete(sm.lastSubmissions, key)
		}
	}
}

// StartCleanupRoutine starts a background goroutine to clean up expired sessions
func (sm *SessionManager) StartCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			sm.CleanupExpiredSessions()
			sm.CleanupOldSubmissions()
		}
	}()
}
