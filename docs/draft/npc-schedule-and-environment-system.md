# NPC Schedule & Environment System

**Status**: Draft/Workshop
**Created**: 2025-12-14
**Related**: 144x-time-progression.md, locations-plan.md

## Overview

This document outlines new systems for NPC schedules, time management during environment traversal, environment encounters, and dungeon discovery mechanics.

---

## 1. NPC Schedule System

### Concept
NPCs have schedules that determine their location throughout the day. Each NPC can be at different locations during different times.

### Schedule Structure

Each NPC has a schedule defining their location by time of day:

```json
{
  "npc_id": "merchant_bob",
  "schedule": [
    {
      "time_range": "06:00-12:00",
      "location": "general_store",
      "activity": "working"
    },
    {
      "time_range": "12:00-13:00",
      "location": "tavern",
      "activity": "lunch"
    },
    {
      "time_range": "13:00-18:00",
      "location": "general_store",
      "activity": "working"
    },
    {
      "time_range": "18:00-22:00",
      "location": "home",
      "activity": "resting"
    },
    {
      "time_range": "22:00-06:00",
      "location": "home",
      "activity": "sleeping"
    }
  ]
}
```

### Implementation Notes
- Time is checked against current game time (from 144x-time-progression.md)
- NPCs appear/disappear from locations based on schedule
- If player tries to interact with NPC at wrong time, show message: "Bob is not here right now"
- Optional: Add hints about NPC location ("Try the tavern around noon")

### Data Storage
- Add `schedule` field to NPC JSON files
- Backend resolves current location based on game time when loading location data

---

## 2. Time Movement & Control System

### Time States

**In Location** (Towns, Cities, Points of Interest):
- Time moves at 144x speed (as per 144x-time-progression.md)
- Always accessible pause/play buttons
- Player can pause time to plan, check inventory, etc.
- Time resumes when player clicks play

**Entering Environment**:
- Time PAUSES automatically
- Player sees environment description and options
- No time passes until player makes movement decision

**Traversing Environment** (Moving between locations):
- Player chooses cardinal direction (N, S, E, W, NE, NW, SE, SW)
- Time RESUMES and advances during travel
- Time advancement based on distance/terrain difficulty
- Encounter checks happen during travel (see section 3)

**In Combat/Encounters**:
- Time pauses during combat resolution
- Small time increment after combat ends

### UI Requirements
- Persistent pause/play controls
- Current time always visible
- Time flow indicator (paused/moving)
- Clear feedback when time state changes

### Implementation Notes
- Build on existing `handleAdvanceTimeAction` from 144x system
- Add time state to game state: `time_paused: bool`
- Movement actions calculate travel time and advance time accordingly
- Pause button sets `time_paused = true`, play sets to `false`

---

## 3. Environment Encounter System

### Environment Properties

Each environment has encounter rates and difficulty settings:

```json
{
  "environment_id": "dark_forest",
  "name": "Dark Forest",
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
  "available_monsters": [
    "wolf",
    "goblin",
    "spider"
  ],
  "terrain_type": "forest",
  "travel_speed_modifier": 0.8
}
```

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

## 4. Dungeon Discovery System

### Concept
Dungeons are special locations discovered while traversing environments. Once discovered, they can be revisited. They are linear with explicit encounters and rewards.

### Dungeon Properties

```json
{
  "dungeon_id": "abandoned_mine",
  "name": "Abandoned Mine",
  "parent_environment": "mountain_pass",
  "discovery": {
    "rarity": "uncommon",
    "chance_per_traversal": 0.15,
    "requires_conditions": []
  },
  "layout": {
    "type": "linear",
    "length": 5,
    "encounters": [
      {
        "step": 1,
        "type": "monster",
        "enemy": "kobold",
        "count": 2
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
        }
      }
    ]
  }
}
```

### Discovery Rarity Tiers
- **Common**: 25-40% chance per traversal
- **Uncommon**: 10-20% chance per traversal
- **Rare**: 3-8% chance per traversal
- **Very Rare**: 1-2% chance per traversal
- **Legendary**: 0.1-0.5% chance per traversal

### Discovery Flow

1. **During Environment Traversal**:
   - Roll for dungeon discovery
   - If discovered: Pause time, show discovery message

2. **Player Choice**:
   - "Enter Now" - Immediately enter dungeon
   - "Mark and Continue" - Save location, continue travel
   - "Ignore" - Continue travel (can still revisit if already discovered)

3. **Dungeon Entry**:
   - Time pauses
   - Player progresses step-by-step through linear layout
   - Each step presents encounter or event
   - No backtracking in linear dungeons

4. **Dungeon Completion**:
   - Mark as completed (can still revisit)
   - Return to environment or parent location
   - Time resumes

### Dungeon State Tracking

Track in save file:
```json
{
  "discovered_dungeons": [
    {
      "dungeon_id": "abandoned_mine",
      "discovered_at": "game_timestamp",
      "completed": true,
      "completion_count": 2
    }
  ]
}
```

### Revisiting Dungeons
- Once discovered, dungeon appears as option when entering parent environment
- "Enter [Dungeon Name]" option added to environment menu
- Can be run multiple times (respawns encounters/loot based on rules)
- Optional: Cooldown timer between runs

### Implementation Notes
- Add dungeon discovery roll to environment traversal
- Store discovered dungeons in save file
- Each environment can have multiple dungeons
- Dungeons with `requires_conditions` only available after prerequisites met
- Linear layout means simple progression: step 1 → step 2 → ... → end

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
- Add `discovered_dungeons` array
- Add `time_paused` state
- NPC schedule data in NPC files

**Location System**:
- Environments reference locations as start/end points
- Dungeons reference parent environment
- NPC schedules reference location IDs

---

## 6. Open Questions / Workshop Items

### Time Mechanics
- Should time advance during dialogue/NPC interactions?
- How much time should different terrain types take to traverse?
- Should there be weather that affects travel time/encounters?

### NPC Schedules
- Should NPCs have different schedules on different days?
- Should special events override regular schedules?
- Do NPCs travel between locations (visible on road) or teleport?

### Encounters
- Should encounter rate decrease if player is much higher level?
- Should player be able to flee from encounters?
- Should repeated encounters in same environment give reduced XP?

### Dungeons
- Should dungeons have consumable loot that doesn't respawn?
- How often can dungeons be re-run?
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

### Phase 1: NPC Schedules
- [ ] Add schedule data structure to NPC JSON schema
- [ ] Implement schedule parser
- [ ] Add current location resolver based on time
- [ ] Update location loading to check NPC schedules
- [ ] Add UI feedback when NPC is elsewhere

### Phase 2: Time Controls
- [ ] Add pause/play buttons to UI
- [ ] Implement time state management
- [ ] Update time advancement to respect pause state
- [ ] Add automatic pause on environment entry
- [ ] Add time flow indicator to UI

### Phase 3: Environment Encounters
- [ ] Add encounter data to environment JSON files
- [ ] Implement encounter probability calculations
- [ ] Add day/night detection
- [ ] Implement difficulty multipliers
- [ ] Connect encounters to combat system

### Phase 4: Dungeon Discovery
- [ ] Create dungeon JSON schema
- [ ] Implement discovery roll system
- [ ] Add dungeon state to save files
- [ ] Create dungeon entry UI
- [ ] Implement linear progression system
- [ ] Add explicit mob spawns
- [ ] Add reward/loot system for dungeons

### Phase 5: Integration & Polish
- [ ] Test all systems together
- [ ] Balance encounter rates
- [ ] Balance dungeon discovery rates
- [ ] Create sample dungeons for each environment
- [ ] Add tutorial/help text for new mechanics
- [ ] Performance testing with many NPCs/schedules

---

**Notes**: This is a working draft. Mechanics subject to change based on playtesting and technical constraints.
