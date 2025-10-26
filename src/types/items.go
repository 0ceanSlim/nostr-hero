package types

// Item represents a game item with all its properties
type Item struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // e.g., "Weapon", "Armor", "Consumable", "Adventuring Gear"
	Weight      int                    `json:"weight"`
	Stack       int                    `json:"stack"`
	Rarity      string                 `json:"rarity"`
	Price       int                    `json:"price"`
	Image       string                 `json:"image"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// InventoryItem represents an item in a player's inventory with quantity
type InventoryItem struct {
	ItemID   string `json:"item"`
	Quantity int    `json:"quantity"`
	Slot     int    `json:"slot,omitempty"` // Position in inventory grid
}

// EquipmentSlots represents all equipment slots
type EquipmentSlots struct {
	MainHand *InventoryItem `json:"mainHand"`
	OffHand  *InventoryItem `json:"offHand"`
	Armor    *InventoryItem `json:"armor"`
	Helmet   *InventoryItem `json:"helmet"`
	Boots    *InventoryItem `json:"boots"`
	Gloves   *InventoryItem `json:"gloves"`
	Ring1    *InventoryItem `json:"ring1"`
	Ring2    *InventoryItem `json:"ring2"`
	Necklace *InventoryItem `json:"necklace"`
	Cloak    *InventoryItem `json:"cloak"`
}

// Inventory represents the complete inventory structure
type Inventory struct {
	GearSlots    *EquipmentSlots  `json:"gear_slots"`
	GeneralSlots []InventoryItem  `json:"general_slots"`
}

// ItemAction represents an action that can be performed on an item
type ItemAction struct {
	Action      string `json:"action"`      // "equip", "use", "drop", "examine"
	DisplayName string `json:"displayName"` // Human-readable name
	IsDefault   bool   `json:"isDefault"`   // Is this the default left-click action?
}

// ItemActionRequest represents a request to perform an action on an item
type ItemActionRequest struct {
	Npub         string `json:"npub"`
	SaveID       string `json:"save_id"`
	ItemID       string `json:"item_id"`
	Action       string `json:"action"`          // "equip", "unequip", "use", "drop", "examine", "move"
	FromSlot     int    `json:"from_slot"`       // Source slot (-1 for equipment)
	ToSlot       int    `json:"to_slot"`         // Destination slot (-1 for equipment)
	FromSlotType string `json:"from_slot_type"`  // "general" or "inventory" (backpack)
	ToSlotType   string `json:"to_slot_type"`    // "general" or "inventory" (backpack)
	FromEquip    string `json:"from_equip"`      // Source equipment slot name (e.g., "mainHand")
	ToEquip      string `json:"to_equip"`        // Destination equipment slot name
	Quantity     int    `json:"quantity"`        // For stackable items
}

// ItemActionResponse represents the result of an item action
type ItemActionResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	NewState interface{} `json:"newState,omitempty"` // Updated inventory/equipment state
	Error    string      `json:"error,omitempty"`
}

// GetItemActions returns available actions for an item based on its type
func GetItemActions(itemType string, isEquipped bool) []ItemAction {
	actions := []ItemAction{}

	// Add actions based on item type
	switch itemType {
	case "Weapon", "Melee Weapon", "Ranged Weapon", "Simple Weapon", "Martial Weapon":
		if !isEquipped {
			actions = append(actions, ItemAction{
				Action:      "equip",
				DisplayName: "Equip",
				IsDefault:   true,
			})
		} else {
			actions = append(actions, ItemAction{
				Action:      "unequip",
				DisplayName: "Unequip",
				IsDefault:   true,
			})
		}

	case "Armor", "Light Armor", "Medium Armor", "Heavy Armor", "Shield":
		if !isEquipped {
			actions = append(actions, ItemAction{
				Action:      "equip",
				DisplayName: "Equip",
				IsDefault:   true,
			})
		} else {
			actions = append(actions, ItemAction{
				Action:      "unequip",
				DisplayName: "Unequip",
				IsDefault:   true,
			})
		}

	case "Ring", "Necklace", "Amulet", "Cloak", "Boots", "Gloves", "Helmet", "Hat":
		if !isEquipped {
			actions = append(actions, ItemAction{
				Action:      "equip",
				DisplayName: "Equip",
				IsDefault:   true,
			})
		} else {
			actions = append(actions, ItemAction{
				Action:      "unequip",
				DisplayName: "Unequip",
				IsDefault:   true,
			})
		}

	case "Potion", "Consumable", "Food":
		actions = append(actions, ItemAction{
			Action:      "use",
			DisplayName: "Use",
			IsDefault:   true,
		})

	default:
		// Default action for other items is examine
		actions = append(actions, ItemAction{
			Action:      "examine",
			DisplayName: "Examine",
			IsDefault:   true,
		})
	}

	// All items can be examined and dropped (unless equipped)
	if !isEquipped {
		actions = append(actions, ItemAction{
			Action:      "drop",
			DisplayName: "Drop",
			IsDefault:   false,
		})
	}

	// Examine is always available
	if itemType != "Potion" && itemType != "Consumable" && itemType != "Food" {
		actions = append(actions, ItemAction{
			Action:      "examine",
			DisplayName: "Examine",
			IsDefault:   false,
		})
	}

	return actions
}

// DetermineEquipmentSlot determines which equipment slot an item should go into
func DetermineEquipmentSlot(itemType string) string {
	switch itemType {
	case "Weapon", "Melee Weapon", "Ranged Weapon", "Simple Weapon", "Martial Weapon":
		return "mainHand"
	case "Shield":
		return "offHand"
	case "Armor", "Light Armor", "Medium Armor", "Heavy Armor":
		return "armor"
	case "Helmet", "Hat":
		return "helmet"
	case "Boots":
		return "boots"
	case "Gloves":
		return "gloves"
	case "Ring":
		return "ring1" // Will need logic to choose ring1 or ring2
	case "Necklace", "Amulet":
		return "necklace"
	case "Cloak", "Cape":
		return "cloak"
	default:
		return "" // Not equippable
	}
}
