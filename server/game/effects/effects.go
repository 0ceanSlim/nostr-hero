package effects

import (
	"encoding/json"
	"fmt"
	"log"

	"nostr-hero/db"
	"nostr-hero/types"
)

// ApplyEffect applies an effect to the character (from game-data/effects/{effectID}.json)
func ApplyEffect(state *types.SaveFile, effectID string) error {
	_, err := ApplyEffectWithMessage(state, effectID)
	return err
}

// ApplyEffectWithMessage applies an effect and returns a message to display
func ApplyEffectWithMessage(state *types.SaveFile, effectID string) (*types.EffectMessage, error) {
	// Load effect data from file
	effectData, err := LoadEffectData(effectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load effect %s: %v", effectID, err)
	}

	// Initialize active_effects if nil
	if state.ActiveEffects == nil {
		state.ActiveEffects = []types.ActiveEffect{}
	}

	// Get effect details
	effects, _ := effectData["effects"].([]interface{})
	_, _ = effectData["name"].(string) // name unused but kept for documentation
	message, _ := effectData["message"].(string)
	color, _ := effectData["color"].(string)
	category, _ := effectData["category"].(string)
	silent, _ := effectData["silent"].(bool)

	if effects == nil {
		return nil, fmt.Errorf("effect %s has no effects array", effectID)
	}

	// Apply each effect component
	for idx, effectRaw := range effects {
		effect, ok := effectRaw.(map[string]interface{})
		if !ok {
			continue
		}

		effectType, _ := effect["type"].(string)
		value, _ := effect["value"].(float64)
		duration, _ := effect["duration"].(float64)
		delay, _ := effect["delay"].(float64)
		tickInterval, _ := effect["tick_interval"].(float64)

		// Determine if this should be an immediate effect or active effect
		// Immediate effects: instant hp/mana/fatigue/hunger changes with no duration/delay/tick
		// Active effects: everything else (stat modifiers, over-time effects, delayed effects)
		isStatModifier := effectType == "strength" || effectType == "dexterity" ||
			effectType == "constitution" || effectType == "intelligence" ||
			effectType == "wisdom" || effectType == "charisma"

		shouldBeActive := tickInterval > 0 || duration > 0 || delay > 0 || isStatModifier

		if !shouldBeActive {
			// Apply immediately (only for instant hp/mana/fatigue/hunger changes)
			ApplyImmediateEffect(state, effectType, int(value))
		} else {
			// Add to active effects for over-time processing or permanent stat modifiers
			activeEffect := types.ActiveEffect{
				EffectID:          effectID,
				EffectIndex:       idx,
				DurationRemaining: duration,
				TotalDuration:     duration, // Store original duration for progress calculation
				DelayRemaining:    delay,
				TickAccumulator:   0.0,
				AppliedAt:         state.TimeOfDay,
			}
			state.ActiveEffects = append(state.ActiveEffects, activeEffect)
		}
	}

	// Return effect message
	effectMsg := &types.EffectMessage{
		Message:  message,
		Color:    color,
		Category: category,
		Silent:   silent,
	}

	return effectMsg, nil
}

// ApplyImmediateEffect applies an instant effect (no duration)
// Note: This function modifies fatigue/hunger but does NOT call status update functions
// to avoid circular dependencies. The caller is responsible for updating penalty effects.
func ApplyImmediateEffect(state *types.SaveFile, effectType string, value int) {
	switch effectType {
	case "hp":
		state.HP += value
		if state.HP > state.MaxHP {
			state.HP = state.MaxHP
		}
		if state.HP < 0 {
			state.HP = 0
		}
	case "mana":
		state.Mana += value
		if state.Mana > state.MaxMana {
			state.Mana = state.MaxMana
		}
		if state.Mana < 0 {
			state.Mana = 0
		}
	case "fatigue":
		// Adjust fatigue level (penalty effects handled by caller)
		state.Fatigue += value
		if state.Fatigue < 0 {
			state.Fatigue = 0
		}
		if state.Fatigue > 10 {
			state.Fatigue = 10
		}
	case "hunger":
		// Adjust hunger level (penalty effects handled by caller)
		state.Hunger += value
		if state.Hunger < 0 {
			state.Hunger = 0
		}
		if state.Hunger > 3 {
			state.Hunger = 3
		}
	}
}

// GetEffectTemplate loads effect template data and returns the specific effect at index
func GetEffectTemplate(effectID string, effectIndex int) (effectType string, value float64, tickInterval float64, name string, err error) {
	effectData, err := LoadEffectData(effectID)
	if err != nil {
		return "", 0, 0, "", fmt.Errorf("failed to load effect %s: %v", effectID, err)
	}

	name, _ = effectData["name"].(string)
	effects, _ := effectData["effects"].([]interface{})

	if effects == nil || effectIndex >= len(effects) {
		return "", 0, 0, name, fmt.Errorf("invalid effect index %d for effect %s", effectIndex, effectID)
	}

	effectObj, ok := effects[effectIndex].(map[string]interface{})
	if !ok {
		return "", 0, 0, name, fmt.Errorf("invalid effect data at index %d", effectIndex)
	}

	effectType, _ = effectObj["type"].(string)
	value, _ = effectObj["value"].(float64)
	tickInterval, _ = effectObj["tick_interval"].(float64)

	return effectType, value, tickInterval, name, nil
}

// TickEffects processes all active effects, applying stat modifiers and ticking down durations
// Returns a slice of messages from effects that triggered (like starvation damage)
func TickEffects(state *types.SaveFile, minutesElapsed int) []types.EffectMessage {
	if len(state.ActiveEffects) == 0 {
		return nil
	}

	var remainingEffects []types.ActiveEffect
	var messages []types.EffectMessage

	for _, activeEffect := range state.ActiveEffects {
		// Load effect template to get type, value, tick_interval
		effectType, value, tickInterval, name, err := GetEffectTemplate(activeEffect.EffectID, activeEffect.EffectIndex)
		if err != nil {
			log.Printf("⚠️ Failed to load effect template for %s: %v", activeEffect.EffectID, err)
			continue
		}

		// Tick down delay first
		if activeEffect.DelayRemaining > 0 {
			activeEffect.DelayRemaining -= float64(minutesElapsed)
			if activeEffect.DelayRemaining > 0 {
				remainingEffects = append(remainingEffects, activeEffect)
				continue
			}
		}

		// Process tick-based effects (damage/healing over time)
		if tickInterval > 0 {
			// For hunger accumulation, use dynamic tick interval based on current hunger level
			if activeEffect.EffectID == "hunger-accumulation-stuffed" ||
				activeEffect.EffectID == "hunger-accumulation-wellfed" ||
				activeEffect.EffectID == "hunger-accumulation-hungry" {
				// Override tick interval based on current hunger level
				switch state.Hunger {
				case 3: // Stuffed
					tickInterval = 360 // 6 hours
				case 2: // Well fed
					tickInterval = 240 // 4 hours
				case 1: // Hungry
					tickInterval = 240 // 4 hours
				case 0: // Starving - no accumulation (handled by starving penalty effect)
					tickInterval = 0
				}
			}

			if tickInterval > 0 {
				activeEffect.TickAccumulator += float64(minutesElapsed)
				for activeEffect.TickAccumulator >= tickInterval {
					ApplyImmediateEffect(state, effectType, int(value))
					activeEffect.TickAccumulator -= tickInterval

					// For starvation damage, show message
					if activeEffect.EffectID == "starving" && effectType == "hp" {
						messages = append(messages, types.EffectMessage{
							Message:  "You're starving! You lose 1 HP from lack of food.",
							Color:    "red",
							Category: "debuff",
							Silent:   false,
						})
						log.Printf("💀 Starvation damage: Player lost 1 HP (current HP: %d)", state.HP)
					}
				}
			}
		}

		// Tick down duration (but don't tick permanent effects with duration == 0)
		if activeEffect.DurationRemaining > 0 {
			activeEffect.DurationRemaining -= float64(minutesElapsed)
		}

		// Keep effect if duration remains or is permanent (0)
		// BUT skip accumulation effects if we've hit the cap
		shouldKeep := activeEffect.DurationRemaining > 0 || activeEffect.DurationRemaining == 0

		// Don't keep fatigue-accumulation if fatigue is maxed
		if activeEffect.EffectID == "fatigue-accumulation" && state.Fatigue >= 10 {
			shouldKeep = false
			log.Printf("🛑 Removing fatigue-accumulation: fatigue at max (10)")
		}

		// Don't keep hunger-accumulation effects if starving
		if (activeEffect.EffectID == "hunger-accumulation-stuffed" ||
			activeEffect.EffectID == "hunger-accumulation-wellfed" ||
			activeEffect.EffectID == "hunger-accumulation-hungry") && state.Hunger <= 0 {
			shouldKeep = false
			log.Printf("🛑 Removing hunger-accumulation: hunger at min (0)")
		}

		if shouldKeep {
			remainingEffects = append(remainingEffects, activeEffect)
		} else if activeEffect.DurationRemaining < 0 {
			log.Printf("⏱️ Effect '%s' expired", name)
		}
	}

	state.ActiveEffects = remainingEffects
	return messages
}

// TickDownEffectDurations reduces duration_remaining for all timed effects
// Used during sleep and other time jumps to properly expire buffs/debuffs
func TickDownEffectDurations(state *types.SaveFile, minutes int) {
	if state.ActiveEffects == nil || minutes <= 0 {
		return
	}

	var remainingEffects []types.ActiveEffect
	for _, effect := range state.ActiveEffects {
		// Skip permanent effects (duration == 0) and system effects
		if effect.DurationRemaining == 0 {
			remainingEffects = append(remainingEffects, effect)
			continue
		}

		// Tick down the duration
		effect.DurationRemaining -= float64(minutes)

		if effect.DurationRemaining > 0 {
			// Effect still active
			remainingEffects = append(remainingEffects, effect)
		} else {
			// Effect expired
			log.Printf("⏱️ Effect '%s' expired during time skip (%d minutes)", effect.EffectID, minutes)
		}
	}

	state.ActiveEffects = remainingEffects
}

// GetActiveStatModifiers calculates total stat modifiers from all active effects
func GetActiveStatModifiers(state *types.SaveFile) map[string]int {
	modifiers := make(map[string]int)

	if state.ActiveEffects == nil {
		return modifiers
	}

	for _, activeEffect := range state.ActiveEffects {
		// Skip effects that haven't started yet (still in delay)
		if activeEffect.DelayRemaining > 0 {
			continue
		}

		// Load effect template to get type and value
		effectType, value, _, _, err := GetEffectTemplate(activeEffect.EffectID, activeEffect.EffectIndex)
		if err != nil {
			log.Printf("⚠️ Failed to load effect template for %s: %v", activeEffect.EffectID, err)
			continue
		}

		// Only apply stat modifiers (not instant effects like hp/mana)
		switch effectType {
		case "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma":
			modifiers[effectType] += int(value)
		}
	}

	return modifiers
}

// LoadEffectData loads effect data from database
func LoadEffectData(effectID string) (map[string]interface{}, error) {
	// Normalize old effect IDs for backward compatibility
	effectID = NormalizeEffectID(effectID)

	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	// Query effect properties from database
	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM effects WHERE id = ?", effectID).Scan(&propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("effect not found in database: %s", effectID)
	}

	// Parse properties JSON
	var effectData map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &effectData); err != nil {
		return nil, fmt.Errorf("failed to parse effect properties: %v", err)
	}

	return effectData, nil
}

// EnrichActiveEffects adds template data (name, category, stat_modifiers) to active effects
// Used when sending active_effects to the frontend for display
func EnrichActiveEffects(activeEffects []types.ActiveEffect) []types.EnrichedEffect {
	enriched := make([]types.EnrichedEffect, 0, len(activeEffects))

	for _, ae := range activeEffects {
		ee := types.EnrichedEffect{
			ActiveEffect:  ae,
			Name:          ae.EffectID, // Default to ID
			Description:   "",
			Category:      "modifier",
			StatModifiers: make(map[string]int),
			TickInterval:  0,
		}

		// Load template data
		effectData, err := LoadEffectData(ae.EffectID)
		if err == nil {
			// Get name
			if name, ok := effectData["name"].(string); ok {
				ee.Name = name
			}

			// Get description
			if desc, ok := effectData["description"].(string); ok {
				ee.Description = desc
			}

			// Get category
			if category, ok := effectData["category"].(string); ok {
				ee.Category = category
			}

			// Get effects array to extract stat modifiers and tick interval
			if effects, ok := effectData["effects"].([]interface{}); ok {
				for _, effectRaw := range effects {
					if effect, ok := effectRaw.(map[string]interface{}); ok {
						effectType, _ := effect["type"].(string)
						value, _ := effect["value"].(float64)
						tickInterval, _ := effect["tick_interval"].(float64)

						// Check if this is a stat modifier
						switch effectType {
						case "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma":
							ee.StatModifiers[effectType] = int(value)
						}

						// Get tick interval if present
						if tickInterval > 0 {
							ee.TickInterval = tickInterval
						}
					}
				}
			}
		}

		enriched = append(enriched, ee)
	}

	return enriched
}

// NormalizeEffectID converts old effect IDs to new ones for backward compatibility
func NormalizeEffectID(effectID string) string {
	oldToNew := map[string]string{
		"hunger-accumulation-well-fed":  "hunger-accumulation-wellfed",
		"hunger-accumulation-full":      "hunger-accumulation-stuffed",
		"hunger-accumulation-satisfied": "hunger-accumulation-wellfed",
		"famished":                      "starving",
	}
	if newID, exists := oldToNew[effectID]; exists {
		return newID
	}
	return effectID
}

// MigrateOldEffectIDs updates all effect IDs in ActiveEffects to use new naming conventions
func MigrateOldEffectIDs(state *types.SaveFile) {
	if state.ActiveEffects == nil {
		return
	}
	for i := range state.ActiveEffects {
		oldID := state.ActiveEffects[i].EffectID
		newID := NormalizeEffectID(oldID)
		if newID != oldID {
			log.Printf("🔄 Migrating effect ID: %s -> %s", oldID, newID)
			state.ActiveEffects[i].EffectID = newID
		}
	}
}
