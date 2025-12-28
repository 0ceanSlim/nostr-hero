package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// BuildingHours represents open/close times for a building
type BuildingHours struct {
	Open  interface{} `json:"open"`  // int (hour 0-23) or "always"
	Close interface{} `json:"close"` // int (hour 0-23) or null
}

// IsBuildingOpen checks if a building is accessible at the given time
// Returns: (isOpen bool, openHour int, closeHour int, err)
func IsBuildingOpen(db *sql.DB, locationID, buildingID string, timeOfDay int) (bool, int, int, error) {
	// Get location data from database
	var propertiesJSON string
	err := db.QueryRow("SELECT properties FROM locations WHERE id = ?", locationID).Scan(&propertiesJSON)
	if err != nil {
		return false, 0, 0, fmt.Errorf("location not found: %s", locationID)
	}

	var locationData map[string]interface{}
	if err := json.Unmarshal([]byte(propertiesJSON), &locationData); err != nil {
		return false, 0, 0, fmt.Errorf("failed to parse location data: %v", err)
	}

	// Navigate to building in districts
	districts, ok := locationData["districts"].(map[string]interface{})
	if !ok {
		return false, 0, 0, fmt.Errorf("location has no districts")
	}

	// Search all districts for the building
	for _, districtData := range districts {
		district, ok := districtData.(map[string]interface{})
		if !ok {
			continue
		}

		buildings, ok := district["buildings"].([]interface{})
		if !ok {
			continue
		}

		for _, buildingData := range buildings {
			building, ok := buildingData.(map[string]interface{})
			if !ok {
				continue
			}

			if building["id"] == buildingID {
				// Found the building, check hours
				hours := BuildingHours{
					Open:  building["open"],
					Close: building["close"],
				}

				// Check if always open
				if openStr, ok := hours.Open.(string); ok && openStr == "always" {
					return true, 0, 24, nil
				}

				// Check if private/never accessible (open: -1)
				if openFloat, ok := hours.Open.(float64); ok && openFloat < 0 {
					return false, -1, -1, nil
				}

				// Get numeric hours
				openHour, ok := hours.Open.(float64)
				if !ok {
					return false, 0, 0, fmt.Errorf("invalid open hour format")
				}

				closeHour := 24.0
				if hours.Close != nil {
					closeHour, ok = hours.Close.(float64)
					if !ok {
						return false, 0, 0, fmt.Errorf("invalid close hour format")
					}
				}

				// Convert timeOfDay (minutes) to hour
				currentHour := timeOfDay / 60

				// Check if building is open
				open := int(openHour)
				close := int(closeHour)

				// Handle wrapping (e.g., 22:00 to 04:00)
				if close < open {
					if currentHour >= open || currentHour < close {
						return true, open, close, nil
					}
				} else {
					if currentHour >= open && currentHour < close {
						return true, open, close, nil
					}
				}

				return false, open, close, nil
			}
		}
	}

	return false, 0, 0, fmt.Errorf("building not found: %s", buildingID)
}
