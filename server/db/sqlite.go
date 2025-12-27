package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDatabase initializes the database connection and validates it exists
// Used by the server - expects database to already be created by migration tool
func InitDatabase() error {
	// Ensure www directory exists
	dataDir := "./www"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create www directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, "game.db")

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found at %s - please run the migration tool first (cd game-data && go run migrate.go)", dbPath)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("✅ Connected to SQLite database at %s", dbPath)

	// Validate that required tables exist
	if err = validateDatabase(); err != nil {
		return fmt.Errorf("database validation failed: %v\nPlease run the migration tool (cd game-data && go run migrate.go)", err)
	}

	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// validateDatabase checks that all required tables exist
func validateDatabase() error {
	requiredTables := []string{
		"items",
		"spells",
		"monsters",
		"locations",
		"npcs",
		"generation_weights",
		"starting_gear",
		"starting_gold",
		"starting_locations",
		"starting_spells",
		"spell_slots_progression",
		"music_tracks",
	}

	for _, table := range requiredTables {
		var exists int
		err := db.QueryRow(`
			SELECT COUNT(*)
			FROM sqlite_master
			WHERE type='table' AND name=?
		`, table).Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check for table %s: %v", table, err)
		}

		if exists == 0 {
			return fmt.Errorf("required table '%s' not found", table)
		}
	}

	log.Println("✅ Database validation passed - all required tables exist")
	return nil
}