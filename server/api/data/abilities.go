package data

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"nostr-hero/db"
)

// AbilityTier represents a scaling tier for an ability
type AbilityTier struct {
	MinLevel         int      `json:"min_level"`
	MaxLevel         int      `json:"max_level"`
	EffectsApplied   []string `json:"effects_applied"`
	Summary          string   `json:"summary"`
	OverrideCost     *int     `json:"override_cost,omitempty"`
	OverrideCooldown *int     `json:"override_cooldown,omitempty"`
}

// Ability represents a martial class ability definition
type Ability struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Class        string        `json:"class"`
	UnlockLevel  int           `json:"unlock_level"`
	ResourceCost int           `json:"resource_cost"`
	ResourceType string        `json:"resource_type"`
	Cooldown     string        `json:"cooldown"`
	Description  string        `json:"description"`
	ScalingTiers []AbilityTier `json:"scaling_tiers"`
}

// AbilityResponse represents an ability with computed tier info
type AbilityResponse struct {
	Ability
	IsUnlocked  bool          `json:"is_unlocked"`
	CurrentTier *AbilityTier  `json:"current_tier"`
	NextTier    *AbilityTier  `json:"next_tier"`
	AllTiers    []AbilityTier `json:"all_tiers"`
}

// AbilitiesListResponse is the API response for abilities list
type AbilitiesListResponse struct {
	Success   bool              `json:"success"`
	Class     string            `json:"class"`
	Level     int               `json:"level"`
	Abilities []AbilityResponse `json:"abilities"`
}

// Valid martial classes that have abilities
var validMartialClasses = map[string]bool{
	"fighter":   true,
	"barbarian": true,
	"monk":      true,
	"rogue":     true,
}

// AbilitiesHandler serves ability data for martial classes
// GET /api/abilities?class={class}&level={level}
func AbilitiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	className := strings.ToLower(r.URL.Query().Get("class"))
	levelStr := r.URL.Query().Get("level")

	if className == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "class parameter is required",
		})
		return
	}

	if !validMartialClasses[className] {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "class must be one of: fighter, barbarian, monk, rogue",
		})
		return
	}

	level := 1
	if levelStr != "" {
		parsed, err := strconv.Atoi(levelStr)
		if err != nil || parsed < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   "level must be a positive integer",
			})
			return
		}
		level = parsed
	}

	// Load abilities from filesystem
	abilities, err := loadAbilitiesForClass(className)
	if err != nil {
		log.Printf("❌ Error loading abilities for %s: %v", className, err)
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to load abilities",
		})
		return
	}

	// Build response with tier calculations
	var responses []AbilityResponse
	for _, ability := range abilities {
		resp := AbilityResponse{
			Ability:    ability,
			IsUnlocked: level >= ability.UnlockLevel,
			AllTiers:   ability.ScalingTiers,
		}

		// Find current and next tier based on level
		for i, tier := range ability.ScalingTiers {
			if level >= tier.MinLevel && level <= tier.MaxLevel {
				resp.CurrentTier = &ability.ScalingTiers[i]
				// Next tier is the one after current
				if i+1 < len(ability.ScalingTiers) {
					resp.NextTier = &ability.ScalingTiers[i+1]
				}
				break
			}
		}

		// If no current tier but ability is unlocked, use last applicable tier
		if resp.CurrentTier == nil && resp.IsUnlocked && len(ability.ScalingTiers) > 0 {
			last := ability.ScalingTiers[len(ability.ScalingTiers)-1]
			resp.CurrentTier = &last
		}

		responses = append(responses, resp)
	}

	writeJSON(w, http.StatusOK, AbilitiesListResponse{
		Success:   true,
		Class:     className,
		Level:     level,
		Abilities: responses,
	})
}

// loadAbilitiesForClass reads all abilities for a given class from the database
func loadAbilitiesForClass(className string) ([]Ability, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	rows, err := database.Query(
		"SELECT id, name, class, unlock_level, resource_cost, COALESCE(resource_type, ''), COALESCE(cooldown, ''), COALESCE(description, ''), COALESCE(properties, '') FROM abilities WHERE class = ?",
		className,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query abilities: %v", err)
	}
	defer rows.Close()

	var abilities []Ability
	for rows.Next() {
		var ability Ability
		var propertiesJSON string

		err := rows.Scan(&ability.ID, &ability.Name, &ability.Class,
			&ability.UnlockLevel, &ability.ResourceCost, &ability.ResourceType,
			&ability.Cooldown, &ability.Description, &propertiesJSON)
		if err != nil {
			log.Printf("⚠️ Error scanning ability row: %v", err)
			continue
		}

		// Parse scaling_tiers from the full properties JSON
		if propertiesJSON != "" {
			var fullAbility Ability
			if err := json.Unmarshal([]byte(propertiesJSON), &fullAbility); err == nil {
				ability.ScalingTiers = fullAbility.ScalingTiers
			} else {
				log.Printf("⚠️ Error parsing ability properties for %s: %v", ability.ID, err)
			}
		}

		abilities = append(abilities, ability)
	}

	// Sort by unlock level
	sortAbilitiesByLevel(abilities)

	return abilities, nil
}

// sortAbilitiesByLevel sorts abilities by their unlock level (ascending)
func sortAbilitiesByLevel(abilities []Ability) {
	for i := 1; i < len(abilities); i++ {
		key := abilities[i]
		j := i - 1
		for j >= 0 && abilities[j].UnlockLevel > key.UnlockLevel {
			abilities[j+1] = abilities[j]
			j--
		}
		abilities[j+1] = key
	}
}

// writeJSON helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("❌ Error encoding JSON response: %v", err)
	}
}
