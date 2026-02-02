# Go Codebase Refactor - TODO

## Completed This Session

### 1. Session Package Extraction ✅

- Created `server/session/` package
- Moved: GameSession, SessionManager, Delta types, snapshot logic
- `api/session_manager.go` now only has HTTP handlers
- `api/delta.go` now only has type aliases for backward compatibility

### 2. Character Generation ✅

- Created `server/game/character/generation.go`
- Moved deterministic char gen from `functions/discoverHero.go`
- `functions/` is now just a backward-compat shim (can delete later)
- `api/dnd.go` now imports `game/character`

### 3. World Package ✅

- Created `server/world/` package for transient server-side state
- Moved `merchant/state_manager.go` → `world/merchant.go`
- Updated `api/shop.go` to use `world.GetMerchantManager()` instead of `merchant.GetManager()`
- Deleted old `merchant/` package
- `world/doc.go` documents purpose: state that persists in memory but NOT in save files

**Future additions to world/** (not yet implemented):
- ground.go - Dropped items per location
- events.go - Temporary world events
- manager.go - Unified world state manager

### 4. Inventory Consolidation ✅

- **Deleted `api/inventory.go`** (2485 lines of dead code)
- Removed `/api/inventory/action` and `/api/inventory/actions` routes from `main.go`
- Frontend uses `/api/game/action` exclusively (via `src/lib/api.js`)
- **Moved container logic to `game/inventory/containers.go`** (514 lines)
- **Moved equipment logic to `game/inventory/equipment.go`** (636 lines)
- **Deleted `api/equipment.go`** and `api/containers.go`
- `api/game_actions.go` is now HTTP routing only - calls `game/inventory/` for logic

**Current `game/inventory/` structure (2086 lines total):**
- `inventory.go` - Use, drop, move, stack, split, add item actions
- `containers.go` - Add/remove items from containers (bags, pouches)
- `equipment.go` - Equip/unequip items

### 5. Shop Logic Extraction ✅

- Created `game/shop/pricing.go` (156 lines) with:
  - `CalculateBuyPrice()` - player buys from merchant
  - `CalculateSellPrice()` - merchant buys from player
  - `ParseIntervalToMinutes()` - time interval utility
- Created `game/inventory/add.go` (204 lines) with:
  - `AddItemToInventory()` - intelligent stacking and slot priority
  - `GetSlotQuantity()` - utility function
- `api/shop.go` now 620 lines (was 939) - HTTP handlers with thin wrappers

---

## Remaining Refactor Tasks

### 6. API Restructure (Optional)

**Goal:** Separate game API from server API (lower priority)

```
api/
├── game/           # Game-specific endpoints
│   ├── actions.go  # POST /api/game/action
│   ├── session.go  # /api/session/*
│   └── shop.go
├── data/           # Static data endpoints
│   ├── gamedata.go
│   └── character.go
└── saves.go
```

### 7. Cleanup

- ~~Delete `functions/` package (now just shims)~~ ✅ DONE
- ~~Fix `api/weights.go` to actually load from DB instead of hardcoded~~ ✅ DONE
- ~~Rework stat generation system~~ ✅ DONE
  - Weighted tier system: 79-81 (25%), 82-86 (65%), 87-89 (10%)
  - Class primary stats minimum raised to 15
  - Individual stats clamped 10-16
  - Deterministic distribution from npub hash

---

## Current Package Summary

**api/** (4704 lines) - HTTP handlers only:
- `game_actions.go` (757) - Routes game actions to `game/` packages
- `shop.go` (620) - Shop HTTP handlers (logic extracted)
- `session_manager.go` (423) - Session HTTP handlers
- `gamedata.go` (520) - Static data HTTP handlers
- `character-creation*.go` (1316) - Character creation handlers
- Others: saves, weights, profile, npcs, abilities, delta, dnd

**game/** - Game logic (2446+ lines):
- `game/inventory/` (2290 lines) - All inventory/equipment logic
  - `inventory.go` - use, drop, move, stack, split
  - `containers.go` - add/remove from containers
  - `equipment.go` - equip/unequip
  - `add.go` - add items with stacking
- `game/shop/` (156 lines) - Shop pricing logic
- `game/character/` - Character generation
- `game/movement/`, `game/status/`, `game/gametime/`, etc.

**session/** - Session state management
**world/** - Transient server-side state (merchant stock)

---

## Notes

### Character Generation (for later)

User wants to revisit skill/stat generation algorithm later.
Current flow:

1. api/dnd.go calls getWeightsFromDB() - HARDCODED, not from DB!
2. Calls character.GenerateCharacter() with weights
3. Uses deterministic RNG seeded from npub

The getWeightsFromDB() function lies - returns hardcoded data.
Should query `generation_weights` table instead.

### Go Report Card Goal

User wants good Go Report Card score after refactor.
Focus on:

- Package organization
- Clear separation of concerns
- Removing dead code
