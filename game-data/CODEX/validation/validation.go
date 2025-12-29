package validation

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Issue represents a validation issue found in game data
type Issue struct {
	Type     string `json:"type"`     // "error", "warning", "info"
	Category string `json:"category"` // "items", "spells", "monsters", etc.
	File     string `json:"file"`
	Field    string `json:"field,omitempty"`
	Message  string `json:"message"`
}

// Result holds the validation results
type Result struct {
	Issues []Issue `json:"issues"`
	Stats  Stats   `json:"stats"`
}

// Stats holds statistics about validation
type Stats struct {
	TotalFiles   int `json:"total_files"`
	ErrorCount   int `json:"error_count"`
	WarningCount int `json:"warning_count"`
	InfoCount    int `json:"info_count"`
}

// ValidateAll runs all validation checks on game data
func ValidateAll() (*Result, error) {
	result := &Result{
		Issues: []Issue{},
	}

	// Validate items
	if itemIssues, err := ValidateItems(); err != nil {
		return nil, err
	} else {
		result.Issues = append(result.Issues, itemIssues...)
	}

	// Validate spells
	if spellIssues, err := ValidateSpells(); err != nil {
		return nil, err
	} else {
		result.Issues = append(result.Issues, spellIssues...)
	}

	// Validate monsters
	if monsterIssues, err := ValidateMonsters(); err != nil {
		return nil, err
	} else {
		result.Issues = append(result.Issues, monsterIssues...)
	}

	// Validate locations
	if locationIssues, err := ValidateLocations(); err != nil {
		return nil, err
	} else {
		result.Issues = append(result.Issues, locationIssues...)
	}

	// Validate NPCs
	if npcIssues, err := ValidateNPCs(); err != nil {
		return nil, err
	} else {
		result.Issues = append(result.Issues, npcIssues...)
	}

	// Calculate stats
	for _, issue := range result.Issues {
		result.Stats.TotalFiles++
		switch issue.Type {
		case "error":
			result.Stats.ErrorCount++
		case "warning":
			result.Stats.WarningCount++
		case "info":
			result.Stats.InfoCount++
		}
	}

	return result, nil
}

// ValidateItems validates all item files
func ValidateItems() ([]Issue, error) {
	issues := []Issue{}
	itemsPath := "../items"

	err := filepath.WalkDir(itemsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			itemIssues := validateItemFile(path)
			issues = append(issues, itemIssues...)
		}
		return nil
	})

	return issues, err
}

func validateItemFile(filePath string) []Issue {
	issues := []Issue{}
	filename := filepath.Base(filePath)
	idFromFilename := strings.TrimSuffix(filename, ".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "items",
			File:     filename,
			Message:  fmt.Sprintf("Failed to read file: %v", err),
		})
		return issues
	}

	var item map[string]interface{}
	if err := json.Unmarshal(data, &item); err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "items",
			File:     filename,
			Message:  fmt.Sprintf("Invalid JSON: %v", err),
		})
		return issues
	}

	// Check required fields
	requiredFields := []string{"id", "name", "type", "price", "weight", "stack", "rarity"}
	for _, field := range requiredFields {
		if _, exists := item[field]; !exists {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "items",
				File:     filename,
				Field:    field,
				Message:  fmt.Sprintf("Missing required field: %s", field),
			})
		}
	}

	// Check ID matches filename
	if id, ok := item["id"].(string); ok {
		if id != idFromFilename {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "items",
				File:     filename,
				Field:    "id",
				Message:  fmt.Sprintf("ID '%s' doesn't match filename '%s'", id, idFromFilename),
			})
		}
	}

	// Check price is non-negative
	if price, ok := item["price"].(float64); ok {
		if price < 0 {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "items",
				File:     filename,
				Field:    "price",
				Message:  "Price cannot be negative",
			})
		}
	}

	// Check weight is non-negative
	if weight, ok := item["weight"].(float64); ok {
		if weight < 0 {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "items",
				File:     filename,
				Field:    "weight",
				Message:  "Weight cannot be negative",
			})
		}
	}

	// Check rarity is valid
	validRarities := map[string]bool{
		"common": true, "uncommon": true, "rare": true,
		"very rare": true, "legendary": true,
	}
	if rarity, ok := item["rarity"].(string); ok {
		if !validRarities[strings.ToLower(rarity)] {
			issues = append(issues, Issue{
				Type:     "warning",
				Category: "items",
				File:     filename,
				Field:    "rarity",
				Message:  fmt.Sprintf("Non-standard rarity: %s", rarity),
			})
		}
	}

	// Check if image exists (warning, not error)
	if image, ok := item["image"].(string); ok && image != "" {
		imagePath := filepath.Join("../../www/res/img/items", idFromFilename+".png")
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			issues = append(issues, Issue{
				Type:     "warning",
				Category: "items",
				File:     filename,
				Field:    "image",
				Message:  "Image file not found",
			})
		}
	}

	return issues
}

// ValidateSpells validates all spell files
func ValidateSpells() ([]Issue, error) {
	issues := []Issue{}
	spellsPath := "../magic/spells"

	err := filepath.WalkDir(spellsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			spellIssues := validateSpellFile(path)
			issues = append(issues, spellIssues...)
		}
		return nil
	})

	return issues, err
}

func validateSpellFile(filePath string) []Issue {
	issues := []Issue{}
	filename := filepath.Base(filePath)
	idFromFilename := strings.TrimSuffix(filename, ".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "spells",
			File:     filename,
			Message:  fmt.Sprintf("Failed to read file: %v", err),
		})
		return issues
	}

	var spell map[string]interface{}
	if err := json.Unmarshal(data, &spell); err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "spells",
			File:     filename,
			Message:  fmt.Sprintf("Invalid JSON: %v", err),
		})
		return issues
	}

	// Check required fields
	requiredFields := []string{"id", "name", "level", "school"}
	for _, field := range requiredFields {
		if _, exists := spell[field]; !exists {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "spells",
				File:     filename,
				Field:    field,
				Message:  fmt.Sprintf("Missing required field: %s", field),
			})
		}
	}

	// Check ID matches filename
	if id, ok := spell["id"].(string); ok {
		if id != idFromFilename {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "spells",
				File:     filename,
				Field:    "id",
				Message:  fmt.Sprintf("ID '%s' doesn't match filename '%s'", id, idFromFilename),
			})
		}
	}

	// Check level is valid (0-9 for D&D 5e)
	if level, ok := spell["level"].(float64); ok {
		if level < 0 || level > 9 {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "spells",
				File:     filename,
				Field:    "level",
				Message:  fmt.Sprintf("Invalid spell level: %v (must be 0-9)", level),
			})
		}
	}

	return issues
}

// ValidateMonsters validates all monster files
func ValidateMonsters() ([]Issue, error) {
	issues := []Issue{}
	monstersPath := "../monsters"

	err := filepath.WalkDir(monstersPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			monsterIssues := validateMonsterFile(path)
			issues = append(issues, monsterIssues...)
		}
		return nil
	})

	return issues, err
}

func validateMonsterFile(filePath string) []Issue {
	issues := []Issue{}
	filename := filepath.Base(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "monsters",
			File:     filename,
			Message:  fmt.Sprintf("Failed to read file: %v", err),
		})
		return issues
	}

	var monster map[string]interface{}
	if err := json.Unmarshal(data, &monster); err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "monsters",
			File:     filename,
			Message:  fmt.Sprintf("Invalid JSON: %v", err),
		})
		return issues
	}

	// Check required fields
	requiredFields := []string{"id", "name"}
	for _, field := range requiredFields {
		if _, exists := monster[field]; !exists {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "monsters",
				File:     filename,
				Field:    field,
				Message:  fmt.Sprintf("Missing required field: %s", field),
			})
		}
	}

	return issues
}

// ValidateLocations validates all location files
func ValidateLocations() ([]Issue, error) {
	issues := []Issue{}
	locationsPath := "../locations"

	// Check cities and environments
	subDirs := []string{"cities", "environments"}
	for _, subDir := range subDirs {
		dirPath := filepath.Join(locationsPath, subDir)
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && strings.HasSuffix(path, ".json") {
				locationIssues := validateLocationFile(path)
				issues = append(issues, locationIssues...)
			}
			return nil
		})

		if err != nil {
			return issues, err
		}
	}

	return issues, nil
}

func validateLocationFile(filePath string) []Issue {
	issues := []Issue{}
	filename := filepath.Base(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "locations",
			File:     filename,
			Message:  fmt.Sprintf("Failed to read file: %v", err),
		})
		return issues
	}

	var location map[string]interface{}
	if err := json.Unmarshal(data, &location); err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "locations",
			File:     filename,
			Message:  fmt.Sprintf("Invalid JSON: %v", err),
		})
		return issues
	}

	// Check required fields
	requiredFields := []string{"id", "name"}
	for _, field := range requiredFields {
		if _, exists := location[field]; !exists {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "locations",
				File:     filename,
				Field:    field,
				Message:  fmt.Sprintf("Missing required field: %s", field),
			})
		}
	}

	return issues
}

// ValidateNPCs validates all NPC files
func ValidateNPCs() ([]Issue, error) {
	issues := []Issue{}
	npcsPath := "../npcs"

	err := filepath.WalkDir(npcsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			npcIssues := validateNPCFile(path)
			issues = append(issues, npcIssues...)
		}
		return nil
	})

	return issues, err
}

func validateNPCFile(filePath string) []Issue {
	issues := []Issue{}
	filename := filepath.Base(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "npcs",
			File:     filename,
			Message:  fmt.Sprintf("Failed to read file: %v", err),
		})
		return issues
	}

	var npc map[string]interface{}
	if err := json.Unmarshal(data, &npc); err != nil {
		issues = append(issues, Issue{
			Type:     "error",
			Category: "npcs",
			File:     filename,
			Message:  fmt.Sprintf("Invalid JSON: %v", err),
		})
		return issues
	}

	// Check required fields
	requiredFields := []string{"id", "name"}
	for _, field := range requiredFields {
		if _, exists := npc[field]; !exists {
			issues = append(issues, Issue{
				Type:     "error",
				Category: "npcs",
				File:     filename,
				Field:    field,
				Message:  fmt.Sprintf("Missing required field: %s", field),
			})
		}
	}

	return issues
}
