# Vault System (House of Keeping)

**Status**: ❌ Not implemented - Planning phase
**Priority**: Medium - QoL feature for inventory management
**Note**: Original concept was "House of Keeping" - simplified to "Vault" for starting cities

## Overview

The **Vault** (originally "House of Keeping") is a universal storage system where players can safely store items beyond inventory limits.

**Simplified Scope**:
- Focus on general vault service for now
- Skip race-specific themes and multi-location mechanics initially
- Just need basic storage in starting cities

**Original Concept** (below):
The House of Keeping was designed as culturally-appropriate "banks" where players can safely store items across different locations. Each racial starting location has its own unique House of Keeping with thematic flavor and naming conventions.

**Current Plan**: See `locations-plan.md` for simplified vault requirements.

## Core Mechanics

### Access Rules
- **Free Access**: Players get automatic free storage access at their race's starting location
- **Paid Access**: Players must pay a one-time "Keeper's Fee" to register at other locations
- **Permanent Rights**: Once registered, players can freely store/withdraw items at that location forever

### Storage Capacity
- **40 slots** per location (configurable via `storage_slots` in NPC config)
- Items stored separately from inventory (won't affect encumbrance while stored)
- Each location's storage is independent

### Registration Fees by Location

| Location | House Name | Fee | Free For |
|----------|-----------|-----|----------|
| Kingdom (Human) | Vault of Crowns | 50g | Human, Half-Elf, Half-Orc, Tiefling |
| City-South (Elf) | Glade of Safekeeping | 60g | Elf |
| Town-Northeast (Dwarf) | Stonevault | 55g | Dwarf |
| Village-West (Halfling/Gnome) | Burrowlock | 45g | Halfling, Gnome |
| Village-Southwest (Orc) | Warhoard | 65g | Orc |
| City-East (Dragonborn) | Ember Vault | 70g | Dragonborn |

## Location-Specific Implementations

### 1. Vault of Crowns (Kingdom - Human)
- **Keeper**: Marcus Goldvault, Royal Custodian
- **Theme**: Official royal institution, bureaucratic and organized
- **Fee Name**: Royal Seal Fee (50g)
- **Lore**: Funded by the crown to encourage trade and travel

### 2. Glade of Safekeeping (City-South - Elf)
- **Keeper**: Silvanthir Moonbough, Warden of Roots
- **Theme**: Nature magic, living trees with hollow storage roots
- **Fee Name**: Offering of Binding (60g)
- **Lore**: Enchanted trees that "sing" your name and hollow a space attuned to your spirit

### 3. Stonevault (Town-Northeast - Dwarf)
- **Keeper**: Thorin Ironledger, Vaultwright
- **Theme**: Ancient clan heritage, carved deep into mountains
- **Fee Name**: Vaultstone Engraving Fee (55g)
- **Lore**: Five generations of dwarven craftsmanship, permanent stonework

### 4. Burrowlock (Village-West - Halfling/Gnome)
- **Keeper**: Rosie Lockwood, Keywarden
- **Theme**: Cozy family burrows beneath homes, cedar-lined
- **Fee Name**: Burrow Key Fee (45g)
- **Lore**: Each family maintains shared storage burrows with unique brass locks

### 5. Warhoard (Village-Southwest - Orc)
- **Keeper**: Grommash Ironhide, Hoardkeeper
- **Theme**: War spoils, communal clan caches, fierce warrior guardian
- **Fee Name**: Tribute of Iron (65g)
- **Lore**: Outsiders must prove worth through tribute; betrayal means death

### 6. Ember Vault (City-East - Dragonborn)
- **Keeper**: Zarithax the Golden, Scalekeeper
- **Theme**: Dragonfire-sealed chests, draconic authority
- **Fee Name**: Flamebinding Rite (70g)
- **Lore**: Vaults bound by dragonfire ceremony, protected by ancient draconic magic

## NPC Data Structure

NPCs are stored in `docs/data/content/npcs/{location}/{npc-id}.json`

### Schema
```json
{
  "id": "npc-id",
  "name": "Full Name",
  "title": "Official Title",
  "race": "Race",
  "location": "location-id",
  "building": "building_id",
  "description": "Visual description and personality",
  "greeting": {
    "first_time": "First meeting dialogue",
    "returning": "Returning player dialogue",
    "native_race": "Greeting for players of keeper's race"
  },
  "dialogue": {
    "node_id": {
      "text": "Dialogue text",
      "action": "action_name",
      "cost": 50,
      "requirements": {
        "gold": 50,
        "not_registered": true
      },
      "success": "Success message",
      "failure": "Failure message",
      "options": ["next_node", "another_node"]
    }
  },
  "storage_config": {
    "free_for_races": ["Race1", "Race2"],
    "registration_fee": 50,
    "storage_slots": 40,
    "slot_type": "standard"
  }
}
```

### Dialogue Actions
- `register_storage` - Pay fee and unlock storage access
- `open_storage` - Open storage interface
- `end_dialogue` - Close dialogue

### Dialogue Requirements
- `gold: N` - Requires N gold
- `registered: true/false` - Whether player is registered at this location
- `not_registered: true` - Shorthand for registered: false
- `not_native: true` - Player is not of a native race

## Location Integration

Buildings reference NPCs in location JSON files:

```json
{
  "id": "vault_of_crowns",
  "name": "The Vault of Crowns",
  "type": "house_of_keeping",
  "description": "Building description",
  "accessible": true,
  "npc": "royal-custodian",
  "actions": ["talk_to_custodian", "access_storage", "register"]
}
```

NPCs are listed in district `npcs` array:
```json
"npcs": ["royal-custodian", "other-npc", "another-npc"]
```

## Backend Implementation TODO

### 1. NPC Data Loading
- [ ] Create `NPCHandler` in `src/api/`
- [ ] Add NPC table to DuckDB schema
- [ ] Load NPC JSON files into database
- [ ] Create API endpoint `/api/npcs/{location}/{npc-id}`
- [ ] Add NPC loading to `GameDataHandler`

### 2. Storage System
- [ ] Add `storage_registrations` table to track player registrations
  - Columns: `npub`, `location_id`, `registered_at`
- [ ] Add `stored_items` table for player storage
  - Columns: `npub`, `location_id`, `slot_number`, `item_id`, `quantity`
- [ ] Create storage API endpoints:
  - `POST /api/storage/register` - Register at location
  - `GET /api/storage/{location}` - Get stored items
  - `POST /api/storage/{location}/deposit` - Store item
  - `POST /api/storage/{location}/withdraw` - Retrieve item

### 3. Dialogue System
- [ ] Create dialogue state machine
- [ ] Handle dialogue requirements checking
- [ ] Process dialogue actions (register_storage, open_storage)
- [ ] Track conversation state in save file

### 4. Save File Integration
Add to save file:
```json
{
  "storage_registrations": ["kingdom", "city-south"],
  "stored_items": {
    "kingdom": {
      "0": {"item": "longsword", "quantity": 1},
      "1": {"item": "health-potion", "quantity": 5}
    }
  }
}
```

## Future Enhancements
- Reputation system affecting fees
- Shared storage for guilds
- Special storage for valuable/magical items
- Storage expansion through quests
- Cross-location item transfer (expensive)
