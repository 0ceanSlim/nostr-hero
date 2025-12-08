package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nostr-hero/src/db"
	"nostr-hero/src/types"
	"strings"
)

// InventoryHandler handles all inventory-related operations
func InventoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.ItemActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Error decoding inventory request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Load the save file
	save, err := LoadSaveByID(req.Npub, req.SaveID)
	if err != nil {
		log.Printf("❌ Error loading save: %v", err)
		sendInventoryError(w, "Save file not found")
		return
	}

	// Route to appropriate action handler
	var response *types.ItemActionResponse
	switch req.Action {
	case "equip":
		response = handleEquipItem(save, &req)
	case "unequip":
		response = handleUnequipItem(save, &req)
	case "use":
		response = handleUseItem(save, &req)
	case "drop":
		response = handleDropItem(save, &req)
	case "examine":
		response = handleExamineItem(save, &req)
	case "move":
		response = handleMoveItem(save, &req)
	case "add":
		response = handleAddItem(save, &req)
	case "stack":
		response = handleStackItem(save, &req)
	case "split":
		response = handleSplitItem(save, &req)
	case "add_to_container":
		response = handleAddToContainer(save, &req)
	case "remove_from_container":
		response = handleRemoveFromContainer(save, &req)
	default:
		response = &types.ItemActionResponse{
			Success: false,
			Error:   "Unknown action: " + req.Action,
			Color:   "red",
		}
	}

	// If action was successful, save the updated state
	if response.Success && req.Action != "examine" {
		if err := writeSaveFile(fmt.Sprintf("data/saves/%s/%s.json", req.Npub, req.SaveID), save); err != nil {
			log.Printf("❌ Error saving updated inventory: %v", err)
			response.Success = false
			response.Error = "Failed to save changes"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleEquipItem equips an item from inventory to equipment slot
func handleEquipItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	log.Printf("⚔️ Equip action: %s from %s[%d] to equipment slot", req.ItemID, req.FromSlotType, req.FromSlot)

	// Get equipment slots
	gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		gearSlots = make(map[string]interface{})
		save.Inventory["gear_slots"] = gearSlots
	}

	// Determine which equipment slot from item's gear_slot property
	equipSlot := req.ToEquip
	isTwoHanded := false

	if equipSlot == "" {
		// Fetch item data from database to get gear_slot property
		database := db.GetDB()
		if database == nil {
			return &types.ItemActionResponse{Success: false, Error: "Database not available"}
		}

		var propertiesJSON string
		var itemType string
		err := database.QueryRow("SELECT properties, item_type FROM items WHERE id = ?", req.ItemID).Scan(&propertiesJSON, &itemType)
		if err != nil {
			log.Printf("❌ Failed to find item '%s' in database: %v", req.ItemID, err)
			return &types.ItemActionResponse{Success: false, Error: fmt.Sprintf("Item '%s' not found in database", req.ItemID)}
		}

		var properties map[string]interface{}
		if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
			return &types.ItemActionResponse{Success: false, Error: "Failed to parse item properties"}
		}

		// Check if item has two-handed tag
		isTwoHanded = false
		if tags, ok := properties["tags"].([]interface{}); ok {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok && tagStr == "two-handed" {
					isTwoHanded = true
					break
				}
			}
		}

		if gearSlotProp, ok := properties["gear_slot"].(string); ok {
			// Map item gear_slot to actual save file slot names
			switch gearSlotProp {
			case "hands":
				// For weapons and shields, determine slot
				if strings.Contains(strings.ToLower(itemType), "shield") || strings.Contains(strings.ToLower(itemType), "armor") {
					equipSlot = "left_arm"
				} else {
					// Weapon logic: check slot availability
					rightArm := gearSlots["right_arm"]
					leftArm := gearSlots["left_arm"]

					rightOccupied := false
					leftOccupied := false

					if rightMap, ok := rightArm.(map[string]interface{}); ok {
						if rightMap["item"] != nil && rightMap["item"] != "" {
							rightOccupied = true
						}
					}
					if leftMap, ok := leftArm.(map[string]interface{}); ok {
						if leftMap["item"] != nil && leftMap["item"] != "" {
							leftOccupied = true
						}
					}

					// Choose slot based on availability
					if !rightOccupied {
						equipSlot = "right_arm"
					} else if !leftOccupied {
						equipSlot = "left_arm"
					} else {
						// Both full, will swap with right_arm
						equipSlot = "right_arm"
					}
				}
			case "armor", "body":
				equipSlot = "armor"
			case "neck", "necklace":
				equipSlot = "necklace"
			case "finger", "ring":
				equipSlot = "ring"
			case "ammunition", "ammo":
				equipSlot = "ammunition"
			case "clothes", "clothing":
				equipSlot = "clothes"
			case "bag", "backpack":
				equipSlot = "bag"
			default:
				equipSlot = gearSlotProp // Use as-is if it matches
			}
		} else {
			return &types.ItemActionResponse{Success: false, Error: "Item does not have a gear_slot property"}
		}
	}

	log.Printf("⚔️ Equipping to slot: %s (two-handed: %v)", equipSlot, isTwoHanded)

	// For two-handed weapons, unequip both hands first
	var itemsToUnequip []map[string]interface{}
	if isTwoHanded {
		// Unequip right_arm if occupied
		if rightMap, ok := gearSlots["right_arm"].(map[string]interface{}); ok {
			if rightMap["item"] != nil && rightMap["item"] != "" {
				itemsToUnequip = append(itemsToUnequip, map[string]interface{}{
					"item":     rightMap["item"],
					"quantity": rightMap["quantity"],
					"from":     "right_arm",
				})
			}
		}
		// Unequip left_arm if occupied
		if leftMap, ok := gearSlots["left_arm"].(map[string]interface{}); ok {
			if leftMap["item"] != nil && leftMap["item"] != "" {
				itemsToUnequip = append(itemsToUnequip, map[string]interface{}{
					"item":     leftMap["item"],
					"quantity": leftMap["quantity"],
					"from":     "left_arm",
				})
			}
		}
	} else {
		// Check if equipping a one-handed weapon when two-handed is equipped
		// (Need to unequip the two-handed weapon)
		// For now, just check if target slot is occupied
		if existing := gearSlots[equipSlot]; existing != nil {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if existingMap["item"] != nil && existingMap["item"] != "" {
					// Will swap with this item
					itemsToUnequip = append(itemsToUnequip, map[string]interface{}{
						"item":     existingMap["item"],
						"quantity": existingMap["quantity"],
						"from":     equipSlot,
					})
				}
			}
		}
	}

	// Find item in inventory (check both general_slots and backpack)
	fromSlotType := req.FromSlotType
	if fromSlotType == "" {
		fromSlotType = "general" // Default for backward compatibility
	}

	var itemData map[string]interface{}
	var sourceInventory []interface{}

	if fromSlotType == "general" {
		generalSlots, ok := save.Inventory["general_slots"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "Invalid general_slots structure"}
		}
		sourceInventory = generalSlots

		if req.FromSlot < 0 || req.FromSlot >= len(sourceInventory) {
			return &types.ItemActionResponse{Success: false, Error: "Invalid source slot"}
		}

		itemMap, ok := sourceInventory[req.FromSlot].(map[string]interface{})
		if !ok || itemMap["item"] != req.ItemID {
			return &types.ItemActionResponse{Success: false, Error: "Item not found in specified slot"}
		}
		itemData = itemMap

	} else if fromSlotType == "inventory" {
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "No backpack found"}
		}
		backpackContents, ok := bag["contents"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "Invalid backpack contents"}
		}
		sourceInventory = backpackContents

		// Search for item by slot number, not array index
		found := false
		for _, slotData := range sourceInventory {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
				if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == req.FromSlot {
					if slotMap["item"] == req.ItemID {
						itemData = slotMap
						found = true
						break
					}
				}
			}
		}
		if !found {
			return &types.ItemActionResponse{Success: false, Error: "Item not found in specified slot"}
		}
	} else {
		return &types.ItemActionResponse{Success: false, Error: "Invalid source slot type"}
	}

	// Find source slot array index if searching backpack by slot number
	sourceArrayIndex := req.FromSlot
	if fromSlotType == "inventory" {
		// Find the array index for this slot
		for i, slotData := range sourceInventory {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
				if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == req.FromSlot {
					sourceArrayIndex = i
					break
				}
			}
		}
	}

	// Add unequipped items back to inventory (swapped items)
	for _, unequipData := range itemsToUnequip {
		// Put the item in the slot we're taking the new item from
		swappedItem := map[string]interface{}{
			"item":     unequipData["item"],
			"quantity": unequipData["quantity"],
			"slot":     req.FromSlot,
		}

		// Special case: If unequipping a bag, preserve its contents
		slotName := unequipData["from"].(string)
		if slotName == "bag" {
			if existingBag, ok := gearSlots["bag"].(map[string]interface{}); ok {
				if contents, ok := existingBag["contents"].([]interface{}); ok {
					swappedItem["contents"] = contents
					log.Printf("📦 Preserving bag contents when swapping (%d items)", len(contents))
				}
			}
		}

		sourceInventory[sourceArrayIndex] = swappedItem

		// Clear the equipment slot
		gearSlots[slotName] = map[string]interface{}{
			"item":     nil,
			"quantity": 0,
		}
	}

	// Move item to equipment slot (only store item and quantity, remove slot field)
	equippedItem := map[string]interface{}{
		"item":     itemData["item"],
		"quantity": itemData["quantity"],
	}

	// Special case: If equipping a bag, preserve its contents
	if equipSlot == "bag" {
		if contents, ok := itemData["contents"].([]interface{}); ok {
			equippedItem["contents"] = contents
			log.Printf("📦 Preserving bag contents when equipping (%d items)", len(contents))
		}
	}

	gearSlots[equipSlot] = equippedItem

	// For two-handed weapons, also clear left_arm
	if isTwoHanded {
		gearSlots["left_arm"] = map[string]interface{}{
			"item":     nil,
			"quantity": 0,
		}
	}

	// Empty the source slot if no swap occurred
	if len(itemsToUnequip) == 0 {
		sourceInventory[sourceArrayIndex] = map[string]interface{}{
			"item":     nil,
			"quantity": 0,
			"slot":     req.FromSlot,
		}
	}

	// Save back to correct location
	if fromSlotType == "general" {
		save.Inventory["general_slots"] = sourceInventory
	} else {
		bag := gearSlots["bag"].(map[string]interface{})
		bag["contents"] = sourceInventory
	}

	log.Printf("✅ Equipped %s to %s (swapped %d items)", req.ItemID, equipSlot, len(itemsToUnequip))

	// Get item name from database for better message
	itemName := req.ItemID
	database := db.GetDB()
	if database != nil {
		var name string
		err := database.QueryRow("SELECT name FROM items WHERE id = ?", req.ItemID).Scan(&name)
		if err == nil {
			itemName = name
		}
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("Equipped %s", itemName),
		Color:    "green",
		NewState: save.Inventory,
	}
}

// handleUnequipItem moves an item from equipment slot to inventory
func handleUnequipItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	log.Printf("🛡️ Unequip action from equipment slot: %s", req.FromEquip)

	// Get equipment slots
	gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid equipment structure",
			Color:   "red",
		}
	}

	// Get the item from equipment slot
	equipSlot := req.FromEquip
	itemData, exists := gearSlots[equipSlot]
	if !exists || itemData == nil {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("No item in equipment slot '%s'", equipSlot),
		}
	}

	// Verify there's actually an item
	itemMap, ok := itemData.(map[string]interface{})
	if !ok || itemMap["item"] == nil || itemMap["item"] == "" {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("No item in equipment slot '%s'", equipSlot),
		}
	}

	itemID := itemMap["item"].(string)
	quantity := 1
	if qty, ok := itemMap["quantity"].(float64); ok {
		quantity = int(qty)
	} else if qty, ok := itemMap["quantity"].(int); ok {
		quantity = qty
	}

	log.Printf("🛡️ Unequipping: %s (quantity: %d)", itemID, quantity)

	// Log initial state of general_slots
	if generalSlots, ok := save.Inventory["general_slots"].([]interface{}); ok {
		log.Printf("📊 INITIAL STATE - General slots count: %d", len(generalSlots))
		for i, slot := range generalSlots {
			if slotMap, ok := slot.(map[string]interface{}); ok {
				log.Printf("   [%d] item=%v, slot=%v", i, slotMap["item"], slotMap["slot"])
			}
		}
	}

	// Special case: If unequipping a bag, it can ONLY go to general slots (not backpack)
	// Bags are containers and containers cannot go in other containers (including themselves)
	emptySlotIndex := -1
	emptySlotType := ""

	if equipSlot == "bag" {
		log.Printf("🎒 Special case: Unequipping bag - searching for empty general slot")
		// Bag can only go to general slots
		generalSlots, ok := save.Inventory["general_slots"].([]interface{})
		if !ok {
			generalSlots = make([]interface{}, 0, 4)
			save.Inventory["general_slots"] = generalSlots
		}

		for i := 0; i < 4; i++ {
			// Extend array if needed
			if i >= len(generalSlots) {
				generalSlots = append(generalSlots, map[string]interface{}{
					"item":     nil,
					"quantity": 0,
					"slot":     i,
				})
				save.Inventory["general_slots"] = generalSlots
			}

			// Check if slot is empty
			if generalSlots[i] == nil {
				emptySlotIndex = i
				emptySlotType = "general"
				log.Printf("✅ Found empty general slot at index %d (nil)", i)
				break
			}

			if slotMap, ok := generalSlots[i].(map[string]interface{}); ok {
				if slotMap["item"] == nil || slotMap["item"] == "" {
					emptySlotIndex = i
					emptySlotType = "general"
					log.Printf("✅ Found empty general slot at index %d (empty item)", i)
					break
				}
			}
		}

		// If no empty general slot, cannot unequip bag
		if emptySlotIndex == -1 {
			log.Printf("❌ No empty general slot found for bag")
			return &types.ItemActionResponse{
				Success: false,
				Error:   "There isn't enough room in your inventory to take off the bag",
				Color:   "red",
			}
		}
		log.Printf("🎒 Bag will be placed in general slot %d", emptySlotIndex)
	} else {
		// For non-bag items, try backpack first, then general slots
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "No backpack found",
				Color:   "red",
			}
		}

		backpackContents, ok := bag["contents"].([]interface{})
		if !ok {
			backpackContents = make([]interface{}, 0, 20)
			bag["contents"] = backpackContents
		}

		// Look for empty slot in backpack (slots 0-19)
		for i := 0; i < 20; i++ {
			// Extend array if needed
			if i >= len(backpackContents) {
				backpackContents = append(backpackContents, map[string]interface{}{
					"item":     nil,
					"quantity": 0,
					"slot":     i,
				})
				bag["contents"] = backpackContents
			}

			// Check if slot is empty
			if backpackContents[i] == nil {
				emptySlotIndex = i
				emptySlotType = "inventory"
				break
			}

			if slotMap, ok := backpackContents[i].(map[string]interface{}); ok {
				if slotMap["item"] == nil || slotMap["item"] == "" {
					emptySlotIndex = i
					emptySlotType = "inventory"
					break
				}
			}
		}

		// If no empty backpack slot, check general slots
		if emptySlotIndex == -1 {
			generalSlots, ok := save.Inventory["general_slots"].([]interface{})
			if !ok {
				generalSlots = make([]interface{}, 0, 4)
				save.Inventory["general_slots"] = generalSlots
			}

			for i := 0; i < 4; i++ {
				// Extend array if needed
				if i >= len(generalSlots) {
					generalSlots = append(generalSlots, map[string]interface{}{
						"item":     nil,
						"quantity": 0,
						"slot":     i,
					})
					save.Inventory["general_slots"] = generalSlots
				}

				// Check if slot is empty
				if generalSlots[i] == nil {
					emptySlotIndex = i
					emptySlotType = "general"
					break
				}

				if slotMap, ok := generalSlots[i].(map[string]interface{}); ok {
					if slotMap["item"] == nil || slotMap["item"] == "" {
						emptySlotIndex = i
						emptySlotType = "general"
						break
					}
				}
			}
		}

		// If no empty slot found, inventory is full
		if emptySlotIndex == -1 {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Your inventory is full",
				Color:   "red",
			}
		}
	}

	// Place item in the empty slot
	newItem := map[string]interface{}{
		"item":     itemID,
		"quantity": quantity,
		"slot":     emptySlotIndex,
	}

	// Special case: If unequipping a bag, preserve its contents
	if equipSlot == "bag" {
		if contents, ok := itemMap["contents"].([]interface{}); ok {
			newItem["contents"] = contents
			log.Printf("📦 Preserving bag contents (%d items)", len(contents))
		}
	}

	if emptySlotType == "inventory" {
		// Sanity check: bags should never go into backpack (they go into general slots only)
		if equipSlot == "bag" {
			log.Printf("❌ ERROR: Attempted to place bag in backpack slot - this should never happen!")
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Internal error: bag cannot go in backpack",
				Color:   "red",
			}
		}

		bag, _ := gearSlots["bag"].(map[string]interface{})
		backpackContents, _ := bag["contents"].([]interface{})
		backpackContents[emptySlotIndex] = newItem
		bag["contents"] = backpackContents
		log.Printf("✅ Moved to backpack slot %d", emptySlotIndex)
	} else {
		log.Printf("📦 Placing item in general slots at index %d", emptySlotIndex)
		generalSlots := save.Inventory["general_slots"].([]interface{})
		log.Printf("📦 General slots before placement: length=%d, item at index %d: %v", len(generalSlots), emptySlotIndex, generalSlots[emptySlotIndex])
		generalSlots[emptySlotIndex] = newItem
		save.Inventory["general_slots"] = generalSlots
		log.Printf("✅ Moved to general slot %d, newItem: %+v", emptySlotIndex, newItem)
		log.Printf("📦 General slots after placement: %+v", generalSlots[emptySlotIndex])

		// Verify it was actually saved
		verifySlot := save.Inventory["general_slots"].([]interface{})[emptySlotIndex]
		log.Printf("🔍 VERIFICATION - Reading back from save.Inventory[general_slots][%d]: %+v", emptySlotIndex, verifySlot)
	}

	// Empty the equipment slot
	log.Printf("🧹 Clearing equipment slot: %s", equipSlot)
	gearSlots[equipSlot] = map[string]interface{}{
		"item":     nil,
		"quantity": 0,
	}
	log.Printf("✅ Equipment slot %s cleared", equipSlot)

	// Final verification log
	log.Printf("🔍 FINAL STATE CHECK:")
	log.Printf("  - Equipment slot '%s': %+v", equipSlot, gearSlots[equipSlot])
	if emptySlotType == "general" {
		finalGeneralSlots := save.Inventory["general_slots"].([]interface{})
		log.Printf("  - General slot %d: %+v", emptySlotIndex, finalGeneralSlots[emptySlotIndex])
	}

	// Get item name from database for better message
	itemName := itemID
	database := db.GetDB()
	if database != nil {
		var name string
		err := database.QueryRow("SELECT name FROM items WHERE id = ?", itemID).Scan(&name)
		if err == nil {
			itemName = name
		}
	}

	// Log what we're about to return
	log.Printf("📤 RETURNING RESPONSE:")
	log.Printf("  - Success: true")
	log.Printf("  - Message: Unequipped %s", itemName)
	if emptySlotType == "general" {
		returnedGeneralSlots := save.Inventory["general_slots"].([]interface{})
		log.Printf("  - NewState general_slots[%d]: %+v", emptySlotIndex, returnedGeneralSlots[emptySlotIndex])
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("Unequipped %s", itemName),
		Color:    "green",
		NewState: save.Inventory,
	}
}

// handleUseItem uses a consumable item
func handleUseItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	// Find item in either general slots or backpack
	var itemSlot int = -1
	var itemData map[string]interface{}
	var inventoryLocation string // "general" or "backpack"
	var inventory []interface{}

	// Determine which inventory to search based on FromSlotType
	if req.FromSlotType == "general" {
		// Search general slots for the item at the specific slot
		if generalSlots, ok := save.Inventory["general_slots"].([]interface{}); ok {
			for i, item := range generalSlots {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Match both item ID and slot number
					if itemMap["item"] == req.ItemID && int(itemMap["slot"].(float64)) == req.FromSlot {
						itemSlot = i
						itemData = itemMap
						inventory = generalSlots
						inventoryLocation = "general"
						break
					}
				}
			}
		}
	} else if req.FromSlotType == "inventory" {
		// Search backpack for the item at the specific slot
		if gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{}); ok {
			if bag, ok := gearSlots["bag"].(map[string]interface{}); ok {
				if backpackSlots, ok := bag["contents"].([]interface{}); ok {
					for i, item := range backpackSlots {
						if itemMap, ok := item.(map[string]interface{}); ok {
							// Match both item ID and slot number
							if itemMap["item"] == req.ItemID && int(itemMap["slot"].(float64)) == req.FromSlot {
								itemSlot = i
								itemData = itemMap
								inventory = backpackSlots
								inventoryLocation = "backpack"
								break
							}
						}
					}
				}
			}
		}
	}

	if itemSlot == -1 {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Item not found in inventory",
		}
	}

	// Query item from database to get properties and effects
	database := db.GetDB()
	if database == nil {
		return &types.ItemActionResponse{Success: false, Error: "Database not available"}
	}

	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM items WHERE id = ?", req.ItemID).Scan(&propertiesJSON)
	if err != nil {
		log.Printf("⚠️ Could not find item %s in database: %v", req.ItemID, err)
		// Continue anyway to allow using the item
		propertiesJSON = "{}"
	}

	// Parse properties to get effects
	var properties map[string]interface{}
	var effectsApplied []string
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err == nil {
		if effects, ok := properties["effects"].([]interface{}); ok {
			// First check if any hunger effect would exceed max (prevent eating when full)
			for _, effect := range effects {
				if effectMap, ok := effect.(map[string]interface{}); ok {
					if effectType, ok := effectMap["type"].(string); ok && effectType == "hunger" {
						if value, ok := effectMap["value"].(float64); ok {
							newHunger := save.Hunger + int(value)
							if newHunger > 3 {
								return &types.ItemActionResponse{
									Success: false,
									Error:   "You're too full to eat",
								}
							}
						}
					}
				}
			}

			// Apply each effect
			for _, effect := range effects {
				if effectMap, ok := effect.(map[string]interface{}); ok {
					effectType, typeOk := effectMap["type"].(string)
					value, valueOk := effectMap["value"].(float64)

					if !typeOk || !valueOk {
						continue
					}

					switch effectType {
					case "hunger":
						save.Hunger += int(value)
						save.Hunger = max(0, min(save.Hunger, 3)) // Clamp to 0-3
						save.HungerCounter = 0                    // Reset hunger counter when eating
						effectsApplied = append(effectsApplied, fmt.Sprintf("Hunger %+d", int(value)))

					case "fatigue":
						save.Fatigue += int(value)
						save.Fatigue = max(0, min(save.Fatigue, 9)) // Clamp to 0-9
						effectsApplied = append(effectsApplied, fmt.Sprintf("Fatigue %+d", int(value)))

					case "hp":
						save.HP += int(value)
						save.HP = max(0, min(save.HP, save.MaxHP)) // Clamp to 0-max_hp
						effectsApplied = append(effectsApplied, fmt.Sprintf("HP %+d", int(value)))

					case "mana":
						save.Mana += int(value)
						save.Mana = max(0, min(save.Mana, save.MaxMana)) // Clamp to 0-max_mana
						effectsApplied = append(effectsApplied, fmt.Sprintf("Mana %+d", int(value)))
					}
				}
			}
		}
	}

	// Handle quantity
	quantity := int(itemData["quantity"].(float64))
	if quantity > 1 {
		itemData["quantity"] = quantity - 1
		inventory[itemSlot] = itemData
	} else {
		// Empty the slot but keep it in the array to preserve slot numbers
		itemData["item"] = nil
		itemData["quantity"] = 0
		inventory[itemSlot] = itemData
	}

	// Save to correct location
	if inventoryLocation == "general" {
		save.Inventory["general_slots"] = inventory
	} else if inventoryLocation == "backpack" {
		if gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{}); ok {
			if bag, ok := gearSlots["bag"].(map[string]interface{}); ok {
				bag["contents"] = inventory
			}
		}
	}

	// Get item name from database for better message
	itemName := req.ItemID
	database2 := db.GetDB()
	if database2 != nil {
		var name string
		err := database2.QueryRow("SELECT name FROM items WHERE id = ?", req.ItemID).Scan(&name)
		if err == nil {
			itemName = name
		}
	}

	// Build message with effects applied
	message := fmt.Sprintf("Used %s", itemName)
	if len(effectsApplied) > 0 {
		message = fmt.Sprintf("Used %s (%s)", itemName, strings.Join(effectsApplied, ", "))
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  message,
		Color:    "blue",
		NewState: save.Inventory,
	}
}

// handleDropItem removes an item from inventory (or reduces quantity if partial drop)
func handleDropItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	log.Printf("🗑️ Drop item: %s from %s[%d], quantity: %d", req.ItemID, req.FromSlotType, req.FromSlot, req.Quantity)

	// Determine which inventory we're dropping from using FromSlotType
	fromSlotType := req.FromSlotType
	if fromSlotType == "" {
		// Default to general if not specified (backward compatibility)
		fromSlotType = "general"
	}

	var inventory []interface{}
	var itemSlot int = req.FromSlot
	var currentQuantity int = 1

	// Get the correct inventory array
	if fromSlotType == "general" {
		generalInventory, ok := save.Inventory["general_slots"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid general_slots structure",
			}
		}
		inventory = generalInventory

		// Validate slot index
		if itemSlot < 0 || itemSlot >= len(inventory) {
			return &types.ItemActionResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid slot index: %d", itemSlot),
			}
		}

		// Get the item at the specific slot
		if inventory[itemSlot] == nil {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Slot is empty",
			}
		}

		itemMap, ok := inventory[itemSlot].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid slot structure",
			}
		}

		// Verify it's the correct item
		if itemMap["item"] != req.ItemID {
			return &types.ItemActionResponse{
				Success: false,
				Error:   fmt.Sprintf("Slot %d contains %v, not %s", itemSlot, itemMap["item"], req.ItemID),
			}
		}

		// Get quantity
		if qty, ok := itemMap["quantity"].(float64); ok {
			currentQuantity = int(qty)
		} else if qty, ok := itemMap["quantity"].(int); ok {
			currentQuantity = qty
		}

	} else if fromSlotType == "inventory" {
		// Get backpack contents
		gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid gear_slots structure",
			}
		}

		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "No backpack found",
			}
		}

		backpackInventory, ok := bag["contents"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid backpack contents",
			}
		}
		inventory = backpackInventory

		// Search for item by slot number, not array index
		arrayIndex := -1
		var itemMap map[string]interface{}
		for i, slotData := range inventory {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
				if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == itemSlot {
					arrayIndex = i
					itemMap = slotMap
					break
				}
			}
		}

		if arrayIndex == -1 {
			return &types.ItemActionResponse{
				Success: false,
				Error:   fmt.Sprintf("Source slot %d is empty", itemSlot),
			}
		}

		// Verify it's the correct item
		if itemMap["item"] != req.ItemID {
			return &types.ItemActionResponse{
				Success: false,
				Error:   fmt.Sprintf("Slot %d contains %v, not %s", itemSlot, itemMap["item"], req.ItemID),
			}
		}

		// Get quantity
		if qty, ok := itemMap["quantity"].(float64); ok {
			currentQuantity = int(qty)
		} else if qty, ok := itemMap["quantity"].(int); ok {
			currentQuantity = qty
		}

		// Update itemSlot to use array index for later operations
		itemSlot = arrayIndex

	} else {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid slot type: " + fromSlotType,
		}
	}

	// Determine how many to drop (default to entire stack)
	dropQuantity := req.Quantity
	if dropQuantity <= 0 {
		dropQuantity = currentQuantity
	}

	log.Printf("🗑️ Dropping %d of %d from %s[%d]", dropQuantity, currentQuantity, fromSlotType, itemSlot)

	// Handle partial vs full drop
	if dropQuantity >= currentQuantity {
		// Drop entire stack - set slot to null (preserve array indices)
		inventory[itemSlot] = map[string]interface{}{
			"item":     nil,
			"quantity": 0,
			"slot":     itemSlot,
		}
	} else {
		// Drop partial stack - reduce quantity
		newQuantity := currentQuantity - dropQuantity
		if itemMap, ok := inventory[itemSlot].(map[string]interface{}); ok {
			itemMap["quantity"] = newQuantity
		}
	}

	// Save the updated inventory back to the save file
	if fromSlotType == "general" {
		save.Inventory["general_slots"] = inventory
	} else {
		gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
		bag := gearSlots["bag"].(map[string]interface{})
		bag["contents"] = inventory
	}

	return &types.ItemActionResponse{
		Success: true,
		Message: fmt.Sprintf("Dropped %d item(s)", dropQuantity),
		NewState: save.Inventory,
	}
}

// handleExamineItem returns item information
func handleExamineItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	// This action doesn't modify state, just returns item info
	// The frontend will fetch item details from cached data
	return &types.ItemActionResponse{
		Success: true,
		Message: "Item examined",
	}
}

// handleMoveItem moves an item between inventory slots
func handleMoveItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	// Determine which inventory areas we're working with
	fromSlotType := req.FromSlotType
	toSlotType := req.ToSlotType

	// Default to "general" if not specified (backward compatibility)
	if fromSlotType == "" {
		fromSlotType = "general"
	}
	if toSlotType == "" {
		toSlotType = "general"
	}

	// Get the source inventory array
	var fromInventory []interface{}
	if fromSlotType == "general" {
		inv, ok := save.Inventory["general_slots"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid general_slots structure",
			}
		}
		fromInventory = inv
	} else if fromSlotType == "inventory" {
		// Get backpack contents from gear_slots.bag.contents
		gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid gear_slots structure",
			}
		}
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid bag structure",
			}
		}
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid bag contents structure",
			}
		}
		fromInventory = contents
	} else {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid source slot type: " + fromSlotType,
		}
	}

	// Extend fromInventory to accommodate the slot if needed (sparse array support)
	maxSlots := 4  // General slots max
	if fromSlotType == "inventory" {
		maxSlots = 20  // Backpack max
	}

	// Extend array with nils if needed
	for len(fromInventory) <= req.FromSlot && len(fromInventory) < maxSlots {
		fromInventory = append(fromInventory, nil)
	}

	// Save the extended array back
	if fromSlotType == "general" {
		save.Inventory["general_slots"] = fromInventory
	} else {
		gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
		bag := gearSlots["bag"].(map[string]interface{})
		bag["contents"] = fromInventory
	}

	// Search for item by slot number, not array index
	fromArrayIndex := -1
	var slotObj map[string]interface{}
	for i, slotData := range fromInventory {
		if slotMap, ok := slotData.(map[string]interface{}); ok {
			if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == req.FromSlot {
				fromArrayIndex = i
				slotObj = slotMap
				break
			}
		}
	}

	if fromArrayIndex == -1 {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Source slot %d is empty", req.FromSlot),
		}
	}

	itemID, _ := slotObj["item"].(string)
	if itemID == "" || slotObj["item"] == nil {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Source slot %d is empty", req.FromSlot),
		}
	}

	// CRITICAL VALIDATION: Check if moving a container to backpack - containers cannot go in backpack
	if toSlotType == "inventory" {
		log.Printf("🔍 VALIDATION CHECK: Is '%s' a container? (destination: backpack)", itemID)

		database := db.GetDB()
		if database == nil {
			log.Printf("❌ CRITICAL: Database not available - cannot validate container restriction")
			return &types.ItemActionResponse{
				Success: false,
				Error:   "System error: Cannot validate item restrictions",
				Color:   "red",
			}
		}

		// Query tags directly from the tags column
		var tagsJSON string
		err := database.QueryRow("SELECT tags FROM items WHERE id = ?", itemID).Scan(&tagsJSON)
		if err != nil {
			log.Printf("❌ CRITICAL: Failed to query tags for %s: %v", itemID, err)
			return &types.ItemActionResponse{
				Success: false,
				Error:   fmt.Sprintf("System error: Cannot find item data for %s", itemID),
				Color:   "red",
			}
		}

		log.Printf("📦 Raw tags JSON from database for '%s': %s", itemID, tagsJSON)

		var tags []interface{}
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			log.Printf("❌ CRITICAL: Failed to parse tags JSON for %s: %v", itemID, err)
			return &types.ItemActionResponse{
				Success: false,
				Error:   "System error: Invalid item data format",
				Color:   "red",
			}
		}

		log.Printf("📦 Parsed tags array for '%s': %v", itemID, tags)

		// Check each tag
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				log.Printf("   🏷️ Found tag: '%s'", tagStr)
				if tagStr == "container" {
					log.Printf("❌ BLOCKED: '%s' has 'container' tag - CANNOT go in backpack!", itemID)
					return &types.ItemActionResponse{
						Success: false,
						Error:   "Containers cannot be stored in the backpack",
						Color:   "red",
					}
				}
			}
		}

		log.Printf("✅ VALIDATION PASSED: '%s' is NOT a container - allowing move to backpack", itemID)
	}

	// Get the destination inventory array
	var toInventory []interface{}
	if toSlotType == "general" {
		inv, ok := save.Inventory["general_slots"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid general_slots structure",
			}
		}
		toInventory = inv
	} else if toSlotType == "inventory" {
		// Get backpack contents from gear_slots.bag.contents
		gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid gear_slots structure",
			}
		}
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid bag structure",
			}
		}
		contents, ok := bag["contents"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{
				Success: false,
				Error:   "Invalid bag contents structure",
			}
		}
		toInventory = contents
	} else {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid destination slot type: " + toSlotType,
		}
	}

	// Validate destination slot
	maxSlots = 4  // General slots max
	if toSlotType == "inventory" {
		maxSlots = 20  // Backpack max
	}

	if req.ToSlot < 0 || req.ToSlot >= maxSlots {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid destination slot: %d (max is %d)", req.ToSlot, maxSlots-1),
		}
	}

	// If moving within the same inventory array
	if fromSlotType == toSlotType {
		// Re-fetch the inventory since we just extended and saved it
		if fromSlotType == "general" {
			fromInventory = save.Inventory["general_slots"].([]interface{})
		} else {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			fromInventory = bag["contents"].([]interface{})
		}

		// Extend to accommodate destination slot if needed
		for len(fromInventory) <= req.ToSlot {
			// Create empty slot object
			emptySlot := map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     len(fromInventory),
			}
			fromInventory = append(fromInventory, emptySlot)
		}

		// fromSlotObj is already set from earlier search
		// Search for destination slot by slot number
		toArrayIndex := -1
		var toSlotObj map[string]interface{}
		for i, slotData := range fromInventory {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
				if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == req.ToSlot {
					toArrayIndex = i
					toSlotObj = slotMap
					break
				}
			}
		}

		// If destination slot doesn't exist, create it
		if toArrayIndex == -1 {
			toSlotObj = map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     req.ToSlot,
			}
			toArrayIndex = len(fromInventory)
			fromInventory = append(fromInventory, toSlotObj)
		}

		// Swap item and quantity, but keep slot numbers correct
		slotObj["item"], toSlotObj["item"] = toSlotObj["item"], slotObj["item"]
		slotObj["quantity"], toSlotObj["quantity"] = toSlotObj["quantity"], slotObj["quantity"]

		// Save back to the correct location
		if fromSlotType == "general" {
			save.Inventory["general_slots"] = fromInventory
		} else {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			bag["contents"] = fromInventory
		}
	} else {
		// Moving between different inventory arrays (general <-> backpack)
		// slotObj already contains the source item

		// Search for destination slot by slot number
		toArrayIndex := -1
		var toSlotObj map[string]interface{}
		for i, slotData := range toInventory {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
				if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == req.ToSlot {
					toArrayIndex = i
					toSlotObj = slotMap
					break
				}
			}
		}

		// If destination slot doesn't exist, create it
		if toArrayIndex == -1 {
			toSlotObj = map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     req.ToSlot,
			}
			toArrayIndex = len(toInventory)
			toInventory = append(toInventory, toSlotObj)
		}

		// If destination slot has an item, swap them
		if toSlotObj["item"] != nil && toSlotObj["item"] != "" {
			// Swap the items - exchange all data but preserve slot numbers
			slotObj["item"], toSlotObj["item"] = toSlotObj["item"], slotObj["item"]
			slotObj["quantity"], toSlotObj["quantity"] = toSlotObj["quantity"], slotObj["quantity"]
		} else {
			// Destination is empty, just move
			toSlotObj["item"] = slotObj["item"]
			toSlotObj["quantity"] = slotObj["quantity"]
			slotObj["item"] = nil
			slotObj["quantity"] = 0
		}

		// Save both arrays back
		if fromSlotType == "general" {
			save.Inventory["general_slots"] = fromInventory
		} else {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			bag["contents"] = fromInventory
		}

		if toSlotType == "general" {
			save.Inventory["general_slots"] = toInventory
		} else {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			bag["contents"] = toInventory
		}
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  "", // Suppress message for move operations
		NewState: save.Inventory,
	}
}

// handleAddItem adds an item to inventory (for picking up from ground)
func handleAddItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	// Get item's stack limit from database
	database := db.GetDB()
	if database == nil {
		return &types.ItemActionResponse{Success: false, Error: "Database not available"}
	}

	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM items WHERE id = ?", req.ItemID).Scan(&propertiesJSON)
	if err != nil {
		log.Printf("⚠️ Could not find item %s in database, assuming stack limit 1", req.ItemID)
		propertiesJSON = "{}"
	}

	// Parse properties to get stack limit
	var properties map[string]interface{}
	stackLimit := 1 // Default stack limit
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err == nil {
		if limit, ok := properties["stack_limit"].(float64); ok {
			stackLimit = int(limit)
		} else if limit, ok := properties["stack"].(float64); ok {
			stackLimit = int(limit)
		}
	}

	log.Printf("📦 Adding %s (quantity: %d, stack limit: %d)", req.ItemID, req.Quantity, stackLimit)

	// Get general_slots and backpack inventories
	generalInventory, ok := save.Inventory["general_slots"].([]interface{})
	if !ok {
		generalInventory = make([]interface{}, 0, 4)
		save.Inventory["general_slots"] = generalInventory
	}

	gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Your inventory is full",
		}
	}

	bag, ok := gearSlots["bag"].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Your inventory is full",
		}
	}

	backpackContents, ok := bag["contents"].([]interface{})
	if !ok {
		backpackContents = make([]interface{}, 0, 20)
		bag["contents"] = backpackContents
	}

	remainingQuantity := req.Quantity

	// STEP 1: Try to stack with existing items (if stackable and not at max)
	if stackLimit > 1 {
		// Check backpack first (most items are here)
		for i, slot := range backpackContents {
			if slot != nil {
				if slotMap, ok := slot.(map[string]interface{}); ok {
					if slotMap["item"] == req.ItemID {
						// Found matching item - check if we can add to it
						currentQty := 1
						if qty, ok := slotMap["quantity"].(float64); ok {
							currentQty = int(qty)
						} else if qty, ok := slotMap["quantity"].(int); ok {
							currentQty = qty
						}

						if currentQty < stackLimit {
							// Can add to this stack
							spaceInStack := stackLimit - currentQty
							addAmount := remainingQuantity
							if addAmount > spaceInStack {
								addAmount = spaceInStack
							}

							slotMap["quantity"] = currentQty + addAmount
							remainingQuantity -= addAmount

							log.Printf("✅ Added %d to existing backpack stack at slot %d (now %d/%d)", addAmount, i, currentQty+addAmount, stackLimit)

							if remainingQuantity <= 0 {
								return &types.ItemActionResponse{
									Success:  true,
									Message:  fmt.Sprintf("Picked up %s", req.ItemID),
									NewState: save.Inventory,
								}
							}
						}
					}
				}
			}
		}

		// Check general slots too
		for i, slot := range generalInventory {
			if slot != nil {
				if slotMap, ok := slot.(map[string]interface{}); ok {
					if slotMap["item"] == req.ItemID {
						currentQty := 1
						if qty, ok := slotMap["quantity"].(float64); ok {
							currentQty = int(qty)
						} else if qty, ok := slotMap["quantity"].(int); ok {
							currentQty = qty
						}

						if currentQty < stackLimit {
							spaceInStack := stackLimit - currentQty
							addAmount := remainingQuantity
							if addAmount > spaceInStack {
								addAmount = spaceInStack
							}

							slotMap["quantity"] = currentQty + addAmount
							remainingQuantity -= addAmount

							log.Printf("✅ Added %d to existing general stack at slot %d (now %d/%d)", addAmount, i, currentQty+addAmount, stackLimit)

							if remainingQuantity <= 0 {
								return &types.ItemActionResponse{
									Success:  true,
									Message:  fmt.Sprintf("Picked up %s", req.ItemID),
									NewState: save.Inventory,
								}
							}
						}
					}
				}
			}
		}
	}

	// STEP 2: If still have quantity remaining, find empty slots
	// Try backpack first (slots 0-19)
	for remainingQuantity > 0 {
		// Find first empty slot in backpack
		emptySlotIndex := -1
		for i := 0; i < 20; i++ {
			// Extend array if needed
			if i >= len(backpackContents) {
				backpackContents = append(backpackContents, map[string]interface{}{
					"item":     nil,
					"quantity": 0,
					"slot":     i,
				})
			}

			if backpackContents[i] == nil {
				emptySlotIndex = i
				break
			}

			if slotMap, ok := backpackContents[i].(map[string]interface{}); ok {
				if slotMap["item"] == nil || slotMap["item"] == "" {
					emptySlotIndex = i
					break
				}
			}
		}

		if emptySlotIndex == -1 {
			// No empty backpack slots, try general slots
			emptySlotIndex = -1
			for i := 0; i < 4; i++ {
				if i >= len(generalInventory) {
					generalInventory = append(generalInventory, map[string]interface{}{
						"item":     nil,
						"quantity": 0,
						"slot":     i,
					})
				}

				if generalInventory[i] == nil {
					emptySlotIndex = i
					break
				}

				if slotMap, ok := generalInventory[i].(map[string]interface{}); ok {
					if slotMap["item"] == nil || slotMap["item"] == "" {
						emptySlotIndex = i
						break
					}
				}
			}

			if emptySlotIndex == -1 {
				// Inventory is completely full
				return &types.ItemActionResponse{
					Success: false,
					Error:   "Your inventory is full",
				}
			}

			// Add to general slot
			addAmount := remainingQuantity
			if addAmount > stackLimit {
				addAmount = stackLimit
			}

			generalInventory[emptySlotIndex] = map[string]interface{}{
				"item":     req.ItemID,
				"quantity": addAmount,
				"slot":     emptySlotIndex,
			}
			save.Inventory["general_slots"] = generalInventory
			remainingQuantity -= addAmount

			log.Printf("✅ Added %d to general slot %d", addAmount, emptySlotIndex)
		} else {
			// Add to backpack slot
			addAmount := remainingQuantity
			if addAmount > stackLimit {
				addAmount = stackLimit
			}

			backpackContents[emptySlotIndex] = map[string]interface{}{
				"item":     req.ItemID,
				"quantity": addAmount,
				"slot":     emptySlotIndex,
			}
			bag["contents"] = backpackContents
			remainingQuantity -= addAmount

			log.Printf("✅ Added %d to backpack slot %d", addAmount, emptySlotIndex)
		}
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("Picked up %s", req.ItemID),
		NewState: save.Inventory,
	}
}

// handleStackItem stacks items together, respecting stack limits
func handleStackItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	log.Printf("📦 Stack action: %s from %s[%d] to %s[%d]", req.ItemID, req.FromSlotType, req.FromSlot, req.ToSlotType, req.ToSlot)

	// Get source and destination inventories
	fromSlotType := req.FromSlotType
	toSlotType := req.ToSlotType

	// Helper to get inventory array
	getInventory := func(slotType string) ([]interface{}, error) {
		if slotType == "general" {
			inv, ok := save.Inventory["general_slots"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("general_slots not found")
			}
			return inv, nil
		} else if slotType == "inventory" {
			gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("gear_slots not found")
			}
			bag, ok := gearSlots["bag"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("bag not found")
			}
			contents, ok := bag["contents"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("bag contents not found")
			}
			return contents, nil
		}
		return nil, fmt.Errorf("unknown slot type: %s", slotType)
	}

	fromInventory, err := getInventory(fromSlotType)
	if err != nil {
		return &types.ItemActionResponse{Success: false, Error: err.Error()}
	}

	toInventory, err := getInventory(toSlotType)
	if err != nil {
		return &types.ItemActionResponse{Success: false, Error: err.Error()}
	}

	// Get source and destination items
	if req.FromSlot >= len(fromInventory) || req.ToSlot >= len(toInventory) {
		return &types.ItemActionResponse{Success: false, Error: "Invalid slot index"}
	}

	fromItem, ok := fromInventory[req.FromSlot].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{Success: false, Error: "Source slot is empty"}
	}

	toItem, ok := toInventory[req.ToSlot].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{Success: false, Error: "Destination slot is empty"}
	}

	// Verify they're the same item
	if fromItem["item"] != toItem["item"] {
		return &types.ItemActionResponse{Success: false, Error: "Cannot stack different items"}
	}

	// Get quantities
	fromQty := 1
	if qty, ok := fromItem["quantity"].(float64); ok {
		fromQty = int(qty)
	} else if qty, ok := fromItem["quantity"].(int); ok {
		fromQty = qty
	}

	toQty := 1
	if qty, ok := toItem["quantity"].(float64); ok {
		toQty = int(qty)
	} else if qty, ok := toItem["quantity"].(int); ok {
		toQty = qty
	}

	// Get stack limit from database
	database := db.GetDB()
	if database == nil {
		return &types.ItemActionResponse{Success: false, Error: "Database not available"}
	}

	var propertiesJSON string
	itemID := fromItem["item"].(string)
	err = database.QueryRow("SELECT properties FROM items WHERE id = ?", itemID).Scan(&propertiesJSON)
	if err != nil {
		return &types.ItemActionResponse{Success: false, Error: "Item not found in database"}
	}

	// Parse properties to get stack limit
	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
		return &types.ItemActionResponse{Success: false, Error: "Failed to parse item properties"}
	}

	stackLimit := 1 // Default stack limit
	if limit, ok := properties["stack_limit"].(float64); ok {
		stackLimit = int(limit)
	} else if limit, ok := properties["stack"].(float64); ok {
		stackLimit = int(limit)
	}

	// Calculate how much can be added to destination
	spaceInDest := stackLimit - toQty
	transferAmount := fromQty

	if transferAmount > spaceInDest {
		transferAmount = spaceInDest
	}

	// Update quantities
	newDestQty := toQty + transferAmount
	newSourceQty := fromQty - transferAmount

	toItem["quantity"] = newDestQty

	if newSourceQty > 0 {
		// Update source quantity (partial stack)
		fromItem["quantity"] = newSourceQty

		// Save both inventories back
		if fromSlotType == "general" {
			save.Inventory["general_slots"] = fromInventory
		} else if fromSlotType == "inventory" {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			bag["contents"] = fromInventory
		}

		if toSlotType == "general" {
			save.Inventory["general_slots"] = toInventory
		} else if toSlotType == "inventory" {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			bag["contents"] = toInventory
		}
	} else {
		// Remove source item (stack fully merged) - create proper empty slot
		fromInventory[req.FromSlot] = map[string]interface{}{
			"item":     nil,
			"quantity": 0,
			"slot":     req.FromSlot,
		}

		// Save source inventory
		if fromSlotType == "general" {
			save.Inventory["general_slots"] = fromInventory
		} else if fromSlotType == "inventory" {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			bag["contents"] = fromInventory
		}

		// Save destination inventory
		if toSlotType == "general" {
			save.Inventory["general_slots"] = toInventory
		} else if toSlotType == "inventory" {
			gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
			bag := gearSlots["bag"].(map[string]interface{})
			bag["contents"] = toInventory
		}
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("Stacked %d items (limit: %d)", transferAmount, stackLimit),
		NewState: save.Inventory,
	}
}

// handleSplitItem splits a stack into two stacks
func handleSplitItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	log.Printf("✂️ Split action: %s from %s[%d] to %s[%d], quantity: %d", req.ItemID, req.FromSlotType, req.FromSlot, req.ToSlotType, req.ToSlot, req.Quantity)

	// Helper to get inventory array
	getInventory := func(slotType string) ([]interface{}, error) {
		if slotType == "general" {
			inv, ok := save.Inventory["general_slots"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("general_slots not found")
			}
			return inv, nil
		} else if slotType == "inventory" {
			gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("gear_slots not found")
			}
			bag, ok := gearSlots["bag"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("bag not found")
			}
			contents, ok := bag["contents"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("bag contents not found")
			}
			return contents, nil
		}
		return nil, fmt.Errorf("unknown slot type: %s", slotType)
	}

	// Get source inventory
	fromInventory, err := getInventory(req.FromSlotType)
	if err != nil {
		return &types.ItemActionResponse{Success: false, Error: err.Error()}
	}

	// Validate source slot
	if req.FromSlot >= len(fromInventory) {
		return &types.ItemActionResponse{Success: false, Error: "Invalid source slot index"}
	}

	// Get source item
	fromItem, ok := fromInventory[req.FromSlot].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{Success: false, Error: "Source slot is empty"}
	}

	// Get current quantity
	currentQty := 1
	if qty, ok := fromItem["quantity"].(float64); ok {
		currentQty = int(qty)
	} else if qty, ok := fromItem["quantity"].(int); ok {
		currentQty = qty
	}

	// Validate split quantity
	splitQty := req.Quantity
	if splitQty <= 0 || splitQty >= currentQty {
		return &types.ItemActionResponse{Success: false, Error: "Invalid split quantity"}
	}

	// Reduce source stack
	newSourceQty := currentQty - splitQty
	fromItem["quantity"] = newSourceQty

	// Create new item for destination slot
	itemID := fromItem["item"].(string)
	newItem := map[string]interface{}{
		"item":     itemID,
		"quantity": splitQty,
		"slot":     req.ToSlot,
	}

	// Get destination inventory
	toInventory, err := getInventory(req.ToSlotType)
	if err != nil {
		return &types.ItemActionResponse{Success: false, Error: err.Error()}
	}

	// Extend destination array if needed
	maxSlots := 4  // General slots max
	if req.ToSlotType == "inventory" {
		maxSlots = 20  // Backpack max
	}

	for len(toInventory) <= req.ToSlot && len(toInventory) < maxSlots {
		// Create proper empty slot object instead of nil
		emptySlot := map[string]interface{}{
			"item":     nil,
			"quantity": 0,
			"slot":     len(toInventory),
		}
		toInventory = append(toInventory, emptySlot)
	}

	// Validate destination slot is empty
	if req.ToSlot >= len(toInventory) {
		return &types.ItemActionResponse{Success: false, Error: "Invalid destination slot"}
	}

	if toInventory[req.ToSlot] != nil {
		if toItem, ok := toInventory[req.ToSlot].(map[string]interface{}); ok {
			if toItem["item"] != nil {
				return &types.ItemActionResponse{Success: false, Error: "Destination slot is not empty"}
			}
		}
	}

	// Place new item in destination
	toInventory[req.ToSlot] = newItem

	// Save source inventory
	if req.FromSlotType == "general" {
		save.Inventory["general_slots"] = fromInventory
	} else if req.FromSlotType == "inventory" {
		gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
		bag := gearSlots["bag"].(map[string]interface{})
		bag["contents"] = fromInventory
	}

	// Save destination inventory
	if req.ToSlotType == "general" {
		save.Inventory["general_slots"] = toInventory
	} else if req.ToSlotType == "inventory" {
		gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
		bag := gearSlots["bag"].(map[string]interface{})
		bag["contents"] = toInventory
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("Split %d items from stack", splitQty),
		NewState: save.Inventory,
	}
}

// handleAddToContainer adds an item from inventory to a container
func handleAddToContainer(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	log.Printf("📦 Add to container: %s from %s[%d] to container at slot %d, container slot %d",
		req.ItemID, req.FromSlotType, req.FromSlot, req.ContainerSlot, req.ToContainerSlot)

	// Get the container from general_slots
	generalSlots, ok := save.Inventory["general_slots"].([]interface{})
	if !ok {
		return &types.ItemActionResponse{Success: false, Error: "General slots not found", Color: "red"}
	}

	// Find the container
	var containerSlot map[string]interface{}
	var containerIndex int
	for i, slot := range generalSlots {
		if slotMap, ok := slot.(map[string]interface{}); ok {
			if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == req.ContainerSlot {
				containerSlot = slotMap
				containerIndex = i
				break
			}
		}
	}

	if containerSlot == nil {
		return &types.ItemActionResponse{Success: false, Error: "Container not found", Color: "red"}
	}

	containerID := containerSlot["item"].(string)

	// Get container properties from database
	database := db.GetDB()
	if database == nil {
		return &types.ItemActionResponse{Success: false, Error: "Database not available", Color: "red"}
	}

	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM items WHERE id = ?", containerID).Scan(&propertiesJSON)
	if err != nil {
		return &types.ItemActionResponse{Success: false, Error: "Container item not found in database", Color: "red"}
	}

	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
		return &types.ItemActionResponse{Success: false, Error: "Failed to parse container properties", Color: "red"}
	}

	// Check if item is a container
	tags, ok := properties["tags"].([]interface{})
	isContainer := false
	if ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok && tagStr == "container" {
				isContainer = true
				break
			}
		}
	}

	if !isContainer {
		return &types.ItemActionResponse{Success: false, Error: "Item is not a container", Color: "red"}
	}

	// Get container slots limit
	containerSlots := 10 // default
	if val, ok := properties["container_slots"].(float64); ok {
		containerSlots = int(val)
	}

	// Get allowed types
	allowedTypes := "any"
	if val, ok := properties["allowed_types"].(string); ok {
		allowedTypes = val
	}

	// Get container contents
	contents, ok := containerSlot["contents"].([]interface{})
	if !ok {
		contents = make([]interface{}, 0)
	}

	// Ensure contents array has enough slots
	for len(contents) < containerSlots {
		contents = append(contents, nil)
	}

	// Check if container is full (count non-null items)
	usedSlots := 0
	for _, item := range contents {
		if item != nil {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["item"] != nil {
					usedSlots++
				}
			}
		}
	}

	if usedSlots >= containerSlots {
		return &types.ItemActionResponse{Success: false, Error: "Container is full", Color: "red"}
	}

	// Get item properties to check if it's a container
	var itemPropertiesJSON string
	err = database.QueryRow("SELECT properties FROM items WHERE id = ?", req.ItemID).Scan(&itemPropertiesJSON)
	if err != nil {
		return &types.ItemActionResponse{Success: false, Error: "Item not found in database", Color: "red"}
	}

	var itemProperties map[string]interface{}
	if err := json.Unmarshal([]byte(itemPropertiesJSON), &itemProperties); err != nil {
		return &types.ItemActionResponse{Success: false, Error: "Failed to parse item properties", Color: "red"}
	}

	// Check if item being added is itself a container - containers cannot go in containers
	if itemTags, ok := itemProperties["tags"].([]interface{}); ok {
		for _, tag := range itemTags {
			if tagStr, ok := tag.(string); ok && tagStr == "container" {
				return &types.ItemActionResponse{
					Success: false,
					Error:   "Containers cannot be stored inside other containers",
					Color:   "red",
				}
			}
		}
	}

	// Validate item type if not "any"
	if allowedTypes != "any" {
		// Normalize allowed type for comparison
		normalizedAllowedType := strings.Replace(allowedTypes, "-", "_", -1)
		normalizedAllowedType = strings.TrimSuffix(normalizedAllowedType, "s")

		// Check item tags
		itemAllowed := false
		if itemTags, ok := itemProperties["tags"].([]interface{}); ok {
			for _, tag := range itemTags {
				if tagStr, ok := tag.(string); ok {
					normalizedTag := strings.Replace(tagStr, "-", "_", -1)
					normalizedTag = strings.TrimSuffix(normalizedTag, "s")
					if normalizedTag == normalizedAllowedType {
						itemAllowed = true
						break
					}
				}
			}
		}

		// Also check item_type field
		if !itemAllowed {
			if itemType, ok := itemProperties["item_type"].(string); ok {
				normalizedItemType := strings.Replace(strings.ToLower(itemType), " ", "_", -1)
				normalizedItemType = strings.Replace(normalizedItemType, "-", "_", -1)
				normalizedItemType = strings.TrimSuffix(normalizedItemType, "s")
				if normalizedItemType == normalizedAllowedType {
					itemAllowed = true
				}
			}
		}

		if !itemAllowed {
			return &types.ItemActionResponse{
				Success: false,
				Error:   fmt.Sprintf("This item cannot be stored in this container (requires %s)", allowedTypes),
				Color:   "red",
			}
		}
	}

	// Find the item in source inventory
	var sourceInventory []interface{}
	var sourceItem map[string]interface{}
	var sourceIndex int

	if req.FromSlotType == "general" {
		sourceInventory = generalSlots
	} else if req.FromSlotType == "inventory" {
		gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "Gear slots not found", Color: "red"}
		}
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "Backpack not found", Color: "red"}
		}
		backpackContents, ok := bag["contents"].([]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "Backpack contents not found", Color: "red"}
		}
		sourceInventory = backpackContents
	} else {
		return &types.ItemActionResponse{Success: false, Error: "Invalid source slot type", Color: "red"}
	}

	// Find source item by matching slot index
	for i, slot := range sourceInventory {
		if slotMap, ok := slot.(map[string]interface{}); ok {
			if slotNum, ok := slotMap["slot"].(float64); ok && int(slotNum) == req.FromSlot {
				if slotMap["item"] == req.ItemID {
					sourceItem = slotMap
					sourceIndex = i
					break
				}
			}
		}
	}

	if sourceItem == nil {
		return &types.ItemActionResponse{Success: false, Error: "Item not found in source inventory", Color: "red"}
	}

	// Get item quantity
	quantity := 1
	if qty, ok := sourceItem["quantity"].(float64); ok {
		quantity = int(qty)
	} else if qty, ok := sourceItem["quantity"].(int); ok {
		quantity = qty
	}

	// Get item name for message
	var itemName string
	err = database.QueryRow("SELECT name FROM items WHERE id = ?", req.ItemID).Scan(&itemName)
	if err != nil {
		itemName = req.ItemID
	}

	// Create item entry for container
	containerItem := map[string]interface{}{
		"item":     req.ItemID,
		"quantity": quantity,
	}

	// Find first empty slot in container
	placedAt := -1
	for i := 0; i < containerSlots; i++ {
		if contents[i] == nil || (contents[i] != nil && contents[i].(map[string]interface{})["item"] == nil) {
			contents[i] = containerItem
			placedAt = i
			break
		}
	}

	if placedAt == -1 {
		return &types.ItemActionResponse{Success: false, Error: "Container is full", Color: "red"}
	}

	// Remove item from source (set to empty slot)
	sourceInventory[sourceIndex] = map[string]interface{}{
		"item":     nil,
		"quantity": 0,
		"slot":     req.FromSlot,
	}

	// Update container contents
	generalSlots[containerIndex].(map[string]interface{})["contents"] = contents

	// Save changes
	save.Inventory["general_slots"] = generalSlots
	if req.FromSlotType == "inventory" {
		gearSlots := save.Inventory["gear_slots"].(map[string]interface{})
		bag := gearSlots["bag"].(map[string]interface{})
		bag["contents"] = sourceInventory
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("Added %dx %s to container", quantity, itemName),
		Color:    "green",
		NewState: save.Inventory,
	}
}

// handleRemoveFromContainer removes an item from a container back to inventory
func handleRemoveFromContainer(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	log.Printf("📤 Remove from container: slot %d from container at slot %d",
		req.FromSlot, req.ContainerSlot)

	var containerSlot map[string]interface{}
	var containerIndex int
	containerLocation := "" // "general" or "equipped"

	// Load general slots early - we'll need them regardless of where the container is
	generalSlots, ok := save.Inventory["general_slots"].([]interface{})
	if !ok {
		return &types.ItemActionResponse{Success: false, Error: "General slots not found", Color: "red"}
	}

	// First, check if the container is equipped in gear_slots.bag
	gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
	if ok {
		if bag, ok := gearSlots["bag"].(map[string]interface{}); ok {
			if bag["item"] != nil && bag["item"] != "" {
				// Bag is equipped - use it
				containerSlot = bag
				containerLocation = "equipped"
				log.Printf("📦 Found container in equipped bag slot")
			}
		}
	}

	// If not equipped, search general_slots
	if containerSlot == nil {

		log.Printf("📦 Searching %d general slots for container at slot %d", len(generalSlots), req.ContainerSlot)

		// Find the container in general slots
		for i, slot := range generalSlots {
			if slotMap, ok := slot.(map[string]interface{}); ok {
				// Check slot number (can be int or float64)
				var slotNum int
				if sn, ok := slotMap["slot"].(float64); ok {
					slotNum = int(sn)
				} else if sn, ok := slotMap["slot"].(int); ok {
					slotNum = sn
				} else {
					// No slot field, use array index
					slotNum = i
				}

				log.Printf("🔍 Slot %d (index %d): item=%v, hasContents=%v", slotNum, i, slotMap["item"], slotMap["contents"] != nil)

				if slotNum == req.ContainerSlot {
					// Verify it has contents (is a container)
					if slotMap["item"] != nil && slotMap["item"] != "" {
						containerSlot = slotMap
						containerIndex = i
						containerLocation = "general"
						log.Printf("📦 Found container '%v' in general slot %d (array index %d)", slotMap["item"], req.ContainerSlot, i)
						break
					} else {
						log.Printf("⚠️ Slot %d is empty, not a container", slotNum)
					}
				}
			}
		}
	}

	if containerSlot == nil {
		return &types.ItemActionResponse{Success: false, Error: "Container not found", Color: "red"}
	}

	// Get container contents
	contents, ok := containerSlot["contents"].([]interface{})
	if !ok {
		return &types.ItemActionResponse{Success: false, Error: "Container has no contents", Color: "red"}
	}

	// Get the item from container (use FromSlot, not ToContainerSlot)
	if req.FromSlot >= len(contents) {
		return &types.ItemActionResponse{Success: false, Error: "Invalid container slot", Color: "red"}
	}

	containerItem, ok := contents[req.FromSlot].(map[string]interface{})
	if !ok || containerItem["item"] == nil {
		return &types.ItemActionResponse{Success: false, Error: "Container slot is empty", Color: "red"}
	}

	itemID := containerItem["item"].(string)
	quantity := 1
	if qty, ok := containerItem["quantity"].(float64); ok {
		quantity = int(qty)
	} else if qty, ok := containerItem["quantity"].(int); ok {
		quantity = qty
	}

	// Try to find empty slot in general_slots first
	emptySlotFound := false

	for i, slot := range generalSlots {
		if slotMap, ok := slot.(map[string]interface{}); ok {
			if slotMap["item"] == nil {
				// Found empty general slot
				generalSlots[i] = map[string]interface{}{
					"item":     itemID,
					"quantity": quantity,
					"slot":     slotMap["slot"],
				}
				emptySlotFound = true
				break
			}
		}
	}

	// If no space in general slots, try backpack
	if !emptySlotFound {
		gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "Gear slots not found", Color: "red"}
		}
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return &types.ItemActionResponse{Success: false, Error: "Backpack not found", Color: "red"}
		}
		backpackContents, ok := bag["contents"].([]interface{})
		if !ok {
			backpackContents = make([]interface{}, 0)
		}

		// Extend backpack if needed
		for len(backpackContents) < 20 {
			backpackContents = append(backpackContents, map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     len(backpackContents),
			})
		}

		// Find empty backpack slot
		for i := 0; i < 20; i++ {
			if backpackContents[i] == nil || backpackContents[i].(map[string]interface{})["item"] == nil {
				backpackContents[i] = map[string]interface{}{
					"item":     itemID,
					"quantity": quantity,
					"slot":     i,
				}
				emptySlotFound = true
				break
			}
		}

		if emptySlotFound {
			bag["contents"] = backpackContents
		}
	}

	if !emptySlotFound {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Inventory is full - cannot remove from container",
			Color:   "red",
		}
	}

	// Remove item from container (set to empty slot with proper structure)
	contents[req.FromSlot] = map[string]interface{}{
		"item":     nil,
		"quantity": 0,
		"slot":     req.FromSlot,
	}

	// Update container in the correct location
	if containerLocation == "equipped" {
		// Update equipped bag
		containerSlot["contents"] = contents
		log.Printf("📦 Updated contents of equipped bag")
	} else {
		// Update general slot container
		generalSlots, _ := save.Inventory["general_slots"].([]interface{})
		generalSlots[containerIndex].(map[string]interface{})["contents"] = contents
		save.Inventory["general_slots"] = generalSlots
		log.Printf("📦 Updated contents of container in general slot %d", containerIndex)
	}

	// Get item name for message
	database := db.GetDB()
	var itemName string
	if database != nil {
		err := database.QueryRow("SELECT name FROM items WHERE id = ?", itemID).Scan(&itemName)
		if err != nil {
			itemName = itemID
		}
	} else {
		itemName = itemID
	}

	return &types.ItemActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("Removed %dx %s from container", quantity, itemName),
		Color:    "green",
		NewState: save.Inventory,
	}
}

// ItemActionsHandler returns available actions for an item
func ItemActionsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract item type from query parameters
	itemType := r.URL.Query().Get("type")
	isEquippedStr := r.URL.Query().Get("equipped")

	isEquipped := strings.ToLower(isEquippedStr) == "true"

	actions := types.GetItemActions(itemType, isEquipped)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"itemType": itemType,
		"actions":  actions,
	})
}

// Helper function to send error response
func sendInventoryError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(types.ItemActionResponse{
		Success: false,
		Error:   message,
	})
}
