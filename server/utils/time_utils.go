package utils

import "nostr-hero/types"

// GetCurrentScheduleSlot finds the active schedule slot for a given time of day
// timeOfDay: minutes from midnight (0-1439)
// Returns: the active slot, or nil if no schedule exists
func GetCurrentScheduleSlot(schedule []types.NPCScheduleSlot, timeOfDay int) *types.NPCScheduleSlot {
	if len(schedule) == 0 {
		return nil // No schedule defined, NPC uses default behavior
	}

	// Normalize time to 0-1439 range
	normalizedTime := timeOfDay % 1440
	if normalizedTime < 0 {
		normalizedTime += 1440
	}

	for i := range schedule {
		slot := &schedule[i]

		// Handle slots that wrap around midnight
		if slot.End < slot.Start {
			// e.g., 22:00 (1320) to 06:00 (360)
			if normalizedTime >= slot.Start || normalizedTime < slot.End {
				return slot
			}
		} else {
			// Normal slot within same day
			if normalizedTime >= slot.Start && normalizedTime < slot.End {
				return slot
			}
		}
	}

	// Shouldn't happen if schedule covers full 24h, but fallback to first slot
	return &schedule[0]
}

// ResolveNPCSchedule returns current schedule state for an NPC
func ResolveNPCSchedule(npc *types.NPCData, timeOfDay int) *types.NPCScheduleInfo {
	currentSlot := GetCurrentScheduleSlot(npc.Schedule, timeOfDay)

	if currentSlot == nil {
		// No schedule - use default location/building from NPC data (backward compatibility)
		return &types.NPCScheduleInfo{
			CurrentSlot:       nil,
			IsAvailable:       true,
			LocationType:      "building",
			LocationID:        npc.Building,
			State:             "working",
			AvailableDialogue: getAllDialogueKeys(npc.Dialogue),
			AvailableActions:  getAllActions(npc),
		}
	}

	return &types.NPCScheduleInfo{
		CurrentSlot:       currentSlot,
		IsAvailable:       len(currentSlot.AvailableActions) > 0 || len(currentSlot.DialogueOptions) > 0,
		LocationType:      currentSlot.LocationType,
		LocationID:        currentSlot.LocationID,
		State:             currentSlot.State,
		AvailableDialogue: currentSlot.DialogueOptions,
		AvailableActions:  currentSlot.AvailableActions,
	}
}

// getAllDialogueKeys extracts all dialogue keys (for backward compat when no schedule)
func getAllDialogueKeys(dialogue map[string]interface{}) []string {
	if dialogue == nil {
		return []string{}
	}
	keys := make([]string, 0, len(dialogue))
	for k := range dialogue {
		keys = append(keys, k)
	}
	return keys
}

// getAllActions extracts all possible actions from NPC config (for backward compat)
func getAllActions(npc *types.NPCData) []string {
	actions := []string{}
	if npc.ShopConfig != nil {
		actions = append(actions, "open_shop", "open_sell")
	}
	if npc.StorageConfig != nil {
		actions = append(actions, "open_storage", "register_storage")
	}
	if npc.InnConfig != nil {
		actions = append(actions, "rent_room")
	}
	return actions
}
