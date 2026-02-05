package staging

import (
	"encoding/json"
	"log"
	"net/http"

	"pubkey-quest/cmd/codex/config"
)

var cfg *config.Config

// SetConfig sets the configuration for handlers
func SetConfig(c *config.Config) {
	cfg = c
}

// HandleStagingInit creates a new staging session
// POST /api/staging/init
// Request: { "npub": "string" }
// Response: { "session_id": "uuid", "expires_at": "timestamp" }
func HandleStagingInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Npub string `json:"npub"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Npub == "" {
		req.Npub = "anonymous"
	}

	session := Manager.CreateSession(req.Npub)
	log.Printf("📝 Created staging session: %s (npub: %s)", session.ID, session.Npub)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"expires_at": session.ExpiresAt,
	})
}

// HandleGetStagingChanges retrieves the list of staged changes
// GET /api/staging/changes?session_id=uuid
// Response: { "changes": [...], "npub": "string", "change_count": 0 }
func HandleGetStagingChanges(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session := Manager.GetSession(sessionID)
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"changes":      session.Changes,
		"npub":         session.Npub,
		"change_count": len(session.Changes),
	})
}

// HandleStagingSubmit creates a GitHub PR from staged changes
// POST /api/staging/submit
// Request: { "session_id": "uuid", "npub": "string" (optional) }
// Response: { "pr_url": "https://github.com/..." }
func HandleStagingSubmit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Npub      string `json:"npub"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	session := Manager.GetSession(req.SessionID)
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if len(session.Changes) == 0 {
		http.Error(w, "No changes to submit", http.StatusBadRequest)
		return
	}

	// Update npub if provided
	if req.Npub != "" {
		session.UpdateNpub(req.Npub)
	}

	// Check rate limit (12-hour cooldown)
	canSubmit, remainingCooldown := Manager.CanSubmit(req.SessionID)
	if !canSubmit {
		hours := int(remainingCooldown.Hours())
		minutes := int(remainingCooldown.Minutes()) % 60
		log.Printf("⚠️ Rate limit exceeded for session %s (cooldown: %dh %dm remaining)", req.SessionID, hours, minutes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":              "Rate limit exceeded",
			"remaining_cooldown": remainingCooldown.String(),
			"retry_after_hours":  hours,
			"retry_after_minutes": minutes,
		})
		return
	}

	// Create GitHub client
	ghClient := NewGitHubClient(cfg)
	if ghClient == nil {
		http.Error(w, "GitHub not configured - check codex-config.yml", http.StatusServiceUnavailable)
		return
	}

	// Create PR
	log.Printf("🚀 Submitting %d changes to GitHub...", len(session.Changes))
	prURL, err := ghClient.CreatePR(session)
	if err != nil {
		log.Printf("❌ Failed to create PR: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Record successful submission for rate limiting
	Manager.RecordSubmission(req.SessionID)

	// Clean up session after successful submission
	Manager.DeleteSession(req.SessionID)
	log.Printf("✅ PR created and session cleaned up: %s", prURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pr_url": prURL,
	})
}

// HandleStagingClear clears all staged changes for a session
// DELETE /api/staging/clear?session_id=uuid
// Response: { "status": "success" }
func HandleStagingClear(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session := Manager.GetSession(sessionID)
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Clear changes by creating a new session with same ID and npub
	npub := session.Npub
	Manager.DeleteSession(sessionID)

	// Create fresh session
	newSession := Manager.CreateSession(npub)
	// Swap IDs to preserve the session ID for the client
	Manager.DeleteSession(newSession.ID)
	newSession.ID = sessionID
	Manager.sessions[sessionID] = newSession

	log.Printf("🗑️ Cleared staging session: %s", sessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// HandleGetMode detects and returns the current operating mode
// GET /api/staging/mode
// Response: { "mode": "direct"|"staging", "github_configured": bool }
func HandleGetMode(w http.ResponseWriter, r *http.Request) {
	mode := DetectMode(r, cfg)
	githubConfigured := cfg.GitHub.Token != ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mode":              string(mode),
		"github_configured": githubConfigured,
	})
}
