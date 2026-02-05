package gameutil

import (
	"fmt"
	"log"

	"pubkey-quest/types"
)

// GetGoldQuantity returns the total gold quantity from inventory
func GetGoldQuantity(state *types.SaveFile) int {
	totalGold := 0

	// Check general slots
	if generalSlots, ok := state.Inventory["general_slots"].([]interface{}); ok {
		for _, slotData := range generalSlots {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
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
	if gearSlots, ok := state.Inventory["gear_slots"].(map[string]interface{}); ok {
		if bag, ok := gearSlots["bag"].(map[string]interface{}); ok {
			if contents, ok := bag["contents"].([]interface{}); ok {
				for _, slotData := range contents {
					if slotMap, ok := slotData.(map[string]interface{}); ok {
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

// DeductGold deducts gold from inventory (returns true if successful)
func DeductGold(state *types.SaveFile, amount int) bool {
	if amount <= 0 {
		return true
	}

	remaining := amount

	// First, deduct from general slots
	if generalSlots, ok := state.Inventory["general_slots"].([]interface{}); ok {
		for _, slotData := range generalSlots {
			if remaining <= 0 {
				break
			}
			if slotMap, ok := slotData.(map[string]interface{}); ok {
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
		if gearSlots, ok := state.Inventory["gear_slots"].(map[string]interface{}); ok {
			if bag, ok := gearSlots["bag"].(map[string]interface{}); ok {
				if contents, ok := bag["contents"].([]interface{}); ok {
					for _, slotData := range contents {
						if remaining <= 0 {
							break
						}
						if slotMap, ok := slotData.(map[string]interface{}); ok {
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

// PlayerHasItem checks if player has an item in inventory
func PlayerHasItem(state *types.SaveFile, itemID string) bool {
	// Check general slots
	if generalSlots, ok := state.Inventory["general_slots"].([]interface{}); ok {
		for _, slotData := range generalSlots {
			if slotMap, ok := slotData.(map[string]interface{}); ok {
				if item, ok := slotMap["item"].(string); ok && item == itemID {
					return true
				}
			}
		}
	}

	// Check backpack
	if gearSlots, ok := state.Inventory["gear_slots"].(map[string]interface{}); ok {
		if bag, ok := gearSlots["bag"].(map[string]interface{}); ok {
			if contents, ok := bag["contents"].([]interface{}); ok {
				for _, slotData := range contents {
					if slotMap, ok := slotData.(map[string]interface{}); ok {
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

// AddGoldToInventory adds gold to the player's inventory
func AddGoldToInventory(inventory map[string]interface{}, goldAmount int) error {
	log.Printf("💰 Adding %dg to inventory", goldAmount)

	// Try general_slots first (type is []interface{})
	generalSlotsRaw, ok := inventory["general_slots"].([]interface{})
	if !ok {
		log.Printf("❌ general_slots type assertion failed, got type: %T", inventory["general_slots"])
		return fmt.Errorf("invalid general_slots format")
	}

	// Try to find existing gold stack or empty slot in general slots
	for i, slotData := range generalSlotsRaw {
		slot, ok := slotData.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if slot has gold already - add to it
		if slot["item"] == "gold-piece" {
			currentQty := 0
			if qty, ok := slot["quantity"].(float64); ok {
				currentQty = int(qty)
			} else if qty, ok := slot["quantity"].(int); ok {
				currentQty = qty
			}
			slot["quantity"] = currentQty + goldAmount
			log.Printf("✅ Added %dg to existing gold stack in general_slots[%d] (new total: %d)", goldAmount, i, currentQty+goldAmount)
			return nil
		}

		// Found empty slot - add gold here
		if slot["item"] == nil || slot["item"] == "" {
			slot["item"] = "gold-piece"
			slot["quantity"] = goldAmount
			log.Printf("✅ Added %dg to general_slots[%d]", goldAmount, i)
			return nil
		}
	}

	// If general slots are full, try backpack
	gearSlots, ok := inventory["gear_slots"].(map[string]interface{})
	if ok {
		bag, ok := gearSlots["bag"].(map[string]interface{})
		if ok {
			contents, ok := bag["contents"].([]interface{})
			if ok {
				for i, slotData := range contents {
					slot, ok := slotData.(map[string]interface{})
					if !ok {
						continue
					}

					// Check if slot has gold already - add to it
					if slot["item"] == "gold-piece" {
						currentQty := 0
						if qty, ok := slot["quantity"].(float64); ok {
							currentQty = int(qty)
						} else if qty, ok := slot["quantity"].(int); ok {
							currentQty = qty
						}
						slot["quantity"] = currentQty + goldAmount
						log.Printf("✅ Added %dg to existing gold stack in backpack[%d] (new total: %d)", goldAmount, i, currentQty+goldAmount)
						return nil
					}

					// Found empty slot
					if slot["item"] == nil || slot["item"] == "" {
						slot["item"] = "gold-piece"
						slot["quantity"] = goldAmount
						log.Printf("✅ Added %dg to backpack[%d]", goldAmount, i)
						return nil
					}
				}
			}
		}
	}

	log.Printf("❌ No empty slots available for gold")
	return fmt.Errorf("no empty slots available for gold")
}
