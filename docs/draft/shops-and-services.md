# Shops and Services System

Consolidated documentation for all shop and service systems.

**Status**: Planning/Design phase - Not yet implemented
**Priority**: High - Core gameplay feature
**Last Updated**: 2025-11-03

---

## Table of Contents

- [General Store System](#general-store-system)
- [Shop Pricing](#shop-pricing)
- [Shop Economy](#shop-economy)
- [Specialized Shops](#specialized-shops)
- [Tavern/Inn System](#taverninn-system)
- [Implementation Plan](#implementation-plan)

---

## General Store System

### Core Mechanics

**Shop Gold System**:
- Merchants start with gold reserve (varies by location)
- Max gold cap prevents infinite accumulation
- Gold regenerates daily
- When out of gold, merchant can't buy from players

**Inventory Stock System**:
- Each item has max stock quantity
- Stock depletes when purchased
- Items restock daily (or configured interval)
- Out of stock = cannot purchase

**Player Transactions**:
- **Sell to merchant**: Player gets buy_price (typically 40-60% of base value)
- **Buy from merchant**: Player pays sell_price (typically 90-120% of base value)

### General Merchants by Location

#### 1. Kingdom - Aldric Goodsworth
- **Location**: Grand Market
- **Race**: Human
- **Personality**: Cheerful, welcoming
- **Economy**:
  - Starting Gold: 500g
  - Max Gold: 1000g
  - Regen: 100g/day
  - Buy: 0.5× | Sell: 1.0×
- **Specialty**: Balanced general supplies

#### 2. Goldenhaven (City-East) - Aurelia Goldscale
- **Location**: Golden Square
- **Race**: Dragonborn
- **Personality**: Professional, quality-focused
- **Economy**:
  - Starting Gold: 800g
  - Max Gold: 1500g
  - Regen: 150g/day
  - Buy: 0.6× | Sell: 1.2×
- **Specialty**: Premium quality goods

#### 3. Verdant (City-South) - Thalindra Greenleaf
- **Location**: Garden Plaza
- **Race**: Elf
- **Personality**: Graceful, nature-focused
- **Economy**:
  - Starting Gold: 600g
  - Max Gold: 1200g
  - Regen: 120g/day
  - Buy: 0.5× | Sell: 1.0×
- **Specialty**: Nature items (herbs, arrows, traps)

#### 4. Ironpeak (Town-NE) - Brogni Ironhand
- **Location**: Miners' Plaza
- **Race**: Dwarf
- **Personality**: Gruff, practical
- **Economy**:
  - Starting Gold: 550g
  - Max Gold: 1100g
  - Regen: 110g/day
  - Buy: 0.5× | Sell: 1.0×
- **Specialty**: Durable mining gear

#### 5. Millhaven (Village-West) - Pip Haversham
- **Location**: Village Center
- **Race**: Halfling
- **Personality**: Cheerful, motherly
- **Economy**:
  - Starting Gold: 300g (lowest)
  - Max Gold: 600g
  - Regen: 75g/day
  - Buy: 0.4× | Sell: 1.0×
- **Specialty**: Basic village supplies

---

## Shop Pricing

### Pricing Multipliers

| Location Type | Buy from Player | Sell to Player | Reasoning |
|---------------|----------------|----------------|-----------|
| **Villages** | 0.4× (40%) | 1.0× (100%) | Limited gold, rural economy |
| **Towns** | 0.5× (50%) | 1.0× (100%) | Standard rates |
| **Cities** | 0.5-0.6× | 1.0-1.2× | Better rates, premium goods |
| **Capital** | 0.5× | 1.0× | Balanced fair market |

### Special Cases

**Stolen Goods**: 0.2× (20% of value) - High risk for merchant
**Damaged Items**: Base price × condition% × multiplier
**Bulk Sales**: Potential discounts for selling many items at once
**Reputation**: Future system may improve prices based on player reputation

---

## Shop Economy

### Gold Regeneration

Merchants regenerate gold daily based on:
- Location tier (village < town < city)
- Merchant wealth
- Trade volume

**Formula**: `daily_regen = base_regen + (sales_yesterday × 0.1)`

### Stock Regeneration

Items restock based on:
- **Common items**: Full restock daily
- **Uncommon items**: 50% restock daily
- **Rare items**: 1-2 items per week
- **Very Rare+**: Special events only

### Economy Balance

**Player Exploit Prevention**:
- Limited merchant gold prevents infinite money loops
- Stock limits prevent buying entire inventory
- Buy/sell price spread prevents arbitrage
- Regeneration timers slow farming

**Merchant Bankruptcy**: If merchant gold < 0, they refuse transactions until next regen

---

## Specialized Shops

### Future Implementation

**Blacksmith**:
- Weapon/armor repairs
- Equipment upgrades
- Metal materials

**Alchemist**:
- Potions and elixirs
- Potion crafting
- Rare ingredients

**Magic Shop**:
- Spell scrolls
- Wands and staves
- Enchanted items (very rare)

**Bowyer/Fletcher**:
- Bows and crossbows
- Arrows and bolts
- Ranged weapon upgrades

---

## Tavern/Inn System

### Core Services

#### 1. Resting
**Purpose**: Recover HP, mana, hunger, fatigue

**Room Tiers**:
- **Poor Room** (2-5g): Basic recovery, dirty, noisy
  - HP: +50%, Mana: +50%, Hunger: +20%, Fatigue: -40%
- **Common Room** (10-15g): Standard recovery
  - HP: +100% (full), Mana: +100% (full), Hunger: +50%, Fatigue: -80%
- **Fine Room** (25-50g): Premium recovery + bonuses
  - HP: +100%, Mana: +100%, Hunger: +80%, Fatigue: -100%, +temp buff

**Meal Options** (with room or separate):
- **Meal included**: Standard fare
- **Hearty feast** (+5-10g): Better hunger/fatigue recovery
- **Skip meal**: Save money, less hunger recovery

#### 2. Tavern Services (Future)

**Drinking/Social**:
- Buy drinks (minor buffs, flavor)
- Talk to NPCs (quests, rumors)
- Gather information about locations

**Entertainment**:
- Bard performances (if player is bard)
- Games of chance (gambling mini-game)
- Arm wrestling, darts, etc.

**Quest Hub**:
- Notice board for side quests
- NPC questgivers
- Bounties and contracts

### Tavern NPCs by Location

See `tavern-npcs-todo.md` and `tavern-performance-design.md` for detailed NPC personalities and performance mechanics (future features).

---

## Implementation Plan

### Phase 1: General Store (Priority 1)
- [ ] Shop UI/interface
- [ ] Buy transaction logic
- [ ] Sell transaction logic
- [ ] Merchant gold system
- [ ] Stock system
- [ ] Price calculations
- [ ] Per-location shop inventories

### Phase 2: Inn/Rest System (Priority 2)
- [ ] Rest UI/interface
- [ ] Room tier selection
- [ ] Recovery calculations
- [ ] Payment system
- [ ] Meal options
- [ ] Time passage mechanic

### Phase 3: Economy Polish (Priority 3)
- [ ] Gold regeneration
- [ ] Stock regeneration
- [ ] Save/load merchant states
- [ ] Balance testing
- [ ] Exploit prevention

### Phase 4: Future Features (Priority 4+)
- [ ] Specialized shops
- [ ] Tavern social features
- [ ] Performance system (bards)
- [ ] Quest board
- [ ] Reputation system

---

## Data Structure

### Shop Save Data
```json
{
  "shops": {
    "kingdom-general-store": {
      "merchant_id": "aldric-goodsworth",
      "current_gold": 750,
      "last_regen": "2025-11-03T00:00:00Z",
      "stock": {
        "rope-hempen-50-feet": 5,
        "torch": 20,
        "health-potion": 3
      }
    }
  }
}
```

### Rest Transaction
```json
{
  "action": "rest",
  "location": "kingdom-inn",
  "room_tier": "common",
  "cost": 12,
  "recovery": {
    "hp": "full",
    "mana": "full",
    "hunger": 50,
    "fatigue": -80
  }
}
```

---

## References

**Source Documents** (consolidated from):
- `shop-system.md` - Core shop mechanics
- `shop-pricing-system.md` - Pricing formulas
- `specialized-shops-design.md` - Future shop types
- `tavern-performance-design.md` - Bard performance system
- `tavern-npcs-todo.md` - Tavern NPC designs

**Related Systems**:
- `locations-plan.md` - Location requirements
- `house-of-keeping-system.md` - Vault/storage system
- Inventory system (implemented)
- Hunger/fatigue system (implemented)

---

**Status**: Ready for implementation
**Next Step**: Implement Phase 1 (General Store)
