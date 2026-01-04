package staging

import (
	"net/http"
	"strings"

	"nostr-hero/codex/config"
)

// Mode represents the operating mode of CODEX
type Mode string

const (
	ModeDirect  Mode = "direct"  // Write to disk immediately
	ModeStaging Mode = "staging" // Stage changes for PR
)

// DetectMode determines whether to use direct save or staging based on request context
func DetectMode(r *http.Request, cfg *config.Config) Mode {
	// If config explicitly sets a mode, use it
	if cfg.Server.StagingMode == "direct" {
		return ModeDirect
	}
	if cfg.Server.StagingMode == "staging" {
		return ModeStaging
	}

	// Auto-detect based on hostname (default behavior)
	host := r.Host
	if strings.HasPrefix(host, "localhost:") || strings.HasPrefix(host, "127.0.0.1:") || host == "localhost" || host == "127.0.0.1" {
		return ModeDirect
	}

	return ModeStaging
}
