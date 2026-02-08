package charactereditor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Editor manages character generation data files
type Editor struct {
	StartingGear      []StartingGearEntry
	BaseHP            json.RawMessage
	StartingGold      json.RawMessage
	GenerationWeights json.RawMessage
	Introductions     json.RawMessage
	StartingLocations json.RawMessage
	StartingSpells    json.RawMessage
	Config            interface{} // *config.Config
	basePath          string
}

// NewEditor creates a new character editor instance
func NewEditor(config interface{}) *Editor {
	return &Editor{
		Config:   config,
		basePath: "game-data/systems/new-character",
	}
}

// LoadAll loads all character generation data files
func (e *Editor) LoadAll() error {
	// Load starting gear (typed struct)
	if err := e.loadStartingGear(); err != nil {
		return fmt.Errorf("failed to load starting-gear.json: %v", err)
	}

	// Load other files as raw JSON
	files := map[string]*json.RawMessage{
		"base-hp.json":           &e.BaseHP,
		"starting-gold.json":     &e.StartingGold,
		"generation-weights.json": &e.GenerationWeights,
		"introductions.json":     &e.Introductions,
		"starting-locations.json": &e.StartingLocations,
		"starting-spells.json":   &e.StartingSpells,
	}

	for filename, target := range files {
		if err := e.loadRawJSON(filename, target); err != nil {
			return fmt.Errorf("failed to load %s: %v", filename, err)
		}
	}

	return nil
}

// loadStartingGear loads and parses starting-gear.json
func (e *Editor) loadStartingGear() error {
	path := filepath.Join(e.basePath, "starting-gear.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &e.StartingGear)
}

// loadRawJSON loads a JSON file as raw bytes
func (e *Editor) loadRawJSON(filename string, target *json.RawMessage) error {
	path := filepath.Join(e.basePath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	*target = json.RawMessage(data)
	return nil
}

// SaveFile writes content to a character data file
func (e *Editor) SaveFile(filename string, content []byte) error {
	path := filepath.Join(e.basePath, filename)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// GetFilePath returns the full path for a character data file
func (e *Editor) GetFilePath(filename string) string {
	return filepath.Join(e.basePath, filename)
}
