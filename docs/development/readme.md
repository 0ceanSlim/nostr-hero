# Pubkey Quest - Development Guide

This guide covers everything you need to get the game server running locally and start contributing.

## Prerequisites

- **Go 1.24+** - Backend server and game logic
- **Node.js 18+** - Frontend build tooling (Vite, Tailwind CSS v4)
- **Air** - Live-reload for Go development: `go install github.com/air-verse/air@latest`
- **swag** - Swagger doc generation: `go install github.com/swaggo/swag/cmd/swag@latest`

## Quick Start

```bash
# 1. Clone the repo
git clone <repo-url> && cd pubkey-quest

# 2. Copy example config and edit it
cp docs/development/examples/example.config.yml config.yml
# Edit config.yml - set your port and enable debug_mode

# 3. Install frontend dependencies
npm install

# 4. Build codex and run database migration
go build -o codex ./cmd/codex    # Linux/macOS
go build -o codex.exe ./cmd/codex  # Windows
./codex -migrate                 # Linux/macOS
codex.exe -migrate               # Windows

# 5. Start the dev server
air
```

The game will be available at `http://localhost:<your-port>`.

That's it — Air handles building the frontend (npm), compiling the Go server, and restarting on changes.

## Running the Server

### With Air (recommended for development)

```bash
air
```

The `.air.toml` config at the project root is committed to the repo so everyone gets the same dev experience. It watches `cmd/` and `types/` for Go file changes and on each rebuild:

1. Runs `npm run build` (Vite + Tailwind CSS v4)
2. Compiles the server from `./cmd/server`
3. Restarts the server binary

Air only triggers on Go file changes. If you're only editing frontend files, use the Vite dev server (below) for faster iteration.

### Manual build and run

```bash
npm run build
go build -o pubkey-quest.exe ./cmd/server
./pubkey-quest.exe
```

## Frontend Development

The frontend uses vanilla JavaScript (ES6 modules) bundled with Vite, and Tailwind CSS v4 for styling.

### Vite Dev Server (fastest for frontend work)

```bash
npm run dev
```

Run this in a separate terminal alongside the Go server. Provides instant hot module replacement (HMR) for JS and CSS changes. Access the game at `http://localhost:5173` — API calls are proxied to your Go backend.

### Watch Mode

```bash
npm run build:watch
```

Continuously rebuilds `www/dist/` on file changes while you run the Go server separately. Useful when you want the Go server handling everything but still want frontend rebuilds.

### Production Build

```bash
npm run build
```

Outputs optimized JS bundles and a single CSS file to `www/dist/`.

## Database

The server expects a SQLite database at `www/game.db`. The JSON files in `game-data/` are the source of truth — the database is a runtime cache.

### Running Migration

Build codex if you haven't already, then run with the `-migrate` flag:

```bash
# Build (one-time, rebuild after pulling codex changes)
go build -o codex ./cmd/codex      # Linux/macOS
go build -o codex.exe ./cmd/codex  # Windows

# Run migration
./codex -migrate       # Linux/macOS
codex.exe -migrate     # Windows
```

Run migration:
- Before the first server start (the server won't start without the database)
- After modifying any JSON files in `game-data/`
- After pulling changes that include game data updates
- If the database becomes corrupted — delete `www/game.db` and re-run migration

### Validating Game Data

```bash
./codex -validate       # Linux/macOS
codex.exe -validate     # Windows
```

Validates all JSON files in `game-data/` and exits with code 0 (pass) or 1 (errors). This runs automatically in CI.

## API Documentation

Swagger annotations are written inline in `cmd/server/api/routes.go`. To regenerate the docs:

```bash
swag init -g cmd/server/api/routes.go -o docs/api/swagger --parseDependency
```

The generated files in `docs/api/` are gitignored and built during deployment. Swagger UI is served at `/swagger/` when the server is running.

## Debug Mode

Set `debug_mode: true` in your `config.yml` to enable:

- **Debug API endpoints**: `/api/debug/sessions` and `/api/debug/state` for inspecting live sessions
- **In-game debug console**: Appears in the bug report modal, shows the live in-memory game state for your character

## Project Structure

```
pubkey-quest/
├── config.yml              # Server config (gitignored, copy from examples)
├── .air.toml               # Air live-reload config (committed)
├── go.mod / go.sum         # Go module (root level)
├── package.json            # npm/Vite config (root level)
│
├── cmd/
│   ├── server/             # Game server
│   │   ├── main.go         # Entry point
│   │   ├── api/            # API handlers + route registration
│   │   ├── app/            # App initialization and lifecycle
│   │   ├── auth/           # Nostr authentication (Grain client)
│   │   ├── cache/          # Caching layer
│   │   ├── db/             # SQLite database connection
│   │   ├── game/           # Game logic (inventory, movement, NPCs, shops, etc.)
│   │   ├── routes/         # HTML page route handlers
│   │   ├── session/        # In-memory session management
│   │   ├── utils/          # Config, templates, utilities
│   │   └── world/          # World and merchant data
│   │
│   └── codex/              # Dev tooling
│       ├── migration/      # JSON → SQLite migration
│       ├── itemeditor/     # Item editor + AI image generation
│       ├── staging/        # GitHub staging integration
│       ├── pixellab/       # Batch image generation
│       └── validation/     # Data validation and cleanup
│
├── types/                  # Shared Go type definitions
│
├── src/                    # Frontend source (ES6 modules)
│   ├── vite.config.js      # Vite bundler config
│   ├── postcss.config.js   # PostCSS/Tailwind config
│   ├── styles/             # CSS source (Tailwind v4)
│   └── ...                 # JS modules (lib, data, state, ui, systems, etc.)
│
├── www/                    # Served to browser
│   ├── game.db             # SQLite database (generated by migration)
│   ├── dist/               # Built JS/CSS (gitignored)
│   ├── views/              # Go HTML templates
│   └── res/                # Static assets (images, fonts)
│
├── game-data/              # Game content JSON (source of truth)
│   ├── items/              # Item definitions
│   ├── magic/              # Spells and spell slots
│   ├── monsters/           # Creature stat blocks
│   ├── locations/          # World map and locations
│   ├── npcs/               # NPCs organized by location
│   └── systems/            # System configs (character gen, music)
│
├── docs/                   # Documentation
│   ├── api/                # Generated Swagger docs (gitignored)
│   ├── development/        # Developer guides (you are here)
│   └── draft/              # Planning and design documents
│
└── data/saves/             # Player save files (gitignored)
```

## Deployment

The test server auto-deploys on every push to main via a **self-hosted GitHub Actions runner**. The `deploy-test.yml` workflow handles the full build pipeline:

1. `npm install` + `npm run build` (frontend)
2. `go build -o codex ./cmd/codex` + `./codex -migrate` (database)
3. `swag init ...` (API docs)
4. `go build -o pubkey-quest ./cmd/server` (server binary with version ldflags)
5. `sudo systemctl restart pubkey-quest-test` + `sudo systemctl restart codex`

A separate `ci.yml` workflow runs on every push and PR (GitHub-hosted runner) to validate builds and run tests — PR merge is blocked on failure.

See [deployment/](./deployment/) for the full CI/CD setup, service templates, and runner configuration.

## Versioning

Version is derived from git state and injected at build time via ldflags. Local dev builds (via `air`) just say `dev`. CI and deploy builds get the commit hash or tag:

```bash
VERSION=$(git describe --tags --always --dirty)
go build -ldflags "-X main.Version=${VERSION}" -o pubkey-quest ./cmd/server
```

Check version: `./pubkey-quest -version` or `./codex -version`.

## Contributing

### Code Style

- **Go**: Run `go fmt` before committing. Follow standard Go conventions.
- **JavaScript**: Vanilla JS only — no frameworks. Event-driven architecture.

### Modifying Game Data

Always edit JSON files in `game-data/` — never edit `www/game.db` directly. Run migration after changes.

### Troubleshooting

| Problem | Fix |
|---------|-----|
| Styling broken | `npm run build` (ensure `npm install` was run first) |
| JS not loading | Check `www/dist/` exists, run `npm run build` |
| Database missing/corrupt | Delete `www/game.db`, rebuild codex and run `codex -migrate` |
| Server won't start | Ensure `www/game.db` exists — build codex and run migration first |
| npm install fails | Ensure Node.js 18+: `node --version` |
