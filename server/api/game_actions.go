package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"time"

	"nostr-hero/db"
	"nostr-hero/types"
	"nostr-hero/utils"
)

// GameAction represents any action a player can take
type GameAction struct {
	Type   string                 `json:"type"`   // "move", "use_item", "equip", "cast_spell", etc.
	Params map[string]any `json:"params"` // Action-specific parameters
}

// GameActionResponse is returned after processing an action
type GameActionResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Color   string                 `json:"color,omitempty"` // Message color (red, green, yellow, white, purple, blue)
	State   *SaveFile              `json:"state,omitempty"` // Updated game state
	Delta   map[string]any         `json:"delta,omitempty"` // Only changed fields (for optimization)
	Data    map[string]interface{} `json:"data,omitempty"`  // Additional response data
	Error   string                 `json:"error,omitempty"`
}

// GameActionHandler handles all game actions
// POST /api/game/action
func GameActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Npub   string     `json:"npub"`
		SaveID string     `json:"save_id"`
		Action GameAction `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("❌ Failed to decode action request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Npub == "" || request.SaveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	// Get session from memory
	session, err := sessionManager.GetSession(request.Npub, request.SaveID)
	if err != nil {
		// Try to load it if not in memory
		session, err = sessionManager.LoadSession(request.Npub, request.SaveID)
		if err != nil {
			log.Printf("❌ Session not found: %v", err)
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
	}

	// Track last action time for player actions (not update_time ticks)
	if request.Action.Type != "update_time" {
		session.LastActionTime = time.Now().Unix()
		session.LastActionGameTime = session.SaveData.TimeOfDay
	}

	// Process the action based on type
	response, err := processGameAction(session, request.Action)
	if err != nil {
		log.Printf("❌ Action failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GameActionResponse{
			Success: false,
			Error:   err.Error(),
			Message: fmt.Sprintf("Failed to process action: %v", err),
		})
		return
	}

	// Update session in memory
	if err := sessionManager.UpdateSession(request.Npub, request.SaveID, session.SaveData); err != nil {
		log.Printf("❌ Failed to update session: %v", err)
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	// Calculate delta from previous snapshot
	delta := session.UpdateSnapshotAndCalculateDelta()
	if delta != nil && !delta.IsEmpty() {
		calculatedDelta := delta.ToMap()

		// Merge calculated delta with any handler-specific delta (like vault_data)
		// This preserves special data added by handlers while also including state changes
		if response.Delta != nil {
			// Handler already set some delta data - merge with calculated
			for key, value := range calculatedDelta {
				response.Delta[key] = value
			}
		} else {
			response.Delta = calculatedDelta
		}
	}

	// Return updated state (and delta if available)
	response.State = &session.SaveData
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processGameAction routes to specific action handlers
func processGameAction(session *GameSession, action GameAction) (*GameActionResponse, error) {
	state := &session.SaveData

	switch action.Type {
	case "move":
		return handleMoveAction(state, action.Params)
	case "use_item":
		return handleUseItemAction(state, action.Params)
	case "equip_item":
		return handleEquipItemAction(state, action.Params)
	case "unequip_item":
		return handleUnequipItemAction(state, action.Params)
	case "drop_item":
		return handleDropItemAction(state, action.Params)
	case "remove_from_inventory":
		return handleRemoveFromInventoryAction(state, action.Params)
	case "pickup_item":
		return handlePickupItemAction(state, action.Params)
	case "cast_spell":
		return handleCastSpellAction(state, action.Params)
	case "rest":
		return handleRestAction(state, action.Params)
	case "advance_time":
		return handleAdvanceTimeAction(state, action.Params)
	case "update_time":
		return handleUpdateTimeAction(state, action.Params)
	case "vault_deposit":
		return handleVaultDepositAction(state, action.Params)
	case "vault_withdraw":
		return handleVaultWithdrawAction(state, action.Params)
	case "move_item":
		return handleMoveItemAction(state, action.Params)
	case "stack_item":
		return handleStackItemAction(state, action.Params)
	case "split_item":
		return handleSplitItemAction(state, action.Params)
	case "add_item":
		return handleAddItemAction(state, action.Params)
	case "add_to_container":
		return handleAddToContainerAction(state, action.Params)
	case "remove_from_container":
		return handleRemoveFromContainerAction(state, action.Params)
	case "enter_building":
		return handleEnterBuildingAction(state, action.Params)
	case "exit_building":
		return handleExitBuildingAction(state, action.Params)
	case "talk_to_npc":
		return handleTalkToNPCAction(state, action.Params)
	case "npc_dialogue_choice":
		return handleNPCDialogueChoiceAction(session, action.Params)
	case "register_vault":
		return handleRegisterVaultAction(state, action.Params)
	case "open_vault":
		return handleOpenVaultAction(state, action.Params)
	case "rent_room":
		return handleRentRoomAction(session, action.Params)
	case "sleep":
		return handleSleepAction(session, action.Params)
	case "wait":
		return handleWaitAction(session, action.Params)
	case "book_show":
		return handleBookShowAction(session, action.Params)
	case "play_show":
		return handlePlayShowAction(session, action.Params)
	case "reset_idle_timer":
		return handleResetIdleTimerAction(session)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// ============================================================================
// ACTION HANDLERS
// ============================================================================

// pluralize returns "s" if count != 1, empty string otherwise
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// handleMoveAction moves the player to a new location
func handleMoveAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	location, ok := params["location"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid location parameter")
	}

	district, _ := params["district"].(string)
	building, _ := params["building"].(string)

	// Validate location exists (TODO: check against location database)

	// Update state
	state.Location = location
	state.District = district
	state.Building = building

	// Add to discovered locations if not already there
	if !slices.Contains(state.LocationsDiscovered, location) {
		state.LocationsDiscovered = append(state.LocationsDiscovered, location)
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Moved to %s", location),
	}, nil
}

// handleUseItemAction uses a consumable item
func handleUseItemAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	slot := -1
	if s, ok := params["slot"].(float64); ok {
		slot = int(s)
	}

	// Find and remove item from inventory (check both general and backpack)
	var itemFound bool
	var effects []string

	// Check general slots
	generalSlots, ok := state.Inventory["general_slots"].([]any)
	if ok {
		for i, slotData := range generalSlots {
			slotMap, ok := slotData.(map[string]any)
			if !ok {
				continue
			}

			if slotMap["item"] == itemID && (slot < 0 || i == slot) {
				itemFound = true

				// Apply item effects based on item ID (hardcoded for common items)
				effects = applyItemEffects(state, itemID)

				// Remove/reduce item quantity
				qty, _ := slotMap["quantity"].(float64)
				if qty > 1 {
					slotMap["quantity"] = qty - 1
				} else {
					slotMap["item"] = nil
					slotMap["quantity"] = 0
				}
				break
			}
		}
	}

	// If not found in general, check backpack
	if !itemFound {
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		contents, ok := bag["contents"].([]any)
		if ok {
			for i, slotData := range contents {
				slotMap, ok := slotData.(map[string]any)
				if !ok {
					continue
				}

				if slotMap["item"] == itemID && (slot < 0 || i == slot) {
					itemFound = true

					// Apply item effects
					effects = applyItemEffects(state, itemID)

					// Remove/reduce item quantity
					qty, _ := slotMap["quantity"].(float64)
					if qty > 1 {
						slotMap["quantity"] = qty - 1
					} else {
						slotMap["item"] = nil
						slotMap["quantity"] = 0
					}
					break
				}
			}
		}
	}

	if !itemFound {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}

	effectMsg := "Used item"
	if len(effects) > 0 {
		effectMsg = fmt.Sprintf("Used %s: %s", itemID, effects[0])
	}

	return &GameActionResponse{
		Success: true,
		Message: effectMsg,
	}, nil
}

// applyItemEffects applies item effects to the character state dynamically from item data
func applyItemEffects(state *SaveFile, itemID string) []string {
	var effectMessages []string

	// Get database connection
	database := db.GetDB()
	if database == nil {
		log.Printf("⚠️ Database not available, cannot apply effects for %s", itemID)
		return []string{"Used"}
	}

	// Query item from database to get properties
	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM items WHERE id = ?", itemID).Scan(&propertiesJSON)
	if err != nil {
		log.Printf("⚠️ Could not find item %s in database: %v", itemID, err)
		return []string{"Used"}
	}

	// Parse properties JSON
	var properties map[string]any
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
		log.Printf("⚠️ Could not parse properties for item %s: %v", itemID, err)
		return []string{"Used"}
	}

	// Check if item has effects array
	effectsRaw, hasEffects := properties["effects"]
	if !hasEffects {
		log.Printf("⚠️ Item %s has no effects defined", itemID)
		return []string{"Used"}
	}

	// Parse effects array
	effectsArray, ok := effectsRaw.([]any)
	if !ok {
		log.Printf("⚠️ Item %s has invalid effects format", itemID)
		return []string{"Used"}
	}

	// Apply each effect
	for _, effectRaw := range effectsArray {
		effectMap, ok := effectRaw.(map[string]any)
		if !ok {
			continue
		}

		effectType, _ := effectMap["type"].(string)
		effectValue, _ := effectMap["value"].(float64) // JSON numbers are float64

		switch effectType {
		case "hp", "health":
			oldHP := state.HP
			state.HP = min(state.MaxHP, state.HP+int(effectValue))
			actualHealed := state.HP - oldHP
			if actualHealed > 0 {
				effectMessages = append(effectMessages, fmt.Sprintf("Healed %d HP", actualHealed))
			}

		case "mana":
			oldMana := state.Mana
			state.Mana = min(state.MaxMana, state.Mana+int(effectValue))
			actualRestored := state.Mana - oldMana
			if actualRestored > 0 {
				effectMessages = append(effectMessages, fmt.Sprintf("Restored %d mana", actualRestored))
			}

		case "hunger":
			oldHunger := state.Hunger
			state.Hunger = min(3, max(0, state.Hunger+int(effectValue)))
			_, _ = updateHungerPenaltyEffects(state)
			ensureHungerAccumulation(state)
			if state.Hunger > oldHunger {
				effectMessages = append(effectMessages, "Hunger restored")
			} else if state.Hunger < oldHunger {
				effectMessages = append(effectMessages, "Hunger decreased")
			}

		case "fatigue":
			oldFatigue := state.Fatigue
			state.Fatigue = max(0, min(10, state.Fatigue+int(effectValue)))

			// Stop accumulation if we've reached max fatigue
			if state.Fatigue >= 10 {
				removeFatigueAccumulation(state)
			} else {
				// Ensure accumulation is active if below max
				ensureFatigueAccumulation(state)
			}

			_, _ = updateFatiguePenaltyEffects(state)
			if state.Fatigue < oldFatigue {
				fatigueReduced := oldFatigue - state.Fatigue
				effectMessages = append(effectMessages, fmt.Sprintf("Fatigue reduced by %d", fatigueReduced))
			} else if state.Fatigue > oldFatigue {
				fatigueIncreased := state.Fatigue - oldFatigue
				effectMessages = append(effectMessages, fmt.Sprintf("Fatigue increased by %d", fatigueIncreased))
			}

		default:
			log.Printf("⚠️ Unknown effect type: %s", effectType)
		}
	}

	if len(effectMessages) == 0 {
		return []string{"Used"}
	}

	return effectMessages
}

// Equipment handlers are now in equipment.go

// handleDropItemAction drops an item from inventory
func handleDropItemAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	slot, _ := params["slot"].(float64)
	slotType, _ := params["slot_type"].(string)
	if slotType == "" {
		slotType = "general" // Default to general slots
	}

	// Get the quantity to drop (default to all if not specified)
	dropQuantity := -1
	if qty, ok := params["quantity"].(float64); ok {
		dropQuantity = int(qty)
	}

	log.Printf("📤 Dropping %s: quantity=%d from %s[%d]", itemID, dropQuantity, slotType, int(slot))

	// Find item in appropriate inventory
	var itemFound bool
	var inventory []any

	switch slotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid inventory structure")
		}
		inventory = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		backpackContents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack structure")
		}
		inventory = backpackContents
	default:
		return nil, fmt.Errorf("invalid slot_type: %s", slotType)
	}

	// Search for item
	for i, slotData := range inventory {
		slotMap, ok := slotData.(map[string]any)
		if !ok {
			continue
		}

		if slotMap["item"] == itemID && (slot < 0 || i == int(slot)) {
			itemFound = true

			// Get current quantity
			currentQty := 1
			if qty, ok := slotMap["quantity"].(float64); ok {
				currentQty = int(qty)
			}

			// Determine how much to drop
			if dropQuantity <= 0 || dropQuantity >= currentQty {
				// Drop entire stack
				slotMap["item"] = nil
				slotMap["quantity"] = 0
				log.Printf("✅ Dropped entire stack of %s (%d items)", itemID, currentQty)
			} else {
				// Drop partial stack (store as int)
				slotMap["quantity"] = currentQty - dropQuantity
				log.Printf("✅ Dropped %d %s (keeping %d)", dropQuantity, itemID, currentQty-dropQuantity)
			}
			break
		}
	}

	if !itemFound {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Dropped %s", itemID),
	}, nil
}

// handleRemoveFromInventoryAction removes an item from inventory (for sell staging)
func handleRemoveFromInventoryAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	fromSlot, _ := params["from_slot"].(float64)
	fromSlotType, _ := params["from_slot_type"].(string)
	if fromSlotType == "" {
		fromSlotType = "general" // Default to general slots
	}

	// Get the quantity to remove (default to 1)
	removeQuantity := 1
	if qty, ok := params["quantity"].(float64); ok {
		removeQuantity = int(qty)
	}

	log.Printf("🛒 Removing %dx %s from %s[%d] for sell staging", removeQuantity, itemID, fromSlotType, int(fromSlot))

	// Find item in appropriate inventory
	var itemFound bool
	var inventory []any

	switch fromSlotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid inventory structure")
		}
		inventory = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		backpackContents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack structure")
		}
		inventory = backpackContents
	default:
		return nil, fmt.Errorf("invalid slot_type: %s", fromSlotType)
	}

	// Search for item at specific slot
	if int(fromSlot) < 0 || int(fromSlot) >= len(inventory) {
		return nil, fmt.Errorf("invalid slot index: %d", int(fromSlot))
	}

	slotMap, ok := inventory[int(fromSlot)].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid slot data at index %d", int(fromSlot))
	}

	if slotMap["item"] != itemID {
		return nil, fmt.Errorf("item mismatch: expected %s, found %v", itemID, slotMap["item"])
	}

	itemFound = true

	// Get current quantity (ensure it's an integer)
	currentQty := 1
	if qty, ok := slotMap["quantity"].(float64); ok {
		currentQty = int(qty)
	} else if qty, ok := slotMap["quantity"].(int); ok {
		currentQty = qty
	}

	log.Printf("🔢 Current quantity at slot: %d (type: %T)", currentQty, slotMap["quantity"])

	if currentQty < removeQuantity {
		return nil, fmt.Errorf("not enough items: have %d, trying to remove %d", currentQty, removeQuantity)
	}

	// Remove from stack (ALWAYS store as int, not float64)
	newQty := currentQty - removeQuantity
	if newQty <= 0 {
		// Remove entire stack
		slotMap["item"] = nil
		slotMap["quantity"] = 0
		log.Printf("✅ Removed entire stack of %s (%d items)", itemID, currentQty)
	} else {
		// Remove partial stack - store as int
		slotMap["quantity"] = newQty
		log.Printf("✅ Removed %d %s (keeping %d) - stored as %T", removeQuantity, itemID, newQty, slotMap["quantity"])
	}

	if !itemFound {
		return nil, fmt.Errorf("item not found: %s at slot %d", itemID, int(fromSlot))
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Removed %dx %s from inventory", removeQuantity, itemID),
	}, nil
}

// handlePickupItemAction picks up an item from the ground
func handlePickupItemAction(_ *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	// TODO: Validate item is on the ground at current location
	// TODO: Find empty inventory slot
	// TODO: Add item to inventory

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Picked up %s", itemID),
	}, nil
}

// handleCastSpellAction casts a spell
func handleCastSpellAction(_ *SaveFile, params map[string]any) (*GameActionResponse, error) {
	spellID, ok := params["spell_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid spell_id parameter")
	}

	// TODO: Validate spell is known
	// TODO: Check mana cost
	// TODO: Apply spell effects
	// TODO: Reduce mana

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Cast %s", spellID),
	}, nil
}

// handleRestAction rests to restore HP/Mana
func handleRestAction(state *SaveFile, _ map[string]any) (*GameActionResponse, error) {
	// Restore HP and Mana
	state.HP = state.MaxHP
	state.Mana = state.MaxMana
	state.Fatigue = 0
	removeFatiguePenaltyEffects(state)

	// Advance time
	state.TimeOfDay = (state.TimeOfDay + 8) % 24 // Rest for 8 hours
	if state.TimeOfDay < 8 {
		state.CurrentDay++
	}

	return &GameActionResponse{
		Success: true,
		Message: "Rested and restored HP/Mana",
	}, nil
}

// handleAdvanceTimeAction advances game time
func handleAdvanceTimeAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	segments, ok := params["segments"].(float64)
	if !ok {
		segments = 1 // Default to 1 time segment
	}

	// Convert hours to minutes (segments is in hours, time_of_day is now in minutes)
	minutesToAdvance := int(segments) * 60

	// Advance time in minutes (0-1439 range)
	state.TimeOfDay += minutesToAdvance
	daysAdvanced := state.TimeOfDay / 1440
	state.CurrentDay += daysAdvanced
	state.TimeOfDay = state.TimeOfDay % 1440

	// Fatigue and hunger are now handled by accumulation effects (fatigue-accumulation, hunger-accumulation-*)
	// No need to manually tick them here

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Advanced %d time segment(s)", int(segments)),
	}, nil
}

// handleUpdateTimeAction syncs time from frontend clock to backend state
// This is the main tick handler that updates buildings, NPCs, and effects
func handleUpdateTimeAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	// Get the new time from frontend
	newTimeOfDay, timeOk := params["time_of_day"].(float64)
	newCurrentDay, dayOk := params["current_day"].(float64)

	if !timeOk || !dayOk {
		return &GameActionResponse{
			Success: false,
			Message: "Missing time parameters",
		}, nil
	}

	// Get session for delta tracking
	npub := state.InternalNpub
	saveID := state.InternalID
	session, err := sessionManager.GetSession(npub, saveID)
	if err != nil {
		// Session not found - still process time but skip delta
		log.Printf("⚠️ Session not found for delta: %s:%s", npub, saveID)
	}

	// Check for auto-pause: if 6+ in-game hours have passed since last player action
	autoPause := false
	if session != nil && session.LastActionGameTime > 0 {
		// Calculate in-game minutes since last action
		newTime := int(newTimeOfDay)
		newDay := int(newCurrentDay)

		// Calculate total minutes elapsed since last action
		var minutesSinceAction int
		if newDay == state.CurrentDay {
			minutesSinceAction = newTime - session.LastActionGameTime
		} else {
			// Handle day wrap
			minutesSinceAction = (1440 - session.LastActionGameTime) + newTime + ((newDay - state.CurrentDay - 1) * 1440)
		}

		// Auto-pause after 6 in-game hours (360 minutes) of idle
		if minutesSinceAction >= 360 {
			autoPause = true
			log.Printf("⏸️ Auto-pause triggered: %d in-game minutes since last action", minutesSinceAction)
		}
	}

	// Calculate time delta
	oldTime := state.TimeOfDay
	oldDay := state.CurrentDay
	newTime := int(newTimeOfDay)
	newDay := int(newCurrentDay)

	// Calculate total minutes elapsed
	var minutesElapsed int
	if newDay == oldDay {
		// Same day
		minutesElapsed = newTime - oldTime
	} else {
		// Day(s) advanced
		minutesElapsed = (1440 - oldTime) + newTime + ((newDay - oldDay - 1) * 1440)
	}

	// Only process if time actually advanced
	if minutesElapsed > 0 {
		// Use advanceTime to properly process effects
		advanceTime(state, minutesElapsed)
	}

	// Check for missed shows (session-only data) and apply penalty
	if session != nil {
		if len(session.BookedShows) > 0 {
			checkMissedShows(state, session, newTime, newDay)
		}
	}

	// Update buildings and NPCs if we have a session
	if session != nil {
		database := db.GetDB()

		// Update building states if needed (every 5 in-game minutes or first call)
		if session.ShouldRefreshBuildings(newTime) {
			buildingStates, err := utils.GetAllBuildingStatesForDistrict(
				database,
				state.Location,
				state.District,
				newTime,
			)
			if err == nil && len(buildingStates) > 0 {
				session.UpdateBuildingStates(buildingStates, newTime)
			}
		}

		// Update NPCs if hour changed
		currentHour := newTime / 60
		if session.ShouldRefreshNPCs(currentHour) {
			npcIDs := GetNPCIDsAtLocation(
				state.Location,
				state.District,
				state.Building,
				newTime,
			)
			// Only log when NPCs actually change (reduces spam)
			if len(npcIDs) > 0 || len(session.NPCsAtLocation) > 0 {
				log.Printf("🧑 NPCs updated: hour=%d, %s/%s/%s, was=%v, now=%v",
					currentHour, state.Location, state.District, state.Building, session.NPCsAtLocation, npcIDs)
			}
			session.UpdateNPCsAtLocation(npcIDs, currentHour)
		}

		// Calculate delta
		delta := session.UpdateSnapshotAndCalculateDelta()
		if delta != nil && delta.NPCs != nil {
			log.Printf("🧑 NPC delta: added=%v, removed=%v", delta.NPCs.Added, delta.NPCs.Removed)
		}
		if delta != nil && !delta.IsEmpty() {
			return &GameActionResponse{
				Success: true,
				Message: "Time updated",
				Delta:   delta.ToMap(),
				Data: map[string]interface{}{
					"time_of_day":     state.TimeOfDay,
					"current_day":     state.CurrentDay,
					"fatigue":         state.Fatigue,
					"hunger":          state.Hunger,
					"hp":              state.HP,
					"active_effects":  enrichActiveEffects(state.ActiveEffects),
					"auto_pause":      autoPause,
				},
			}, nil
		}
	}

	// Return updated state so frontend can sync (fallback if no session/delta)
	return &GameActionResponse{
		Success: true,
		Message: "Time updated",
		Data: map[string]interface{}{
			"time_of_day":     state.TimeOfDay,
			"current_day":     state.CurrentDay,
			"fatigue":         state.Fatigue,
			"hunger":          state.Hunger,
			"hp":              state.HP,
			"active_effects":  enrichActiveEffects(state.ActiveEffects),
			"auto_pause":      autoPause,
		},
	}, nil
}

// handleVaultDepositAction deposits items into vault (uses existing move_item action for vault transfers)
func handleVaultDepositAction(_ *SaveFile, _ map[string]any) (*GameActionResponse, error) {
	// Vaults work like containers - use the container system
	// This is handled by frontend calling move_item or add_to_container with vault as destination
	return &GameActionResponse{
		Success: true,
		Message: "Item deposited to vault",
	}, nil
}

// handleVaultWithdrawAction withdraws items from vault (uses existing move_item action for vault transfers)
func handleVaultWithdrawAction(_ *SaveFile, _ map[string]any) (*GameActionResponse, error) {
	// Vaults work like containers - use the container system
	// This is handled by frontend calling move_item or remove_from_container with vault as source
	return &GameActionResponse{
		Success: true,
		Message: "Item withdrawn from vault",
	}, nil
}

// handleMoveItemAction moves/swaps items between inventory slots
func handleMoveItemAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	fromSlot := int(params["from_slot"].(float64))
	toSlot := int(params["to_slot"].(float64))
	fromSlotType, _ := params["from_slot_type"].(string)
	toSlotType, _ := params["to_slot_type"].(string)

	// Get the appropriate slot arrays
	var fromSlots, toSlots []any
	var vaultBuilding string

	// Get vault building ID if dealing with vault
	if params["vault_building"] != nil {
		vaultBuilding, _ = params["vault_building"].(string)
	}

	// Get from slots
	switch fromSlotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		fromSlots = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		contents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		fromSlots = contents
	case "vault":
		vault := getVaultForLocation(state, vaultBuilding)
		if vault == nil {
			return nil, fmt.Errorf("vault not found for building: %s", vaultBuilding)
		}
		slots, ok := vault["slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid vault slots")
		}
		fromSlots = slots
	}

	// Get to slots
	switch toSlotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		toSlots = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		contents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		toSlots = contents
	case "vault":
		vault := getVaultForLocation(state, vaultBuilding)
		if vault == nil {
			return nil, fmt.Errorf("vault not found for building: %s", vaultBuilding)
		}
		slots, ok := vault["slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid vault slots")
		}
		toSlots = slots
	}

	// CRITICAL VALIDATION: Containers cannot go into backpack
	if toSlotType == "inventory" {
		log.Printf("🔍 VALIDATION CHECK: Is '%s' a container? (destination: backpack)", itemID)

		database := db.GetDB()
		if database == nil {
			log.Printf("❌ CRITICAL: Database not available")
			return &GameActionResponse{
				Success: false,
				Error:   "System error: Cannot validate item restrictions",
				Color:   "red",
			}, nil
		}

		// Query tags from database
		var tagsJSON string
		err := database.QueryRow("SELECT tags FROM items WHERE id = ?", itemID).Scan(&tagsJSON)
		if err != nil {
			log.Printf("❌ CRITICAL: Failed to query tags for %s: %v", itemID, err)
			return &GameActionResponse{
				Success: false,
				Error:   fmt.Sprintf("System error: Cannot find item %s", itemID),
				Color:   "red",
			}, nil
		}

		log.Printf("📦 Raw tags JSON from database for '%s': %s", itemID, tagsJSON)

		var tags []any
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			log.Printf("❌ CRITICAL: Failed to parse tags JSON for %s: %v", itemID, err)
			return &GameActionResponse{
				Success: false,
				Error:   "System error: Invalid item data format",
				Color:   "red",
			}, nil
		}

		log.Printf("📦 Parsed tags array for '%s': %v", itemID, tags)

		// Check each tag
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				log.Printf("   🏷️ Found tag: '%s'", tagStr)
				if tagStr == "container" {
					log.Printf("❌ BLOCKED: '%s' has 'container' tag - CANNOT go in backpack!", itemID)
					return &GameActionResponse{
						Success: false,
						Error:   "Containers cannot be stored in the backpack",
						Color:   "red",
					}, nil
				}
			}
		}

		log.Printf("✅ VALIDATION PASSED: '%s' is NOT a container - allowing move to backpack", itemID)
	}

	// ADDITIONAL VALIDATION: Check displaced item in swap scenarios
	// If dragging FROM backpack and swapping with a container, the container would go INTO backpack (not allowed!)
	log.Printf("🔍 Checking swap validation: fromSlotType=%s, toSlotType=%s", fromSlotType, toSlotType)
	if fromSlotType == "inventory" && toSlotType != "inventory" {
		log.Printf("🔍 Condition met: dragging FROM backpack TO %s", toSlotType)
		// Check if we're swapping (destination slot is not empty)
		if toSlots != nil && toSlot < len(toSlots) {
			log.Printf("🔍 Checking destination slot %d (toSlots length: %d)", toSlot, len(toSlots))
			if destSlot, ok := toSlots[toSlot].(map[string]any); ok {
				log.Printf("🔍 Destination slot data: %+v", destSlot)
				if destItem, ok := destSlot["item"].(string); ok && destItem != "" {
					// There's an item in the destination - this is a swap
					// Check if the displaced item is a container
					log.Printf("🔍 SWAP VALIDATION: Checking if displaced item '%s' is a container (would go to backpack)", destItem)

					database := db.GetDB()
					if database != nil {
						var tagsJSON string
						err := database.QueryRow("SELECT tags FROM items WHERE id = ?", destItem).Scan(&tagsJSON)
						if err == nil {
							var tags []any
							if err := json.Unmarshal([]byte(tagsJSON), &tags); err == nil {
								for _, tag := range tags {
									if tagStr, ok := tag.(string); ok && tagStr == "container" {
										log.Printf("❌ BLOCKED: Displaced item '%s' is a container and cannot go in backpack via swap!", destItem)
										return &GameActionResponse{
											Success: false,
											Error:   "Containers cannot be stored in the backpack",
											Color:   "red",
										}, nil
									}
								}
							}
						}
					}
					log.Printf("✅ Swap validated: Displaced item '%s' is not a container", destItem)
				}
			}
		}
	}

	// Swap items
	if fromSlots != nil && toSlots != nil && fromSlot >= 0 && toSlot >= 0 {
		if fromSlotType == toSlotType {
			// Same array, just swap within it
			fromSlots[fromSlot], fromSlots[toSlot] = fromSlots[toSlot], fromSlots[fromSlot]

			// Update the "slot" field in each swapped item
			if fromSlotMap, ok := fromSlots[fromSlot].(map[string]any); ok {
				fromSlotMap["slot"] = fromSlot
			}
			if toSlotMap, ok := fromSlots[toSlot].(map[string]any); ok {
				toSlotMap["slot"] = toSlot
			}
		} else {
			// Different arrays, swap between them
			temp := fromSlots[fromSlot]
			fromSlots[fromSlot] = toSlots[toSlot]
			toSlots[toSlot] = temp

			// Update the "slot" field in each swapped item
			if fromSlotMap, ok := fromSlots[fromSlot].(map[string]any); ok {
				fromSlotMap["slot"] = fromSlot
			}
			if toSlotMap, ok := toSlots[toSlot].(map[string]any); ok {
				toSlotMap["slot"] = toSlot
			}
		}

		log.Printf("✅ Swapped slots: %s[%d] ↔ %s[%d]", fromSlotType, fromSlot, toSlotType, toSlot)
	}

	// If vault was involved, return updated vault data
	delta := map[string]any{}
	if fromSlotType == "vault" || toSlotType == "vault" {
		log.Printf("🏦 Vault involved: from=%s, to=%s, building=%s", fromSlotType, toSlotType, vaultBuilding)
		vault := getVaultForLocation(state, vaultBuilding)
		if vault != nil {
			delta["vault_data"] = vault
			log.Printf("✅ Returning updated vault data with %d slots", len(vault["slots"].([]any)))
		} else {
			log.Printf("⚠️ Vault not found for building: %s", vaultBuilding)
		}
	}

	response := &GameActionResponse{
		Success: true,
		Message: "", // Suppressed - no need to show success message for moves
	}
	if len(delta) > 0 {
		response.Delta = delta
	}

	return response, nil
}

// handleStackItemAction stacks items together
func handleStackItemAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	fromSlot := int(params["from_slot"].(float64))
	toSlot := int(params["to_slot"].(float64))
	fromSlotType, _ := params["from_slot_type"].(string)
	toSlotType, _ := params["to_slot_type"].(string)

	// Get source slots
	var fromSlots []any
	switch fromSlotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		fromSlots = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		contents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		fromSlots = contents
	default:
		return nil, fmt.Errorf("invalid source slot type: %s", fromSlotType)
	}

	// Get destination slots
	var toSlots []any
	switch toSlotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		toSlots = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		contents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		toSlots = contents
	default:
		return nil, fmt.Errorf("invalid destination slot type: %s", toSlotType)
	}

	// Get source and destination items
	fromSlotMap, ok := fromSlots[fromSlot].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid source slot data")
	}

	toSlotMap, ok := toSlots[toSlot].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid destination slot data")
	}

	// Verify both slots have the same item
	if fromSlotMap["item"] != itemID || toSlotMap["item"] != itemID {
		return nil, fmt.Errorf("items don't match for stacking")
	}

	// Get item data to check max stack
	itemData, err := db.GetItemByID(itemID)
	if err != nil {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}

	// Parse max stack from properties
	maxStack := 1
	if itemData.Properties != "" {
		var properties map[string]any
		if err := json.Unmarshal([]byte(itemData.Properties), &properties); err == nil {
			if val, ok := properties["stack"].(float64); ok {
				maxStack = int(val)
			}
		}
	}

	// Get quantities - handle both int and float64 types
	var fromQty, toQty int

	// Convert from slot quantity (handle both int and float64)
	switch v := fromSlotMap["quantity"].(type) {
	case float64:
		fromQty = int(v)
	case int:
		fromQty = v
	default:
		fromQty = 0
	}

	// Convert to slot quantity (handle both int and float64)
	switch v := toSlotMap["quantity"].(type) {
	case float64:
		toQty = int(v)
	case int:
		toQty = v
	default:
		toQty = 0
	}

	log.Printf("📊 Stack quantities - From slot: qty=%v (type=%T), To slot: qty=%v (type=%T), max stack: %d",
		fromSlotMap["quantity"], fromSlotMap["quantity"],
		toSlotMap["quantity"], toSlotMap["quantity"], maxStack)
	log.Printf("📊 Converted - fromQty=%d, toQty=%d", fromQty, toQty)

	// Check if destination is already at max
	if toQty >= maxStack {
		return nil, fmt.Errorf("destination stack is full (max %d)", maxStack)
	}

	// Calculate how much can be added
	canAdd := maxStack - toQty
	if canAdd > fromQty {
		canAdd = fromQty
	}

	// Update destination slot
	toSlotMap["quantity"] = toQty + canAdd

	// Update or clear source slot
	remaining := fromQty - canAdd
	if remaining > 0 {
		fromSlotMap["quantity"] = remaining
		log.Printf("✅ Stacked %s: moved %d from %d to %d (now %d total, %d remaining in source)", itemID, canAdd, fromQty, toQty, toQty+canAdd, remaining)
	} else {
		fromSlotMap["item"] = nil
		fromSlotMap["quantity"] = 0
		log.Printf("✅ Stacked %s: %d + %d = %d (source cleared)", itemID, fromQty, toQty, toQty+canAdd)
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Stacked items (%d total)", toQty+canAdd),
	}, nil
}

// handleSplitItemAction splits a stack into two stacks
func handleSplitItemAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	fromSlot := int(params["from_slot"].(float64))
	toSlot := int(params["to_slot"].(float64))
	fromSlotType, _ := params["from_slot_type"].(string)
	toSlotType, _ := params["to_slot_type"].(string)
	splitQuantity := int(params["quantity"].(float64))

	log.Printf("✂️ Splitting %s: %d from %s[%d] to %s[%d]", itemID, splitQuantity, fromSlotType, fromSlot, toSlotType, toSlot)

	// Get source slot
	var fromSlots []any
	switch fromSlotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		fromSlots = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		contents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		fromSlots = contents
	default:
		return nil, fmt.Errorf("invalid source slot type: %s", fromSlotType)
	}

	// Get destination slot
	var toSlots []any
	switch toSlotType {
	case "general":
		generalSlots, ok := state.Inventory["general_slots"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		toSlots = generalSlots
	case "inventory":
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]any)
		bag, _ := gearSlots["bag"].(map[string]any)
		contents, ok := bag["contents"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		toSlots = contents
	default:
		return nil, fmt.Errorf("invalid destination slot type: %s", toSlotType)
	}

	// Validate slots exist
	if fromSlot < 0 || fromSlot >= len(fromSlots) {
		return nil, fmt.Errorf("invalid from slot: %d", fromSlot)
	}
	if toSlot < 0 || toSlot >= len(toSlots) {
		return nil, fmt.Errorf("invalid to slot: %d", toSlot)
	}

	// Get source item
	fromSlotMap, ok := fromSlots[fromSlot].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid source slot data")
	}

	// Verify item ID matches
	if fromSlotMap["item"] != itemID {
		return nil, fmt.Errorf("item mismatch in source slot")
	}

	// Get current quantity (handle both int and float64 types)
	var currentQty int
	switch v := fromSlotMap["quantity"].(type) {
	case float64:
		currentQty = int(v)
	case int:
		currentQty = v
	default:
		return nil, fmt.Errorf("invalid quantity in source slot")
	}

	// Validate split quantity
	if splitQuantity <= 0 || splitQuantity >= currentQty {
		return nil, fmt.Errorf("invalid split quantity: %d (current: %d)", splitQuantity, currentQty)
	}

	// Check destination slot is empty
	toSlotMap, ok := toSlots[toSlot].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid destination slot data")
	}
	if toSlotMap["item"] != nil && toSlotMap["item"] != "" {
		return nil, fmt.Errorf("destination slot is not empty")
	}

	// Perform split
	remainingQty := currentQty - splitQuantity

	// Update source slot (store as int, not float64)
	fromSlotMap["quantity"] = remainingQty

	// Update destination slot (store as int, not float64)
	toSlotMap["item"] = itemID
	toSlotMap["quantity"] = splitQuantity
	toSlotMap["slot"] = toSlot

	log.Printf("✅ Split complete: %s (%d remaining in slot %d, %d in new slot %d)", itemID, remainingQty, fromSlot, splitQuantity, toSlot)

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Split %d items into new stack", splitQuantity),
	}, nil
}

// handleAddItemAction adds an item to inventory
func handleAddItemAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	quantity := 1
	if q, ok := params["quantity"].(float64); ok {
		quantity = int(q)
	}

	log.Printf("➕ Adding %dx %s to inventory", quantity, itemID)

	// Try general slots first
	generalSlots, ok := state.Inventory["general_slots"].([]any)
	if ok {
		for i, slotData := range generalSlots {
			slotMap, ok := slotData.(map[string]any)
			if !ok {
				continue
			}

			// Check if slot is empty
			if slotMap["item"] == nil || slotMap["item"] == "" {
				// Add item to this slot (ensure quantity is int)
				slotMap["item"] = itemID
				slotMap["quantity"] = int(quantity)
				log.Printf("✅ Added %dx %s to general_slots[%d] (type: %T)", quantity, itemID, i, slotMap["quantity"])

				return &GameActionResponse{
					Success: true,
					Message: fmt.Sprintf("Added %dx %s", quantity, itemID),
				}, nil
			}
		}
	}

	// Try backpack if general slots are full
	gearSlots, ok := state.Inventory["gear_slots"].(map[string]any)
	if ok {
		bag, ok := gearSlots["bag"].(map[string]any)
		if ok {
			backpack, ok := bag["contents"].([]any)
			if ok {
				for i, slotData := range backpack {
					slotMap, ok := slotData.(map[string]any)
					if !ok {
						continue
					}

					// Check if slot is empty
					if slotMap["item"] == nil || slotMap["item"] == "" {
						// Add item to this slot (ensure quantity is int)
						slotMap["item"] = itemID
						slotMap["quantity"] = int(quantity)
						log.Printf("✅ Added %dx %s to backpack[%d] (type: %T)", quantity, itemID, i, slotMap["quantity"])

						return &GameActionResponse{
							Success: true,
							Message: fmt.Sprintf("Added %dx %s", quantity, itemID),
						}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("inventory is full")
}

// GetGameStateHandler returns the current game state
// GET /api/game/state?npub={npub}&save_id={saveID}
func GetGameStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	npub := r.URL.Query().Get("npub")
	saveID := r.URL.Query().Get("save_id")

	if npub == "" || saveID == "" {
		http.Error(w, "Missing npub or save_id", http.StatusBadRequest)
		return
	}

	// Get session from memory
	session, err := sessionManager.GetSession(npub, saveID)
	if err != nil {
		// Try to load it if not in memory
		session, err = sessionManager.LoadSession(npub, saveID)
		if err != nil {
			log.Printf("❌ Failed to get session: %v", err)
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
	}

	// Include session-specific data in the response
	stateWithSession := map[string]any{
		"success": true,
		"state": map[string]any{
			// Include all SaveData fields
			"d":                    session.SaveData.D,
			"created_at":           session.SaveData.CreatedAt,
			"race":                 session.SaveData.Race,
			"class":                session.SaveData.Class,
			"background":           session.SaveData.Background,
			"alignment":            session.SaveData.Alignment,
			"experience":           session.SaveData.Experience,
			"hp":                   session.SaveData.HP,
			"max_hp":               session.SaveData.MaxHP,
			"mana":                 session.SaveData.Mana,
			"max_mana":             session.SaveData.MaxMana,
			"fatigue":              session.SaveData.Fatigue,
			"hunger":               session.SaveData.Hunger,
			"stats":                session.SaveData.Stats,
			"location":             session.SaveData.Location,
			"district":             session.SaveData.District,
			"building":             session.SaveData.Building,
			"current_day":          session.SaveData.CurrentDay,
			"time_of_day":          session.SaveData.TimeOfDay,
			"inventory":            session.SaveData.Inventory,
			"vaults":               session.SaveData.Vaults,
			"known_spells":         session.SaveData.KnownSpells,
			"spell_slots":          session.SaveData.SpellSlots,
			"locations_discovered": session.SaveData.LocationsDiscovered,
			"music_tracks_unlocked": session.SaveData.MusicTracksUnlocked,
			"active_effects":       session.SaveData.ActiveEffects,

			// Add session-specific data
			"rented_rooms":    session.RentedRooms,
			"booked_shows":    session.BookedShows,
			"performed_shows": session.PerformedShows,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stateWithSession)
}

// handleEnterBuildingAction enters a building
func handleEnterBuildingAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	buildingID, ok := params["building_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid building_id parameter")
	}

	// Check if building is open
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	isOpen, openMinutes, closeMinutes, err := utils.IsBuildingOpen(database, state.Location, buildingID, state.TimeOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to check building hours: %v", err)
	}

	if !isOpen {
		// Convert minutes to hours:minutes format for display
		openHour := openMinutes / 60
		openMin := openMinutes % 60
		closeHour := closeMinutes / 60
		closeMin := closeMinutes % 60

		// Format times with AM/PM
		formatTime := func(hour, min int) string {
			period := "AM"
			displayHour := hour
			if hour >= 12 {
				period = "PM"
				if hour > 12 {
					displayHour = hour - 12
				}
			}
			if displayHour == 0 {
				displayHour = 12
			}
			return fmt.Sprintf("%d:%02d %s", displayHour, min, period)
		}

		return &GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("The building is closed. Open hours: %s - %s", formatTime(openHour, openMin), formatTime(closeHour, closeMin)),
			Color:   "red",
		}, nil
	}

	// Update state to include building
	state.Building = buildingID

	log.Printf("🏛️ Entered building: %s", buildingID)

	return &GameActionResponse{
		Success: true,
		Message: "Entered building",
		Color:   "blue",
	}, nil
}

// handleExitBuildingAction exits a building
func handleExitBuildingAction(state *SaveFile, _ map[string]any) (*GameActionResponse, error) {
	// Update state to remove building (back outdoors)
	state.Building = ""

	log.Printf("🚪 Exited building")

	// Check fatigue level to warn user
	message := "Exited building"
	if state.Fatigue > 0 {
		message = fmt.Sprintf("Exited building (Fatigue: %d)", state.Fatigue)
	}

	return &GameActionResponse{
		Success: true,
		Message: message,
		Color:   "blue",
	}, nil
}

// handleTalkToNPCAction initiates dialogue with an NPC
func handleTalkToNPCAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	npcID, ok := params["npc_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid npc_id parameter")
	}

	// Get NPC data from database
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM npcs WHERE id = ?", npcID).Scan(&propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("NPC not found: %s", npcID)
	}

	// Parse NPC properties into NPCData struct
	var npcData types.NPCData
	if err := json.Unmarshal([]byte(propertiesJSON), &npcData); err != nil {
		return nil, fmt.Errorf("failed to parse NPC data: %v", err)
	}

	// Resolve NPC schedule based on current time
	scheduleInfo := utils.ResolveNPCSchedule(&npcData, state.TimeOfDay)

	// Determine location type from location ID
	locationType := utils.DetermineLocationType(scheduleInfo.Location)

	// Check if player is at NPC's current location
	playerAtLocation := false
	if locationType == "building" && state.Building == scheduleInfo.Location {
		playerAtLocation = true
	} else if locationType == "district" && state.Building == "" {
		// Construct full district ID from location + district (e.g., "kingdom" + "center" = "kingdom-center")
		playerDistrictID := fmt.Sprintf("%s-%s", state.Location, state.District)
		if playerDistrictID == scheduleInfo.Location {
			playerAtLocation = true
		}
	}

	if !playerAtLocation {
		return &GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("%s is not here at this time.", npcData.Name),
			Color:   "yellow",
		}, nil
	}

	// Check if NPC is available for interaction
	if !scheduleInfo.IsAvailable {
		return &GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("%s is busy right now.", npcData.Name),
			Color:   "yellow",
		}, nil
	}

	// Determine greeting based on state
	greetingText := ""
	isRegistered := isVaultRegistered(state, scheduleInfo.Location)
	isNativeRace := isNativeRaceForLocation(state.Race, state.Location)

	if isNativeRace {
		greetingText, _ = npcData.Greeting["native_race"]
	} else if isRegistered {
		greetingText, _ = npcData.Greeting["returning"]
	} else {
		greetingText, _ = npcData.Greeting["first_time"]
	}

	// Get initial dialogue node (first from available options)
	var dialogueNode string
	var dialogueText string
	var options []string

	if len(scheduleInfo.AvailableDialogue) > 0 {
		dialogueNode = scheduleInfo.AvailableDialogue[0]

		// Get dialogue content
		if nodeData, ok := npcData.Dialogue[dialogueNode].(map[string]any); ok {
			dialogueText, _ = nodeData["text"].(string)
			if opts, ok := nodeData["options"].([]any); ok {
				for _, opt := range opts {
					if optStr, ok := opt.(string); ok {
						// Filter options based on requirements
						optionNode, _ := npcData.Dialogue[optStr].(map[string]any)
						if optionNode != nil {
							requirements, _ := optionNode["requirements"].(map[string]any)
							if checkDialogueRequirements(state, requirements) {
								options = append(options, optStr)
							}
						} else {
							options = append(options, optStr)
						}
					}
				}
			}
		}
	} else {
		return &GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("%s is busy right now.", npcData.Name),
			Color:   "yellow",
		}, nil
	}

	log.Printf("💬 %s: %s (schedule state: %s, showing %d options)", npcID, greetingText, scheduleInfo.State, len(options))

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("%s\n\n%s", greetingText, dialogueText),
		Color:   "yellow",
		Delta: map[string]any{
			"npc_dialogue": map[string]any{
				"npc_id":         npcID,
				"node":           dialogueNode,
				"text":           dialogueText,
				"options":        options,
				"schedule_state": scheduleInfo.State,
			},
		},
	}, nil
}

// handleNPCDialogueChoiceAction processes player's dialogue choice
func handleNPCDialogueChoiceAction(session *GameSession, params map[string]any) (*GameActionResponse, error) {
	state := &session.SaveData
	npcID, _ := params["npc_id"].(string)
	choice, _ := params["choice"].(string)

	// Get NPC data from database
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM npcs WHERE id = ?", npcID).Scan(&propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("NPC not found: %s", npcID)
	}

	// Parse NPC properties
	var npcData map[string]any
	if err := json.Unmarshal([]byte(propertiesJSON), &npcData); err != nil {
		return nil, fmt.Errorf("failed to parse NPC data: %v", err)
	}

	dialogue, _ := npcData["dialogue"].(map[string]any)
	choiceNode, _ := dialogue[choice].(map[string]any)

	if choiceNode == nil {
		return nil, fmt.Errorf("invalid dialogue choice: %s", choice)
	}

	// Check requirements
	requirements, _ := choiceNode["requirements"].(map[string]any)
	if !checkDialogueRequirements(state, requirements) {
		return &GameActionResponse{
			Success: false,
			Message: "Requirements not met for this dialogue option",
			Color:   "red",
		}, nil
	}

	// Get action if any
	action, _ := choiceNode["action"].(string)
	responseText, _ := choiceNode["text"].(string)

	// Process action
	var actionResult string
	switch action {
	case "register_storage":
		cost, _ := choiceNode["cost"].(float64)
		goldAmount := getGoldQuantity(state)
		if goldAmount >= int(cost) {
			if deductGold(state, int(cost)) {
				registerVault(state, state.Building)
				actionResult, _ = choiceNode["success"].(string)
			} else {
				log.Printf("⚠️ Failed to deduct gold even though we had enough")
				actionResult, _ = choiceNode["failure"].(string)
			}
		} else {
			actionResult, _ = choiceNode["failure"].(string)
		}

	case "rent_room":
		cost, _ := choiceNode["cost"].(float64)
		result, err := handleRentRoomAction(session, map[string]any{"cost": cost})
		if err != nil {
			log.Printf("❌ Failed to rent room: %v", err)
			actionResult, _ = choiceNode["failure"].(string)
		} else if result.Success {
			actionResult, _ = choiceNode["success"].(string)
		} else {
			// Failed but not an error (e.g., not enough gold, already rented)
			actionResult = result.Message
		}

	case "open_storage":
		// Return vault data
		vault := getVaultForLocation(state, state.Building)
		if vault == nil {
			log.Printf("❌ Vault not found for building: %s (Location: %s)", state.Building, state.Location)
			log.Printf("📦 Available vaults: %+v", state.Vaults)
			return &GameActionResponse{
				Success: false,
				Message: "Vault not found for this building",
				Color:   "error",
			}, nil
		}
		log.Printf("✅ Opening vault for building: %s", state.Building)
		return &GameActionResponse{
			Success: true,
			Message: responseText,
			Color:   "yellow",
			Delta: map[string]any{
				"open_vault": vault,
				"npc_dialogue": map[string]any{
					"action": "close",
				},
			},
		}, nil

	case "open_shop":
		// Return signal to open shop UI
		log.Printf("✅ Opening shop for merchant: %s", npcID)
		return &GameActionResponse{
			Success: true,
			Message: responseText,
			Color:   "yellow",
			Delta: map[string]any{
				"open_shop": npcID,
				"npc_dialogue": map[string]any{
					"action": "close",
				},
			},
		}, nil

	case "open_sell":
		// Return signal to open shop UI in sell mode
		log.Printf("✅ Opening shop (sell mode) for merchant: %s", npcID)
		return &GameActionResponse{
			Success: true,
			Message: responseText,
			Color:   "yellow",
			Delta: map[string]any{
				"open_shop": npcID,
				"shop_tab":  "sell",
				"npc_dialogue": map[string]any{
					"action": "close",
				},
			},
		}, nil

	case "book_show":
		// Load show configuration from NPC data
		log.Printf("✅ Getting available shows for NPC: %s", npcID)

		// Get current building/venue for tracking
		venueID := state.Building
		if venueID == "" {
			return &GameActionResponse{
				Success: false,
				Message: "Not in a building",
				Color:   "red",
			}, nil
		}

		// Get show configuration from NPC
		showConfig, ok := npcData["show_config"].(map[string]any)
		if !ok {
			log.Printf("❌ NPC %s does not have show_config", npcID)
			return &GameActionResponse{
				Success: false,
				Message: "This NPC doesn't offer show bookings",
				Color:   "red",
			}, nil
		}

		// Get day of week
		dayOfWeek := utils.GetDayOfWeek(state.CurrentDay)
		dayStr := fmt.Sprintf("%d", dayOfWeek)

		// Get available shows for this day
		availableShows, ok := showConfig["shows_by_day"].(map[string]any)[dayStr].([]any)
		if !ok || len(availableShows) == 0 {
			return &GameActionResponse{
				Success: false,
				Message: "No shows available today",
				Color:   "yellow",
				Delta: map[string]any{
					"npc_dialogue": map[string]any{
						"action": "close",
					},
				},
			}, nil
		}

		// Check booking deadline
		bookingDeadline := int(showConfig["booking_deadline"].(float64))
		showTime := int(showConfig["show_time"].(float64))

		if state.TimeOfDay >= bookingDeadline && state.TimeOfDay < showTime {
			return &GameActionResponse{
				Success: false,
				Message: fmt.Sprintf("Too late to book! Booking closes at %d:%02d.", bookingDeadline/60, bookingDeadline%60),
				Color:   "red",
				Delta: map[string]any{
					"npc_dialogue": map[string]any{
						"action": "close",
					},
				},
			}, nil
		}

		// Check if already booked a show for today (session-only data)
		if session.BookedShows != nil {
			for _, booking := range session.BookedShows {
				if day, ok := booking["day"].(int); ok && day == state.CurrentDay {
					return &GameActionResponse{
						Success: false,
						Message: "You've already booked a show for today",
						Color:   "yellow",
						Delta: map[string]any{
							"npc_dialogue": map[string]any{
								"action": "close",
							},
						},
					}, nil
				}
			}
		}

		// Check instrument availability for each show (but include all shows)
		var allShows []map[string]any
		hasAnyBookableShow := false
		for _, show := range availableShows {
			if showMap, ok := show.(map[string]any); ok {
				requiredInstruments, _ := showMap["required_instruments"].([]any)
				hasAllInstruments := true
				var missingInstruments []string
				for _, instrRaw := range requiredInstruments {
					instrument, _ := instrRaw.(string)
					if !playerHasItem(state, instrument) {
						hasAllInstruments = false
						missingInstruments = append(missingInstruments, instrument)
					}
				}

				// Add instrument availability info to the show
				showWithAvailability := make(map[string]any)
				for k, v := range showMap {
					showWithAvailability[k] = v
				}
				showWithAvailability["can_book"] = hasAllInstruments
				showWithAvailability["missing_instruments"] = missingInstruments

				allShows = append(allShows, showWithAvailability)

				if hasAllInstruments {
					hasAnyBookableShow = true
				}
			}
		}

		// Modify message if player can't book any shows
		displayMessage := responseText
		if !hasAnyBookableShow {
			displayMessage = "Here are tonight's available shows. You'll need the right instruments to book a performance."
		}

		// Return show selection UI with all shows
		return &GameActionResponse{
			Success: true,
			Message: displayMessage,
			Color:   "yellow",
			Delta: map[string]any{
				"show_booking": map[string]any{
					"npc_id":         npcID,
					"venue_id":       venueID,
					"available_shows": allShows,
					"show_time":      showTime,
				},
				"npc_dialogue": map[string]any{
					"action": "close",
				},
			},
		}, nil

	case "end_dialogue":
		return &GameActionResponse{
			Success: true,
			Message: responseText,
			Color:   "yellow",
			Delta: map[string]any{
				"npc_dialogue": map[string]any{
					"action": "close",
				},
			},
		}, nil
	}

	// Get next options and filter based on requirements
	options, _ := choiceNode["options"].([]any)
	var optionsList []string
	for _, opt := range options {
		if optStr, ok := opt.(string); ok {
			// Get the option node to check requirements
			optionNode, _ := dialogue[optStr].(map[string]any)
			if optionNode != nil {
				requirements, _ := optionNode["requirements"].(map[string]any)
				// Only include option if requirements are met
				if checkDialogueRequirements(state, requirements) {
					optionsList = append(optionsList, optStr)
				} else {
					log.Printf("🚫 Filtered out option '%s' (requirements not met)", optStr)
				}
			} else {
				// No requirements, include it
				optionsList = append(optionsList, optStr)
			}
		}
	}

	displayText := responseText
	if actionResult != "" {
		displayText = actionResult
	}

	return &GameActionResponse{
		Success: true,
		Message: displayText,
		Color:   "yellow",
		Delta: map[string]any{
			"npc_dialogue": map[string]any{
				"npc_id":  npcID,
				"node":    choice,
				"text":    displayText,
				"options": optionsList,
			},
		},
	}, nil
}

// Helper: Check if vault is registered at location
func isVaultRegistered(state *SaveFile, buildingID string) bool {
	if state.Vaults == nil {
		return false
	}
	for _, vault := range state.Vaults {
		// Check new format (building field)
		if building, ok := vault["building"].(string); ok {
			if building == buildingID {
				return true
			}
		} else if location, ok := vault["location"].(string); ok {
			// Check old format (location field) - match if we're at that location
			if location == state.Location {
				return true
			}
		}
	}
	return false
}

// Helper: Check if race is native to location
func isNativeRaceForLocation(race, location string) bool {
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

// Helper: Get total gold quantity from inventory
func getGoldQuantity(state *SaveFile) int {
	totalGold := 0

	// Check general slots
	if generalSlots, ok := state.Inventory["general_slots"].([]any); ok {
		for _, slotData := range generalSlots {
			if slotMap, ok := slotData.(map[string]any); ok {
				if itemID, ok := slotMap["item"].(string); ok && itemID == "gold-piece" {
					// Handle both int and float64 types
					switch v := slotMap["quantity"].(type) {
					case float64:
						totalGold += int(v)
					case int:
						totalGold += v
					}
				}
			}
		}
	}

	// Check backpack
	if gearSlots, ok := state.Inventory["gear_slots"].(map[string]any); ok {
		if bag, ok := gearSlots["bag"].(map[string]any); ok {
			if contents, ok := bag["contents"].([]any); ok {
				for _, slotData := range contents {
					if slotMap, ok := slotData.(map[string]any); ok {
						if itemID, ok := slotMap["item"].(string); ok && itemID == "gold-piece" {
							// Handle both int and float64 types
							switch v := slotMap["quantity"].(type) {
							case float64:
								totalGold += int(v)
							case int:
								totalGold += v
							}
						}
					}
				}
			}
		}
	}

	return totalGold
}

// Helper: Deduct gold from inventory (returns true if successful)
func deductGold(state *SaveFile, amount int) bool {
	if amount <= 0 {
		return true
	}

	remaining := amount

	// First, deduct from general slots
	if generalSlots, ok := state.Inventory["general_slots"].([]any); ok {
		for _, slotData := range generalSlots {
			if remaining <= 0 {
				break
			}
			if slotMap, ok := slotData.(map[string]any); ok {
				if itemID, ok := slotMap["item"].(string); ok && itemID == "gold-piece" {
					// Handle both int and float64 types
					var currentQty int
					switch v := slotMap["quantity"].(type) {
					case float64:
						currentQty = int(v)
					case int:
						currentQty = v
					default:
						continue
					}

					if currentQty >= remaining {
						// This slot has enough gold (store as int)
						slotMap["quantity"] = currentQty - remaining
						if currentQty == remaining {
							// Clear slot if depleted
							slotMap["item"] = nil
							slotMap["quantity"] = 0
						}
						remaining = 0
					} else {
						// Take all gold from this slot
						remaining -= currentQty
						slotMap["item"] = nil
						slotMap["quantity"] = 0
					}
				}
			}
		}
	}

	// Then, deduct from backpack if needed
	if remaining > 0 {
		if gearSlots, ok := state.Inventory["gear_slots"].(map[string]any); ok {
			if bag, ok := gearSlots["bag"].(map[string]any); ok {
				if contents, ok := bag["contents"].([]any); ok {
					for _, slotData := range contents {
						if remaining <= 0 {
							break
						}
						if slotMap, ok := slotData.(map[string]any); ok {
							if itemID, ok := slotMap["item"].(string); ok && itemID == "gold-piece" {
								// Handle both int and float64 types
								var currentQty int
								switch v := slotMap["quantity"].(type) {
								case float64:
									currentQty = int(v)
								case int:
									currentQty = v
								default:
									continue
								}

								if currentQty >= remaining {
									// This slot has enough gold (store as int)
									slotMap["quantity"] = currentQty - remaining
									if currentQty == remaining {
										slotMap["item"] = nil
										slotMap["quantity"] = 0
									}
									remaining = 0
								} else {
									remaining -= currentQty
									slotMap["item"] = nil
									slotMap["quantity"] = 0
								}
							}
						}
					}
				}
			}
		}
	}

	return remaining == 0
}

// Helper: Check dialogue requirements
func checkDialogueRequirements(state *SaveFile, requirements map[string]any) bool {
	if requirements == nil {
		return true
	}

	if notNative, ok := requirements["not_native"].(bool); ok && notNative {
		if isNativeRaceForLocation(state.Race, state.Location) {
			return false
		}
	}

	if notRegistered, ok := requirements["not_registered"].(bool); ok && notRegistered {
		if isVaultRegistered(state, state.Building) {
			return false
		}
	}

	if registered, ok := requirements["registered"].(bool); ok && registered {
		if !isVaultRegistered(state, state.Building) {
			return false
		}
	}

	if goldReq, ok := requirements["gold"].(float64); ok {
		if getGoldQuantity(state) < int(goldReq) {
			return false
		}
	}

	return true
}

// Helper: Register vault at location
func registerVault(state *SaveFile, buildingID string) {
	if state.Vaults == nil {
		state.Vaults = []map[string]any{}
	}

	// Check if already registered
	for _, vault := range state.Vaults {
		if building, ok := vault["building"].(string); ok && building == buildingID {
			return // Already registered
		}
	}

	// Create new vault with 40 empty slots
	slots := make([]map[string]any, 40)
	for i := range 40 {
		slots[i] = map[string]any{
			"slot":     i,
			"item":     nil,
			"quantity": 0,
		}
	}

	vault := map[string]any{
		"building": buildingID,
		"slots":    slots,
	}

	state.Vaults = append(state.Vaults, vault)
	log.Printf("✅ Registered vault at %s", buildingID)
}

// Helper: Get vault for location
func getVaultForLocation(state *SaveFile, buildingID string) map[string]any {
	if state.Vaults == nil {
		return nil
	}

	for _, vault := range state.Vaults {
		// Check new format (building field)
		if building, ok := vault["building"].(string); ok && building == buildingID {
			return vault
		}
		// Check old format (location field) - return if we're at that location
		if location, ok := vault["location"].(string); ok && location == state.Location {
			return vault
		}
	}

	return nil
}

// handleRegisterVaultAction registers a vault (called after payment)
func handleRegisterVaultAction(state *SaveFile, _ map[string]any) (*GameActionResponse, error) {
	buildingID := state.Building
	if buildingID == "" {
		return nil, fmt.Errorf("not in a building")
	}

	registerVault(state, buildingID)

	return &GameActionResponse{
		Success: true,
		Message: "Vault registered successfully",
		Color:   "green",
	}, nil
}

// handleOpenVaultAction returns vault data for UI
func handleOpenVaultAction(state *SaveFile, _ map[string]any) (*GameActionResponse, error) {
	buildingID := state.Building
	if buildingID == "" {
		return nil, fmt.Errorf("not in a building")
	}

	vault := getVaultForLocation(state, buildingID)
	if vault == nil {
		return nil, fmt.Errorf("no vault registered at this location")
	}

	return &GameActionResponse{
		Success: true,
		Message: "Vault opened",
		Delta: map[string]any{
			"vault": vault,
		},
	}, nil
}

// handleAddToContainerAction adds an item to a container
func handleAddToContainerAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	// Convert params to ItemActionRequest
	itemID, _ := params["item_id"].(string)
	fromSlot := -1
	if fs, ok := params["from_slot"].(float64); ok {
		fromSlot = int(fs)
	}
	fromSlotType, _ := params["from_slot_type"].(string)
	containerSlot := -1
	if cs, ok := params["container_slot"].(float64); ok {
		containerSlot = int(cs)
	}
	containerSlotType, _ := params["container_slot_type"].(string)
	toContainerSlot := -1
	if tcs, ok := params["to_container_slot"].(float64); ok {
		toContainerSlot = int(tcs)
	}

	req := &types.ItemActionRequest{
		ItemID:          itemID,
		Action:          "add_to_container",
		FromSlot:        fromSlot,
		FromSlotType:    fromSlotType,
		ContainerSlot:   containerSlot,
		ToSlotType:      containerSlotType, // Use ToSlotType for container location
		ToContainerSlot: toContainerSlot,
	}

	// Call the inventory handler
	response := handleAddToContainer(state, req)

	// Convert ItemActionResponse to GameActionResponse
	if response.Success {
		return &GameActionResponse{
			Success: true,
			Message: response.Message,
			Color:   response.Color,
			State:   state, // Return the updated state
		}, nil
	}

	return nil, fmt.Errorf("%s", response.Error)
}

// handleRemoveFromContainerAction removes an item from a container
func handleRemoveFromContainerAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	// Convert params to ItemActionRequest
	itemID, _ := params["item_id"].(string)
	containerSlot := -1
	if cs, ok := params["container_slot"].(float64); ok {
		containerSlot = int(cs)
	}
	fromContainerSlot := -1
	if fcs, ok := params["from_container_slot"].(float64); ok {
		fromContainerSlot = int(fcs)
	}
	containerSlotType, _ := params["container_slot_type"].(string)

	req := &types.ItemActionRequest{
		ItemID:        itemID,
		Action:        "remove_from_container",
		ContainerSlot: containerSlot,
		FromSlot:      fromContainerSlot, // Use FromSlot for the container slot index
		FromSlotType:  containerSlotType, // Pass the container location type
	}

	// Call the inventory handler
	response := handleRemoveFromContainer(state, req)

	// Convert ItemActionResponse to GameActionResponse
	if response.Success {
		return &GameActionResponse{
			Success: true,
			Message: response.Message,
			Color:   response.Color,
			State:   state, // Return the updated state
		}, nil
	}

	return nil, fmt.Errorf("%s", response.Error)
}

// ============================================================================
// TAVERN ACTIONS
// ============================================================================

// handleRentRoomAction rents a room at an inn/tavern
func handleRentRoomAction(session *GameSession, params map[string]any) (*GameActionResponse, error) {
	state := &session.SaveData
	buildingID := state.Building
	if buildingID == "" {
		return nil, fmt.Errorf("not in a building")
	}

	// Get cost from params (from NPC dialogue config)
	cost := 50 // Default cost
	if c, ok := params["cost"].(float64); ok {
		cost = int(c)
	}

	// Check if player has enough gold
	goldAmount := getGoldQuantity(state)
	log.Printf("🪙 Player has %d gold, room costs %d gold", goldAmount, cost)
	if goldAmount < cost {
		return &GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("You need %d gold to rent a room. You have %d gold.", cost, goldAmount),
			Color:   "red",
		}, nil
	}

	// Deduct gold
	log.Printf("💰 Attempting to deduct %d gold...", cost)
	if !deductGold(state, cost) {
		log.Printf("❌ Failed to deduct gold!")
		return &GameActionResponse{
			Success: false,
			Message: "Failed to deduct gold for room rental",
			Color:   "red",
		}, nil
	}

	// Verify gold was deducted
	newGoldAmount := getGoldQuantity(state)
	log.Printf("✅ Gold deducted successfully. Old: %d, New: %d", goldAmount, newGoldAmount)

	// Initialize rented rooms if needed (session-only data)
	if session.RentedRooms == nil {
		session.RentedRooms = []map[string]any{}
	}

	// Check if already rented at this building
	for _, room := range session.RentedRooms {
		if building, ok := room["building"].(string); ok && building == buildingID {
			return &GameActionResponse{
				Success: false,
				Message: "You already have a room rented here",
				Color:   "yellow",
			}, nil
		}
	}

	// Add rented room (expires at end of next day - 23:59)
	expirationDay := state.CurrentDay + 1
	expirationTime := 1439 // 23:59

	room := map[string]any{
		"building":        buildingID,
		"expiration_day":  expirationDay,
		"expiration_time": expirationTime,
	}

	session.RentedRooms = append(session.RentedRooms, room)

	log.Printf("🏠 Rented room at %s for %d gold (expires day %d at %d)", buildingID, cost, expirationDay, expirationTime)

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Rented a room for %d gold. You can sleep here until tomorrow night.", cost),
		Color:   "green",
		Data: map[string]interface{}{
			"rented_rooms": session.RentedRooms,
		},
	}, nil
}

// handleSleepAction sleeps in a rented room
func handleSleepAction(session *GameSession, _ map[string]any) (*GameActionResponse, error) {
	state := &session.SaveData
	buildingID := state.Building
	if buildingID == "" {
		return nil, fmt.Errorf("not in a building")
	}

	// Check if player has a rented room here and find the index (session-only data)
	hasRoom := false
	roomIndex := -1
	if session.RentedRooms != nil {
		for i, room := range session.RentedRooms {
			if building, ok := room["building"].(string); ok && building == buildingID {
				// Check if expired
				expDay, _ := room["expiration_day"].(int)
				expTime, _ := room["expiration_time"].(int)

				if state.CurrentDay > expDay || (state.CurrentDay == expDay && state.TimeOfDay > expTime) {
					// Room expired, remove it
					session.RentedRooms = append(session.RentedRooms[:i], session.RentedRooms[i+1:]...)
					return &GameActionResponse{
						Success: false,
						Message: "Your room rental has expired. Please rent another room.",
						Color:   "yellow",
					}, nil
				}

				hasRoom = true
				roomIndex = i
				break
			}
		}
	}

	if !hasRoom {
		return &GameActionResponse{
			Success: false,
			Message: "You don't have a room rented here",
			Color:   "red",
		}, nil
	}

	// Calculate sleep quality based on current time (late bedtime = poor sleep)
	// Ideal bedtime: before midnight (0-359 minutes or 1320-1439 minutes)
	// Late bedtime: after midnight (360-720 minutes) - fatigue penalty
	poorSleep := false
	if state.TimeOfDay >= 360 && state.TimeOfDay <= 720 {
		poorSleep = true
	}

	// Calculate how many minutes will be slept
	oldTime := state.TimeOfDay
	targetTime := 360 // 6 AM
	var minutesSlept int
	if oldTime >= targetTime {
		// Already past 6 AM, sleep until 6 AM next day (e.g., 10 PM = 1320 mins, sleep 8h40m)
		minutesSlept = (1440 - oldTime) + targetTime
		state.CurrentDay++
	} else {
		// Before 6 AM, sleep until 6 AM same day
		minutesSlept = targetTime - oldTime
	}
	state.TimeOfDay = targetTime

	// Tick down duration-based effects for the time slept
	// This handles buffs/debuffs like performance-high that expire over time
	tickDownEffectDurations(state, minutesSlept)

	// Reset fatigue based on sleep quality
	if poorSleep {
		state.Fatigue = 1 // Poor sleep - still a bit tired
		log.Printf("😴 Poor sleep due to late bedtime (fatigue level 1)")
	} else {
		state.Fatigue = 0 // Good sleep - fully rested
		log.Printf("😴 Good sleep (fully rested)")
	}
	resetFatigueAccumulator(state)
	_, _ = updateFatiguePenaltyEffects(state)

	// Reset hunger (well fed after waking up)
	state.Hunger = 2
	resetHungerAccumulator(state)
	_, _ = updateHungerPenaltyEffects(state)
	ensureHungerAccumulation(state)

	// Restore HP and Mana fully
	state.HP = state.MaxHP
	state.Mana = state.MaxMana

	// Remove the rented room after sleeping (room is used up)
	if roomIndex >= 0 && roomIndex < len(session.RentedRooms) {
		session.RentedRooms = append(session.RentedRooms[:roomIndex], session.RentedRooms[roomIndex+1:]...)
		log.Printf("🚪 Room rental at %s has been used and removed", buildingID)
	}

	sleepMessage := "You wake up refreshed at 6 AM."
	if poorSleep {
		sleepMessage = "You wake up at 6 AM, but didn't sleep well due to going to bed late."
	}

	// Update building states and NPCs after sleep (time jump)
	database := db.GetDB()
	if database != nil {
		newTime := state.TimeOfDay
		currentHour := newTime / 60

		// Refresh building states
		buildingStates, err := utils.GetAllBuildingStatesForDistrict(
			database,
			state.Location,
			state.District,
			newTime,
		)
		if err == nil && len(buildingStates) > 0 {
			session.UpdateBuildingStates(buildingStates, newTime)
		}

		// Refresh NPCs
		npcIDs := GetNPCIDsAtLocation(
			state.Location,
			state.District,
			state.Building,
			newTime,
		)
		session.UpdateNPCsAtLocation(npcIDs, currentHour)
	}

	// Calculate delta for frontend updates
	delta := session.UpdateSnapshotAndCalculateDelta()

	return &GameActionResponse{
		Success: true,
		Message: sleepMessage,
		Color:   "green",
		Delta:   delta.ToMap(),
		Data: map[string]interface{}{
			"time_of_day":   state.TimeOfDay,
			"current_day":   state.CurrentDay,
			"fatigue":       state.Fatigue,
			"hunger":        state.Hunger,
			"hp":            state.HP,
			"max_hp":        state.MaxHP,
			"mana":          state.Mana,
			"max_mana":      state.MaxMana,
			"rented_rooms":  session.RentedRooms, // Send updated rooms so frontend knows room was used
		},
	}, nil
}

// handleWaitAction waits for a specified amount of time
// Accepts either "minutes" (15-360) or "hours" (1-6) for backwards compatibility
func handleWaitAction(session *GameSession, params map[string]any) (*GameActionResponse, error) {
	state := &session.SaveData

	var minutesToAdvance int

	// Check for minutes first (more granular), fall back to hours
	if minutesFloat, ok := params["minutes"].(float64); ok {
		minutesToAdvance = int(minutesFloat)
		// Validate minutes (15-360 in 15 minute increments, 6 hours max)
		if minutesToAdvance < 15 || minutesToAdvance > 360 {
			return &GameActionResponse{
				Success: false,
				Message: "You can only wait between 15 minutes and 6 hours",
				Color:   "red",
			}, nil
		}
	} else if hoursFloat, ok := params["hours"].(float64); ok {
		hours := int(hoursFloat)
		// Validate hours (1-6 hours max)
		if hours < 1 || hours > 6 {
			return &GameActionResponse{
				Success: false,
				Message: "You can only wait between 1 and 6 hours",
				Color:   "red",
			}, nil
		}
		minutesToAdvance = hours * 60
	} else {
		return nil, fmt.Errorf("hours or minutes parameter is required")
	}

	// Track fatigue/hunger before wait
	oldFatigue := state.Fatigue
	oldHunger := state.Hunger

	// Advance time and process all effects (effects system handles fatigue/hunger)
	timeMessages := advanceTime(state, minutesToAdvance)

	// Update building states and NPCs after time jump
	database := db.GetDB()
	if database != nil {
		newTime := state.TimeOfDay
		currentHour := newTime / 60

		// Refresh building states
		buildingStates, err := utils.GetAllBuildingStatesForDistrict(
			database,
			state.Location,
			state.District,
			newTime,
		)
		if err == nil && len(buildingStates) > 0 {
			session.UpdateBuildingStates(buildingStates, newTime)
		}

		// Refresh NPCs
		npcIDs := GetNPCIDsAtLocation(
			state.Location,
			state.District,
			state.Building,
			newTime,
		)
		session.UpdateNPCsAtLocation(npcIDs, currentHour)
	}

	// Format message based on wait duration
	var message string
	hours := minutesToAdvance / 60
	mins := minutesToAdvance % 60
	if hours > 0 && mins > 0 {
		message = fmt.Sprintf("You waited %d hour%s and %d minute%s.",
			hours, pluralize(hours), mins, pluralize(mins))
	} else if hours > 0 {
		message = fmt.Sprintf("You waited %d hour%s.", hours, pluralize(hours))
	} else {
		message = fmt.Sprintf("You waited %d minute%s.", mins, pluralize(mins))
	}

	// Add explicit fatigue/hunger change messages
	if state.Fatigue != oldFatigue {
		fatigueChange := state.Fatigue - oldFatigue
		if fatigueChange > 0 {
			message += fmt.Sprintf("\n\n💤 Your fatigue increased by %d (now %d/10)", fatigueChange, state.Fatigue)
		} else {
			message += fmt.Sprintf("\n\n✨ Your fatigue decreased by %d (now %d/10)", -fatigueChange, state.Fatigue)
		}
	}

	if state.Hunger != oldHunger {
		hungerChange := state.Hunger - oldHunger
		hungerNames := map[int]string{0: "Famished", 1: "Hungry", 2: "Satisfied", 3: "Full"}
		if hungerChange < 0 {
			message += fmt.Sprintf("\n\n🍽️ You're feeling hungrier (now %s)", hungerNames[state.Hunger])
		}
	}

	// Append any time-based messages (like starvation damage)
	if len(timeMessages) > 0 {
		for _, msg := range timeMessages {
			if !msg.Silent {
				message += "\n\n" + msg.Message
			}
		}
	}

	log.Printf("⏱️ Waited %d minutes - Time: %d, Fatigue: %d→%d, Hunger: %d→%d", minutesToAdvance, state.TimeOfDay, oldFatigue, state.Fatigue, oldHunger, state.Hunger)

	// Calculate delta for UI updates
	delta := session.UpdateSnapshotAndCalculateDelta()

	return &GameActionResponse{
		Success: true,
		Message: message,
		Color:   "yellow",
		Delta:   delta.ToMap(),
		Data: map[string]interface{}{
			"time_of_day": state.TimeOfDay,
			"current_day": state.CurrentDay,
			"fatigue":     state.Fatigue,
			"hunger":      state.Hunger,
			"hp":          state.HP,
		},
	}, nil
}

// handleResetIdleTimerAction resets the auto-pause idle timer
// Called when the play button is pressed to prevent immediate re-triggering of auto-pause
func handleResetIdleTimerAction(session *GameSession) (*GameActionResponse, error) {
	// Reset the idle tracking to current time
	session.LastActionTime = time.Now().Unix()
	session.LastActionGameTime = session.SaveData.TimeOfDay

	log.Printf("⏱️ Idle timer reset - LastActionGameTime: %d", session.LastActionGameTime)

	return &GameActionResponse{
		Success: true,
		Message: "Idle timer reset",
	}, nil
}

// handleBookShowAction books a performance at a tavern
func handleBookShowAction(session *GameSession, params map[string]any) (*GameActionResponse, error) {
	state := &session.SaveData
	showID, ok := params["show_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing show_id parameter")
	}

	npcID, ok := params["npc_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing npc_id parameter")
	}

	venueID := state.Building
	if venueID == "" {
		return nil, fmt.Errorf("not in a building")
	}

	// Get NPC data from database
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM npcs WHERE id = ?", npcID).Scan(&propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("NPC not found: %s", npcID)
	}

	// Parse NPC properties
	var npcData map[string]any
	if err := json.Unmarshal([]byte(propertiesJSON), &npcData); err != nil {
		return nil, fmt.Errorf("failed to parse NPC data: %v", err)
	}

	// Load show configuration from NPC data
	showConfig, ok := npcData["show_config"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("NPC %s does not have show_config", npcID)
	}

	// Get day of week
	dayOfWeek := utils.GetDayOfWeek(state.CurrentDay)
	dayStr := fmt.Sprintf("%d", dayOfWeek)

	// Get available shows for this day
	availableShows, ok := showConfig["shows_by_day"].(map[string]any)[dayStr].([]any)
	if !ok || len(availableShows) == 0 {
		return &GameActionResponse{
			Success: false,
			Message: "No shows available today",
			Color:   "yellow",
		}, nil
	}

	// Find the requested show
	var selectedShow map[string]any
	for _, show := range availableShows {
		if showMap, ok := show.(map[string]any); ok {
			if id, ok := showMap["id"].(string); ok && id == showID {
				selectedShow = showMap
				break
			}
		}
	}

	if selectedShow == nil {
		return &GameActionResponse{
			Success: false,
			Message: "Show not available today",
			Color:   "yellow",
		}, nil
	}

	// Check booking deadline (must book before specified time, e.g., 8 PM = 1200 minutes)
	showTime := int(showConfig["show_time"].(float64))
	bookingDeadline := int(showConfig["booking_deadline"].(float64))

	// Check if current time is past the booking deadline
	if state.TimeOfDay >= bookingDeadline && state.TimeOfDay < showTime {
		return &GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("Too late to book! Booking closes at %d:%02d.", bookingDeadline/60, bookingDeadline%60),
			Color:   "red",
		}, nil
	}

	// Check if player has required instruments
	requiredInstruments, _ := selectedShow["required_instruments"].([]any)
	for _, instrRaw := range requiredInstruments {
		instrument, _ := instrRaw.(string)
		if !playerHasItem(state, instrument) {
			return &GameActionResponse{
				Success: false,
				Message: fmt.Sprintf("You need a %s to perform this show", instrument),
				Color:   "red",
			}, nil
		}
	}

	// Check if already booked a show for today (session-only data)
	if session.BookedShows == nil {
		session.BookedShows = []map[string]any{}
	}

	for _, booking := range session.BookedShows {
		if day, ok := booking["day"].(int); ok && day == state.CurrentDay {
			return &GameActionResponse{
				Success: false,
				Message: "You've already booked a show for today",
				Color:   "yellow",
			}, nil
		}
	}

	// Book the show
	booking := map[string]any{
		"show_id":    showID,
		"venue_id":   venueID,
		"day":        state.CurrentDay,
		"show_time":  showTime,
		"performed":  false,
		"show_data":  selectedShow, // Store show data for later
	}

	session.BookedShows = append(session.BookedShows, booking)

	showName, _ := selectedShow["name"].(string)
	log.Printf("🎭 Booked show '%s' (ID: %s) at %s for day %d", showName, showID, venueID, state.CurrentDay)

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Booked '%s' for tonight at 9 PM!", showName),
		Color:   "green",
		Data: map[string]interface{}{
			"booked_shows": session.BookedShows,
		},
	}, nil
}

// handlePlayShowAction performs a booked show
func handlePlayShowAction(session *GameSession, _ map[string]any) (*GameActionResponse, error) {
	state := &session.SaveData
	log.Printf("🎭 handlePlayShowAction called - building: %s, day: %d, time: %d", state.Building, state.CurrentDay, state.TimeOfDay)

	venueID := state.Building
	if venueID == "" {
		return nil, fmt.Errorf("not in a building")
	}

	// Find booked show for today at this venue (session-only data)
	var booking map[string]any
	var bookingIndex int
	if session.BookedShows != nil {
		for i, b := range session.BookedShows {
			day, _ := b["day"].(int)
			venue, _ := b["venue_id"].(string)
			performed, _ := b["performed"].(bool)

			if day == state.CurrentDay && venue == venueID && !performed {
				booking = b
				bookingIndex = i
				break
			}
		}
	}

	if booking == nil {
		return &GameActionResponse{
			Success: false,
			Message: "You don't have a show booked for tonight",
			Color:   "yellow",
		}, nil
	}

	// Check if it's show time (must be within 30 minutes of show time)
	showTime := getIntValue(booking, "show_time", 0)
	timeDiff := state.TimeOfDay - showTime
	if timeDiff < 0 || timeDiff > 30 {
		return &GameActionResponse{
			Success: false,
			Message: "It's not show time yet (show starts at 9 PM)",
			Color:   "yellow",
		}, nil
	}

	// Get show data
	showData, _ := booking["show_data"].(map[string]any)
	if showData == nil {
		return nil, fmt.Errorf("show data not found in booking")
	}

	// Get required instruments and find the first one the player has
	requiredInstruments, _ := showData["required_instruments"].([]any)
	var instrumentID string
	for _, instrRaw := range requiredInstruments {
		instr, _ := instrRaw.(string)
		if playerHasItem(state, instr) {
			instrumentID = instr
			break
		}
	}

	if instrumentID == "" {
		return &GameActionResponse{
			Success: false,
			Message: "You don't have the required instrument!",
			Color:   "red",
		}, nil
	}

	// Load instrument difficulty data
	instrumentData, err := loadInstrumentData(instrumentID)
	if err != nil {
		log.Printf("⚠️ Failed to load instrument data for %s, using defaults: %v", instrumentID, err)
		// Use safe defaults if instrument not found
		instrumentData = map[string]interface{}{
			"base_success":       70.0,
			"charisma_modifier":  5.0,
		}
	}

	// Get charisma stat (with active effect modifiers)
	charisma := 10 // Default
	if stats, ok := state.Stats["Charisma"]; ok {
		if charInt, ok := stats.(int); ok {
			charisma = charInt
		} else if charFloat, ok := stats.(float64); ok {
			charisma = int(charFloat)
		}
	}

	// Apply active charisma modifiers from effects
	statModifiers := getActiveStatModifiers(state)
	charisma += statModifiers["charisma"]

	// Calculate performance success chance
	baseSuccess := getFloatValue(instrumentData, "base_success", 50.0)
	charismaMod := getFloatValue(instrumentData, "charisma_modifier", 5.0)
	successChance := baseSuccess + (float64(charisma-10) * charismaMod)

	// Clamp success chance between 5% and 95%
	if successChance < 5 {
		successChance = 5
	}
	if successChance > 95 {
		successChance = 95
	}

	// Perform charisma check (roll 1-100)
	roll := rand.Intn(100) + 1
	performanceSuccess := float64(roll) <= successChance

	log.Printf("🎲 Performance check: roll=%d, success_threshold=%.0f%% - %v", roll, successChance, performanceSuccess)

	// Calculate rewards
	baseGold := getIntValue(showData, "base_gold", 0)
	baseXP := getIntValue(showData, "base_xp", 0)
	charismaBonus := getIntValue(showData, "charisma_gold_bonus", 0)

	// Calculate total gold (charisma bonus applies to points above 10)
	charismaMod2 := charisma - 10
	if charismaMod2 < 0 {
		charismaMod2 = 0
	}
	totalGold := baseGold + (charismaMod2 * charismaBonus)

	// Add gold to inventory (always get paid)
	if err := addGoldToInventory(state.Inventory, totalGold); err != nil {
		log.Printf("⚠️ Failed to add gold to inventory: %v", err)
	}

	// Only award XP on successful performance
	var resultMessage string
	var resultColor string
	if performanceSuccess {
		state.Experience += baseXP
		// Apply performance-high effect (+2 charisma for 12 hours)
		if err := applyEffect(state, "performance-high"); err != nil {
			log.Printf("⚠️ Failed to apply performance-high effect: %v", err)
			resultMessage = fmt.Sprintf("🎵 Excellent performance! The crowd loved it! Earned %d gold and %d XP!", totalGold, baseXP)
		} else {
			resultMessage = fmt.Sprintf("🎵 Excellent performance! The crowd loved it! Earned %d gold and %d XP. You feel confident! (+2 Charisma for 12 hours)", totalGold, baseXP)
		}
		resultColor = "green"
		log.Printf("✅ Performance success! Earned %d gold, %d XP", totalGold, baseXP)
	} else {
		// Apply stage-fright effect (-1 charisma for 12 hours)
		if err := applyEffect(state, "stage-fright"); err != nil {
			log.Printf("⚠️ Failed to apply stage-fright effect: %v", err)
			resultMessage = fmt.Sprintf("😰 The performance was lackluster. Earned %d gold but no experience.", totalGold)
		} else {
			resultMessage = fmt.Sprintf("😰 The performance was lackluster. Earned %d gold but no experience. You feel shaken. (-1 Charisma for 12 hours)", totalGold)
		}
		resultColor = "yellow"
		log.Printf("❌ Performance failure! Earned %d gold (no XP)", totalGold)
	}

	// Advance time by 60 minutes (1 hour performance)
	oldTime := state.TimeOfDay
	timeMessages := advanceTime(state, 60)
	log.Printf("⏰ Time advanced from %d to %d (60 minutes)", oldTime, state.TimeOfDay)

	// Append any time-based messages (like starvation damage) to result
	if len(timeMessages) > 0 {
		for _, msg := range timeMessages {
			if !msg.Silent {
				resultMessage += "\n\n" + msg.Message
			}
		}
	}

	// Mark show as performed (session-only data)
	session.BookedShows[bookingIndex]["performed"] = true

	// Add to performed shows list for daily tracking (session-only data)
	if session.PerformedShows == nil {
		session.PerformedShows = []string{}
	}
	showID, _ := booking["show_id"].(string)
	performedKey := fmt.Sprintf("%s_%d", showID, state.CurrentDay)
	session.PerformedShows = append(session.PerformedShows, performedKey)

	return &GameActionResponse{
		Success: true,
		Message: resultMessage,
		Color:   resultColor,
	}, nil
}

// checkMissedShows checks for any booked shows that the player missed and applies a penalty
func checkMissedShows(state *SaveFile, session *GameSession, currentTime int, currentDay int) {
	if session.BookedShows == nil {
		return
	}

	for i, booking := range session.BookedShows {
		// Skip already performed or already penalized shows
		performed, _ := booking["performed"].(bool)
		penalized, _ := booking["penalized"].(bool)
		if performed || penalized {
			continue
		}

		bookingDay := getIntValue(booking, "day", -1)
		showTime := getIntValue(booking, "show_time", 0)
		showEndTime := showTime + 60 // 1-hour show window (9-10pm)

		// Check if we've passed the show window
		// Case 1: Same day, past the show end time
		// Case 2: Day has advanced (definitely missed)
		showMissed := false
		if bookingDay == currentDay && currentTime > showEndTime {
			showMissed = true
		} else if currentDay > bookingDay {
			showMissed = true
		}

		if showMissed {
			log.Printf("🎭 Player missed booked show! Day %d, show was at %d, current time is day %d at %d",
				bookingDay, showTime, currentDay, currentTime)

			// Apply no-show penalty effect (-2 charisma for 24 hours)
			if err := applyEffect(state, "no-show"); err != nil {
				log.Printf("⚠️ Failed to apply no-show effect: %v", err)
			} else {
				log.Printf("😔 Applied no-show penalty: -2 Charisma for 24 hours")
			}

			// Mark as penalized so we don't keep applying the penalty
			session.BookedShows[i]["penalized"] = true
		}
	}
}

// loadInstrumentData loads instrument difficulty data from database
func loadInstrumentData(instrumentID string) (map[string]interface{}, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	// Query item properties from database
	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM items WHERE id = ?", instrumentID).Scan(&propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("instrument not found in database: %s", instrumentID)
	}

	// Parse properties JSON
	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
		return nil, fmt.Errorf("failed to parse instrument properties: %v", err)
	}

	// Extract performance data
	performance, ok := properties["performance"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("instrument %s has no performance data", instrumentID)
	}

	return performance, nil
}

// Helper: Check if player has an item in inventory
func playerHasItem(state *SaveFile, itemID string) bool {
	// Check general slots
	if generalSlots, ok := state.Inventory["general_slots"].([]any); ok {
		for _, slotData := range generalSlots {
			if slotMap, ok := slotData.(map[string]any); ok {
				if item, ok := slotMap["item"].(string); ok && item == itemID {
					return true
				}
			}
		}
	}

	// Check backpack
	if gearSlots, ok := state.Inventory["gear_slots"].(map[string]any); ok {
		if bag, ok := gearSlots["bag"].(map[string]any); ok {
			if contents, ok := bag["contents"].([]any); ok {
				for _, slotData := range contents {
					if slotMap, ok := slotData.(map[string]any); ok {
						if item, ok := slotMap["item"].(string); ok && item == itemID {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// ============================================================================
// EFFECTS SYSTEM
// ============================================================================

// EffectMessage contains the message to display when an effect is applied
type EffectMessage struct {
	Message  string
	Color    string
	Category string
	Silent   bool
}

// applyEffect applies an effect to the character (from game-data/effects/{effectID}.json)
func applyEffect(state *SaveFile, effectID string) error {
	_, err := applyEffectWithMessage(state, effectID)
	return err
}

// applyEffectWithMessage applies an effect and returns a message to display
func applyEffectWithMessage(state *SaveFile, effectID string) (*EffectMessage, error) {
	// Load effect data from file
	effectData, err := loadEffectData(effectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load effect %s: %v", effectID, err)
	}

	// Initialize active_effects if nil
	if state.ActiveEffects == nil {
		state.ActiveEffects = []ActiveEffect{}
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
			applyImmediateEffect(state, effectType, int(value))
		} else {
			// Add to active effects for over-time processing or permanent stat modifiers
			activeEffect := ActiveEffect{
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
	effectMsg := &EffectMessage{
		Message:  message,
		Color:    color,
		Category: category,
		Silent:   silent,
	}

	return effectMsg, nil
}

// applyImmediateEffect applies an instant effect (no duration)
func applyImmediateEffect(state *SaveFile, effectType string, value int) {
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
		// Adjust fatigue level and update penalty effects
		state.Fatigue += value
		if state.Fatigue < 0 {
			state.Fatigue = 0
		}
		if state.Fatigue > 10 {
			state.Fatigue = 10
		}

		// Stop accumulation if we've reached max fatigue
		if state.Fatigue >= 10 {
			removeFatigueAccumulation(state)
		} else {
			// Ensure accumulation is active if below max
			ensureFatigueAccumulation(state)
		}

		_, _ = updateFatiguePenaltyEffects(state)
	case "hunger":
		// Adjust hunger level and update penalty effects
		state.Hunger += value
		if state.Hunger < 0 {
			state.Hunger = 0
		}
		if state.Hunger > 3 {
			state.Hunger = 3
		}
		_, _ = updateHungerPenaltyEffects(state)
		ensureHungerAccumulation(state)
	}
}

// Fatigue and Hunger System Initialization

// normalizeEffectID converts old effect IDs to new ones for backward compatibility
func normalizeEffectID(effectID string) string {
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

// migrateOldEffectIDs updates all effect IDs in ActiveEffects to use new naming conventions
func migrateOldEffectIDs(state *SaveFile) {
	if state.ActiveEffects == nil {
		return
	}
	for i := range state.ActiveEffects {
		oldID := state.ActiveEffects[i].EffectID
		newID := normalizeEffectID(oldID)
		if newID != oldID {
			log.Printf("🔄 Migrating effect ID: %s -> %s", oldID, newID)
			state.ActiveEffects[i].EffectID = newID
		}
	}
}

// initializeFatigueHungerEffects ensures all accumulation and penalty effects are properly set
// This should be called when loading a save or after modifying fatigue/hunger values
func initializeFatigueHungerEffects(state *SaveFile) error {
	// Migrate old effect IDs to new ones for backward compatibility
	migrateOldEffectIDs(state)

	// Ensure fatigue accumulation effect is present
	if err := ensureFatigueAccumulation(state); err != nil {
		return fmt.Errorf("failed to ensure fatigue accumulation: %w", err)
	}

	// Ensure hunger accumulation effect is present
	if err := ensureHungerAccumulation(state); err != nil {
		return fmt.Errorf("failed to ensure hunger accumulation: %w", err)
	}

	// Apply penalty effects based on current levels
	if _, err := updateFatiguePenaltyEffects(state); err != nil {
		return fmt.Errorf("failed to update fatigue penalty effects: %w", err)
	}

	if _, err := updateHungerPenaltyEffects(state); err != nil {
		return fmt.Errorf("failed to update hunger penalty effects: %w", err)
	}

	return nil
}

// Fatigue management functions (penalty effects applied based on numeric fatigue level)

// updateFatiguePenaltyEffects applies appropriate penalty effects based on fatigue level
// New thresholds: 0-5 (no penalty), 6 (tired), 8 (very tired), 9 (fatigued), 10 (exhaustion)
func updateFatiguePenaltyEffects(state *SaveFile) (*EffectMessage, error) {
	// Remove all existing fatigue penalty effects
	removeFatiguePenaltyEffects(state)

	// Apply penalty effect based on current fatigue level
	switch {
	case state.Fatigue >= 10:
		return applyEffectWithMessage(state, "exhaustion")
	case state.Fatigue == 9:
		return applyEffectWithMessage(state, "fatigued")
	case state.Fatigue == 8:
		return applyEffectWithMessage(state, "very-tired")
	case state.Fatigue >= 6:
		return applyEffectWithMessage(state, "tired")
	default:
		// Fatigue 0-5: No fatigue penalty (fresh)
		return nil, nil
	}
}

// removeFatiguePenaltyEffects removes all fatigue penalty effects
func removeFatiguePenaltyEffects(state *SaveFile) {
	var remainingEffects []ActiveEffect
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

// ensureFatigueAccumulation ensures the fatigue accumulation effect is active
// Only adds if fatigue < 10 (stops accumulation at max)
func ensureFatigueAccumulation(state *SaveFile) error {
	// Don't accumulate if already at max fatigue
	if state.Fatigue >= 10 {
		removeFatigueAccumulation(state)
		return nil
	}

	// Check if already present
	for _, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID == "fatigue-accumulation" {
			return nil // Already present
		}
	}

	// Apply it
	return applyEffect(state, "fatigue-accumulation")
}

// removeFatigueAccumulation removes the fatigue accumulation effect
func removeFatigueAccumulation(state *SaveFile) {
	var remainingEffects []ActiveEffect
	for _, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID != "fatigue-accumulation" {
			remainingEffects = append(remainingEffects, activeEffect)
		}
	}
	state.ActiveEffects = remainingEffects
}

// resetFatigueAccumulator resets the tick accumulator for fatigue accumulation effect
func resetFatigueAccumulator(state *SaveFile) {
	for i, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID == "fatigue-accumulation" {
			state.ActiveEffects[i].TickAccumulator = 0
			return
		}
	}
}

// Hunger management functions (penalty effects applied based on numeric hunger level)

// updateHungerPenaltyEffects applies appropriate penalty effects based on hunger level
// 3/3 "Stuffed": +1 CON, -1 STR, -1 DEX
// 2/3 "Well Fed": No effect (baseline)
// 1/3 "Hungry": -1 DEX only
// 0/3 "Famished": -1 HP every 4 hours
func updateHungerPenaltyEffects(state *SaveFile) (*EffectMessage, error) {
	// Remove all existing hunger penalty effects
	removeHungerPenaltyEffects(state)

	// Apply penalty/bonus effect based on current hunger level
	switch state.Hunger {
	case 0:
		return applyEffectWithMessage(state, "starving")
	case 1:
		return applyEffectWithMessage(state, "hungry")
	case 2:
		// Well fed - no effect (baseline)
		return nil, nil
	case 3:
		return applyEffectWithMessage(state, "stuffed")
	default:
		// Clamp to valid range
		if state.Hunger < 0 {
			state.Hunger = 0
			return applyEffectWithMessage(state, "starving")
		}
		state.Hunger = 3
		return applyEffectWithMessage(state, "stuffed")
	}
}

// removeHungerPenaltyEffects removes all hunger penalty effects
func removeHungerPenaltyEffects(state *SaveFile) {
	var remainingEffects []ActiveEffect
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

// ensureHungerAccumulation ensures hunger accumulation effect is present (no swapping needed)
func ensureHungerAccumulation(state *SaveFile) error {
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

	return applyEffect(state, effectID)
}

// removeHungerAccumulation removes all hunger accumulation effects
func removeHungerAccumulation(state *SaveFile) {
	var remainingEffects []ActiveEffect
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

// resetHungerAccumulator resets the tick accumulator for hunger accumulation effects
func resetHungerAccumulator(state *SaveFile) {
	for i, activeEffect := range state.ActiveEffects {
		if activeEffect.EffectID == "hunger-accumulation-stuffed" ||
			activeEffect.EffectID == "hunger-accumulation-wellfed" ||
			activeEffect.EffectID == "hunger-accumulation-hungry" {
			state.ActiveEffects[i].TickAccumulator = 0
			return
		}
	}
}

// getEffectTemplate loads effect template data and returns the specific effect at index
func getEffectTemplate(effectID string, effectIndex int) (effectType string, value float64, tickInterval float64, name string, err error) {
	effectData, err := loadEffectData(effectID)
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

// tickEffects processes all active effects, applying stat modifiers and ticking down durations
// Returns a slice of messages from effects that triggered (like starvation damage)
func tickEffects(state *SaveFile, minutesElapsed int) []EffectMessage {
	if len(state.ActiveEffects) == 0 {
		return nil
	}

	var remainingEffects []ActiveEffect
	var messages []EffectMessage

	for _, activeEffect := range state.ActiveEffects {
		// Load effect template to get type, value, tick_interval
		effectType, value, tickInterval, name, err := getEffectTemplate(activeEffect.EffectID, activeEffect.EffectIndex)
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
					applyImmediateEffect(state, effectType, int(value))
					activeEffect.TickAccumulator -= tickInterval

					// For starvation damage, show message
					if activeEffect.EffectID == "starving" && effectType == "hp" {
						messages = append(messages, EffectMessage{
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

// tickDownEffectDurations reduces duration_remaining for all timed effects
// Used during sleep and other time jumps to properly expire buffs/debuffs
func tickDownEffectDurations(state *SaveFile, minutes int) {
	if state.ActiveEffects == nil || minutes <= 0 {
		return
	}

	var remainingEffects []ActiveEffect
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

// getActiveStatModifiers calculates total stat modifiers from all active effects
func getActiveStatModifiers(state *SaveFile) map[string]int {
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
		effectType, value, _, _, err := getEffectTemplate(activeEffect.EffectID, activeEffect.EffectIndex)
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

// loadEffectData loads effect data from database
func loadEffectData(effectID string) (map[string]interface{}, error) {
	// Normalize old effect IDs for backward compatibility
	effectID = normalizeEffectID(effectID)

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

// enrichActiveEffects adds template data (name, category, stat_modifiers) to active effects
// Used when sending active_effects to the frontend for display
func enrichActiveEffects(activeEffects []ActiveEffect) []EnrichedEffect {
	enriched := make([]EnrichedEffect, 0, len(activeEffects))

	for _, ae := range activeEffects {
		ee := EnrichedEffect{
			ActiveEffect:  ae,
			Name:          ae.EffectID, // Default to ID
			Category:      "modifier",
			StatModifiers: make(map[string]int),
			TickInterval:  0,
		}

		// Load template data
		effectData, err := loadEffectData(ae.EffectID)
		if err == nil {
			// Get name
			if name, ok := effectData["name"].(string); ok {
				ee.Name = name
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

// advanceTime advances the game time by the specified minutes and ticks all time-based systems
// Returns messages from any effects that triggered (like starvation damage)
func advanceTime(state *SaveFile, minutes int) []EffectMessage {
	oldTime := state.TimeOfDay
	oldDay := state.CurrentDay

	// Advance time
	state.TimeOfDay += minutes

	// Handle day wrap-around
	if state.TimeOfDay >= 1440 {
		daysAdvanced := state.TimeOfDay / 1440
		state.CurrentDay += daysAdvanced
		state.TimeOfDay = state.TimeOfDay % 1440
	}

	// Tick active effects (includes fatigue/hunger accumulation effects)
	messages := tickEffects(state, minutes)

	// Update penalty effects based on current fatigue/hunger levels
	fatigueMsg, _ := updateFatiguePenaltyEffects(state)
	hungerMsg, _ := updateHungerPenaltyEffects(state)

	// Add any new penalty effect messages
	if fatigueMsg != nil && !fatigueMsg.Silent {
		messages = append(messages, *fatigueMsg)
	}
	if hungerMsg != nil && !hungerMsg.Silent {
		messages = append(messages, *hungerMsg)
	}

	if state.CurrentDay != oldDay {
		log.Printf("📅 Day advanced from %d to %d", oldDay, state.CurrentDay)
	}

	// Only log time changes when hour changes (reduces log spam)
	oldHour := oldTime / 60
	newHour := state.TimeOfDay / 60
	if newHour != oldHour || state.CurrentDay != oldDay {
		log.Printf("⏰ Hour changed: %02d:00 -> %02d:00 (Day %d)", oldHour, newHour, state.CurrentDay)
	}

	return messages
}

// ============================================================================
// TYPE CONVERSION HELPERS
// ============================================================================

// getIntValue safely extracts an int from a map, handling both int and float64
func getIntValue(m map[string]any, key string, defaultValue int) int {
	val, ok := m[key]
	if !ok {
		return defaultValue
	}

	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return defaultValue
	}
}

// getFloatValue safely extracts a float64 from a map, handling both int and float64
func getFloatValue(m map[string]any, key string, defaultValue float64) float64 {
	val, ok := m[key]
	if !ok {
		return defaultValue
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultValue
	}
}
