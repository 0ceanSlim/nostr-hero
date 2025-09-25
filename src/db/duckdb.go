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

// InitDatabase initializes the DuckDB database connection
func InitDatabase() error {
	// Ensure www directory exists
	dataDir := "./www"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create www directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, "game.db")

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Connected to SQLite database at %s", dbPath)

	// Create tables
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
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

// createTables creates all the necessary tables for static game data
func createTables() error {
	tables := []string{
		// Items table for weapons, armor, consumables, etc.
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			item_type TEXT NOT NULL,
			properties TEXT,
			tags TEXT,
			rarity TEXT DEFAULT 'common',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Spells table
		`CREATE TABLE IF NOT EXISTS spells (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			level INTEGER NOT NULL,
			school TEXT NOT NULL,
			damage TEXT,
			mana_cost INTEGER,
			classes TEXT,
			properties TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Monsters table
		`CREATE TABLE IF NOT EXISTS monsters (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			challenge_rating REAL,
			stats TEXT,
			actions TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Locations table
		`CREATE TABLE IF NOT EXISTS locations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			location_type TEXT,
			description TEXT,
			properties TEXT,
			connections TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Character classes table for deterministic creation
		`CREATE TABLE IF NOT EXISTS character_classes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			hit_die INTEGER,
			spell_progression TEXT,
			starting_equipment TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Races table for character generation
		`CREATE TABLE IF NOT EXISTS races (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			ability_modifiers TEXT,
			traits TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Equipment packs table
		`CREATE TABLE IF NOT EXISTS equipment_packs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			items TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	log.Println("Database tables created successfully")
	return nil
}