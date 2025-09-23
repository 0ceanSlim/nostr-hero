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
- Players encounter **random events**, **monsters**, and **exploration opportunities**
- Travel requires **rations** and has **travel time/fatigue mechanics**
- After 4-10 activities, players can choose to continue or return to a city

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
- **Level Cap**: 20 (demo cap at 299 XP to prevent leveling)
- **Skill Points**: Gained at levels 3, 5, 7, 9, 11, 13, 15, 17, 19, 20
- **Health**: Hit die max + CON modifier at level 1, +half hit die per level
- **Carry Weight**: (STR × 5) base, ×10 with backpack
- **Storage**: 4 slots without backpack, 20 slots with backpack

### Combat System
- **Turn-based** D&D 5e-inspired combat
- **Damage Modifiers**: (Stat - 10) ÷ 2
- **Weapon Properties**: Finesse, Range, Heavy, etc.
- **Monster Encounters**: Based on environment and CR scaling

### Resources & Survival
- **Rations**: Required for travel and sleeping
- **Sleep**: Requires bedroll, restores ½ health, consumes 1 ration
- **City Entry Fees**: Required for most cities (except starting city)
- **Gold Weight**: 1 GP = 0.0006 lbs

### Travel System
- **Walking**: Move between adjacent locations through environments
- **Cart Travel**: Direct city-to-city transport (costs gold + rations)
- **Random Encounters**: Monsters based on environment type
- **Activity Tracking**: 4-10 activities trigger rest/continue choice

## Technical Architecture

### HTMX + Server-Side Rendering
- **All game state** managed server-side
- **DOM updates** via HTMX swaps
- **Turn-based actions** as form submissions
- **Real-time updates** through server responses

### Data Structure

#### Location JSON Format
```json
{
  "id": "kingdom-center",
  "name": "Kingdom Center",
  "type": "city_district",
  "parent_city": "kingdom",
  "description": "The bustling heart of the kingdom...",
  "connections": {
    "north": "kingdom-north",
    "south": "kingdom-south",
    "east": "kingdom-east",
    "west": "kingdom-west"
  },
  "buildings": [
    {
      "name": "Royal Palace",
      "type": "landmark",
      "description": "Seat of royal power",
      "accessible": true
    },
    {
      "name": "Grand Market",
      "type": "shop",
      "description": "Central marketplace",
      "accessible": true,
      "shop_type": "general"
    }
  ],
  "npcs": ["royal_guard", "merchant_leader"],
  "special_features": []
}
```

#### Environment JSON Format
```json
{
  "id": "forest-kingdom-to-south-city",
  "name": "Darkwood Forest",
  "type": "environment",
  "environment_type": "forest",
  "connects": ["kingdom-south", "south-city-north"],
  "description": "Ancient trees tower overhead...",
  "travel_time": 3,
  "encounter_tables": {
    "common": ["goblin", "wolf", "bear"],
    "uncommon": ["owlbear", "treant"],
    "rare": ["green_dragon_wyrmling"]
  },
  "resources": ["herbs", "lumber", "game"],
  "random_events": ["lost_traveler", "ancient_shrine", "bandit_camp"]
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

Goblin: ████████░░ (8/10 HP)
You:    ██████████ (15/15 HP)

Your Turn:
→ Attack with Sword (1d8+2)
→ Cast Spell
→ Use Item
→ Attempt to Flee
```

## Content Progression

### Starter Experience
1. **Character Creation**: Background selection influences starting equipment/gold
2. **Kingdom Tutorial**: Learn movement, building interaction, basic UI
3. **First Environment**: Guided forest exploration with simple encounters
4. **First Combat**: Tutorial battle with scaling difficulty
5. **Return to City**: Learn about rest, shopping, city services

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

## Questions for Clarification

1. **Building Interaction**: How detailed should individual building interactions be? Should each shop have its own interface or simplified "trade" actions?

2. **Combat Complexity**: Should we implement full D&D 5e combat rules (initiative, actions/bonus actions, spell slots) or simplified turn-based combat?

3. **Inventory Management**: How detailed should item management be? Visual grid or simple list?

4. **Social Features**: Should there be player-to-player interactions, or is this purely single-player?

5. **Save System**: Should character progression sync across sessions via Nostr, or local storage with periodic Nostr backups?

This design provides a foundation for a rich, text-based RPG experience while maintaining the simplicity needed for HTMX implementation. The modular location system allows for easy world expansion and the turn-based mechanics ensure smooth server-side state management.