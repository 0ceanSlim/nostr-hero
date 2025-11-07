package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// STARTING GEAR LOADING
// ============================================================================

func loadStartingGearForClass(class string) (*StartingGearData, error) {
	data, err := os.ReadFile("docs/data/character/starting-gear.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read starting-gear.json: %v", err)
	}

	var allGear []StartingGearData
	if err := json.Unmarshal(data, &allGear); err != nil {
		return nil, fmt.Errorf("failed to parse starting-gear.json: %v", err)
	}

	for _, gear := range allGear {
		if gear.Class == class {
			return &gear, nil
		}
	}

	return nil, fmt.Errorf("no starting gear found for class: %s", class)
}

// ============================================================================
// INVENTORY BUILDING
// ============================================================================

func buildInventoryFromChoices(database *sql.DB, startingGear *StartingGearData, choices map[string]string, packChoice string) (map[string]interface{}, error) {
	// 1. Collect all items from given_items + player choices
	allItems := []ItemWithQty{}

	// Add given items
	allItems = append(allItems, startingGear.StartingGear.GivenItems...)

	// Log what we received
	log.Printf("📦 Received equipment choices: %+v", choices)
	log.Printf("📦 Pack choice: %s", packChoice)

	// Add selected equipment
	for i, equipChoice := range startingGear.StartingGear.EquipmentChoices {
		choiceKey := fmt.Sprintf("choice-%d", i)
		selectedID := choices[choiceKey]

		if selectedID == "" {
			log.Printf("⚠️  No selection for %s (available keys: %v)", choiceKey, getKeys(choices))
			continue
		}

		// Check if it's a JSON array (complex weapon choice)
		if selectedID[0] == '[' {
			var weaponList [][]interface{}
			if err := json.Unmarshal([]byte(selectedID), &weaponList); err == nil {
				// Successfully parsed as complex choice
				for _, weapon := range weaponList {
					if len(weapon) >= 2 {
						itemID, ok1 := weapon[0].(string)
						qty, ok2 := weapon[1].(float64)
						if ok1 && ok2 {
							allItems = append(allItems, ItemWithQty{Item: itemID, Quantity: int(qty)})
						}
					}
				}
				continue
			}
		}

		// Find the selected option
		for _, option := range equipChoice.Options {
			if option.Type == "single" && option.Item == selectedID {
				allItems = append(allItems, ItemWithQty{Item: option.Item, Quantity: option.Quantity})
				break
			} else if option.Type == "bundle" {
				// Check if this bundle's first item matches (simplified check)
				if len(option.Items) > 0 && option.Items[0].Item == selectedID {
					allItems = append(allItems, option.Items...)
					break
				}
			}
		}
	}

	// Add pack choice
	if packChoice != "" {
		allItems = append(allItems, ItemWithQty{Item: packChoice, Quantity: 1})
	}

	// 2. Stack items (respect stack limits)
	stackedItems, err := stackItems(database, allItems)
	if err != nil {
		return nil, fmt.Errorf("failed to stack items: %v", err)
	}

	// 3. Build inventory structure with auto-equipping
	inventory, err := createInventoryStructure(database, stackedItems)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory: %v", err)
	}

	return inventory, nil
}

// ============================================================================
// ITEM STACKING
// ============================================================================

func stackItems(database *sql.DB, items []ItemWithQty) ([]ItemWithQty, error) {
	stacked := []ItemWithQty{}

	for _, item := range items {
		// Get item data to check stack limit
		stackLimit, err := getItemStackLimit(database, item.Item)
		if err != nil {
			log.Printf("⚠️  Failed to get stack limit for %s: %v", item.Item, err)
			stackLimit = 1 // Default
		}

		remainingQty := item.Quantity

		// Try to add to existing stacks
		for i := range stacked {
			if stacked[i].Item == item.Item && stacked[i].Quantity < stackLimit {
				canAdd := min(remainingQty, stackLimit-stacked[i].Quantity)
				stacked[i].Quantity += canAdd
				remainingQty -= canAdd

				if remainingQty <= 0 {
					break
				}
			}
		}

		// Create new stacks for remaining quantity
		for remainingQty > 0 {
			stackSize := min(remainingQty, stackLimit)
			stacked = append(stacked, ItemWithQty{Item: item.Item, Quantity: stackSize})
			remainingQty -= stackSize
		}
	}

	return stacked, nil
}

func getItemStackLimit(database *sql.DB, itemID string) (int, error) {
	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM items WHERE id = ?", itemID).Scan(&propertiesJSON)
	if err != nil {
		return 1, err
	}

	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
		return 1, err
	}

	if stack, ok := properties["stack"].(float64); ok {
		return int(stack), nil
	}

	return 1, nil
}

// ============================================================================
// INVENTORY STRUCTURE CREATION
// ============================================================================

func createInventoryStructure(database *sql.DB, items []ItemWithQty) (map[string]interface{}, error) {
	inventory := map[string]interface{}{
		"general_slots": []map[string]interface{}{
			{"slot": 0, "item": nil, "quantity": 0},
			{"slot": 1, "item": nil, "quantity": 0},
			{"slot": 2, "item": nil, "quantity": 0},
			{"slot": 3, "item": nil, "quantity": 0},
		},
		"gear_slots": map[string]interface{}{
			"bag":        map[string]interface{}{"item": nil, "quantity": 0},
			"left_arm":   map[string]interface{}{"item": nil, "quantity": 0},
			"right_arm":  map[string]interface{}{"item": nil, "quantity": 0},
			"armor":      map[string]interface{}{"item": nil, "quantity": 0},
			"necklace":   map[string]interface{}{"item": nil, "quantity": 0},
			"ring":       map[string]interface{}{"item": nil, "quantity": 0},
			"ammunition": map[string]interface{}{"item": nil, "quantity": 0},
			"clothes":    map[string]interface{}{"item": nil, "quantity": 0},
		},
	}

	gearSlots := inventory["gear_slots"].(map[string]interface{})
	generalSlots := inventory["general_slots"].([]map[string]interface{})

	remainingItems := []ItemWithQty{}
	twoHandedEquipped := false

	// 1. Handle packs first (auto-unpack to bag slot)
	for _, item := range items {
		if isPackItem(item.Item) {
			contents, err := unpackItem(database, item.Item)
			if err != nil {
				log.Printf("⚠️  Failed to unpack %s: %v", item.Item, err)
				continue
			}

			gearSlots["bag"] = map[string]interface{}{
				"item":     "backpack",
				"quantity": 1,
				"contents": contents,
			}
		} else {
			remainingItems = append(remainingItems, item)
		}
	}

	// 2. Auto-equip equipment items
	for i := len(remainingItems) - 1; i >= 0; i-- {
		item := remainingItems[i]
		itemData, err := getItemData(database, item.Item)
		if err != nil {
			log.Printf("⚠️  Failed to get item data for %s: %v", item.Item, err)
			continue
		}

		gearSlot := getGearSlot(itemData)
		if gearSlot == "" {
			continue // Not equipment
		}

		equipped := false

		switch gearSlot {
		case "armor", "necklace", "ring", "ammunition", "clothes":
			if gearSlots[gearSlot].(map[string]interface{})["item"] == nil {
				gearSlots[gearSlot] = map[string]interface{}{
					"item":     item.Item,
					"quantity": item.Quantity,
				}
				equipped = true
			}

		case "hands":
			isTwoHanded := hasTags(itemData, []string{"two-handed"})

			if isTwoHanded {
				if gearSlots["right_arm"].(map[string]interface{})["item"] == nil {
					gearSlots["right_arm"] = map[string]interface{}{
						"item":     item.Item,
						"quantity": item.Quantity,
					}
					twoHandedEquipped = true
					equipped = true
				}
			} else {
				if gearSlots["right_arm"].(map[string]interface{})["item"] == nil {
					gearSlots["right_arm"] = map[string]interface{}{
						"item":     item.Item,
						"quantity": item.Quantity,
					}
					equipped = true
				} else if gearSlots["left_arm"].(map[string]interface{})["item"] == nil && !twoHandedEquipped {
					gearSlots["left_arm"] = map[string]interface{}{
						"item":     item.Item,
						"quantity": item.Quantity,
					}
					equipped = true
				}
			}
		}

		if equipped {
			// Remove from remaining items
			remainingItems = append(remainingItems[:i], remainingItems[i+1:]...)
		}
	}

	// 3. Put remaining items in general slots or backpack
	generalSlotIndex := 0
	for _, item := range remainingItems {
		itemData, _ := getItemData(database, item.Item)
		isContainer := hasTags(itemData, []string{"container"})

		// Containers go in general slots, not in bag
		if isContainer {
			if generalSlotIndex < 4 {
				generalSlots[generalSlotIndex] = map[string]interface{}{
					"slot":     generalSlotIndex,
					"item":     item.Item,
					"quantity": item.Quantity,
				}
				generalSlotIndex++
			}
		} else {
			// Try to add to backpack first
			if bag, ok := gearSlots["bag"].(map[string]interface{}); ok && bag["item"] != nil {
				if contents, ok := bag["contents"].([]map[string]interface{}); ok {
					placed := false
					for i, slot := range contents {
						if slot["item"] == nil {
							contents[i] = map[string]interface{}{
								"slot":     slot["slot"],
								"item":     item.Item,
								"quantity": item.Quantity,
							}
							placed = true
							break
						}
					}
					if placed {
						continue
					}
				}
			}

			// Backpack full or doesn't exist, use general slots
			if generalSlotIndex < 4 {
				generalSlots[generalSlotIndex] = map[string]interface{}{
					"slot":     generalSlotIndex,
					"item":     item.Item,
					"quantity": item.Quantity,
				}
				generalSlotIndex++
			}
		}
	}

	inventory["general_slots"] = generalSlots
	inventory["gear_slots"] = gearSlots

	return inventory, nil
}

// ============================================================================
// PACK UNPACKING
// ============================================================================

func isPackItem(itemID string) bool {
	return itemID == "explorers-pack" || itemID == "priests-pack" ||
	       itemID == "scholars-pack" || itemID == "dungeoneers-pack" ||
	       itemID == "diplomats-pack" || itemID == "entertainers-pack" ||
	       itemID == "burglars-pack" || itemID == "druid-pack"
}

func unpackItem(database *sql.DB, packID string) ([]map[string]interface{}, error) {
	var propertiesJSON string
	err := database.QueryRow("SELECT properties FROM items WHERE id = ?", packID).Scan(&propertiesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to query pack: %v", err)
	}

	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
		return nil, fmt.Errorf("failed to parse properties: %v", err)
	}

	contentsRaw, ok := properties["contents"]
	if !ok {
		return nil, fmt.Errorf("pack has no contents field")
	}

	// Convert contents to [][]interface{}
	contentsArray, ok := contentsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("contents is not an array")
	}

	slots := []map[string]interface{}{}
	slotIndex := 0

	for _, contentItem := range contentsArray {
		itemArray, ok := contentItem.([]interface{})
		if !ok || len(itemArray) < 2 {
			continue
		}

		itemID, ok1 := itemArray[0].(string)
		qty, ok2 := itemArray[1].(float64)
		if !ok1 || !ok2 {
			continue
		}

		// Skip backpack and pack items
		if itemID == "backpack" || isPackItem(itemID) {
			continue
		}

		slots = append(slots, map[string]interface{}{
			"slot":     slotIndex,
			"item":     itemID,
			"quantity": int(qty),
		})
		slotIndex++
	}

	// Fill remaining slots to 20
	for slotIndex < 20 {
		slots = append(slots, map[string]interface{}{
			"slot":     slotIndex,
			"item":     nil,
			"quantity": 0,
		})
		slotIndex++
	}

	return slots, nil
}

// ============================================================================
// ITEM HELPERS
// ============================================================================

func getItemData(database *sql.DB, itemID string) (map[string]interface{}, error) {
	var propertiesJSON, tagsJSON string
	var name, description, itemType, rarity string

	err := database.QueryRow(`
		SELECT name, description, item_type, properties, tags, rarity
		FROM items WHERE id = ?
	`, itemID).Scan(&name, &description, &itemType, &propertiesJSON, &tagsJSON, &rarity)

	if err != nil {
		return nil, err
	}

	var properties map[string]interface{}
	var tags []interface{}

	json.Unmarshal([]byte(propertiesJSON), &properties)
	json.Unmarshal([]byte(tagsJSON), &tags)

	return map[string]interface{}{
		"id":         itemID,
		"name":       name,
		"description": description,
		"item_type":  itemType,
		"properties": properties,
		"tags":       tags,
		"rarity":     rarity,
	}, nil
}

func getGearSlot(itemData map[string]interface{}) string {
	if properties, ok := itemData["properties"].(map[string]interface{}); ok {
		if gearSlot, ok := properties["gear_slot"].(string); ok {
			return gearSlot
		}
	}
	return ""
}

func hasTags(itemData map[string]interface{}, searchTags []string) bool {
	tags, ok := itemData["tags"].([]interface{})
	if !ok {
		return false
	}

	for _, tag := range tags {
		tagStr, ok := tag.(string)
		if !ok {
			continue
		}
		for _, search := range searchTags {
			if tagStr == search {
				return true
			}
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getKeys(m map[string]string) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ============================================================================
// SPELL SLOTS AND KNOWN SPELLS
// ============================================================================

func generateSpellSlots(class string) (map[string]interface{}, error) {
	data, err := os.ReadFile("docs/data/character/spell-slots.json")
	if err != nil {
		return nil, err
	}

	// Parse as generic map first to skip the "description" field
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	classKey := strings.ToLower(class)
	classDataRaw, ok := rawData[classKey]
	if !ok {
		return make(map[string]interface{}), nil // Not a caster
	}

	classData, ok := classDataRaw.(map[string]interface{})
	if !ok {
		return make(map[string]interface{}), nil
	}

	level1SlotsRaw, ok := classData["1"]
	if !ok {
		return make(map[string]interface{}), nil
	}

	level1Slots, ok := level1SlotsRaw.(map[string]interface{})
	if !ok {
		return make(map[string]interface{}), nil
	}

	spellSlots := make(map[string]interface{})

	for slotType, countRaw := range level1Slots {
		count, ok := countRaw.(float64) // JSON numbers are float64
		if !ok || count <= 0 {
			continue
		}

		slots := []map[string]interface{}{}
		for i := 0; i < int(count); i++ {
			slots = append(slots, map[string]interface{}{
				"slot":     i,
				"spell":    nil,
				"quantity": 0,
			})
		}
		spellSlots[slotType] = slots
	}

	return spellSlots, nil
}

func loadKnownSpells(class string) ([]string, error) {
	data, err := os.ReadFile("docs/data/character/starting-spells.json")
	if err != nil {
		return nil, err
	}

	var allSpells map[string]map[string][]string
	if err := json.Unmarshal(data, &allSpells); err != nil {
		return nil, err
	}

	classKey := strings.ToLower(class)
	classSpells, ok := allSpells[classKey]
	if !ok {
		return []string{}, nil // Not a caster
	}

	knownSpells := []string{}

	if cantrips, ok := classSpells["cantrips"]; ok {
		knownSpells = append(knownSpells, cantrips...)
	}

	if level1, ok := classSpells["level1"]; ok {
		knownSpells = append(knownSpells, level1...)
	}

	return knownSpells, nil
}

// ============================================================================
// LOCATION HELPERS
// ============================================================================

func getStartingCityForRace(race string) (string, error) {
	data, err := os.ReadFile("docs/data/character/racial-starting-cities.json")
	if err != nil {
		return "millhaven", err
	}

	var racialCities struct {
		RacialStartingCities map[string]string `json:"racial_starting_cities"`
	}

	if err := json.Unmarshal(data, &racialCities); err != nil {
		return "millhaven", err
	}

	if city, ok := racialCities.RacialStartingCities[race]; ok {
		return city, nil
	}

	return "millhaven", nil
}

func getDisplayNamesForSave(database *sql.DB, locationID, districtKey, buildingID string) (string, string, string) {
	var name, propertiesJSON string
	err := database.QueryRow("SELECT name, properties FROM locations WHERE id = ?", locationID).Scan(&name, &propertiesJSON)

	if err != nil {
		log.Printf("⚠️  Failed to get location name: %v", err)
		return locationID, districtKey, buildingID
	}

	locationName := name

	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
		return locationName, districtKey, buildingID
	}

	districtName := districtKey
	buildingName := buildingID

	if districts, ok := properties["districts"].(map[string]interface{}); ok {
		if district, ok := districts[districtKey].(map[string]interface{}); ok {
			if dName, ok := district["name"].(string); ok {
				districtName = dName
			}

			if buildingID != "" {
				if buildings, ok := district["buildings"].([]interface{}); ok {
					for _, bldg := range buildings {
						if building, ok := bldg.(map[string]interface{}); ok {
							if bid, ok := building["id"].(string); ok && bid == buildingID {
								if bname, ok := building["name"].(string); ok {
									buildingName = bname
								}
							}
						}
					}
				}
			}
		}
	}

	return locationName, districtName, buildingName
}

func getMusicTrackForLocation(locationID string) string {
	data, err := os.ReadFile("docs/data/content/music.json")
	if err != nil {
		log.Printf("⚠️  Failed to read music.json: %v", err)
		return ""
	}

	var musicData struct {
		Tracks []struct {
			Title      string  `json:"title"`
			File       string  `json:"file"`
			UnlocksAt  *string `json:"unlocks_at"`
			AutoUnlock bool    `json:"auto_unlock"`
		} `json:"tracks"`
	}

	if err := json.Unmarshal(data, &musicData); err != nil {
		log.Printf("⚠️  Failed to parse music.json: %v", err)
		return ""
	}

	// Find track for this location
	for _, track := range musicData.Tracks {
		if track.UnlocksAt != nil && *track.UnlocksAt == locationID {
			return track.Title
		}
	}

	return ""
}

func getAutoUnlockMusicTracks() []string {
	data, err := os.ReadFile("docs/data/content/music.json")
	if err != nil {
		log.Printf("⚠️  Failed to read music.json: %v", err)
		return []string{}
	}

	var musicData struct {
		Tracks []struct {
			Title      string  `json:"title"`
			File       string  `json:"file"`
			UnlocksAt  *string `json:"unlocks_at"`
			AutoUnlock bool    `json:"auto_unlock"`
		} `json:"tracks"`
	}

	if err := json.Unmarshal(data, &musicData); err != nil {
		log.Printf("⚠️  Failed to parse music.json: %v", err)
		return []string{}
	}

	tracks := []string{}
	for _, track := range musicData.Tracks {
		if track.AutoUnlock {
			tracks = append(tracks, track.Title)
		}
	}

	return tracks
}

// ============================================================================
// VAULT GENERATION
// ============================================================================

func generateStartingVault(locationID string) map[string]interface{} {
	slots := []map[string]interface{}{}
	for i := 0; i < 40; i++ {
		slots = append(slots, map[string]interface{}{
			"slot":     i,
			"item":     nil,
			"quantity": 0,
		})
	}

	return map[string]interface{}{
		"location": locationID,
		"slots":    slots,
	}
}

// ============================================================================
// STARTING GOLD
// ============================================================================

func getStartingGold(background string) (int, error) {
	data, err := os.ReadFile("docs/data/character/starting-gold.json")
	if err != nil {
		return 1000, err
	}

	var goldData map[string][][]interface{}
	if err := json.Unmarshal(data, &goldData); err != nil {
		return 1000, err
	}

	goldList, ok := goldData["starting-gold"]
	if !ok {
		return 1000, fmt.Errorf("no starting-gold key")
	}

	for _, entry := range goldList {
		if len(entry) >= 2 {
			if bg, ok := entry[0].(string); ok && bg == background {
				if gold, ok := entry[1].(float64); ok {
					return int(gold), nil
				}
			}
		}
	}

	return 1000, nil
}

// ============================================================================
// SAVE TO DISK
// ============================================================================

func saveToDisk(npub string, saveFile *SaveFile) error {
	savesDir := filepath.Join(SavesDirectory, npub)
	if err := os.MkdirAll(savesDir, 0755); err != nil {
		return fmt.Errorf("failed to create saves directory: %v", err)
	}

	savePath := filepath.Join(savesDir, saveFile.InternalID+".json")
	return writeSaveFile(savePath, saveFile)
}
