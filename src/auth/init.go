package auth

import (
	"log"

	"nostr-hero/src/utils"
)

// InitializeGrainClient initializes only the session manager for Nostr Hero
func InitializeGrainClient(config *utils.Config) error {
	log.Println("🎮 Initializing Grain session manager for Nostr Hero...")

	// For now, we only need the session manager from grain
	// We can initialize full client later if needed for relay connections
	log.Println("✅ Grain session manager ready for Nostr Hero")
	return nil
}

// ShutdownGrainClient gracefully shuts down the grain client
func ShutdownGrainClient() error {
	log.Println("🎮 Shutting down Grain client...")
	return nil
}

// getDefaultRelays returns default Nostr relays for the game
func getDefaultRelays(config *utils.Config) []string {
	// You can configure these in your config file later
	defaultRelays := []string{
		"wss://relay.damus.io",
		"wss://nos.lol",
		"wss://relay.nostr.band",
		"wss://nostr.happytavern.co",
		"wss://relay.snort.social",
	}

	// TODO: Add config option to override default relays
	// if config.Nostr.DefaultRelays != nil {
	//     return config.Nostr.DefaultRelays
	// }

	return defaultRelays
}