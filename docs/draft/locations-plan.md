# Locations System Plan

Current plan for location-based services in starting cities.

## Starting City Requirements

Each starting city needs these core services:

### 1. General Store
**Purpose**: Buy/sell basic items and equipment

**Features**:
- Buy items from curated inventory
- Sell items from player inventory
- Pricing based on shop-pricing-system.md
- Stock based on location tier/size

**TODO**:
- Implement shop UI
- Create shop inventory lists per city
- Implement buy/sell transactions
- Add price calculations

### 2. Vault (House of Keeping)
**Purpose**: Long-term item storage

**Features**:
- Store items beyond inventory limits
- Safe storage (items don't decay/disappear)
- Per-character storage (not shared between saves)
- Access from any vault location

**TODO**:
- Implement vault UI
- Add storage capacity limits
- Create vault data structure in save files
- Implement deposit/withdraw mechanics

**See**: `house-of-keeping-system.md` for initial design (needs updates)

### 3. Inn/Tavern
**Purpose**: Rest, recover, and social hub

**Features**:
- Rest to recover HP/mana/hunger/fatigue
- Cost to rest based on room quality
- Social interactions (future: NPCs, quests)
- Performance opportunities (future: bard performances)

**TODO**:
- Implement rest system
- Add pricing tiers (poor/common/fine rooms)
- Create rest recovery calculations
- Add basic tavern UI

**See**: `tavern-performance-design.md` and `tavern-npcs-todo.md` for future features

## Implementation Priority

1. **General Store** - Highest priority, needed for basic gameplay
2. **Inn/Tavern** - Medium priority, needed for survival mechanics
3. **Vault** - Lower priority, nice-to-have for inventory management

## Future Services

These may be added to larger cities later:
- Blacksmith (equipment upgrades/repairs)
- Alchemist (potion crafting)
- Magic shop (spell scrolls, wands)
- Guild halls (class-specific quests)
- Training grounds (skill improvements)

## Current Status

- [x] Location data structure defined
- [x] Location JSON files created
- [ ] Shop system implementation
- [ ] Vault system implementation
- [ ] Inn/rest system implementation
- [ ] Location-specific UI screens

---

**Last Updated**: 2025-11-03
**Status**: Planning phase
**Related Docs**:
- `shop-pricing-system.md`
- `shop-system.md`
- `specialized-shops-design.md`
- `house-of-keeping-system.md` (vault)
- `tavern-performance-design.md`
- `tavern-npcs-todo.md`
