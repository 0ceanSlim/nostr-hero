package itemeditor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nostr-hero/codex/pixellab"

	"github.com/gorilla/mux"
)

// HandleItemEditor renders the item editor UI
func (e *Editor) HandleItemEditor(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/item-editor.html")
}

// HandleGetItems returns all items as JSON
func (e *Editor) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e.Items)
}

// HandleGetItem returns a specific item
func (e *Editor) HandleGetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.Items[filename]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// HandleSaveItem saves an item to disk
func (e *Editor) HandleSaveItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save to file
	if err := e.SaveItemToFile(filename, &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update in memory
	e.Items[filename] = &item

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleValidate validates all items
func (e *Editor) HandleValidate(w http.ResponseWriter, r *http.Request) {
	issues := []string{}

	for filename, item := range e.Items {
		if item.ID != filename {
			issues = append(issues, fmt.Sprintf("%s: ID '%s' doesn't match filename", filename, item.ID))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issues": issues,
	})
}

// HandleGetTypes returns all unique item types
func (e *Editor) HandleGetTypes(w http.ResponseWriter, r *http.Request) {
	types := make(map[string]bool)
	for _, item := range e.Items {
		if item.Type != "" {
			types[item.Type] = true
		}
	}

	typeList := make([]string, 0, len(types))
	for t := range types {
		typeList = append(typeList, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(typeList)
}

// HandleGetTags returns all unique tags
func (e *Editor) HandleGetTags(w http.ResponseWriter, r *http.Request) {
	tags := make(map[string]bool)
	for _, item := range e.Items {
		for _, tag := range item.Tags {
			if tag != "" {
				tags[tag] = true
			}
		}
	}

	tagList := make([]string, 0, len(tags))
	for t := range tags {
		tagList = append(tagList, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tagList)
}

// PixelLab image generation handlers

// HandleGetBalance gets the PixelLab account balance
func (e *Editor) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	if e.PixelLabClient == nil {
		http.Error(w, "PixelLab client not initialized", http.StatusServiceUnavailable)
		return
	}

	balance, err := e.PixelLabClient.GetBalance()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// HandleGenerateImage generates an image for an item
func (e *Editor) HandleGenerateImage(w http.ResponseWriter, r *http.Request) {
	if e.PixelLabClient == nil {
		http.Error(w, "PixelLab client not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.Items[filename]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	var req struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = "bitforge"
	}

	log.Printf("🎨 Generating image for %s using %s...", item.Name, req.Model)

	prompt := pixellab.GeneratePrompt(item.Name, item.Description, item.AIDescription, item.Rarity)
	negativePrompt := pixellab.NegativePrompt()

	result, err := e.PixelLabClient.GenerateImage(prompt, negativePrompt, req.Model)
	if err != nil {
		log.Printf("❌ Error generating image: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save image with timestamp to history folder
	timestamp := time.Now().Format("20060102_150405")
	historyDir := filepath.Join("../../www/res/img/items/_history", item.ID)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	historyFile := filepath.Join(historyDir, fmt.Sprintf("%s_%s.png", timestamp, req.Model))
	imageData, err := base64.StdEncoding.DecodeString(result.Image.Base64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(historyFile, imageData, 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Image generated successfully ($%.4f) - saved to history", result.Usage.USD)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"cost":      result.Usage.USD,
		"imagePath": historyFile,
		"imageData": result.Image.Base64,
		"prompt":    prompt,
	})
}

// HandleGetImage checks if an image exists for an item
func (e *Editor) HandleGetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.Items[filename]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Check if image exists
	imagePath := filepath.Join("../../www/res/img/items", item.ID+".png")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"exists": false,
		})
		return
	}

	// Also check for images in history
	historyDir := filepath.Join("../../www/res/img/items/_history", item.ID)
	var historyFiles []string
	if entries, err := os.ReadDir(historyDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
				historyFiles = append(historyFiles, entry.Name())
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists":       true,
		"path":         imagePath,
		"historyFiles": historyFiles,
	})
}

// HandleAcceptImage accepts a generated image
func (e *Editor) HandleAcceptImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.Items[filename]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	var req struct {
		ImageData string `json:"imageData"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Decode the base64 image data
	imageData, err := base64.StdEncoding.DecodeString(req.ImageData)
	if err != nil {
		http.Error(w, "Invalid image data", http.StatusBadRequest)
		return
	}

	// Save to main items directory
	mainImagePath := filepath.Join("../../www/res/img/items", item.ID+".png")
	if err := os.WriteFile(mainImagePath, imageData, 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Image accepted for %s", item.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"path":    mainImagePath,
	})
}
