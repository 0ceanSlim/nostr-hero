package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ActiveEffect stores only runtime state for an active effect (template data is in database)
type ActiveEffect struct {
	EffectID          string  `json:"effect_id"`          // ID of effect template (e.g., "performance-high")
	EffectIndex       int     `json:"effect_index"`       // Index in effect's effects array (0 for single-effect templates)
	DurationRemaining float64 `json:"duration_remaining"` // Minutes remaining until effect expires
	TotalDuration     float64 `json:"total_duration"`     // Original duration when effect was applied (for progress calculation)
	DelayRemaining    float64 `json:"delay_remaining"`    // Minutes remaining before effect starts
	TickAccumulator   float64 `json:"tick_accumulator"`   // Time accumulated since last tick (for periodic effects)
	AppliedAt         int     `json:"applied_at"`         // Time of day (minutes) when effect was applied
}

type SaveFile struct {
	D                   string                 `json:"d"`
	CreatedAt           string                 `json:"created_at"`
	Race                string                 `json:"race"`
	Class               string                 `json:"class"`
	Background          string                 `json:"background"`
	Alignment           string                 `json:"alignment"`
	Experience          int                    `json:"experience"`
	HP                  int                    `json:"hp"`
	MaxHP               int                    `json:"max_hp"`
	Mana                int                    `json:"mana"`
	MaxMana             int                    `json:"max_mana"`
	Fatigue             int                    `json:"fatigue"`             // Fatigue level (0-10+), penalties applied via effects
	Hunger              int                    `json:"hunger"`              // Hunger level (0-3: 0=Famished, 1=Hungry, 2=Satisfied, 3=Full), penalties applied via effects
	Stats               map[string]interface{} `json:"stats"`
	Location            string                 `json:"location"`     // City ID (e.g., "kingdom", "village-west")
	District            string                 `json:"district"`     // District key (e.g., "center", "north", "south")
	Building            string                 `json:"building"`     // Building ID or empty for outdoors
	CurrentDay          int                    `json:"current_day"`
	TimeOfDay           int                    `json:"time_of_day"` // Minutes in current day (0-1439, where 720=noon, 0=midnight)
	Inventory           map[string]interface{} `json:"inventory"`
	Vaults              []map[string]interface{} `json:"vaults"`
	KnownSpells         []string               `json:"known_spells"`
	SpellSlots          map[string]interface{} `json:"spell_slots"`
	LocationsDiscovered []string                   `json:"locations_discovered"`
	MusicTracksUnlocked []string                   `json:"music_tracks_unlocked"`
	ActiveEffects       []ActiveEffect             `json:"active_effects,omitempty"` // Compact format: only runtime state, template data in DB
	InternalID          string                     `json:"-"`                        // Not serialized, used internally for file naming
	InternalNpub        string                     `json:"-"`                        // Not serialized, used internally for directory structure
}

const SavesDirectory = "data/saves"

// Initialize saves directory
func init() {
	if err := os.MkdirAll(SavesDirectory, 0755); err != nil {
		log.Printf("Warning: Failed to create saves directory: %v", err)
	}
}

// SavesHandler handles save file operations
func SavesHandler(w http.ResponseWriter, r *http.Request) {
	// Extract npub from URL path: /api/saves/{npub}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/saves/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Missing npub in URL", http.StatusBadRequest)
		return
	}

	npub := pathParts[0]

	switch r.Method {
	case "GET":
		handleGetSaves(w, r, npub)
	case "POST":
		handleCreateSave(w, r, npub)
	case "DELETE":
		if len(pathParts) < 2 {
			http.Error(w, "Missing save ID", http.StatusBadRequest)
			return
		}
		handleDeleteSave(w, r, npub, pathParts[1])
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get all saves for a user
func handleGetSaves(w http.ResponseWriter, _ *http.Request, npub string) {
	log.Printf("📂 Loading saves for npub: %s", npub)

	savesDir := filepath.Join(SavesDirectory, npub)
	if _, err := os.Stat(savesDir); os.IsNotExist(err) {
		// No saves directory exists for this user
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SaveFile{})
		return
	}

	files, err := ioutil.ReadDir(savesDir)
	if err != nil {
		log.Printf("❌ Error reading saves directory: %v", err)
		http.Error(w, "Failed to read saves", http.StatusInternalServerError)
		return
	}

	var saves []SaveFile
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			savePath := filepath.Join(savesDir, file.Name())
			if saveData, err := loadSaveFile(savePath); err == nil {
				saves = append(saves, *saveData)
			} else {
				log.Printf("⚠️ Failed to load save file %s: %v", file.Name(), err)
			}
		}
	}

	log.Printf("✅ Found %d saves for npub: %s", len(saves), npub)
	w.Header().Set("Content-Type", "application/json")

	// Convert saves to include id field in JSON output
	savesWithID := make([]map[string]interface{}, 0, len(saves))
	for _, save := range saves {
		saveMap := make(map[string]interface{})
		saveMap["id"] = save.InternalID
		saveMap["d"] = save.D
		saveMap["created_at"] = save.CreatedAt
		saveMap["race"] = save.Race
		saveMap["class"] = save.Class
		saveMap["background"] = save.Background
		saveMap["alignment"] = save.Alignment
		saveMap["experience"] = save.Experience
		saveMap["hp"] = save.HP
		saveMap["max_hp"] = save.MaxHP
		saveMap["mana"] = save.Mana
		saveMap["max_mana"] = save.MaxMana
		saveMap["fatigue"] = save.Fatigue
		saveMap["hunger"] = save.Hunger
		saveMap["stats"] = save.Stats
		saveMap["location"] = save.Location
		saveMap["district"] = save.District
		saveMap["building"] = save.Building
		saveMap["inventory"] = save.Inventory
		saveMap["vaults"] = save.Vaults
		saveMap["known_spells"] = save.KnownSpells
		saveMap["spell_slots"] = save.SpellSlots
		saveMap["locations_discovered"] = save.LocationsDiscovered
		saveMap["music_tracks_unlocked"] = save.MusicTracksUnlocked
		saveMap["current_day"] = save.CurrentDay
		saveMap["time_of_day"] = save.TimeOfDay
		savesWithID = append(savesWithID, saveMap)
	}

	json.NewEncoder(w).Encode(savesWithID)
}

// Create or update a save
func handleCreateSave(w http.ResponseWriter, r *http.Request, npub string) {
	// First decode into a flexible map to handle any structure
	var rawData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		log.Printf("❌ Error decoding save data: %v", err)
		http.Error(w, "Invalid save data", http.StatusBadRequest)
		return
	}


	// Convert back to JSON and then decode into SaveFile struct
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		log.Printf("❌ Error marshaling save data: %v", err)
		http.Error(w, "Invalid save data", http.StatusInternalServerError)
		return
	}

	var saveData SaveFile
	if err := json.Unmarshal(jsonData, &saveData); err != nil {
		log.Printf("❌ Error unmarshaling save data: %v", err)
		http.Error(w, "Invalid save data", http.StatusBadRequest)
		return
	}

	// Set internal metadata (not serialized to JSON)
	saveData.InternalNpub = npub

	// Check if 'id' was provided in the request (for overwrites)
	if id, ok := rawData["id"].(string); ok && id != "" {
		saveData.InternalID = id
		log.Printf("📝 Overwriting existing save: %s", id)
	} else if saveData.InternalID == "" {
		// Generate new save ID only if none provided
		saveData.InternalID = fmt.Sprintf("save_%d", time.Now().Unix())
		saveData.CreatedAt = time.Now().UTC().Format(time.RFC3339)
		log.Printf("✨ Creating new save: %s", saveData.InternalID)
	}

	// Ensure saves directory exists for this user
	userSavesDir := filepath.Join(SavesDirectory, npub)
	if err := os.MkdirAll(userSavesDir, 0755); err != nil {
		log.Printf("❌ Error creating saves directory: %v", err)
		http.Error(w, "Failed to create saves directory", http.StatusInternalServerError)
		return
	}

	// Write save file
	savePath := filepath.Join(userSavesDir, saveData.InternalID+".json")
	if err := writeSaveFile(savePath, &saveData); err != nil {
		log.Printf("❌ Error writing save file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Saved game for npub: %s, save ID: %s", npub, saveData.InternalID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"save_id": saveData.InternalID,
		"message": "Game saved successfully",
	})
}

// Delete a save
func handleDeleteSave(w http.ResponseWriter, _ *http.Request, npub string, saveID string) {
	savePath := filepath.Join(SavesDirectory, npub, saveID+".json")

	if err := os.Remove(savePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Save file not found", http.StatusNotFound)
		} else {
			log.Printf("❌ Error deleting save file: %v", err)
			http.Error(w, "Failed to delete save", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("🗑️ Deleted save: %s for npub: %s", saveID, npub)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Save deleted successfully",
	})
}

// Load specific save by ID
func LoadSaveByID(npub, saveID string) (*SaveFile, error) {
	savePath := filepath.Join(SavesDirectory, npub, saveID+".json")
	return loadSaveFile(savePath)
}

// Helper functions
func loadSaveFile(path string) (*SaveFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var save SaveFile
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, err
	}

	// Extract internal ID from filename
	filename := filepath.Base(path)
	save.InternalID = strings.TrimSuffix(filename, ".json")

	// Extract npub from directory path
	dir := filepath.Dir(path)
	save.InternalNpub = filepath.Base(dir)

	return &save, nil
}

func writeSaveFile(path string, save *SaveFile) error {
	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// Get save file info without full game state (for listings)
func GetSaveInfo(npub, saveID string) (*SaveFile, error) {
	save, err := LoadSaveByID(npub, saveID)
	if err != nil {
		return nil, err
	}

	// Return save info for listings
	return save, nil
}