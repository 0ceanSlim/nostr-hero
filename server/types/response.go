package types

// GameActionResponse is returned after processing an action
type GameActionResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Color   string                 `json:"color,omitempty"` // Message color (red, green, yellow, white, purple, blue)
	State   *SaveFile              `json:"state,omitempty"` // Updated game state
	Delta   map[string]interface{} `json:"delta,omitempty"` // Only changed fields (for optimization)
	Data    map[string]interface{} `json:"data,omitempty"`  // Additional response data
	Error   string                 `json:"error,omitempty"`
}

// EffectMessage contains the message to display when an effect is applied
type EffectMessage struct {
	Message  string
	Color    string
	Category string
	Silent   bool
}

// GameAction represents any action a player can take
type GameAction struct {
	Type   string                 `json:"type"`   // "move", "use_item", "equip", "cast_spell", etc.
	Params map[string]interface{} `json:"params"` // Action-specific parameters
}

// NPCDeltaInfo provides access to NPC delta information
type NPCDeltaInfo struct {
	Added   []string
	Removed []string
}

// DeltaProvider is an interface for objects that can convert to a map (used for delta updates)
type DeltaProvider interface {
	ToMap() map[string]interface{}
	IsEmpty() bool
	GetNPCs() *NPCDeltaInfo
}
