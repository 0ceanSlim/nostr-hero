package charactereditor

// StartingGearEntry represents the starting gear for a specific class
type StartingGearEntry struct {
	Class        string       `json:"class"`
	StartingGear StartingGear `json:"starting_gear"`
}

// StartingGear contains all equipment choices for a class
type StartingGear struct {
	EquipmentChoices []EquipmentChoice `json:"equipment_choices"`
	PackChoice       *PackChoice       `json:"pack_choice,omitempty"`
	GivenItems       []ItemQuantity    `json:"given_items"`
}

// EquipmentChoice represents a choice between multiple equipment options
type EquipmentChoice struct {
	Description string   `json:"description"`
	Options     []Option `json:"options"`
}

// Option represents one choice in an equipment selection
type Option struct {
	Type     string         `json:"type"` // "single", "bundle", "multi_slot"
	Item     string         `json:"item,omitempty"`
	Quantity int            `json:"quantity,omitempty"`
	Items    []ItemQuantity `json:"items,omitempty"` // For bundle
	Slots    []Slot         `json:"slots,omitempty"` // For multi_slot
}

// Slot represents a slot in a multi_slot option
type Slot struct {
	Type     string   `json:"type"` // "weapon_choice" or "fixed"
	Options  []string `json:"options,omitempty"`  // For weapon_choice
	Item     string   `json:"item,omitempty"`     // For fixed
	Quantity int      `json:"quantity,omitempty"` // For fixed
}

// ItemQuantity represents an item with a quantity
type ItemQuantity struct {
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

// PackChoice represents a choice between equipment packs
type PackChoice struct {
	Description string   `json:"description"`
	Options     []string `json:"options"`
}
