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
5. Install npm dependencies

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
# Starts the server and automatically restarts it on any .go file changes.
air
```

Alternatively, you can build and run it manually:
```bash
go build -o nostr-hero.exe
./nostr-hero.exe
```

### Frontend (CSS)

The project uses Tailwind CSS, which is now integrated with the npm build system.

#### CSS Development (Watch Mode)

To automatically recompile CSS when you make styling changes:

```bash
# Watch and recompile CSS on changes
npm run dev:css
```

**Or run both JavaScript and CSS watchers together:**
```bash
# Run Vite + Tailwind CSS together (recommended for full-stack development)
npm run dev:full
```

This will run both the Vite dev server and Tailwind CSS watcher concurrently in a single terminal.

#### CSS Production Build

CSS is automatically built when you run the production build:

```bash
# Builds both JavaScript and CSS
npm run build
```

To build only CSS:
```bash
npm run build:css
```

#### Option 2: Use the Tailwind Play CDN (Quick Backend Development)

If you're only working on backend Go code and don't want to run any Node.js processes, you can use the Tailwind Play CDN. Edit `www/views/templates/layout.html`:

**Replace this:**
```html
<link href="/res/style/output.css?v=20250109-23" rel="stylesheet" />
```

**With this:**
```html
<script src="https://cdn.tailwindcss.com"></script>
```

For more information, see the [styling readme](./www/res/style/readme.md).

### Frontend (JavaScript)

The project uses modern ES6 modules with Vite for bundling and development. There are two modes of operation:

#### Development Mode (with Hot Module Replacement)

For active JavaScript development, use Vite's dev server which provides instant hot module replacement (HMR):

```bash
# In a separate terminal from your 'air' process
npm run dev
```

**What this does:**
- Starts Vite dev server on http://localhost:5173
- Proxies API calls to your Go backend (port from config.yml)
- Provides instant updates when you edit JavaScript files
- Includes source maps for easy debugging
- Shows all debug logs in console

**Note:** When using Vite dev server, you'll access the game at http://localhost:5173 (Vite), not your Go server port.

#### Production Build

To build optimized production bundles:

```bash
# Build minified, optimized bundles
npm run build

# Preview production build
npm run preview
```

**Production build benefits:**
- Minified and tree-shaken (71% smaller than development)
- Debug logs removed (only errors shown)
- Hashed filenames for cache busting
- Code splitting for faster loading

**Build outputs:**
- Development: `www/dist/dev/` (readable, with source maps)
- Production: `www/dist/prod/` (minified, with hashes)

**Clean build directory:**
```bash
npm run clean
```

For more details on the JavaScript architecture, see [JAVASCRIPT-REFACTOR.md](../../JAVASCRIPT-REFACTOR.md).

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
├── main.go                 # Server entry point
├── config.yml             # Local server configuration (gitignored)
│
├── src/                   # Go backend source code
│   ├── api/              # API handlers for game logic
│   ├── db/               # Database layer (SQLite)
│   ├── routes/           # HTTP route definitions
│   └── ...
│
├── www/                   # All frontend assets served to the browser
│   ├── game.db           # SQLite database (generated on startup)
│   ├── views/            # Go HTML templates
│   ├── scripts/          # Frontend JavaScript files
│   └── res/              # Static resources (CSS, images, fonts)
│
├── docs/
│   ├── data/            # The game's raw data (SOURCE OF TRUTH)
│   ├── development/     # Developer documentation (you are here)
│   │   ├── deployment/  # Deployment scripts & server setup docs
│   │   └── tools/       # Custom developer tool documentation
│   └── draft/           # Archived planning and design documents
│
└── data/saves/          # Local player save files (gitignored)
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

**Always edit the JSON files in `docs/data/` - never edit the database file (`www/game.db`) directly.** The database is completely rebuilt from the JSON files every time the server starts.

### Common Troubleshooting

- **Styling is broken:** Your `www/res/style/output.css` file may be missing or out of date. See the [Frontend (CSS)](#frontend-css) section above for instructions on how to regenerate it, or consult the [styling readme](./www/res/style/readme.md).
- **Database is corrupted:** If you encounter strange data-related errors, delete `www/game.db` and restart the server. It will be rebuilt automatically.

---

## Getting Help

- Check existing documentation in `docs/development/`.
- Review code comments in `src/` and `www/scripts/`.
- Look at example items in `docs/data/equipment/items/`.
- Check session notes in `docs/draft/` for historical context.

---

**Last Updated**: 2025-12-07