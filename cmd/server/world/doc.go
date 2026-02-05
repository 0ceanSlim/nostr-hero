// Package world manages transient server-side world state that persists across
// player sessions but is NOT saved to player save files.
//
// This package handles:
//   - Merchant inventories and gold (per-player, resets on restock timers)
//   - Ground items / dropped items (future)
//   - World events and temporary state (future)
//   - Any game state that should be server-authoritative and session-scoped
//
// Key principles:
//   - State lives in server memory only (not in save files)
//   - State is per-player where appropriate (e.g., merchant stock)
//   - State has built-in expiration/reset mechanisms
//   - Players cannot cheat by manipulating save files
//
// Current files:
//   - merchant.go: MerchantStateManager for per-player merchant inventories/gold
//
// Future additions:
//   - ground.go: Dropped items per location
//   - events.go: Temporary world events
//   - manager.go: Unified world state manager
package world
