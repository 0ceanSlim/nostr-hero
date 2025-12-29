package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents application configuration
type Config struct {
	PixelLab struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"pixellab"`
}

// Load loads the application configuration from config.yml
func Load() (*Config, error) {
	// Look for config.yml in the project root (two directories up from CODEX)
	configPath := "../../config.yml"
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
