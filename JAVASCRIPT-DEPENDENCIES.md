# JavaScript Dependency Map
**Date**: 2025-12-15

## Overview

This document maps the dependencies between JavaScript files to understand the load order requirements and identify tight coupling.

## Current Load Order (game.html)

```
1. session-manager.js    # Core: Auth & session
2. game-api.js           # Core: API calls
3. item-helpers.js       # Utils: Item data loading
4. character-helpers.js  # Utils: Vault/location helpers
5. game-ui.js            # Pages: Main UI (EXPORTS MANY FUNCTIONS)
6. auth.js               # Core: Authentication
7. character-generator.js# Systems: Character generation
8. game-state.js         # Systems: State management (EXPORTS MANY FUNCTIONS)
9. game-logic.js         # Systems: Game rules
10. save-system.js       # Systems: Save/load
11. container-system.js  # Systems: Container management
12. inventory-interactions.js # Systems: Inventory drag/drop
13. startup.js           # Pages: Initialization
```

## Dependency Graph

### Core Layer (No dependencies on other layers)

#### session-manager.js
- **Exports**:
  - `window.sessionManager` - Session state & auth
- **Imports**: None (standalone)
- **Used By**: All files that need authentication

#### game-api.js
- **Exports**:
  - `window.gameAPI` - API call wrappers
- **Depends On**:
  - `session-manager.js` (for npub/session)
- **Used By**: Any file making API calls

#### auth.js
- **Exports**: Various auth functions
- **Depends On**:
  - `session-manager.js`
- **Used By**: Login/logout flows

### Utils Layer (Shared helpers, minimal dependencies)

#### item-helpers.js
- **Exports**:
  - `loadItemsFromDatabase()` - Fetch all items
  - `getItemById(itemId)` - Get single item (async)
- **Depends On**: None (direct fetch calls)
- **Used By**: Many files (game-ui, game-intro, equipment-selection, etc.)

#### character-helpers.js
- **Exports**:
  - Vault & location helper functions
- **Depends On**: Unknown (need to audit)
- **Used By**: game-state, game-ui

#### inventory-helpers.js
- **Exports**: Inventory creation utilities
- **Depends On**: Unknown (need to audit)
- **Used By**: game-state, character-generator

### Systems Layer (Game logic & state)

#### game-state.js (⚠️ CRITICAL - Many exports)
- **Exports**:
  - `window.getGameState()` - Get current state
  - `window.updateGameState(state)` - Update state
  - `window.refreshGameState()` - Reload state
  - `getItemById(itemId)` - Cached item lookup
  - `getSpellById(spellId)` - Cached spell lookup
  - `getLocationById(locationId)` - Cached location lookup
  - `getMonsterById(monsterId)` - Cached monster lookup
  - `getPackById(packId)` - Cached pack lookup
  - `getNPCById(npcId)` - Cached NPC lookup
  - `initializeGame()` - Game init
  - Ground item functions
- **Depends On**:
  - `session-manager.js` (for npub)
  - `item-helpers.js` (for loading items)
  - `character-helpers.js`
  - `game-ui.js` (calls showMessage, updateCharacterDisplay, etc.)
- **Used By**: ALMOST EVERY FILE (central state)

#### game-logic.js
- **Exports**:
  - `moveToLocation(locationId)` - Travel logic
  - `useItem(itemId, slot)` - Item usage
- **Depends On**:
  - `game-state.js` (getGameState, updateGameState)
  - `game-ui.js` (showActionText)
- **Used By**: UI event handlers

#### save-system.js
- **Exports**:
  - `window.saveGame()` - Save current game
  - Auto-save interval
- **Depends On**:
  - `session-manager.js`
  - `game-state.js`
  - `game-api.js`
  - `game-ui.js` (showMessage)
- **Used By**: Manual save button, auto-save timer

#### character-generator.js
- **Exports**: Character generation functions
- **Depends On**:
  - API calls (direct fetch)
- **Used By**: new-game.js, game-intro.js

#### inventory-interactions.js
- **Exports**:
  - `window.inventoryInteractions` - Drag/drop handlers
- **Depends On**:
  - `game-state.js`
  - `game-ui.js` (showMessage)
  - `item-helpers.js`
- **Used By**: Inventory UI event handlers

#### container-system.js
- **Exports**: Container management functions
- **Depends On**:
  - `game-state.js`
  - `game-ui.js` (showMessage)
- **Used By**: Container UI

### Pages Layer (UI & initialization)

#### game-ui.js (⚠️ CRITICAL - Too many exports)
- **Exports**:
  - `window.showMessage(text, type, duration)` - Toast messages
  - `window.showActionText(text, color)` - Game log messages
  - `window.addGameLog(message)` - Add to game log
  - `updateCharacterDisplay()` - Render character UI
  - `updateInventoryDisplay()` - Render inventory
  - `updateSpellsDisplay()` - Render spells
  - `displayCurrentLocation()` - Render location
  - `updateTimeDisplay()` - Render time/fatigue
  - `updateAllDisplays()` - Render everything
  - `enterBuilding(buildingId)` - Building entry
  - `exitBuilding()` - Building exit
  - `talkToNPC(npcId)` - NPC dialogue
  - `showVaultUI(vaultData)` - Vault UI
  - `closeVaultUI()` - Close vault
  - `openGroundModal()` - Ground items modal
  - `closeGroundModal()` - Close ground modal
  - `refreshGroundModal()` - Update ground display
  - `pickupGroundItem(itemId)` - Pickup ground item
- **Depends On**:
  - `game-state.js` (getGameState, updateGameState, getItemById, etc.)
  - `item-helpers.js`
  - `character-helpers.js`
- **Used By**: ALMOST EVERY FILE (central UI)

#### startup.js
- **Exports**: None (initialization only)
- **Depends On**:
  - `session-manager.js`
  - `game-state.js` (initializeGame)
  - `game-ui.js` (updateAllDisplays)
- **Used By**: None (runs on page load)

#### game-intro.js
- **Exports**: Functions for intro flow
- **Depends On**:
  - `session-manager.js`
  - `character-generator.js`
  - `item-helpers.js` (duplicates getItemImageName, getItemStats)
- **Used By**: game-intro.html page

#### new-game.js
- **Exports**: Functions for new game screen
- **Depends On**:
  - `session-manager.js`
  - `character-generator.js`
  - `game-ui.js` (showMessage)
- **Used By**: new-game.html page

### Components Layer

#### continue-button.js
- **Exports**: `window.createContinueButton(text, delay, onClick)`
- **Depends On**: None
- **Used By**: game-intro.js

#### back-button.js
- **Exports**: `window.createBackButton(onClick)`
- **Depends On**: None
- **Used By**: equipment-selection.js

## Critical Dependencies (High Coupling)

### game-ui.js is depended on by:
1. game-state.js (calls showMessage, updateCharacterDisplay, etc.)
2. game-logic.js (calls showActionText)
3. save-system.js (calls showMessage)
4. inventory-interactions.js (calls showMessage)
5. container-system.js (calls showMessage)
6. startup.js (calls updateAllDisplays)
7. new-game.js (calls showMessage)
8. auth.js (calls showMessage)
9. session-manager.js (calls showMessage)
10. nostr-connect.js (calls showMessage)

**Problem**: game-ui.js exports too many functions, making it a bottleneck. Every file depends on it.

### game-state.js is depended on by:
1. game-ui.js (calls getGameState, updateGameState, getItemById, etc.)
2. game-logic.js (calls getGameState, updateGameState)
3. save-system.js (calls getGameState)
4. inventory-interactions.js (calls getGameState, updateGameState, getItemById)
5. container-system.js (calls getGameState, updateGameState)
6. startup.js (calls initializeGame)
7. game-intro.js (calls getGameState, updateGameState)
8. new-game.js (calls getGameState)

**Problem**: game-state.js is the central state store (acceptable for state management).

## Circular Dependencies

### game-ui.js ↔ game-state.js
- game-state.js imports from game-ui.js (showMessage, updateCharacterDisplay)
- game-ui.js imports from game-state.js (getGameState, updateGameState, getItemById)

**Problem**: Circular dependency can cause initialization issues. Need to break this cycle.

**Solution**: Extract message/display functions into separate modules that both can import.

## Duplicate Function Calls

### Functions that fetch items:
1. `item-helpers.js::getItemById()` - Async, fetches from API
2. `game-state.js::getItemById()` - Sync, reads from cached DOM
3. `game-ui.js::getItemByIdAsync()` - Async, fetches from DB

**Problem**: Three different ways to get items. Confusing and error-prone.

**Solution**: Unify into single `getItemById(itemId, cached = true)` function:
- If `cached=true`, read from DOM cache (fast)
- If `cached=false`, fetch from API (fresh data)

### Functions that show messages:
1. `game-ui.js::showMessage()` - Toast notifications
2. `game-ui.js::showActionText()` - Colored game log
3. `game-ui.js::addGameLog()` - Plain game log

**Problem**: All in game-ui.js, making it a central dependency.

**Solution**: Extract to `ui/messages.js`, reducing game-ui.js dependencies.

## Proposed Refactoring

### Phase 1: Break Circular Dependencies

**Extract message system**:
- Create `ui/messages.js` with:
  - `showMessage()`
  - `showActionText()`
  - `addGameLog()`
- Update all files to import from `ui/messages.js` instead of `game-ui.js`
- Remove message functions from `game-ui.js`

**Result**: game-state.js no longer depends on game-ui.js

### Phase 2: Extract Display Functions

**Create display modules**:
- `ui/character-display.js` - updateCharacterDisplay(), updateStatsTab()
- `ui/inventory-display.js` - updateInventoryDisplay(), updateSpellsDisplay()
- `ui/location-display.js` - displayCurrentLocation(), playLocationMusic()
- `ui/npc-dialogue.js` - NPC dialogue functions
- `ui/vault-ui.js` - Vault UI functions
- `ui/ground-items-ui.js` - Ground items UI functions
- `ui/time-display.js` - Time display functions

**Result**: game-ui.js is now split into focused modules. Files can import only what they need.

### Phase 3: Consolidate Item Helpers

**Unify item fetching**:
- Keep `item-helpers.js` as single source
- Add `getItemById(itemId, cached = true)` to handle both cases
- Update all files to use unified function
- Remove duplicate implementations

**Result**: Single, clear way to get item data.

### New Dependency Graph (After Refactoring)

```
Core Layer (no dependencies)
  ├── session-manager.js
  ├── game-api.js
  └── auth.js

Utils Layer (depends on Core)
  ├── item-helpers.js (unified)
  ├── character-helpers.js
  ├── inventory-helpers.js
  └── ui-helpers.js (NEW - button creation, formatters)

Systems Layer (depends on Core + Utils)
  ├── game-state.js (no UI dependencies!)
  ├── game-logic.js
  ├── save-system.js
  ├── character-generator.js
  ├── inventory-interactions.js
  └── container-system.js

UI Layer (depends on Core + Utils + Systems)
  ├── messages.js (NEW - extracted)
  ├── character-display.js (NEW - extracted)
  ├── inventory-display.js (NEW - extracted)
  ├── location-display.js (NEW - extracted)
  ├── npc-dialogue.js (NEW - extracted)
  ├── vault-ui.js (NEW - extracted)
  ├── ground-items-ui.js (NEW - extracted)
  ├── time-display.js (NEW - extracted)
  └── combat-ui.js (NEW - extracted)

Pages Layer (depends on all layers)
  ├── game-intro.js (uses UI modules)
  ├── equipment-selection.js (uses utils)
  ├── new-game.js (uses UI modules)
  └── startup.js (orchestrates initialization)

Components Layer (minimal dependencies)
  ├── continue-button.js
  └── back-button.js
```

**Benefits**:
- ✅ No circular dependencies
- ✅ Clear layered architecture
- ✅ Each module has single responsibility
- ✅ Files import only what they need
- ✅ Easier to test and maintain
