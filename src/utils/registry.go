package utils

import (
	"encoding/json"
	"os"

	"nostr-hero/src/types"
)

const registryFile = "we/data/alpha-registry.json"

func ReadRegistry() ([]types.RegistryEntry, error) {
	var registry []types.RegistryEntry

	file, err := os.ReadFile(registryFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.RegistryEntry{}, nil // Return empty list if file doesn't exist
		}
		return nil, err
	}

	err = json.Unmarshal(file, &registry)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func WriteRegistry(registry []types.RegistryEntry) error {
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(registryFile, data, 0644)
}

func IsNpubInRegistry(npub string, registry []types.RegistryEntry) bool {
	for _, entry := range registry {
		if entry.Npub == npub {
			return true
		}
	}
	return false
}
