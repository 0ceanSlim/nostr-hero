package character

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0ceanslim/grain/client/core/tools"

	gamecharacter "nostr-hero/game/character"
	"nostr-hero/types"
	"nostr-hero/utils"
)

// CharacterResponse represents the generated character response
// swagger:model CharacterResponse
type CharacterResponse struct {
	Npub      string                 `json:"npub" example:"npub1..."`
	Pubkey    string                 `json:"pubkey" example:"abc123..."`
	Character map[string]interface{} `json:"character"`
}

// CharacterHandler godoc
// @Summary      Generate character
// @Description  Generates a deterministic character based on Nostr npub. The same npub always produces the same character.
// @Tags         Character
// @Produce      json
// @Param        npub  query     string  true  "Nostr public key (npub format)"
// @Success      200   {object}  CharacterResponse
// @Failure      400   {string}  string  "Missing or invalid npub"
// @Failure      500   {string}  string  "Server error"
// @Router       /character [get]
func CharacterHandler(w http.ResponseWriter, r *http.Request) {
	npub := r.URL.Query().Get("npub")
	if npub == "" {
		http.Error(w, "Missing npub parameter", http.StatusBadRequest)
		return
	}

	pubKey, err := tools.DecodeNpub(npub)
	if err != nil {
		http.Error(w, "Invalid npub", http.StatusBadRequest)
		return
	}

	// Load weight data from DuckDB (same as weights API endpoint)
	weightDataMap, err := GetWeightsFromDB()
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
	generatedChar := gamecharacter.GenerateCharacter(pubKey, &weightData)

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
