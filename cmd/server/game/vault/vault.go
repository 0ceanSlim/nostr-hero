package vault

import (
	"log"

	"pubkey-quest/types"
)

// IsVaultRegistered checks if a vault is registered at the specified building
func IsVaultRegistered(state *types.SaveFile, buildingID string) bool {
	if state.Vaults == nil {
		return false
	}
	for _, vault := range state.Vaults {
		// Check new format (building field)
		if building, ok := vault["building"].(string); ok {
			if building == buildingID {
				return true
			}
		} else if location, ok := vault["location"].(string); ok {
			// Check old format (location field) - match if we're at that location
			if location == state.Location {
				return true
			}
		}
	}
	return false
}

// RegisterVault registers a new vault at the specified building
func RegisterVault(state *types.SaveFile, buildingID string) {
	if state.Vaults == nil {
		state.Vaults = []map[string]interface{}{}
	}

	// Check if already registered
	for _, vault := range state.Vaults {
		if building, ok := vault["building"].(string); ok && building == buildingID {
			return // Already registered
		}
	}

	// Create new vault with 40 empty slots
	slots := make([]map[string]interface{}, 40)
	for i := range 40 {
		slots[i] = map[string]interface{}{
			"slot":     i,
			"item":     nil,
			"quantity": 0,
		}
	}

	vault := map[string]interface{}{
		"building": buildingID,
		"slots":    slots,
	}

	state.Vaults = append(state.Vaults, vault)
	log.Printf("✅ Registered vault at %s", buildingID)
}

// GetVaultForLocation returns the vault for the specified building
func GetVaultForLocation(state *types.SaveFile, buildingID string) map[string]interface{} {
	if state.Vaults == nil {
		return nil
	}

	for _, vault := range state.Vaults {
		// Check new format (building field)
		if building, ok := vault["building"].(string); ok && building == buildingID {
			return vault
		}
		// Check old format (location field) - return if we're at that location
		if location, ok := vault["location"].(string); ok && location == state.Location {
			return vault
		}
	}

	return nil
}

// HandleVaultDepositAction deposits items into vault (uses existing move_item action for vault transfers)
func HandleVaultDepositAction(_ *types.SaveFile, _ map[string]interface{}) (*types.GameActionResponse, error) {
	// Vaults work like containers - use the container system
	// This is handled by frontend calling move_item or add_to_container with vault as destination
	return &types.GameActionResponse{
		Success: true,
		Message: "Item deposited to vault",
	}, nil
}

// HandleVaultWithdrawAction withdraws items from vault (uses existing move_item action for vault transfers)
func HandleVaultWithdrawAction(_ *types.SaveFile, _ map[string]interface{}) (*types.GameActionResponse, error) {
	// Vaults work like containers - use the container system
	// This is handled by frontend calling move_item or remove_from_container with vault as source
	return &types.GameActionResponse{
		Success: true,
		Message: "Item withdrawn from vault",
	}, nil
}

// HandleRegisterVaultAction registers a vault (called after payment)
func HandleRegisterVaultAction(state *types.SaveFile, _ map[string]interface{}) (*types.GameActionResponse, error) {
	buildingID := state.Building
	if buildingID == "" {
		return &types.GameActionResponse{
			Success: false,
			Error:   "not in a building",
			Color:   "red",
		}, nil
	}

	RegisterVault(state, buildingID)

	return &types.GameActionResponse{
		Success: true,
		Message: "Vault registered successfully",
		Color:   "green",
	}, nil
}

// HandleOpenVaultAction returns vault data for UI
func HandleOpenVaultAction(state *types.SaveFile, _ map[string]interface{}) (*types.GameActionResponse, error) {
	buildingID := state.Building
	if buildingID == "" {
		return &types.GameActionResponse{
			Success: false,
			Error:   "not in a building",
			Color:   "red",
		}, nil
	}

	vault := GetVaultForLocation(state, buildingID)
	if vault == nil {
		return &types.GameActionResponse{
			Success: false,
			Error:   "no vault registered at this location",
			Color:   "red",
		}, nil
	}

	return &types.GameActionResponse{
		Success: true,
		Message: "Vault opened",
		Delta: map[string]interface{}{
			"vault": vault,
		},
	}, nil
}
