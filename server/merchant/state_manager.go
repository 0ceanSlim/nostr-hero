package merchant

import (
	"log"
	"sync"
	"time"
)

// MerchantInventoryItem represents current stock for a single item
type MerchantInventoryItem struct {
	ItemID       string `json:"item_id"`
	CurrentStock int    `json:"current_stock"`
	MaxStock     int    `json:"max_stock"`
}

// MerchantState represents the current state of a merchant for a specific player
type MerchantState struct {
	MerchantID      string                           `json:"merchant_id"`
	CurrentGold     int                              `json:"current_gold"`
	Inventory       map[string]*MerchantInventoryItem `json:"inventory"` // item_id -> stock info
	LastRestock     time.Time                        `json:"last_restock"`
	RestockInterval int                              `json:"restock_interval"` // Minutes
}

// MerchantStateManager manages merchant states per player
type MerchantStateManager struct {
	// Per-player merchant states: playerNpub -> merchantID -> MerchantState
	states map[string]map[string]*MerchantState
	mu     sync.RWMutex

	// Future: Global item tracking (specific items shared across all players)
	// globalItems map[string]int // item_id -> global_stock
}

var manager *MerchantStateManager
var once sync.Once

// GetManager returns the singleton merchant state manager
func GetManager() *MerchantStateManager {
	once.Do(func() {
		manager = &MerchantStateManager{
			states: make(map[string]map[string]*MerchantState),
		}
		log.Println("🏪 Merchant State Manager initialized")
	})
	return manager
}

// GetMerchantState gets or initializes merchant state for a player
// Returns the state and whether a restock occurred
func (m *MerchantStateManager) GetMerchantState(npub, merchantID string, initialGold int, initialInventory []MerchantInventoryItem, restockInterval int) (*MerchantState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize player's merchant map if needed
	if m.states[npub] == nil {
		m.states[npub] = make(map[string]*MerchantState)
	}

	// Check if merchant state exists for this player
	state, exists := m.states[npub][merchantID]
	restocked := false

	if !exists {
		// First time seeing this merchant for this player - initialize
		log.Printf("🆕 Initializing merchant state: %s for player %s", merchantID, npub[:12])
		state = &MerchantState{
			MerchantID:      merchantID,
			CurrentGold:     initialGold,
			Inventory:       make(map[string]*MerchantInventoryItem),
			LastRestock:     time.Now(),
			RestockInterval: restockInterval,
		}

		// Copy initial inventory
		for _, item := range initialInventory {
			state.Inventory[item.ItemID] = &MerchantInventoryItem{
				ItemID:       item.ItemID,
				CurrentStock: item.CurrentStock,
				MaxStock:     item.MaxStock,
			}
		}

		m.states[npub][merchantID] = state
	} else {
		// Check if restock is due (based on real-world time)
		minutesSinceRestock := time.Since(state.LastRestock).Minutes()
		if minutesSinceRestock >= float64(state.RestockInterval) {
			log.Printf("🔄 Restocking merchant: %s for player %s (%.0f min since last restock)", merchantID, npub[:12], minutesSinceRestock)
			m.restockMerchant(state, initialGold, initialInventory)
			restocked = true
		}
	}

	return state, restocked
}

// restockMerchant restores merchant inventory and gold
func (m *MerchantStateManager) restockMerchant(state *MerchantState, restoreGold int, restoreInventory []MerchantInventoryItem) {
	// Restore gold to starting value
	state.CurrentGold = restoreGold

	// Restore inventory to max stock
	for _, item := range restoreInventory {
		if invItem, exists := state.Inventory[item.ItemID]; exists {
			invItem.CurrentStock = invItem.MaxStock
		} else {
			// Item was added to shop config since last time - add it
			state.Inventory[item.ItemID] = &MerchantInventoryItem{
				ItemID:       item.ItemID,
				CurrentStock: item.CurrentStock,
				MaxStock:     item.MaxStock,
			}
		}
	}

	state.LastRestock = time.Now()
	log.Printf("✅ Merchant restocked: %s (Gold: %d)", state.MerchantID, state.CurrentGold)
}

// UpdateMerchantInventory updates stock and gold after a transaction
func (m *MerchantStateManager) UpdateMerchantInventory(npub, merchantID, itemID string, quantityChange int, goldChange int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[npub][merchantID]
	if !exists {
		log.Printf("❌ Merchant state not found: %s for player %s", merchantID, npub[:12])
		return nil // Silently fail - will reinitialize next time
	}

	// Update stock
	if invItem, exists := state.Inventory[itemID]; exists {
		invItem.CurrentStock += quantityChange
		if invItem.CurrentStock < 0 {
			invItem.CurrentStock = 0
		}
	}

	// Update gold
	state.CurrentGold += goldChange

	log.Printf("💰 Merchant updated: %s | Item: %s (%+d) | Gold: %d (%+d)", merchantID, itemID, quantityChange, state.CurrentGold, goldChange)
	return nil
}

// GetTimeUntilRestock returns minutes until next restock
func (m *MerchantStateManager) GetTimeUntilRestock(npub, merchantID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.states[npub] == nil || m.states[npub][merchantID] == nil {
		return 0
	}

	state := m.states[npub][merchantID]
	elapsed := time.Since(state.LastRestock).Minutes()
	remaining := float64(state.RestockInterval) - elapsed

	if remaining < 0 {
		return 0
	}
	return remaining
}

// CleanupPlayerStates removes merchant states for a player (called when player logs out)
func (m *MerchantStateManager) CleanupPlayerStates(npub string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, npub)
	log.Printf("🧹 Cleaned up merchant states for player: %s", npub[:12])
}

// GetAllMerchantStatesForPlayer returns all merchant states for debugging
func (m *MerchantStateManager) GetAllMerchantStatesForPlayer(npub string) map[string]*MerchantState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.states[npub] == nil {
		return nil
	}

	// Return a copy to prevent external modification
	copy := make(map[string]*MerchantState)
	for k, v := range m.states[npub] {
		copy[k] = v
	}
	return copy
}
