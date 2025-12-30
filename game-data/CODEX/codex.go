package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"nostr-hero/codex/config"
	itemeditor "nostr-hero/codex/item-editor"
	"nostr-hero/codex/migration"
	"nostr-hero/codex/pixellab"
	"nostr-hero/codex/validation"

	"github.com/gorilla/mux"
)

var editor *itemeditor.Editor

func main() {
	// Command-line flags
	migrateFlag := flag.Bool("migrate", false, "Run database migration and exit")
	flag.Parse()

	// Handle migration flag
	if *migrateFlag {
		fmt.Println("🔄 Running database migration...")
		dbPath := "../../www/game.db"

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

	if err := editor.LoadItems(); err != nil {
		log.Fatal(err)
	}

	// Try to load config for PixelLab
	cfg, err := config.Load()
	if err != nil {
		log.Printf("⚠️ Error loading config: %v - continuing without image generation", err)
	} else if cfg != nil {
		editor.PixelLabClient = pixellab.NewClient(cfg.PixelLab.APIKey)
		log.Printf("✅ PixelLab client initialized")
	}

	r := mux.NewRouter()

	// Home page
	r.HandleFunc("/", handleHome).Methods("GET")

	// Item editor routes
	r.HandleFunc("/tools/item-editor", editor.HandleItemEditor).Methods("GET")
	r.HandleFunc("/api/items", editor.HandleGetItems).Methods("GET")
	r.HandleFunc("/api/items/{filename}", editor.HandleGetItem).Methods("GET")
	r.HandleFunc("/api/items/{filename}", editor.HandleSaveItem).Methods("PUT")
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

	// Static files
	r.PathPrefix("/www/").Handler(http.StripPrefix("/www/", http.FileServer(http.Dir("../../www/"))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("🎯 CODEX - Content Organization & Data Entry eXperience")
	fmt.Println("🔧 Starting on http://localhost:8080")
	fmt.Println("Opening browser...")

	// Open browser automatically
	go func() {
		time.Sleep(500 * time.Millisecond)
		url := "http://localhost:8080"
		var err error

		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", url).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		case "darwin":
			err = exec.Command("open", url).Start()
		}

		if err != nil {
			fmt.Printf("Please open your browser to: %s\n", url)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", r))
}

// Home page handler
func handleHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/home.html")
}

// Database migration handlers
var migrationStatus migration.Status
var migrationRunning bool

func handleDatabaseMigration(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/database-migration.html")
}

func handleMigrateStart(w http.ResponseWriter, r *http.Request) {
	if migrationRunning {
		http.Error(w, "Migration already in progress", http.StatusConflict)
		return
	}

	migrationRunning = true
	go func() {
		defer func() { migrationRunning = false }()

		dbPath := "../../../www/game.db"
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
	http.ServeFile(w, r, "templates/validation.html")
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
