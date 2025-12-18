package api

import (
	"encoding/json"
	"fmt"
	"log"
	"nostr-hero/db"
)

// ============================================================================
// EQUIPMENT ACTION HANDLERS
// All equipping/unequipping logic for the in-memory session system
// ============================================================================

// handleEquipItemAction equips an item from inventory
func handleEquipItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	itemID, ok := params["item_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid item_id parameter")
	}

	// Get from_slot as int (can come as float64 from JSON)
	var fromSlot int
	if fs, ok := params["from_slot"].(float64); ok {
		fromSlot = int(fs)
	} else if fs, ok := params["from_slot"].(int); ok {
		fromSlot = fs
	} else {
		return nil, fmt.Errorf("missing or invalid from_slot parameter")
	}

	fromSlotType, _ := params["from_slot_type"].(string)
	if fromSlotType == "" {
		fromSlotType = "general" // Default
	}

	// Equipment slot can be provided or auto-determined
	equipSlot, _ := params["equipment_slot"].(string)
	isTwoHanded := false

	log.Printf("⚔️ Equip action: %s from %s[%d] to equipment slot", itemID, fromSlotType, fromSlot)

	// Get equipment slots
	gearSlots, ok := state.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		gearSlots = make(map[string]interface{})
		state.Inventory["gear_slots"] = gearSlots
	}

	// If no equipment slot specified, determine from item properties
	if equipSlot == "" {
		// Fetch item data from database to get gear_slot property
		database := db.GetDB()
		if database == nil {
			return nil, fmt.Errorf("database not available")
		}

		var propertiesJSON string
		var tagsJSON string
		var itemType string
		err := database.QueryRow("SELECT properties, tags, item_type FROM items WHERE id = ?", itemID).Scan(&propertiesJSON, &tagsJSON, &itemType)
		if err != nil {
			return nil, fmt.Errorf("item '%s' not found in database: %v", itemID, err)
		}

		var properties map[string]interface{}
		if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
			return nil, fmt.Errorf("failed to parse item properties")
		}

		// Check for two-handed
		var tags []interface{}
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err == nil {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok && tagStr == "two-handed" {
					isTwoHanded = true
					break
				}
			}
		}

		// Determine equipment slot
		if gearSlotProp, ok := properties["gear_slot"].(string); ok {
			switch gearSlotProp {
			case "hands":
				// For weapons and shields
				if itemType == "Shield" {
					equipSlot = "left_arm"
				} else {
					// Weapon - check availability
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

					// Choose slot
					if !rightOccupied {
						equipSlot = "right_arm"
					} else if !leftOccupied {
						equipSlot = "left_arm"
					} else {
						// Both full, swap with right
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
				equipSlot = gearSlotProp
			}
		} else {
			return nil, fmt.Errorf("item does not have a gear_slot property")
		}
	}

	log.Printf("⚔️ Equipping to slot: %s (two-handed: %v)", equipSlot, isTwoHanded)

	// Handle two-handed weapons - unequip both hands
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
		// Check if target slot is occupied
		if existing := gearSlots[equipSlot]; existing != nil {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				if existingMap["item"] != nil && existingMap["item"] != "" {
					existingItemID := existingMap["item"].(string)

					// Will swap with this item
					itemsToUnequip = append(itemsToUnequip, map[string]interface{}{
						"item":     existingItemID,
						"quantity": existingMap["quantity"],
						"from":     equipSlot,
					})

					// CRITICAL: Check if existing item is two-handed
					// If so, also unequip from the OTHER hand slot
					log.Printf("🔍 Checking if existing item '%s' is two-handed", existingItemID)
					database := db.GetDB()
					if database != nil {
						var tagsJSON string
						err := database.QueryRow("SELECT tags FROM items WHERE id = ?", existingItemID).Scan(&tagsJSON)
						if err == nil {
							log.Printf("🔍 Tags for %s: %s", existingItemID, tagsJSON)
							var tags []interface{}
							if err := json.Unmarshal([]byte(tagsJSON), &tags); err == nil {
								log.Printf("🔍 Parsed %d tags for %s", len(tags), existingItemID)
								for _, tag := range tags {
									if tagStr, ok := tag.(string); ok {
										log.Printf("🔍 Tag: %s", tagStr)
										if tagStr == "two-handed" {
											// Existing item is two-handed - also clear the other hand
											otherSlot := ""
											if equipSlot == "right_arm" {
												otherSlot = "left_arm"
											} else if equipSlot == "left_arm" {
												otherSlot = "right_arm"
											}

											if otherSlot != "" {
												log.Printf("🔍 Checking other slot: %s", otherSlot)
												if otherHand, ok := gearSlots[otherSlot].(map[string]interface{}); ok {
													log.Printf("🔍 Other hand has: %v", otherHand["item"])
													if otherHand["item"] == existingItemID {
														log.Printf("🗡️ Existing two-handed weapon detected - also clearing %s", otherSlot)
														// Don't add to itemsToUnequip (already added above)
														// Just clear it directly
														gearSlots[otherSlot] = map[string]interface{}{
															"item":     nil,
															"quantity": 0,
														}
													} else {
														log.Printf("⚠️ Other hand has different item: %v (expected %s)", otherHand["item"], existingItemID)
													}
												} else {
													log.Printf("⚠️ Could not get other hand slot map")
												}
											}
											break
										}
									}
								}
							} else {
								log.Printf("⚠️ Failed to parse tags JSON: %v", err)
							}
						} else {
							log.Printf("⚠️ Failed to query tags for %s: %v", existingItemID, err)
						}
					} else {
						log.Printf("⚠️ Database not available")
					}
				}
			}
		}
	}

	// Find item in inventory
	var itemData map[string]interface{}
	var sourceInventory []interface{}
	var sourceArrayIndex int

	if fromSlotType == "general" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid general_slots structure")
		}
		sourceInventory = generalSlots

		if fromSlot < 0 || fromSlot >= len(sourceInventory) {
			return nil, fmt.Errorf("invalid source slot")
		}

		itemMap, ok := sourceInventory[fromSlot].(map[string]interface{})
		if !ok || itemMap["item"] != itemID {
			return nil, fmt.Errorf("item not found in specified slot")
		}
		itemData = itemMap
		sourceArrayIndex = fromSlot

	} else if fromSlotType == "inventory" {
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no backpack found")
		}
		backpackContents, ok := bag["contents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid backpack contents")
		}
		sourceInventory = backpackContents

		// Search for item by slot number
		found := false
		for i, slotData := range sourceInventory {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
				var slotNum int
				if sn, ok := slotMap["slot"].(float64); ok {
					slotNum = int(sn)
				} else if sn, ok := slotMap["slot"].(int); ok {
					slotNum = sn
				}
				if slotNum == fromSlot && slotMap["item"] == itemID {
					itemData = slotMap
					sourceArrayIndex = i
					found = true
					break
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("item not found in specified slot")
		}
	} else {
		return nil, fmt.Errorf("invalid source slot type")
	}

	// Add unequipped items back to inventory (swapped items)
	for i, unequipData := range itemsToUnequip {
		var targetSlot int
		var targetInventory []interface{}

		if i == 0 {
			// First item goes in the source slot
			targetSlot = sourceArrayIndex
			targetInventory = sourceInventory
		} else {
			// Additional items (e.g., second weapon from two-handed swap) need empty slots
			// Try to find empty slot in same inventory
			foundEmpty := false
			for j, slot := range sourceInventory {
				if slotMap, ok := slot.(map[string]interface{}); ok {
					if slotMap["item"] == nil || slotMap["item"] == "" {
						targetSlot = j
						targetInventory = sourceInventory
						foundEmpty = true
						break
					}
				}
			}

			if !foundEmpty {
				// No empty slot found - need to find in backpack or general slots
				if fromSlotType == "general" {
					// Source was general, try backpack
					bag, ok := gearSlots["bag"].(map[string]interface{})
					if ok {
						backpackContents, ok := bag["contents"].([]interface{})
						if ok {
							for j := 0; j < 20; j++ {
								if j >= len(backpackContents) {
									backpackContents = append(backpackContents, map[string]interface{}{
										"item":     nil,
										"quantity": 0,
										"slot":     j,
									})
									bag["contents"] = backpackContents
								}
								if slotMap, ok := backpackContents[j].(map[string]interface{}); ok {
									if slotMap["item"] == nil || slotMap["item"] == "" {
										targetSlot = j
										targetInventory = backpackContents
										foundEmpty = true
										break
									}
								}
							}
						}
					}
				} else {
					// Source was backpack, try general slots
					generalSlots, ok := state.Inventory["general_slots"].([]interface{})
					if ok {
						for j := 0; j < 4; j++ {
							if j >= len(generalSlots) {
								generalSlots = append(generalSlots, map[string]interface{}{
									"item":     nil,
									"quantity": 0,
									"slot":     j,
								})
								state.Inventory["general_slots"] = generalSlots
							}
							if slotMap, ok := generalSlots[j].(map[string]interface{}); ok {
								if slotMap["item"] == nil || slotMap["item"] == "" {
									targetSlot = j
									targetInventory = generalSlots
									foundEmpty = true
									break
								}
							}
						}
					}
				}

				if !foundEmpty {
					return nil, fmt.Errorf("inventory is full, cannot unequip both items")
				}
			}
		}

		// Place the unequipped item in target slot
		targetInventory[targetSlot] = map[string]interface{}{
			"item":     unequipData["item"],
			"quantity": unequipData["quantity"],
			"slot":     targetSlot,
		}

		// Clear the equipment slot
		slotName := unequipData["from"].(string)
		gearSlots[slotName] = map[string]interface{}{
			"item":     nil,
			"quantity": 0,
		}

		log.Printf("✅ Unequipped %s from %s to inventory slot %d", unequipData["item"], slotName, targetSlot)
	}

	// Move item to equipment slot
	if isTwoHanded {
		// Two-handed weapons always go to right_arm and occupy both hands
		gearSlots["right_arm"] = map[string]interface{}{
			"item":     itemData["item"],
			"quantity": itemData["quantity"],
		}
		// Set left_arm to same item to show it's occupied
		gearSlots["left_arm"] = map[string]interface{}{
			"item":     itemData["item"],
			"quantity": itemData["quantity"],
		}
		log.Printf("✅ Equipped two-handed %s to both hands", itemData["item"])
	} else {
		// One-handed weapons go to specified slot
		equippedItem := map[string]interface{}{
			"item":     itemData["item"],
			"quantity": itemData["quantity"],
		}

		// CRITICAL: If equipping a bag, preserve its contents
		if equipSlot == "bag" {
			if contents, ok := itemData["contents"].([]interface{}); ok {
				equippedItem["contents"] = contents
				log.Printf("📦 Preserving bag contents when equipping (%d items)", len(contents))
			} else {
				// Initialize empty contents if none exist
				equippedItem["contents"] = make([]interface{}, 0, 20)
				log.Printf("📦 Initializing empty bag contents")
			}
		}

		gearSlots[equipSlot] = equippedItem
	}

	// Empty the source slot if no swap occurred
	if len(itemsToUnequip) == 0 {
		sourceInventory[sourceArrayIndex] = map[string]interface{}{
			"item":     nil,
			"quantity": 0,
			"slot":     fromSlot,
		}
	}

	// Save back to correct location
	if fromSlotType == "general" {
		state.Inventory["general_slots"] = sourceInventory
	} else {
		bag := gearSlots["bag"].(map[string]interface{})
		bag["contents"] = sourceInventory
	}

	if isTwoHanded {
		log.Printf("✅ Equipped %s to both hands (swapped %d items)", itemID, len(itemsToUnequip))
	} else {
		log.Printf("✅ Equipped %s to %s (swapped %d items)", itemID, equipSlot, len(itemsToUnequip))
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Equipped %s", itemID),
	}, nil
}

// handleUnequipItemAction unequips an item to inventory
func handleUnequipItemAction(state *SaveFile, params map[string]interface{}) (*GameActionResponse, error) {
	// Check both equipment_slot and from_equip for compatibility
	equipSlot, ok := params["equipment_slot"].(string)
	if !ok || equipSlot == "" {
		equipSlot, ok = params["from_equip"].(string)
		if !ok || equipSlot == "" {
			return nil, fmt.Errorf("missing or invalid equipment_slot/from_equip parameter")
		}
	}

	log.Printf("🛡️ Unequip action from equipment slot: %s", equipSlot)

	// Get equipment slots
	gearSlots, ok := state.Inventory["gear_slots"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid equipment structure")
	}

	// Get the item from equipment slot
	itemData, exists := gearSlots[equipSlot]
	if !exists || itemData == nil {
		return nil, fmt.Errorf("no item in equipment slot '%s'", equipSlot)
	}

	// Verify there's actually an item
	itemMap, ok := itemData.(map[string]interface{})
	if !ok || itemMap["item"] == nil || itemMap["item"] == "" {
		return nil, fmt.Errorf("no item in equipment slot '%s'", equipSlot)
	}

	itemID := itemMap["item"].(string)
	quantity := 1
	if qty, ok := itemMap["quantity"].(float64); ok {
		quantity = int(qty)
	} else if qty, ok := itemMap["quantity"].(int); ok {
		quantity = qty
	}

	log.Printf("🛡️ Unequipping: %s (quantity: %d)", itemID, quantity)

	// SPECIAL CASE: Unequipping the bag itself
	// The bag can ONLY go to general slots (not into itself!)
	emptySlotIndex := -1
	emptySlotType := ""

	if equipSlot == "bag" {
		log.Printf("🎒 Special case: Unequipping bag - can ONLY go to general slots")

		// Search ONLY general slots for the bag
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			generalSlots = make([]interface{}, 0, 4)
			state.Inventory["general_slots"] = generalSlots
		}

		for i := 0; i < 4; i++ {
			// Extend array if needed
			if i >= len(generalSlots) {
				generalSlots = append(generalSlots, map[string]interface{}{
					"item":     nil,
					"quantity": 0,
					"slot":     i,
				})
				state.Inventory["general_slots"] = generalSlots
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

		// If no empty general slot found, bag cannot be unequipped
		if emptySlotIndex == -1 {
			return &GameActionResponse{
				Success: false,
				Error:   "There isn't enough room in your inventory to take off the bag",
				Color:   "red",
			}, nil
		}

		log.Printf("✅ Found empty general slot at index %d for bag", emptySlotIndex)
	} else {
		// NORMAL CASE: Unequipping other items
		// Try backpack first, then general slots

		bag, ok := gearSlots["bag"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no backpack found")
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
	}

	// If no empty backpack slot (for normal items), check general slots
	if emptySlotIndex == -1 && equipSlot != "bag" {
		generalSlots, ok := state.Inventory["general_slots"].([]interface{})
		if !ok {
			generalSlots = make([]interface{}, 0, 4)
			state.Inventory["general_slots"] = generalSlots
		}

		for i := 0; i < 4; i++ {
			// Extend array if needed
			if i >= len(generalSlots) {
				generalSlots = append(generalSlots, map[string]interface{}{
					"item":     nil,
					"quantity": 0,
					"slot":     i,
				})
				state.Inventory["general_slots"] = generalSlots
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
		return nil, fmt.Errorf("your inventory is full")
	}

	// Place item in the empty slot
	newItem := map[string]interface{}{
		"item":     itemID,
		"quantity": quantity,
		"slot":     emptySlotIndex,
	}

	// CRITICAL: If unequipping a bag, preserve its contents
	if equipSlot == "bag" {
		if contents, ok := itemMap["contents"].([]interface{}); ok {
			newItem["contents"] = contents
			log.Printf("📦 Preserving bag contents when unequipping (%d items)", len(contents))
		}
	}

	if emptySlotType == "inventory" {
		bag := gearSlots["bag"].(map[string]interface{})
		backpackContents := bag["contents"].([]interface{})
		backpackContents[emptySlotIndex] = newItem
		bag["contents"] = backpackContents
		log.Printf("✅ Moved to backpack slot %d", emptySlotIndex)
	} else {
		generalSlots := state.Inventory["general_slots"].([]interface{})
		generalSlots[emptySlotIndex] = newItem
		state.Inventory["general_slots"] = generalSlots
		log.Printf("✅ Moved to general slot %d", emptySlotIndex)
	}

	// Empty the equipment slot
	gearSlots[equipSlot] = map[string]interface{}{
		"item":     nil,
		"quantity": 0,
	}

	// Check if this is a two-handed weapon (item in both hands)
	// If so, clear the other hand too
	if equipSlot == "right_arm" {
		if leftArm, ok := gearSlots["left_arm"].(map[string]interface{}); ok {
			if leftArm["item"] == itemID {
				gearSlots["left_arm"] = map[string]interface{}{
					"item":     nil,
					"quantity": 0,
				}
				log.Printf("✅ Also cleared left_arm (two-handed weapon)")
			}
		}
	} else if equipSlot == "left_arm" {
		if rightArm, ok := gearSlots["right_arm"].(map[string]interface{}); ok {
			if rightArm["item"] == itemID {
				gearSlots["right_arm"] = map[string]interface{}{
					"item":     nil,
					"quantity": 0,
				}
				log.Printf("✅ Also cleared right_arm (two-handed weapon)")
			}
		}
	}

	return &GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Unequipped %s", itemID),
	}, nil
}
