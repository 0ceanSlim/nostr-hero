package charactereditor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pubkey-quest/cmd/codex/config"
	"pubkey-quest/cmd/codex/staging"
)

// HandleGetAllData returns all character generation data
func (e *Editor) HandleGetAllData(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"starting_gear":       e.StartingGear,
		"base_hp":             e.BaseHP,
		"starting_gold":       e.StartingGold,
		"generation_weights":  e.GenerationWeights,
		"introductions":       e.Introductions,
		"starting_locations":  e.StartingLocations,
		"starting_spells":     e.StartingSpells,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetStartingGear returns starting gear data
func (e *Editor) HandleGetStartingGear(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e.StartingGear)
}

// HandleSaveStartingGear saves starting gear data (staging-aware)
func (e *Editor) HandleSaveStartingGear(w http.ResponseWriter, r *http.Request) {
	var newData []StartingGearEntry
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate item IDs exist
	if err := e.validateStartingGear(newData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Detect mode
	cfg := e.Config.(*config.Config)
	mode := staging.DetectMode(r, cfg)
	sessionID := r.Header.Get("X-Session-ID")

	filePath := e.GetFilePath("starting-gear.json")
	newContent, err := json.MarshalIndent(newData, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// Add newline at end of file
	newContent = append(newContent, '\n')

	if mode == staging.ModeDirect {
		// Direct mode: save immediately
		if err := e.SaveFile("starting-gear.json", newContent); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save: %v", err), http.StatusInternalServerError)
			return
		}
		e.StartingGear = newData

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "saved",
			"mode":   "direct",
		})
	} else {
		// Staging mode: add to session
		session := staging.Manager.GetSession(sessionID)
		if session == nil {
			http.Error(w, "Invalid session", http.StatusBadRequest)
			return
		}

		oldContent, _ := os.ReadFile(filePath)
		gitPath := strings.ReplaceAll(filePath, "\\", "/")

		session.AddChange(staging.Change{
			Type:       staging.ChangeUpdate,
			FilePath:   gitPath,
			OldContent: oldContent,
			NewContent: newContent,
			Timestamp:  time.Now(),
		})

		e.StartingGear = newData

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "staged",
			"mode":    "staging",
			"changes": len(session.Changes),
		})
	}
}

// validateStartingGear validates that all item IDs exist
func (e *Editor) validateStartingGear(data []StartingGearEntry) error {
	// Build item ID set from game-data/items/
	validItems := make(map[string]bool)
	itemsPath := "game-data/items"

	err := filepath.Walk(itemsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			itemID := strings.TrimSuffix(filepath.Base(path), ".json")
			validItems[itemID] = true
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load items: %v", err)
	}

	// Validate each class's starting gear
	for _, entry := range data {
		for _, choice := range entry.StartingGear.EquipmentChoices {
			for _, option := range choice.Options {
				if err := validateOption(option, validItems); err != nil {
					return fmt.Errorf("class %s: %v", entry.Class, err)
				}
			}
		}

		// Validate given items
		for _, item := range entry.StartingGear.GivenItems {
			if !validItems[item.Item] {
				return fmt.Errorf("class %s: invalid given item ID '%s'", entry.Class, item.Item)
			}
		}

		// Validate pack choices (if present)
		if entry.StartingGear.PackChoice != nil {
			for _, packID := range entry.StartingGear.PackChoice.Options {
				if !validItems[packID] {
					return fmt.Errorf("class %s: invalid pack ID '%s'", entry.Class, packID)
				}
			}
		}
	}

	return nil
}

// validateOption validates a single equipment option
func validateOption(opt Option, validItems map[string]bool) error {
	switch opt.Type {
	case "single":
		if opt.Item != "" && !validItems[opt.Item] {
			return fmt.Errorf("invalid item ID '%s'", opt.Item)
		}
	case "bundle":
		for _, item := range opt.Items {
			if !validItems[item.Item] {
				return fmt.Errorf("invalid item ID '%s'", item.Item)
			}
		}
	case "multi_slot":
		for _, slot := range opt.Slots {
			if slot.Type == "weapon_choice" {
				for _, weaponID := range slot.Options {
					if !validItems[weaponID] {
						return fmt.Errorf("invalid weapon ID '%s'", weaponID)
					}
				}
			} else if slot.Type == "fixed" {
				if slot.Item != "" && !validItems[slot.Item] {
					return fmt.Errorf("invalid item ID '%s'", slot.Item)
				}
			}
		}
	default:
		return fmt.Errorf("invalid option type '%s'", opt.Type)
	}
	return nil
}

// HandleGetOtherFile returns a specific character data file as raw JSON
func (e *Editor) HandleGetOtherFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data json.RawMessage
		switch filename {
		case "base-hp":
			data = e.BaseHP
		case "starting-gold":
			data = e.StartingGold
		case "generation-weights":
			data = e.GenerationWeights
		case "introductions":
			data = e.Introductions
		case "starting-locations":
			data = e.StartingLocations
		case "starting-spells":
			data = e.StartingSpells
		default:
			http.Error(w, "Unknown file", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

// HandleSaveOtherFile saves a character data file (staging-aware, basic implementation)
func (e *Editor) HandleSaveOtherFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newData json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Detect mode
		cfg := e.Config.(*config.Config)
		mode := staging.DetectMode(r, cfg)
		sessionID := r.Header.Get("X-Session-ID")

		jsonFilename := filename + ".json"
		filePath := e.GetFilePath(jsonFilename)

		// Pretty-print JSON
		var prettyJSON interface{}
		json.Unmarshal(newData, &prettyJSON)
		newContent, err := json.MarshalIndent(prettyJSON, "", "  ")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
			return
		}
		newContent = append(newContent, '\n')

		if mode == staging.ModeDirect {
			if err := e.SaveFile(jsonFilename, newContent); err != nil {
				http.Error(w, fmt.Sprintf("Failed to save: %v", err), http.StatusInternalServerError)
				return
			}

			// Update in-memory data
			switch filename {
			case "base-hp":
				e.BaseHP = newData
			case "starting-gold":
				e.StartingGold = newData
			case "generation-weights":
				e.GenerationWeights = newData
			case "introductions":
				e.Introductions = newData
			case "starting-locations":
				e.StartingLocations = newData
			case "starting-spells":
				e.StartingSpells = newData
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "saved",
				"mode":   "direct",
			})
		} else {
			session := staging.Manager.GetSession(sessionID)
			if session == nil {
				http.Error(w, "Invalid session", http.StatusBadRequest)
				return
			}

			oldContent, _ := os.ReadFile(filePath)
			gitPath := strings.ReplaceAll(filePath, "\\", "/")

			session.AddChange(staging.Change{
				Type:       staging.ChangeUpdate,
				FilePath:   gitPath,
				OldContent: oldContent,
				NewContent: newContent,
				Timestamp:  time.Now(),
			})

			// Update in-memory data
			switch filename {
			case "base-hp":
				e.BaseHP = newData
			case "starting-gold":
				e.StartingGold = newData
			case "generation-weights":
				e.GenerationWeights = newData
			case "introductions":
				e.Introductions = newData
			case "starting-locations":
				e.StartingLocations = newData
			case "starting-spells":
				e.StartingSpells = newData
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "staged",
				"mode":    "staging",
				"changes": len(session.Changes),
			})
		}
	}
}
