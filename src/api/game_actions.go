package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// GameAction represents any action a player can take
type GameAction struct {
	Type   string                 `json:"type"`   // "move", "use_item", "equip", "cast_spell", etc.
	Params map[string]interface{} `json:"params"` // Action-specific parameters
}

// GameActionResponse is returned after processing an action
type GameActionResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	State   *SaveFile              `json:"state,omitempty"`   // Updated game state
	Delta   map[string]interface{} `json:"delta,omitempty"`   // Only changed fields (for optimization)
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
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// ============================================================================
// ACTION HANDLERS
// ============================================================================

// handleMoveAction moves the player to a new location
func handleMoveAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
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
	discovered := false
	for _, loc := range state.LocationsDiscovered {
		if loc == location {
			discovered = true
			break
		}
	}
	if !discovered {
		state.LocationsDiscovered = append(state.LocationsDiscovered, location)
	}

	// Advance time by 1 segment when moving locations
	timeParams := map[string]interface{}{
		"segments": float64(1),
	}
	_, err := handleAdvanceTimeAction(state, timeParams)
	if err != nil {
		log.Printf("⚠️ Failed to advance time: %v", err)
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Moved to %s", location),
	}, nil
}

// handleUseItemAction uses a consumable item
func handleUseItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
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
	generalSlots, ok := state.Inventory["general_slots"].([]interface{})
	if ok {
		for i, slotData := range generalSlots {
			slotMap, ok := slotData.(map[string]interface{})
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
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]interface{})
		bag, _ := gearSlots["bag"].(map[string]interface{})
		contents, ok := bag["contents"].([]interface{})
		if ok {
			for i, slotData := range contents {
				slotMap, ok := slotData.(map[string]interface{})
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

// applyItemEffects applies item effects to the character state
func applyItemEffects(state *SaveFile, itemID string) []string {
	var effects []string

	// Hardcoded item effects for common consumables
	// TODO: Load from database/JSON
	switch itemID {
	case "health-potion", "potion-of-healing":
		healAmount := 10
		state.HP = min(state.MaxHP, state.HP+healAmount)
		effects = append(effects, fmt.Sprintf("Healed %d HP", healAmount))

	case "greater-health-potion":
		healAmount := 20
		state.HP = min(state.MaxHP, state.HP+healAmount)
		effects = append(effects, fmt.Sprintf("Healed %d HP", healAmount))

	case "mana-potion":
		manaAmount := 10
		state.Mana = min(state.MaxMana, state.Mana+manaAmount)
		effects = append(effects, fmt.Sprintf("Restored %d mana", manaAmount))

	case "rations":
		state.Hunger = min(3, state.Hunger+1)
		effects = append(effects, "Hunger restored")

	case "waterskin":
		// Minor fatigue reduction
		state.Fatigue = max(0, state.Fatigue-1)
		effects = append(effects, "Feeling refreshed")

	default:
		effects = append(effects, "Used")
	}

	return effects
}

// handleEquipItemAction equips an item from inventory
func handleEquipItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	toEquip, ok := params["equipment_slot"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid equipment_slot parameter")
	}

	fromSlot := int(params["from_slot"].(float64))
	fromSlotType, _ := params["from_slot_type"].(string)

	log.Printf("⚔️ Equipping %s from %s[%d] to %s", itemID, fromSlotType, fromSlot, toEquip)

	// Get inventory structure
	gearSlots, ok := state.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		gearSlots = make(map[string]interface{})
		state.Inventory["gear_slots"] = gearSlots
	}

	// Check if something is already equipped in that slot
	currentlyEquipped := gearSlots[toEquip]
	if currentlyEquipped != nil {
		log.Printf("⚠️ Slot %s is occupied, will be unequipped", toEquip)
		// Unequip current item to first available inventory slot
		// For now, we'll just overwrite (TODO: add auto-unequip logic)
	}

	// Remove item from source inventory
	var sourceItem map[string]interface{}

	if fromSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok || fromSlot >= len(generalSlots) {
			return nil, fmt.Errorf("invalid source slot")
		}
		slotMap, ok := generalSlots[fromSlot].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid slot data")
		}
		sourceItem = slotMap
	} else if fromSlotType == "inventory" {
		// From backpack
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("backpack not found")
		}
		contents, ok := bag["contents"].([]interface{})
		if !ok || fromSlot >= len(contents) {
			return nil, fmt.Errorf("invalid backpack slot")
		}
		slotMap, ok := contents[fromSlot].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid slot data")
		}
		sourceItem = slotMap
	} else {
		return nil, fmt.Errorf("invalid source slot type: %s", fromSlotType)
	}

	// Verify the item matches
	if sourceItem["item"] != itemID {
		return nil, fmt.Errorf("item mismatch: expected %s, found %v", itemID, sourceItem["item"])
	}

	// Equip the item
	gearSlots[toEquip] = map[string]interface{}{
		"item":     itemID,
		"quantity": 1,
	}

	// Clear source slot
	sourceItem["item"] = nil
	sourceItem["quantity"] = 0

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Equipped %s", itemID),
	}, nil
}

// handleUnequipItemAction unequips an item to inventory
func handleUnequipItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	slot, ok := params["equipment_slot"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid equipment_slot parameter")
	}

	// Get gear slots
	inventory, ok := state.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid inventory structure")
	}

	// Check if slot has an item
	currentItem := inventory[slot]
	if currentItem == nil {
		return nil, fmt.Errorf("no item equipped in slot %s", slot)
	}

	// TODO: Find empty slot in general inventory
	// TODO: Move item to general inventory
	// For now, just clear the equipment slot
	inventory[slot] = nil

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Unequipped item from %s", slot),
	}, nil
}

// handleDropItemAction drops an item from inventory
func handleDropItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	slot, _ := params["slot"].(float64)

	// Find and remove item
	inventory, ok := state.Inventory["general_slots"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid inventory structure")
	}

	for i, slotData := range inventory {
		slotMap, ok := slotData.(map[string]interface{})
		if !ok {
			continue
		}

		if slotMap["item"] == itemID && (slot < 0 || i == int(slot)) {
			slotMap["item"] = nil
			slotMap["quantity"] = 0
			break
		}
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Dropped %s", itemID),
	}, nil
}

// handlePickupItemAction picks up an item from the ground
func handlePickupItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
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
func handleCastSpellAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
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
func handleRestAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	// Restore HP and Mana
	state.HP = state.MaxHP
	state.Mana = state.MaxMana
	state.Fatigue = 0

	// Advance time
	state.TimeOfDay = (state.TimeOfDay + 8) % 12 // Rest for 8 time segments
	if state.TimeOfDay < 8 {
		state.CurrentDay++
	}

	return &GameActionResponse{
		Success: true,
		Message: "Rested and restored HP/Mana",
	}, nil
}

// handleAdvanceTimeAction advances game time
func handleAdvanceTimeAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	segments, ok := params["segments"].(float64)
	if !ok {
		segments = 1 // Default to 1 time segment
	}

	newTime := state.TimeOfDay + int(segments)
	daysAdvanced := newTime / 12
	state.CurrentDay += daysAdvanced
	state.TimeOfDay = newTime % 12

	// Update fatigue counter
	state.FatigueCounter += int(segments)
	if state.FatigueCounter >= 2 {
		state.Fatigue++
		state.FatigueCounter = 0
	}

	// Update hunger counter
	state.HungerCounter += int(segments)
	hungerThreshold := 3
	if state.Hunger <= 1 {
		hungerThreshold = 6 // Slower when already hungry
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

// handleVaultDepositAction deposits items into vault
func handleVaultDepositAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	quantity, _ := params["quantity"].(float64)
	if quantity <= 0 {
		quantity = 1
	}

	// TODO: Validate player is at vault location
	// TODO: Find item in inventory
	// TODO: Find empty vault slot
	// TODO: Move item from inventory to vault

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Deposited %s to vault", itemID),
	}, nil
}

// handleVaultWithdrawAction withdraws items from vault
func handleVaultWithdrawAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	quantity, _ := params["quantity"].(float64)
	if quantity <= 0 {
		quantity = 1
	}

	// TODO: Validate player is at vault location
	// TODO: Find item in vault
	// TODO: Find empty inventory slot
	// TODO: Move item from vault to inventory

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Withdrew %s from vault", itemID),
	}, nil
}

// handleMoveItemAction moves/swaps items between inventory slots
func handleMoveItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	fromSlot := int(params["from_slot"].(float64))
	toSlot := int(params["to_slot"].(float64))
	fromSlotType, _ := params["from_slot_type"].(string)
	toSlotType, _ := params["to_slot_type"].(string)

	log.Printf("🔀 Moving %s from %s[%d] to %s[%d]", itemID, fromSlotType, fromSlot, toSlotType, toSlot)

	// Get the appropriate slot arrays
	var fromSlots, toSlots []interface{}

	if fromSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		fromSlots = generalSlots
	} else if fromSlotType == "inventory" {
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]interface{})
		bag, _ := gearSlots["bag"].(map[string]interface{})
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		fromSlots = contents
	}

	if toSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		toSlots = generalSlots
	} else if toSlotType == "inventory" {
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]interface{})
		bag, _ := gearSlots["bag"].(map[string]interface{})
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		toSlots = contents
	}

	// Swap items
	if fromSlots != nil && toSlots != nil && fromSlot >= 0 && toSlot >= 0 {
		if fromSlotType == toSlotType {
			// Same array, just swap within it
			fromSlots[fromSlot], fromSlots[toSlot] = fromSlots[toSlot], fromSlots[fromSlot]

			// Update the "slot" field in each swapped item
			if fromSlotMap, ok := fromSlots[fromSlot].(map[string]interface{}); ok {
				fromSlotMap["slot"] = fromSlot
			}
			if toSlotMap, ok := fromSlots[toSlot].(map[string]interface{}); ok {
				toSlotMap["slot"] = toSlot
			}
		} else {
			// Different arrays, swap between them
			temp := fromSlots[fromSlot]
			fromSlots[fromSlot] = toSlots[toSlot]
			toSlots[toSlot] = temp

			// Update the "slot" field in each swapped item
			if fromSlotMap, ok := fromSlots[fromSlot].(map[string]interface{}); ok {
				fromSlotMap["slot"] = fromSlot
			}
			if toSlotMap, ok := toSlots[toSlot].(map[string]interface{}); ok {
				toSlotMap["slot"] = toSlot
			}
		}

		log.Printf("✅ Swapped slots: %s[%d] ↔ %s[%d]", fromSlotType, fromSlot, toSlotType, toSlot)
	}

	return &GameActionResponse{
		Success: true,
		Message: "Item moved",
	}, nil
}

// handleStackItemAction stacks items together
func handleStackItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	fromSlot := int(params["from_slot"].(float64))
	toSlot := int(params["to_slot"].(float64))
	fromSlotType, _ := params["from_slot_type"].(string)
	toSlotType, _ := params["to_slot_type"].(string)

	log.Printf("📦 Stacking %s from %s[%d] to %s[%d]", itemID, fromSlotType, fromSlot, toSlotType, toSlot)

	// Get source slots
	var fromSlots []interface{}
	if fromSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		fromSlots = generalSlots
	} else if fromSlotType == "inventory" {
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]interface{})
		bag, _ := gearSlots["bag"].(map[string]interface{})
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		fromSlots = contents
	} else {
		return nil, fmt.Errorf("invalid source slot type: %s", fromSlotType)
	}

	// Get destination slots
	var toSlots []interface{}
	if toSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		toSlots = generalSlots
	} else if toSlotType == "inventory" {
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]interface{})
		bag, _ := gearSlots["bag"].(map[string]interface{})
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		toSlots = contents
	} else {
		return nil, fmt.Errorf("invalid destination slot type: %s", toSlotType)
	}

	// Get source and destination items
	fromSlotMap, ok := fromSlots[fromSlot].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid source slot data")
	}

	toSlotMap, ok := toSlots[toSlot].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid destination slot data")
	}

	// Verify both slots have the same item
	if fromSlotMap["item"] != itemID || toSlotMap["item"] != itemID {
		return nil, fmt.Errorf("items don't match for stacking")
	}

	// Get quantities
	fromQty, _ := fromSlotMap["quantity"].(float64)
	toQty, _ := toSlotMap["quantity"].(float64)

	// Combine stacks
	newQty := int(fromQty) + int(toQty)

	// Update destination slot
	toSlotMap["quantity"] = float64(newQty)

	// Clear source slot
	fromSlotMap["item"] = nil
	fromSlotMap["quantity"] = 0

	log.Printf("✅ Stacked %s: %d + %d = %d", itemID, int(fromQty), int(toQty), newQty)

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Stacked items (%d total)", newQty),
	}, nil
}

// handleSplitItemAction splits a stack into two stacks
func handleSplitItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	fromSlot := int(params["from_slot"].(float64))
	toSlot := int(params["to_slot"].(float64))
	fromSlotType, _ := params["from_slot_type"].(string)
	toSlotType, _ := params["to_slot_type"].(string)
	splitQuantity := int(params["quantity"].(float64))

	log.Printf("✂️ Splitting %s: %d from %s[%d] to %s[%d]", itemID, splitQuantity, fromSlotType, fromSlot, toSlotType, toSlot)

	// Get source slot
	var fromSlots []interface{}
	if fromSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		fromSlots = generalSlots
	} else if fromSlotType == "inventory" {
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]interface{})
		bag, _ := gearSlots["bag"].(map[string]interface{})
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		fromSlots = contents
	} else {
		return nil, fmt.Errorf("invalid source slot type: %s", fromSlotType)
	}

	// Get destination slot
	var toSlots []interface{}
	if toSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid general slots")
		}
		toSlots = generalSlots
	} else if toSlotType == "inventory" {
		gearSlots, _ := state.Inventory["gear_slots"].(map[string]interface{})
		bag, _ := gearSlots["bag"].(map[string]interface{})
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid backpack")
		}
		toSlots = contents
	} else {
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
	fromSlotMap, ok := fromSlots[fromSlot].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid source slot data")
	}

	// Verify item ID matches
	if fromSlotMap["item"] != itemID {
		return nil, fmt.Errorf("item mismatch in source slot")
	}

	// Get current quantity
	currentQty, ok := fromSlotMap["quantity"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid quantity in source slot")
	}

	// Validate split quantity
	if splitQuantity <= 0 || splitQuantity >= int(currentQty) {
		return nil, fmt.Errorf("invalid split quantity: %d (current: %d)", splitQuantity, int(currentQty))
	}

	// Check destination slot is empty
	toSlotMap, ok := toSlots[toSlot].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid destination slot data")
	}
	if toSlotMap["item"] != nil && toSlotMap["item"] != "" {
		return nil, fmt.Errorf("destination slot is not empty")
	}

	// Perform split
	remainingQty := int(currentQty) - splitQuantity

	// Update source slot
	fromSlotMap["quantity"] = float64(remainingQty)

	// Update destination slot
	toSlotMap["item"] = itemID
	toSlotMap["quantity"] = float64(splitQuantity)
	toSlotMap["slot"] = toSlot

	log.Printf("✅ Split complete: %s (%d remaining in slot %d, %d in new slot %d)", itemID, remainingQty, fromSlot, splitQuantity, toSlot)

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Split %d items into new stack", splitQuantity),
	}, nil
}

// handleAddItemAction adds an item to inventory
func handleAddItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, _ := params["item_id"].(string)
	quantity := 1
	if q, ok := params["quantity"].(float64); ok {
		quantity = int(q)
	}

	log.Printf("➕ Adding %dx %s to inventory", quantity, itemID)

	// Find first empty slot in general inventory
	generalSlots, ok := state.Inventory["general_slots"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid general slots")
	}

	for i, slotData := range generalSlots {
		slotMap, ok := slotData.(map[string]interface{})
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"state":   session.SaveData,
	})
}
