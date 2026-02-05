package api

import (
	"pubkey-quest/cmd/server/session"
	"pubkey-quest/types"
)

// Type aliases for backward compatibility - all delta types now live in session package

// Delta represents changes between states for surgical frontend updates.
type Delta = session.Delta

// ShowReadyDelta holds changes to show readiness state
type ShowReadyDelta = session.ShowReadyDelta

// CharacterDelta holds changes to character stats
type CharacterDelta = session.CharacterDelta

// NPCDelta holds changes to NPCs at current location
type NPCDelta = session.NPCDelta

// BuildingDelta holds changes to building states (open/closed)
type BuildingDelta = session.BuildingDelta

// InventoryDelta holds changes to inventory slots
type InventoryDelta = session.InventoryDelta

// InventorySlotDelta holds changes to a single inventory slot
type InventorySlotDelta = session.InventorySlotDelta

// EquipmentDelta holds changes to equipment slots
type EquipmentDelta = session.EquipmentDelta

// LocationDelta holds changes to location state
type LocationDelta = session.LocationDelta

// EffectsDelta holds changes to active effects
type EffectsDelta = session.EffectsDelta

// SessionSnapshot captures state at a point in time for delta calculation
type SessionSnapshot = session.SessionSnapshot

// InventorySlotSnapshot captures state of a single inventory slot
type InventorySlotSnapshot = session.InventorySlotSnapshot

// CreateSnapshot creates a snapshot from current session state
func CreateSnapshot(save *types.SaveFile, npcs []string, buildings map[string]bool) *SessionSnapshot {
	return session.CreateSnapshot(save, npcs, buildings)
}

// CalculateDelta compares two snapshots and returns only the changes
func CalculateDelta(old, new *SessionSnapshot) *Delta {
	return session.CalculateDelta(old, new)
}
