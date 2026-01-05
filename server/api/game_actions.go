package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"

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
	Delta   map[string]any `json:"delta,omitempty"` // Only changed fields (for optimization)
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

	// Process the action based on type
	response, err := processGameAction(&session.SaveData, request.Action)
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

	log.Printf("✅ Action processed: %s for %s", request.Action.Type, request.SaveID)

	// Return updated state
	response.State = &session.SaveData
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processGameAction routes to specific action handlers
func processGameAction(state *SaveFile, action GameAction) (*GameActionResponse, error) {
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
		return handleNPCDialogueChoiceAction(state, action.Params)
	case "register_vault":
		return handleRegisterVaultAction(state, action.Params)
	case "open_vault":
		return handleOpenVaultAction(state, action.Params)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// ============================================================================
// ACTION HANDLERS
// ============================================================================

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
			if state.Hunger > oldHunger {
				effectMessages = append(effectMessages, "Hunger restored")
			} else if state.Hunger < oldHunger {
				effectMessages = append(effectMessages, "Hunger decreased")
			}

		case "fatigue":
			oldFatigue := state.Fatigue
			state.Fatigue = max(0, state.Fatigue+int(effectValue))
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

	// Update fatigue counter (increments every 240 minutes = 4 hours)
	state.FatigueCounter += float64(minutesToAdvance)
	if state.FatigueCounter >= 240.0 {
		state.Fatigue++
		state.FatigueCounter = 0
	}

	// Update hunger counter (decreases every 360 minutes = 6 hours, or 720 minutes = 12 hours if already hungry)
	state.HungerCounter += float64(minutesToAdvance)
	hungerThreshold := 360.0 // 6 hours in minutes
	if state.Hunger <= 1 {
		hungerThreshold = 720.0 // 12 hours in minutes (slower when already hungry)
	}
	if state.HungerCounter >= hungerThreshold {
		if state.Hunger > 0 {
			state.Hunger--
		}
		state.HungerCounter = 0
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Advanced %d time segment(s)", int(segments)),
	}, nil
}

// handleUpdateTimeAction syncs time from frontend clock to backend state
func handleUpdateTimeAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
	// Update time_of_day (minutes in day, 0-1439)
	if timeOfDay, ok := params["time_of_day"].(float64); ok {
		state.TimeOfDay = int(timeOfDay)
	}

	// Update current_day
	if currentDay, ok := params["current_day"].(float64); ok {
		state.CurrentDay = int(currentDay)
	}

	// Update fatigue counter (in minutes)
	if fatigueCounter, ok := params["fatigue_counter"].(float64); ok {
		state.FatigueCounter = fatigueCounter
	}

	// Update hunger counter (in minutes)
	if hungerCounter, ok := params["hunger_counter"].(float64); ok {
		state.HungerCounter = hungerCounter
	}

	// Update fatigue level
	if fatigue, ok := params["fatigue"].(float64); ok {
		state.Fatigue = int(fatigue)
	}

	// Update hunger level
	if hunger, ok := params["hunger"].(float64); ok {
		state.Hunger = int(hunger)
	}

	return &GameActionResponse{
		Success: true,
		Message: "Time updated",
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

	log.Printf("🔀 Moving %s from %s[%d] to %s[%d]", itemID, fromSlotType, fromSlot, toSlotType, toSlot)

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
		vault := getVaultForLocation(state, vaultBuilding)
		if vault != nil {
			delta["vault_data"] = vault
			log.Printf("✅ Returning updated vault data in response")
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

	log.Printf("📦 Stacking %s from %s[%d] to %s[%d]", itemID, fromSlotType, fromSlot, toSlotType, toSlot)

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

	// Find first empty slot in general inventory
	generalSlots, ok := state.Inventory["general_slots"].([]any)
	if !ok {
		return nil, fmt.Errorf("invalid general slots")
	}

	for i, slotData := range generalSlots {
		slotMap, ok := slotData.(map[string]any)
		if !ok {
			continue
		}

		// Check if slot is empty
		if slotMap["item"] == nil || slotMap["item"] == "" {
			// Add item to this slot
			slotMap["item"] = itemID
			slotMap["quantity"] = quantity
			log.Printf("✅ Added item to slot %d", i)

			return &GameActionResponse{
				Success: true,
				Message: fmt.Sprintf("Added %dx %s", quantity, itemID),
			}, nil
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"state":   session.SaveData,
	})
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

	isOpen, openHour, closeHour, err := utils.IsBuildingOpen(database, state.Location, buildingID, state.TimeOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to check building hours: %v", err)
	}

	if !isOpen {
		return &GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("The building is closed. Open hours: %02d:00 - %02d:00", openHour, closeHour),
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

	// Check if fatigue increased to warn user
	message := "Exited building"
	if state.Fatigue > 0 && state.FatigueCounter == 0 {
		// Fatigue just increased
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

	// Check if player is at NPC's current location
	playerAtLocation := false
	if scheduleInfo.LocationType == "building" && state.Building == scheduleInfo.LocationID {
		playerAtLocation = true
	} else if scheduleInfo.LocationType == "district" && state.Building == "" {
		// Construct full district ID from location + district (e.g., "kingdom" + "center" = "kingdom-center")
		playerDistrictID := fmt.Sprintf("%s-%s", state.Location, state.District)
		if playerDistrictID == scheduleInfo.LocationID {
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
	isRegistered := isVaultRegistered(state, scheduleInfo.LocationID)
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
func handleNPCDialogueChoiceAction(state *SaveFile, params map[string]any) (*GameActionResponse, error) {
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
		log.Printf("🔍 isVaultRegistered: No vaults array found")
		return false
	}
	log.Printf("🔍 isVaultRegistered: Checking building '%s' (location: '%s') against %d vaults", buildingID, state.Location, len(state.Vaults))
	for i, vault := range state.Vaults {
		// Check new format (building field)
		if building, ok := vault["building"].(string); ok {
			log.Printf("  - Vault %d: building = '%s' (new format)", i, building)
			if building == buildingID {
				log.Printf("  ✅ Match found (by building)!")
				return true
			}
		} else if location, ok := vault["location"].(string); ok {
			// Check old format (location field) - match if we're at that location
			log.Printf("  - Vault %d: location = '%s' (old format)", i, location)
			if location == state.Location {
				log.Printf("  ✅ Match found (by location - old format)!")
				return true
			}
		} else {
			log.Printf("  - Vault %d: no building or location field", i)
		}
	}
	log.Printf("  ❌ No match found for building '%s'", buildingID)
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
