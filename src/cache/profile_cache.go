package cache

import (
	"sync"
	"time"
)

// ProfileData represents cached Nostr profile metadata
type ProfileData struct {
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
	Picture     string `json:"picture"`
	About       string `json:"about"`
	NIP05       string `json:"nip05"`
	LUD16       string `json:"lud16"`
}

// CachedProfile stores profile data with expiration
type CachedProfile struct {
	Data      ProfileData
	ExpiresAt time.Time
}

// ProfileCache is a thread-safe in-memory cache for Nostr profiles
type ProfileCache struct {
	mu       sync.RWMutex
	profiles map[string]CachedProfile // map[pubkey]CachedProfile
	ttl      time.Duration
}

var (
	// GlobalProfileCache is the global profile cache instance
	GlobalProfileCache *ProfileCache
)

// InitProfileCache initializes the global profile cache
func InitProfileCache(ttl time.Duration) {
	GlobalProfileCache = &ProfileCache{
		profiles: make(map[string]CachedProfile),
		ttl:      ttl,
	}

	// Start cleanup goroutine
	go GlobalProfileCache.cleanupExpired()
}

// Get retrieves a profile from cache if it exists and is not expired
func (pc *ProfileCache) Get(pubkey string) (ProfileData, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	cached, exists := pc.profiles[pubkey]
	if !exists {
		return ProfileData{}, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return ProfileData{}, false
	}

	return cached.Data, true
}

// Set stores a profile in cache with TTL
func (pc *ProfileCache) Set(pubkey string, profile ProfileData) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.profiles[pubkey] = CachedProfile{
		Data:      profile,
		ExpiresAt: time.Now().Add(pc.ttl),
	}
}

// Delete removes a profile from cache
func (pc *ProfileCache) Delete(pubkey string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	delete(pc.profiles, pubkey)
}

// Clear removes all profiles from cache
func (pc *ProfileCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.profiles = make(map[string]CachedProfile)
}

// cleanupExpired periodically removes expired entries
func (pc *ProfileCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pc.mu.Lock()
		now := time.Now()
		for pubkey, cached := range pc.profiles {
			if now.After(cached.ExpiresAt) {
				delete(pc.profiles, pubkey)
			}
		}
		pc.mu.Unlock()
	}
}

// Stats returns cache statistics
func (pc *ProfileCache) Stats() map[string]interface{} {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return map[string]interface{}{
		"total_cached": len(pc.profiles),
		"ttl_hours":    pc.ttl.Hours(),
	}
}
