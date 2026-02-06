# NPC Schedule & Environment System

**Status**: Draft/Workshop
**Created**: 2026-02-05
**Related**: 144x-time-progression.md, locations-plan.md

## Overview

This document outlines new systems for time management during environment traversal, environment encounters, and dungeon discovery mechanics.

---

## 1. Time Movement & Control System

### Time States

**Entering Environment**:

- Time PAUSES automatically
- Player sees environment description and directional options
- No time passes until player makes movement decision

**Traversing Environment** (Moving between locations):

- Player chooses from 2 cardinal directions (N/S or E/W depending on environment)
- Progress bar shows advancement toward destination
- Time moves at 144x speed during traversal
- Travel time modified by: terrain, weather, encumbrance
- Player can: stop, continue forward, or turn back
- Encounter checks happen during travel (see section 3)

**In Combat/Encounters**:

- Time pauses during combat resolution
- Small time increment after combat ends

---

## 3. Environment Encounter System

### Environment Properties

Each environment has traversal settings, encounter rates, and difficulty settings:

```json
{
  "environment_id": "dark_forest",
  "name": "Dark Forest",
  "navigation": {
    "axis": "north_south",
    "north_destination": "mountain_pass",
    "south_destination": "kingdom",
    "base_travel_time_minutes": 45,
    "terrain_modifier": 1.2,
    "weather_effects": {
      "clear": 1.0,
      "rain": 1.3,
      "storm": 1.5
    }
  },
  "encounter_rates": {
    "day": {
      "monsters_per_hour": 2.5,
      "difficulty_multiplier": 1.0
    },
    "night": {
      "monsters_per_hour": 4.0,
      "difficulty_multiplier": 1.5
    }
  },
  "available_monsters": ["wolf", "goblin", "spider"],
  "terrain_type": "forest"
}
```

**Navigation Axis**:

- `north_south`: Player can travel North or South
- `east_west`: Player can travel East or West
- Only 2 directions available per environment

**Travel Time Calculation**:

```
actual_time = base_travel_time_minutes
              × terrain_modifier
              × weather_modifier
              × encumbrance_modifier
```

**Player Options During Traversal**:

- Continue toward destination (progress increases)
- Stop at current position (can discover dungeons here)
- Turn back toward origin (progress decreases)

### Encounter Mechanics

**Encounter Probability**:

- Calculate based on time spent traveling in environment
- Formula: `encounter_chance = (monsters_per_hour * hours_traveled) / random_factor`
- Check day vs night based on current game time
- Difficulty multiplier affects monster level/stats

**Day/Night Determination**:

- Day: 06:00 - 18:00
- Night: 18:00 - 06:00
- Uses current game time from time progression system

**Encounter Resolution**:

1. Roll for encounter during travel
2. If encounter triggers:
   - Pause time
   - Select random monster from `available_monsters`
   - Apply difficulty multiplier to monster stats
   - Enter combat
3. After combat, time resumes

### Implementation Notes

- Add encounter data to environment JSON files
- Calculate encounters when processing movement action
- Time-of-day affects both rate and difficulty
- Difficulty multiplier scales monster HP, damage, XP

---

## 4. Dungeon Discovery System (change to point of interest with dungeons being a primary poi but also will cover other utilities in unique pois since they can be revisited once discovered)

### Concept

Dungeons are special locations positioned at specific intervals within environments. They have static mobs and linear event sequences. Once discovered, they can be revisited by stopping at that interval during traversal.

### Dungeon Properties

```json
{
  "dungeon_id": "abandoned_mine",
  "name": "Abandoned Mine",
  "parent_environment": "mountain_pass",
  "location": {
    "direction": "south",
    "distance_minutes": 30,
    "description": "30 minutes south into the mountain pass"
  },
  "access_requirements": [
    {
      "item_id": "rope-hempen-50-feet",
      "consumed": false,
      "description": "Requires rope to descend into the mine"
    }
  ],
  "discovery": {
    "base_chance": 0.2,
    "level_modifier": 0.05,
    "max_chance": 0.95,
    "calculation": "base_chance + (player_level × level_modifier)"
  },
  "completion_effect": {
    "type": "none",
    "description": "No traversal progression"
  },
  "layout": {
    "type": "linear",
    "length": 5,
    "encounters": [
      {
        "step": 1,
        "type": "monster",
        "enemy": "kobold",
        "count": 2,
        "static": true
      },
      {
        "step": 3,
        "type": "treasure",
        "item": "iron_ore",
        "quantity": 3
      },
      {
        "step": 5,
        "type": "boss",
        "enemy": "kobold_chief",
        "reward": {
          "item": "rusty_key",
          "location_unlock": "secret_cache"
        },
        "static": true
      }
    ]
  }
}
```

**Access Requirements Example** (Mountain Cave with progression):

```json
{
  "dungeon_id": "mountain_cave_passage",
  "name": "Mountain Cave Passage",
  "parent_environment": "mountain_pass",
  "location": {
    "direction": "north",
    "distance_minutes": 20,
    "description": "20 minutes north into the mountain pass"
  },
  "access_requirements": [
    {
      "item_id": "torch",
      "consumed": true,
      "description": "Need a torch to see in the dark cave"
    }
  ],
  "completion_effect": {
    "type": "traversal_progress",
    "amount_minutes": 40,
    "description": "Shortcut through the mountain, saves 40 minutes of travel"
  },
  "layout": {
    "type": "linear",
    "length": 3,
    "encounters": [
      {
        "step": 2,
        "type": "monster",
        "enemy": "bat",
        "count": 3,
        "static": true
      }
    ]
  }
}
```

**Location Within Environment**:

- `direction`: Which direction from origin (north, south, east, west)
- `distance_minutes`: How many minutes of travel from origin
- Dungeons exist at specific intervals along the environment path

**Access Requirements**:

- `item_id`: Required item in player's inventory
- `consumed`: If true, item is consumed on entry; if false, item just needs to be present
- `description`: Flavor text explaining why item is needed
- Player must have ALL required items to enter
- If requirements not met, show message explaining what's needed

**Completion Effects**:

- `type: "none"`: No special effect, just loot/XP from encounters
- `type: "traversal_progress"`: Advances player position in environment
  - `amount_minutes`: How many minutes forward the player is moved
  - Useful for shortcuts, passages through mountains, etc.
  - Player emerges from dungeon further along the path
- Other types could be added: location unlocks, story flags, etc.

**Discovery Calculation**:

```
discovery_chance = base_chance + (player_level × level_modifier)
discovery_chance = min(discovery_chance, max_chance)
```

Example:

- Base chance: 20%
- Level modifier: 5% per level
- Player level 3: 20% + (3 × 5%) = 35% chance
- Player level 15: 20% + (15 × 5%) = 95% (capped at max_chance)

### Discovery Flow

1. **During Environment Traversal**:
   - Player stops at a position during travel
   - System checks if player is within dungeon interval (±2 minutes)
   - If position matches dungeon location and not yet discovered:
     - Roll for discovery based on player level
     - If successful: Show discovery message, mark as discovered

2. **Player Choice After Discovery**:
   - "Enter Now" - Check access requirements, enter if met
   - "Mark and Continue" - Save location, continue travel
   - Dungeon is permanently marked on this environment path

3. **Revisiting Dungeons**:
   - During traversal, if player stops at discovered dungeon's interval
   - "Enter [Dungeon Name]" option appears
   - Access requirements checked on every entry attempt
   - Can be entered multiple times (static encounters respawn)

4. **Dungeon Entry**:
   - Time pauses
   - Player progresses step-by-step through linear layout
   - Each step presents encounter or event
   - Static encounters (marked `"static": true`) always spawn
   - No backtracking in linear dungeons

5. **Dungeon Completion**:
   - Add dungeon_id to locations_discovered (if not already there)
   - Apply completion effect:
     - **none**: Return to same position in environment
     - **traversal_progress**: Advance player position by amount_minutes
   - Time resumes from new position

### Dungeon State Tracking

Track in save file (same field as other locations):

```json
{
  "locations_discovered": [
    "kingdom",
    "dark_forest",
    "abandoned_mine",
    "mountain_cave_passage"
  ],
  "current_environment_position": {
    "environment_id": "mountain_pass",
    "position_minutes": 15,
    "direction": "south",
    "origin_location": "kingdom"
  }
}
```

**Discovery Tracking**:

- Dungeons are stored as simple IDs in `locations_discovered` array
- No separate dungeon tracking structure needed
- Dungeon position and requirements defined in dungeon JSON file
- Check if dungeon ID exists in `locations_discovered` to determine if discovered

### Revisiting Dungeons

- When player stops during traversal, check if current position matches any dungeon's location
- Cross-reference with `locations_discovered` to see if dungeon is known
- If discovered: Show "Enter [Dungeon Name]" option
- If not discovered: Roll for discovery
- Can be run multiple times (static encounters always respawn)
- Check access requirements on each entry attempt

### Implementation Notes

- Track player position during environment traversal (minutes from origin)
- Load all dungeons for current environment from JSON files
- Check dungeon positions when player stops (±2 minute tolerance)
- Discovery roll only happens if dungeon not in `locations_discovered`
- On discovery, add dungeon_id to `locations_discovered` array
- Each environment can have multiple dungeons at different intervals
- Linear layout means simple progression: step 1 → step 2 → ... → end
- Static mobs always respawn, dynamic loot may or may not respawn
- Access requirements checked before allowing entry

---

## 5. Integration Points

### With Existing Systems

**144x Time Progression**:

- Pause/play controls work with existing time advancement
- Environment traversal uses same time calculation
- NPC schedules read from same game time

**Combat System**:

- Encounters trigger existing combat flow
- Difficulty multipliers affect monster stats
- Dungeon bosses use same combat mechanics

**Save System**:

- Add `discovered_dungeons` array (thjis to change, i think I have a general discoveries array taht will track them all nested under that property)

**Location System**:

- Environments reference locations as start/end points
- Dungeons reference parent environment

---

## 6. Open Questions / Workshop Items

### Time Mechanics

- Should weather be dynamic or based on time/location? both. it should be synamic but also defined by location. ie blizzard in the artic.

### Environment Traversal

- Should player see exact percentage progress or vague descriptions? exact percentage, like a progress bar

### Encounters

- Should encounter rate decrease if player is much higher level? no env has same general encounter, general encounter mob scales with player level, points of interest can have statically defined mobs
- Should player be able to flee from encounters? yes, combat notes required, I wrote extensively about combat at some point
- Should repeated encounters in same environment give reduced XP? nah, I hope it's dynamic enough to not repeat too often

### Dungeons

- ✅ **RESOLVED**: Dungeons at specific intervals (e.g., 30 min south)
- ✅ **RESOLVED**: Discovery based on player level formula
- ✅ **RESOLVED**: Static mobs always respawn
- ✅ **RESOLVED**: Access requirements (items needed, consumed or not)
- ✅ **RESOLVED**: Completion effects (none vs traversal_progress)
- ✅ **RESOLVED**: State tracked in same field as locations (simple IDs)
- Should dungeons have consumable loot that doesn't respawn?
- How often can dungeons be re-run (cooldown)?
- Should dungeon difficulty scale with player level?
- Should there be branching paths in some dungeons (non-linear)?

---

## 7. Future Enhancements

### Possible Additions

- **Dynamic Events**: Random events during travel (merchant on road, ambush, etc.)
- **Camping**: Ability to make camp during travel to rest/recover
- **Fast Travel**: Unlock fast travel between discovered locations
- **Environmental Hazards**: Weather, terrain effects on travel
- **Dungeon Maps**: Show progress through dungeon visually
- **Multiplayer Dungeons**: Party-based dungeon runs
- **Procedural Dungeons**: Generate random layouts instead of fixed
- **Time-of-day Dungeons**: Some dungeons only accessible at certain times

---

## Implementation Checklist

### Phase 3: Environment Navigation System

- [ ] Add navigation data to environment JSON (axis, destinations, base_travel_time)
- [ ] Implement 2-direction navigation (N/S or E/W only)
- [ ] Create progress bar UI for traversal
- [ ] Track player position in environment (minutes from origin)
- [ ] Add stop/continue/turn back options during traversal
- [ ] Implement travel time calculation (base × terrain × weather × encumbrance)
- [ ] Time advances at 144x during traversal

### Phase 4: Environment Encounters

- [ ] Add encounter data to environment JSON files
- [ ] Implement encounter probability during traversal
- [ ] Add day/night detection (6am-6pm vs 6pm-6am)
- [ ] Implement difficulty multipliers for night encounters
- [ ] Connect encounters to combat system
- [ ] Pause time when encounter triggers

### Phase 5: Dungeon Discovery & Positioning

- [ ] Create dungeon JSON schema with location intervals
- [ ] Add access_requirements field (item_id, consumed, description)
- [ ] Add completion_effect field (type: none | traversal_progress)
- [ ] Implement position-based discovery (±2 minute tolerance)
- [ ] Implement level-based discovery formula (base + level × modifier)
- [ ] Add current_environment_position to save file
- [ ] Use existing locations_discovered array for dungeon IDs
- [ ] Check for dungeons when player stops during traversal
- [ ] Show "Enter [Dungeon Name]" option if at discovered dungeon position

### Phase 6: Dungeon Access & Requirements

- [ ] Implement access requirement checking before dungeon entry
- [ ] Check player inventory for required items
- [ ] Consume items if marked as consumed in requirements
- [ ] Show requirement messages if player lacks items
- [ ] Allow entry only if all requirements met

### Phase 7: Dungeon Linear Progression

- [ ] Create dungeon entry UI
- [ ] Implement step-by-step linear progression
- [ ] Add static mob spawns (marked encounters always appear)
- [ ] Add treasure/reward system for dungeon steps
- [ ] Allow re-entry at discovered positions (respawn static mobs)

### Phase 8: Dungeon Completion Effects

- [ ] Implement "none" completion type (just return to position)
- [ ] Implement "traversal_progress" completion type
- [ ] Advance player position by amount_minutes when traversal_progress dungeon completed
- [ ] Update progress bar after dungeon completion with progression
- [ ] Test shortcuts properly advance player toward destination

### Phase 9: Integration & Polish

- [ ] Test all systems together
- [ ] Balance travel times for each environment
- [ ] Balance encounter rates (day vs night)
- [ ] Balance dungeon discovery rates by level
- [ ] Create 2-3 sample dungeons per environment type
- [ ] Add tutorial/help text for new mechanics
- [ ] Test position tracking and dungeon revisiting
- [ ] Performance testing with traversal calculations

---

**Notes**: This is a working draft. Mechanics subject to change based on playtesting and technical constraints.
