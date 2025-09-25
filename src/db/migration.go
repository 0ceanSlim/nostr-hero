package db

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"os"
)

// MigrateFromJSON imports all JSON data into the database
func MigrateFromJSON() error {
	log.Println("Starting JSON to DuckDB migration...")

	// Migrate equipment packs
	if err := migrateEquipmentPacks(); err != nil {
		return fmt.Errorf("failed to migrate equipment packs: %v", err)
	}

	// Migrate character data
	if err := migrateCharacterData(); err != nil {
		return fmt.Errorf("failed to migrate character data: %v", err)
	}

	// Migrate content data (monsters, locations)
	if err := migrateContentData(); err != nil {
		return fmt.Errorf("failed to migrate content data: %v", err)
	}

	log.Println("Migration completed successfully!")
	return nil
}

// migrateEquipmentPacks migrates equipment packs from JSON
func migrateEquipmentPacks() error {
	log.Println("Migrating equipment packs...")

	packsPath := filepath.Join("docs", "data", "equipment", "packs.json")
	data, err := os.ReadFile(packsPath)
	if err != nil {
		return fmt.Errorf("failed to read packs.json: %v", err)
	}

	var packsData struct {
		Packs map[string][][2]interface{} `json:"packs"`
	}

	if err := json.Unmarshal(data, &packsData); err != nil {
		return fmt.Errorf("failed to unmarshal packs data: %v", err)
	}

	// Clear existing data
	if _, err := db.Exec("DELETE FROM equipment_packs"); err != nil {
		return fmt.Errorf("failed to clear equipment_packs table: %v", err)
	}

	// Insert equipment packs
	stmt := `INSERT INTO equipment_packs (id, name, items) VALUES (?, ?, ?)`
	for name, items := range packsData.Packs {
		id := strings.ToLower(strings.ReplaceAll(name, "'", ""))
		id = strings.ReplaceAll(id, " ", "-")

		itemsJSON, err := json.Marshal(items)
		if err != nil {
			return fmt.Errorf("failed to marshal items for pack %s: %v", name, err)
		}

		if _, err := db.Exec(stmt, id, name, string(itemsJSON)); err != nil {
			return fmt.Errorf("failed to insert pack %s: %v", name, err)
		}
	}

	log.Printf("Migrated %d equipment packs", len(packsData.Packs))
	return nil
}

// migrateCharacterData migrates character-related JSON files
func migrateCharacterData() error {
	log.Println("Migrating character data...")

	characterDataPath := filepath.Join("docs", "data", "character")

	// Define the files we want to migrate for character data
	characterFiles := map[string]string{
		"advancement.json":           "character_advancement",
		"base-hp.json":              "character_base_hp",
		"racial-starting-cities.json": "racial_starting_cities",
		"spell-progression.json":     "spell_progression",
		"starting-gear.json":         "starting_gear",
		"starting-gold.json":         "starting_gold",
		"starting-spells.json":       "starting_spells",
	}

	for filename, tableName := range characterFiles {
		filePath := filepath.Join(characterDataPath, filename)
		if err := migrateGenericJSON(filePath, tableName); err != nil {
			log.Printf("Warning: failed to migrate %s: %v", filename, err)
		}
	}

	return nil
}

// migrateContentData migrates monsters, locations, and other content
func migrateContentData() error {
	log.Println("Migrating content data...")

	// Migrate monsters
	if err := migrateMonsters(); err != nil {
		return fmt.Errorf("failed to migrate monsters: %v", err)
	}

	// Migrate locations
	if err := migrateLocations(); err != nil {
		return fmt.Errorf("failed to migrate locations: %v", err)
	}

	return nil
}

// migrateMonsters migrates all monster JSON files
func migrateMonsters() error {
	monstersPath := filepath.Join("docs", "data", "content", "monsters")

	// Clear existing monsters
	if _, err := db.Exec("DELETE FROM monsters"); err != nil {
		return fmt.Errorf("failed to clear monsters table: %v", err)
	}

	count := 0
	err := filepath.WalkDir(monstersPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			if err := migrateMonsterFile(path); err != nil {
				log.Printf("Warning: failed to migrate monster file %s: %v", path, err)
			} else {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk monsters directory: %v", err)
	}

	log.Printf("Migrated %d monsters", count)
	return nil
}

// migrateMonsterFile migrates a single monster JSON file
func migrateMonsterFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var monster map[string]interface{}
	if err := json.Unmarshal(data, &monster); err != nil {
		return err
	}

	// Extract base filename as ID
	id := strings.TrimSuffix(filepath.Base(filePath), ".json")

	// Convert monster data to required fields
	name, _ := monster["name"].(string)
	challengeRating, _ := monster["challenge_rating"].(float64)

	// Serialize stats and actions as JSON
	statsJSON, _ := json.Marshal(monster)
	actionsJSON, _ := json.Marshal(map[string]interface{}{}) // Empty for now

	stmt := `INSERT INTO monsters (id, name, challenge_rating, stats, actions) VALUES (?, ?, ?, ?, ?)`
	_, err = db.Exec(stmt, id, name, challengeRating, string(statsJSON), string(actionsJSON))
	return err
}

// migrateLocations migrates location data
func migrateLocations() error {
	locationsPath := filepath.Join("docs", "data", "content", "locations")

	// Clear existing locations
	if _, err := db.Exec("DELETE FROM locations"); err != nil {
		return fmt.Errorf("failed to clear locations table: %v", err)
	}

	count := 0

	// Walk through cities and environments
	subDirs := []string{"cities", "environments"}
	for _, subDir := range subDirs {
		dirPath := filepath.Join(locationsPath, subDir)
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && strings.HasSuffix(path, ".json") {
				locationType := subDir[:len(subDir)-1] // Remove 's' from cities/environments
				if err := migrateLocationFile(path, locationType); err != nil {
					log.Printf("Warning: failed to migrate location file %s: %v", path, err)
				} else {
					count++
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("Warning: failed to walk %s directory: %v", subDir, err)
		}
	}

	log.Printf("Migrated %d locations", count)
	return nil
}

// migrateLocationFile migrates a single location JSON file
func migrateLocationFile(filePath, locationType string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var location map[string]interface{}
	if err := json.Unmarshal(data, &location); err != nil {
		return err
	}

	// Extract base filename as ID
	id := strings.TrimSuffix(filepath.Base(filePath), ".json")

	name, _ := location["name"].(string)
	description, _ := location["description"].(string)

	// Serialize all properties as JSON
	propertiesJSON, _ := json.Marshal(location)
	connectionsJSON, _ := json.Marshal(location["connections"])

	stmt := `INSERT INTO locations (id, name, location_type, description, properties, connections)
			 VALUES (?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(stmt, id, name, locationType, description, string(propertiesJSON), string(connectionsJSON))
	return err
}

// migrateGenericJSON migrates a generic JSON file to a dynamically created table
func migrateGenericJSON(filePath, tableName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Create table if it doesn't exist
	createSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id VARCHAR PRIMARY KEY,
		data JSON,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`, tableName)

	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}

	// Clear existing data
	if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)); err != nil {
		return fmt.Errorf("failed to clear table %s: %v", tableName, err)
	}

	// Insert data
	id := strings.TrimSuffix(filepath.Base(filePath), ".json")
	stmt := fmt.Sprintf(`INSERT INTO %s (id, data) VALUES (?, ?)`, tableName)
	_, err = db.Exec(stmt, id, string(data))

	if err != nil {
		return fmt.Errorf("failed to insert into %s: %v", tableName, err)
	}

	return nil
}