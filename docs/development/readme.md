# Nostr Hero - Development Guide

Complete guide for developing, building, and running Nostr Hero.

## Table of Contents

- [Quick Start](#quick-start)
- [Running the Game](#running-the-game)
- [Building](#building)
- [Development Tools](#development-tools)
- [Project Structure](#project-structure)
- [Game Systems](#game-systems)
- [Contributing](#contributing)

---

## Quick Start

### Prerequisites
- **Go 1.21+** (for backend)
- **Node.js** (optional, for frontend tooling)
- **Make** (optional, for build commands)

### Run Development Server
```bash
# Using Air (live reload - recommended)
air

# Or build and run manually
go build -o nostr-hero.exe
./nostr-hero.exe
```

The game will be available at `http://localhost:8080` (or port specified in `config.yml`)

---

## Running the Game

### Development Mode (Live Reload)

Using [Air](https://github.com/cosmtrek/air) for automatic rebuild on file changes:

```bash
air
```

**Configuration**: See [example.air.toml](./example.air.toml) for reference configuration.

Copy to repo root as `.air.toml` and customize as needed.

### Production Mode

```bash
# Build
go build -o nostr-hero.exe

# Run
./nostr-hero.exe
```

### Configuration

Edit `config.yml` to configure:
- Server port
- Nostr authentication settings
- PixelLab API key (for image generation tools)

---

## Building

### Backend (Go)

```bash
# Build for current platform
go build -o nostr-hero.exe

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o nostr-hero
GOOS=windows GOARCH=amd64 go build -o nostr-hero.exe
GOOS=darwin GOARCH=amd64 go build -o nostr-hero
```

### Database Migration

The game automatically migrates JSON data to SQLite on startup:

1. Server starts (`main.go`)
2. `MigrateFromJSON()` runs (`src/db/migration.go`)
3. Reads all JSON files from `docs/data/`
4. Populates `www/game.db`

**Note**: JSON files in `docs/data/` are the source of truth. The database is just a runtime cache.

---

## Development Tools

**Complete tools documentation**: [tools/readme.md](./tools/readme.md)

### Quick Reference

| Tool | Purpose | Quick Start |
|------|---------|-------------|
| **Item Editor** | Edit items + AI image generation | `cd docs/development/tools && make run-item-editor` |
| **PixelLab Generator** | Batch generate item images | `cd docs/development/tools && make pixellab-gen` |
| **Monster Manager** | Manage creature stats | `python docs/development/tools/monster_manager.py` |
| **World Map Visualizer** | Visualize locations | `python docs/development/tools/world_map_visualizer.py` |

**See [tools/readme.md](./tools/readme.md) for complete documentation, commands, and workflows.**

---

## Project Structure

```
nostr-hero/
├── main.go                 # Server entry point
├── config.yml             # Server configuration
│
├── src/                   # Go backend
│   ├── api/              # API handlers
│   ├── auth/             # Nostr authentication
│   ├── db/               # Database layer
│   ├── types/            # Data structures
│   ├── routes/           # HTTP routes
│   ├── utils/            # Utilities
│   └── functions/        # Game logic
│
├── www/                   # Frontend
│   ├── game.db           # SQLite database (generated)
│   ├── views/            # HTML templates
│   ├── scripts/          # JavaScript
│   │   ├── core/        # Session, auth
│   │   ├── systems/     # Game systems
│   │   ├── pages/       # UI logic
│   │   └── components/  # Reusable UI
│   └── res/             # Static resources
│       ├── css/
│       └── img/
│
├── docs/
│   ├── data/            # Game content (SOURCE OF TRUTH)
│   │   ├── character/   # Character generation
│   │   ├── equipment/   # Items database
│   │   └── content/     # Spells, monsters, locations
│   ├── development/     # Development docs (you are here)
│   └── draft/          # Archived planning docs
│
└── data/saves/          # Player save files (gitignored)
```

---

## Game Systems

### Core Systems

- **Character Generation** - Deterministic based on Nostr npub
- **Save System** - JSON-based with auto-save
- **Inventory System** - Drag-and-drop with equipment slots
- **Spell System** - D&D 5e-inspired spell slots
- **Location System** - Travel between cities/areas
- **Hunger/Fatigue** - Survival mechanics
- **Shop System** - Buy/sell items

### Data Flow

```
JSON Files (docs/data/)
    ↓ [Migration on startup]
SQLite (www/game.db)
    ↓ [API endpoints]
Go Backend (src/api/)
    ↓ [HTTP/JSON]
JavaScript Frontend (www/scripts/)
    ↓ [Save system]
Save Files (data/saves/{npub}/)
```

### Authentication

- Uses Nostr for identity (NIP-07 browser extensions)
- Grain client handles Nostr protocol
- No traditional passwords
- Session stored client-side

---

## Contributing

### Code Style

**Go Backend:**
- Use `go fmt`
- Follow standard Go conventions
- Add comments for exported functions

**JavaScript Frontend:**
- Vanilla JS (no frameworks)
- Event-driven architecture
- DOM-based state management (transitional)

### Adding New Items

1. Create JSON file: `docs/data/equipment/items/{item-id}.json`
2. Add pixel art: `www/res/img/items/{item-id}.png`
3. Restart server (migration loads new item)

See [Item Editor docs](./ITEM-EDITOR.md) for using the visual editor.

### Adding New Locations

1. Create JSON: `docs/data/content/locations/{location-id}.json`
2. Add background: `www/res/img/locations/{location-id}.png`
3. Update connections in related locations

### Modifying Game Data

**Always edit the JSON files in `docs/data/` - never edit the database directly.**

The database is regenerated from JSON on every server restart.

---

## Common Tasks

### Testing Changes

User typically runs and tests manually. The development server (via `air`) auto-reloads on changes.

### Database Issues

If database gets corrupted or out of sync:

```bash
# Delete the database
rm www/game.db

# Restart server (will regenerate from JSON)
air
```

### Adding API Endpoints

1. Define handler in `src/api/`
2. Register route in `main.go`
3. Update frontend to call new endpoint

### Debugging

- Backend logs go to console
- Frontend errors in browser DevTools console
- Check `data/saves/` for save file issues

---

## Architecture Principles

### Go-First Approach

- ✅ **All game logic in Go** (validation, calculations, state)
- ✅ **JavaScript only for DOM** (rendering, input, UI)
- ✅ **Backend is authoritative** (frontend can't cheat)

### Why Go-First?

1. **Security** - Backend validates all actions
2. **Consistency** - Single source of truth
3. **Testability** - Go easier to unit test
4. **Performance** - Compiled language for heavy calculations
5. **Future-proof** - Easier to add multiplayer, server AI, etc.

---

## Resources

- [Development Tools](./tools/readme.md) - Complete tools documentation
- [Item Editor Guide](./tools/item-editor.md) - Visual item editing with AI image generation
- [Example Air Config](./example.air.toml) - Live reload configuration
- [Project Architecture](../../CLAUDE.md) - Complete codebase documentation

---

## Getting Help

- Check existing documentation in `docs/development/`
- Review code comments in `src/` and `www/scripts/`
- Look at example items in `docs/data/equipment/items/`
- Check session notes in `docs/draft/` for historical context

---

**Last Updated**: 2025-11-03
