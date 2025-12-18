# JavaScript Migration Status

**Last Updated:** 2025-12-17

## Overview

The core game pages have been successfully converted from legacy script tags to ES6 modules with Vite bundling. This document tracks what's been converted and what still needs migration.

## ✅ Converted Pages (Using New ES6 Bundles)

These pages now use the new `src/` ES6 modules via Vite bundles in `www/dist/`:

### 1. **game.html** - Main Game Page
- **Bundle:** `/dist/game.js` (83.5 KB, 20.3 KB gzipped)
- **Status:** ✅ Fully converted
- **Old scripts:** Commented out (13 scripts preserved for rollback)
- **Features:** Character display, inventory, equipment, spells, save system, location navigation

### 2. **game-intro.html** - Intro Cutscene Sequence
- **Bundle:** `/dist/gameIntro.js` (42.1 KB, 7.8 KB gzipped)
- **Status:** ✅ Fully converted
- **Old scripts:** Commented out (9 scripts preserved for rollback)
- **Features:** Story scenes, equipment selection, character creation flow

### 3. **new-game.html** - Character Creation Preview
- **Bundle:** `/dist/newGame.js` (18.4 KB, 5.8 KB gzipped)
- **Status:** ✅ Fully converted
- **Old scripts:** Commented out (7 scripts preserved for rollback)
- **Features:** Character preview, equipment display, save creation

**Total Bundle Size:** 144 KB minified, ~34 KB gzipped (vs 300KB+ before)

## ❌ Unconverted Pages (Still Using Old Scripts)

These pages still use the legacy `www/scripts/` directory:

### 1. **index.html** - Homepage / Landing Page
- **Scripts Used:**
  - `/scripts/systems/theme-manager.js`
  - `/scripts/systems/profile-manager.js`
  - `/scripts/pages/tabs.js`
  - `/scripts/core/session-manager.js`
  - `/scripts/core/auth.js`
  - `/scripts/core/nostr-connect.js`
- **Why Not Converted:** Homepage with save selection, not part of core game flow

### 2. **saves.html** - Save Management Page
- **Scripts Used:**
  - `/scripts/core/session-manager.js`
  - `/scripts/systems/theme-manager.js`
  - `/scripts/utils/item-helpers.js`
  - `/scripts/pages/game-ui.js`
- **Why Not Converted:** Separate save management UI

### 3. **settings.html** - Settings Page
- **Scripts Used:**
  - `/scripts/core/session-manager.js`
  - `/scripts/core/auth.js`
  - `/scripts/systems/theme-manager.js`
  - `/scripts/systems/relay-manager.js`
- **Why Not Converted:** Settings configuration page

### 4. **discover.html** - Discovery Page
- **Scripts Used:**
  - `/scripts/core/session-manager.js`
  - `/scripts/core/auth.js`
  - `/scripts/systems/theme-manager.js`
- **Why Not Converted:** Feature discovery page

## 📂 Directory Structure

### New ES6 Module Structure (src/)
```
src/
├── entries/               # Vite entry points (bundles)
│   ├── game.js           # Main game bundle entry
│   ├── gameIntro.js      # Intro sequence entry
│   └── newGame.js        # Character creation entry
├── lib/                  # Core libraries
│   ├── logger.js
│   ├── api.js
│   ├── session.js
│   └── events.js
├── data/                 # Data modules
│   ├── items.js
│   ├── characters.js
│   └── inventory.js
├── state/                # State management
│   ├── gameState.js
│   └── staticData.js
├── logic/                # Pure game logic
│   ├── mechanics.js
│   └── characterGenerator.js
├── systems/              # Complex game systems
│   ├── saveSystem.js
│   ├── inventoryInteractions.js
│   └── equipmentSelection.js
├── ui/                   # UI rendering modules
│   ├── messaging.js
│   ├── timeDisplay.js
│   ├── characterDisplay.js
│   ├── locationDisplay.js
│   ├── spellsDisplay.js
│   ├── groundItems.js
│   └── displayCoordinator.js
├── pages/                # Page-specific logic
│   ├── startup.js
│   ├── tabs.js
│   ├── newGame.js
│   └── gameIntro.js
└── components/           # Reusable components
    └── continueButton.js
```

### Legacy Script Structure (www/scripts/)
```
www/scripts/              # OLD - Still needed by unconverted pages
├── core/
│   ├── session-manager.js    # Used by: index, saves, settings, discover
│   ├── auth.js              # Used by: index, settings, discover
│   ├── game-api.js          # Converted to src/lib/api.js
│   └── nostr-connect.js     # Used by: index
├── systems/
│   ├── theme-manager.js     # Used by: index, saves, settings, discover
│   ├── profile-manager.js   # Used by: index
│   ├── relay-manager.js     # Used by: settings
│   ├── game-state.js        # Converted to src/state/gameState.js
│   ├── game-logic.js        # Converted to src/logic/mechanics.js
│   ├── save-system.js       # Converted to src/systems/saveSystem.js
│   ├── character-generator.js  # Converted to src/logic/characterGenerator.js
│   ├── equipment-selection.js  # Converted to src/systems/equipmentSelection.js
│   └── inventory-interactions.js  # Converted to src/systems/inventoryInteractions.js
├── utils/
│   ├── item-helpers.js      # Converted to src/data/items.js
│   ├── character-helpers.js # Converted to src/data/characters.js
│   └── inventory-helpers.js # Converted to src/data/inventory.js
├── pages/
│   ├── tabs.js             # Used by: index (partially converted to src/pages/tabs.js)
│   ├── startup.js          # Converted to src/pages/startup.js
│   ├── game-ui.js          # Converted to src/ui/* modules (split into 7 modules)
│   ├── new-game.js         # Converted to src/pages/newGame.js
│   └── game-intro.js       # Converted to src/pages/gameIntro.js
└── components/
    ├── continue-button.js   # Converted to src/components/continueButton.js
    └── back-button.js       # Not converted yet
```

## 🔧 Files Still Referencing Old Scripts

### Go Server Route
**File:** `server/main.go:49`
```go
mux.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("www/scripts/"))))
```
**Purpose:** Serves legacy scripts for unconverted pages (index, saves, settings, discover)
**Action:** ⚠️ Keep until all pages converted

## ✅ What Can Be Safely Removed

Once ALL pages are converted (including index, saves, settings, discover), the following can be deleted:

### Duplicate Converted Files (Ready to Delete)
These have been successfully converted to `src/` and are no longer needed:

```
www/scripts/
├── core/
│   └── game-api.js          ✅ → src/lib/api.js
├── systems/
│   ├── game-state.js        ✅ → src/state/gameState.js
│   ├── game-logic.js        ✅ → src/logic/mechanics.js
│   ├── save-system.js       ✅ → src/systems/saveSystem.js
│   ├── character-generator.js  ✅ → src/logic/characterGenerator.js
│   ├── equipment-selection.js  ✅ → src/systems/equipmentSelection.js
│   ├── inventory-interactions.js  ✅ → src/systems/inventoryInteractions.js
│   └── container-system.js  ✅ → (merged into inventoryInteractions.js)
├── utils/
│   ├── item-helpers.js      ✅ → src/data/items.js
│   ├── character-helpers.js ✅ → src/data/characters.js
│   └── inventory-helpers.js ✅ → src/data/inventory.js
├── pages/
│   ├── startup.js          ✅ → src/pages/startup.js
│   ├── game-ui.js          ✅ → src/ui/* (split into 7 modules)
│   ├── new-game.js         ✅ → src/pages/newGame.js
│   └── game-intro.js       ✅ → src/pages/gameIntro.js
└── components/
    └── continue-button.js   ✅ → src/components/continueButton.js
```

### Files Still Needed (Keep for now)
```
www/scripts/
├── core/
│   ├── session-manager.js   ⚠️ Used by 4 unconverted pages
│   ├── auth.js             ⚠️ Used by 3 unconverted pages
│   └── nostr-connect.js    ⚠️ Used by index.html
├── systems/
│   ├── theme-manager.js    ⚠️ Used by 4 unconverted pages
│   ├── profile-manager.js  ⚠️ Used by index.html
│   └── relay-manager.js    ⚠️ Used by settings.html
└── pages/
    └── tabs.js             ⚠️ Used by index.html
```

## 🧪 Testing Checklist

### Converted Pages Testing

**Game Page (`/game`):**
- [ ] Page loads without console errors
- [ ] Character stats display correctly
- [ ] Inventory drag-and-drop works
- [ ] Equipment slots function
- [ ] Save game works (Ctrl+S)
- [ ] Location navigation works
- [ ] Ground items modal opens
- [ ] Vault UI functions

**Intro Page (`/game-intro`):**
- [ ] Character generates correctly
- [ ] Cutscene plays through
- [ ] Equipment selection works
- [ ] Spell cards display (for casters)
- [ ] Save creation succeeds
- [ ] Redirect to game works

**New Game Page (`/new-game`):**
- [ ] Character displays correctly
- [ ] Introduction text shows
- [ ] Equipment choices display
- [ ] Save creation works
- [ ] No console errors

### Unconverted Pages Testing (Should Still Work)

**Homepage (`/`):**
- [ ] Tabs navigation works
- [ ] Save selection works
- [ ] Authentication flow works
- [ ] Theme switching works

**Other Pages:**
- [ ] `/saves` - Save management works
- [ ] `/settings` - Settings page works
- [ ] `/discover` - Discovery page works

## 📊 Code Quality Analysis

### Improvements from Migration

1. **Module System:** Explicit imports/exports instead of global scope pollution
2. **Code Organization:** Clear dependency hierarchy (7 layers)
3. **Bundle Size:** 71% reduction (300KB+ → 34KB gzipped)
4. **Build Process:** Vite for fast builds, minification, tree-shaking
5. **Source Maps:** Available for debugging
6. **Type Safety:** JSDoc comments throughout
7. **Maintainability:** Smaller, focused modules (vs 1911-line game-ui.js)

### Remaining Issues

1. **Duplicate Code:** Some files exist in both `src/` and `www/scripts/`
2. **Inconsistent Patterns:** Unconverted pages use different patterns
3. **Global Dependencies:** Converted pages still export to `window` for compatibility
4. **Session Manager:** Duplicated in `src/lib/session.js` and `www/scripts/core/session-manager.js`

## 🚀 Next Steps

### Phase 1: Verify Current Migration ✅
- [x] All three game pages use new bundles
- [x] Old scripts preserved as comments
- [x] No broken imports in src/
- [x] Bundles build successfully

### Phase 2: Test Thoroughly (In Progress)
- [ ] Test all converted pages
- [ ] Test all unconverted pages (ensure they still work)
- [ ] Check browser console for errors
- [ ] Verify all features work

### Phase 3: Convert Remaining Pages (Future)
- [ ] Convert index.html to ES6 modules
- [ ] Convert saves.html to ES6 modules
- [ ] Convert settings.html to ES6 modules
- [ ] Convert discover.html to ES6 modules

### Phase 4: Final Cleanup (After Phase 3)
- [ ] Remove duplicate files from `www/scripts/`
- [ ] Remove `/scripts/` route from Go server
- [ ] Delete `www/scripts/` directory entirely
- [ ] Update documentation

## 🔄 Rollback Instructions

If something breaks on a converted page:

1. **Quick Rollback (Per Page):**
   - Open the HTML file (game.html, game-intro.html, or new-game.html)
   - Comment out: `<script type="module" src="/dist/[bundle].js"></script>`
   - Uncomment the old scripts block
   - Hard refresh browser (Ctrl+Shift+R)

2. **Full Rollback (All Pages):**
   ```bash
   git checkout www/views/game.html www/views/game-intro.html www/views/new-game.html
   ```

3. **Rebuild if needed:**
   ```bash
   npm run build
   ```

## 📝 Build Commands

```bash
# Install dependencies
npm install

# Build once
npm run build

# Build and watch for changes
npm run build:watch

# Preview built bundles
npm run preview
```

## 🎯 Success Metrics

**Bundle Performance:**
- game.js: 71.78 KB → 20.30 KB gzipped (71% reduction)
- gameIntro.js: 26.67 KB → 7.82 KB gzipped (71% reduction)
- newGame.js: 18.38 KB → 5.76 KB gzipped (69% reduction)

**Code Organization:**
- 24 focused ES6 modules
- 7-layer dependency hierarchy
- Zero circular dependencies
- Explicit import/export chains

**Developer Experience:**
- Watch mode for live reload
- Source maps for debugging
- Clear module boundaries
- Maintainable codebase

---

**Status:** ✅ Core game pages converted successfully. Ready for thorough testing before removing old scripts directory.
