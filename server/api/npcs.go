package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"nostr-hero/db"
	"nostr-hero/types"
	"nostr-hero/utils"
)

// NPCLocationResponse represents NPC visibility at a location
type NPCLocationResponse struct {
	NPCID          string `json:"npc_id"`
	Name           string `json:"name"`
	Title          string `json:"title"`
	LocationType   string `json:"location_type"` // "building" or "district"
	LocationID     string `json:"location_id"`
	State          string `json:"state"`
	IsInteractable bool   `json:"is_interactable"`
}

// GetNPCsAtLocationHandler returns NPCs visible at player's current location and time
// GET /api/npcs/at-location?location={locationID}&district={districtID}&building={buildingID}&time={timeOfDay}
func GetNPCsAtLocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	locationID := r.URL.Query().Get("location")
	districtID := r.URL.Query().Get("district")
	buildingID := r.URL.Query().Get("building")
	timeOfDay := 720 // Default noon

	if t := r.URL.Query().Get("time"); t != "" {
		fmt.Sscanf(t, "%d", &timeOfDay)
	}

	database := db.GetDB()
	if database == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	// Get all NPCs for this location
	rows, err := database.Query("SELECT id, properties FROM npcs WHERE location = ?", locationID)
	if err != nil {
		http.Error(w, "Failed to query NPCs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	visibleNPCs := []NPCLocationResponse{}

	for rows.Next() {
		var npcID, propertiesJSON string
		if err := rows.Scan(&npcID, &propertiesJSON); err != nil {
			continue
		}

		var npcData types.NPCData
		if err := json.Unmarshal([]byte(propertiesJSON), &npcData); err != nil {
			continue
		}

		// Resolve schedule
		scheduleInfo := utils.ResolveNPCSchedule(&npcData, timeOfDay)

		// Determine location type from location ID
		locationType := utils.DetermineLocationType(scheduleInfo.Location)

		// Check if NPC is at player's current location
		isVisible := false
		if buildingID != "" && locationType == "building" && scheduleInfo.Location == buildingID {
			isVisible = true
		} else if buildingID == "" && locationType == "district" && scheduleInfo.Location == districtID {
			isVisible = true
		}

		if isVisible {
			visibleNPCs = append(visibleNPCs, NPCLocationResponse{
				NPCID:          npcID,
				Name:           npcData.Name,
				Title:          npcData.Title,
				LocationType:   locationType,
				LocationID:     scheduleInfo.Location,
				State:          scheduleInfo.State,
				IsInteractable: scheduleInfo.IsAvailable,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visibleNPCs)
}
