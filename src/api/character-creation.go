package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"nostr-hero/src/db"
	"nostr-hero/src/functions"
	"nostr-hero/src/types"
	"nostr-hero/src/utils"
)

// ============================================================================
// REQUEST/RESPONSE TYPES
// ============================================================================

// CreateCharacterRequest represents the frontend's simple equipment choices
type CreateCharacterRequest struct {
	Npub              string            `json:"npub"`
	Name              string            `json:"name"`
	EquipmentChoices  map[string]string `json:"equipment_choices"` // e.g., {"choice-0": "scimitar", "choice-1": "shield"}
	PackChoice        string            `json:"pack_choice"`       // e.g., "druid-pack"
}

// CreateCharacterResponse returns the save ID and full character data
type CreateCharacterResponse struct {
	Success   bool        `json:"success"`
	SaveID    string      `json:"save_id"`
	Character interface{} `json:"character,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// EquipmentChoice represents a single choice from starting-gear.json
type EquipmentChoice struct {
	Description string            `json:"description"`
	Options     []EquipmentOption `json:"options"`
}

// EquipmentOption represents one option in a choice
type EquipmentOption struct {
	Type     string          `json:"type"` // "single", "bundle", "multi_slot"
	Item     string          `json:"item,omitempty"`
	Quantity int             `json:"quantity,omitempty"`
	Items    []ItemWithQty   `json:"items,omitempty"`  // For bundles
	Slots    []MultiSlotItem `json:"slots,omitempty"`  // For multi_slot
}

// ItemWithQty represents an item with quantity
type ItemWithQty struct {
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

// MultiSlotItem represents a slot in a multi_slot choice
type MultiSlotItem struct {
	Type     string   `json:"type"` // "weapon_choice" or "fixed"
	Options  []string `json:"options,omitempty"`
	Item     string   `json:"item,omitempty"`
	Quantity int      `json:"quantity,omitempty"`
}

// StartingGearData represents the class-specific gear from JSON
type StartingGearData struct {
	Class        string `json:"class"`
	StartingGear struct {
		EquipmentChoices []EquipmentChoice `json:"equipment_choices"`
		PackChoice       *struct {
			Description string   `json:"description"`
			Options     []string `json:"options"`
		} `json:"pack_choice"`
		GivenItems []ItemWithQty `json:"given_items"`
	} `json:"starting_gear"`
}

// ============================================================================
// API HANDLER
// ============================================================================

// CreateCharacterHandler creates a new character and save file
func CreateCharacterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Error decoding request: %v", err)
		respondWithError(w, "Invalid request data")
		return
	}

	log.Printf("🎮 Creating character for npub: %s, name: %s", req.Npub, req.Name)

	// 1. Decode npub and generate character
	pubKey, err := utils.DecodeNpub(req.Npub)
	if err != nil {
		respondWithError(w, "Invalid npub")
		return
	}

	weightData, err := getWeightsFromDB()
	if err != nil {
		respondWithError(w, "Failed to load weight data: "+err.Error())
		return
	}

	weightDataJSON, _ := json.Marshal(weightData)
	var weightDataStruct types.WeightData
	json.Unmarshal(weightDataJSON, &weightDataStruct)

	character := functions.GenerateCharacter(pubKey, &weightDataStruct)

	// 2. Load starting gear data
	startingGear, err := loadStartingGearForClass(character.Class)
	if err != nil {
		respondWithError(w, "Failed to load starting gear: "+err.Error())
		return
	}

	// 3. Build inventory from equipment choices
	database := db.GetDB()
	if database == nil {
		respondWithError(w, "Database not available")
		return
	}

	inventory, err := buildInventoryFromChoices(database, startingGear, req.EquipmentChoices, req.PackChoice)
	if err != nil {
		respondWithError(w, "Failed to build inventory: "+err.Error())
		return
	}

	// 4. Generate spell slots
	spellSlots, err := generateSpellSlots(character.Class)
	if err != nil {
		log.Printf("⚠️  Failed to generate spell slots: %v", err)
		spellSlots = make(map[string]interface{})
	}

	// 5. Load known spells
	knownSpells, err := loadKnownSpells(character.Class)
	if err != nil {
		log.Printf("⚠️  Failed to load known spells: %v", err)
		knownSpells = []string{}
	}

	// 6. Determine starting location based on race
	startingCity, err := getStartingCityForRace(character.Race)
	if err != nil {
		log.Printf("⚠️  Failed to get starting city: %v", err)
		startingCity = "millhaven"
	}

	// 7. Generate starting vault
	startingVault := generateStartingVault(startingCity)
	vaults := []map[string]interface{}{startingVault}

	// 8. Calculate HP and Mana
	hp := calculateHP(character.Stats["Constitution"], character.Class)
	mana := calculateMana(character.Stats, character.Class)

	// 9. Get starting gold
	gold, err := getStartingGold(character.Background)
	if err != nil {
		log.Printf("⚠️  Failed to get starting gold: %v", err)
		gold = 1000 // Default
	}

	// 10. Use location IDs directly (not display names)
	// startingCity is already an ID like "millhaven", "verdant", etc.
	locationID := startingCity
	districtKey := "center" // All characters start in the center district
	buildingID := ""        // Start outdoors

	// 11. Get music tracks (auto-unlock + location track)
	musicTracks := getAutoUnlockMusicTracks()
	locationMusic := getMusicTrackForLocation(startingCity)
	if locationMusic != "" {
		musicTracks = append(musicTracks, locationMusic)
	}

	// 12. Convert stats to interface{} map
	statsInterface := make(map[string]interface{})
	for k, v := range character.Stats {
		statsInterface[k] = v
	}

	// 13. Create save file
	saveFile := SaveFile{
		D:                   req.Name,
		CreatedAt:           time.Now().UTC().Format(time.RFC3339),
		Race:                character.Race,
		Class:               character.Class,
		Background:          character.Background,
		Alignment:           character.Alignment,
		Experience:          0,
		HP:                  hp,
		MaxHP:               hp,
		Mana:                mana,
		MaxMana:             mana,
		Fatigue:             0,
		FatigueCounter:      0,  // ← ADDED
		Hunger:              1,  // ← ADDED (1 = Hungry)
		HungerCounter:       0,  // ← ADDED
		Gold:                gold,
		Stats:               statsInterface,
		Location:            locationID,    // Use ID, not display name
		District:            districtKey,   // Use key, not display name
		Building:            buildingID,    // Use ID, not display name
		CurrentDay:          1,
		TimeOfDay:           6, // Highnoon
		Inventory:           inventory,
		Vaults:              vaults,
		KnownSpells:         knownSpells,
		SpellSlots:          spellSlots,
		LocationsDiscovered: []string{startingCity}, // Only the city ID, not districts
		MusicTracksUnlocked: musicTracks,
		InternalNpub:        req.Npub,
		InternalID:          fmt.Sprintf("save_%d", time.Now().Unix()),
	}

	// 12. Save to disk
	if err := saveToDisk(req.Npub, &saveFile); err != nil {
		respondWithError(w, "Failed to save character: "+err.Error())
		return
	}

	log.Printf("✅ Character created successfully: %s", saveFile.InternalID)

	// 13. Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateCharacterResponse{
		Success: true,
		SaveID:  saveFile.InternalID,
		Character: map[string]interface{}{
			"name":       saveFile.D,
			"race":       saveFile.Race,
			"class":      saveFile.Class,
			"background": saveFile.Background,
			"alignment":  saveFile.Alignment,
			"hp":         saveFile.HP,
			"max_hp":     saveFile.MaxHP,
			"mana":       saveFile.Mana,
			"max_mana":   saveFile.MaxMana,
			"gold":       saveFile.Gold,
			"stats":      saveFile.Stats,
		},
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func respondWithError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(CreateCharacterResponse{
		Success: false,
		Error:   message,
	})
}

// Calculate HP based on class and constitution
func calculateHP(constitution int, class string) int {
	hitDice := map[string]int{
		"Barbarian": 12,
		"Fighter":   10,
		"Paladin":   10,
		"Monk":      8,
		"Ranger":    10,
		"Rogue":     8,
		"Bard":      8,
		"Cleric":    8,
		"Druid":     8,
		"Sorcerer":  6,
		"Warlock":   8,
		"Wizard":    6,
	}

	hitDie := hitDice[class]
	if hitDie == 0 {
		hitDie = 8 // Default
	}

	conModifier := (constitution - 10) / 2
	return hitDie + conModifier
}

// Calculate mana based on class and stats
func calculateMana(stats map[string]int, class string) int {
	spellcasters := map[string]string{
		"Wizard":   "Intelligence",
		"Sorcerer": "Charisma",
		"Warlock":  "Charisma",
		"Bard":     "Charisma",
		"Cleric":   "Wisdom",
		"Druid":    "Wisdom",
		"Paladin":  "Charisma",
		"Ranger":   "Wisdom",
	}

	spellcastingStat, isCaster := spellcasters[class]
	if !isCaster {
		return 0
	}

	statValue := stats[spellcastingStat]
	statModifier := (statValue - 10) / 2

	mana := statModifier + 1
	if mana < 0 {
		mana = 0
	}

	return mana
}
