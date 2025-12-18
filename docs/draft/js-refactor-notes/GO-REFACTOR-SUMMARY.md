# Go Backend Refactor - Complete

**Date**: 2025-12-16
**Status**: ✅ Complete and Tested

---

## What Changed

Reorganized the Go backend structure to cleanly separate backend and frontend code.

### Before

```
nostr-hero/
├── main.go                    # In root
├── go.mod                     # In root
├── go.sum                     # In root
└── src/                       # Go backend packages
    ├── api/
    ├── auth/
    ├── db/
    └── ...
```

**Problems:**
- `src/` was ambiguous (could be Go or JavaScript)
- main.go in root was messy
- Confusing when adding JavaScript src/ directory

### After

```
nostr-hero/
├── server/                    # All Go code here
│   ├── main.go                # Server entry point
│   ├── go.mod                 # Go module definition
│   ├── go.sum                 # Go dependencies
│   │
│   ├── api/                   # API handlers
│   ├── auth/                  # Authentication
│   ├── cache/                 # Caching layer
│   ├── db/                    # Database layer
│   ├── routes/                # Route handlers
│   ├── types/                 # Type definitions
│   ├── utils/                 # Utility functions
│   ├── handlers/              # HTTP handlers
│   └── functions/             # Game logic
│
└── src/                       # JavaScript ES6 modules
    ├── lib/
    ├── data/
    ├── pages/
    └── ...
```

**Benefits:**
- ✅ Clear separation: `server/` = Go, `src/` = JavaScript
- ✅ All Go code in one place
- ✅ Follows common convention for Go+JS monorepos
- ✅ Cleaner root directory

---

## Changes Made

### 1. Moved Files

```bash
# Moved Go files into server/
mv main.go go.mod go.sum server/
```

### 2. Updated Import Paths

**Before** (when main.go was in root):
```go
import (
    "nostr-hero/src/api"
    "nostr-hero/src/auth"
    "nostr-hero/src/db"
    // ...
)
```

**After** (with main.go and go.mod in server/):
```go
import (
    "nostr-hero/api"
    "nostr-hero/auth"
    "nostr-hero/db"
    // ...
)
```

**Why?**
- go.mod is now in `server/`, so that's the module root
- Packages are relative to the module root
- No need for `server/` prefix since we're already in the module

### 3. Updated Dependencies

```go
// Removed local replace directive
// Before:
replace github.com/0ceanslim/grain => ../grain

// After:
// (removed - uses actual GitHub dependency)
```

Then ran `go mod tidy` to download dependencies from GitHub.

### 4. Updated Build Configuration

**`docs/development/example.air.toml`:**

```toml
# Before:
cmd = "go build -o ./tmp/main.exe ."
include_dir = ["src"]

# After:
cmd = "go build -o ./tmp/main.exe ./server"
include_dir = ["server"]
exclude_dir = [..., "src", "node_modules"]
```

**Why?**
- Build from server/ directory
- Watch server/ for changes (not src/ which is now JavaScript)
- Exclude JavaScript src/ and node_modules/ from Go file watching

### 5. Updated Documentation

**Files updated:**
- `CLAUDE.md` - Project documentation
- `BUILD-SYSTEM-SUMMARY.md` - Build system overview
- `docs/development/readme.md` - Development guide (already had JS section)

All now reflect:
- `server/` for Go backend
- `src/` for JavaScript frontend
- Updated build commands and directory structure

---

## File Changes Summary

**24 Go files updated:**
All import paths changed from `nostr-hero/src/` to `nostr-hero/`.

```
server/api/weights.go
server/api/game_actions.go
server/api/inventory.go
server/api/equipment.go
server/api/character-creation.go
server/api/profile.go
server/api/gamedata.go
server/api/dnd.go
server/auth/grain.go
server/auth/init.go
server/routes/settings.go
server/routes/game.go
server/routes/new-game.go
server/routes/saves.go
server/routes/loadSave.go
server/routes/legacy-registry.go
server/routes/alpha-registry.go
server/routes/index.go
server/routes/discover.go
server/utils/helpers.go
server/utils/fetchUserRelays.go
server/utils/fetchUserMetaData.go
server/utils/registry.go
server/functions/discoverHero.go
```

---

## Build & Run

### Development (with Air)

From project root:

```bash
# Copy example configs (if not done yet)
cd docs/development
make setup

# Start server with live reload
air
```

Air will:
1. Watch `server/` directory for .go file changes
2. Rebuild on changes: `go build -o ./tmp/main.exe ./server`
3. Automatically restart the server

### Manual Build

From project root:

```bash
# Build from server/ directory
go build -o nostr-hero.exe ./server

# Run
./nostr-hero.exe
```

### From Server Directory

You can also work directly in server/:

```bash
cd server

# Build
go build -o ../nostr-hero.exe .

# Or with Air (need to adjust .air.toml root)
# Not recommended - stay in project root
```

---

## Testing

✅ **Build Test Passed**

```bash
cd server && go build -o ../tmp/test-build.exe .
```

**Result:** Successfully built 22MB binary

No errors, all imports resolved correctly.

---

## Developer Workflow

### Backend Development

1. **Edit Go code** in `server/` directory
2. **Air auto-reloads** on save
3. **Check logs** in terminal

### Frontend Development

1. **Edit JavaScript** in `src/` directory
2. **Vite HMR** updates browser instantly
3. **Check console** in browser DevTools

### Full-Stack Development

Run in parallel (two terminals):

```bash
# Terminal 1: Go backend with Air
air

# Terminal 2: JavaScript with Vite + CSS
npm run dev:full
```

---

## Migration Notes

### For Other Developers

If you have uncommitted changes or local modifications:

1. **Backup your work**
2. **Pull latest changes**
3. **Update your Air config:**
   ```bash
   cd docs/development
   make setup  # Recreates .air.toml with correct paths
   ```
4. **Rebuild:**
   ```bash
   cd server
   go mod tidy
   go build -o ../nostr-hero.exe .
   ```

### Common Issues

**Issue: "package nostr-hero/server/api not found"**

**Solution:** Imports should be `nostr-hero/api` not `nostr-hero/server/api`
- go.mod is in server/, so that's the module root
- Already fixed in this refactor

**Issue: "grain dependency not found"**

**Solution:** Run `go mod tidy` in server/ directory
- Downloads dependencies from GitHub
- Updates go.sum

**Issue: "Air not building"**

**Solution:** Update .air.toml
```bash
cd docs/development
make setup
```

---

## What's Next

The Go refactor is complete. Next steps:

1. **Continue JavaScript refactor** - Convert `www/scripts/` to `src/` ES6 modules
2. **Test integration** - Ensure Go backend serves new JavaScript bundles
3. **Update HTML templates** - Use bundled JS instead of script tags

See `JAVASCRIPT-REFACTOR.md` for the full JavaScript migration plan.

---

## Final Structure

```
nostr-hero/
├── server/                      # ✅ Go backend (complete)
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   └── [api, auth, db, routes, types, utils, ...]
│
├── src/                         # 🚧 JavaScript frontend (in progress)
│   ├── lib/logger.js            # ✅ Created
│   ├── config/constants.js      # ✅ Created
│   ├── pages/test.js            # ✅ Created
│   └── [rest of structure]      # ⏳ To be migrated from www/scripts/
│
├── www/
│   ├── dist/                    # ✅ Build output (gitignored)
│   ├── scripts/                 # ⏳ Old JS (will be deleted after migration)
│   └── [views, res, ...]
│
└── docs/development/
    ├── Makefile                 # ✅ Automated setup
    ├── example.air.toml         # ✅ Updated for server/
    ├── example.package.json     # ✅ npm config
    └── example.vite.config.js   # ✅ Vite config
```

**Status:**
- ✅ Go refactor complete
- ✅ Build system setup complete
- 🚧 JavaScript conversion in progress
- ⏳ Full migration pending

