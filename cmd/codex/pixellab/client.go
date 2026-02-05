package pixellab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Response from PixelLab API
type Response struct {
	Usage struct {
		USD float64 `json:"usd"`
	} `json:"usage"`
	Image struct {
		Base64 string `json:"base64"`
	} `json:"image"`
}

// Balance response from PixelLab
type Balance struct {
	Type string  `json:"type"`
	USD  float64 `json:"usd"`
}

// Client for PixelLab image generation
type Client struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// NewClient creates a new PixelLab client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: "https://api.pixellab.ai/v1",
		Client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// GetBalance checks API balance
func (c *Client) GetBalance() (*Balance, error) {
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

	var balance Balance
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// GenerateImage generates a pixel art image
func (c *Client) GenerateImage(description, negativePrompt string, model string) (*Response, error) {
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

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GeneratePrompt creates a fantasy prompt from item properties
func GeneratePrompt(name, description, aiDescription, rarity string) string {
	// Priority 1: Use ai_description if available
	if aiDescription != "" && len(aiDescription) > 10 {
		return aiDescription
	}

	// Priority 2: Use regular description if available
	if description != "" && len(description) > 10 {
		return description
	}

	// Fallback: Generate from item name and rarity
	rarityLower := strings.ToLower(rarity)
	var rarityDesc string
	switch rarityLower {
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

	return fmt.Sprintf("%s %s", rarityDesc, name)
}

// NegativePrompt returns the standard negative prompt for pixel art
func NegativePrompt() string {
	return "blurry, fuzzy, soft, antialiased, smooth, low quality, modern, realistic, photograph, 3d render, low resolution, text, letters, words, people, characters, faces, anime, cartoon"
}
