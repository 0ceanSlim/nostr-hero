# ⚔️ Nostr Hero

A web-based D&D-style RPG that generates your unique character from your Nostr identity and lets you adventure through a persistent world stored on Nostr relays.

## Overview

Nostr Hero is a nostalgic RPG experience that derives a deterministic character from your Nostr public key. Your cryptographic identity becomes an adventurer with unique stats, equipment, and abilities. The game combines classic D&D 5e mechanics with Nostr's decentralized protocol to create a persistent, cross-client RPG experience.

## Current Development Status (Pre-Alpha)

The game is currently in **very early pre-alpha development**. Most systems are planned or in initial implementation.

### ✅ Implemented So Far

- **Deterministic Character Generation**:
  - Race, class, background, and alignment derived from Nostr pubkey
  - D&D 5e ability scores (STR, DEX, CON, INT, WIS, CHA)
  - Weighted distribution for realistic race/class combinations

- **Character Introduction System**:
  - Unique narrative introductions for each class/background combination
  - Contextual storytelling that reflects character origins
  - Atmospheric scene-setting with class-specific imagery

- **Starting Equipment System**:
  - Complex class-based equipment selection algorithm
  - Background-specific bonus equipment
  - Intelligent gear choices (e.g., spellcasters get component pouches, rangers get survival gear)
  - Deterministic selection ensures same character always gets same loadout

- **Starting Spell System**:
  - Class-specific spell lists for all spellcasting classes
  - Cantrip selection based on class mechanics
  - Level 1 spell loadout unique to each spellcaster
  - Spell slots properly allocated per class rules

- **Character Initialization**:
  - Starting HP calculated from class hit die + CON modifier
  - Starting mana derived from spellcasting ability + level
  - Starting gold based on class and background
  - All stats properly initialized and balanced

- **Game UI (Visual Only)**:
  - Dark Win95-themed retro interface with beveled edges
  - Tabbed panels: Equipment, Inventory, Spells, Quests, Stats, Music
  - Character stat display (HP, mana, fatigue, encumbrance bars)
  - Equipment slot visualization
  - Backpack grid (20 slots)
  - Spell slot interface
  - **Note**: UI displays save data but lacks interactive functionality

- **Content Database**:
  - 200+ items from D&D 5e SRD (weapons, armor, gear, tools)
  - Pixel art sprites (64x64) for items
  - Full D&D 5e spell database
  - Monster database
  - Location data
  - Item editor GUI tool for content creation

- **Save System (Basic)**:
  - Local JSON saves tied to npub
  - Save file generation from character creation
  - UI instantiation from save data

- **Authentication**:
  - Nostr login via NIP-07 browser extensions or Amber (Android)
  - Grain authentication client integration

### 🚧 Not Yet Implemented (Pre-Alpha → Alpha Goals)

The following systems are **designed/specified** but need initial implementation:

- **Inventory System Logic**:
  - Moving items between slots
  - Equipping/unequipping gear
  - Item stacking
  - Weight/encumbrance calculations
  - Dropping/destroying items

- **Spell System Logic**:
  - Casting spells
  - Mana consumption
  - Spell slot tracking
  - Spell preparation mechanics

- **Game UI Interactivity**:
  - Clickable items
  - Drag-and-drop
  - Context menus
  - Tooltips with item details

- **Location System**:
  - Scene rendering
  - Location transitions
  - Environment traversal

- **Combat System**:
  - Turn-based combat
  - Dice rolling mechanics
  - Enemy AI
  - Loot drops

- **NPC Systems**:
  - NPC interactions
  - Dialogue system
  - Shops (buy/sell)
  - Storage systems
  - Inn/rest mechanics

### 📋 Development Roadmap

#### Pre-Alpha → Alpha Goals
Focus: **Core Gameplay Mechanics**

- Implement inventory management logic
- Implement spell casting and mana system
- Build location/scene system with navigation
- Create environment traversal mechanics
- Implement combat encounters
- Add basic NPC interactions (shop, storage, inn)
- Complete save/load functionality
- Playtesting and balance adjustments

#### Alpha → Beta Goals
Focus: **Content & Nostr Integration**

- **Quest System**:
  - Handcrafted quest chains (all quests are manually designed)
  - Universal quests available to all players
  - Main story quests accessible regardless of character
  - Character-specific quests (race/class/background exclusive)
  - Exclusive quest rewards tailored to specific classes/roles
  - **Note**: No main story or critical content locked by character type

- **Full Nostr Integration**:
  - **Relay-based Saves**: Store save state as Nostr events (cross-client compatible)
  - **Save Validation**: Server validates saves against official list (modded vs unmodded tracking)
  - **Dungeon Master npub**: Official account for game announcements, player DMs, community engagement
  - **Nostr Badges**: Award achievements as NIP-58 badges
  - **Valid Saves List**: NIP-51 list event tracking legitimate save event IDs
  - **In-game Nostr Features**:
    - Write kind 1 notes using parchment and ink pen
    - Write long-form articles (kind 30023) using books
    - More creative integrations TBD

- **Cross-client Gameplay**:
  - Play on any Nostr Hero client with the same character
  - Community-built clients and mods
  - Official validation against canonical ruleset

## Tech Stack

### Backend (Go)
- **Web Server**: HTTP server with Go template rendering
- **Character Generation**: Deterministic algorithm based on Nostr pubkey
- **Authentication**: Grain client for Nostr auth (NIP-07, Amber)
- **API Endpoints**: REST API for game data, saves, and character info
- **Future**: DuckDB for analytics and event validation

### Frontend
- **Vanilla JavaScript**: Pure JS for game logic and state management
- **TailwindCSS**: Dark Win95-inspired retro theme
- **Go Templates**: Server-side rendering
- **Session Management**: Client-side Nostr session handling

### Data Format
- **JSON-based**: All game content (items, spells, monsters, locations)
- **Modular Structure**: Easy to extend with new content
- **Save Format**: Flat JSON structure (will migrate to Nostr events)

## Game Data

All game content is stored as JSON in `docs/data/`:

- **Items**: `docs/data/equipment/items/` - Weapons, armor, adventuring gear, tools
  - 200+ items from D&D 5e SRD
  - Pixel art sprites for each item
  - Stats, costs, weights, and descriptions

- **Spells**: `docs/data/content/spells/` - Full D&D 5e spell list
  - Spell effects and mechanics
  - Mana costs and casting requirements
  - Class-specific spell access

- **Monsters**: `docs/data/content/monsters/` - Creature database
  - Stats and abilities
  - Loot tables
  - AI behaviors (planned)

- **Locations**: `docs/data/content/locations/` - World map data
  - Location descriptions
  - Connected environments
  - NPC spawn points

- **Character Data**: `docs/data/character/` - Character generation tables
  - Race/class/background definitions
  - Starting equipment packs
  - Ability score distributions

## Save System

**Current (Pre-Alpha)**: Local saves stored in `data/saves/{npub}/save_{timestamp}.json`

```json
{
  "d": "character-name",
  "created_at": "2025-10-08T23:09:49Z",
  "race": "Human",
  "class": "Bard",
  "background": "Sage",
  "alignment": "Neutral Evil",
  "level": 1,
  "experience": 0,
  "hp": 7,
  "max_hp": 7,
  "mana": 3,
  "max_mana": 3,
  "fatigue": 0,
  "gold": 1000,
  "stats": {
    "strength": 13,
    "dexterity": 12,
    "constitution": 8,
    "intelligence": 9,
    "wisdom": 12,
    "charisma": 14
  },
  "location": "kingdom",
  "inventory": {
    "gear_slots": { /* equipment */ },
    "general_slots": [ /* 4 general slots */ ]
  },
  "known_spells": [ /* spell IDs */ ],
  "spell_slots": { /* prepared spells */ },
  "locations_discovered": [ "kingdom" ],
  "music_tracks_unlocked": [ "kingdom-theme" ]
}
```

**Future (Beta)**: Nostr event-based saves with cross-client compatibility

## Contributing

Contributions are welcome! Here's how you can help:

1. **Report Bugs**: Use the [bug report template](https://github.com/0ceanSlim/nostr-hero/issues/new?template=bug_report.md)
2. **Suggest Features**: Open an issue with your idea
3. **Add Content**: Create new items, spells, monsters, or locations (JSON format)
4. **Improve Code**: Submit PRs for bug fixes or enhancements
5. **Playtesting**: Try the game and provide feedback (when playable builds are available)

**Content Creators**: Check out `docs/development/tools/` for the item editor and other content creation tools.

## Getting Started

See [Development Documentation](docs/development/) for setup instructions and development guides.

## License

This project is Open Source and licensed under the MIT License. See the [LICENSE](license) file for details.

---

Open Source and made with 💦 by [OceanSlim](https://njump.me/npub1zmc6qyqdfnllhnzzxr5wpepfpnzcf8q6m3jdveflmgruqvd3qa9sjv7f60)
