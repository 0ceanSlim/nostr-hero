package status

import (
	"log"

	"nostr-hero/game/effects"
	"nostr-hero/types"
)

// UpdateFatiguePenaltyEffects applies appropriate penalty effects based on fatigue level
// New thresholds: 0-5 (no penalty), 6 (tired), 8 (very tired), 9 (fatigued), 10 (exhaustion)
func UpdateFatiguePenaltyEffects(state *types.SaveFile) (*types.EffectMessage, error) {
	// Remove all existing fatigue penalty effects
	RemoveFatiguePenaltyEffects(state)

	// Apply penalty effect based on current fatigue level
	switch {
	case state.Fatigue >= 10:
		return effects.ApplyEffectWithMessage(state, "exhaustion")
	case state.Fatigue == 9:
		return effects.ApplyEffectWithMessage(state, "fatigued")
	case state.Fatigue == 8:
		return effects.ApplyEffectWithMessage(state, "very-tired")
	case state.Fatigue >= 6:
		return effects.ApplyEffectWithMessage(state, "tired")
	default:
		// Fatigue 0-5: No fatigue penalty (fresh)
		return nil, nil
	}
}

// RemoveFatiguePenaltyEffects removes all fatigue penalty effects
func RemoveFatiguePenaltyEffects(state *types.SaveFile) {
	var remainingEffects []types.ActiveEffect
	for _, activeEffect := range state.ActiveEffects {
		// Keep non-fatigue-penalty effects
		if activeEffect.EffectID != "tired" &&
			activeEffect.EffectID != "very-tired" &&
			activeEffect.EffectID != "fatigued" &&
			activeEffect.EffectID != "exhaustion" {
			remainingEffects = append(remainingEffects, activeEffect)
		}
	}
	state.ActiveEffects = remainingEffects
}

// EnsureFatigueAccumulation ensures the fatigue accumulation effect is active
// Only adds if fatigue < 10 (stops accumulation at max)
func EnsureFatigueAccumulation(state *types.SaveFile) error {
	// Don't accumulate if already at max fatigue
	if state.Fatigue >= 10 {
		RemoveFatigueAccumulation(state)
		return nil
	}

	// Check if already present
	for _, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID == "fatigue-accumulation" {
			return nil // Already present
		}
	}

	// Apply it
	return effects.ApplyEffect(state, "fatigue-accumulation")
}

// RemoveFatigueAccumulation removes the fatigue accumulation effect
func RemoveFatigueAccumulation(state *types.SaveFile) {
	var remainingEffects []types.ActiveEffect
	for _, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID != "fatigue-accumulation" {
			remainingEffects = append(remainingEffects, activeEffect)
		}
	}
	state.ActiveEffects = remainingEffects
}

// ResetFatigueAccumulator resets the tick accumulator for fatigue accumulation effect
func ResetFatigueAccumulator(state *types.SaveFile) {
	for i, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID == "fatigue-accumulation" {
			state.ActiveEffects[i].TickAccumulator = 0
			return
		}
	}
}

// HandleFatigueChange processes a fatigue change and updates related effects
func HandleFatigueChange(state *types.SaveFile) {
	// Stop accumulation if we've reached max fatigue
	if state.Fatigue >= 10 {
		RemoveFatigueAccumulation(state)
	} else {
		// Ensure accumulation is active if below max
		if err := EnsureFatigueAccumulation(state); err != nil {
			log.Printf("⚠️ Failed to ensure fatigue accumulation: %v", err)
		}
	}

	// Update penalty effects
	if _, err := UpdateFatiguePenaltyEffects(state); err != nil {
		log.Printf("⚠️ Failed to update fatigue penalty effects: %v", err)
	}
}
