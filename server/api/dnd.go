package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"nostr-hero/game/character"
	"nostr-hero/types"
	"nostr-hero/utils"
)

func CharacterHandler(w http.ResponseWriter, r *http.Request) {
	npub := r.URL.Query().Get("npub")
	if npub == "" {
		http.Error(w, "Missing npub parameter", http.StatusBadRequest)
		return
	}

	pubKey, err := utils.DecodeNpub(npub)
	if err != nil {
		http.Error(w, "Invalid npub", http.StatusBadRequest)
		return
	}

	// Load weight data from DuckDB (same as weights API endpoint)
	weightDataMap, err := getWeightsFromDB()
	if err != nil {
		http.Error(w, "Error loading weight data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert map to WeightData struct
	weightDataJSON, err := json.Marshal(weightDataMap)
	if err != nil {
		http.Error(w, "Error marshaling weight data", http.StatusInternalServerError)
		return
	}

	var weightData types.WeightData
	if err := json.Unmarshal(weightDataJSON, &weightData); err != nil {
		http.Error(w, "Error unmarshaling weight data", http.StatusInternalServerError)
		return
	}

	// Generate character using the loaded weight data
	generatedChar := character.GenerateCharacter(pubKey, &weightData)

	registry, err := utils.ReadRegistry()
	if err != nil {
		http.Error(w, "Error reading registry", http.StatusInternalServerError)
		return
	}

	if !utils.IsNpubInRegistry(npub, registry) {
		newEntry := types.RegistryEntry{
			Npub:      npub,
			PubKey:    pubKey,
			Character: generatedChar,
		}
		registry = append(registry, newEntry)
		if err := utils.WriteRegistry(registry); err != nil {
			fmt.Println("Error writing to registry:", err)
		}
		fmt.Println("Logged new entry for npub:", npub)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"npub":      npub,
		"pubkey":    pubKey,
		"character": generatedChar,
	})
}
