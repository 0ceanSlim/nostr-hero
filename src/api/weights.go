package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"nostr-hero/src/db"
)

// WeightsHandler serves character generation weights from DuckDB
func WeightsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get weights data from DuckDB
	weightsData, err := getWeightsFromDB()
	if err != nil {
		http.Error(w, "Failed to load weights data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weightsData)
}

// getWeightsFromDB retrieves all character generation weights from DuckDB
func getWeightsFromDB() (map[string]interface{}, error) {
	// For now, we'll construct the weights structure manually from the migrated data
	// In the future, this could be stored as a single JSON document in the database

	weights := map[string]interface{}{
		"Races": []string{
			"Human", "Elf", "Dwarf", "Halfling", "Gnome", "Orc",
			"Half-Elf", "Dragonborn", "Tiefling", "Half-Orc",
		},
		"RaceWeights": []int{36, 12, 14, 8, 6, 10, 7, 2, 3, 2},
		"classWeightsByRace": map[string]map[string]int{
			"Human": {
				"Fighter": 12, "Rogue": 8, "Bard": 8, "Cleric": 8,
				"Paladin": 8, "Ranger": 8, "Monk": 8, "Barbarian": 8,
				"Druid": 8, "Sorcerer": 8, "Warlock": 8, "Wizard": 8,
			},
			"Elf": {
				"Wizard": 16, "Ranger": 16, "Rogue": 10, "Bard": 8,
				"Sorcerer": 8, "Fighter": 8, "Druid": 8, "Cleric": 8,
				"Monk": 8, "Warlock": 8, "Paladin": 4, "Barbarian": 4,
			},
			"Dwarf": {
				"Fighter": 18, "Barbarian": 18, "Cleric": 14, "Paladin": 14,
				"Monk": 8, "Ranger": 8, "Druid": 8, "Rogue": 2,
				"Sorcerer": 2, "Bard": 2, "Warlock": 2, "Wizard": 2,
			},
			"Halfling": {
				"Rogue": 20, "Ranger": 14, "Fighter": 14, "Monk": 10,
				"Bard": 10, "Sorcerer": 10, "Warlock": 10, "Cleric": 10,
				"Druid": 10, "Barbarian": 4, "Paladin": 4, "Wizard": 4,
			},
			"Gnome": {
				"Wizard": 20, "Rogue": 20, "Sorcerer": 8, "Fighter": 8,
				"Warlock": 8, "Bard": 8, "Ranger": 6, "Monk": 6,
				"Paladin": 6, "Cleric": 4, "Druid": 4, "Barbarian": 2,
			},
			"Orc": {
				"Barbarian": 22, "Fighter": 16, "Rogue": 12, "Druid": 12,
				"Warlock": 6, "Ranger": 6, "Monk": 6, "Cleric": 6,
				"Paladin": 6, "Sorcerer": 4, "Bard": 4, "Wizard": 4,
			},
			"Half-Elf": {
				"Bard": 14, "Sorcerer": 14, "Monk": 12, "Paladin": 10,
				"Rogue": 10, "Fighter": 8, "Cleric": 8, "Druid": 8,
				"Ranger": 8, "Wizard": 8, "Warlock": 6, "Barbarian": 4,
			},
			"Dragonborn": {
				"Paladin": 22, "Sorcerer": 14, "Warlock": 14, "Bard": 14,
				"Fighter": 8, "Barbarian": 8, "Monk": 8, "Rogue": 2,
				"Cleric": 2, "Druid": 2, "Ranger": 2, "Wizard": 2,
			},
			"Tiefling": {
				"Sorcerer": 18, "Warlock": 18, "Wizard": 10, "Bard": 10,
				"Paladin": 8, "Fighter": 8, "Rogue": 8, "Monk": 8,
				"Ranger": 8, "Cleric": 2, "Druid": 2, "Barbarian": 2,
			},
			"Half-Orc": {
				"Barbarian": 22, "Fighter": 22, "Monk": 10, "Cleric": 10,
				"Paladin": 10, "Warlock": 10, "Bard": 6, "Ranger": 6,
				"Druid": 6, "Rogue": 2, "Sorcerer": 2, "Wizard": 2,
			},
		},
		"BackgroundWeightsByClass": map[string]map[string]int{
			"Paladin": {
				"Acolyte": 13, "Folk Hero": 13, "Gladiator": 12,
				"Guard": 12, "Knight": 12, "Noble": 12, "Soldier": 13, "Wayfarer": 13,
			},
			"Sorcerer": {
				"Charlatan": 16, "Entertainer": 12, "Hermit": 11,
				"Noble": 12, "Sage": 14, "Urchin": 11, "Wayfarer": 11, "Scribe": 13,
			},
			"Warlock": {
				"Charlatan": 17, "Criminal": 9, "Spy": 8, "Hermit": 17,
				"Outlander": 16, "Sage": 16, "Sailor": 7, "Urchin": 10,
			},
			"Bard": {
				"Charlatan": 13, "Entertainer": 13, "Artisan": 6, "Merchant": 7,
				"Noble": 13, "Sage": 12, "Sailor": 12, "Urchin": 12, "Wayfarer": 12,
			},
			"Fighter": {
				"Folk Hero": 15, "Gladiator": 15, "Knight": 16, "Outlander": 12,
				"Pirate": 7, "Sailor": 6, "Soldier": 15, "Farmer": 13,
			},
			"Barbarian": {
				"Folk Hero": 20, "Outlander": 23, "Pirate": 20,
				"Soldier": 21, "Urchin": 16,
			},
			"Monk": {
				"Acolyte": 16, "Hermit": 25, "Outlander": 13,
				"Sage": 18, "Wayfarer": 10, "Scribe": 18,
			},
			"Rogue": {
				"Charlatan": 17, "Criminal": 17, "Artisan": 8, "Merchant": 9,
				"Pirate": 16, "Sailor": 16, "Urchin": 17,
			},
			"Cleric": {
				"Acolyte": 21, "Folk Hero": 16, "Hermit": 21,
				"Sage": 13, "Wayfarer": 18, "Scribe": 13,
			},
			"Druid": {
				"Farmer": 20, "Hermit": 20, "Outlander": 20,
				"Sage": 20, "Wayfarer": 20,
			},
			"Ranger": {
				"Farmer": 17, "Folk Hero": 17, "Guide": 17,
				"Outlander": 17, "Sailor": 16, "Soldier": 16,
			},
			"Wizard": {
				"Acolyte": 16, "Artisan": 10, "Merchant": 10,
				"Hermit": 10, "Noble": 20, "Sage": 17, "Scribe": 17,
			},
		},
		"Alignments": []string{
			"Lawful Good", "Neutral Good", "Chaotic Good",
			"Lawful Neutral", "True Neutral", "Chaotic Neutral",
			"Lawful Evil", "Neutral Evil", "Chaotic Evil",
		},
		"AlignmentWeights": []int{10, 10, 10, 10, 20, 10, 10, 10, 10},
	}

	return weights, nil
}

// IntroductionsHandler serves character introduction data from database
func IntroductionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get introductions data from database
	introductions, err := getIntroductionsFromDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load introductions: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(introductions))
}

// getIntroductionsFromDB retrieves introductions data from database
func getIntroductionsFromDB() (string, error) {
	database := db.GetDB()
	if database == nil {
		return "", fmt.Errorf("database not available")
	}

	var dataJSON string
	err := database.QueryRow("SELECT data FROM introductions WHERE id = 'introductions'").Scan(&dataJSON)
	if err != nil {
		return "", fmt.Errorf("failed to query introductions: %v", err)
	}
	return dataJSON, nil
}

// StartingGearHandler serves starting equipment data from database
func StartingGearHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get starting gear data from database
	startingGear, err := getStartingGearFromDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load starting gear: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(startingGear))
}

// getStartingGearFromDB retrieves starting gear data from database
func getStartingGearFromDB() (string, error) {
	database := db.GetDB()
	if database == nil {
		return "", fmt.Errorf("database not available")
	}

	var dataJSON string
	err := database.QueryRow("SELECT data FROM starting_gear WHERE id = 'starting-gear'").Scan(&dataJSON)
	if err != nil {
		return "", fmt.Errorf("failed to query starting gear: %v", err)
	}
	return dataJSON, nil
}