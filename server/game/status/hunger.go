package status

import (
	"log"

	"nostr-hero/game/effects"
	"nostr-hero/types"
)

// UpdateHungerPenaltyEffects applies appropriate penalty effects based on hunger level
// 3/3 "Stuffed": +1 CON, -1 STR, -1 DEX
// 2/3 "Well Fed": No effect (baseline)
// 1/3 "Hungry": -1 DEX only
// 0/3 "Famished": -1 HP every 4 hours
func UpdateHungerPenaltyEffects(state *types.SaveFile) (*types.EffectMessage, error) {
	// Remove all existing hunger penalty effects
	RemoveHungerPenaltyEffects(state)

	// Apply penalty/bonus effect based on current hunger level
	switch state.Hunger {
	case 0:
		return effects.ApplyEffectWithMessage(state, "starving")
	case 1:
		return effects.ApplyEffectWithMessage(state, "hungry")
	case 2:
		// Well fed - no effect (baseline)
		return nil, nil
	case 3:
		return effects.ApplyEffectWithMessage(state, "stuffed")
	default:
		// Clamp to valid range
		if state.Hunger < 0 {
			state.Hunger = 0
			return effects.ApplyEffectWithMessage(state, "starving")
		}
		state.Hunger = 3
		return effects.ApplyEffectWithMessage(state, "stuffed")
	}
}

// RemoveHungerPenaltyEffects removes all hunger penalty effects
func RemoveHungerPenaltyEffects(state *types.SaveFile) {
	var remainingEffects []types.ActiveEffect
	for _, activeEffect := range state.ActiveEffects {
		// Keep non-hunger-penalty effects
		if activeEffect.EffectID != "starving" &&
			activeEffect.EffectID != "hungry" &&
			activeEffect.EffectID != "stuffed" {
			remainingEffects = append(remainingEffects, activeEffect)
		}
	}
	state.ActiveEffects = remainingEffects
}

// EnsureHungerAccumulation ensures hunger accumulation effect is present (no swapping needed)
func EnsureHungerAccumulation(state *types.SaveFile) error {
	// Check if hunger accumulation effect already exists
	for _, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID == "hunger-accumulation-stuffed" ||
			activeEffect.EffectID == "hunger-accumulation-wellfed" ||
			activeEffect.EffectID == "hunger-accumulation-hungry" {
			// Already present - don't remove/re-add (preserves tick_accumulator)
			return nil
		}
	}

	// Apply initial hunger accumulation effect based on current hunger level
	var effectID string
	switch state.Hunger {
	case 3:
		effectID = "hunger-accumulation-stuffed"
	case 2:
		effectID = "hunger-accumulation-wellfed"
	case 1:
		effectID = "hunger-accumulation-hungry"
	case 0:
		// Don't apply hunger decrease accumulation when famished (hunger stays at 0)
		return nil
	default:
		return nil
	}

	return effects.ApplyEffect(state, effectID)
}

// RemoveHungerAccumulation removes all hunger accumulation effects
func RemoveHungerAccumulation(state *types.SaveFile) {
	var remainingEffects []types.ActiveEffect
	for _, activeEffect := range state.ActiveEffects {
		// Keep non-hunger-accumulation effects
		if activeEffect.EffectID != "hunger-accumulation-stuffed" &&
			activeEffect.EffectID != "hunger-accumulation-wellfed" &&
			activeEffect.EffectID != "hunger-accumulation-hungry" {
			remainingEffects = append(remainingEffects, activeEffect)
		}
	}
	state.ActiveEffects = remainingEffects
}

// ResetHungerAccumulator resets the tick accumulator for hunger accumulation effects
func ResetHungerAccumulator(state *types.SaveFile) {
	for i, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID == "hunger-accumulation-stuffed" ||
			activeEffect.EffectID == "hunger-accumulation-wellfed" ||
			activeEffect.EffectID == "hunger-accumulation-hungry" {
			state.ActiveEffects[i].TickAccumulator = 0
			return
		}
	}
}

// HandleHungerChange processes a hunger change and updates related effects
func HandleHungerChange(state *types.SaveFile) {
	// Update penalty effects
	if _, err := UpdateHungerPenaltyEffects(state); err != nil {
		log.Printf("⚠️ Failed to update hunger penalty effects: %v", err)
	}

	// Ensure accumulation is active
	if err := EnsureHungerAccumulation(state); err != nil {
		log.Printf("⚠️ Failed to ensure hunger accumulation: %v", err)
	}
}
