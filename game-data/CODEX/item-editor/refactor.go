package itemeditor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Reference represents a reference to an item ID found in other files
type Reference struct {
	File     string `json:"file"`
	Location string `json:"location"`
	OldID    string `json:"oldId"`
	NewID    string `json:"newId"`
}

// RefactorPreview shows what will change during ID refactoring
type RefactorPreview struct {
	OldID      string      `json:"oldId"`
	NewID      string      `json:"newId"`
	References []Reference `json:"references"`
	WillRename string      `json:"willRename"`
}

// StartingGear represents the starting gear structure
type StartingGear []ClassData

type ClassData struct {
	Class        string     `json:"class"`
	StartingGear []GearItem `json:"starting_gear"`
}

type GearItem struct {
	Given  interface{}   `json:"given,omitempty"`
	Option []interface{} `json:"option,omitempty"`
}

// HandleRefactorPreview generates a preview of ID refactoring
func (e *Editor) HandleRefactorPreview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename string `json:"filename"`
		OldID    string `json:"oldId"`
		NewID    string `json:"newId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	preview, err := e.generateRefactorPreview(req.OldID, req.NewID, req.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// HandleRefactorApply applies an ID refactoring
func (e *Editor) HandleRefactorApply(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename string           `json:"filename"`
		Preview  *RefactorPreview `json:"preview"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := e.applyRefactor(req.Preview, req.Filename); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload items
	e.LoadItems()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// generateRefactorPreview generates a preview of what will change
func (e *Editor) generateRefactorPreview(oldID, newID, filename string) (*RefactorPreview, error) {
	preview := &RefactorPreview{
		OldID:      oldID,
		NewID:      newID,
		References: []Reference{},
		WillRename: fmt.Sprintf("%s.json → %s.json", filename, newID),
	}

	// Scan starting gear for references
	if refs, err := e.scanStartingGearReferences(oldID, newID); err != nil {
		return nil, fmt.Errorf("error scanning starting gear: %v", err)
	} else {
		preview.References = append(preview.References, refs...)
	}

	// Scan pack contents for references
	if refs, err := e.scanPackReferences(oldID, newID); err != nil {
		return nil, fmt.Errorf("error scanning pack contents: %v", err)
	} else {
		preview.References = append(preview.References, refs...)
	}

	return preview, nil
}

// scanStartingGearReferences scans starting-gear.json for references
func (e *Editor) scanStartingGearReferences(oldID, newID string) ([]Reference, error) {
	startingGearPath := "game-data/systems/new-character/starting-gear.json"

	data, err := os.ReadFile(startingGearPath)
	if err != nil {
		return nil, err
	}

	var startingGear StartingGear
	if err := json.Unmarshal(data, &startingGear); err != nil {
		return nil, err
	}

	var references []Reference

	for classIndex, classData := range startingGear {
		for gearIndex, gearItem := range classData.StartingGear {
			// Check given items
			if gearItem.Given != nil {
				references = append(references, e.scanGivenReferences(gearItem.Given, oldID, newID, fmt.Sprintf("[%d].starting_gear[%d].given", classIndex, gearIndex))...)
			}

			// Check option items
			references = append(references, e.scanOptionReferences(gearItem.Option, oldID, newID, fmt.Sprintf("[%d].starting_gear[%d].option", classIndex, gearIndex))...)
		}
	}

	return references, nil
}

// scanGivenReferences scans given items for references
func (e *Editor) scanGivenReferences(given interface{}, oldID, newID, basePath string) []Reference {
	var references []Reference

	switch g := given.(type) {
	case []interface{}:
		for i, item := range g {
			if subArray, ok := item.([]interface{}); ok && len(subArray) >= 2 {
				if itemID, ok := subArray[0].(string); ok && itemID == oldID {
					references = append(references, Reference{
						File:     "starting-gear.json",
						Location: fmt.Sprintf("%s[%d][0]", basePath, i),
						OldID:    oldID,
						NewID:    newID,
					})
				}
			} else if i == 0 {
				if itemID, ok := item.(string); ok && itemID == oldID {
					references = append(references, Reference{
						File:     "starting-gear.json",
						Location: fmt.Sprintf("%s[0]", basePath),
						OldID:    oldID,
						NewID:    newID,
					})
				}
			}
		}
	}

	return references
}

// scanOptionReferences scans option items for references
func (e *Editor) scanOptionReferences(option interface{}, oldID, newID, basePath string) []Reference {
	var references []Reference

	switch opt := option.(type) {
	case []interface{}:
		for i, item := range opt {
			if subArray, ok := item.([]interface{}); ok && len(subArray) >= 2 {
				if itemID, ok := subArray[0].(string); ok && itemID == oldID {
					references = append(references, Reference{
						File:     "starting-gear.json",
						Location: fmt.Sprintf("%s[%d][0]", basePath, i),
						OldID:    oldID,
						NewID:    newID,
					})
				}
			} else {
				references = append(references, e.scanOptionReferences(item, oldID, newID, fmt.Sprintf("%s[%d]", basePath, i))...)
			}
		}
	}

	return references
}

// scanPackReferences scans pack contents for references
func (e *Editor) scanPackReferences(oldID, newID string) ([]Reference, error) {
	var references []Reference

	for filename, item := range e.Items {
		if item.Contents != nil {
			for contentIndex, content := range item.Contents {
				if len(content) >= 2 {
					if itemID, ok := content[0].(string); ok && itemID == oldID {
						references = append(references, Reference{
							File:     fmt.Sprintf("%s.json", filename),
							Location: fmt.Sprintf("contents[%d][0]", contentIndex),
							OldID:    oldID,
							NewID:    newID,
						})
					}
				}
			}
		}
	}

	return references, nil
}

// applyRefactor applies the refactoring changes
func (e *Editor) applyRefactor(preview *RefactorPreview, filename string) error {
	// 1. Update the item's ID
	item := e.Items[filename]
	item.ID = preview.NewID

	// 2. Save the item with new filename
	newFilename := preview.NewID
	if err := e.SaveItemToFile(newFilename, item); err != nil {
		return fmt.Errorf("failed to save item with new filename: %v", err)
	}

	// 3. Delete old file
	oldFilePath := filepath.Join("game-data/items", filename+".json")
	if err := os.Remove(oldFilePath); err != nil {
		return fmt.Errorf("failed to remove old file: %v", err)
	}

	// 4. Update starting gear references
	if err := e.updateStartingGearReferences(preview); err != nil {
		return fmt.Errorf("failed to update starting gear: %v", err)
	}

	// 5. Update pack references
	if err := e.updatePackReferences(preview); err != nil {
		return fmt.Errorf("failed to update pack contents: %v", err)
	}

	// 6. Update internal state
	delete(e.Items, filename)
	e.Items[newFilename] = item

	return nil
}

// updateStartingGearReferences updates references in starting-gear.json
func (e *Editor) updateStartingGearReferences(preview *RefactorPreview) error {
	startingGearPath := "game-data/systems/new-character/starting-gear.json"

	data, err := os.ReadFile(startingGearPath)
	if err != nil {
		return err
	}

	// Simple string replacement
	oldPattern := fmt.Sprintf(`"%s"`, preview.OldID)
	newPattern := fmt.Sprintf(`"%s"`, preview.NewID)

	updatedData := strings.ReplaceAll(string(data), oldPattern, newPattern)

	return os.WriteFile(startingGearPath, []byte(updatedData), 0644)
}

// updatePackReferences updates references in pack contents
func (e *Editor) updatePackReferences(preview *RefactorPreview) error {
	for filename, item := range e.Items {
		updated := false

		if item.Contents != nil {
			for _, content := range item.Contents {
				if len(content) >= 2 {
					if itemID, ok := content[0].(string); ok && itemID == preview.OldID {
						content[0] = preview.NewID
						updated = true
					}
				}
			}
		}

		if updated {
			if err := e.SaveItemToFile(filename, item); err != nil {
				return fmt.Errorf("failed to update %s: %v", filename, err)
			}
		}
	}

	return nil
}
