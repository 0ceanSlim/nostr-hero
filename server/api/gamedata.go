package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"nostr-hero/db"
)

// GameData represents the complete static game data bundle
type GameData struct {
	Items     []Item     `json:"items"`
	Spells    []Spell    `json:"spells"`
	Monsters  []Monster  `json:"monsters"`
	Locations []Location `json:"locations"`
	Packs     []Pack     `json:"packs"`
}

// Item represents game items
type Item struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ItemType    string                 `json:"item_type"`
	Properties  map[string]interface{} `json:"properties"`
	Tags        []string               `json:"tags"`
	Rarity      string                 `json:"rarity"`
}

// Spell represents game spells
type Spell struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Description string                `json:"description"`
	Level      int                    `json:"level"`
	School     string                 `json:"school"`
	Damage     string                 `json:"damage"`
	ManaCost   int                    `json:"mana_cost"`
	Classes    []string               `json:"classes"`
	Properties map[string]interface{} `json:"properties"`
}

// Monster represents game monsters
type Monster struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	ChallengeRating float64                `json:"challenge_rating"`
	Stats           map[string]interface{} `json:"stats"`
	Actions         map[string]interface{} `json:"actions"`
}

// Location represents game locations
type Location struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	LocationType string                 `json:"location_type"`
	Description  string                 `json:"description"`
	Image        string                 `json:"image,omitempty"`
	Music        string                 `json:"music,omitempty"`
	Properties   map[string]interface{} `json:"properties"`
	Connections  []string               `json:"connections"`
}

// Pack represents equipment packs
type Pack struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Items []interface{} `json:"items"`
}

// NPC represents non-player characters
type NPC struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"`
	Race        string                 `json:"race,omitempty"`
	Location    string                 `json:"location,omitempty"`
	Building    string                 `json:"building,omitempty"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties"`
}

// GameDataHandler serves all game data in one request for efficient loading
func GameDataHandler(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	if database == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	gameData := GameData{}

	// Load all static data in parallel
	errChan := make(chan error, 5)

	go func() {
		var err error
		gameData.Items, err = LoadAllItems(database)
		errChan <- err
	}()

	go func() {
		var err error
		gameData.Spells, err = LoadAllSpells(database)
		errChan <- err
	}()

	go func() {
		var err error
		gameData.Monsters, err = LoadAllMonsters(database)
		errChan <- err
	}()

	go func() {
		var err error
		gameData.Locations, err = LoadAllLocations(database)
		errChan <- err
	}()

	go func() {
		var err error
		gameData.Packs, err = LoadAllPacks(database)
		errChan <- err
	}()

	// Wait for all operations to complete
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			log.Printf("Error loading game data: %v", err)
			http.Error(w, "Failed to load game data", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	if err := json.NewEncoder(w).Encode(gameData); err != nil {
		log.Printf("Error encoding game data: %v", err)
		http.Error(w, "Failed to encode game data", http.StatusInternalServerError)
		return
	}

	log.Printf("Served game data: %d items, %d spells, %d monsters, %d locations, %d packs",
		len(gameData.Items), len(gameData.Spells), len(gameData.Monsters), len(gameData.Locations), len(gameData.Packs))
}

// Individual endpoints for specific data types
func ItemsHandler(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	if database == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	items, err := LoadAllItems(database)
	if err != nil {
		log.Printf("Error loading items: %v", err)
		http.Error(w, "Failed to load items", http.StatusInternalServerError)
		return
	}

	// Filter by name if provided (name is actually the item ID from starting-gear.json)
	nameQuery := r.URL.Query().Get("name")
	if nameQuery != "" {
		log.Printf("Filtering items by ID: '%s'", nameQuery)
		var filteredItems []Item
		for _, item := range items {
			// Match by ID (the item filename without .json)
			if item.ID == nameQuery {
				log.Printf("  ✓ Match found: ID='%s', Name='%s'", item.ID, item.Name)
				filteredItems = append(filteredItems, item)
			}
		}
		if len(filteredItems) == 0 {
			log.Printf("  ⚠️ No items matched ID '%s'. Checking first 5 items in database:", nameQuery)
			for i := 0; i < 5 && i < len(items); i++ {
				log.Printf("    - ID: '%s', Name: '%s'", items[i].ID, items[i].Name)
			}
		}
		items = filteredItems
		log.Printf("Returning %d filtered items", len(items))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func SpellsHandler(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	if database == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	// Extract spell ID from URL path if present (e.g., /api/spells/fire-bolt)
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/spells/"), "/")
	spellID := ""
	if len(pathParts) > 0 && pathParts[0] != "" {
		spellID = pathParts[0]
	}

	spells, err := LoadAllSpells(database)
	if err != nil {
		log.Printf("Error loading spells: %v", err)
		http.Error(w, "Failed to load spells", http.StatusInternalServerError)
		return
	}

	// Filter by ID if provided
	if spellID != "" {
		for _, spell := range spells {
			if spell.ID == spellID {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(spell)
				return
			}
		}
		http.Error(w, "Spell not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spells)
}

func MonstersHandler(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	if database == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	monsters, err := LoadAllMonsters(database)
	if err != nil {
		log.Printf("Error loading monsters: %v", err)
		http.Error(w, "Failed to load monsters", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monsters)
}

func LocationsHandler(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	if database == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	locations, err := LoadAllLocations(database)
	if err != nil {
		log.Printf("Error loading locations: %v", err)
		http.Error(w, "Failed to load locations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

func NPCsHandler(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	if database == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	npcs, err := LoadAllNPCs(database)
	if err != nil {
		log.Printf("Error loading NPCs: %v", err)
		http.Error(w, "Failed to load NPCs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(npcs)
}

// Database loading functions
func LoadAllItems(database *sql.DB) ([]Item, error) {
	rows, err := database.Query("SELECT id, name, description, item_type, properties, tags, rarity FROM items")
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %v", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var propertiesJSON, tagsJSON string

		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.ItemType,
			&propertiesJSON, &tagsJSON, &item.Rarity)
		if err != nil {
			log.Printf("Error scanning item row: %v", err)
			continue
		}

		// Parse JSON fields
		if propertiesJSON != "" {
			json.Unmarshal([]byte(propertiesJSON), &item.Properties)
		}
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &item.Tags)
		}

		items = append(items, item)
	}

	return items, nil
}

func LoadAllSpells(database *sql.DB) ([]Spell, error) {
	rows, err := database.Query("SELECT id, name, description, level, school, COALESCE(damage, ''), mana_cost, COALESCE(classes, ''), COALESCE(properties, '') FROM spells")
	if err != nil {
		return nil, fmt.Errorf("failed to query spells: %v", err)
	}
	defer rows.Close()

	var spells []Spell
	for rows.Next() {
		var spell Spell
		var propertiesJSON, classesJSON, damage string

		err := rows.Scan(&spell.ID, &spell.Name, &spell.Description, &spell.Level,
			&spell.School, &damage, &spell.ManaCost, &classesJSON, &propertiesJSON)
		if err != nil {
			log.Printf("Error scanning spell row: %v", err)
			continue
		}

		spell.Damage = damage

		// Parse JSON fields
		if propertiesJSON != "" {
			json.Unmarshal([]byte(propertiesJSON), &spell.Properties)
		}
		if classesJSON != "" {
			json.Unmarshal([]byte(classesJSON), &spell.Classes)
		}

		spells = append(spells, spell)
	}

	return spells, nil
}

func LoadAllMonsters(database *sql.DB) ([]Monster, error) {
	rows, err := database.Query("SELECT id, name, challenge_rating, stats, actions FROM monsters")
	if err != nil {
		return nil, fmt.Errorf("failed to query monsters: %v", err)
	}
	defer rows.Close()

	var monsters []Monster
	for rows.Next() {
		var monster Monster
		var statsJSON, actionsJSON string

		err := rows.Scan(&monster.ID, &monster.Name, &monster.ChallengeRating, &statsJSON, &actionsJSON)
		if err != nil {
			log.Printf("Error scanning monster row: %v", err)
			continue
		}

		// Parse JSON fields
		json.Unmarshal([]byte(statsJSON), &monster.Stats)
		json.Unmarshal([]byte(actionsJSON), &monster.Actions)

		monsters = append(monsters, monster)
	}

	return monsters, nil
}

func LoadAllLocations(database *sql.DB) ([]Location, error) {
	rows, err := database.Query("SELECT id, name, COALESCE(location_type, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(music, ''), COALESCE(properties, ''), COALESCE(connections, '') FROM locations")
	if err != nil {
		return nil, fmt.Errorf("failed to query locations: %v", err)
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var location Location
		var propertiesJSON, connectionsJSON string

		err := rows.Scan(&location.ID, &location.Name, &location.LocationType, &location.Description,
			&location.Image, &location.Music, &propertiesJSON, &connectionsJSON)
		if err != nil {
			log.Printf("Error scanning location row: %v", err)
			continue
		}

		// Parse JSON fields
		if propertiesJSON != "" {
			json.Unmarshal([]byte(propertiesJSON), &location.Properties)
		}
		if connectionsJSON != "" {
			json.Unmarshal([]byte(connectionsJSON), &location.Connections)
		}

		locations = append(locations, location)
	}

	return locations, nil
}

func LoadAllPacks(database *sql.DB) ([]Pack, error) {
	rows, err := database.Query("SELECT id, name, items FROM equipment_packs")
	if err != nil {
		return nil, fmt.Errorf("failed to query equipment packs: %v", err)
	}
	defer rows.Close()

	var packs []Pack
	for rows.Next() {
		var pack Pack
		var itemsJSON string

		err := rows.Scan(&pack.ID, &pack.Name, &itemsJSON)
		if err != nil {
			log.Printf("Error scanning pack row: %v", err)
			continue
		}

		// Parse JSON field
		json.Unmarshal([]byte(itemsJSON), &pack.Items)

		packs = append(packs, pack)
	}

	return packs, nil
}

func LoadAllNPCs(database *sql.DB) ([]NPC, error) {
	rows, err := database.Query("SELECT id, name, title, race, location, building, description, properties FROM npcs")
	if err != nil {
		return nil, fmt.Errorf("failed to query NPCs: %v", err)
	}
	defer rows.Close()

	var npcs []NPC
	for rows.Next() {
		var npc NPC
		var propertiesJSON string
		var title, race, location, building, description sql.NullString

		err := rows.Scan(&npc.ID, &npc.Name, &title, &race, &location, &building, &description, &propertiesJSON)
		if err != nil {
			log.Printf("Error scanning NPC row: %v", err)
			continue
		}

		// Handle nullable fields
		if title.Valid {
			npc.Title = title.String
		}
		if race.Valid {
			npc.Race = race.String
		}
		if location.Valid {
			npc.Location = location.String
		}
		if building.Valid {
			npc.Building = building.String
		}
		if description.Valid {
			npc.Description = description.String
		}

		// Parse JSON field
		if propertiesJSON != "" {
			json.Unmarshal([]byte(propertiesJSON), &npc.Properties)
		}

		npcs = append(npcs, npc)
	}

	return npcs, nil
}