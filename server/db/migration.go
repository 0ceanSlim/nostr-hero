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

	// Migrate character data
	if err := migrateCharacterData(); err != nil {
		return fmt.Errorf("failed to migrate character data: %v", err)
	}

	// Migrate items (includes packs as items now)
	if err := migrateItems(); err != nil {
		return fmt.Errorf("failed to migrate items: %v", err)
	}

	// Migrate spells
	if err := migrateSpells(); err != nil {
		return fmt.Errorf("failed to migrate spells: %v", err)
	}

	// Migrate content data (monsters, locations, NPCs)
	if err := migrateContentData(); err != nil {
		return fmt.Errorf("failed to migrate content data: %v", err)
	}

	log.Println("Migration completed successfully!")
	return nil
}

// migrateCharacterData migrates character-related JSON files
func migrateCharacterData() error {
	log.Println("Migrating character data...")

	characterDataPath := filepath.Join("game-data", "systems", "new-character")

	// Define the files we want to migrate for character data
	characterFiles := map[string]string{
		"base-hp.json":              "character_base_hp",
		"generation-weights.json":   "generation_weights",
		"introductions.json":        "introductions",
		"starting-gear.json":        "starting_gear",
		"starting-gold.json":        "starting_gold",
		"starting-locations.json":   "starting_locations",
		"starting-spells.json":      "starting_spells",
	}

	for filename, tableName := range characterFiles {
		filePath := filepath.Join(characterDataPath, filename)
		if err := migrateGenericJSON(filePath, tableName); err != nil {
			log.Printf("Warning: failed to migrate %s: %v", filename, err)
		}
	}

	return nil
}

// migrateItems migrates all item JSON files
func migrateItems() error {
	log.Println("Migrating items...")

	itemsPath := filepath.Join("game-data", "items")

	// Clear existing items
	if _, err := db.Exec("DELETE FROM items"); err != nil {
		return fmt.Errorf("failed to clear items table: %v", err)
	}

	count := 0
	err := filepath.WalkDir(itemsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			if err := migrateItemFile(path); err != nil {
				log.Printf("Warning: failed to migrate item file %s: %v", path, err)
			} else {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk items directory: %v", err)
	}

	log.Printf("Migrated %d items", count)
	return nil
}

// migrateItemFile migrates a single item JSON file
func migrateItemFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var item map[string]interface{}
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}

	// Extract base filename as ID
	id := strings.TrimSuffix(filepath.Base(filePath), ".json")

	// Convert item data to required fields
	name, _ := item["name"].(string)
	description, _ := item["description"].(string)
	itemType, _ := item["type"].(string)
	rarity, _ := item["rarity"].(string)

	// Extract tags as JSON
	tagsJSON, _ := json.Marshal(item["tags"])

	// Serialize all properties as JSON for the properties field
	propertiesJSON, _ := json.Marshal(item)

	stmt := `INSERT INTO items (id, name, description, item_type, properties, tags, rarity) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(stmt, id, name, description, itemType, string(propertiesJSON), string(tagsJSON), rarity)
	return err
}

// migrateSpells migrates all spell JSON files
func migrateSpells() error {
	log.Println("Migrating spells...")

	spellsPath := filepath.Join("game-data", "magic", "spells")

	// Clear existing spells
	if _, err := db.Exec("DELETE FROM spells"); err != nil {
		return fmt.Errorf("failed to clear spells table: %v", err)
	}

	count := 0
	err := filepath.WalkDir(spellsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			if err := migrateSpellFile(path); err != nil {
				log.Printf("Warning: failed to migrate spell file %s: %v", path, err)
			} else {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk spells directory: %v", err)
	}

	log.Printf("Migrated %d spells", count)
	return nil
}

// migrateSpellFile migrates a single spell JSON file
func migrateSpellFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var spell map[string]interface{}
	if err := json.Unmarshal(data, &spell); err != nil {
		return err
	}

	// Extract base filename as ID
	id := strings.TrimSuffix(filepath.Base(filePath), ".json")

	// Convert spell data to required fields
	name, _ := spell["name"].(string)
	description, _ := spell["description"].(string)
	level, _ := spell["level"].(float64)
	school, _ := spell["school"].(string)
	damage, _ := spell["damage"].(string)
	manaCostFloat, _ := spell["mana_cost"].(float64)
	manaCost := int(manaCostFloat)

	// Extract classes as JSON
	classesJSON, _ := json.Marshal(spell["classes"])

	// Serialize all properties as JSON for the properties field
	propertiesJSON, _ := json.Marshal(spell)

	stmt := `INSERT INTO spells (id, name, description, level, school, damage, mana_cost, classes, properties)
	         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(stmt, id, name, description, int(level), school, damage, manaCost, string(classesJSON), string(propertiesJSON))
	return err
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

	// Migrate NPCs
	if err := migrateNPCs(); err != nil {
		return fmt.Errorf("failed to migrate NPCs: %v", err)
	}

	return nil
}

// migrateMonsters migrates all monster JSON files
func migrateMonsters() error {
	monstersPath := filepath.Join("game-data", "monsters")

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
	locationsPath := filepath.Join("game-data", "locations")

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
	image, _ := location["image"].(string)
	music, _ := location["music"].(string)

	// Serialize all properties as JSON
	propertiesJSON, _ := json.Marshal(location)
	connectionsJSON, _ := json.Marshal(location["connections"])

	stmt := `INSERT INTO locations (id, name, location_type, description, image, music, properties, connections)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(stmt, id, name, locationType, description, image, music, string(propertiesJSON), string(connectionsJSON))
	return err
}

// migrateNPCs migrates all NPC JSON files from all location subdirectories
func migrateNPCs() error {
	npcsPath := filepath.Join("game-data", "npcs")

	// Clear existing NPCs
	if _, err := db.Exec("DELETE FROM npcs"); err != nil {
		return fmt.Errorf("failed to clear npcs table: %v", err)
	}

	count := 0

	// Walk through all subdirectories (kingdom, millhaven, etc.)
	err := filepath.WalkDir(npcsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			if err := migrateNPCFile(path); err != nil {
				log.Printf("Warning: failed to migrate NPC file %s: %v", path, err)
			} else {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk NPCs directory: %v", err)
	}

	log.Printf("Migrated %d NPCs", count)
	return nil
}

// migrateNPCFile migrates a single NPC JSON file
func migrateNPCFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var npc map[string]interface{}
	if err := json.Unmarshal(data, &npc); err != nil {
		return err
	}

	// Extract base filename as ID
	id := strings.TrimSuffix(filepath.Base(filePath), ".json")

	// Convert NPC data to required fields
	name, _ := npc["name"].(string)
	title, _ := npc["title"].(string)
	race, _ := npc["race"].(string)
	location, _ := npc["location"].(string)
	building, _ := npc["building"].(string)
	description, _ := npc["description"].(string)

	// Serialize all properties as JSON
	propertiesJSON, _ := json.Marshal(npc)

	stmt := `INSERT INTO npcs (id, name, title, race, location, building, description, properties)
	         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(stmt, id, name, title, race, location, building, description, string(propertiesJSON))
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