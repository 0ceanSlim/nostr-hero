# Nostr Hero: Text-Based Turn-Based RPG

## Game Overview

Nostr Hero is a text-based, turn-based RPG built with HTMX and server-side rendering. Players navigate through a medieval fantasy world, exploring cities and wilderness environments, engaging in combat, and developing their character through D&D 5e-inspired mechanics.

## Core Game Loop

### 1. Location-Based Navigation

The game world consists of **Cities** and **Environments** connected by a grid-based movement system.

#### City Navigation

- Each city has a **Center** and directional **Districts** (North, South, East, West)
- Players can move between districts using cardinal directions
- Each district contains **Buildings** with specific functions (shops, taverns, guilds, etc.)
- Districts that border other locations allow **exit to environments**

#### Environment Travel

- **Environments** are wilderness areas between cities
- **Daily Actions**: Each day choose to travel forward, travel back, rest, or forage
- **Travel Progress**: Environments have minimum travel times (e.g., 3 days for forest)
- **Encounters**: Daily chance for monsters, events, or discoveries
- **Resource Management**: Each day requires rations, builds fatigue
- **Two-Way Travel**: Can advance toward destination or retreat to origin at any time

### 2. World Structure

```
                    Town----------Arctic---------Town
                      |                           |
                      |                           |
                    Swamp                      Mountain
                      |                           |
                      |                           |
Village---Hill-----KINGDOM---------Urban---------City----Desert---Village
                      |
                      |
                    Forest
                      |
                      |
                    City---------Swamp---------Village
                      |
                      |
                   Coastal
                      |
                      |
                   Village
```

#### Location Types:

- **Kingdom**: Central hub, 4 connections (starter city)
- **City**: Major settlement, 3 connections
- **Town**: Medium settlement, 2 connections
- **Village**: Small settlement, 1 connection

#### Environment Types:

- Arctic, Coastal, Desert, Forest, Grassland, Hill, Mountain, Swamp, Underwater, Urban

## Game Mechanics

### Character Progression

- **Level Cap**: 20 with exponential XP scaling making higher levels extremely time-consuming
- **Health**: Hit die max + CON modifier at level 1, +half hit die average per level
- **Attributes**: STR, DEX, CON, INT, WIS, CHA with D&D-style modifiers
- **Skills**: Individual skill progression separate from character level

### Inventory & Equipment System

- **Base Capacity**: 4 inventory slots + equipment slots
- **Equipment Slots**: Armor, Right Hand, Left Hand, Necklace, Ring
- **Carry Weight**: Base (STR × 5), backpack doubles total capacity
- **Slot Expansion**: Certain items increase available slots when equipped
  - Component Pouch: +4 slots (total 8)
  - Backpack: +20 slots (total 24)

### Combat System

- **Range-Based Tactical Combat**: Distance 0-6 with weapon effectiveness zones
- **Initiative**: Determined by fatigue-based stealth checks
  - **Player Advantage**: Fresh players can spot enemies first (distance 5, can skip)
  - **Normal Encounter**: Moderate fatigue leads to simultaneous detection (distance 3)
  - **Monster Ambush**: Exhausted players get surprised (distance 1, monster first)
- **Turn Structure**: Optional movement + main action
- **Actions**: Attack, Defend (+2 AC), Advance, Retreat, Use Item, Flee
- **Fatigue Effects**: Reduces stealth, combat rolls, and starting distance
- **Weapon Properties**: Range values (0-6), finesse, heavy, ammunition, etc.

#### Encounter Types & Starting Positions

- **Player Advantage** (sneak up): Start at distance 5, player goes first, can skip fight
- **Normal Encounter**: Start at distance 3, player goes first
- **Monster Ambush**: Start at distance 1, monster goes first
- **Fatigue Effect**: Exhausted players start 1-2 distances closer

#### Combat Actions

- **Movement**: Step forward (-1 distance) or step back (+1 distance)
- **Melee Attack**: Effective at range 0-1, uses STR modifier
- **Ranged Attack**: Uses weapon range (1-6), uses DEX modifier, requires ammunition
- **Cast Spell**: Use prepared spells with varying ranges and effects
- **Defend**: +2 AC for one turn
- **Use Item**: Consume potions, activate magic items
- **Flee**: Escape chance based on speed, adds 1 fatigue

#### Spell Casting in Combat

- **Spell Ranges**: Each spell has specific range limitations
  - **Touch Spells** (range 0): Cure Wounds, Shocking Grasp
  - **Close Range** (1-2): Fire Bolt, Ray of Frost
  - **Medium Range** (3-4): Magic Missile (automatic hit)
  - **Long Range** (5-6): Some high-level spells
- **Mana Cost**: Each spell consumes mana points
  - **Cantrips**: 0 mana (unlimited use)
  - **Level 1 Spells**: 1 mana each
  - **Higher Levels**: Increasing mana costs
- **Material Components**: Some spells require specific items
  - **Fireball**: Requires "bat-guano-and-sulfur" (consumed on use)
  - **Cure Wounds**: Requires "blessed-water" (consumed on use)
  - **Component Pouch**: Provides materials for spells without specific requirements
- **Spell Effects**: Different spell types affect combat differently
  - **Damage Spells**: Magic Missile (automatic hit), Fire Bolt (ranged attack)
  - **Defensive Spells**: Shield (+5 AC for one turn)
  - **Healing Spells**: Cure Wounds (restore HP)
  - **Utility Spells**: Various battlefield effects
- **Spell Preparation**: Only prepared spells can be cast
  - **Wizards**: Must prepare spells from spellbook each day
  - **Clerics**: Can prepare any spell from their domain list
  - **Sorcerers**: Know limited spells, always available

#### Range Effectiveness

- **Distance 0**: Grappling, both creatures engaged
- **Distance 1**: Optimal for melee weapons (swords, axes)
- **Distance 2-3**: Close range for bows, crossbows, thrown weapons
- **Distance 4-6**: Long range for longbows, heavy crossbows
- **Beyond Range**: Cannot attack or severe penalties

### Currency & Resources

- **Gold Only**: All currency is in Gold Pieces (GP), no copper/silver/platinum
- **Gold Values**: Based on D&D copper piece rates (no decimals)
  - Rations: 5 GP, Sword: 150 GP, Inn Stay: 20 GP, etc.
- **Gold Weight**: 1 GP = 0.0006 lbs
- **Rations**: Required for travel and sleeping
- **Sleep**: Requires bedroll, restores ½ health, consumes 1 ration
- **City Entry Fees**: Required for most cities (except starting city)

### Fatigue System (0-10 Scale)

- **Fresh (0-2)**: No penalties
- **Tired (3-4)**: -1 to combat rolls
- **Weary (5-6)**: -2 to combat rolls
- **Exhausted (7-8)**: -3 to combat rolls, start combat 1 distance closer
- **Critically Fatigued (9-10)**: -5 to combat rolls, start combat 2 distances closer, 1 HP damage per day

#### Fatigue Sources

- **Daily Travel**: +1 fatigue per day
- **Hard/Extreme Environments**: +1 additional fatigue per day
- **Terrain Hazards**: +1 fatigue when encountered
- **No Rations**: +2 fatigue per day
- **Combat**: +1 fatigue per encounter

#### Fatigue Recovery

- **Rest with Bedroll**: -2 fatigue, restores ½ max HP, requires bedroll + 1 ration, takes 1 day
- **Rest with Shelter**: -3 fatigue, restores ½ max HP, requires bedroll + 1 ration + shelter location, takes 1 day
- **Poor Rest**: -1 fatigue, restores ¼ max HP, requires 1 ration (no bedroll), takes 1 day
- **Consume Ration**: -1 fatigue, can be done anytime (once per day), does not restore HP
- **No Rest**: +2 fatigue per day without rations, no HP recovery

### Travel System

- **Environment-Based**: Each environment has `travel_time` (minimum days) and `travel_difficulty`
- **Daily Actions**: Travel forward/back (1 fatigue + 1 ration), Rest, or Forage
- **Terrain Hazards**: Daily chance for delays based on environment difficulty
- **Random Encounters**: Monsters and events based on environment type
- **Two-Way Travel**: Can advance toward destination or retreat to origin

## Technical Architecture

### HTMX + Server-Side Rendering

- **All game state** managed server-side in Go
- **DOM updates** via HTMX swaps for real-time interaction
- **Turn-based actions** as form submissions
- **Session management** with character data persistence
- **Combat calculations** processed server-side with range/fatigue mechanics

### Core Game Systems Integration

- **Fatigue System**: Affects stealth, combat rolls, and encounter outcomes
- **Range-Based Combat**: Distance 0-6 with weapon effectiveness zones
- **Travel Mechanics**: Daily actions with resource consumption and encounter rolls
- **Inventory Management**: Equipment slots + expandable storage system
- **Gem System**: Procedural drops with size multipliers and CR scaling

### Data Structure

#### Environment JSON Format

```json
{
  "id": "forest-kingdom",
  "name": "Darkwood Forest",
  "type": "environment",
  "environment_type": "forest",
  "connects": ["kingdom-south", "south-city-north"],
  "description": "Ancient trees tower overhead...",
  "travel_time": 3,
  "travel_difficulty": "moderate",
  "required_rations": 1,
  "encounter_table": {
    "daily_encounter_chance": 0.35,
    "encounter_types": {
      "monster": 0.7,
      "event": 0.2,
      "discovery": 0.1
    },
    "monster_tiers": {
      "common": {"weight": 60, "monsters": ["goblin", "wolf"]},
      "uncommon": {"weight": 30, "monsters": ["owlbear", "treant"]},
      "rare": {"weight": 10, "monsters": ["green-dragon-wyrmling"]}
    }
  },
  "rest_spots": [{"name": "Hollow Tree", "shelter": true}],
  "weather_effects": {...},
  "resources": {...}
}
```

#### Monster JSON Format

```json
{
  "id": 17,
  "name": "Goblin",
  "challenge_rating": 0.25,
  "hit_points": 7,
  "armor_class": 15,
  "encounter_data": {
    "encounter_weight": 15,
    "rarity_tier": "common",
    "min_group_size": 1,
    "max_group_size": 4,
    "ambush_chance": 0.2
  },
  "loot_table": {
    "gold_range": [1, 6],
    "items": [{ "name": "Dagger", "chance": 0.3 }],
    "gem_table": { "enabled": true, "cr_modifier": 0.25 }
  },
  "enviornment": ["forest", "grassland", "mountain"]
}
```

#### Character Save Format (Nostr Event)

```json
{
  "character": {
    "race": "Human",
    "class": "Cleric",
    "stats": {"strength": 8, "dexterity": 12, ...},
    "hp": {"current": 10, "max": 10},
    "fatigue": 3,
    "location": "kingdom-center",
    "inventory": [["Healing Potion", 3], ["Rations", 5]],
    "equipment": {
      "armor": "Chain Mail",
      "weapon_right": "Mace",
      "weapon_left": "Shield",
      "ring": null,
      "necklace": "Holy Symbol"
    }
  }
}
```

## UI/UX Design

### City Interface

```
=== Kingdom Center ===
You stand in the bustling heart of the kingdom. Noble carriages
roll past merchants hawking their wares...

Available Actions:
→ North District
→ South District
→ East District
→ West District
→ Enter Grand Market
→ Visit Royal Palace
→ Check Character Status
```

### Environment Interface

```
=== Darkwood Forest ===
Ancient oaks stretch toward a canopy that blocks most sunlight.
You hear rustling in the underbrush...

Travel Progress: ████░░ (Day 2 of 3)

Available Actions:
→ Continue Forward
→ Search for Resources
→ Rest (Requires Bedroll & Rations)
→ Return to Kingdom
```

### Combat Interface

```
=== Combat ===
A goblin emerges from behind a tree, weapon drawn!
Distance: 3 (Normal encounter)

Goblin: ████████░░ (8/10 HP)
You:    ██████████ (15/15 HP) [Tired: -1 to rolls]
Mana:   ████░░░░░░ (4/10 MP)

Your Turn:
→ Advance (distance 2)
→ Retreat (distance 4)
→ Attack with Sword (range 0-1, need to close distance)
→ Attack with Shortbow (range 2-4, effective at current distance)
→ Cast Fire Bolt (1-2 range, need to close distance, 0 mana)
→ Cast Magic Missile (3-4 range, effective, 1 mana, auto-hit)
→ Cast Shield (defensive, +5 AC this turn, 1 mana)
→ Defend (+2 AC this turn)
→ Use Health Potion
→ Attempt to Flee (cannot flee at this distance)

Spell Components:
- Component Pouch: Available for general spells
- Bat Guano & Sulfur: 2 remaining (for Fireball)
```

## Content Progression

### Character Generation

- **Deterministic Creation**: Characters generated from hashed Nostr key
- **Race-Specific Starting Locations**: Each race begins in their appropriate homeland
- **Background-Based Equipment**: Starting gear determined by character background
- **Personal Opening Story**: Unique introduction based on background and class combination

### Starter Experience

1. **Personal Introduction**: Background-specific narrative about caretaker's death and departure
2. **Equipment Selection**: Choose class-appropriate gear left by caretaker
3. **Starting Location**: Begin in racial homeland (Kingdom for Humans, etc.)
4. **Tutorial Integration**: Learn movement, building interaction, basic UI through natural gameplay
5. **First Environment**: Guided exploration based on chosen path from starting city

### World Expansion

- **Unlock new cities** through exploration
- **Discover hidden locations** via special events
- **Environmental storytelling** through location descriptions
- **Dynamic events** based on player choices and progression

## Special Features

### Nostr Integration

- **Character saves** on Nostr relays
- **Legacy character bonuses** for returning players
- **Tavern VIP section** for happytavern NIP-05 holders
- **Social sharing** of achievements and discoveries

### Audio Design

- **Environment-specific** background music
- **Combat music** for battle sequences
- **Ambient sounds** for immersive experience

### Building Interactions

- **Simple NPC Housing**: Buildings primarily contain NPCs with basic interaction options
- **Shops**: Buy/sell interface with limited inventory
- **Taverns**: Rest, rumors, and social gathering points
- **Services**: Healing, repairs, information, and quest givers
- **Future Expansion**: More complex building interactions planned

### NPCs

- **Basic Dialogue**: Simple conversation trees with 2-4 response options
- **Services**: Trading, healing, information, quest distribution
- **Faction Relationships**: NPCs may react differently based on player reputation
- **Future Development**: More complex NPC interactions and relationships planned

## Social Features & Multiplayer

### Current Implementation

- **Single Player**: No social features in initial release
- **Planned Features**:
  - Party System: Travel environments together with other players
  - PvP Combat: Player vs player encounters in certain zones
  - Guild System: Player organizations with shared goals

### Magic & Spell System

#### Current Implementation
- **17 Implemented Spells**: From cantrips to high-level spells
  - Cantrips: Fire Bolt, Ray of Frost, Shocking Grasp, etc.
  - Level 1: Magic Missile, Shield, Cure Wounds, etc.
  - Higher Levels: Fireball, Thunderwave, etc.
- **Material Components**: Custom system using specific items
  - Replaced traditional V/S/M with material-only requirements
  - "bat-guano-and-sulfur" for explosive spells
  - "blessed-water" for divine magic
  - "cursed-bone-dust" for necromancy
  - Component pouches for spells without specific materials
- **Class Integration**: Spell lists integrated with character classes
  - Wizards: Spellbook system with learnable spells
  - Clerics: Domain-based spell access
  - Sorcerers: Known spells system
  - Each class has appropriate starting spells
- **Mana System**: Custom mana costs instead of spell slots
  - Cantrips: 0 mana
  - Level 1: 1 mana
  - Scaling costs for higher levels
- **Spell Properties**: Damage, range, school, casting time all defined

#### Spell Data Structure
```json
{
  "name": "Magic Missile",
  "level": 1,
  "school": "evocation",
  "mana_cost": 1,
  "damage": "3d4 + 3",
  "spell_attack": "automatic",
  "material_component": null,
  "classes": ["sorcerer", "wizard"]
}
```

### Save System & Architecture

#### Current Implementation
- **DOM State Management**: All game state stored in hidden DOM elements during play
- **Manual Save System**: Players click "Save Game" to persist state to Nostr relays
- **Nostr Integration**: Character data stored as addressable Nostr events
- **JSON Format**: Complete save state with character, inventory, spells, location
- **Cross-Device**: Access character from any device with Nostr keys
- **Working Character Creation**: Full HTMX-based character creation with gear selection

#### Current Tech Stack
- **Backend**: Go server with minimal APIs for data serving
- **Database**: DuckDB for static game data (items, spells, monsters, locations)
- **Frontend**: HTMX + Hyperscript + JavaScript for DOM state management
- **Save Storage**: Nostr relays for persistent character saves
- **File Structure**: Organized data folders (character/, systems/, equipment/, content/)

#### Implementation Status
- ✅ **Character Discovery & Generation**: Working deterministic character creation
- ✅ **Spell System**: 17 spells with material components integrated
- ✅ **Item System**: 184+ items with full properties and tags
- ✅ **Character Creation UI**: Dynamic HTMX interface with gear selection
- 🚧 **Database Migration**: Moving from JSON files to DuckDB
- 📋 **Game Interface**: Planning DOM-based gameplay mechanics
- ⏸️ **Combat System**: Designed but awaiting implementation
- ⏸️ **Location System**: Designed but awaiting implementation

This design provides a foundation for a rich, text-based RPG experience using modern web technologies. The DOM-based state management with manual saves to Nostr relays creates a unique hybrid of client-side responsiveness and decentralized persistence, perfect for the HTMX/Hyperscript approach.
