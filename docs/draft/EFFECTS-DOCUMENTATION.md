# Effects System Documentation

## Overview

The effects system provides a flexible way to apply temporary or permanent stat modifiers to characters. Effects are defined as JSON templates in `game-data/effects/` and stored in the database via migration.

## Effect Structure

```json
{
  "id": "effect-id",
  "name": "Display Name",
  "description": "What this effect does",
  "effects": [
    {
      "type": "stat-type",
      "value": 2,
      "duration": 720,
      "delay": 0,
      "tick_interval": 60
    }
  ],
  "icon": "icon-name",
  "color": "color-name"
}
```

### Effect Properties

- **id**: Unique identifier (kebab-case)
- **name**: Display name shown to players
- **description**: Explanation of what the effect does
- **effects**: Array of individual stat modifiers
  - **type**: Stat to modify (`strength`, `dexterity`, `constitution`, `intelligence`, `wisdom`, `charisma`, `hp`, `mana`, `fatigue`, `hunger`)
  - **value**: Amount to modify (positive = buff, negative = debuff)
  - **duration**: How long the effect lasts in minutes (0 = permanent until removed)
  - **delay**: Minutes before effect starts (optional, default 0)
  - **tick_interval**: How often periodic effects trigger in minutes (optional, for damage/healing over time)
- **icon**: Icon identifier for UI display
- **color**: Color for UI display (`green`, `yellow`, `orange`, `red`, `purple`, `blue`)

## Save File Format

Effects are stored in save files using a compact format that only includes runtime state:

```json
{
  "active_effects": [
    {
      "effect_id": "performance-high",
      "effect_index": 0,
      "duration_remaining": 650.5,
      "delay_remaining": 0,
      "tick_accumulator": 0,
      "applied_at": 720
    }
  ]
}
```

Template data (name, description, stat type, base value) is looked up from the database when needed. This keeps save files small and allows effect definitions to be updated without invalidating saves.

## System Effects

### Fatigue System (Replaced Fields: `fatigue`, `fatigue_counter`)

Fatigue is now tracked via effects instead of dedicated fields:

- **fatigue-level-1**: Minor tiredness, -1 DEX
- **fatigue-level-2**: Getting tired, -2 DEX, -1 CON
- **fatigue-level-3**: Very tired, -3 DEX, -2 CON, -1 WIS

Fatigue increases every 240 minutes (4 hours) of activity. Sleeping removes all fatigue effects.

### Hunger System (Replaced Fields: `hunger`, `hunger_counter`)

Hunger is now tracked via effects:

- **hunger-famished** (Level 0): Starving, -3 STR, -3 DEX, -2 CON, -2 INT
- **hunger-hungry** (Level 1): Hungry, -1 STR, -1 CON
- *(Level 2: Satisfied - no effect, baseline)*
- **hunger-full** (Level 3): Well fed, +1 CON

Hunger decreases over time based on current level:
- Full (3): Decreases every 720 minutes (12 hours)
- Satisfied (2): Decreases every 480 minutes (8 hours)
- Hungry (1): Decreases every 360 minutes (6 hours)
- Famished (0): Stays at 0

### Exhaustion Effects

- **exhaustion**: Severe exhaustion, massive penalties to all stats (-4 STR/DEX, -3 CON, -2 INT/WIS/CHA)
- **overwright**: Magical exhaustion, balanced penalties to all physical and mental stats (-3 to five stats)

These effects represent extreme states beyond normal fatigue and may be applied by specific game events (combat, magic use, environmental hazards).

## Performance Effects

- **performance-high**: +2 CHA for 720 minutes (12 hours) after successful performance
- **stage-fright**: -1 CHA for 720 minutes (12 hours) after failed performance

## Status Effects

- **poison**: Periodic damage over time (example of tick_interval usage)
- **drunk**: Temporary stat penalties from alcohol

## Usage in Code

### Applying Effects

```go
// Apply an effect by ID
err := applyEffect(state, "performance-high")
```

### Checking Active Effects

```go
// Get current fatigue level
level := getCurrentFatigueLevel(state)

// Get current hunger level
level := getCurrentHungerLevel(state)
```

### Removing Effects

```go
// Remove all fatigue effects
removeFatigueEffects(state)

// Remove all hunger effects
removeHungerEffects(state)
```

### Time-Based Ticking

```go
// Tick all active effects
tickEffects(state, minutesElapsed)

// Tick fatigue system
tickFatigue(state, minutesElapsed)

// Tick hunger system
tickHunger(state, minutesElapsed)
```

## Stats Tab Display Format

Effects should be displayed in the stats tab grouped by type:

### Active Buffs (Positive Effects)
```
🟢 Performance High - 8h 32m remaining
   +2 Charisma
🟢 Well Fed - Permanent
   +1 Constitution
```

### Active Debuffs (Negative Effects)
```
🟠 Fatigue (Level 2) - Permanent
   -2 Dexterity, -1 Constitution
🔴 Hungry - Permanent
   -1 Strength, -1 Constitution
```

### Display Format Rules

1. **Color Coding**:
   - 🟢 Green: Positive effects (buffs)
   - 🟡 Yellow: Minor negative effects
   - 🟠 Orange: Moderate negative effects
   - 🔴 Red: Severe negative effects

2. **Duration Display**:
   - Permanent effects (duration = 0): Show "Permanent"
   - Timed effects: Show remaining time in hours and minutes (e.g., "8h 32m remaining")
   - Effects with delay: Show "Starts in Xh Ym"

3. **Stat Modifiers**:
   - Show all stat changes in a single line
   - Use +/- notation clearly
   - Group multiple modifiers with commas

4. **Grouping**:
   - System effects (fatigue, hunger) shown first
   - Buffs (positive effects) shown next
   - Debuffs (negative effects) shown last
   - Within each group, sort by severity (most impactful first)

## Creating New Effects

1. Create JSON file in `game-data/effects/`
2. Follow naming convention: `effect-name.json` (kebab-case)
3. Define all required fields (id, name, description, effects array)
4. Run migration to load into database: `cd game-data && go run migrate.go`
5. Test effect application in-game

## Migration

Effects are migrated from JSON to SQLite on server startup via `game-data/CODEX/migration/migration.go`. The migration:

1. Reads all JSON files from `game-data/effects/`
2. Clears the `effects` table
3. Inserts effect data into database

To force a re-migration, delete `www/game.db` and restart the server, or run `cd game-data && go run migrate.go`.

## Backend Implementation

Effects logic is primarily in `server/api/game_actions.go`:

- `applyEffect()` - Apply effect template to character
- `tickEffects()` - Process all active effects over time
- `getEffectTemplate()` - Load effect template from database
- `getActiveStatModifiers()` - Calculate total stat bonuses from effects
- Fatigue management: `getCurrentFatigueLevel()`, `updateFatigueLevel()`, `removeFatigueEffects()`, `tickFatigue()`
- Hunger management: `getCurrentHungerLevel()`, `updateHungerLevel()`, `removeHungerEffects()`, `tickHunger()`

## Frontend Integration

Effects are managed entirely on the backend. The frontend should:

1. Display active effects in the stats tab
2. Show effect icons/indicators on the character sheet
3. Update displays when `active_effects` changes in game state

Effect template data (names, descriptions, icons) should be fetched from the backend API, not hardcoded in frontend.
