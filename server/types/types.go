package types

type RegistryEntry struct {
	Npub      string    `json:"npub"`
	PubKey    string    `json:"pubkey"`
	Character Character `json:"character"`
}

type RaceClassWeight struct {
	Race   string
	Class  string
	Weight int
}

type RaceBackgroundWeight struct {
	Race       string
	Background string
	Weight     int
}

type Character struct {
	Race       string         `json:"race"`
	Class      string         `json:"class"`
	Background string         `json:"background"`
	Alignment  string         `json:"alignment"`
	Stats      map[string]int `json:"stats"`
}

type WeightData struct {
	Races                    []string                  `json:"Races"`
	RaceWeights              []int                     `json:"RaceWeights"`
	ClassWeightsByRace       map[string]map[string]int `json:"classWeightsByRace"`
	BackgroundWeightsByClass map[string]map[string]int `json:"backgroundWeightsByClass"`
	Alignments               []string                  `json:"Alignments"`
	AlignmentWeights         []int                     `json:"AlignmentWeights"`
}

// Weighted option structure
type WeightedOption struct {
	Name   string
	Weight int
}

// NPCScheduleSlot represents a time period in an NPC's schedule
type NPCScheduleSlot struct {
	Start            int      `json:"start"`              // Minutes from midnight (0-1439)
	End              int      `json:"end"`                // Minutes from midnight (0-1439, wraps to next day if < start)
	LocationType     string   `json:"location_type"`      // "building" or "district"
	LocationID       string   `json:"location_id"`        // Building ID or district ID
	State            string   `json:"state"`              // "sleeping", "working", "traveling", "home"
	DialogueOptions  []string `json:"dialogue_options"`   // Which dialogue nodes are available
	AvailableActions []string `json:"available_actions"`  // Which actions can be performed (open_shop, etc.)
}

// NPCData represents the full NPC structure from database
type NPCData struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Title         string                 `json:"title,omitempty"`
	Race          string                 `json:"race,omitempty"`
	Location      string                 `json:"location,omitempty"`      // Default location (backward compat)
	Building      string                 `json:"building,omitempty"`      // Default building (backward compat)
	Description   string                 `json:"description,omitempty"`
	Greeting      map[string]string      `json:"greeting,omitempty"`
	Dialogue      map[string]interface{} `json:"dialogue,omitempty"`
	Schedule      []NPCScheduleSlot      `json:"schedule,omitempty"`      // Optional schedule
	ShopConfig    map[string]interface{} `json:"shop_config,omitempty"`
	StorageConfig map[string]interface{} `json:"storage_config,omitempty"`
	InnConfig     map[string]interface{} `json:"inn_config,omitempty"`
}

// NPCScheduleInfo represents the resolved schedule state for an NPC at a given time
type NPCScheduleInfo struct {
	CurrentSlot       *NPCScheduleSlot `json:"current_slot"`
	IsAvailable       bool             `json:"is_available"`
	LocationType      string           `json:"location_type"`
	LocationID        string           `json:"location_id"`
	State             string           `json:"state"`
	AvailableDialogue []string         `json:"available_dialogue"`
	AvailableActions  []string         `json:"available_actions"`
}

// Shop-related types

// ShopInventoryItem represents an item in a shop's inventory
type ShopInventoryItem struct {
	ItemID          string `json:"item_id"`
	Stock           int    `json:"stock"`
	MaxStock        int    `json:"max_stock"`
	RestockRate     int    `json:"restock_rate"`
	RestockInterval string `json:"restock_interval"`
}

// ShopConfig represents the static configuration from NPC JSON
type ShopConfig struct {
	ShopType            string              `json:"shop_type"`
	BuysItems           bool                `json:"buys_items"`
	BuyPriceMultiplier  float64             `json:"buy_price_multiplier"`  // What merchant pays player
	SellPriceMultiplier float64             `json:"sell_price_multiplier"` // What player pays merchant
	StartingGold        int                 `json:"starting_gold"`
	MaxGold             int                 `json:"max_gold"`
	GoldRegenRate       int                 `json:"gold_regen_rate"`
	GoldRegenInterval   string              `json:"gold_regen_interval"`
	Inventory           []ShopInventoryItem `json:"inventory"`
}

// ShopState represents the runtime state of a shop (stored in save file)
type ShopState struct {
	MerchantID  string         `json:"merchant_id"`
	CurrentGold int            `json:"current_gold"`
	LastRegen   string         `json:"last_regen"` // ISO timestamp
	Stock       map[string]int `json:"stock"`      // item_id -> current quantity
}

// ShopTransaction represents a buy or sell transaction
type ShopTransaction struct {
	Npub       string `json:"npub"`
	SaveID     string `json:"save_id"`
	MerchantID string `json:"merchant_id"`
	ItemID     string `json:"item_id"`
	Quantity   int    `json:"quantity"`
	Action     string `json:"action"` // "buy" or "sell"
}