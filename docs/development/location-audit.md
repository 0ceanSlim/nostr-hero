# Location Content Audit

## Overview
This document inventories all buildings and NPCs currently defined in location files, categorizing them by implementation status.

## Status Categories
- ✅ **IMPLEMENTED**: NPC JSON exists with full dialogue
- 🏗️ **ESSENTIAL**: Core gameplay NPCs that need implementation
- 🎨 **FLAVOR**: Optional atmosphere NPCs for immersion
- ❌ **REMOVE**: Redundant or unnecessary placeholders

---

## Kingdom (Human Starting City)

### Center District
**Buildings:**
- ✅ Vault of Crowns (house_of_keeping) → royal-custodian
- Royal Palace (landmark) - inaccessible
- Grand Market (shop:general)
- Adventurers Guild (guild)

**NPCs:**
- ✅ royal-custodian (IMPLEMENTED)
- 🏗️ royal_guard (ESSENTIAL) - Palace guards, entry fee enforcement
- 🏗️ merchant_leader (ESSENTIAL) - Grand Market shop keeper
- 🏗️ guild_master (ESSENTIAL) - Quest giver, adventurer registration

### North District
**Buildings:**
- Royal Library (library)
- Noble Quarter (residential) - inaccessible
- Mage Tower (magic_shop)

**NPCs:**
- 🏗️ court_wizard (ESSENTIAL) - Magic shop, spell purchases
- 🏗️ librarian (ESSENTIAL) - Research, lore information
- 🎨 noble_messenger (FLAVOR) - Quest delivery, flavor

### South District
**Buildings:**
- Main Gate (gate)
- The Wayward Traveler Inn (inn)
- Royal Stables (transport)

**NPCs:**
- 🏗️ gate_captain (ESSENTIAL) - Entry/exit to forest environment
- 🏗️ innkeeper (ESSENTIAL) - Rest, food, rumors
- 🏗️ cart_driver (ESSENTIAL) - Fast travel between cities

### East District
**Buildings:**
- Grand Forge (blacksmith)
- Artisan Quarter (crafting)
- Royal Mint (bank)

**NPCs:**
- 🏗️ master_smith (ESSENTIAL) - Weapons/armor shop, repairs
- 🏗️ bank_clerk (ESSENTIAL) - Gold deposit/withdrawal (different from storage)
- 🎨 artisan_leader (FLAVOR) - Crafting lore, commission items

### West District
**Buildings:**
- Western Trading Post (shop:exotic)
- The Golden Griffin Tavern (tavern)
- Caravan Guild Office (transport)

**NPCs:**
- 🏗️ caravan_master (ESSENTIAL) - Long-distance travel, caravan jobs
- 🏗️ tavern_keeper (ESSENTIAL) - Drinks, rumors, music
- 🏗️ exotic_trader (ESSENTIAL) - Rare items shop

---

## City-East / Goldenhaven (Dragonborn Starting City)

### Center District (Golden Square)
**Buildings:**
- ✅ Ember Vault (house_of_keeping) → scalekeeper
- Grand Exchange (bank)
- Golden Bazaar (shop:luxury)
- Merchant Princes Guild (guild)

**NPCs:**
- ✅ scalekeeper (IMPLEMENTED)
- 🏗️ guild_prince (ESSENTIAL) - Merchant guild, trade contracts
- 🏗️ exchange_master (ESSENTIAL) - Currency exchange, banking
- 🏗️ luxury_dealer (ESSENTIAL) - High-end items shop

### West District (Western Gate)
**Buildings:**
- Customs House (civic)
- Caravan Rest Inn (inn)

**NPCs:**
- 🏗️ customs_officer (ESSENTIAL) - Entry fees, inspections
- 🏗️ caravan_innkeeper (ESSENTIAL) - Rest, caravan services

### North District (Mountain District)
**Buildings:**
- Mountain Ore Exchange (shop:mining)

**NPCs:**
- 🏗️ ore_broker (ESSENTIAL) - Mining goods, ore sales

### East District (Eastern Port)
**Buildings:**
- Harbor Master's Office (transport)
- The Salty Anchor (tavern)

**NPCs:**
- 🏗️ harbor_master (ESSENTIAL) - Ship travel, cargo
- 🎨 ship_captain (FLAVOR) - Sailing stories, distant lands

---

## City-South / Verdant City (Elf Starting City)

### Center District (Garden Plaza)
**Buildings:**
- ✅ Glade of Safekeeping (house_of_keeping) → warden-of-roots
- The Verdant Archives (library)
- Circle of the Ancient Oak (guild:druid)
- Garden of Remedies (shop:herbalism)

**NPCs:**
- ✅ warden-of-roots (IMPLEMENTED)
- 🏗️ chief_druid (ESSENTIAL) - Druid guild, nature magic
- 🏗️ master_herbalist (ESSENTIAL) - Herbal shop, potions
- 🎨 nature_scholar (FLAVOR) - Library research, nature lore

### North District (Forest Quarter)
**Buildings:**
- Forest Rangers Lodge (guild)

**NPCs:**
- 🏗️ ranger_captain (ESSENTIAL) - Guides, forest navigation

### West District (Artisan Quarter)
**Buildings:**
- Living Wood Workshop (crafting)

**NPCs:**
- 🏗️ master_woodcarver (ESSENTIAL) - Natural items, staves

### South District (Coastal Road)
**Buildings:**
- Seaside Trading Post (shop:coastal)

**NPCs:**
- 🏗️ coastal_trader (ESSENTIAL) - Coastal goods shop

---

## Town-Northeast / Ironpeak (Dwarf Starting Town)

### Center District (Miners' Plaza)
**Buildings:**
- ✅ Stonevault (house_of_keeping) → vaultwright
- The Pickaxe & Pint (inn)
- Mountain Ore Exchange (shop:mining)
- Miners' Guild Hall (guild)

**NPCs:**
- ✅ vaultwright (IMPLEMENTED)
- 🏗️ guild_foreman (ESSENTIAL) - Mining guild, claims
- 🏗️ ore_merchant (ESSENTIAL) - Ore shop
- 🏗️ tavern_keeper (ESSENTIAL) - Inn services

### South District (Valley Approach)
**Buildings:**
- Mine Cart Station (transport)

**NPCs:**
- 🏗️ cart_master (ESSENTIAL) - Mine cart transport

---

## Town-North (Unmapped Starting Race)

### Center District
**Buildings:**
- General Store (shop)
- The Northern Hearth Inn (inn)
- Town Hall (civic)

**NPCs:**
- 🎨 town_elder (FLAVOR) - Town leadership, local quests
- 🏗️ fur_trader (ESSENTIAL) - Cold weather gear shop
- 🏗️ tavern_keeper (ESSENTIAL) - Inn services

### South District (Southern Gate)
**Buildings:**
- Guard Tower (military)

**NPCs:**
- 🏗️ guard_captain (ESSENTIAL) - Gate control

---

## Village-West / Millhaven (Halfling/Gnome Starting Village)

### Center District
**Buildings:**
- ✅ Burrowlock (house_of_keeping) → keywarden
- The Grain & Grape Inn (inn)
- Haversham's General Store (shop:general)
- Shrine of the Harvest (shrine)

**NPCs:**
- ✅ keywarden (IMPLEMENTED)
- 🎨 village_elder (FLAVOR) - Local quests, village history
- 🏗️ innkeeper (ESSENTIAL) - Inn services
- 🏗️ shopkeeper (ESSENTIAL) - General goods shop
- 🎨 local_priest (FLAVOR) - Blessings, shrine services

### East District (Eastern Farms)
**Buildings:**
- Millwright's Farm (farm)
- Crossroads Stone (landmark)

**NPCs:**
- 🎨 village_farmer (FLAVOR) - Farm goods, rural life
- 🎨 traveling_merchant (FLAVOR) - Random goods

---

## Village-Southwest / Marshlight (Orc Starting Village)

### Center District (Stilt Houses)
**Buildings:**
- ✅ Warhoard (house_of_keeping) → hoardkeeper
- The Glowing Lantern (inn)
- Swamp Treasures (shop:swamp_gear)
- Shrine of Peaceful Spirits (shrine)

**NPCs:**
- ✅ hoardkeeper (IMPLEMENTED)
- 🎨 village_shaman (FLAVOR) - Shrine keeper, spiritual guidance
- 🏗️ bog_trader (ESSENTIAL) - Swamp gear shop
- 🏗️ marsh_guide (ESSENTIAL) - Swamp navigation

### East District (Eastern Dock)
**Buildings:**
- Marsh Boat Rental (transport)

**NPCs:**
- 🏗️ boat_keeper (ESSENTIAL) - Boat transport

---

## Village-South (Unmapped Starting Race)

### Center District (Harbor Village)
**Buildings:**
- Fisher's Market (shop)
- The Anchor Inn (inn)
- Sea Temple (shrine)

**NPCs:**
- 🏗️ harbor_master (ESSENTIAL) - Boat services
- 🎨 old_fisherman (FLAVOR) - Fishing tales, local lore
- 🎨 sea_priest (FLAVOR) - Temple services

### East District (Trade Road)
**Buildings:**
- Trading Post (shop)

**NPCs:**
- 🏗️ trade_merchant (ESSENTIAL) - General shop

---

## Village-Southeast (Unmapped Starting Race)

### Center District (Desert Village)
**Buildings:**
- Oasis Inn (inn)
- Desert Supplies (shop)
- Village Well (landmark)

**NPCs:**
- 🎨 village_elder (FLAVOR) - Local leadership
- 🏗️ desert_guide (ESSENTIAL) - Desert navigation
- 🏗️ water_keeper (ESSENTIAL) - Water supplies

### West District (Caravan Stop)
**Buildings:**
- Caravan Rest (transport)

**NPCs:**
- 🏗️ caravan_master (ESSENTIAL) - Caravan travel

---

## Summary Statistics

### By Status
- ✅ **Implemented**: 6 NPCs (all Keepers)
- 🏗️ **Essential**: 46 NPCs (core gameplay)
- 🎨 **Flavor**: 12 NPCs (optional atmosphere)
- **Total Placeholder NPCs**: 58

### By Function
**Essential NPCs by Role:**
- Shop Keepers: 15
- Guild Masters: 6
- Inn/Tavern Keepers: 7
- Transport/Travel: 7
- Guards/Officials: 5
- Services (library, bank, etc.): 6

### Recommendation
**Phase 1 - Core Services (Priority):**
1. All shop NPCs (players need to buy/sell)
2. Inn/tavern keepers (rest mechanics)
3. Guild masters (quest system)

**Phase 2 - Travel & Exploration:**
1. Transport NPCs (fast travel)
2. Guides (environment navigation)
3. Gate/customs officials (access control)

**Phase 3 - Flavor & Immersion:**
1. Flavor NPCs (can be added later)
2. Or simply remove if not needed for MVP
