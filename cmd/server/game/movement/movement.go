package movement

import (
	"fmt"
	"log"
	"slices"

	"pubkey-quest/cmd/server/db"
	"pubkey-quest/cmd/server/game/building"
	"pubkey-quest/types"
)

// HandleMoveAction moves to a new location
func HandleMoveAction(state *types.SaveFile, params map[string]interface{}) (*types.GameActionResponse, error) {
	location, ok := params["location"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid location parameter")
	}

	district, _ := params["district"].(string)
	building, _ := params["building"].(string)

	// Update state
	state.Location = location
	state.District = district
	state.Building = building

	// Add to discovered locations if not already there
	if !slices.Contains(state.LocationsDiscovered, location) {
		state.LocationsDiscovered = append(state.LocationsDiscovered, location)
	}

	return &types.GameActionResponse{
		Success: true,
		Message: fmt.Sprintf("Moved to %s", location),
	}, nil
}

// HandleEnterBuildingAction enters a building
func HandleEnterBuildingAction(state *types.SaveFile, params map[string]interface{}) (*types.GameActionResponse, error) {
	buildingID, ok := params["building_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid building_id parameter")
	}

	// Check if building is open
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not available")
	}

	isOpen, openMinutes, closeMinutes, err := building.IsBuildingOpen(database, state.Location, buildingID, state.TimeOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to check building hours: %v", err)
	}

	if !isOpen {
		// Convert minutes to hours:minutes format for display
		openHour := openMinutes / 60
		openMin := openMinutes % 60
		closeHour := closeMinutes / 60
		closeMin := closeMinutes % 60

		// Format times with AM/PM
		formatTime := func(hour, min int) string {
			period := "AM"
			displayHour := hour
			if hour >= 12 {
				period = "PM"
				if hour > 12 {
					displayHour = hour - 12
				}
			}
			if displayHour == 0 {
				displayHour = 12
			}
			return fmt.Sprintf("%d:%02d %s", displayHour, min, period)
		}

		return &types.GameActionResponse{
			Success: false,
			Message: fmt.Sprintf("The building is closed. Open hours: %s - %s", formatTime(openHour, openMin), formatTime(closeHour, closeMin)),
			Color:   "red",
		}, nil
	}

	// Update state to include building
	state.Building = buildingID

	log.Printf("🏛️ Entered building: %s", buildingID)

	return &types.GameActionResponse{
		Success: true,
		Message: "Entered building",
		Color:   "blue",
	}, nil
}

// HandleExitBuildingAction exits a building
func HandleExitBuildingAction(state *types.SaveFile, _ map[string]interface{}) (*types.GameActionResponse, error) {
	// Update state to remove building (back outdoors)
	state.Building = ""

	log.Printf("🚪 Exited building")

	// Check fatigue level to warn user
	message := "Exited building"
	if state.Fatigue > 0 {
		message = fmt.Sprintf("Exited building (Fatigue: %d)", state.Fatigue)
	}

	return &types.GameActionResponse{
		Success: true,
		Message: message,
		Color:   "blue",
	}, nil
}
