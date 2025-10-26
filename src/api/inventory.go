package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// handleDropItem removes an item from inventory
func handleDropItem(save *SaveFile, req *types.ItemActionRequest) *types.ItemActionResponse {
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
	for i, item := range inventory {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemMap["item"] == req.ItemID {
				itemSlot = i
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

	// Remove item from inventory
	newInventory := make([]interface{}, 0, len(inventory)-1)
	for i, item := range inventory {
		if i != itemSlot {
			newInventory = append(newInventory, item)
		}
	}
	save.Inventory["general_slots"] = newInventory

	return &types.ItemActionResponse{
		Success: true,
		Message: "Item dropped",
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

		// Get source and destination slot objects
		fromSlotObj := fromInventory[req.FromSlot].(map[string]interface{})
		toSlotObj := fromInventory[req.ToSlot].(map[string]interface{})

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
			toInventory = append(toInventory, nil)
		}

		// If destination slot has an item, swap them
		if req.ToSlot < len(toInventory) && toInventory[req.ToSlot] != nil {
			fromInventory[req.FromSlot], toInventory[req.ToSlot] = toInventory[req.ToSlot], fromInventory[req.FromSlot]
		} else {
			// Just move the item
			toInventory[req.ToSlot] = item
			fromInventory[req.FromSlot] = nil
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
