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
	default:
		response = &types.ItemActionResponse{
			Success: false,
			Error:   "Unknown action: " + req.Action,
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
	// Get inventory from save
	inventory, ok := save.Inventory["general_slots"].([]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid inventory structure",
		}
	}

	// Get equipment slots
	gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		gearSlots = make(map[string]interface{})
		save.Inventory["gear_slots"] = gearSlots
	}

	// Find item in inventory
	var itemSlot int = -1
	var itemData map[string]interface{}
	for i, item := range inventory {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemMap["item"] == req.ItemID {
				itemSlot = i
				itemData = itemMap
				break
			}
		}
	}

	if itemSlot == -1 {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Item not found in inventory",
		}
	}

	// Determine which equipment slot this item should go into
	equipSlot := req.ToEquip
	if equipSlot == "" {
		// Auto-determine slot based on item type
		// This would require fetching item data from database
		equipSlot = "mainHand" // Default for now
	}

	// Check if slot is occupied
	if existing := gearSlots[equipSlot]; existing != nil {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Equipment slot '%s' is already occupied", equipSlot),
			Message: "Unequip the current item first",
		}
	}

	// Move item from inventory to equipment
	gearSlots[equipSlot] = itemData

	// Remove from inventory (maintain array structure)
	newInventory := make([]interface{}, 0, len(inventory)-1)
	for i, item := range inventory {
		if i != itemSlot {
			newInventory = append(newInventory, item)
		}
	}
	save.Inventory["general_slots"] = newInventory

	return &types.ItemActionResponse{
		Success: true,
		Message: "Item equipped successfully",
		NewState: save.Inventory,
	}
}

// handleUnequipItem moves an item from equipment slot to inventory
func handleUnequipItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	// Get inventory from save
	inventory, ok := save.Inventory["general_slots"].([]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid inventory structure",
		}
	}

	// Check if inventory has space (max 20 slots)
	if len(inventory) >= 20 {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Inventory is full",
		}
	}

	// Get equipment slots
	gearSlots, ok := save.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid equipment structure",
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

	// Add to inventory
	inventory = append(inventory, itemData)
	save.Inventory["general_slots"] = inventory

	// Remove from equipment slot
	gearSlots[equipSlot] = nil

	return &types.ItemActionResponse{
		Success: true,
		Message: "Item unequipped successfully",
		NewState: save.Inventory,
	}
}

// handleUseItem uses a consumable item
func handleUseItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
	// Get inventory from save
	inventory, ok := save.Inventory["general_slots"].([]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Invalid inventory structure",
		}
	}

	// Find item in inventory
	var itemSlot int = -1
	var itemData map[string]interface{}
	for i, item := range inventory {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemMap["item"] == req.ItemID {
				itemSlot = i
				itemData = itemMap
				break
			}
		}
	}

	if itemSlot == -1 {
		return &types.ItemActionResponse{
			Success: false,
			Error:   "Item not found in inventory",
		}
	}

	// TODO: Apply item effects (healing, buffs, etc.)
	// For now, just remove the item

	// Handle quantity
	quantity := int(itemData["quantity"].(float64))
	if quantity > 1 {
		itemData["quantity"] = quantity - 1
		inventory[itemSlot] = itemData
	} else {
		// Remove item completely
		newInventory := make([]interface{}, 0, len(inventory)-1)
		for i, item := range inventory {
			if i != itemSlot {
				newInventory = append(newInventory, item)
			}
		}
		inventory = newInventory
	}

	save.Inventory["general_slots"] = inventory

	return &types.ItemActionResponse{
		Success: true,
		Message: "Item used",
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

	// Validate source slot has an item
	if req.FromSlot < 0 || req.FromSlot >= len(fromInventory) {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid source slot: %d", req.FromSlot),
		}
	}

	// Check if the slot object exists
	if fromInventory[req.FromSlot] == nil {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Source slot %d is empty", req.FromSlot),
		}
	}

	// The inventory stores items as objects with {item, quantity, slot} fields
	// Check if the slot's "item" field is nil
	slotObj, ok := fromInventory[req.FromSlot].(map[string]interface{})
	if !ok {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid slot structure at %d", req.FromSlot),
		}
	}

	itemID, _ := slotObj["item"].(string)
	if itemID == "" || slotObj["item"] == nil {
		return &types.ItemActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Source slot %d is empty", req.FromSlot),
		}
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

		// Get source and destination slot objects (handle nil slots)
		var fromSlotObj map[string]interface{}
		if fromInventory[req.FromSlot] != nil {
			fromSlotObj = fromInventory[req.FromSlot].(map[string]interface{})
		} else {
			// Create empty slot object if nil
			fromSlotObj = map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     req.FromSlot,
			}
			fromInventory[req.FromSlot] = fromSlotObj
		}

		var toSlotObj map[string]interface{}
		if fromInventory[req.ToSlot] != nil {
			toSlotObj = fromInventory[req.ToSlot].(map[string]interface{})
		} else {
			// Create empty slot object if nil
			toSlotObj = map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     req.ToSlot,
			}
			fromInventory[req.ToSlot] = toSlotObj
		}

		// Swap item and quantity, but keep slot numbers correct
		fromSlotObj["item"], toSlotObj["item"] = toSlotObj["item"], fromSlotObj["item"]
		fromSlotObj["quantity"], toSlotObj["quantity"] = toSlotObj["quantity"], fromSlotObj["quantity"]

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
		item := fromInventory[req.FromSlot]

		// Extend destination array if needed
		for len(toInventory) <= req.ToSlot {
			// Create proper empty slot object instead of nil
			emptySlot := map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     len(toInventory),
			}
			toInventory = append(toInventory, emptySlot)
		}

		// If destination slot has an item, swap them
		if req.ToSlot < len(toInventory) && toInventory[req.ToSlot] != nil {
			// Check if destination slot actually has an item
			if destSlot, ok := toInventory[req.ToSlot].(map[string]interface{}); ok {
				if destSlot["item"] != nil && destSlot["item"] != "" {
					// Swap the items
					fromInventory[req.FromSlot], toInventory[req.ToSlot] = toInventory[req.ToSlot], fromInventory[req.FromSlot]
				} else {
					// Destination is empty, just move
					toInventory[req.ToSlot] = item
					fromInventory[req.FromSlot] = map[string]interface{}{
						"item":     nil,
						"quantity": 0,
						"slot":     req.FromSlot,
					}
				}
			}
		} else {
			// Just move the item
			toInventory[req.ToSlot] = item
			fromInventory[req.FromSlot] = map[string]interface{}{
				"item":     nil,
				"quantity": 0,
				"slot":     req.FromSlot,
			}
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
		Success: true,
		Message: "Item moved",
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
