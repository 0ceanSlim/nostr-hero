package character

import (
	"encoding/json"
	"fmt"
	"net/http"

	"nostr-hero/db"
)

// WeightsHandler serves character generation weights from DuckDB
func WeightsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get weights data from DuckDB
	weightsData, err := GetWeightsFromDB()
	if err != nil {
		http.Error(w, "Failed to load weights data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weightsData)
}

// GetWeightsFromDB retrieves all character generation weights from database
func GetWeightsFromDB() (map[string]interface{}, error) {
	weights, err := db.GetGenerationWeights()
	if err != nil {
		return nil, err
	}

	// Convert struct to map[string]interface{} for JSON serialization
	return map[string]interface{}{
		"Races":                    weights.Races,
		"RaceWeights":              weights.RaceWeights,
		"classWeightsByRace":       weights.ClassWeightsByRace,
		"BackgroundWeightsByClass": weights.BackgroundWeightsByClass,
		"Alignments":               weights.Alignments,
		"AlignmentWeights":         weights.AlignmentWeights,
	}, nil
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
