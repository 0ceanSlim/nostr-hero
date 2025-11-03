package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

// Item represents a game item with all possible fields
type Item struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	AIDescription string              `json:"ai_description,omitempty"`
	Price         int                 `json:"price"`
	Type          string              `json:"type"`
	Weight        float64             `json:"weight"`
	Stack         int                 `json:"stack"`
	GearSlot      string              `json:"gear_slot,omitempty"`
	Rarity        string              `json:"rarity"`
	Tags          []string            `json:"tags,omitempty"`
	Contents      [][]interface{}     `json:"contents,omitempty"`
	AC            interface{}         `json:"ac,omitempty"`
	Damage        interface{}         `json:"damage,omitempty"`
	DamageType    string              `json:"damage-type,omitempty"`
	Heal          interface{}         `json:"heal,omitempty"`
	Ammunition    string              `json:"ammunition,omitempty"`
	Range         string              `json:"range,omitempty"`
	RangeLong     string              `json:"range-long,omitempty"`
	Img           string              `json:"img,omitempty"`
	Image         string              `json:"image,omitempty"`
	Notes         []string            `json:"notes,omitempty"`
	// Dynamic fields for any additional properties
	Extra         map[string]interface{} `json:"-"`
}

// Reference represents a reference to an item ID found in other files
type Reference struct {
	File     string `json:"file"`
	Location string `json:"location"`
	OldID    string `json:"oldId"`
	NewID    string `json:"newId"`
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

// RefactorPreview shows what will change during ID refactoring
type RefactorPreview struct {
	OldID      string      `json:"oldId"`
	NewID      string      `json:"newId"`
	References []Reference `json:"references"`
	WillRename string      `json:"willRename"`
}

// Config for PixelLab API
type Config struct {
	PixelLab struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"pixellab"`
}

// PixelLabResponse from API
type PixelLabResponse struct {
	Usage struct {
		USD float64 `json:"usd"`
	} `json:"usage"`
	Image struct {
		Base64 string `json:"base64"`
	} `json:"image"`
}

// BalanceResponse from PixelLab
type BalanceResponse struct {
	Type string  `json:"type"`
	USD  float64 `json:"usd"`
}

// PixelLabClient for image generation
type PixelLabClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// NewPixelLabClient creates a new PixelLab client
func NewPixelLabClient(apiKey string) *PixelLabClient {
	return &PixelLabClient{
		APIKey:  apiKey,
		BaseURL: "https://api.pixellab.ai/v1",
		Client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// GetBalance checks API balance
func (c *PixelLabClient) GetBalance() (*BalanceResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/balance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var balance BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// GenerateImage generates a pixel art image
func (c *PixelLabClient) GenerateImage(description, negativePrompt string, model string) (*PixelLabResponse, error) {
	var endpoint string
	var payload map[string]interface{}

	basePayload := map[string]interface{}{
		"description": description,
		"image_size": map[string]int{
			"width":  32,
			"height": 32,
		},
		"no_background": true,
		"detail":        "highly detailed",
		"outline":       "single color black outline",
	}

	if negativePrompt != "" {
		basePayload["negative_description"] = negativePrompt
	}

	switch model {
	case "bitforge":
		endpoint = "/generate-image-bitforge"
		payload = basePayload
	case "pixflux":
		endpoint = "/generate-image-pixflux"
		payload = basePayload
	default:
		return nil, fmt.Errorf("unsupported model: %s", model)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PixelLabResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ItemEditor holds the application state
type ItemEditor struct {
	items         map[string]*Item
	pixellabClient *PixelLabClient
}

var editor *ItemEditor

func loadConfig() (*Config, error) {
	configPath := "config.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("⚠️ config.yml not found - image generation will be disabled")
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.PixelLab.APIKey == "" {
		log.Printf("⚠️ pixellab.api_key not found in config.yml - image generation will be disabled")
		return nil, nil
	}

	return &config, nil
}

func generateFantasyPrompt(item *Item) string {
	// Priority 1: Use ai_description if available
	if item.AIDescription != "" && len(item.AIDescription) > 10 {
		return item.AIDescription
	}

	// Priority 2: Use regular description if available
	if item.Description != "" && len(item.Description) > 10 {
		return item.Description
	}

	// Fallback: Generate from item name and type
	rarity := strings.ToLower(item.Rarity)
	var rarityDesc string
	switch rarity {
	case "common":
		rarityDesc = "simple, basic"
	case "uncommon":
		rarityDesc = "well-crafted, slightly ornate"
	case "rare":
		rarityDesc = "ornate, decorated"
	case "very rare":
		rarityDesc = "highly ornate, magical aura"
	case "legendary":
		rarityDesc = "legendary, glowing, magical effects"
	default:
		rarityDesc = "well-made"
	}

	return fmt.Sprintf("%s %s", rarityDesc, item.Name)
}

func generateNegativePrompt() string {
	return "blurry, fuzzy, soft, antialiased, smooth, low quality, modern, realistic, photograph, 3d render, low resolution, text, letters, words, people, characters, faces, anime, cartoon"
}

func main() {
	editor = &ItemEditor{
		items: make(map[string]*Item),
	}

	if err := editor.loadItems(); err != nil {
		log.Fatal(err)
	}

	// Try to load config for PixelLab
	config, err := loadConfig()
	if err != nil {
		log.Printf("⚠️ Error loading config: %v - continuing without image generation", err)
	} else if config != nil {
		editor.pixellabClient = NewPixelLabClient(config.PixelLab.APIKey)
		log.Printf("✅ PixelLab client initialized")
	}

	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/items", editor.handleGetItems).Methods("GET")
	r.HandleFunc("/api/items/{filename}", editor.handleGetItem).Methods("GET")
	r.HandleFunc("/api/items/{filename}", editor.handleSaveItem).Methods("PUT")
	r.HandleFunc("/api/validate", editor.handleValidate).Methods("GET")
	r.HandleFunc("/api/types", editor.handleGetTypes).Methods("GET")
	r.HandleFunc("/api/tags", editor.handleGetTags).Methods("GET")
	r.HandleFunc("/api/refactor/preview", editor.handleRefactorPreview).Methods("POST")
	r.HandleFunc("/api/refactor/apply", editor.handleRefactorApply).Methods("POST")

	// Image generation routes
	r.HandleFunc("/api/balance", editor.handleGetBalance).Methods("GET")
	r.HandleFunc("/api/items/{filename}/generate-image", editor.handleGenerateImage).Methods("POST")
	r.HandleFunc("/api/items/{filename}/image", editor.handleGetImage).Methods("GET")
	r.HandleFunc("/api/items/{filename}/accept-image", editor.handleAcceptImage).Methods("POST")

	// Static files
	r.PathPrefix("/www/").Handler(http.StripPrefix("/www/", http.FileServer(http.Dir("./www/"))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/", editor.handleIndex).Methods("GET")

	fmt.Println("🔧 Item Editor starting on http://localhost:8080")
	fmt.Println("Opening browser...")

	// Open browser automatically
	go func() {
		url := "http://localhost:8080"
		var err error

		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", url).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		case "darwin":
			err = exec.Command("open", url).Start()
		}

		if err != nil {
			fmt.Printf("Please open your browser to: %s\n", url)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", r))
}

func (e *ItemEditor) loadItems() error {
	itemsDir := "docs/data/equipment/items"
	e.items = make(map[string]*Item)

	err := filepath.WalkDir(itemsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var item Item
			if err := json.Unmarshal(data, &item); err != nil {
				return fmt.Errorf("error parsing %s: %v", path, err)
			}

			filename := strings.TrimSuffix(filepath.Base(path), ".json")
			e.items[filename] = &item
		}

		return nil
	})

	return err
}

func (e *ItemEditor) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Item Editor - Terminal Theme</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            background-color: #121212;
            color: #ffffff;
            font-size: 14px;
            line-height: 1.4;
        }

        .container {
            display: flex;
            height: 100vh;
        }

        .sidebar {
            width: 350px;
            background-color: #1e1e1e;
            border-right: 1px solid #3d3d3d;
            padding: 20px;
            overflow-y: auto;
        }

        .main-content {
            flex: 1;
            padding: 20px;
            overflow-y: auto;
        }

        .title {
            color: #50fa7b;
            font-size: 18px;
            font-weight: bold;
            margin-bottom: 20px;
            padding: 10px;
            background-color: #2d2d2d;
            border-radius: 4px;
        }

        .search-box {
            width: 100%;
            padding: 10px;
            background-color: #2d2d2d;
            border: 1px solid #3d3d3d;
            color: #ffffff;
            border-radius: 4px;
            margin-bottom: 15px;
        }

        .search-box:focus {
            outline: none;
            border-color: #50fa7b;
        }

        .item-list {
            list-style: none;
        }

        .item-item {
            padding: 10px;
            margin-bottom: 5px;
            background-color: #2d2d2d;
            border-radius: 4px;
            cursor: pointer;
            transition: background-color 0.2s;
        }

        .item-item:hover {
            background-color: #3d3d3d;
        }

        .item-item.selected {
            background-color: #44475a;
            border-left: 3px solid #50fa7b;
        }

        .item-name {
            color: #f8f8f2;
            font-weight: bold;
        }

        .item-id {
            color: #6272a4;
            font-size: 12px;
        }

        .item-status {
            float: right;
            font-size: 12px;
        }

        .status-valid {
            color: #50fa7b;
        }

        .status-invalid {
            color: #ff5555;
        }

        .form-group {
            margin-bottom: 15px;
        }

        .form-label {
            display: block;
            margin-bottom: 5px;
            color: #f1fa8c;
            font-weight: bold;
        }

        .form-input {
            width: 100%;
            padding: 10px;
            background-color: #2d2d2d;
            border: 1px solid #3d3d3d;
            color: #ffffff;
            border-radius: 4px;
        }

        .form-input:focus {
            outline: none;
            border-color: #50fa7b;
        }

        .form-textarea {
            height: 100px;
            resize: vertical;
        }

        .button {
            padding: 10px 20px;
            background-color: #50fa7b;
            color: #121212;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: bold;
            margin-right: 10px;
            margin-bottom: 10px;
        }

        .button:hover {
            background-color: #5af78e;
        }

        .button-secondary {
            background-color: #6272a4;
            color: #ffffff;
        }

        .button-secondary:hover {
            background-color: #7285b7;
        }

        .button-danger {
            background-color: #ff5555;
            color: #ffffff;
        }

        .button-danger:hover {
            background-color: #ff6b6b;
        }

        .status-bar {
            background-color: #1e1e1e;
            padding: 10px 20px;
            border-top: 1px solid #3d3d3d;
            position: fixed;
            bottom: 0;
            left: 0;
            right: 0;
        }

        .grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
        }

        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.8);
            z-index: 1000;
        }

        .modal-content {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background-color: #1e1e1e;
            padding: 30px;
            border-radius: 8px;
            border: 1px solid #3d3d3d;
            max-width: 600px;
            max-height: 80vh;
            overflow-y: auto;
            width: 90%;
        }

        .modal-title {
            color: #50fa7b;
            font-size: 18px;
            font-weight: bold;
            margin-bottom: 20px;
        }

        .reference-item {
            background-color: #2d2d2d;
            padding: 10px;
            margin-bottom: 5px;
            border-radius: 4px;
            font-size: 12px;
        }

        .reference-file {
            color: #8be9fd;
            font-weight: bold;
        }

        .reference-location {
            color: #6272a4;
        }

        .tag-chip, .note-chip {
            display: inline-block;
            background-color: #44475a;
            color: #f8f8f2;
            padding: 5px 10px;
            margin: 2px;
            border-radius: 15px;
            font-size: 12px;
            border: 1px solid #6272a4;
        }

        .tag-chip:hover, .note-chip:hover {
            background-color: #6272a4;
        }

        .chip-remove {
            margin-left: 8px;
            color: #ff5555;
            cursor: pointer;
            font-weight: bold;
        }

        .chip-remove:hover {
            color: #ff8888;
        }

        .chips-container {
            margin-bottom: 10px;
            min-height: 30px;
            border: 1px solid #3d3d3d;
            border-radius: 4px;
            padding: 5px;
            background-color: #2d2d2d;
        }

        .hidden {
            display: none !important;
        }

        .loading {
            text-align: center;
            color: #6272a4;
            padding: 20px;
        }

        @keyframes blink {
            0%, 50% { opacity: 1; }
            51%, 100% { opacity: 0; }
        }

        .cursor {
            animation: blink 1s infinite;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="sidebar">
            <div class="title">🔧 Item Editor</div>

            <input type="text" class="search-box" id="searchBox" placeholder="Search items...">

            <div class="form-group">
                <label class="form-label">Filter by Type:</label>
                <select class="form-input" id="typeFilter">
                    <option value="">All Types</option>
                </select>
            </div>

            <div class="form-group">
                <label class="form-label">Filter by Tag:</label>
                <select class="form-input" id="tagFilter">
                    <option value="">All Tags</option>
                </select>
            </div>

            <div class="form-group">
                <button class="button button-secondary" onclick="validateAllItems()">Validate All</button>
                <button class="button button-secondary" onclick="refreshItems()">Refresh</button>
            </div>

            <ul class="item-list" id="itemList">
                <li class="loading">Loading items<span class="cursor">...</span></li>
            </ul>
        </div>

        <div class="main-content">
            <div class="title" id="editTitle">Select an item to edit</div>

            <div id="editForm" class="hidden">
                <!-- Image Preview and Generation Section -->
                <div class="form-group" style="border: 1px solid #3d3d3d; border-radius: 4px; padding: 15px; margin-bottom: 20px; background-color: #1e1e1e;">
                    <label class="form-label">Item Image:</label>

                    <div style="display: flex; gap: 20px; align-items: flex-start;">
                        <!-- Current Image Preview -->
                        <div id="imagePreviewContainer" style="flex: 0 0 128px;">
                            <div style="width: 128px; height: 128px; background-color: #2d2d2d; border: 1px solid #3d3d3d; border-radius: 4px; display: flex; align-items: center; justify-content: center; image-rendering: pixelated;">
                                <img id="currentImagePreview" style="max-width: 128px; max-height: 128px; image-rendering: pixelated; display: none;" />
                                <div id="noImagePlaceholder" style="color: #6272a4; text-align: center; font-size: 12px;">No image</div>
                            </div>
                        </div>

                        <!-- Image Generation Controls -->
                        <div style="flex: 1;">
                            <div id="pixellabStatus" style="margin-bottom: 10px; color: #6272a4; font-size: 12px;">
                                <span id="balanceDisplay"></span>
                            </div>

                            <div style="display: flex; gap: 10px; margin-bottom: 10px;">
                                <select id="modelSelect" class="form-input" style="width: 150px;">
                                    <option value="bitforge">Bitforge (~$0.03)</option>
                                    <option value="pixflux">Pixflux (~$0.05)</option>
                                </select>
                                <button class="button" onclick="generateImage()" id="generateImageBtn">
                                    🎨 Generate Image
                                </button>
                                <button class="button button-secondary" onclick="refreshBalance()">
                                    🔄 Refresh Balance
                                </button>
                            </div>

                            <div id="generationStatus" style="display: none; padding: 10px; background-color: #2d2d2d; border-radius: 4px; margin-bottom: 10px;">
                                <div id="generationMessage" style="color: #f1fa8c;"></div>
                                <div id="generationCost" style="color: #50fa7b; font-size: 12px; margin-top: 5px;"></div>
                            </div>

                            <!-- Generated Image Preview -->
                            <div id="generatedImageContainer" style="display: none;">
                                <div style="margin-bottom: 10px;">
                                    <label class="form-label">Generated Image:</label>
                                    <div style="width: 128px; height: 128px; background-color: #2d2d2d; border: 1px solid #50fa7b; border-radius: 4px; display: flex; align-items: center; justify-content: center; image-rendering: pixelated;">
                                        <img id="generatedImagePreview" style="max-width: 128px; max-height: 128px; image-rendering: pixelated;" />
                                    </div>
                                </div>
                                <div style="display: flex; gap: 10px;">
                                    <button class="button" onclick="acceptGeneratedImage()">
                                        ✓ Use This Image
                                    </button>
                                    <button class="button button-danger" onclick="discardGeneratedImage()">
                                        ✗ Discard
                                    </button>
                                </div>
                                <div id="promptUsed" style="margin-top: 10px; padding: 8px; background-color: #2d2d2d; border-radius: 4px; font-size: 11px; color: #6272a4;">
                                    <strong>Prompt:</strong> <span id="promptText"></span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div id="dynamicFields"></div>

                <div class="form-group">
                    <label class="form-label">Tags:</label>
                    <div id="tagsContainer">
                        <div id="tagsList" class="chips-container"></div>
                        <input type="text" class="form-input" id="newTagInput" placeholder="Add tag..." list="tagDatalist" style="margin-top: 10px;" onkeypress="if(event.key==='Enter'){addTag(); event.preventDefault();}">
                        <datalist id="tagDatalist"></datalist>
                        <button type="button" class="button button-secondary" onclick="addTag()" style="margin-top: 5px;">+ Add Tag</button>
                    </div>
                </div>

                <div class="form-group">
                    <label class="form-label">Internal Notes:</label>
                    <div id="notesContainer">
                        <div id="notesList" class="chips-container"></div>
                        <input type="text" class="form-input" id="newNoteInput" placeholder="Add internal note..." style="margin-top: 10px;" onkeypress="if(event.key==='Enter'){addNote(); event.preventDefault();}">
                        <button type="button" class="button button-secondary" onclick="addNote()" style="margin-top: 5px;">+ Add Note</button>
                    </div>
                </div>

                <div class="form-group">
                    <button class="button button-secondary" onclick="showAddFieldDialog()">+ Add Field</button>
                </div>

                <div class="form-group">
                    <button class="button" onclick="saveItem()">💾 Save Changes</button>
                    <button class="button button-secondary" onclick="showRefactorDialog()">🔄 Refactor ID</button>
                </div>
            </div>
        </div>
    </div>

    <div class="status-bar">
        <span id="statusText">Ready</span>
    </div>

    <!-- Refactor Modal -->
    <div class="modal" id="refactorModal">
        <div class="modal-content">
            <div class="modal-title">🔄 Global ID Refactoring</div>

            <div class="form-group">
                <label class="form-label">Current ID:</label>
                <input type="text" class="form-input" id="currentId" readonly>
            </div>

            <div class="form-group">
                <label class="form-label">New ID:</label>
                <input type="text" class="form-input" id="newId">
            </div>

            <div class="form-group">
                <p>⚠️ This will update ALL references across:</p>
                <ul style="margin-left: 20px; margin-top: 10px; color: #6272a4;">
                    <li>Item filename</li>
                    <li>Starting gear entries</li>
                    <li>Pack contents</li>
                    <li>Any other item references</li>
                </ul>
            </div>

            <div class="form-group">
                <button class="button" onclick="previewRefactor()">Preview Changes</button>
                <button class="button button-secondary" onclick="closeRefactorModal()">Cancel</button>
            </div>
        </div>
    </div>

    <!-- Preview Modal -->
    <div class="modal" id="previewModal">
        <div class="modal-content">
            <div class="modal-title">🔄 Refactor Preview</div>

            <div id="previewContent">
                <div class="loading">Generating preview<span class="cursor">...</span></div>
            </div>

            <div class="form-group">
                <button class="button" onclick="applyRefactor()">✓ Apply Refactor</button>
                <button class="button button-danger" onclick="closePreviewModal()">✗ Cancel</button>
            </div>
        </div>
    </div>

    <!-- Add Field Modal -->
    <div class="modal" id="addFieldModal">
        <div class="modal-content">
            <div class="modal-title">➕ Add New Field</div>

            <div class="form-group">
                <label class="form-label">Field Name:</label>
                <input type="text" class="form-input" id="newFieldName" placeholder="e.g., damage-type, range">
            </div>

            <div class="form-group">
                <label class="form-label">Field Value:</label>
                <input type="text" class="form-input" id="newFieldValue" placeholder="Field value">
            </div>

            <div class="form-group">
                <button class="button" onclick="addField()">✓ Add Field</button>
                <button class="button button-danger" onclick="closeAddFieldModal()">✗ Cancel</button>
            </div>
        </div>
    </div>

    <script>
        let currentItems = {};
        let selectedFilename = null;
        let refactorPreviewData = null;
        let allTypes = [];
        let allTags = [];
        let currentTags = [];
        let currentNotes = [];
        let currentItemData = {};

        // Standard field order - these fields appear first in this order
        const STANDARD_FIELDS = [
            'id', 'name', 'description', 'ai_description', 'price', 'type', 'weight', 'stack',
            'gear_slot', 'rarity', 'ac', 'damage', 'damage-type', 'heal',
            'ammunition', 'range', 'range-long', 'img', 'image'
        ];

        // Fields that should always exist (will be added if missing)
        const REQUIRED_FIELDS = {
            'id': '',
            'name': '',
            'description': '',
            'price': 0,
            'type': '',
            'weight': 0,
            'stack': 1,
            'rarity': 'common',
            'tags': [],
            'notes': []
        };

        // Load items on page load
        document.addEventListener('DOMContentLoaded', function() {
            loadItems();
            loadTypes();
            loadTags();

            // Search and filter functionality
            document.getElementById('searchBox').addEventListener('input', filterItems);
            document.getElementById('typeFilter').addEventListener('change', filterItems);
            document.getElementById('tagFilter').addEventListener('change', filterItems);
        });

        async function loadItems() {
            try {
                const response = await fetch('/api/items');
                currentItems = await response.json();
                renderItemList();
                updateStatus('Loaded ' + Object.keys(currentItems).length + ' items');
            } catch (error) {
                updateStatus('Error loading items: ' + error.message);
            }
        }

        async function loadTypes() {
            try {
                const response = await fetch('/api/types');
                allTypes = await response.json();
                const typeFilter = document.getElementById('typeFilter');
                typeFilter.innerHTML = '<option value="">All Types</option>';
                allTypes.forEach(type => {
                    const option = document.createElement('option');
                    option.value = type;
                    option.textContent = type;
                    typeFilter.appendChild(option);
                });
            } catch (error) {
                console.error('Error loading types:', error);
            }
        }

        async function loadTags() {
            try {
                const response = await fetch('/api/tags');
                allTags = await response.json();

                // Update filter dropdown
                const tagFilter = document.getElementById('tagFilter');
                tagFilter.innerHTML = '<option value="">All Tags</option>';
                allTags.forEach(tag => {
                    const option = document.createElement('option');
                    option.value = tag;
                    option.textContent = tag;
                    tagFilter.appendChild(option);
                });

                // Update datalist for tag input
                const tagDatalist = document.getElementById('tagDatalist');
                if (tagDatalist) {
                    tagDatalist.innerHTML = '';
                    allTags.forEach(tag => {
                        const option = document.createElement('option');
                        option.value = tag;
                        tagDatalist.appendChild(option);
                    });
                }
            } catch (error) {
                console.error('Error loading tags:', error);
            }
        }

        function renderItemList() {
            filterItems(); // Use the filter function to render the list
        }

        function filterItems() {
            const searchQuery = document.getElementById('searchBox').value.toLowerCase();
            const typeFilter = document.getElementById('typeFilter').value;
            const tagFilter = document.getElementById('tagFilter').value;

            const itemList = document.getElementById('itemList');
            itemList.innerHTML = '';

            Object.keys(currentItems).forEach(filename => {
                const item = currentItems[filename];

                // Check search query
                const matchesSearch = !searchQuery ||
                    item.name.toLowerCase().includes(searchQuery) ||
                    item.id.toLowerCase().includes(searchQuery) ||
                    (item.description && item.description.toLowerCase().includes(searchQuery));

                // Check type filter
                const matchesType = !typeFilter || item.type === typeFilter;

                // Check tag filter
                const matchesTag = !tagFilter || (item.tags && item.tags.includes(tagFilter));

                if (matchesSearch && matchesType && matchesTag) {
                    const li = document.createElement('li');
                    li.className = 'item-item';
                    li.onclick = () => selectItem(filename);

                    if (selectedFilename === filename) {
                        li.classList.add('selected');
                    }

                    const validationStatus = item.id === filename ? '✓' : '✗';
                    li.innerHTML =
                        '<div class="item-name">' + item.name + '</div>' +
                        '<div class="item-id">' + item.id + '</div>' +
                        '<div class="item-status">' + validationStatus + '</div>';

                    itemList.appendChild(li);
                }
            });
        }

        async function selectItem(filename) {
            // Remove previous selection
            document.querySelectorAll('.item-item').forEach(item => {
                item.classList.remove('selected');
            });

            // Add selection to clicked item
            event.target.closest('.item-item').classList.add('selected');

            selectedFilename = filename;
            currentItemData = currentItems[filename];

            // Update title
            document.getElementById('editTitle').textContent = 'Editing: ' + currentItemData.name;

            // Ensure all required fields exist
            Object.keys(REQUIRED_FIELDS).forEach(field => {
                if (currentItemData[field] === undefined || currentItemData[field] === null) {
                    currentItemData[field] = REQUIRED_FIELDS[field];
                }
            });

            // Handle tags and notes separately
            currentTags = currentItemData.tags ? [...currentItemData.tags] : [];
            currentNotes = currentItemData.notes ? [...currentItemData.notes] : [];

            // Render dynamic fields
            renderDynamicFields();
            renderTags();
            renderNotes();

            // Load current image
            loadCurrentImage();

            // Show edit form
            document.getElementById('editForm').classList.remove('hidden');

            updateStatus('Selected: ' + currentItemData.name);
        }

        function renderDynamicFields() {
            console.log('renderDynamicFields called');
            const container = document.getElementById('dynamicFields');

            if (!container) {
                console.error('dynamicFields container not found!');
                return;
            }

            console.log('Container found, clearing...');
            container.innerHTML = '';

            if (!currentItemData || Object.keys(currentItemData).length === 0) {
                console.log('No currentItemData to render');
                container.innerHTML = '<div class="form-group"><label>No item data available</label></div>';
                return;
            }

            console.log('Rendering fields for:', currentItemData);

            // Skip tags and notes - they're handled separately
            const skipFields = ['tags', 'notes'];

            // Create ordered list: standard fields first, then custom fields
            const standardFieldsToShow = STANDARD_FIELDS.filter(field =>
                currentItemData.hasOwnProperty(field) && !skipFields.includes(field)
            );

            const customFields = Object.keys(currentItemData).filter(field =>
                !STANDARD_FIELDS.includes(field) && !skipFields.includes(field)
            );

            const orderedFields = [...standardFieldsToShow, ...customFields];

            orderedFields.forEach(key => {

                console.log('Creating field for:', key, currentItemData[key]);

                try {
                    const fieldGroup = document.createElement('div');
                    fieldGroup.className = 'form-group';

                const label = document.createElement('label');
                label.className = 'form-label';
                label.textContent = formatFieldName(key) + ':';

                const inputContainer = document.createElement('div');
                inputContainer.style.display = 'flex';
                inputContainer.style.gap = '10px';

                let input;
                if (key === 'description') {
                    input = document.createElement('textarea');
                    input.className = 'form-input form-textarea';
                } else {
                    input = document.createElement('input');
                    input.className = 'form-input';
                    input.type = getInputType(key);
                }

                input.id = 'field_' + key;
                input.value = formatFieldValue(currentItemData[key]);

                const removeBtn = document.createElement('button');
                removeBtn.className = 'button button-danger';
                removeBtn.textContent = '✗';
                removeBtn.style.padding = '5px 10px';
                removeBtn.onclick = () => removeField(key);

                    inputContainer.appendChild(input);
                    inputContainer.appendChild(removeBtn);

                    fieldGroup.appendChild(label);
                    fieldGroup.appendChild(inputContainer);
                    container.appendChild(fieldGroup);

                    console.log('Successfully created field for:', key);
                } catch (error) {
                    console.error('Error creating field for', key, ':', error);
                }
            });

            console.log('Finished rendering fields. Container children:', container.children.length);
        }

        function formatFieldName(key) {
            return key.replace(/-/g, ' ').replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
        }

        function getInputType(key) {
            if (key.includes('price') || key === 'stack') return 'number';
            if (key === 'weight') return 'number';
            return 'text';
        }

        function formatFieldValue(value) {
            if (value === null || value === undefined) return '';
            if (typeof value === 'object') return JSON.stringify(value);
            return String(value);
        }

        function removeField(key) {
            if (confirm('Remove field "' + key + '"?')) {
                delete currentItemData[key];
                renderDynamicFields();
            }
        }

        function renderTags() {
            const tagsList = document.getElementById('tagsList');
            if (!tagsList) {
                console.error('tagsList element not found');
                return;
            }
            tagsList.innerHTML = '';

            currentTags.forEach(tag => {
                const chip = document.createElement('span');
                chip.className = 'tag-chip';
                chip.innerHTML = tag + '<span class="chip-remove" onclick="removeTag(\'' + tag + '\')">&times;</span>';
                tagsList.appendChild(chip);
            });
        }

        function addTag() {
            const input = document.getElementById('newTagInput');
            const tag = input.value.trim();

            if (tag && !currentTags.includes(tag)) {
                currentTags.push(tag);
                renderTags();
                input.value = '';

                // Add to global tags list if new
                if (!allTags.includes(tag)) {
                    allTags.push(tag);
                    loadTags(); // Refresh the dropdown
                }
            }
        }

        function removeTag(tag) {
            currentTags = currentTags.filter(t => t !== tag);
            renderTags();
        }

        function renderNotes() {
            const notesList = document.getElementById('notesList');
            if (!notesList) {
                console.error('notesList element not found');
                return;
            }
            notesList.innerHTML = '';

            currentNotes.forEach(note => {
                const chip = document.createElement('span');
                chip.className = 'note-chip';
                chip.innerHTML = note + '<span class="chip-remove" onclick="removeNote(\'' + note + '\')">&times;</span>';
                notesList.appendChild(chip);
            });
        }

        function addNote() {
            const input = document.getElementById('newNoteInput');
            const note = input.value.trim();

            if (note && !currentNotes.includes(note)) {
                currentNotes.push(note);
                renderNotes();
                input.value = '';
            }
        }

        function removeNote(note) {
            currentNotes = currentNotes.filter(n => n !== note);
            renderNotes();
        }

        function showAddFieldDialog() {
            document.getElementById('newFieldName').value = '';
            document.getElementById('newFieldValue').value = '';
            document.getElementById('addFieldModal').style.display = 'block';
        }

        function closeAddFieldModal() {
            document.getElementById('addFieldModal').style.display = 'none';
        }

        function addField() {
            const fieldName = document.getElementById('newFieldName').value.trim();
            const fieldValue = document.getElementById('newFieldValue').value.trim();

            if (!fieldName) {
                alert('Field name is required');
                return;
            }

            if (currentItemData.hasOwnProperty(fieldName)) {
                alert('Field "' + fieldName + '" already exists');
                return;
            }

            // Try to parse as number if it looks like one
            let parsedValue = fieldValue;
            if (fieldValue && !isNaN(fieldValue) && !isNaN(parseFloat(fieldValue))) {
                parsedValue = parseFloat(fieldValue);
                if (Number.isInteger(parsedValue)) {
                    parsedValue = parseInt(fieldValue);
                }
            }

            currentItemData[fieldName] = parsedValue;
            renderDynamicFields();
            closeAddFieldModal();
        }

        async function saveItem() {
            if (!selectedFilename) {
                updateStatus('No item selected');
                return;
            }

            // Update currentItemData with form values
            Object.keys(currentItemData).forEach(key => {
                if (key === 'tags' || key === 'notes') return; // Skip - handled separately

                const input = document.getElementById('field_' + key);
                if (input) {
                    let value = input.value;

                    // Convert to appropriate type
                    if (input.type === 'number') {
                        value = input.step && input.step.includes('.') ? parseFloat(value) : parseInt(value);
                        if (isNaN(value)) value = 0;
                    } else if (value === '') {
                        value = null;
                    }

                    currentItemData[key] = value;
                }
            });

            // Add tags and notes (always include these fields)
            currentItemData.tags = currentTags;
            currentItemData.notes = currentNotes;

            // Create properly ordered object for saving
            const orderedItem = {};

            // Add standard fields first (in order)
            STANDARD_FIELDS.forEach(field => {
                if (currentItemData.hasOwnProperty(field)) {
                    orderedItem[field] = currentItemData[field];
                }
            });

            // Add custom fields
            Object.keys(currentItemData).forEach(field => {
                if (!STANDARD_FIELDS.includes(field) && field !== 'tags' && field !== 'notes') {
                    orderedItem[field] = currentItemData[field];
                }
            });

            // Add tags and notes at the end
            orderedItem.tags = currentTags;
            orderedItem.notes = currentNotes;

            try {
                const response = await fetch('/api/items/' + selectedFilename, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(orderedItem)
                });

                if (response.ok) {
                    currentItems[selectedFilename] = currentItemData;
                    renderItemList();
                    updateStatus('✓ Item saved successfully');

                    // Refresh types and tags lists in case new ones were added
                    loadTypes();
                    loadTags();
                } else {
                    const error = await response.text();
                    updateStatus('Error saving item: ' + error);
                }
            } catch (error) {
                updateStatus('Error saving item: ' + error.message);
            }
        }

        async function validateAllItems() {
            try {
                const response = await fetch('/api/validate');
                const result = await response.json();

                if (result.issues && result.issues.length > 0) {
                    updateStatus('Found ' + result.issues.length + ' validation issues: ' + result.issues.join(', '));
                } else {
                    updateStatus('✓ All items pass validation!');
                }
            } catch (error) {
                updateStatus('Error validating items: ' + error.message);
            }
        }

        async function refreshItems() {
            await loadItems();
        }

        function showRefactorDialog() {
            if (!selectedFilename) {
                updateStatus('No item selected');
                return;
            }

            document.getElementById('currentId').value = currentItems[selectedFilename].id;
            document.getElementById('newId').value = currentItems[selectedFilename].id;
            document.getElementById('refactorModal').style.display = 'block';
        }

        function closeRefactorModal() {
            document.getElementById('refactorModal').style.display = 'none';
        }

        async function previewRefactor() {
            const oldId = document.getElementById('currentId').value;
            const newId = document.getElementById('newId').value;

            if (!newId || newId === oldId) {
                updateStatus('Invalid or unchanged ID');
                return;
            }

            try {
                const response = await fetch('/api/refactor/preview', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        filename: selectedFilename,
                        oldId: oldId,
                        newId: newId
                    })
                });

                if (response.ok) {
                    refactorPreviewData = await response.json();
                    renderPreview(refactorPreviewData);

                    // Close refactor modal and show preview
                    closeRefactorModal();
                    document.getElementById('previewModal').style.display = 'block';
                } else {
                    const error = await response.text();
                    updateStatus('Error generating preview: ' + error);
                }
            } catch (error) {
                updateStatus('Error generating preview: ' + error.message);
            }
        }

        function renderPreview(preview) {
            const content = document.getElementById('previewContent');

            let html =
                '<div style="margin-bottom: 20px;">' +
                    '<div style="color: #50fa7b; font-weight: bold; font-size: 16px;">' +
                        '🔄 ' + preview.oldId + ' → ' + preview.newId +
                    '</div>' +
                '</div>' +

                '<div style="margin-bottom: 20px;">' +
                    '<div style="color: #f1fa8c; font-weight: bold;">📁 File rename:</div>' +
                    '<div style="color: #6272a4; margin-left: 10px;">' + preview.willRename + '</div>' +
                '</div>';

            if (preview.references && preview.references.length > 0) {
                html +=
                    '<div style="margin-bottom: 20px;">' +
                        '<div style="color: #f1fa8c; font-weight: bold;">📝 Will update ' + preview.references.length + ' references:</div>' +
                        '<div style="margin-top: 10px;">';

                preview.references.forEach(ref => {
                    html +=
                        '<div class="reference-item">' +
                            '<div class="reference-file">• ' + ref.file + '</div>' +
                            '<div class="reference-location">  ' + ref.location + '</div>' +
                        '</div>';
                });

                html += '</div></div>';
            } else {
                html +=
                    '<div style="margin-bottom: 20px;">' +
                        '<div style="color: #f1fa8c; font-weight: bold;">📝 No references found to update.</div>' +
                    '</div>';
            }

            html +=
                '<div style="margin-top: 20px; padding: 15px; background-color: #2d2d2d; border-radius: 4px;">' +
                    '<span style="color: #50fa7b;">✓ Safe to apply</span> or ' +
                    '<span style="color: #ff5555;">✗ Cancel</span>' +
                '</div>';

            content.innerHTML = html;
        }

        function closePreviewModal() {
            document.getElementById('previewModal').style.display = 'none';
            refactorPreviewData = null;
        }

        async function applyRefactor() {
            if (!refactorPreviewData) {
                updateStatus('No refactor data available');
                return;
            }

            try {
                const response = await fetch('/api/refactor/apply', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        filename: selectedFilename,
                        preview: refactorPreviewData
                    })
                });

                if (response.ok) {
                    updateStatus('✓ Refactor applied successfully!');
                    closePreviewModal();

                    // Reload items and select the renamed item
                    await loadItems();
                    if (refactorPreviewData && refactorPreviewData.newId && refactorPreviewData.newId in currentItems) {
                        selectItem(refactorPreviewData.newId);
                    }
                } else {
                    const error = await response.text();
                    updateStatus('Error applying refactor: ' + error);
                }
            } catch (error) {
                updateStatus('Error applying refactor: ' + error.message);
            }
        }

        function updateStatus(message) {
            document.getElementById('statusText').textContent = message;
        }

        // Image generation functions
        let generatedImageData = null;

        async function refreshBalance() {
            try {
                const response = await fetch('/api/balance');
                if (response.ok) {
                    const balance = await response.json();
                    document.getElementById('balanceDisplay').textContent =
                        'Balance: $' + balance.usd.toFixed(4) + ' USD';
                    updateStatus('Balance refreshed');
                } else {
                    document.getElementById('balanceDisplay').textContent =
                        'PixelLab API not configured';
                }
            } catch (error) {
                document.getElementById('balanceDisplay').textContent =
                    'PixelLab API unavailable';
                console.error('Error fetching balance:', error);
            }
        }

        async function loadCurrentImage() {
            if (!selectedFilename) return;

            const response = await fetch('/api/items/' + selectedFilename + '/image');
            const data = await response.json();

            const preview = document.getElementById('currentImagePreview');
            const placeholder = document.getElementById('noImagePlaceholder');

            if (data.exists) {
                preview.src = '/' + data.path.replace(/\\/g, '/');
                preview.style.display = 'block';
                placeholder.style.display = 'none';
            } else {
                preview.style.display = 'none';
                placeholder.style.display = 'block';
            }
        }

        async function generateImage() {
            if (!selectedFilename) {
                updateStatus('No item selected');
                return;
            }

            const model = document.getElementById('modelSelect').value;
            const btn = document.getElementById('generateImageBtn');
            const statusDiv = document.getElementById('generationStatus');
            const messageDiv = document.getElementById('generationMessage');
            const costDiv = document.getElementById('generationCost');

            btn.disabled = true;
            btn.textContent = '⏳ Generating...';
            statusDiv.style.display = 'block';
            messageDiv.textContent = 'Generating pixel art image...';
            costDiv.textContent = '';

            try {
                const response = await fetch('/api/items/' + selectedFilename + '/generate-image', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ model: model })
                });

                if (response.ok) {
                    const result = await response.json();
                    generatedImageData = result;

                    messageDiv.textContent = '✓ Image generated successfully!';
                    costDiv.textContent = 'Cost: $' + result.cost.toFixed(4) + ' USD';

                    // Show the generated image
                    const generatedPreview = document.getElementById('generatedImagePreview');
                    generatedPreview.src = 'data:image/png;base64,' + result.imageData;
                    document.getElementById('generatedImageContainer').style.display = 'block';
                    document.getElementById('promptText').textContent = result.prompt;

                    updateStatus('Image generated successfully');
                    refreshBalance(); // Update balance
                } else {
                    const error = await response.text();
                    messageDiv.textContent = '✗ Error: ' + error;
                    messageDiv.style.color = '#ff5555';
                    updateStatus('Error generating image');
                }
            } catch (error) {
                messageDiv.textContent = '✗ Error: ' + error.message;
                messageDiv.style.color = '#ff5555';
                updateStatus('Error generating image');
            } finally {
                btn.disabled = false;
                btn.textContent = '🎨 Generate Image';
            }
        }

        async function acceptGeneratedImage() {
            if (!generatedImageData || !selectedFilename) return;

            try {
                // Call backend to copy the image to the main directory
                const response = await fetch('/api/items/' + selectedFilename + '/accept-image', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        imageData: generatedImageData.imageData
                    })
                });

                if (!response.ok) {
                    throw new Error('Failed to save image');
                }

                const result = await response.json();

                // Update the item's image field
                const itemId = currentItemData.id;
                const imagePathField = document.getElementById('field_image');
                if (imagePathField) {
                    imagePathField.value = '/res/img/items/' + itemId + '.png';
                }

                // Hide the generated image container
                document.getElementById('generatedImageContainer').style.display = 'none';
                document.getElementById('generationStatus').style.display = 'none';

                // Update the current image preview
                const preview = document.getElementById('currentImagePreview');
                const placeholder = document.getElementById('noImagePlaceholder');
                preview.src = 'data:image/png;base64,' + generatedImageData.imageData;
                preview.style.display = 'block';
                placeholder.style.display = 'none';

                updateStatus('Image accepted and saved to ' + result.path);
                generatedImageData = null;
            } catch (error) {
                updateStatus('Error accepting image: ' + error.message);
            }
        }

        function discardGeneratedImage() {
            document.getElementById('generatedImageContainer').style.display = 'none';
            document.getElementById('generationStatus').style.display = 'none';
            generatedImageData = null;
            updateStatus('Generated image discarded');
        }

        // Load balance on page load
        document.addEventListener('DOMContentLoaded', function() {
            refreshBalance();
        });

        // Close modals when clicking outside
        window.onclick = function(event) {
            const refactorModal = document.getElementById('refactorModal');
            const previewModal = document.getElementById('previewModal');
            const addFieldModal = document.getElementById('addFieldModal');

            if (event.target === refactorModal) {
                closeRefactorModal();
            }
            if (event.target === previewModal) {
                closePreviewModal();
            }
            if (event.target === addFieldModal) {
                closeAddFieldModal();
            }
        }
    </script>
</body>
</html>`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
}

func (e *ItemEditor) handleGetItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e.items)
}

func (e *ItemEditor) handleGetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.items[filename]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (e *ItemEditor) handleSaveItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save to file
	if err := e.saveItemToFile(filename, &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update in memory
	e.items[filename] = &item

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (e *ItemEditor) handleValidate(w http.ResponseWriter, r *http.Request) {
	issues := []string{}

	for filename, item := range e.items {
		if item.ID != filename {
			issues = append(issues, fmt.Sprintf("%s: ID '%s' doesn't match filename", filename, item.ID))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issues": issues,
	})
}

func (e *ItemEditor) handleGetTypes(w http.ResponseWriter, r *http.Request) {
	types := make(map[string]bool)
	for _, item := range e.items {
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

func (e *ItemEditor) handleGetTags(w http.ResponseWriter, r *http.Request) {
	tags := make(map[string]bool)
	for _, item := range e.items {
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

func (e *ItemEditor) handleRefactorPreview(w http.ResponseWriter, r *http.Request) {
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

func (e *ItemEditor) handleRefactorApply(w http.ResponseWriter, r *http.Request) {
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
	e.loadItems()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (e *ItemEditor) saveItemToFile(filename string, item *Item) error {
	itemsDir := "docs/data/equipment/items"
	filepath := filepath.Join(itemsDir, filename+".json")

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func (e *ItemEditor) generateRefactorPreview(oldID, newID, filename string) (*RefactorPreview, error) {
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

func (e *ItemEditor) scanStartingGearReferences(oldID, newID string) ([]Reference, error) {
	startingGearPath := "docs/data/character/starting-gear.json"

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

func (e *ItemEditor) scanGivenReferences(given interface{}, oldID, newID, basePath string) []Reference {
	var references []Reference

	switch g := given.(type) {
	case []interface{}:
		// Handle both formats: [["item", 1], ["item", 2]] and ["item", 10]
		for i, item := range g {
			if subArray, ok := item.([]interface{}); ok && len(subArray) >= 2 {
				// Standard format: ["item", quantity]
				if itemID, ok := subArray[0].(string); ok && itemID == oldID {
					references = append(references, Reference{
						File:     "starting-gear.json",
						Location: fmt.Sprintf("%s[%d][0]", basePath, i),
						OldID:    oldID,
						NewID:    newID,
					})
				}
			} else if i == 0 {
				// Special format: ["item", quantity] (direct array)
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

func (e *ItemEditor) scanOptionReferences(option interface{}, oldID, newID, basePath string) []Reference {
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

func (e *ItemEditor) scanPackReferences(oldID, newID string) ([]Reference, error) {
	var references []Reference

	for filename, item := range e.items {
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

func (e *ItemEditor) applyRefactor(preview *RefactorPreview, filename string) error {
	// 1. Update the item's ID
	item := e.items[filename]
	item.ID = preview.NewID

	// 2. Save the item with new filename
	newFilename := preview.NewID
	if err := e.saveItemToFile(newFilename, item); err != nil {
		return fmt.Errorf("failed to save item with new filename: %v", err)
	}

	// 3. Delete old file
	oldFilePath := filepath.Join("docs/data/equipment/items", filename+".json")
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
	delete(e.items, filename)
	e.items[newFilename] = item

	return nil
}

func (e *ItemEditor) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	if e.pixellabClient == nil {
		http.Error(w, "PixelLab client not initialized", http.StatusServiceUnavailable)
		return
	}

	balance, err := e.pixellabClient.GetBalance()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (e *ItemEditor) handleGenerateImage(w http.ResponseWriter, r *http.Request) {
	if e.pixellabClient == nil {
		http.Error(w, "PixelLab client not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.items[filename]
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

	prompt := generateFantasyPrompt(item)
	negativePrompt := generateNegativePrompt()

	result, err := e.pixellabClient.GenerateImage(prompt, negativePrompt, req.Model)
	if err != nil {
		log.Printf("❌ Error generating image: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save image with timestamp to history folder
	timestamp := time.Now().Format("20060102_150405")
	historyDir := filepath.Join("www/res/img/items/_history", item.ID)
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

func (e *ItemEditor) handleGetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.items[filename]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Check if image exists
	imagePath := filepath.Join("www/res/img/items", item.ID+".png")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"exists": false,
		})
		return
	}

	// Also check for images in history
	historyDir := filepath.Join("www/res/img/items/_history", item.ID)
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

func (e *ItemEditor) handleAcceptImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	item, exists := e.items[filename]
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
	mainImagePath := filepath.Join("www/res/img/items", item.ID+".png")
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

func (e *ItemEditor) updateStartingGearReferences(preview *RefactorPreview) error {
	startingGearPath := "docs/data/character/starting-gear.json"

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

func (e *ItemEditor) updatePackReferences(preview *RefactorPreview) error {
	for filename, item := range e.items {
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
			if err := e.saveItemToFile(filename, item); err != nil {
				return fmt.Errorf("failed to update %s: %v", filename, err)
			}
		}
	}

	return nil
}