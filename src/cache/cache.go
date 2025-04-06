package cache

import (
	"sync"
	"time"
)

type UserCache struct {
	mu     sync.RWMutex
	data   map[string]CachedUserData
	expiry time.Duration
}

type CachedUserData struct {
	Metadata  string    // JSON serialized metadata
	Mailboxes    string    // JSON serialized relays
	Timestamp time.Time // Store time of insertion
}

var cache = &UserCache{
	data:   make(map[string]CachedUserData),
	expiry: 10 * time.Minute, // Auto-expire in 10 min
}

// Store user data
func SetUserData(publicKey string, metadata, mailboxes string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.data[publicKey] = CachedUserData{
		Metadata:  metadata,
		Mailboxes:    mailboxes,
		Timestamp: time.Now(),
	}
}

// Retrieve user data
func GetUserData(publicKey string) (CachedUserData, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	data, exists := cache.data[publicKey]
	if !exists || time.Since(data.Timestamp) > cache.expiry {
		return CachedUserData{}, false
	}
	return data, true
}
