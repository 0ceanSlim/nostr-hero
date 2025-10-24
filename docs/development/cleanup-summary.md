# Location Cleanup Summary

## What Was Done

All placeholder NPCs have been removed from location files. Only the 6 implemented Keeper NPCs remain.

## Current State

### NPCs Remaining (Implemented)
- ✅ `royal-custodian` (Kingdom)
- ✅ `scalekeeper` (City-East / Goldenhaven)
- ✅ `warden-of-roots` (City-South / Verdant)
- ✅ `vaultwright` (Town-Northeast / Ironpeak)
- ✅ `keywarden` (Village-West / Millhaven)
- ✅ `hoardkeeper` (Village-Southwest / Marshlight)

### NPCs Removed (58 total)
All placeholder NPCs have been removed from the following locations:
- Kingdom (5 districts) - removed 15 NPCs
- City-East / Goldenhaven (4 districts) - removed 7 NPCs
- City-South / Verdant (4 districts) - removed 6 NPCs
- Town-Northeast / Ironpeak (2 districts) - removed 4 NPCs
- Town-North / Frosthold (2 districts) - removed 4 NPCs
- Village-West / Millhaven (2 districts) - removed 6 NPCs
- Village-Southwest / Marshlight (2 districts) - removed 4 NPCs
- Village-South / Saltwind (2 districts) - removed 4 NPCs
- Village-Southeast / Dusthaven (2 districts) - removed 4 NPCs

### Buildings Status
**All buildings remain intact** - only NPC references were removed. Buildings still exist with their descriptions and actions, they just don't have NPCs staffing them yet.

## Next Steps for Implementation

### Priority 1: Core Shop NPCs
Players need to buy/sell items for basic gameplay:
- Weapon/armor shops (blacksmiths)
- General goods stores
- Magic/spell shops
- Specialty shops (herbalism, mining, etc.)

**Estimated**: 15 shop keeper NPCs

### Priority 2: Rest & Recovery
Players need to rest and recover:
- Inn keepers
- Tavern keepers

**Estimated**: 7 inn/tavern NPCs

### Priority 3: Travel & Navigation
For world exploration:
- Cart/caravan masters (fast travel)
- Guides (environment navigation)
- Harbor masters (boat travel)

**Estimated**: 7 transport NPCs

### Priority 4: Quests & Guilds
For quest system:
- Guild masters (Adventurers, Merchants, Druids, etc.)
- Town guards/officials (entry control, quest givers)

**Estimated**: 11 guild/official NPCs

### Priority 5: Services
Banking, research, crafting support:
- Bank clerks (gold storage - different from item storage)
- Librarians (research/lore)
- Craftsmen (commission items)

**Estimated**: 6 service NPCs

### Optional: Flavor NPCs
Atmosphere and immersion (can be added last or omitted):
- Village elders
- Scholars
- Priests/shamans
- Flavor merchants

**Estimated**: 12 flavor NPCs

## Implementation Pattern

Based on the Keeper NPCs, the pattern for implementing new NPCs is:

1. **Create NPC JSON** in `docs/data/content/npcs/{location}/{npc-id}.json`
2. **Add to location** in the building's `npc` field and district's `npcs` array
3. **Backend support**:
   - Add NPC loading to database
   - Create API endpoint for NPC data
   - Implement dialogue system
   - Implement NPC-specific actions (shop, inn, transport, etc.)

## Files Modified

### Location Files Cleaned:
- `docs/data/content/locations/cities/kingdom.json`
- `docs/data/content/locations/cities/city-east.json`
- `docs/data/content/locations/cities/city-south.json`
- `docs/data/content/locations/cities/town-northeast.json`
- `docs/data/content/locations/cities/town-north.json`
- `docs/data/content/locations/cities/village-west.json`
- `docs/data/content/locations/cities/village-southwest.json`
- `docs/data/content/locations/cities/village-south.json`
- `docs/data/content/locations/cities/village-southeast.json`

### Documentation Created:
- `docs/development/house-of-keeping-system.md` - Complete system documentation
- `docs/development/location-audit.md` - Full NPC inventory and categorization
- `docs/development/cleanup-summary.md` - This file

## Recommendation for Development

**Phase 1 (MVP)**: Implement 5-10 core shop NPCs in main cities
- Focus on Kingdom and starting cities
- Bare minimum: weapon shop, general store, inn
- Get buy/sell/rest mechanics working

**Phase 2**: Expand to all starting cities
- Ensure each starting location has basic services
- Add guild NPCs for quest system
- Implement travel NPCs for world exploration

**Phase 3**: Fill out secondary locations
- Add NPCs to non-starting towns/villages
- Add flavor NPCs for immersion
- Polish dialogue and interactions

This cleanup gives you a clean slate to implement NPCs systematically without placeholder clutter!
