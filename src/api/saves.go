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
	Fatigue             int                    `json:"fatigue"`
	Gold                int                    `json:"gold"`
	Stats               map[string]interface{} `json:"stats"`
	Location            string                 `json:"location"`
	Inventory           map[string]interface{} `json:"inventory"`
	KnownSpells         []string               `json:"known_spells"`
	SpellSlots          map[string]interface{} `json:"spell_slots"`
	LocationsDiscovered []string               `json:"locations_discovered"`
	MusicTracksUnlocked []string               `json:"music_tracks_unlocked"`
	InternalID          string                 `json:"-"` // Not serialized, used internally for file naming
	InternalNpub        string                 `json:"-"` // Not serialized, used internally for directory structure
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
func handleGetSaves(w http.ResponseWriter, r *http.Request, npub string) {
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
		saveMap["gold"] = save.Gold
		saveMap["stats"] = save.Stats
		saveMap["location"] = save.Location
		saveMap["inventory"] = save.Inventory
		saveMap["known_spells"] = save.KnownSpells
		saveMap["spell_slots"] = save.SpellSlots
		saveMap["locations_discovered"] = save.LocationsDiscovered
		saveMap["music_tracks_unlocked"] = save.MusicTracksUnlocked
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

	// Log the incoming data for debugging
	log.Printf("📥 Received save data: %+v", rawData)

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

	if saveData.InternalID == "" {
		// Generate new save ID
		saveData.InternalID = fmt.Sprintf("save_%d", time.Now().Unix())
		saveData.CreatedAt = time.Now().UTC().Format(time.RFC3339)
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
func handleDeleteSave(w http.ResponseWriter, r *http.Request, npub string, saveID string) {
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