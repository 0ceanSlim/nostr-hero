package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"nostr-hero/db"
	"nostr-hero/merchant"
	"nostr-hero/types"
)

// ShopHandler handles shop-related operations
// Routes:
// GET /api/shop/{merchant_id} - Get shop data (config + inventory with prices)
// POST /api/shop/buy - Buy items from shop
// POST /api/shop/sell - Sell items to shop
func ShopHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/shop/"), "/")

	switch r.Method {
	case "GET":
		if len(pathParts) > 0 && pathParts[0] != "" {
			handleGetShop(w, r, pathParts[0])
		} else {
			http.Error(w, "Missing merchant ID", http.StatusBadRequest)
		}
	case "POST":
		if len(pathParts) > 0 {
			switch pathParts[0] {
			case "buy":
				handleBuyFromShop(w, r)
			case "sell":
				handleSellToShop(w, r)
			default:
				http.Error(w, "Invalid shop action", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Missing action", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get shop data including inventory with prices
func handleGetShop(w http.ResponseWriter, r *http.Request, merchantID string) {
	// Get npub from query parameter
	npub := r.URL.Query().Get("npub")
	if npub == "" {
		http.Error(w, "Missing npub parameter", http.StatusBadRequest)
		return
	}

	log.Printf("📂 Loading shop data for merchant: %s (player: %s)", merchantID, npub[:12])

	// Get NPC data from database
	npcData, err := db.GetNPCByID(merchantID)
	if err != nil {
		log.Printf("❌ Error loading NPC: %v", err)
		http.Error(w, "Merchant not found", http.StatusNotFound)
		return
	}

	// Parse shop config - it's already in the ShopConfig field
	configJSON, err := json.Marshal(npcData.ShopConfig)
	if err != nil {
		log.Printf("❌ Error marshaling shop config: %v", err)
		http.Error(w, "Invalid shop configuration", http.StatusInternalServerError)
		return
	}

	var shopConfig types.ShopConfig
	if err := json.Unmarshal(configJSON, &shopConfig); err != nil {
		log.Printf("❌ Error parsing shop config: %v", err)
		http.Error(w, "Invalid shop configuration", http.StatusInternalServerError)
		return
	}

	// Initialize merchant inventory items for state manager
	initialInventory := make([]merchant.MerchantInventoryItem, 0)
	for _, invItem := range shopConfig.Inventory {
		initialInventory = append(initialInventory, merchant.MerchantInventoryItem{
			ItemID:       invItem.ItemID,
			CurrentStock: invItem.Stock,
			MaxStock:     invItem.MaxStock,
		})
	}

	// Get or initialize merchant state for this player
	restockInterval := 1440 // Default: 24 hours (in minutes)
	if shopConfig.RestockInterval > 0 {
		restockInterval = shopConfig.RestockInterval
	}

	merchantManager := merchant.GetManager()
	merchantState, restocked := merchantManager.GetMerchantState(npub, merchantID, shopConfig.StartingGold, initialInventory, restockInterval)

	// Get item prices and current stock from merchant state
	itemsWithPrices := make([]map[string]any, 0)
	for _, invItem := range shopConfig.Inventory {
		item, err := db.GetItemByID(invItem.ItemID)
		if err != nil {
			log.Printf("⚠️ Item not found: %s", invItem.ItemID)
			continue
		}

		// Get current stock from merchant state
		currentStock := 0
		if stateItem, exists := merchantState.Inventory[invItem.ItemID]; exists {
			currentStock = stateItem.CurrentStock
		}

		// Calculate buy/sell prices
		basePrice := item.Value
		buyPrice := int(float64(basePrice) * shopConfig.SellPriceMultiplier)  // Player pays this
		sellPrice := int(float64(basePrice) * shopConfig.BuyPriceMultiplier)  // Merchant pays player this

		itemsWithPrices = append(itemsWithPrices, map[string]any{
			"item_id":     invItem.ItemID,
			"name":        item.Name,
			"description": item.Description,
			"type":        item.Type,
			"value":       basePrice,
			"buy_price":   buyPrice,   // What player pays to buy
			"sell_price":  sellPrice,  // What player gets when selling
			"stock":       currentStock, // Current stock from merchant state
			"max_stock":   invItem.MaxStock,
		})
	}

	// Calculate time until next restock
	timeUntilRestock := merchantManager.GetTimeUntilRestock(npub, merchantID)

	response := map[string]any{
		"merchant_id":           merchantID,
		"merchant_name":         npcData.Name,
		"shop_type":             shopConfig.ShopType,
		"buys_items":            shopConfig.BuysItems,
		"current_gold":          merchantState.CurrentGold, // Current gold from state
		"max_gold":              shopConfig.MaxGold,
		"buy_price_multiplier":  shopConfig.BuyPriceMultiplier,
		"sell_price_multiplier": shopConfig.SellPriceMultiplier,
		"inventory":             itemsWithPrices,
		"restock_interval":      restockInterval, // Minutes between restocks
		"time_until_restock":    int(timeUntilRestock), // Minutes until next restock
		"just_restocked":        restocked, // Whether merchant just restocked
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ Loaded shop data for merchant: %s", merchantID)
}

// Buy items from shop
func handleBuyFromShop(w http.ResponseWriter, r *http.Request) {
	var transaction types.ShopTransaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Invalid transaction data",
		})
		return
	}

	log.Printf("🛒 Processing buy: %s buying %dx %s from %s", transaction.Npub, transaction.Quantity, transaction.ItemID, transaction.MerchantID)

	// Get session from memory (not disk!)
	session, err := sessionManager.GetSession(transaction.Npub, transaction.SaveID)
	if err != nil {
		log.Printf("❌ Session not found: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Session not found",
		})
		return
	}

	save := &session.SaveData

	// Get merchant data
	npcData, err := db.GetNPCByID(transaction.MerchantID)
	if err != nil {
		log.Printf("❌ Error loading merchant: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Merchant not found",
		})
		return
	}

	// Parse shop config
	configJSON, _ := json.Marshal(npcData.ShopConfig)
	var shopConfig types.ShopConfig
	json.Unmarshal(configJSON, &shopConfig)

	// Find item in shop inventory config
	var shopItem *types.ShopInventoryItem
	for i := range shopConfig.Inventory {
		if shopConfig.Inventory[i].ItemID == transaction.ItemID {
			shopItem = &shopConfig.Inventory[i]
			break
		}
	}

	if shopItem == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Item not in shop inventory",
		})
		return
	}

	// Get merchant state to check current stock
	initialInventory := make([]merchant.MerchantInventoryItem, 0)
	for _, invItem := range shopConfig.Inventory {
		initialInventory = append(initialInventory, merchant.MerchantInventoryItem{
			ItemID:       invItem.ItemID,
			CurrentStock: invItem.Stock,
			MaxStock:     invItem.MaxStock,
		})
	}

	restockInterval := 1440
	if shopConfig.RestockInterval > 0 {
		restockInterval = shopConfig.RestockInterval
	}

	merchantManager := merchant.GetManager()
	merchantState, _ := merchantManager.GetMerchantState(transaction.Npub, transaction.MerchantID, shopConfig.StartingGold, initialInventory, restockInterval)

	// Check current stock from merchant state
	currentStock := 0
	if stateItem, exists := merchantState.Inventory[transaction.ItemID]; exists {
		currentStock = stateItem.CurrentStock
	}

	if currentStock < transaction.Quantity {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Not enough stock (available: %d)", currentStock),
		})
		return
	}

	// Get item data for price
	item, err := db.GetItemByID(transaction.ItemID)
	if err != nil {
		log.Printf("❌ Error loading item: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Item not found",
		})
		return
	}

	// Calculate total cost
	buyPrice := int(float64(item.Value) * shopConfig.SellPriceMultiplier)
	totalCost := buyPrice * transaction.Quantity

	// Check player gold (using existing helper function)
	playerGold := getGoldQuantity(save)

	if playerGold < totalCost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Not enough gold (need %d, have %d)", totalCost, playerGold),
		})
		return
	}

	// Try to add items to inventory first to see how many fit
	itemsAdded, err := addItemToInventory(save, transaction.ItemID, transaction.Quantity)
	if err != nil && itemsAdded == 0 {
		// No items could be added
		log.Printf("❌ Error adding item to inventory: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   err.Error(),
			"message": "No room in inventory",
		})
		return
	}

	// Calculate actual cost for items that fit
	actualCost := buyPrice * itemsAdded

	// Deduct gold for items that were added
	if !deductGold(save, actualCost) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Failed to deduct gold",
		})
		return
	}

	// Update session in memory (not disk!)
	if err := sessionManager.UpdateSession(transaction.Npub, transaction.SaveID, session.SaveData); err != nil {
		log.Printf("❌ Failed to update session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Failed to update session",
		})
		return
	}

	// Update merchant state (deduct stock, add gold from player)
	merchantManager.UpdateMerchantInventory(transaction.Npub, transaction.MerchantID, transaction.ItemID, -itemsAdded, actualCost)

	// Build response message
	var message string
	if itemsAdded < transaction.Quantity {
		message = fmt.Sprintf("Bought %dx %s for %dg (inventory full - %d didn't fit)", itemsAdded, item.Name, actualCost, transaction.Quantity-itemsAdded)
		log.Printf("⚠️ Partial buy: %s bought %dx %s for %dg (%d didn't fit - inventory full)", transaction.Npub, itemsAdded, transaction.ItemID, actualCost, transaction.Quantity-itemsAdded)
	} else {
		message = fmt.Sprintf("Bought %dx %s for %dg", itemsAdded, item.Name, actualCost)
		log.Printf("✅ Buy successful: %s bought %dx %s for %dg (IN MEMORY)", transaction.Npub, itemsAdded, transaction.ItemID, actualCost)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":      true,
		"message":      message,
		"gold_spent":   actualCost,
		"new_gold":     playerGold - actualCost,
		"items_bought": itemsAdded,
	})
}

// Sell items to shop
func handleSellToShop(w http.ResponseWriter, r *http.Request) {
	var transaction types.ShopTransaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Invalid transaction data",
		})
		return
	}

	log.Printf("💰 Processing sell: %s selling %dx %s to %s", transaction.Npub, transaction.Quantity, transaction.ItemID, transaction.MerchantID)

	// Get session from memory (not disk!)
	session, err := sessionManager.GetSession(transaction.Npub, transaction.SaveID)
	if err != nil {
		log.Printf("❌ Session not found: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Session not found",
		})
		return
	}

	save := &session.SaveData

	// Get merchant data
	npcData, err := db.GetNPCByID(transaction.MerchantID)
	if err != nil {
		log.Printf("❌ Error loading merchant: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Merchant not found",
		})
		return
	}

	// Parse shop config
	configJSON, _ := json.Marshal(npcData.ShopConfig)
	var shopConfig types.ShopConfig
	json.Unmarshal(configJSON, &shopConfig)

	// Check if merchant buys items
	if !shopConfig.BuysItems {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "This merchant doesn't buy items",
		})
		return
	}

	// Get item data for price
	item, err := db.GetItemByID(transaction.ItemID)
	if err != nil {
		log.Printf("❌ Error loading item: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Item not found",
		})
		return
	}

	// Calculate total value
	sellPrice := int(float64(item.Value) * shopConfig.BuyPriceMultiplier)
	totalValue := sellPrice * transaction.Quantity

	// Get merchant state to check current gold
	initialInventory := make([]merchant.MerchantInventoryItem, 0)
	for _, invItem := range shopConfig.Inventory {
		initialInventory = append(initialInventory, merchant.MerchantInventoryItem{
			ItemID:       invItem.ItemID,
			CurrentStock: invItem.Stock,
			MaxStock:     invItem.MaxStock,
		})
	}

	restockInterval := 1440
	if shopConfig.RestockInterval > 0 {
		restockInterval = shopConfig.RestockInterval
	}

	merchantManager := merchant.GetManager()
	merchantState, _ := merchantManager.GetMerchantState(transaction.Npub, transaction.MerchantID, shopConfig.StartingGold, initialInventory, restockInterval)

	// Check merchant gold from state
	merchantGold := merchantState.CurrentGold
	if merchantGold < totalValue {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Merchant doesn't have enough gold (needs %d, has %d)", totalValue, merchantGold),
		})
		return
	}

	// Check player has items and remove them
	if err := removeItemFromInventory(save, transaction.ItemID, transaction.Quantity); err != nil {
		log.Printf("❌ Error removing item from inventory: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Add gold to player (using existing helper function)
	if err := addGoldToInventory(save.Inventory, totalValue); err != nil {
		log.Printf("❌ Error adding gold to inventory: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Failed to add gold",
		})
		return
	}

	playerGold := getGoldQuantity(save)

	// Update session in memory (not disk!)
	if err := sessionManager.UpdateSession(transaction.Npub, transaction.SaveID, session.SaveData); err != nil {
		log.Printf("❌ Failed to update session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Failed to update session",
		})
		return
	}

	// Update merchant state (add stock, deduct gold paid to player)
	merchantManager.UpdateMerchantInventory(transaction.Npub, transaction.MerchantID, transaction.ItemID, transaction.Quantity, -totalValue)

	log.Printf("✅ Sell successful: %s sold %dx %s for %dg (IN MEMORY)", transaction.Npub, transaction.Quantity, transaction.ItemID, totalValue)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":     true,
		"message":     fmt.Sprintf("Sold %dx %s for %dg", transaction.Quantity, item.Name, totalValue),
		"gold_earned": totalValue,
		"new_gold":    playerGold + totalValue,
		"items_sold":  transaction.Quantity,
	})
}

// Helper: Add item to player inventory with intelligent stacking and slot priority
// Returns: (itemsAdded int, error)
func addItemToInventory(save *SaveFile, itemID string, quantity int) (int, error) {
	log.Printf("🔧 addItemToInventory called: itemID=%s, quantity=%d", itemID, quantity)

	inventory := save.Inventory

	// Get item data to check max stack size
	itemData, err := db.GetItemByID(itemID)
	if err != nil {
		return 0, fmt.Errorf("item not found: %s", itemID)
	}

	// Parse stack size from properties JSON
	maxStack := 1 // Default to 1
	if itemData.Properties != "" {
		var properties map[string]any
		if err := json.Unmarshal([]byte(itemData.Properties), &properties); err == nil {
			if val, ok := properties["stack"].(float64); ok {
				maxStack = int(val)
			}
		}
	}

	log.Printf("📦 Adding %dx %s to inventory (max stack: %d)", quantity, itemID, maxStack)

	// Get inventory slots
	generalSlots, ok := inventory["general_slots"].([]any)
	if !ok {
		return 0, fmt.Errorf("invalid inventory structure")
	}

	gearSlots, ok := inventory["gear_slots"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("invalid gear slots")
	}

	bag, _ := gearSlots["bag"].(map[string]any)
	backpackSlots, _ := bag["contents"].([]any)

	remaining := quantity
	totalAdded := 0

	log.Printf("🔍 Inventory state: %d general slots, %d backpack slots", len(generalSlots), len(backpackSlots))

	// STEP 1: Try to stack with existing items in backpack first
	if backpackSlots != nil {
		log.Printf("🔍 Checking backpack for existing %s stacks...", itemID)
		for i, slotData := range backpackSlots {
			if remaining <= 0 {
				break
			}

			slot, ok := slotData.(map[string]any)
			if !ok {
				continue
			}

			if slot["item"] != itemID {
				log.Printf("  🔍 backpack[%d]: %v (not a match)", i, slot["item"])
				continue
			}

			log.Printf("  ✨ backpack[%d]: Found existing %s!", i, itemID)
			currentQty := getSlotQuantity(slot)
			if currentQty >= maxStack {
				continue // Already at max stack
			}

			canAdd := maxStack - currentQty
			if canAdd > remaining {
				canAdd = remaining
			}

			slot["quantity"] = currentQty + canAdd
			backpackSlots[i] = slot
			remaining -= canAdd
			totalAdded += canAdd
			log.Printf("  ✅ Stacked %d in backpack[%d] (now %d)", canAdd, i, currentQty+canAdd)
		}
	}

	// STEP 2: Try to stack with existing items in general slots
	for i, slotData := range generalSlots {
		if remaining <= 0 {
			break
		}

		slot, ok := slotData.(map[string]any)
		if !ok || slot["item"] != itemID {
			continue
		}

		currentQty := getSlotQuantity(slot)
		if currentQty >= maxStack {
			continue // Already at max stack
		}

		canAdd := maxStack - currentQty
		if canAdd > remaining {
			canAdd = remaining
		}

		slot["quantity"] = currentQty + canAdd
		generalSlots[i] = slot
		remaining -= canAdd
		totalAdded += canAdd
		log.Printf("  ✅ Stacked %d in general[%d] (now %d)", canAdd, i, currentQty+canAdd)
	}

	// STEP 3: Fill empty backpack slots
	for i, slotData := range backpackSlots {
		if remaining <= 0 {
			break
		}

		slot, ok := slotData.(map[string]any)
		if !ok {
			continue
		}

		if slot["item"] == nil || slot["item"] == "" {
			toAdd := remaining
			if toAdd > maxStack {
				toAdd = maxStack
			}

			slot["item"] = itemID
			slot["quantity"] = toAdd
			backpackSlots[i] = slot
			remaining -= toAdd
			totalAdded += toAdd
			log.Printf("  ✅ Added %d to empty backpack[%d]", toAdd, i)
		}
	}

	// STEP 4: Fill empty general slots
	for i, slotData := range generalSlots {
		if remaining <= 0 {
			break
		}

		slot, ok := slotData.(map[string]any)
		if !ok {
			continue
		}

		if slot["item"] == nil || slot["item"] == "" {
			toAdd := remaining
			if toAdd > maxStack {
				toAdd = maxStack
			}

			slot["item"] = itemID
			slot["quantity"] = toAdd
			generalSlots[i] = slot
			remaining -= toAdd
			totalAdded += toAdd
			log.Printf("  ✅ Added %d to empty general[%d]", toAdd, i)
		}
	}

	// Update inventory
	if backpackSlots != nil {
		bag["contents"] = backpackSlots
		gearSlots["bag"] = bag
	}
	inventory["general_slots"] = generalSlots
	inventory["gear_slots"] = gearSlots

	if remaining > 0 {
		log.Printf("⚠️ Inventory full - added %d/%d items (%d couldn't fit)", totalAdded, quantity, remaining)
		if totalAdded == 0 {
			return 0, fmt.Errorf("no room in inventory")
		}
		return totalAdded, nil // Partial success
	}

	log.Printf("✅ Successfully added all %d items", totalAdded)
	return totalAdded, nil
}

// Helper: Get quantity from slot (handles both int and float64)
func getSlotQuantity(slot map[string]any) int {
	switch v := slot["quantity"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

// Helper: Remove item from player inventory
func removeItemFromInventory(save *SaveFile, itemID string, quantity int) error {
	inventory := save.Inventory

	// Get general_slots
	generalSlots, ok := inventory["general_slots"].([]any)
	if !ok {
		return fmt.Errorf("invalid inventory structure")
	}

	// Find item and check quantity
	for i, slotData := range generalSlots {
		slot, ok := slotData.(map[string]any)
		if !ok {
			continue
		}

		if slot["item"] == itemID {
			currentQty := 0
			if qty, ok := slot["quantity"].(float64); ok {
				currentQty = int(qty)
			}

			if currentQty < quantity {
				return fmt.Errorf("not enough items (have %d, need %d)", currentQty, quantity)
			}

			newQty := currentQty - quantity
			if newQty <= 0 {
				slot["item"] = nil
				slot["quantity"] = 0
			} else {
				slot["quantity"] = newQty
			}

			generalSlots[i] = slot
			inventory["general_slots"] = generalSlots
			return nil
		}
	}

	return fmt.Errorf("item not found in inventory")
}
