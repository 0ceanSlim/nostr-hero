package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents application configuration
type Config struct {
	Server struct {
		Port        int    `yaml:"port"`
		StagingMode string `yaml:"staging_mode"` // "auto", "direct", or "staging"
	} `yaml:"server"`
	GitHub struct {
		Token     string `yaml:"token"`
		RepoOwner string `yaml:"repo_owner"`
		RepoName  string `yaml:"repo_name"`
	} `yaml:"github"`
	PixelLab struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"pixellab"`
}

// Load loads the application configuration from codex-config.yml
func Load() (*Config, error) {
	// Look for codex-config.yml in the current directory (running from root)
	configPath := "./codex-config.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("⚠️ codex-config.yml not found - using defaults")
		// Return default config
		config := &Config{}
		config.Server.Port = 8080
		config.Server.StagingMode = "auto"
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set default port if not specified
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	// Set default staging mode if not specified
	if config.Server.StagingMode == "" {
		config.Server.StagingMode = "auto"
	}

	if config.PixelLab.APIKey == "" {
		log.Printf("⚠️ pixellab.api_key not found in config.yml - image generation will be disabled")
	}

	if config.GitHub.Token == "" {
		log.Printf("⚠️ github.token not found in config.yml - PR staging will be disabled")
	}

	return &config, nil
}
