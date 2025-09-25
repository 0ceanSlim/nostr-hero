package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds server-related configurations
type ServerConfig struct {
	Port     int    `yaml:"port"`
	TLS      bool   `yaml:"tls"`
	AppTitle string `yaml:"app_title"`
}

// Config holds the full application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
}

// Global variable to hold the config after loading
var AppConfig Config

// LoadConfig reads the YAML config file into the AppConfig struct
func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}
