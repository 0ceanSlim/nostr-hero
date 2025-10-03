package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Config struct {
	PixelLab struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"pixellab"`
}

type Item struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	AIDescription string   `json:"ai_description"`
	Type          string   `json:"type"`
	Rarity        string   `json:"rarity"`
	Tags          []string `json:"tags"`
}

type PixelLabResponse struct {
	Usage struct {
		USD float64 `json:"usd"`
	} `json:"usage"`
	Image struct {
		Base64 string `json:"base64"`
	} `json:"image"`
}

type BalanceResponse struct {
	Type string  `json:"type"`
	USD  float64 `json:"usd"`
}

type PixelLabClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func NewPixelLabClient(apiKey string) *PixelLabClient {
	return &PixelLabClient{
		APIKey:  apiKey,
		BaseURL: "https://api.pixellab.ai/v1",
		Client:  &http.Client{Timeout: 60 * time.Second},
	}
}

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

func loadConfig() (*Config, error) {
	configPath := "config.yml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
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
		return nil, fmt.Errorf("pixellab.api_key not found in config.yml")
	}

	return &config, nil
}

func loadItems() ([]Item, error) {
	itemsDir := "docs/data/equipment/items"
	items := []Item{}

	err := filepath.Walk(itemsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var item Item
		if err := json.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("error parsing %s: %v", path, err)
		}

		items = append(items, item)
		return nil
	})

	return items, err
}

func filterItemsByType(items []Item, itemType string) []Item {
	if itemType == "" {
		return items
	}

	var filtered []Item
	for _, item := range items {
		// Try exact match first (case insensitive)
		if strings.EqualFold(item.Type, itemType) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func findItemByID(items []Item, itemID string) *Item {
	for _, item := range items {
		if strings.EqualFold(item.ID, itemID) {
			return &item
		}
	}
	return nil
}

func generateFantasyPrompt(item Item) string {
	baseStyle := ""

	itemID := strings.ToLower(item.ID)
	itemName := strings.ToLower(item.Name)
	itemType := strings.ToLower(item.Type)
	rarity := strings.ToLower(item.Rarity)
	description := strings.ToLower(item.Description)

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

	var itemDesc string

	// Priority 1: Use ai_description if available
	if item.AIDescription != "" && len(item.AIDescription) > 10 {
		itemDesc = fmt.Sprintf("%s %s", rarityDesc, item.AIDescription)
	} else if description != "" && len(description) > 10 {
		// Priority 2: Use regular description if available
		itemDesc = fmt.Sprintf("%s %s", rarityDesc, description)
	} else {
		// Fallback to specific hardcoded descriptions for common items
		switch itemID {
		case "acid":
			itemDesc = "small glass vial filled with green corrosive acid, bubbling liquid, cork stopper"
		case "alchemists-fire":
			itemDesc = "glass flask with orange glowing liquid fire, cork stopper, magical flames"
		case "alchemists-supplies":
			itemDesc = "alchemist's kit with bottles, vials, mortar and pestle, brewing equipment"
		case "abacus":
			itemDesc = "wooden counting frame with sliding beads, calculation tool"
		case "arrows", "arrow":
			itemDesc = "bundle of wooden arrows with metal tips and feather fletching"
		case "longbow":
			itemDesc = "tall wooden longbow with bowstring, recurved ends"
		case "longsword":
			itemDesc = "medieval longsword with straight blade, crossguard, leather wrapped handle"
		case "dagger":
			itemDesc = "short dagger with pointed blade, simple crossguard, wrapped handle"
		case "battleaxe":
			itemDesc = "single-bladed battle axe with wooden handle, metal axe head"
		case "greatsword":
			itemDesc = "large two-handed sword with long blade, extended crossguard"
		case "hammer":
			itemDesc = "war hammer with heavy metal head, wooden handle"
		case "shield":
			itemDesc = "round wooden shield with metal boss, leather straps"
		case "healing":
			itemDesc = "red healing potion in glass bottle with cork, magical red liquid"
		case "gold-piece":
			itemDesc = "shiny gold coin with royal crest, metallic gleam"
		default:
			// Final fallback based on item type and name
			if strings.Contains(itemType, "weapon") || strings.Contains(itemName, "sword") || strings.Contains(itemName, "bow") || strings.Contains(itemName, "axe") || strings.Contains(itemName, "hammer") || strings.Contains(itemName, "dagger") {
				itemDesc = fmt.Sprintf("%s %s weapon with handle and blade", rarityDesc, itemName)
			} else if strings.Contains(itemType, "armor") || strings.Contains(itemName, "shield") || strings.Contains(itemName, "helm") {
				itemDesc = fmt.Sprintf("%s %s armor piece", rarityDesc, itemName)
			} else if strings.Contains(itemType, "tool") || strings.Contains(itemName, "supplies") {
				itemDesc = fmt.Sprintf("%s %s tool or equipment", rarityDesc, itemName)
			} else if strings.Contains(itemName, "potion") || strings.Contains(itemName, "elixir") {
				itemDesc = fmt.Sprintf("glass bottle with %s liquid, cork stopper", itemName)
			} else if strings.Contains(itemName, "scroll") {
				itemDesc = fmt.Sprintf("rolled parchment scroll with writing")
			} else if strings.Contains(itemName, "book") || strings.Contains(itemName, "tome") {
				itemDesc = fmt.Sprintf("leather-bound book with pages")
			} else if strings.Contains(itemName, "ring") {
				itemDesc = fmt.Sprintf("metal ring with gem or inscription")
			} else if strings.Contains(itemName, "amulet") || strings.Contains(itemName, "pendant") {
				itemDesc = fmt.Sprintf("pendant amulet on chain")
			} else {
				itemDesc = fmt.Sprintf("%s %s", rarityDesc, itemName)
			}
		}
	}

	if baseStyle == "" {
		return itemDesc
	}
	return fmt.Sprintf("%s, %s", baseStyle, itemDesc)
}

func generateNegativePrompt() string {
	return "blurry, fuzzy, soft, antialiased, smooth, low quality, modern, realistic, photograph, 3d render, low resolution, text, letters, words, people, characters, faces, anime, cartoon"
}

func saveImage(base64Data, filename, runDir string) error {
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	outputDir := filepath.Join("www/res/img/items", runDir, "png")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, filename+".png")
	return os.WriteFile(outputPath, imageData, 0644)
}

var rootCmd = &cobra.Command{
	Use:   "pixellab-generator",
	Short: "Generate pixel art images for game items using PixelLab API",
	Long:  "A tool to generate 32x32 pixel art images for fantasy game items using the PixelLab API",
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Check current API balance",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		client := NewPixelLabClient(config.PixelLab.APIKey)
		balance, err := client.GetBalance()
		if err != nil {
			fmt.Printf("Error getting balance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current balance: $%.4f USD\n", balance.USD)
	},
}

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Estimate costs for generating images",
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetInt("count")
		model, _ := cmd.Flags().GetString("model")

		items, err := loadItems()
		if err != nil {
			fmt.Printf("Error loading items: %v\n", err)
			os.Exit(1)
		}

		if count == 0 || count > len(items) {
			count = len(items)
		}

		// Rough estimate based on typical PixelLab pricing
		var estimatedCostPerImage float64 = 0.05 // $0.05 per image estimate
		if model == "bitforge" {
			estimatedCostPerImage = 0.03 // Bitforge might be cheaper
		}

		totalCost := float64(count) * estimatedCostPerImage

		fmt.Printf("Dry run estimate:\n")
		fmt.Printf("Items to generate: %d\n", count)
		fmt.Printf("Model: %s\n", model)
		fmt.Printf("Estimated cost: $%.4f USD\n", totalCost)
		fmt.Printf("\nNote: This is an estimate. Actual costs may vary.\n")
	},
}

var generateTypeCmd = &cobra.Command{
	Use:   "generate-type [type]",
	Short: "Generate images for items of a specific type",
	Long: `Generate images for items of a specific type.

Available types:
  "Adventuring Gear"      - Ropes, torches, tools, etc.
  "Martial Melee Weapons" - Longswords, battleaxes, etc.
  "Simple Melee Weapons"  - Clubs, daggers, etc.
  "Martial Ranged Weapons"- Longbows, crossbows, etc.
  "Simple Ranged Weapons" - Slings, darts, etc.
  "Light Armor"           - Leather, studded leather
  "Medium Armor"          - Chain mail, scale mail
  "Heavy Armor"           - Plate, splint armor
  "Ammunition"            - Arrows, bolts, bullets
  "Tools"                 - Artisan tools, thieves' tools
  "Clothes"               - Common, fine, traveler's clothes
  "Potion"                - Healing potions, elixirs

Examples:
  pixellab-gen generate-type "Martial Melee Weapons" --model bitforge
  pixellab-gen generate-type "Adventuring Gear" --count 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemType := args[0]
		count, _ := cmd.Flags().GetInt("count")
		maxBalance, _ := cmd.Flags().GetBool("max-balance")
		model, _ := cmd.Flags().GetString("model")

		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		client := NewPixelLabClient(config.PixelLab.APIKey)

		// Check balance first
		balance, err := client.GetBalance()
		if err != nil {
			fmt.Printf("Error getting balance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current balance: $%.4f USD\n", balance.USD)

		allItems, err := loadItems()
		if err != nil {
			fmt.Printf("Error loading items: %v\n", err)
			os.Exit(1)
		}

		// Filter items by type
		items := filterItemsByType(allItems, itemType)
		if len(items) == 0 {
			fmt.Printf("No items found for type '%s'\n", itemType)
			fmt.Println("Available types: \"Adventuring Gear\", \"Martial Melee Weapons\", \"Simple Melee Weapons\", etc.")
			os.Exit(1)
		}

		fmt.Printf("Found %d items of type '%s'\n", len(items), itemType)

		if maxBalance {
			// Estimate how many images we can generate with current balance
			estimatedCostPerImage := 0.05
			if model == "bitforge" {
				estimatedCostPerImage = 0.03
			}
			maxCount := int(balance.USD / estimatedCostPerImage)
			if maxCount > len(items) {
				maxCount = len(items)
			}
			count = maxCount
			fmt.Printf("Generating maximum possible items with current balance: %d\n", count)
		} else if count == 0 || count > len(items) {
			count = len(items)
		}

		if count == 0 {
			fmt.Println("No items to generate.")
			return
		}

		// Create run-specific directory with timestamp and type
		runDir := fmt.Sprintf("run_%s_%s_%s", strings.ReplaceAll(itemType, " ", "_"), model, time.Now().Format("20060102_150405"))
		fmt.Printf("Generating %d items of type '%s' using %s model...\n", count, itemType, model)
		fmt.Printf("Output directory: www/res/img/items/%s/png/\n\n", runDir)

		totalCost := 0.0
		successful := 0

		for i := 0; i < count; i++ {
			item := items[i]
			fmt.Printf("[%d/%d] Generating %s...", i+1, count, item.Name)

			prompt := generateFantasyPrompt(item)
			negativePrompt := generateNegativePrompt()

			// Debug: Show the actual prompt being sent
			fmt.Printf("\n  DEBUG: Prompt = %s\n", prompt)

			result, err := client.GenerateImage(prompt, negativePrompt, model)
			if err != nil {
				fmt.Printf(" ERROR: %v\n", err)
				continue
			}

			if err := saveImage(result.Image.Base64, item.ID, runDir); err != nil {
				fmt.Printf(" ERROR saving: %v\n", err)
				continue
			}

			totalCost += result.Usage.USD
			successful++
			fmt.Printf(" SUCCESS ($%.4f)\n", result.Usage.USD)

			// Small delay to be respectful to the API
			time.Sleep(1 * time.Second)
		}

		fmt.Printf("\n=== Generation Complete ===\n")
		fmt.Printf("Successfully generated: %d/%d images\n", successful, count)
		fmt.Printf("Total cost: $%.4f USD\n", totalCost)
		fmt.Printf("Remaining balance: $%.4f USD\n", balance.USD-totalCost)
		fmt.Printf("Images saved to: www/res/img/items/%s/png/\n", runDir)
	},
}

var generateIDCmd = &cobra.Command{
	Use:   "generate-id [item-id]",
	Short: "Generate image for a specific item by ID",
	Long: `Generate image for a specific item by its ID.

Examples:
  pixellab-gen generate-id longsword --model bitforge
  pixellab-gen generate-id acid --model pixflux`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemID := args[0]
		model, _ := cmd.Flags().GetString("model")

		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		client := NewPixelLabClient(config.PixelLab.APIKey)

		// Check balance first
		balance, err := client.GetBalance()
		if err != nil {
			fmt.Printf("Error getting balance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current balance: $%.4f USD\n", balance.USD)

		allItems, err := loadItems()
		if err != nil {
			fmt.Printf("Error loading items: %v\n", err)
			os.Exit(1)
		}

		// Find item by ID
		item := findItemByID(allItems, itemID)
		if item == nil {
			fmt.Printf("Item with ID '%s' not found\n", itemID)
			os.Exit(1)
		}

		fmt.Printf("Found item: %s (%s)\n", item.Name, item.Type)

		// Create run-specific directory with timestamp and item ID
		runDir := fmt.Sprintf("run_%s_%s_%s", itemID, model, time.Now().Format("20060102_150405"))
		fmt.Printf("Generating '%s' using %s model...\n", item.Name, model)
		fmt.Printf("Output directory: www/res/img/items/%s/png/\n\n", runDir)

		fmt.Printf("Generating %s...", item.Name)

		prompt := generateFantasyPrompt(*item)
		negativePrompt := generateNegativePrompt()

		// Debug: Show the actual prompt being sent
		fmt.Printf("\n  DEBUG: Prompt = %s\n", prompt)

		result, err := client.GenerateImage(prompt, negativePrompt, model)
		if err != nil {
			fmt.Printf(" ERROR: %v\n", err)
			os.Exit(1)
		}

		if err := saveImage(result.Image.Base64, item.ID, runDir); err != nil {
			fmt.Printf(" ERROR saving: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf(" SUCCESS ($%.4f)\n", result.Usage.USD)

		fmt.Printf("\n=== Generation Complete ===\n")
		fmt.Printf("Successfully generated: %s\n", item.Name)
		fmt.Printf("Total cost: $%.4f USD\n", result.Usage.USD)
		fmt.Printf("Remaining balance: $%.4f USD\n", balance.USD-result.Usage.USD)
		fmt.Printf("Image saved to: www/res/img/items/%s/png/%s.png\n", runDir, item.ID)
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate images for items",
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetInt("count")
		maxBalance, _ := cmd.Flags().GetBool("max-balance")
		model, _ := cmd.Flags().GetString("model")

		config, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		client := NewPixelLabClient(config.PixelLab.APIKey)

		// Check balance first
		balance, err := client.GetBalance()
		if err != nil {
			fmt.Printf("Error getting balance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current balance: $%.4f USD\n", balance.USD)

		items, err := loadItems()
		if err != nil {
			fmt.Printf("Error loading items: %v\n", err)
			os.Exit(1)
		}

		if maxBalance {
			// Estimate how many images we can generate with current balance
			estimatedCostPerImage := 0.05
			if model == "bitforge" {
				estimatedCostPerImage = 0.03
			}
			count = int(balance.USD / estimatedCostPerImage)
			if count > len(items) {
				count = len(items)
			}
			fmt.Printf("Generating maximum possible items with current balance: %d\n", count)
		} else if count == 0 || count > len(items) {
			count = len(items)
		}

		if count == 0 {
			fmt.Println("No items to generate.")
			return
		}

		// Create run-specific directory with timestamp
		runDir := fmt.Sprintf("run_%s_%s", model, time.Now().Format("20060102_150405"))
		fmt.Printf("Generating %d items using %s model...\n", count, model)
		fmt.Printf("Output directory: www/res/img/items/%s/png/\n\n", runDir)

		totalCost := 0.0
		successful := 0

		for i := 0; i < count; i++ {
			item := items[i]
			fmt.Printf("[%d/%d] Generating %s...", i+1, count, item.Name)

			prompt := generateFantasyPrompt(item)
			negativePrompt := generateNegativePrompt()

			// Debug: Show the actual prompt being sent
			fmt.Printf("\n  DEBUG: Prompt = %s\n", prompt)

			result, err := client.GenerateImage(prompt, negativePrompt, model)
			if err != nil {
				fmt.Printf(" ERROR: %v\n", err)
				continue
			}

			if err := saveImage(result.Image.Base64, item.ID, runDir); err != nil {
				fmt.Printf(" ERROR saving: %v\n", err)
				continue
			}

			totalCost += result.Usage.USD
			successful++
			fmt.Printf(" SUCCESS ($%.4f)\n", result.Usage.USD)

			// Small delay to be respectful to the API
			time.Sleep(1 * time.Second)
		}

		fmt.Printf("\n=== Generation Complete ===\n")
		fmt.Printf("Successfully generated: %d/%d images\n", successful, count)
		fmt.Printf("Total cost: $%.4f USD\n", totalCost)
		fmt.Printf("Remaining balance: $%.4f USD\n", balance.USD-totalCost)
		fmt.Printf("Images saved to: www/res/img/items/%s/png/\n", runDir)
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(dryRunCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(generateTypeCmd)
	rootCmd.AddCommand(generateIDCmd)

	// Flags for dry-run
	dryRunCmd.Flags().IntP("count", "c", 0, "Number of items to estimate (0 = all)")
	dryRunCmd.Flags().StringP("model", "m", "pixflux", "Model to use (pixflux or bitforge)")

	// Flags for generate
	generateCmd.Flags().IntP("count", "c", 0, "Number of items to generate (0 = all)")
	generateCmd.Flags().BoolP("max-balance", "b", false, "Generate maximum possible with current balance")
	generateCmd.Flags().StringP("model", "m", "pixflux", "Model to use (pixflux or bitforge)")

	// Flags for generate-type
	generateTypeCmd.Flags().IntP("count", "c", 0, "Number of items to generate (0 = all)")
	generateTypeCmd.Flags().BoolP("max-balance", "b", false, "Generate maximum possible with current balance")
	generateTypeCmd.Flags().StringP("model", "m", "pixflux", "Model to use (pixflux or bitforge)")

	// Flags for generate-id
	generateIDCmd.Flags().StringP("model", "m", "pixflux", "Model to use (pixflux or bitforge)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}