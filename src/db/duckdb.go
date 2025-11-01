package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	// Run migrations first
	if err = runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

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

// runMigrations runs database schema migrations
func runMigrations() error {
	// Check if locations table exists and has the image/music columns
	var columnCount int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM pragma_table_info('locations')
		WHERE name IN ('image', 'music')
	`).Scan(&columnCount)

	if err == nil && columnCount < 2 {
		log.Println("Running migration: Adding image and music columns to locations table")

		// Add image column if it doesn't exist
		_, err = db.Exec("ALTER TABLE locations ADD COLUMN image TEXT")
		if err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return fmt.Errorf("failed to add image column: %v", err)
		}

		// Add music column if it doesn't exist
		_, err = db.Exec("ALTER TABLE locations ADD COLUMN music TEXT")
		if err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return fmt.Errorf("failed to add music column: %v", err)
		}

		log.Println("Successfully added image and music columns to locations table")
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
			image TEXT,
			music TEXT,
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

		// NPCs table
		`CREATE TABLE IF NOT EXISTS npcs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			title TEXT,
			race TEXT,
			location TEXT,
			building TEXT,
			description TEXT,
			properties TEXT,
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