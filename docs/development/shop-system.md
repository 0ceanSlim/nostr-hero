# Shop System Documentation

## Overview

The shop system allows players to buy items from merchants and sell items from their inventory. Each shop has limited gold and stock that regenerates over time, creating a dynamic economy.

## Core Mechanics

### Shop Gold System
- **Starting Gold**: Merchants start with a gold reserve (varies by location/merchant type)
- **Max Gold**: Maximum gold a merchant can accumulate
- **Gold Depletion**: Merchants lose gold when buying items from players
- **Gold Regeneration**: Merchants gain gold over time (daily restock)
- **Out of Gold**: When a merchant runs out of gold, they cannot buy items from players

### Inventory Stock System
- **Stock Limits**: Each item has a maximum stock quantity
- **Stock Depletion**: Items are removed from stock when purchased
- **Restocking**: Items regenerate daily (or on configured interval)
- **Out of Stock**: Items cannot be purchased when stock reaches zero

### Pricing
- **Sell Price**: Base item price × sell_price_multiplier
- **Buy Price**: Base item price × buy_price_multiplier
- **Player Sells to Merchant**: Player receives buy_price (typically 40-60% of base)
- **Player Buys from Merchant**: Player pays sell_price (typically 90-120% of base)

## General Store NPCs

### 1. Kingdom - Aldric Goodsworth (general-merchant)
**Location**: Kingdom Center, Grand Market
**Race**: Human
**Personality**: Cheerful, welcoming, business-minded

**Economy**:
- Starting Gold: 500g
- Max Gold: 1000g
- Regen: 100g/day
- Buy Multiplier: 0.5× (pays 50% of item value)
- Sell Multiplier: 1.0× (sells at base price)

**Specialty**: Balanced general supplies for urban adventurers

---

### 2. City-East (Goldenhaven) - Aurelia Goldscale (trade-merchant)
**Location**: Golden Square, Goldscale Trading House
**Race**: Dragonborn
**Personality**: Professional, quality-focused, slightly expensive

**Economy**:
- Starting Gold: 800g
- Max Gold: 1500g
- Regen: 150g/day
- Buy Multiplier: 0.6× (pays 60% - better than most)
- Sell Multiplier: 1.2× (sells 20% above base - premium pricing)

**Specialty**: High-quality goods, silk rope, spyglass, premium items

---

### 3. City-South (Verdant) - Thalindra Greenleaf (nature-supplier)
**Location**: Garden Plaza, Greenleaf Provisions
**Race**: Elf
**Personality**: Graceful, nature-focused, respectful

**Economy**:
- Starting Gold: 600g
- Max Gold: 1200g
- Regen: 120g/day
- Buy Multiplier: 0.5×
- Sell Multiplier: 1.0×

**Specialty**: Nature items (antitoxin, herbalism kit, hunting traps, nets, arrows)

---

### 4. Town-Northeast (Ironpeak) - Brogni Ironhand (mining-supplier)
**Location**: Miners' Plaza, Ironhand's Supplies
**Race**: Dwarf
**Personality**: Gruff, practical, dwarven quality

**Economy**:
- Starting Gold: 550g
- Max Gold: 1100g
- Regen: 110g/day
- Buy Multiplier: 0.5×
- Sell Multiplier: 1.0×

**Specialty**: Durable gear (extra torches, pitons, lanterns, mining tools, chains)

---

### 5. Village-West (Millhaven) - Pip Haversham (village-shopkeeper)
**Location**: Village Center, Haversham's General Store
**Race**: Halfling
**Personality**: Cheerful, motherly, small-village friendly

**Economy**:
- Starting Gold: 300g (lowest)
- Max Gold: 700g
- Regen: 80g/day
- Buy Multiplier: 0.4× (pays only 40% - worst prices)
- Sell Multiplier: 0.9× (sells 10% below base - cheapest)

**Specialty**: Rural goods (extra rations, candles, soap, buckets, fishing tackle, sacks)

---

### 6. Village-Southwest (Marshlight) - Grokmar the Shrewd (swamp-trader)
**Location**: Stilt Houses, Swamp Treasures
**Race**: Orc
**Personality**: Lean, calculating, survival-focused

**Economy**:
- Starting Gold: 400g
- Max Gold: 900g
- Regen: 90g/day
- Buy Multiplier: 0.45×
- Sell Multiplier: 0.95×

**Specialty**: Swamp survival (antitoxin, nets, poles, waterproof bags, insect repellent)

---

## NPC JSON Schema

```json
{
  "id": "npc-id",
  "name": "Full Name",
  "title": "Title",
  "race": "Race",
  "location": "location-id",
  "building": "building_id",
  "description": "Visual and personality description",
  "greeting": {
    "first_time": "First meeting",
    "returning": "Subsequent visits",
    "low_gold": "When merchant is low on gold"
  },
  "dialogue": {
    "main_menu": {
      "text": "What brings you here?",
      "options": ["browse_shop", "sell_items", "goodbye"]
    },
    "browse_shop": {
      "text": "Have a look!",
      "action": "open_shop",
      "options": ["main_menu"]
    },
    "sell_items": {
      "text": "What are you selling?",
      "action": "open_sell",
      "requirements": {
        "shop_gold": 1
      },
      "failure": "I'm out of coin.",
      "options": ["main_menu"]
    },
    "goodbye": {
      "text": "Farewell!",
      "action": "end_dialogue"
    }
  },
  "shop_config": {
    "shop_type": "general",
    "buys_items": true,
    "buy_price_multiplier": 0.5,
    "sell_price_multiplier": 1.0,
    "starting_gold": 500,
    "max_gold": 1000,
    "gold_regen_rate": 100,
    "gold_regen_interval": "daily",
    "inventory": [
      {
        "item_id": "rope-hemp",
        "stock": 10,
        "max_stock": 10,
        "restock_rate": 5,
        "restock_interval": "daily"
      }
    ]
  }
}
```

## Common Inventory Items

All general stores stock basic adventuring gear:

**Universal Items** (all shops):
- Rope (hemp or silk)
- Torches
- Rations
- Waterskin
- Backpack
- Bedroll
- Tinderbox
- Oil Flask

**Regional Variations**:
- **Urban** (Kingdom, Goldenhaven): Premium items, lanterns, spyglasses
- **Nature** (Verdant): Antitoxin, herbalism kits, hunting traps, arrows
- **Mountain** (Ironpeak): Extra torches/pitons, mining tools, chains
- **Rural** (Millhaven): Candles, soap, fishing tackle, farm goods
- **Swamp** (Marshlight): Waterproof bags, insect repellent, poles

## Backend Implementation

### Database Tables

#### shop_state
Tracks dynamic shop gold and stock:
```sql
CREATE TABLE shop_state (
  npc_id TEXT PRIMARY KEY,
  current_gold INTEGER,
  last_gold_regen TIMESTAMP,
  updated_at TIMESTAMP
);
```

#### shop_inventory
Tracks current stock per shop:
```sql
CREATE TABLE shop_inventory (
  npc_id TEXT,
  item_id TEXT,
  current_stock INTEGER,
  last_restock TIMESTAMP,
  PRIMARY KEY (npc_id, item_id)
);
```

#### player_transactions
Tracks buy/sell history:
```sql
CREATE TABLE player_transactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  npub TEXT,
  npc_id TEXT,
  item_id TEXT,
  quantity INTEGER,
  transaction_type TEXT, -- 'buy' or 'sell'
  gold_amount INTEGER,
  timestamp TIMESTAMP
);
```

### API Endpoints

#### GET /api/shop/:npc_id
Returns shop data:
```json
{
  "npc": { /* NPC data */ },
  "current_gold": 450,
  "max_gold": 1000,
  "inventory": [
    {
      "item_id": "rope-hemp",
      "item": { /* full item data */ },
      "current_stock": 7,
      "max_stock": 10,
      "sell_price": 100,
      "buy_price": 50
    }
  ]
}
```

#### POST /api/shop/:npc_id/buy
Player purchases item:
```json
{
  "item_id": "rope-hemp",
  "quantity": 2
}
```

Response:
```json
{
  "success": true,
  "gold_spent": 200,
  "items_received": 2,
  "new_stock": 5
}
```

#### POST /api/shop/:npc_id/sell
Player sells item:
```json
{
  "item_id": "longsword",
  "quantity": 1
}
```

Response:
```json
{
  "success": true,
  "gold_received": 75,
  "shop_gold_remaining": 375
}
```

### Restock Logic

**Daily Reset** (runs at server midnight or on first shop access per day):
```javascript
function restockShop(npcId, shopConfig) {
  // Regenerate gold
  const newGold = Math.min(
    currentGold + shopConfig.gold_regen_rate,
    shopConfig.max_gold
  );

  // Restock items
  for (const item of shopConfig.inventory) {
    const newStock = Math.min(
      currentStock + item.restock_rate,
      item.max_stock
    );
    updateShopInventory(npcId, item.item_id, newStock);
  }
}
```

### Save File Integration

Player's save file should track:
```json
{
  "shop_interactions": {
    "general-merchant": {
      "first_visit": "2025-10-16T10:30:00Z",
      "last_visit": "2025-10-20T14:22:00Z",
      "total_purchases": 15,
      "total_sales": 8
    }
  }
}
```

## Economy Balance

### Price Ranges
- **Best Buy Prices**: Goldenhaven (60%)
- **Standard Buy Prices**: Most shops (50%)
- **Worst Buy Prices**: Millhaven (40%)

- **Cheapest Sell Prices**: Millhaven (90%)
- **Standard Sell Prices**: Most shops (100%)
- **Most Expensive**: Goldenhaven (120%)

### Gold Availability
- **Richest**: Goldenhaven (800-1500g)
- **Standard**: Kingdom, Verdant, Ironpeak (500-1200g)
- **Poorest**: Millhaven, Marshlight (300-900g)

### Design Philosophy
- **Urban merchants** have more gold and charge premium prices
- **Rural merchants** have less gold but offer discounts
- **Specialized merchants** (Goldenhaven, Verdant) have unique items
- **Stock limits** prevent players from buying unlimited quantities
- **Gold limits** prevent players from selling entire inventory at once

## Future Enhancements
- Dynamic pricing based on supply/demand
- Reputation discounts
- Bulk purchase discounts
- Special orders (pay upfront, wait for delivery)
- Regional price differences (ore cheap in mining town, expensive elsewhere)
- Merchant quests for better prices
