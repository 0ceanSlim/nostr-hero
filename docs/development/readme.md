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

There are two main components to the application: the Go backend and the frontend CSS. While the Quick Start gets you running with a pre-compiled `output.css`, for active frontend development, you'll need to manage the CSS compilation yourself.

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

The project uses Tailwind CSS. There are two ways to handle CSS during development.

#### Option 1: Use the Tailwind Play CDN (Easiest)

For a "zero-install" frontend experience, you can use the Tailwind Play CDN. This is a script that compiles the necessary styles directly in your browser. This is great for backend work or quick UI tweaks without needing Node.js.

To use it, edit `www/views/templates/layout.html` and replace the `output.css` link with the CDN script:

**Replace this:**
```html
<link href="/res/style/output.css?v=20250109-23" rel="stylesheet" />
```

**With this:**
```html
<script src="https://cdn.tailwindcss.com"></script>
```

#### Option 2: Use the Tailwind CLI (for Active Styling)

If you are actively working on the game's styles, you should use the Tailwind CLI to watch for changes and rebuild your CSS file automatically. This requires Node.js.

In a separate terminal from your `air` process, run the following command:
```bash
npx tailwindcss -i ./www/res/style/input.css -o ./www/res/style/output.css --watch
```
This will watch for changes in the HTML and JS files and keep `output.css` up to date. For more information, see the [styling readme](./www/res/style/readme.md).

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