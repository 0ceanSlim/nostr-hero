# Hunger System Implementation Plan

## Overview
Add hunger tracking (0-3 scale) with automatic time-based decay and ration consumption. When famished (hunger = 0), take -1 HP damage every 4 hours. Cannot eat when full.

---

## Phase 1: Data Structure Updates

### 1.1 Add Hunger Fields to Save File Backend
**File:** `src/api/saves.go`
- Add `Hunger int` field to `SaveFile` struct (after Fatigue)
- Add `HungerCounter int` field (tracks segments since last hunger decay)
- Add `HungerDamageCounter int` field (tracks segments since last HP damage when famished)
- Add JSON tags: `json:"hunger"`, `json:"hunger_counter"`, `json:"hunger_damage_counter"`
- Default hunger value: 1 (Hungry state)

### 1.2 Update Rations Item Definition
**File:** `docs/data/equipment/items/rations.json`
- Add `"consumable"` to `tags` array
- Add new `effects` field:
  ```json
  "effects": [
    {"type": "hunger", "value": 1},
    {"type": "fatigue", "value": -3}
  ]
  ```
- Change `type` from `"Adventuring Gear"` to `"Food"` (enables Use action)

---

## Phase 2: Frontend State & UI

### 2.1 Add Hunger to Game State
**File:** `www/scripts/systems/game-state.js`
- Add hunger, hunger_counter, hunger_damage_counter to state initialization
- Ensure values persist in save/load operations

### 2.2 Add Hunger Display to UI
**File:** `www/scripts/pages/game-ui.js`
- Add hunger display next to fatigue display
- Show icons/text for hunger levels:
  - 0 = Famished 💀
  - 1 = Hungry 🍖
  - 2 = Satisfied ✅
  - 3 = Full 🍗

---

## Phase 3: Hunger Decay Logic

### 3.1 Implement Time-Based Hunger Decay
**File:** `www/scripts/systems/game-logic.js` (in time advancement function)

Add hunger tracking alongside existing fatigue:
```javascript
// Hunger decay
let hungerCounter = newCharacterState.hunger_counter || 0;
let currentHunger = newCharacterState.hunger || 1;

hungerCounter += 1;

// Decay rates: 3→2 or 2→1 every 3 segments (6h), 1→0 every 6 segments (12h)
if (currentHunger > 1 && hungerCounter >= 3) {
    currentHunger -= 1;
    hungerCounter = 0;
    showMessage('🍖 You feel hungry', 'warning');
} else if (currentHunger === 1 && hungerCounter >= 6) {
    currentHunger = 0;
    hungerCounter = 0;
    showMessage('💀 You are famished!', 'error');
}

newCharacterState.hunger = currentHunger;
newCharacterState.hunger_counter = hungerCounter;
```

### 3.2 Implement Famished HP Damage
**File:** `www/scripts/systems/game-logic.js` (in time advancement function)

Add HP damage when hunger = 0:
```javascript
// HP damage when famished (every 2 segments = 4 hours)
if (currentHunger === 0) {
    let damageCounter = newCharacterState.hunger_damage_counter || 0;
    damageCounter += 1;

    if (damageCounter >= 2) {
        newCharacterState.hp = Math.max(0, newCharacterState.hp - 1);
        damageCounter = 0;
        showMessage('💀 Starvation damages you! (-1 HP)', 'error');
    }

    newCharacterState.hunger_damage_counter = damageCounter;
} else {
    // Reset damage counter if not famished
    newCharacterState.hunger_damage_counter = 0;
}
```

---

## Phase 4: Item Effect System

### 4.1 Backend: Apply Item Effects on Use
**File:** `src/api/inventory.go` (replace TODO at line 428 in handleUseItem)

Implement generic effect system:
1. Query item from database to get properties JSON
2. Parse `effects` array from properties
3. **Check if hunger effect would exceed max (3)** - if so, return error "You're too full to eat"
4. Apply each effect:
   - `"hunger"` → Modify `save.Hunger` (clamp 0-3)
   - `"fatigue"` → Modify `save.Fatigue` (clamp 0-9)
   - `"hp"` → Modify `save.HP` (clamp 0-max_hp)
   - `"mana"` → Modify `save.Mana` (clamp 0-max_mana)
5. Return success with effects applied message

### 4.2 Frontend: Remove Hardcoded Rations Logic
**File:** `www/scripts/systems/game-logic.js`

Remove hardcoded rations check (lines 214-218) - backend now handles all effects generically.

---

## Phase 5: Validation & Limits

### 5.1 Hunger/Fatigue/HP Clamping
Ensure all updates clamp values correctly:
- **Hunger**: 0-3 (can't go below 0 or above 3)
- **Fatigue**: 0-9
- **HP**: 0-max_hp
- **Mana**: 0-max_mana

### 5.2 Prevent Eating When Full
**File:** `src/api/inventory.go` (in handleUseItem, before applying effects)

Check if item has hunger effect and current hunger = 3:
- If so, return error: `"You're too full to eat"`
- Prevents consuming the item
- Player cannot waste rations when already full

---

## Testing Checklist

1. ✅ New character starts with hunger = 1
2. ✅ Hunger decreases: 3→2 (3 segments), 2→1 (3 segments), 1→0 (6 segments)
3. ✅ Eating rations: hunger +1, fatigue -3
4. ✅ Cannot eat when hunger = 3 (shows error message)
5. ✅ Famished: -1 HP every 2 segments (4 hours)
6. ✅ HP damage stops when hunger increases above 0
7. ✅ UI displays hunger correctly
8. ✅ Save/load preserves hunger values
9. ✅ Counters reset properly after decay/damage

---

## Files to Modify

**Backend (Go):**
- `src/api/saves.go` - Add Hunger, HungerCounter, HungerDamageCounter fields
- `src/api/inventory.go` - Implement item effects + full check in handleUseItem

**Frontend (JavaScript):**
- `www/scripts/systems/game-logic.js` - Hunger decay, HP damage, remove hardcoded rations
- `www/scripts/pages/game-ui.js` - Hunger display
- `www/scripts/systems/game-state.js` - Add hunger fields to state

**Data (JSON):**
- `docs/data/equipment/items/rations.json` - Add effects, change type to "Food"

---

## Summary
Simple hunger system with automatic decay, ration consumption for restoration, starvation damage at hunger = 0, and prevention of eating when full (hunger = 3).

---

## Hunger Mechanics Reference

### Hunger Scale
| Level | Status    | Description                           |
|-------|-----------|---------------------------------------|
| 0     | Famished  | -1 HP every 4 hours (2 segments)      |
| 1     | Hungry    | Default state after sleep/start       |
| 2     | Satisfied | Comfortable state                     |
| 3     | Full      | Cannot eat more food                  |

### Hunger Decay Rates
- **3 → 2**: Every 3 segments (6 hours)
- **2 → 1**: Every 3 segments (6 hours)
- **1 → 0**: Every 6 segments (12 hours)

### Ration Effects
- **Hunger**: +1 (cannot exceed 3)
- **Fatigue**: -3 (cannot go below 0)

### Time Segments (0-11)
| Segment | Time     | Notes          |
|---------|----------|----------------|
| 0       | Midnight | Start of day   |
| 1       | 2 AM     |                |
| 2       | 4 AM     |                |
| 3       | 6 AM     | Dawn           |
| 4       | 8 AM     |                |
| 5       | 10 AM    |                |
| 6       | Noon     | High noon      |
| 7       | 2 PM     |                |
| 8       | 4 PM     |                |
| 9       | 6 PM     | Evening        |
| 10      | 8 PM     |                |
| 11      | 10 PM    |                |
