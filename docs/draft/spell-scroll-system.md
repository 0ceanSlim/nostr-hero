# Spell Scroll Crafting System

## Overview

A crafting system that allows magic-using classes to prepare spell scrolls at specific locations using components and gold. These scrolls can be used by any magic class in combat, bypassing spell slot limitations while still requiring mana.

**Key Benefits**:
- Wizards can craft scrolls for Wizard spells
- Clerics can craft scrolls for Cleric spells
- BUT any magic class can USE any prepared scroll (cross-class utility)
- No save file tracking needed - all gated by class/level/skill checks
- Location-based crafting adds exploration incentive
- Gold/component sink for economy balance

---

## Core Mechanics

### Crafting Flow
```
1. Player visits crafting location (Wizard Tower, Temple, etc.)
2. UI shows scrolls they CAN craft (filtered by class/level/intelligence)
3. Player selects scroll to craft
4. Backend validates: class, level, skill, components, gold, location
5. Scroll created and added to inventory
6. Components and gold consumed
```

### Usage Flow
```
1. Player equips spellbook (container) or keeps scrolls in inventory
2. During combat, scroll appears as usable item
3. Player uses scroll → casts spell (costs mana, consumes scroll)
4. ANY magic class can use ANY scroll (not limited to their class spells)
```

### Gating System (No Save Tracking)

**Recipe visibility determined by**:
- `class` - Only Wizards see Wizard scroll recipes, only Clerics see Cleric recipes
- `level` - Higher level scrolls require higher character level
- `intelligence` - Complex scrolls require minimum INT stat
- `location` - Must be at appropriate crafting station

**No need to track**: "known recipes", "discovered scrolls", etc. The API just checks current character stats.

### Success Chance System

Crafting scrolls has a **chance to fail** based on intelligence:

**Formula**:
```
success_chance = base_success_chance + ((player_int - min_int) * int_scaling_bonus)
Capped at maximum 95% (always some risk)
```

**Key Points**:
- Each spell has a `base_success_chance` (at minimum intelligence)
- Each spell has an `int_scaling_bonus` (% gained per INT point above minimum)
- Higher intelligence = better success rate, but never guaranteed
- Simple spells (level 1) have higher base success rates (75-85%)
- Complex spells (level 5+) have lower base success rates (50-60%)

**On Failure**:
- All components consumed (destroyed in failed attempt)
- All gold consumed (represents time/materials wasted)
- No scroll created
- Player shown failure message with encouragement to try again

**Example Success Rate Progressions**:

| Intelligence | Fireball (60% base, 3%/INT) | Magic Missile (80% base, 2%/INT) | Teleport (40% base, 4.5%/INT) |
|--------------|----------------------------|----------------------------------|-------------------------------|
| 13 (min)     | 60%                        | -                                | -                             |
| 10 (min)     | -                          | 80%                              | -                             |
| 16 (min)     | -                          | -                                | 40%                           |
| 14           | 63%                        | 88%                              | -                             |
| 15           | 66%                        | 90%                              | -                             |
| 16           | 69%                        | 92%                              | 40%                           |
| 18           | 75%                        | 95% (capped)                     | 49%                           |
| 20           | 81%                        | 95% (capped)                     | 58%                           |
| 22           | 87%                        | 95% (capped)                     | 67%                           |
| 24           | 93%                        | 95% (capped)                     | 76%                           |
| 26+          | 95% (capped)               | 95% (capped)                     | 85%                           |
| 28+          | 95% (capped)               | 95% (capped)                     | 94%                           |
| 30+          | 95% (capped)               | 95% (capped)                     | 95% (capped)                  |

**Key Insight**: Simple level 1 spells cap quickly, while complex high-level spells reward extreme intelligence investment.

---

## Data Structures

### Scroll Recipe (`game-data/magic/scroll-recipes/`)

Each recipe is a JSON file defining crafting requirements:

```json
{
  "id": "scroll-recipe-fireball",
  "spell_id": "fireball",
  "name": "Scroll of Fireball Recipe",
  "description": "Inscribe a Fireball spell onto a scroll for later use",

  "crafting_requirements": {
    "classes_allowed": ["wizard", "sorcerer"],
    "min_level": 5,
    "min_intelligence": 13,
    "locations_allowed": ["wizards-tower-goldenhaven", "wizards-tower-kingdom", "arcane-academy"]
  },

  "components": [
    {"item_id": "sulfur", "quantity": 3},
    {"item_id": "bat-guano", "quantity": 1},
    {"item_id": "parchment", "quantity": 1}
  ],

  "gold_cost": 50,
  "crafted_item_id": "scroll-fireball",

  "success_chance": {
    "base_chance": 0.60,
    "int_scaling_bonus": 0.03,
    "max_chance": 0.95
  }
}
```

### Prepared Scroll Item (`game-data/items/scrolls/`)

The actual item created when crafting succeeds:

```json
{
  "id": "scroll-fireball",
  "name": "Scroll of Fireball",
  "type": "scroll",
  "subtype": "arcane",
  "spell_id": "fireball",
  "description": "A prepared scroll containing the Fireball spell. Using it casts the spell and consumes the scroll. Can be used by any magic class.",
  "weight": 0.1,
  "value": 100,
  "stackable": false,
  "usable": true,
  "usable_in_combat": true,
  "mana_cost": 3,
  "image": "scroll-fire",
  "rarity": "uncommon"
}
```

### Component Items (Already Exist)

Using existing `spell-components` and other items:
- `sulfur`
- `bat-guano`
- `pearl-dust`
- `iron-filings`
- `holy-water`
- `mistletoe`
- `parchment` (base writing material for all scrolls)
- etc.

### Spellbook Container (`game-data/items/containers/`)

```json
{
  "id": "spellbook",
  "name": "Spellbook",
  "type": "container",
  "container_type": "spellbook",
  "description": "A leather-bound tome for storing prepared spell scrolls. Can hold up to 20 scrolls.",
  "weight": 3,
  "value": 50,
  "capacity": 20,
  "allowed_item_types": ["scroll"],
  "equip_slot": "offHand",
  "image": "spellbook",
  "rarity": "common"
}
```

---

## Curated Scroll List

### Wizard/Sorcerer Scrolls (Arcane)

| Scroll | Spell Level | Min Char Level | Min INT | Base Success | INT Scaling | Components | Gold Cost | Locations |
|--------|-------------|----------------|---------|--------------|-------------|------------|-----------|-----------|
| Magic Missile | 1 | 1 | 10 | 80% | +2%/INT | Iron Filings x2, Parchment | 25g | Any Wizard Tower |
| Shield | 1 | 1 | 10 | 85% | +2%/INT | Glass Shard, Parchment | 25g | Any Wizard Tower |
| Mage Armor | 1 | 1 | 10 | 80% | +2%/INT | Leather Scrap, Parchment | 20g | Any Wizard Tower |
| Detect Magic | 1 | 1 | 10 | 85% | +2%/INT | Crystal Shard, Parchment | 15g | Any Wizard Tower, Library |
| Identify | 1 | 1 | 12 | 70% | +2.5%/INT | Pearl Dust x1, Parchment | 30g | Wizard Tower, Library |
| Fireball | 3 | 5 | 13 | 60% | +3%/INT | Sulfur x3, Bat Guano, Parchment | 50g | Wizard Tower |
| Lightning Bolt | 3 | 5 | 13 | 60% | +3%/INT | Storm Essence, Copper Wire, Parchment | 60g | Wizard Tower |
| Counterspell | 3 | 5 | 14 | 55% | +3.5%/INT | Null Stone Fragment, Parchment | 100g | Arcane Academy |
| Haste | 3 | 5 | 13 | 65% | +3%/INT | Quicksilver, Parchment | 70g | Wizard Tower |
| Fly | 3 | 5 | 13 | 65% | +3%/INT | Wing of Bat, Parchment | 50g | Wizard Tower |
| Cone of Cold | 5 | 9 | 15 | 50% | +4%/INT | Frost Crystal x3, Parchment | 120g | Ice Mage Tower |
| Wall of Force | 5 | 9 | 15 | 45% | +4%/INT | Adamantine Dust, Parchment | 150g | Arcane Academy |
| Teleport | 7 | 13 | 16 | 40% | +4.5%/INT | Lodestone x2, Parchment | 250g | Arcane Academy |

### Cleric/Paladin Scrolls (Divine)

| Scroll | Spell Level | Min Char Level | Min INT | Base Success | INT Scaling | Components | Gold Cost | Locations |
|--------|-------------|----------------|---------|--------------|-------------|------------|-----------|-----------|
| Cure Wounds | 1 | 1 | 10 | 85% | +2%/INT | Holy Water, Parchment | 30g | Temple, Shrine |
| Bless | 1 | 1 | 10 | 80% | +2%/INT | Sacred Incense, Parchment | 25g | Temple, Shrine |
| Shield of Faith | 1 | 1 | 10 | 80% | +2%/INT | Silver Dust, Parchment | 25g | Temple |
| Healing Word | 1 | 1 | 10 | 85% | +2%/INT | Dried Herbs, Parchment | 20g | Temple, Druid Grove |
| Lesser Restoration | 2 | 3 | 11 | 70% | +2.5%/INT | Diamond Dust x1, Parchment | 40g | Temple |
| Prayer of Healing | 2 | 3 | 11 | 70% | +2.5%/INT | Holy Water x2, Parchment | 50g | Temple |
| Revivify | 3 | 5 | 13 | 50% | +3.5%/INT | Diamond (300gp value), Parchment | 300g | Major Temple only |
| Dispel Magic | 3 | 5 | 13 | 60% | +3%/INT | Powdered Iron, Parchment | 60g | Temple |
| Remove Curse | 3 | 5 | 13 | 60% | +3%/INT | Sacred Oil, Parchment | 70g | Temple |
| Greater Restoration | 5 | 9 | 15 | 45% | +4%/INT | Diamond Dust x5, Parchment | 200g | Major Temple only |

### Druid/Ranger Scrolls (Nature)

| Scroll | Spell Level | Min Char Level | Min INT | Base Success | INT Scaling | Components | Gold Cost | Locations |
|--------|-------------|----------------|---------|--------------|-------------|------------|-----------|-----------|
| Entangle | 1 | 1 | 10 | 80% | +2%/INT | Vine Clippings, Parchment | 20g | Druid Grove |
| Goodberry | 1 | 1 | 10 | 85% | +2%/INT | Berries x5, Parchment | 15g | Druid Grove |
| Speak with Animals | 1 | 1 | 10 | 85% | +2%/INT | Animal Fur, Parchment | 20g | Druid Grove |
| Barkskin | 2 | 3 | 11 | 75% | +2.5%/INT | Oak Bark, Parchment | 30g | Druid Grove |
| Pass Without Trace | 2 | 3 | 11 | 70% | +2.5%/INT | Ashes of Mistletoe, Parchment | 40g | Druid Grove |
| Call Lightning | 3 | 5 | 13 | 60% | +3%/INT | Storm Water, Parchment | 50g | Druid Grove (outdoor) |
| Plant Growth | 3 | 5 | 13 | 65% | +3%/INT | Seeds x10, Parchment | 40g | Druid Grove |

---

## Crafting Locations

### Location Types and Available Recipes

| Location ID | Type | Scrolls Available |
|-------------|------|-------------------|
| `wizards-tower-goldenhaven` | Wizard Tower | All Wizard/Sorcerer scrolls (level appropriate) |
| `wizards-tower-kingdom` | Wizard Tower | All Wizard/Sorcerer scrolls (level appropriate) |
| `arcane-academy` | Advanced Arcane | High-level Wizard scrolls (level 5+) |
| `temple-kingdom` | Temple | All Cleric/Paladin scrolls |
| `temple-goldenhaven` | Temple | All Cleric/Paladin scrolls |
| `major-temple-capital` | Major Temple | All Divine scrolls + Revivify/Greater Restoration |
| `druid-grove-forest` | Druid Grove | All Druid/Ranger scrolls |
| `ice-mage-tower` | Specialized Arcane | Cone of Cold, other ice spells |

### Location Data Addition

Add to existing location JSON files:

```json
{
  "id": "wizards-tower-goldenhaven",
  "name": "Wizard's Tower",
  "crafting_stations": [
    {
      "type": "scroll_crafting_arcane",
      "description": "A well-stocked inscription desk with arcane focuses"
    }
  ]
}
```

---

## Component Sourcing

### Purchasable Components (Shops)

| Item | Shop Type | Price | Locations |
|------|-----------|-------|-----------|
| Parchment | General Store, Scribe | 5g | Most cities |
| Iron Filings | Blacksmith | 3g | Most cities |
| Copper Wire | Blacksmith | 5g | Most cities |
| Holy Water | Temple | 10g | Religious locations |
| Dried Herbs | Apothecary | 5g | Most cities |
| Glass Shard | General Store | 2g | Most cities |
| Leather Scrap | Tanner | 3g | Most cities |

### Monster Drops

| Item | Monster Type | Drop Rate | Locations |
|------|--------------|-----------|-----------|
| Bat Guano | Giant Bat | 60% | Caves, Ruins |
| Sulfur | Fire Elemental | 40% | Volcanic areas |
| Wing of Bat | Giant Bat | 30% | Caves |
| Storm Essence | Air Elemental | 50% | Mountains, Storm zones |
| Frost Crystal | Ice Elemental | 50% | Frozen areas |
| Null Stone Fragment | Magical Constructs | 20% | Wizard towers (enemy) |

### Gathering (Future Feature)

| Item | Gathering Node | Locations |
|------|----------------|-----------|
| Mistletoe | Ancient Oak Trees | Druid groves |
| Berries | Berry Bushes | Forests |
| Oak Bark | Oak Trees | Forests |
| Vine Clippings | Vines | Jungles, Forests |

### Quest Rewards

| Item | Quest Giver | Quest Type |
|------|-------------|------------|
| Diamond Dust x5 | High Priest | Major divine quest |
| Adamantine Dust | Archmage | Arcane research quest |
| Lodestone x2 | Merchant Guild | Trade route quest |

---

## Backend API Design

### Endpoint: Get Available Scroll Recipes

**Request**: `GET /api/scrolls/recipes?class={class}&level={level}&intelligence={int}&location={location_id}`

**Response**:
```json
{
  "success": true,
  "recipes": [
    {
      "id": "scroll-recipe-fireball",
      "spell_id": "fireball",
      "name": "Scroll of Fireball Recipe",
      "components": [
        {"item_id": "sulfur", "quantity": 3, "player_has": 5},
        {"item_id": "bat-guano", "quantity": 1, "player_has": 2},
        {"item_id": "parchment", "quantity": 1, "player_has": 3}
      ],
      "gold_cost": 50,
      "success_chance": {
        "player_chance": 0.66,
        "base_chance": 0.60,
        "int_bonus": 0.06,
        "player_int": 15,
        "min_int": 13
      },
      "can_craft": true,
      "missing_requirements": []
    }
  ]
}
```

### Endpoint: Craft Scroll

**Request**: `POST /api/scrolls/craft`

**Body**:
```json
{
  "npub": "npub1...",
  "save_id": "save_1234567890",
  "recipe_id": "scroll-recipe-fireball",
  "location_id": "wizards-tower-goldenhaven"
}
```

**Backend Validation & Crafting**:
1. Load character from save
2. Load recipe from database
3. Check class matches `classes_allowed`
4. Check level >= `min_level`
5. Check intelligence >= `min_intelligence`
6. Check location matches `locations_allowed`
7. Check inventory has all components
8. Check player has gold_cost
9. Calculate success chance: `base_chance + ((player_int - min_int) * int_scaling)` (capped at 95%)
10. Roll random number (0.0 to 1.0)
11. If roll <= success_chance: SUCCESS
    - Add crafted scroll to inventory
    - Remove components and gold
    - Return success response
12. If roll > success_chance: FAILURE
    - Remove components and gold (wasted in failed attempt)
    - Return failure response
13. Save game state

**Success Response** (scroll crafted):
```json
{
  "success": true,
  "crafted": true,
  "message": "Successfully crafted Scroll of Fireball!",
  "scroll_created": {
    "id": "scroll-fireball",
    "name": "Scroll of Fireball"
  },
  "success_chance": 0.66,
  "new_gold": 184,
  "components_consumed": [
    {"item_id": "sulfur", "quantity": 3},
    {"item_id": "bat-guano", "quantity": 1},
    {"item_id": "parchment", "quantity": 1}
  ]
}
```

**Failure Response** (crafting failed, components lost):
```json
{
  "success": true,
  "crafted": false,
  "message": "Crafting failed! The scroll's inscription faded as you wrote it. Your components were consumed in the failed attempt.",
  "success_chance": 0.66,
  "new_gold": 184,
  "components_consumed": [
    {"item_id": "sulfur", "quantity": 3},
    {"item_id": "bat-guano", "quantity": 1},
    {"item_id": "parchment", "quantity": 1}
  ]
}
```

**Error Response** (validation failed, nothing consumed):
```json
{
  "success": false,
  "error": "Insufficient intelligence (requires 13, you have 10)"
}
```

---

## Frontend UI Design

### Crafting Interface

When player clicks "Craft Scrolls" at a valid location:

```
┌────────────────────────────────────────────────────────────┐
│  📜 Scroll Crafting - Wizard's Tower                       │
│  Your Intelligence: 15 (+2 modifier)                       │
├────────────────────────────────────────────────────────────┤
│  Available Recipes (Filtered by your class/level/INT)      │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐ │
│  │ [🔥] Scroll of Fireball                              │ │
│  │      Level 5+ | Intelligence 13+ | Wizard/Sorcerer   │ │
│  │      Success Chance: 66% (Base 60% + 6% from INT)    │ │
│  │      Components:                                      │ │
│  │      ✅ Sulfur x3 (you have: 5)                       │ │
│  │      ✅ Bat Guano x1 (you have: 2)                    │ │
│  │      ✅ Parchment x1 (you have: 3)                    │ │
│  │      Gold: 50g (you have: 234g)                       │ │
│  │      ⚠️ Failure consumes all components and gold      │ │
│  │      [Craft] [Examine Spell]                          │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐ │
│  │ [⚡] Scroll of Lightning Bolt                         │ │
│  │      Level 5+ | Intelligence 13+ | Wizard/Sorcerer   │ │
│  │      Success Chance: 66% (Base 60% + 6% from INT)    │ │
│  │      Components:                                      │ │
│  │      ❌ Storm Essence x1 (you have: 0)                │ │
│  │      ✅ Copper Wire x1 (you have: 3)                  │ │
│  │      ✅ Parchment x1 (you have: 3)                    │ │
│  │      Gold: 60g (you have: 234g)                       │ │
│  │      [Craft] (disabled - missing components)          │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐ │
│  │ [🌟] Scroll of Magic Missile                         │ │
│  │      Level 1+ | Intelligence 10+ | Wizard/Sorcerer   │ │
│  │      Success Chance: 90% (Base 80% + 10% from INT)   │ │
│  │      Components:                                      │ │
│  │      ✅ Iron Filings x2 (you have: 4)                 │ │
│  │      ✅ Parchment x1 (you have: 3)                    │ │
│  │      Gold: 25g (you have: 234g)                       │ │
│  │      [Craft] [Examine Spell]                          │ │
│  └──────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────┘
```

### Spellbook Container View

```
┌────────────────────────────────────────────────────────────┐
│  📖 Spellbook (8/20 scrolls)                               │
├────────────────────────────────────────────────────────────┤
│  [🔥 Fireball]        [🔥 Fireball]      [⚡ Lightning]    │
│  [🛡️ Shield]          [🎯 Magic Missile] [✨ Identify]     │
│  [💚 Cure Wounds]     [🌟 Bless]         [Empty]           │
│  [Empty]              [Empty]            [Empty]           │
└────────────────────────────────────────────────────────────┘

Right-click: Use in Combat | Move to Inventory | Examine
```

---

## Implementation Phases

### Phase 1: Data Setup
- [ ] Create scroll recipe JSON files (start with 15-20 recipes)
- [ ] Create prepared scroll item JSON files
- [ ] Create spellbook container item
- [ ] Verify all component items exist (parchment, spell-components, etc.)
- [ ] Update location JSONs with crafting_stations

### Phase 2: Database Migration
- [ ] Add `scroll_recipes` table to migration
- [ ] Add `crafting_stations` field to locations table
- [ ] Migrate scroll recipe data on startup

### Phase 3: Backend API
- [ ] `GET /api/scrolls/recipes` - Filter and return available recipes
- [ ] `POST /api/scrolls/craft` - Validate and execute crafting
- [ ] Add validation logic (class/level/int/location/components/gold)

### Phase 4: Frontend UI
- [ ] Create crafting interface component
- [ ] Add "Craft Scrolls" button to locations with crafting stations
- [ ] Display available recipes with component status
- [ ] Handle craft button click → API call → update inventory

### Phase 5: Container Integration
- [ ] Implement spellbook container logic (like existing backpack)
- [ ] Allow scrolls to be stored in spellbook
- [ ] UI for spellbook contents

### Phase 6: Combat Integration (Later)
- [ ] Add scroll usage to combat action menu
- [ ] Validate mana cost on use
- [ ] Consume scroll on successful cast
- [ ] Use existing spell effect logic

---

## Balance Considerations

### Preventing Scroll Spam

**Success Chance**: Intelligence-based crafting success (40-85% base, up to 95% max)
- Lower INT characters will fail frequently
- Even high INT characters can fail (95% cap means 5% failure)
- Failed attempts still consume components and gold
- Creates risk/reward tension

**Mana Cost**: Scrolls still cost mana to use (prevents unlimited spam)

**Gold Sink**: Higher level scrolls are expensive (50g-300g each)
- Multiply by success rate: might need 2-3 attempts to succeed
- Real cost of Fireball scroll at 66% success: ~75g average

**Component Scarcity**: Rare components require exploration/combat
- Bat Guano: Need to farm Giant Bats
- Storm Essence: Need to fight Air Elementals
- Diamonds: Expensive to buy or quest rewards
- Failed crafts make rare components even more valuable

**Location Restriction**: Must return to city/tower to craft (can't craft in dungeon)

**Inventory Space**: Scrolls take space (can't hoard 100 fireballs)

### Advantages Over Spell Slots

✅ **Cross-class usage** - Fighter can use Fireball scroll (if they have mana somehow)

✅ **No slot limit** - Cast 5 Fireballs if you have 5 scrolls and mana

✅ **Preparation** - Stock up before tough boss fight

✅ **Trading** - Can give/sell scrolls to other players (future multiplayer)

### Disadvantages vs Spell Slots

❌ **Expensive** - Costs gold + components + time to craft

❌ **Consumable** - One-time use, then gone

❌ **Still needs mana** - Can't bypass mana system

❌ **Requires preparation** - Must plan ahead and return to cities

---

## Example Scenarios

### Scenario 1: Wizard Preparing for Dungeon

1. Player is level 5 Wizard with INT 15
2. Visits Wizard Tower in Goldenhaven
3. Opens crafting interface, sees Fireball scroll recipe
4. Fireball shows 66% success chance (60% base + 6% from INT)
5. Has components for Fireball, attempts to craft
6. **Attempt 1**: Success! Scroll created (50g + components consumed)
7. **Attempt 2**: Failure! Components and gold wasted, no scroll created
8. Farms more components, tries again
9. **Attempt 3**: Success! Second scroll created
10. Stores both scrolls in spellbook (equipped in offHand)
11. Heads to dungeon
12. During boss fight, uses both Fireball scrolls + regular spell slots
13. Survives thanks to extra firepower

### Scenario 2: Cleric Helping Party

1. Player is level 7 Cleric
2. Crafts 3x Cure Wounds scrolls, 2x Lesser Restoration scrolls
3. Gives 1 Cure Wounds scroll to party's Fighter
4. Fighter gets poisoned in battle, uses scroll (doesn't have mana, but scroll provides it? or requires mana potion?)
5. **Note**: Need to decide if non-casters can use scrolls without mana

### Scenario 3: Component Farming & Intelligence Scaling

1. Player is level 9 Wizard with INT 18
2. Wants to craft Cone of Cold scroll (requires INT 15)
3. Success chance: 50% base + ((18-15) * 4%) = 62%
4. Needs Frost Crystal x3 (from Ice Elementals)
5. Travels to Frozen Wastes
6. Defeats Ice Elementals, collects 6 Frost Crystals
7. Returns to Ice Mage Tower
8. **First attempt**: Success! (62% chance paid off)
9. Player levels up to 10, increases INT to 20
10. New success chance: 50% + ((20-15) * 4%) = 70%
11. Crafts second Cone of Cold scroll with higher success rate

---

## Design Decisions Made

### ✅ Success Chance System (Implemented)

**Decided**: Intelligence-based success chance with scaling
- Each spell has base success rate (at minimum INT)
- Success improves with higher INT (+2-4.5% per point above minimum)
- Capped at 95% maximum (always some risk)
- Failure consumes all components and gold
- Creates meaningful risk/reward and incentivizes INT investment

### Open Questions (Still To Decide)

### 1. Non-Caster Classes Using Scrolls?

**Option A**: Any class can use scrolls if they have mana (strict D&D)
- Problem: Most non-casters have 0 mana
- Solution: Mana potions become valuable for scroll use

**Option B**: Only magic classes can use scrolls (simpler)
- Wizard, Sorcerer, Cleric, Druid, Paladin, Ranger
- Fighter, Barbarian, Rogue cannot use scrolls

**Recommendation**: Start with Option B (magic classes only), add Option A later if mana potions exist

### 2. Scroll Storage Limits?

**Option A**: Spellbook required (holds 20 scrolls)
- Forces players to manage inventory
- Spellbook takes equipment slot

**Option B**: Scrolls can go anywhere in inventory
- Simpler, more flexible
- Less realistic

**Recommendation**: Option A (spellbook container) - adds strategic depth

### 3. Pre-made Scrolls as Loot?

**Option A**: Players can ONLY get scrolls by crafting

**Option B**: Rarely find pre-made scrolls in loot
- Wizard enemies might drop 1-2 scrolls
- Treasure chests might have scrolls
- Makes scrolls more accessible early game

**Recommendation**: Option B - rare loot drops make scrolls accessible before players can craft

---

## Future Enhancements

### Phase 2 Features (Post-MVP)

**Scroll Variants**:
- Empowered Scrolls (cast at higher level, cost more components)
- Quickcast Scrolls (use as bonus action, very expensive)

**Advanced Crafting**:
- Batch crafting (craft 5 scrolls at once, bulk discount)
- Crafting skill progression (faster crafting, lower costs)

**NPC Services**:
- Hire NPC to craft scrolls for you (at premium price)
- NPC transmutation services (convert components)

**Spellbook Upgrades**:
- Apprentice Spellbook (10 scrolls)
- Master Spellbook (30 scrolls)
- Legendary Spellbook (50 scrolls, reduces mana cost by 1)

**Trading/Economy**:
- Sell scrolls to NPCs or other players
- Scroll merchants in major cities
- Dynamic pricing based on rarity

---

## Summary

This spell scroll system adds:
- **Crafting depth** - Meaningful component gathering and gold sink
- **Intelligence-based progression** - Higher INT = better success rates, incentivizes stat investment
- **Risk/reward gameplay** - Success chance (40-95%) creates tension and makes scrolls feel earned
- **Strategic preparation** - Plan ahead for tough encounters, account for crafting failures
- **Cross-class utility** - Clerics can use Wizard scrolls and vice versa
- **Exploration incentive** - Travel to specific locations for rare scrolls and components
- **No save bloat** - All gated by current stats (class/level/INT), no recipe tracking needed

**Key Design Philosophy**:
- Scrolls are powerful but risky to craft
- Intelligence investment directly improves success rates
- Failure is punishing (lose components) but teaches resource management
- Creates meaningful gold and component sinks for economy

Next steps: Implement Phase 1 (data setup) with ~20 core scroll recipes covering combat, healing, and utility spells.
