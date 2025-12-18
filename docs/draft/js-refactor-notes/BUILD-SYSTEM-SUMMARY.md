# JavaScript Build System - Setup Summary

**Date**: 2025-12-16
**Status**: Phase 1-2 Complete, Ready for Review

---

## What We've Built

A modern ES6 module build system with Vite for bundling and development.

### Directory Structure

```
nostr-hero/
├── server/                       # Go backend source
│   ├── main.go                   # Server entry point
│   ├── go.mod                    # Go module definition
│   ├── go.sum                    # Go dependencies
│   ├── api/
│   ├── auth/
│   ├── db/
│   ├── routes/
│   ├── types/
│   ├── utils/
│   ├── cache/
│   ├── handlers/
│   └── functions/
│
├── src/                          # JavaScript source (new, ES6 modules)
│   ├── lib/                      # Layer 1: Core libraries (no dependencies)
│   │   ├── logger.js            # ✅ Smart logging (dev vs prod)
│   │   ├── events.js            # Event bus (TODO)
│   │   ├── session.js           # Session management (TODO)
│   │   └── auth.js              # Authentication (TODO)
│   │
│   ├── data/                     # Layer 2: Data layer
│   ├── state/                    # Layer 3: State management
│   ├── logic/                    # Layer 4: Pure game logic
│   ├── systems/                  # Layer 5: Complex systems
│   │   ├── inventory/
│   │   ├── containers/
│   │   ├── equipment/
│   │   └── nostr-integration/
│   ├── ui/                       # Layer 6: UI rendering
│   │   ├── core/
│   │   ├── character/
│   │   ├── inventory/
│   │   ├── location/
│   │   ├── modals/
│   │   ├── combat/
│   │   └── widgets/
│   ├── pages/                    # Layer 7: Page entry points
│   │   ├── test.js              # ✅ Build system test
│   │   ├── game.js              # TODO
│   │   ├── intro.js             # TODO
│   │   ├── new-game.js          # TODO
│   │   └── index.js             # TODO
│   ├── components/               # Layer 8: Reusable components
│   │   └── buttons/
│   └── config/
│       └── constants.js          # ✅ Game constants
│
├── www/                          # Public folder (unchanged)
│   ├── dist/                     # Build output (gitignored)
│   │   ├── dev/                  # Development builds
│   │   └── prod/                 # Production builds
│   ├── views/                    # HTML templates
│   ├── res/                      # Static resources
│   └── scripts/                  # Old vanilla JS (will be deleted after migration)
│
├── docs/development/             # Development configs (examples)
│   ├── Makefile                  # ✅ First-time setup automation
│   ├── example.package.json      # ✅ npm config
│   ├── example.vite.config.js    # ✅ Vite bundler config
│   ├── example.air.toml          # Air config (existing)
│   └── readme.md                 # ✅ Updated with JS + CSS workflows
│
├── test.html                     # ✅ Build system test page
├── .gitignore                    # ✅ Updated for node_modules/, dist/, etc.
└── JAVASCRIPT-REFACTOR.md        # ✅ Updated with correct paths

# These will be created by Makefile:
├── package.json                  # (gitignored, copied from example)
├── vite.config.js                # (gitignored, copied from example)
└── node_modules/                 # (gitignored, npm install)
```

---

## Key Files Created

### 1. Build System Configuration

**`docs/development/example.package.json`**
- npm scripts for development and production
- Integrated Tailwind CSS compilation
- Vite + Tailwind + Concurrently dependencies

**Key scripts:**
- `npm run dev` - Vite dev server (HMR)
- `npm run dev:css` - Tailwind CSS watch mode
- `npm run dev:full` - Both Vite + CSS together
- `npm run build` - Production build (minified, tree-shaken)
- `npm run clean` - Clean dist/

**`docs/development/example.vite.config.js`**
- Entry points for test, game, intro, new-game, index pages
- Development vs production modes
- Source maps in development
- Minification in production
- Code splitting (vendor, ui-core chunks)
- Path aliases (@lib, @data, @ui, etc.)
- Proxy to Go backend (port 8080)

**`docs/development/Makefile`**
- `make setup` - Complete first-time setup
- `make setup-configs` - Copy example configs to root
- `make install-deps` - Install npm dependencies
- `make clean-configs` - Remove local configs

### 2. JavaScript Source Files

**`src/lib/logger.js`** (Layer 1 - Foundation)
- Smart logging system
- Development: Shows all logs (debug, info, warn, error)
- Production: Shows only errors
- Uses `__DEV__` and `__PROD__` globals from Vite

**`src/config/constants.js`**
- Game configuration constants
- API URLs, inventory config, UI settings

**`src/pages/test.js`**
- Build system verification
- Tests logger, module imports, DOM manipulation
- Confirms dev vs prod modes work

### 3. Test Page

**`test.html`**
- Standalone test page for build system
- Instructions for development and production builds
- Expected console output documented
- Can run with: `npm run dev` → visit http://localhost:5173/test.html

### 4. Documentation Updates

**`docs/development/readme.md`**
- Added automated setup instructions (Makefile)
- Added JavaScript development section
- Added CSS compilation via npm scripts
- Integrated workflows for Vite + Tailwind

**`.gitignore`**
- Excludes `node_modules/`
- Excludes `www/dist/`
- Excludes local configs (package.json, vite.config.js)
- Keeps examples in docs/development/

---

## How It Works

### Development Workflow

1. **First-time setup:**
   ```bash
   cd docs/development
   make setup
   ```

2. **Start Go backend:**
   ```bash
   air
   ```

3. **Start Vite dev server:**
   ```bash
   npm run dev
   # Access at http://localhost:5173
   ```

4. **Or start both Vite + CSS together:**
   ```bash
   npm run dev:full
   ```

### Development Features

- **Hot Module Replacement (HMR)**: Edit JS → instant browser update
- **Source Maps**: Debug original source code, not bundled code
- **All logs shown**: console.debug, info, warn, error all visible
- **Fast**: Native ES modules, no bundling needed
- **Proxy**: API calls automatically forwarded to Go backend

### Production Build

```bash
# Build optimized bundles
npm run build

# Preview production build
npm run preview
```

### Production Features

- **Minified**: esbuild minification
- **Tree-Shaken**: Dead code automatically removed
- **Code Split**: Vendor and UI chunks for better caching
- **Hashed Filenames**: Automatic cache busting (e.g., `game.a3f9b2c8.js`)
- **Only Errors Logged**: Debug/info/warn logs stripped out

---

## Testing the Build System

### Quick Test

1. Run the Makefile setup:
   ```bash
   cd docs/development
   make setup
   ```

2. Start Vite dev server:
   ```bash
   npm run dev
   ```

3. Open http://localhost:5173/test.html

4. Check for:
   - ✅ Green "Build system test successful!" banner
   - ✅ Console logs showing correct environment
   - ✅ No errors in console

### Expected Console Output (Development)

```
[Nostr Hero] ℹ️ Nostr Hero v1.0.0 - Build System Test
[Nostr Hero] 🐛 This debug message should only appear in development
[Nostr Hero] ⚠️ This warning should only appear in development
[Nostr Hero] ❌ This error should appear in both dev and production
[Nostr Hero] 🐛 Running in DEVELOPMENT mode
[Nostr Hero] 🐛 Debug logs are enabled
[Nostr Hero] ℹ️ Build system version: 1.0.0
[Nostr Hero] ℹ️ DOM manipulation test successful
```

### Expected Console Output (Production)

```
[Nostr Hero] ❌ This error should appear in both dev and production
```

---

## Next Steps

Before continuing with the full conversion, **please review and test**:

### 1. Test the Setup

```bash
cd docs/development
make setup
npm run dev
# Visit http://localhost:5173/test.html
```

### 2. Verify Structure

- [x] `server/` contains Go code
- [x] `src/` contains new JS modules
- [x] `www/scripts/` still exists (old code, not touched yet)
- [x] Build config examples in `docs/development/`

### 3. Check Documentation

- [x] `docs/development/readme.md` - Updated with new workflows
- [x] `docs/development/Makefile` - Automation scripts
- [x] `JAVASCRIPT-REFACTOR.md` - Updated with correct paths

---

## Once Approved, We'll Continue With:

### Phase 3: Convert Existing JavaScript to ES6 Modules

Layer by layer conversion of `www/scripts/` to `src/`:

1. **Layer 1 (lib/)**: Core libraries - session, auth, API client
2. **Layer 2 (data/)**: Item data, character data, profiles
3. **Layer 3 (state/)**: Game state (breaking circular dependencies!)
4. **Layer 4 (logic/)**: Character generation, game rules
5. **Layer 5 (systems/)**: Inventory, equipment, containers
6. **Layer 6 (ui/)**: Split game-ui.js (1,911 lines) into 15+ modules
7. **Layer 7 (pages/)**: game.js, intro.js, new-game.js entry points
8. **Layer 8 (components/)**: Reusable UI components

### Phase 4: Update HTML Templates

Update `www/views/` templates to use bundled modules instead of 13+ script tags.

### Phase 5: Test Everything

Build and test both development and production modes.

### Phase 6: Cleanup

Delete old `www/scripts/` directory once everything is migrated and tested.

---

## Questions to Answer

1. **Does the Makefile setup work correctly?**
2. **Does `npm run dev` start the Vite server?**
3. **Does the test page show the green banner?**
4. **Are console logs correct in dev mode?**
5. **Does the directory structure make sense?**
6. **Is the documentation clear?**

Once you approve this foundation, we'll proceed with converting the existing JavaScript codebase layer by layer!
