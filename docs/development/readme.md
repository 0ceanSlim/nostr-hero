# Nostr Hero - Development Guide

This guide provides a complete overview for developers looking to build, run, and contribute to the Nostr Hero project. It is designed to get you from a fresh clone to a running development server as quickly as possible.

## Table of Contents

- [Quick Start](#quick-start)
- [Running the Game](#running-the-game)
- [Building & Deployment](#building--deployment)
- [Development Tools](#development-tools)
- [Project Structure](#project-structure)
- [Game Systems](#game-systems)
- [Contributing](#contributing)

---

## Quick Start

### Option A: Automated Setup (Recommended)

For the fastest setup, use the included Makefile which will copy all necessary config files and install dependencies.

**Prerequisites:**
- **Go 1.21+** (for the backend server)
- **Node.js 18+** (for JavaScript development)
- **Air** (for live-reloading): `go install github.com/cosmtrek/air@latest`

**Setup:**
```bash
# Run from the docs/development directory
cd docs/development
make setup
```

This will:
1. Copy `example.config.yml` → `config.yml`
2. Copy `example.air.toml` → `.air.toml`
3. Copy `example.package.json` → `package.json`
4. Copy `example.vite.config.js` → `vite.config.js`
5. Copy `example.postcss.config.js` → `postcss.config.js`
6. Install npm dependencies

**Makefile Commands:**
```bash
# View all available commands
make help

# Copy configs only (without installing npm packages)
make setup-configs

# Install/update npm dependencies
make install-deps

# Reset to clean state (removes local configs, keeps examples)
make clean-configs
```

### Option B: Manual Setup

This approach is designed to get the game running with the least amount of setup, using the pre-compiled CSS file.

### 1. Prerequisites
- **Go 1.21+** (for the backend server)
- **Air** (for live-reloading): `go install github.com/cosmtrek/air@latest`

### 2. Configure the Server

Copy the example configuration file. This file is gitignored, so your local settings won't be committed.

```bash
cp example.config.yml config.yml
```

Edit `config.yml` to set your desired `port`. It is also highly recommended to enable `debug_mode` for development:

```yaml
server:
  port: 8585
  debug_mode: true # Recommended for development
```

### 3. Run the Server

That's it! The only process you need is the Go backend.

```bash
# This will automatically recompile and restart the server when Go files change.
air
```

The game will now be available at `http://localhost:8585` (or the port you specified).

**Having styling issues?** If the game looks unstyled, see the [Frontend (CSS)](#frontend-css) section for more options.

---

## Running the Game

There are three main components to the application: the Go backend, frontend JavaScript, and frontend CSS. While the Quick Start gets you running with pre-compiled assets, for active frontend development, you'll need to manage JavaScript bundling and CSS compilation yourself.

### Backend (Go)

The backend serves the game files and handles all game logic. The recommended way to run it is with [Air](https://github.com/cosmtrek/air).

```bash
# Starts the server and automatically restarts it on any .go file changes
# Also runs 'npm run build' to rebuild frontend assets before each Go build
air
```

Air watches the `server/` directory for Go file changes and:
1. Runs `npm run build` (Vite + Tailwind CSS v4)
2. Builds the Go binary from `server/main.go`
3. Restarts the server

Alternatively, you can build and run it manually:
```bash
# Build frontend assets first
npm run build

# Build Go binary from server/ directory
cd server
go build -o ../nostr-hero main.go
cd ..

# Run the binary
./nostr-hero
```

### Frontend (CSS)

The project uses **Tailwind CSS v4** integrated with Vite via the `@tailwindcss/vite` plugin. CSS is automatically compiled as part of the Vite build process.

#### CSS Development

**Option 1: Use Air (recommended for backend-focused development)**
```bash
# Air automatically rebuilds CSS on every Go file change
air
```

**Option 2: Use Vite dev server (recommended for frontend-focused development)**
```bash
# Instant hot module replacement for CSS changes
npm run dev
```

**CSS Configuration:**
- Source: `src/styles/main.css` (imports Tailwind CSS v4)
- Output: `www/dist/main.css` (single CSS file for all pages)
- Config: `postcss.config.js` (Tailwind + Autoprefixer)

The CSS file is configured in `src/styles/main.css`:
```css
@import "tailwindcss";

/* Tailwind content sources */
@source "../../www/views/**/*.html";
@source "../**/*.{js,jsx,ts,tsx}";
```

Tailwind v4 uses CSS-based configuration instead of JavaScript config files.

### Frontend (JavaScript)

The project uses modern ES6 modules with Vite for bundling and development.

#### Development Modes

**Option 1: Air (single terminal, rebuilds on Go changes)**
```bash
# Watches Go files, rebuilds frontend + backend on change
air
```
- Rebuilds both frontend and backend when you edit Go files
- Slower but simpler (single command)
- Access game at your Go server port (from config.yml)

**Option 2: Vite dev server (recommended for frontend development)**
```bash
# In a separate terminal from your Go server
npm run dev
```
- Instant hot module replacement (HMR) for JS/CSS changes
- Faster iteration for frontend development
- Proxies API calls to your Go backend
- Access game at http://localhost:5173 (Vite dev server)

#### Production Build

```bash
# Build optimized bundles for production
npm run build
```

**What gets built:**
- All JavaScript entry points from `src/entries/`
- Single CSS file (`main.css`) from `src/styles/main.css`
- Output directory: `www/dist/`

**Entry points:**
- `index.js` - Home page
- `game.js` - Main game interface
- `gameIntro.js` - Game introduction
- `newGame.js` - Character creation
- `settings.js` - Settings page
- `discover.js` - Discovery page
- `saves.js` - Save management

#### Watch Mode (for manual testing)

```bash
# Continuously rebuild on file changes
npm run build:watch
```

Useful if you want to run the Go server directly and have Vite rebuild in the background.

### Debug Mode

You can enable debug mode by setting `debug_mode: true` in your `config.yml`. When enabled, it provides two key features:

1.  **Debug API Endpoints**: Activates `/api/debug/sessions` and `/api/debug/state` to inspect live game sessions from the backend.
2.  **In-Game Debug Console**: Adds a "Debug Console" to the bug report modal in the game UI, allowing you to view the live in-memory game state for your character. This is very useful for debugging inventory and state issues.

---

## Building & Deployment

### Building Binaries

To create a production-ready binary of the application, use the standard `go build` command.

```bash
# Build for your current platform
go build -o nostr-hero.exe

# Example: Build for Linux
GOOS=linux GOARCH=amd64 go build -o nostr-hero
```

### Database Migration

The game automatically migrates all game data from the JSON files in `docs/data/` into a single SQLite database (`www/game.db`) on every startup. The JSON files are the source of truth; the database is just a performant runtime cache.

### Deployment Strategy

This project uses a multi-server environment (Test, Official, Modded) with automated deployment triggers. For a complete overview of the server architecture, versioning strategy, and build automation, see the detailed deployment guide:

- **[Deployment & Versioning Strategy](./deployment/readme.md)**

---

## Development Tools

This project includes several custom tools to aid in development.

**Complete tools documentation**: [tools/readme.md](./tools/readme.md)

| Tool                 | Purpose                          | Quick Start                                               |
| -------------------- | -------------------------------- | --------------------------------------------------------- |
| **Item Editor**      | Edit items + AI image generation | `cd docs/development/tools && make run-item-editor`         |
| **PixelLab Generator** | Batch generate item images       | `cd docs/development/tools && make pixellab-gen`          |
| **Monster Manager**  | Manage creature stats            | `python docs/development/tools/monster_manager.py`        |
| **World Map Visualizer** | Visualize locations              | `python docs/development/tools/world_map_visualizer.py` |

---

## Project Structure

```
nostr-hero/
├── config.yml             # Local server configuration (gitignored)
│
├── server/                # Go backend source code
│   ├── main.go           # Server entry point
│   ├── go.mod            # Go module definition
│   ├── api/              # API handlers for game logic
│   ├── db/               # Database layer (SQLite)
│   ├── routes/           # HTTP route definitions
│   ├── auth/             # Nostr authentication
│   ├── types/            # Data structures
│   └── utils/            # Utility functions
│
├── src/                   # Frontend JavaScript source code
│   ├── entries/          # Entry points (one per page)
│   ├── lib/              # Core libraries
│   ├── data/             # Data layer
│   ├── state/            # State management
│   ├── logic/            # Game logic
│   ├── systems/          # Complex systems
│   ├── ui/               # UI rendering
│   ├── components/       # Reusable components
│   ├── config/           # Configuration
│   └── styles/           # CSS source (Tailwind v4)
│
├── www/                   # All frontend assets served to the browser
│   ├── game.db           # SQLite database (generated on startup)
│   ├── dist/             # Built JavaScript and CSS (gitignored)
│   ├── views/            # Go HTML templates
│   └── res/              # Static resources (images, fonts)
│
├── game-data/             # The game's raw data (SOURCE OF TRUTH)
│   ├── items/            # 200+ individual item files
│   ├── magic/            # Spells and spell slots
│   ├── monsters/         # Creature stat blocks
│   ├── locations/        # World map and locations
│   ├── npcs/             # NPCs by location
│   └── systems/          # System configurations
│
├── docs/
│   ├── development/      # Developer documentation (you are here)
│   │   ├── examples/    # Example config files
│   │   ├── deployment/  # Deployment scripts & server setup
│   │   └── tools/       # Custom developer tools
│   └── draft/           # Archived planning and design documents
│
└── data/saves/           # Local player save files (gitignored)
```

---

## Contributing

### Code Style

**Go Backend:**
- Run `go fmt` before committing.
- Follow standard Go conventions and idioms.

**JavaScript Frontend:**
- Vanilla JS (no frameworks).
- Follow an event-driven architecture where possible.

### Modifying Game Data

**Always edit the JSON files in `game-data/` - never edit the database file (`www/game.db`) directly.** The database is completely rebuilt from the JSON files every time the server starts.

### Common Troubleshooting

- **Styling is broken:** Run `npm run build` to regenerate the CSS file. Make sure you have run `npm install` first.
- **JavaScript not loading:** Check that `www/dist/` contains the built files. Run `npm run build` if needed.
- **Database is corrupted:** Delete `www/game.db` and restart the server. It will be rebuilt automatically from JSON files.
- **Go build fails:** Make sure you're building from the `server/` directory: `cd server && go build -o ../nostr-hero main.go`
- **npm install fails:** Make sure you have Node.js 18+ installed: `node --version`

---

## Getting Help

- Check existing documentation in `docs/development/`.
- Review code comments in `server/` and `src/`.
- Look at example items in `game-data/items/`.
- Check session notes in `docs/draft/` for historical context.
- See deployment documentation in `docs/development/deployment/`.

---

**Last Updated**: 2025-12-18