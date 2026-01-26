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

// EquipmentSlots represents all equipment slots (12 slots in 3x4 grid)
// Layout: neck/head/ammo, mainhand/chest/offhand, ring1/legs/ring2, gloves/boots/bag
type EquipmentSlots struct {
	Neck     *InventoryItem `json:"neck"`
	Head     *InventoryItem `json:"head"`
	Ammo     *InventoryItem `json:"ammo"`
	Mainhand *InventoryItem `json:"mainhand"`
	Chest    *InventoryItem `json:"chest"`
	Offhand  *InventoryItem `json:"offhand"`
	Ring1    *InventoryItem `json:"ring1"`
	Legs     *InventoryItem `json:"legs"`
	Ring2    *InventoryItem `json:"ring2"`
	Gloves   *InventoryItem `json:"gloves"`
	Boots    *InventoryItem `json:"boots"`
	Bag      *InventoryItem `json:"bag"`
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
	Npub            string `json:"npub"`
	SaveID          string `json:"save_id"`
	ItemID          string `json:"item_id"`
	Action          string `json:"action"`             // "equip", "unequip", "use", "drop", "examine", "move", "add_to_container", "remove_from_container"
	FromSlot        int    `json:"from_slot"`          // Source slot (-1 for equipment)
	ToSlot          int    `json:"to_slot"`            // Destination slot (-1 for equipment)
	FromSlotType    string `json:"from_slot_type"`     // "general" or "inventory" (backpack)
	ToSlotType      string `json:"to_slot_type"`       // "general" or "inventory" (backpack)
	FromEquip       string `json:"from_equip"`         // Source equipment slot name (e.g., "mainHand")
	ToEquip         string `json:"to_equip"`           // Destination equipment slot name
	Quantity        int    `json:"quantity"`           // For stackable items
	ContainerSlot   int    `json:"container_slot"`     // Which inventory slot has the container
	ToContainerSlot int    `json:"to_container_slot"`  // Which slot within the container
}

// ItemActionResponse represents the result of an item action
type ItemActionResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Color    string      `json:"color,omitempty"`    // Color for message: 'red', 'green', 'yellow', 'white', 'purple', 'blue'
	NewState interface{} `json:"newState,omitempty"` // Updated inventory/equipment state
	Error    string      `json:"error,omitempty"`
}

// GetItemActions returns available actions for an item based on its tags and type
// Equipment items (those with "equipment" tag) can be equipped/unequipped
// Consumables can be used, other items default to examine
func GetItemActions(itemType string, isEquipped bool, hasEquipmentTag bool) []ItemAction {
	actions := []ItemAction{}

	// Equipment items (determined by "equipment" tag, not type)
	if hasEquipmentTag {
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
	} else {
		// Non-equipment items - check for consumables
		switch itemType {
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
	}

	// All items can be dropped (unless equipped)
	if !isEquipped {
		actions = append(actions, ItemAction{
			Action:      "drop",
			DisplayName: "Drop",
			IsDefault:   false,
		})
	}

	// Examine is always available (unless it's already the default action)
	if hasEquipmentTag || (itemType != "Potion" && itemType != "Consumable" && itemType != "Food") {
		actions = append(actions, ItemAction{
			Action:      "examine",
			DisplayName: "Examine",
			IsDefault:   false,
		})
	}

	return actions
}

// MapGearSlotToEquipmentSlot maps an item's gear_slot property to equipment slot(s)
// Returns the primary slot name. For slots with alternatives (hands, ring),
// the inventory handler determines the actual slot based on availability.
// Items must have the "equipment" tag to be equippable.
func MapGearSlotToEquipmentSlot(gearSlot string) string {
	switch gearSlot {
	// Hand slots
	case "hands":
		return "mainhand" // Primary slot, inventory handler tries offhand if full
	case "mainhand":
		return "mainhand"
	case "offhand":
		return "offhand"
	// Body slots
	case "chest", "armor", "body":
		return "chest"
	case "head", "helmet", "hat":
		return "head"
	case "legs", "leg", "greaves":
		return "legs"
	case "gloves", "glove", "gauntlets":
		return "gloves"
	case "boots", "boot", "feet":
		return "boots"
	// Accessory slots
	case "neck", "necklace", "amulet":
		return "neck"
	case "ring", "finger", "ring1", "ring2":
		return "ring1" // Primary slot, inventory handler tries ring2 if full
	// Other slots
	case "ammo", "ammunition":
		return "ammo"
	case "bag", "backpack":
		return "bag"
	default:
		return "" // Not a valid equipment slot
	}
}
