package npc

import (
	"slices"

	"nostr-hero/game/gameutil"
	"nostr-hero/game/vault"
	"nostr-hero/types"
)

// IsNativeRaceForLocation checks if a race is native to a location
func IsNativeRaceForLocation(race, location string) bool {
	nativeRaces := map[string][]string{
		"kingdom":           {"Human", "Half-Elf", "Half-Orc", "Tiefling"},
		"village-southwest": {"Orc"},
		"forest-kingdom":    {"Elf"},
		"hill-kingdom":      {"Dwarf"},
		"village-west":      {"Halfling"},
	}

	races, ok := nativeRaces[location]
	if !ok {
		return false
	}

	return slices.Contains(races, race)
}

// CheckDialogueRequirements checks if dialogue requirements are met
func CheckDialogueRequirements(state *types.SaveFile, requirements map[string]interface{}) bool {
	if requirements == nil {
		return true
	}

	if notNative, ok := requirements["not_native"].(bool); ok && notNative {
		if IsNativeRaceForLocation(state.Race, state.Location) {
			return false
		}
	}

	if notRegistered, ok := requirements["not_registered"].(bool); ok && notRegistered {
		if vault.IsVaultRegistered(state, state.Building) {
			return false
		}
	}

	if registered, ok := requirements["registered"].(bool); ok && registered {
		if !vault.IsVaultRegistered(state, state.Building) {
			return false
		}
	}

	if goldReq, ok := requirements["gold"].(float64); ok {
		if gameutil.GetGoldQuantity(state) < int(goldReq) {
			return false
		}
	}

	return true
}
