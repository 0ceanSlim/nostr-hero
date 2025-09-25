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
	ID         string                 `json:"id"`
	Npub       string                 `json:"npub"`
	Character  map[string]interface{} `json:"character"`
	GameState  map[string]interface{} `json:"gameState"`
	Location   string                 `json:"location"`
	LastPlayed string                 `json:"last_played"`
	CreatedAt  string                 `json:"created_at"`
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
	json.NewEncoder(w).Encode(saves)
}

// Create or update a save
func handleCreateSave(w http.ResponseWriter, r *http.Request, npub string) {
	var saveData SaveFile
	if err := json.NewDecoder(r.Body).Decode(&saveData); err != nil {
		log.Printf("❌ Error decoding save data: %v", err)
		http.Error(w, "Invalid save data", http.StatusBadRequest)
		return
	}

	// Set metadata
	saveData.Npub = npub
	saveData.LastPlayed = time.Now().UTC().Format(time.RFC3339)

	if saveData.ID == "" {
		// Generate new save ID
		saveData.ID = fmt.Sprintf("save_%d", time.Now().Unix())
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
	savePath := filepath.Join(userSavesDir, saveData.ID+".json")
	if err := writeSaveFile(savePath, &saveData); err != nil {
		log.Printf("❌ Error writing save file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Saved game for npub: %s, save ID: %s", npub, saveData.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"save_id": saveData.ID,
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

	// Return save without full game state for performance
	return &SaveFile{
		ID:         save.ID,
		Npub:       save.Npub,
		Character:  save.Character,
		Location:   save.Location,
		LastPlayed: save.LastPlayed,
		CreatedAt:  save.CreatedAt,
	}, nil
}