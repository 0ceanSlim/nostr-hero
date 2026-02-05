package status

import (
	"fmt"

	"pubkey-quest/cmd/server/game/effects"
	"pubkey-quest/types"
)

// InitializeFatigueHungerEffects ensures all accumulation and penalty effects are properly set
// This should be called when loading a save or after modifying fatigue/hunger values
func InitializeFatigueHungerEffects(state *types.SaveFile) error {
	// Migrate old effect IDs to new ones for backward compatibility
	effects.MigrateOldEffectIDs(state)

	// Ensure fatigue accumulation effect is present
	if err := EnsureFatigueAccumulation(state); err != nil {
		return fmt.Errorf("failed to ensure fatigue accumulation: %w", err)
	}

	// Ensure hunger accumulation effect is present
	if err := EnsureHungerAccumulation(state); err != nil {
		return fmt.Errorf("failed to ensure hunger accumulation: %w", err)
	}

	// Apply penalty effects based on current levels
	if _, err := UpdateFatiguePenaltyEffects(state); err != nil {
		return fmt.Errorf("failed to update fatigue penalty effects: %w", err)
	}

	if _, err := UpdateHungerPenaltyEffects(state); err != nil {
		return fmt.Errorf("failed to update hunger penalty effects: %w", err)
	}

	// Apply encumbrance effects based on current weight
	if _, err := UpdateEncumbrancePenaltyEffects(state); err != nil {
		return fmt.Errorf("failed to update encumbrance penalty effects: %w", err)
	}

	return nil
}
