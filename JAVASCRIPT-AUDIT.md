# JavaScript Cleanup Audit
**Date**: 2025-12-15
**Status**: In Progress

## Overview

Total JavaScript: **12,596 lines** across 24 files
Top 7 files account for **8,667 lines (69%)** of all code

## File Size Analysis

### Large Files (Need Splitting)
| File | Lines | Purpose | Issues |
|------|-------|---------|--------|
| `pages/game-ui.js` | 1,911 | Main UI rendering | Too many responsibilities, should be split |
| `pages/game-intro.js` | 1,679 | Character intro sequence | Has duplicate functions (see below) |
| `systems/equipment-selection.js` | 1,638 | Equipment choice UI | Has duplicate functions (see below) |
| `systems/inventory-interactions.js` | 1,393 | Inventory drag/drop | Could be optimized |
| `core/session-manager.js` | 713 | Authentication/session | Reasonable size |
| `pages/new-game.js` | 667 | New game screen | Reasonable size |
| `systems/game-state.js` | 666 | State management | Reasonable size |

## Critical Issues Found

### 1. Duplicate Functions

#### `getItemImageName()` - EXACT DUPLICATE
- **Location 1**: `systems/equipment-selection.js:1397`
- **Location 2**: `pages/game-intro.js:273`
- **Action**: Extract to `utils/item-helpers.js`

#### `getItemStats()` - EXACT DUPLICATE
- **Location 1**: `systems/equipment-selection.js:1526`
- **Location 2**: `pages/game-intro.js:621`
- **Action**: Extract to `utils/item-helpers.js`

#### `getItemById()` - MULTIPLE VERSIONS
- **Version 1**: `utils/item-helpers.js:37` (async, fetches from database)
- **Version 2**: `systems/game-state.js:221` (sync, reads from cached DOM)
- **Version 3**: `pages/game-ui.js:54` (async variant called `getItemByIdAsync`)
- **Action**: Consolidate into single helper with optional async mode

### 2. Function Responsibilities in game-ui.js (1,911 lines)

#### Message/Log System (Lines 11-103)
- `addGameLog()` - Add message to log
- `showActionText()` - Colored log messages
- `showMessage()` - Toast notifications

#### Character Display (Lines 117-787)
- `calculateMaxCapacity()` - Weight calculations
- `calculateAndDisplayWeight()` - Weight display
- `updateCharacterDisplay()` - Main character UI update (400+ lines!)
- `updateStatsTab()` - Stats tab rendering

#### Inventory/Spells Display (Lines 788-919)
- `updateInventoryDisplay()` - Render inventory
- `updateSpellsDisplay()` - Render spells

#### Location System (Lines 920-1456)
- `playLocationMusic()` - Music management
- `displayCurrentLocation()` - Location rendering (250+ lines!)
- `createActionButton()` - Button creation utility
- `createLocationButton()` - Location button utility
- `isBuildingOpen()` - Building hours check
- `showBuildingClosedMessage()` - Closed building message
- `enterBuilding()` - Building entry logic
- `exitBuilding()` - Building exit logic

#### NPC/Dialogue System (Lines 1300-1456)
- `talkToNPC()` - Initiate NPC dialogue
- `showNPCDialogue()` - Dialogue UI
- `formatDialogueOption()` - Format dialogue choices
- `selectDialogueOption()` - Handle dialogue choice
- `closeNPCDialogue()` - Close dialogue

#### Vault System (Lines 1457-1586)
- `showVaultUI()` - Display vault interface
- `createVaultSlot()` - Create vault slot elements
- `closeVaultUI()` - Close vault

#### Ground Items System (Lines 1726-1906)
- `openGroundModal()` - Show ground items modal
- `closeGroundModal()` - Hide ground items modal
- `refreshGroundModal()` - Update ground items display
- `pickupGroundItem()` - Pick up item from ground

#### Combat System (Lines 1588-1596)
- `updateCombatInterface()` - Combat UI (stub)

#### Shop/Tavern (Lines 1597-1608)
- `openShop()` - Shop interface (stub)
- `openTavern()` - Tavern interface (stub)

#### Time System (Lines 1609-1661)
- `updateTimeDisplay()` - Time/fatigue display
- `formatTime()` - Time formatting

#### Utility (Lines 1663-1716)
- `updateAllDisplays()` - Update all UI elements
- `saveGameToRelay()` - Save to Nostr relay (stub)
- `loadAdvancementData()` - Load advancement data

## Proposed Reorganization

### New Module Structure

```
www/scripts/
├── core/                          # Core systems (no changes)
│   ├── session-manager.js         # 713 lines ✅
│   ├── auth.js                    # 647 lines ✅
│   ├── game-api.js                # 224 lines ✅
│   └── nostr-connect.js           # 456 lines ✅
│
├── utils/                         # Shared utilities
│   ├── item-helpers.js            # 73 lines → Expand
│   │   + Add: getItemImageName(), getItemStats()
│   │   + Add: unified getItemById() function
│   ├── character-helpers.js       # 109 lines ✅
│   ├── inventory-helpers.js       # 338 lines ✅
│   └── ui-helpers.js              # NEW - Button creation, formatters
│       + createActionButton()
│       + createLocationButton()
│       + formatDialogueOption()
│       + formatTime()
│
├── systems/                       # Game systems
│   ├── game-state.js              # 666 lines → Reduce
│   │   - Remove: getItemById() (move to utils)
│   ├── game-logic.js              # 154 lines ✅
│   ├── save-system.js             # 45 lines ✅
│   ├── character-generator.js     # 447 lines ✅
│   ├── inventory-interactions.js  # 1,393 lines → Review
│   ├── container-system.js        # 396 lines ✅
│   ├── profile-manager.js         # 214 lines ✅
│   ├── relay-manager.js           # 194 lines ✅
│   └── theme-manager.js           # 153 lines ✅
│
├── ui/                            # NEW - UI rendering modules
│   ├── messages.js                # NEW - Extracted from game-ui.js
│   │   + addGameLog()
│   │   + showActionText()
│   │   + showMessage()
│   ├── character-display.js       # NEW - Extracted from game-ui.js
│   │   + updateCharacterDisplay()
│   │   + updateStatsTab()
│   │   + calculateMaxCapacity()
│   │   + calculateAndDisplayWeight()
│   ├── inventory-display.js       # NEW - Extracted from game-ui.js
│   │   + updateInventoryDisplay()
│   │   + updateSpellsDisplay()
│   ├── location-display.js        # NEW - Extracted from game-ui.js
│   │   + displayCurrentLocation()
│   │   + playLocationMusic()
│   │   + enterBuilding()
│   │   + exitBuilding()
│   │   + isBuildingOpen()
│   │   + showBuildingClosedMessage()
│   ├── npc-dialogue.js            # NEW - Extracted from game-ui.js
│   │   + talkToNPC()
│   │   + showNPCDialogue()
│   │   + selectDialogueOption()
│   │   + closeNPCDialogue()
│   ├── vault-ui.js                # NEW - Extracted from game-ui.js
│   │   + showVaultUI()
│   │   + createVaultSlot()
│   │   + closeVaultUI()
│   ├── ground-items-ui.js         # NEW - Extracted from game-ui.js
│   │   + openGroundModal()
│   │   + closeGroundModal()
│   │   + refreshGroundModal()
│   │   + pickupGroundItem()
│   ├── time-display.js            # NEW - Extracted from game-ui.js
│   │   + updateTimeDisplay()
│   └── combat-ui.js               # NEW - Extracted from game-ui.js (stub)
│       + updateCombatInterface()
│
├── pages/                         # Page-specific logic
│   ├── game-intro.js              # 1,679 lines → Reduce
│   │   - Remove: getItemImageName(), getItemStats() (use utils)
│   ├── equipment-selection.js     # 1,638 lines → Reduce (move to systems/)
│   │   - Remove: getItemImageName(), getItemStats() (use utils)
│   ├── new-game.js                # 667 lines ✅
│   ├── startup.js                 # 302 lines ✅
│   └── tabs.js                    # 124 lines ✅
│
└── components/                    # Reusable components
    ├── continue-button.js         # 35 lines ✅
    └── back-button.js             # 18 lines ✅
```

### Load Order (game.html)

**New proposed order:**
```html
<!-- Core -->
<script src="/scripts/core/session-manager.js"></script>
<script src="/scripts/core/game-api.js"></script>
<script src="/scripts/core/auth.js"></script>

<!-- Utils (shared helpers) -->
<script src="/scripts/utils/item-helpers.js"></script>
<script src="/scripts/utils/character-helpers.js"></script>
<script src="/scripts/utils/inventory-helpers.js"></script>
<script src="/scripts/utils/ui-helpers.js"></script>

<!-- Systems (game logic) -->
<script src="/scripts/systems/game-state.js"></script>
<script src="/scripts/systems/game-logic.js"></script>
<script src="/scripts/systems/save-system.js"></script>
<script src="/scripts/systems/character-generator.js"></script>
<script src="/scripts/systems/container-system.js"></script>
<script src="/scripts/systems/inventory-interactions.js"></script>

<!-- UI (rendering modules) -->
<script src="/scripts/ui/messages.js"></script>
<script src="/scripts/ui/character-display.js"></script>
<script src="/scripts/ui/inventory-display.js"></script>
<script src="/scripts/ui/location-display.js"></script>
<script src="/scripts/ui/npc-dialogue.js"></script>
<script src="/scripts/ui/vault-ui.js"></script>
<script src="/scripts/ui/ground-items-ui.js"></script>
<script src="/scripts/ui/time-display.js"></script>
<script src="/scripts/ui/combat-ui.js"></script>

<!-- Page initialization -->
<script src="/scripts/pages/startup.js"></script>
```

## Action Plan

### Phase 1: Extract Duplicates (CURRENT)
- [x] Delete unused `pages/login.js`
- [ ] Extract `getItemImageName()` to `utils/item-helpers.js`
- [ ] Extract `getItemStats()` to `utils/item-helpers.js`
- [ ] Update `game-intro.js` to use utils version
- [ ] Update `equipment-selection.js` to use utils version
- [ ] Consolidate `getItemById()` variants

### Phase 2: Split game-ui.js
- [ ] Create `ui/` directory
- [ ] Extract messages module (100 lines)
- [ ] Extract character-display module (670 lines)
- [ ] Extract inventory-display module (130 lines)
- [ ] Extract location-display module (250 lines)
- [ ] Extract npc-dialogue module (156 lines)
- [ ] Extract vault-ui module (129 lines)
- [ ] Extract ground-items-ui module (180 lines)
- [ ] Extract time-display module (52 lines)
- [ ] Extract combat-ui module (9 lines)
- [ ] Delete original game-ui.js

### Phase 3: Create ui-helpers.js
- [ ] Extract button creation utilities
- [ ] Extract formatting utilities

### Phase 4: Update HTML Files
- [ ] Update game.html with new script paths
- [ ] Update game-intro.html if needed
- [ ] Update new-game.html if needed
- [ ] Test all pages load correctly

### Phase 5: Clean Up Remaining Files
- [ ] Review game-intro.js for more cleanup
- [ ] Review equipment-selection.js (consider moving to systems/)
- [ ] Review inventory-interactions.js for optimization

## Expected Results

- **game-ui.js**: 1,911 lines → **DELETED** (split into 9 modules)
- **game-intro.js**: 1,679 lines → ~1,550 lines (-129 lines from duplicates)
- **equipment-selection.js**: 1,638 lines → ~1,510 lines (-128 lines from duplicates)
- **item-helpers.js**: 73 lines → ~330 lines (+257 from consolidation)
- **Total reduction**: From 12,596 to ~11,800 lines (-796 lines)
- **Better organization**: Clear separation of concerns
- **No more duplicates**: Single source of truth for shared functions

## Risk Assessment

**Low Risk**: Extracting duplicate utility functions
**Medium Risk**: Splitting game-ui.js (many functions exported to window)
**High Risk**: Changing load order could break dependencies

**Mitigation**:
- Test after each extraction
- Keep window exports consistent
- Update one HTML file at a time
