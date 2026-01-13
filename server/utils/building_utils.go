package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// BuildingHours represents open/close times for a building
type BuildingHours struct {
	Open  interface{} `json:"open"`  // int (minutes 0-1439) or "always"
	Close interface{} `json:"close"` // int (minutes 0-1439) or null
}

// IsBuildingOpen checks if a building is accessible at the given time
// Returns: (isOpen bool, openMinutes int, closeMinutes int, err)
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
					return true, 0, 1440, nil
				}

				// Check if private/never accessible (open: -1)
				if openFloat, ok := hours.Open.(float64); ok && openFloat < 0 {
					return false, -1, -1, nil
				}

				// Get numeric minutes (0-1439)
				openMinutes, ok := hours.Open.(float64)
				if !ok {
					return false, 0, 0, fmt.Errorf("invalid open time format")
				}

				closeMinutes := 1440.0 // Default to end of day
				if hours.Close != nil {
					closeMinutes, ok = hours.Close.(float64)
					if !ok {
						return false, 0, 0, fmt.Errorf("invalid close time format")
					}
				}

				// Current time is already in minutes (0-1439)
				currentTimeMinutes := timeOfDay

				// Check if building is open
				open := int(openMinutes)
				close := int(closeMinutes)

				// Handle wrapping (e.g., 1020 minutes [5 PM] to 480 minutes [8 AM] for overnight)
				if close < open {
					// Overnight hours - open if current time >= open OR current time < close
					if currentTimeMinutes >= open || currentTimeMinutes < close {
						return true, open, close, nil
					}
				} else {
					// Normal hours - open if current time >= open AND current time < close
					if currentTimeMinutes >= open && currentTimeMinutes < close {
						return true, open, close, nil
					}
				}

				return false, open, close, nil
			}
		}
	}

	return false, 0, 0, fmt.Errorf("building not found: %s", buildingID)
}
