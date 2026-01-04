package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"nostr-hero/codex/config"
	itemeditor "nostr-hero/codex/item-editor"
	"nostr-hero/codex/migration"
	"nostr-hero/codex/pixellab"
	"nostr-hero/codex/staging"
	"nostr-hero/codex/validation"

	"github.com/gorilla/mux"
)

var editor *itemeditor.Editor

func main() {
	// Command-line flags
	migrateFlag := flag.Bool("migrate", false, "Run database migration and exit")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	if cfg == nil {
		log.Fatal("❌ codex-config.yml not found")
	}

	// Handle migration flag
	if *migrateFlag {
		fmt.Println("🔄 Running database migration...")
		dbPath := "./www/game.db"

		err := migration.Migrate(dbPath, func(status migration.Status) {
			if status.Progress > 0 {
				fmt.Printf("  %s (%d/%d)\n", status.Message, status.Progress, status.Total)
			} else {
				fmt.Printf("  %s\n", status.Message)
			}
		})

		if err != nil {
			fmt.Printf("❌ Migration failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Migration completed successfully!")
		os.Exit(0)
	}

	// Initialize item editor
	editor = itemeditor.New()
	editor.Config = cfg

	if err := editor.LoadItems(); err != nil {
		log.Fatal(err)
	}

	// Initialize PixelLab if API key is configured
	if cfg.PixelLab.APIKey != "" {
		editor.PixelLabClient = pixellab.NewClient(cfg.PixelLab.APIKey)
		log.Printf("✅ PixelLab client initialized")
	}

	// Initialize staging system
	staging.SetConfig(cfg)
	staging.Manager.StartCleanupRoutine()
	log.Printf("✅ Staging system initialized")

	r := mux.NewRouter()

	// Home page
	r.HandleFunc("/", handleHome).Methods("GET")

	// Item editor routes
	r.HandleFunc("/tools/item-editor", editor.HandleItemEditor).Methods("GET")
	r.HandleFunc("/api/items", editor.HandleGetItems).Methods("GET")
	r.HandleFunc("/api/items/{filename}", editor.HandleGetItem).Methods("GET")
	r.HandleFunc("/api/items/{filename}", editor.HandleSaveItem).Methods("PUT")
	r.HandleFunc("/api/items/{filename}", editor.HandleDeleteItem).Methods("DELETE")
	r.HandleFunc("/api/validate", editor.HandleValidate).Methods("GET")
	r.HandleFunc("/api/types", editor.HandleGetTypes).Methods("GET")
	r.HandleFunc("/api/tags", editor.HandleGetTags).Methods("GET")
	r.HandleFunc("/api/refactor/preview", editor.HandleRefactorPreview).Methods("POST")
	r.HandleFunc("/api/refactor/apply", editor.HandleRefactorApply).Methods("POST")
	r.HandleFunc("/api/balance", editor.HandleGetBalance).Methods("GET")
	r.HandleFunc("/api/items/{filename}/generate-image", editor.HandleGenerateImage).Methods("POST")
	r.HandleFunc("/api/items/{filename}/image", editor.HandleGetImage).Methods("GET")
	r.HandleFunc("/api/items/{filename}/accept-image", editor.HandleAcceptImage).Methods("POST")

	// Database migration routes
	r.HandleFunc("/tools/database-migration", handleDatabaseMigration).Methods("GET")
	r.HandleFunc("/api/migrate/start", handleMigrateStart).Methods("POST")
	r.HandleFunc("/api/migrate/status", handleMigrateStatus).Methods("GET")

	// Validation routes
	r.HandleFunc("/tools/validation", handleValidationTool).Methods("GET")
	r.HandleFunc("/api/validation/run", handleValidationRun).Methods("POST")
	r.HandleFunc("/api/validation/cleanup", handleCleanupRun).Methods("POST")

	// Staging routes
	r.HandleFunc("/api/staging/init", staging.HandleStagingInit).Methods("POST")
	r.HandleFunc("/api/staging/changes", staging.HandleGetStagingChanges).Methods("GET")
	r.HandleFunc("/api/staging/submit", staging.HandleStagingSubmit).Methods("POST")
	r.HandleFunc("/api/staging/clear", staging.HandleStagingClear).Methods("DELETE")
	r.HandleFunc("/api/staging/mode", staging.HandleGetMode).Methods("GET")

	// Static files - serve from main www directory (running from root)
	r.PathPrefix("/www/").Handler(http.StripPrefix("/www/", http.FileServer(http.Dir("www"))))
	r.PathPrefix("/game-res/").Handler(http.StripPrefix("/game-res/", http.FileServer(http.Dir("www/res"))))

	// CODEX built assets - served from www/dist/codex/
	r.PathPrefix("/dist/codex/").Handler(http.StripPrefix("/dist/codex/", http.FileServer(http.Dir("www/dist/codex"))))

	// CODEX resources (if any static files in res/)
	r.PathPrefix("/codex-res/").Handler(http.StripPrefix("/codex-res/", http.FileServer(http.Dir("game-data/CODEX/res"))))

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Println("🎯 CODEX - Content Organization & Data Entry eXperience")
	fmt.Printf("🚀 Server starting on http://localhost%s\n", port)

	log.Fatal(http.ListenAndServe(port, r))
}

// Home page handler
func handleHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "game-data/CODEX/html/home-new.html")
}

// Database migration handlers
var migrationStatus migration.Status
var migrationRunning bool

func handleDatabaseMigration(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "game-data/CODEX/html/database-migration.html")
}

func handleMigrateStart(w http.ResponseWriter, r *http.Request) {
	if migrationRunning {
		http.Error(w, "Migration already in progress", http.StatusConflict)
		return
	}

	migrationRunning = true
	go func() {
		defer func() { migrationRunning = false }()

		dbPath := "./www/game.db"
		err := migration.Migrate(dbPath, func(status migration.Status) {
			migrationStatus = status
		})

		if err != nil {
			migrationStatus = migration.Status{
				Step:  "error",
				Error: err.Error(),
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func handleMigrateStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(migrationStatus)
}

// Validation handlers
func handleValidationTool(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "game-data/CODEX/html/validation.html")
}

func handleValidationRun(w http.ResponseWriter, r *http.Request) {
	result, err := validation.ValidateAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleCleanupRun(w http.ResponseWriter, r *http.Request) {
	// Check for dry_run parameter
	dryRun := r.URL.Query().Get("dry_run") == "true"

	result, err := validation.CleanupAllItems(dryRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
