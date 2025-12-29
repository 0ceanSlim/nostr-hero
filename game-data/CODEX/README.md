# CODEX - Content Organization & Data Entry eXperience

**The official all-in-one tool for managing Nostr Hero game data.**

CODEX is a comprehensive web-based GUI tool for editing, validating, and managing all game content for Nostr Hero. It provides a terminal-themed interface for developers and contributors to work with game data files.

## Features

### 🎯 Current Features (v1.0)

#### Item Editor
- **Visual Item Browser** - Browse all game items with search and filters
- **Full CRUD Operations** - Create, read, update, and delete items
- **Global ID Refactoring** - Safely rename item IDs across all game data
- **Validation** - Real-time validation of item data
- **Tag Management** - Add and manage item tags
- **Field Management** - Add/remove custom fields dynamically

#### Image Generation (PixelLab Integration)
- **AI-Generated Pixel Art** - Generate item sprites using PixelLab API
- **Multiple Models** - Choose between Bitforge and Pixflux
- **Image History** - Keeps history of all generated images
- **Preview & Accept** - Review generated images before accepting

#### Database Migration
- **CLI Migration** - Run migrations from command line with `--migrate` flag
- **GUI Migration** - Visual interface for migrating JSON to SQLite with progress tracking
- **Status Monitoring** - Real-time migration progress updates

#### Data Validation
- **Comprehensive Checks** - Validate all game data for errors and inconsistencies
- **Categorized Issues** - Errors, warnings, and info messages
- **Detailed Reports** - File-by-file breakdown of validation issues

### 🚀 Planned Features

- **Spell Editor** - Edit D&D 5e spells
- **Monster Editor** - Manage creature stat blocks
- **Location Editor** - Edit world map and locations
- **NPC Editor** - Manage NPC data
- **Bulk Operations** - Edit multiple items at once
- **Export/Import** - Backup and restore game data

## Installation

### Prerequisites

- Go 1.21 or later
- PixelLab API key (optional, for image generation)

### Setup

1. Navigate to the CODEX directory:
   ```bash
   cd game-data/CODEX
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. (Optional) Configure PixelLab API:
   - Add your API key to `config.yml` in the project root:
   ```yaml
   pixellab:
     api_key: "your-api-key-here"
   ```

## Usage

### Running CODEX

From the `game-data/CODEX` directory:

**Start the GUI (default)**:
```bash
go run .
# or
./codex.exe
```

The tool will:
1. Start a web server on `http://localhost:8080`
2. Automatically open your browser
3. Load all items from `game-data/items/`

**Run database migration from command line**:
```bash
./codex.exe --migrate
```

This will migrate all JSON game data to `www/game.db` and exit. Useful for CI/CD pipelines or automated builds.

### Interface Overview

#### Left Sidebar
- **Search Box** - Filter items by name, ID, or description
- **Type Filter** - Filter by item type (weapon, armor, etc.)
- **Tag Filter** - Filter by tags
- **Validate All** - Run validation on all items
- **Refresh** - Reload items from disk

#### Main Panel
- **Item Editor** - Edit all item properties
- **Image Manager** - Preview, generate, and accept item images
- **Tag Manager** - Add/remove tags
- **Notes Manager** - Add internal development notes
- **Save Changes** - Write changes to JSON file
- **Refactor ID** - Globally rename an item ID

### Key Workflows

#### Editing an Item

1. Click an item in the sidebar
2. Modify fields in the main panel
3. Click "💾 Save Changes"

#### Refactoring an Item ID

1. Select the item to rename
2. Click "🔄 Refactor ID"
3. Enter the new ID
4. Click "Preview Changes" to see what will update
5. Review the preview (shows all references)
6. Click "✓ Apply Refactor" to execute

This will automatically update:
- The item filename
- The item's `id` field
- All references in `starting-gear.json`
- All references in pack contents
- Any other item references

#### Generating Item Images

1. Select an item
2. Choose a model (Bitforge or Pixflux)
3. Click "🎨 Generate Image"
4. Review the generated image
5. Click "✓ Use This Image" to accept, or "✗ Discard" to try again

Images are saved to:
- History: `www/res/img/items/_history/{item-id}/{timestamp}_{model}.png`
- Main: `www/res/img/items/{item-id}.png` (after accepting)

## Architecture

### Directory Structure

```
game-data/CODEX/
├── codex.go                      # Main entry point
├── go.mod, go.sum                # Go dependencies
├── README.md                     # This file
│
├── templates/                    # HTML templates
│   ├── home.html
│   ├── item-editor.html
│   ├── database-migration.html
│   └── validation.html
│
├── static/                       # Static assets
│   ├── css/                      # Stylesheets
│   │   ├── codex.css
│   │   ├── tool.css
│   │   └── validation.css
│   └── js/                       # JavaScript modules
│       ├── item-editor.js
│       ├── database-migration.js
│       └── validation.js
│
├── item-editor/                  # Item editor package
│   ├── editor.go
│   ├── handlers.go
│   └── refactor.go
│
├── config/                       # Configuration
│   └── config.go
│
├── pixellab/                     # Image generation
│   └── client.go
│
├── migration/                    # Database migration
│   └── migration.go
│
└── validation/                   # Data validation
    └── validation.go
```

### Data Flow

```
game-data/items/*.json  <--  CODEX  -->  www/res/img/items/*.png
                              |
                              v
              game-data/systems/new-character/starting-gear.json
```

CODEX operates directly on the JSON files in `game-data/`, which are the source of truth. Changes are written immediately to disk.

### Path Resolution

CODEX runs from `game-data/CODEX/`, so paths are relative:
- Items: `../items/`
- Starting Gear: `../systems/new-character/starting-gear.json`
- Images: `../../www/res/img/items/`
- Config: `../../config.yml`

## API Endpoints

CODEX provides a REST API for the frontend:

### Item Management
- `GET /api/items` - List all items
- `GET /api/items/{filename}` - Get specific item
- `PUT /api/items/{filename}` - Update item
- `GET /api/validate` - Validate all items
- `GET /api/types` - Get all item types
- `GET /api/tags` - Get all tags

### Refactoring
- `POST /api/refactor/preview` - Preview ID refactor
- `POST /api/refactor/apply` - Apply ID refactor

### Image Generation
- `GET /api/balance` - Get PixelLab account balance
- `POST /api/items/{filename}/generate-image` - Generate image
- `GET /api/items/{filename}/image` - Get image info
- `POST /api/items/{filename}/accept-image` - Accept generated image

## Development

### Adding New Editors

To add a new editor (e.g., for spells or monsters):

1. Add data structures to `main.go`
2. Add load/save functions
3. Add API endpoints
4. Add frontend UI in the HTML template
5. Wire up JavaScript event handlers

### Extending Validation

Add validation rules in `handleValidate()`:

```go
func (e *ItemEditor) handleValidate(w http.ResponseWriter, r *http.Request) {
	issues := []string{}

	for filename, item := range e.items {
		// Add your validation checks here
		if item.Price < 0 {
			issues = append(issues, fmt.Sprintf("%s: negative price", filename))
		}
	}

	// ...
}
```

## Troubleshooting

### Items not loading
- Ensure you're running from `game-data/CODEX/`
- Check that `../items/` directory exists
- Check console for JSON parsing errors

### Image generation fails
- Verify `config.yml` exists in project root
- Check PixelLab API key is valid
- Check account balance with "🔄 Refresh Balance"

### Refactor not updating all references
- Ensure starting-gear.json path is correct
- Check console for error messages
- Verify the item ID format matches

## Contributing

### Adding Items
1. Use CODEX to create new items
2. Generate pixel art images
3. Test in-game
4. Commit both JSON and PNG files

### Reporting Issues
- Check validation output first
- Include the item JSON that causes issues
- Note any console errors

## Credits

CODEX is part of the Nostr Hero project.

- **Terminal Theme**: Dracula-inspired color scheme
- **Image Generation**: PixelLab API
- **Web Framework**: Gorilla Mux

## License

Same as Nostr Hero project.

---

**Last Updated**: 2025-12-28
**Version**: 1.0.0
**Maintainer**: Nostr Hero Development Team
