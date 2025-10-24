# Shop Pricing System - Charisma-Based

## Core Pricing Formula

### Base Item Value
Every item has a defined `base_value` in gold. This is the "true market value" of the item.

---

## GENERAL STORES

General stores buy ANY item and sell common goods.

### Selling Items (Player sells TO general store)
- **At CHA 20**: Player receives 100% of base value (1.0×)
- **At CHA 10**: Player receives 50% of base value (0.5×)
- **Formula**: `sell_price = base_value × (0.5 + (CHA - 10) × 0.05)`
- **Cap**: Maximum sell price is 100% of base value (even with CHA > 20)

### Buying Items (Player buys FROM general store)
- **At CHA 20**: Player pays 100% of base value (1.0×)
- **At CHA 10**: Player pays 150% of base value (1.5×)
- **Formula**: `buy_price = max(base_value × (1.5 - (CHA - 10) × 0.05), base_value)`
- **Cap**: Minimum buy price is 100% of base value (even with CHA > 20)
- **Note**: Charisma caps at 20 for characters

---

## SPECIALIZED SHOPS

Specialized shops ONLY buy/sell items in their specialty category.

### Selling Items (Player sells TO specialized shop)
- Specialized shops pay **10% MORE** than general stores for items in their specialty
- **At CHA 20**: Player receives 100% of base value (1.0× after cap)
- **At CHA 18.2**: Player receives ~100% of base value (0.91 × 1.1 = 1.0)
- **At CHA 10**: Player receives 55% of base value (0.55×)
- **Formula**: `sell_price = min(base_value × (0.5 + (CHA - 10) × 0.05) × 1.1, base_value)`
- **Cap**: Maximum sell price is 100% of base value (same as general stores)
- **Restriction**: Shop will ONLY buy items from their specialty inventory

### Buying Items (Player buys FROM specialized shop)
- Same as general stores (standard CHA pricing)
- **At CHA 20**: Player pays 100% of base value (1.0×)
- **At CHA 10**: Player pays 150% of base value (1.5×)
- **Formula**: `buy_price = max(base_value × (1.5 - (CHA - 10) × 0.05), base_value)`
- **Cap**: Minimum buy price is 100% of base value (never below)
- **Restriction**: Shop will ONLY sell items from their specialty inventory

## Charisma Scaling Table

| CHA | General: Sell | General: Buy | Specialist: Sell | Specialist: Buy | Example (10g item) |
|-----|--------------|--------------|-----------------|----------------|-------------------|
| 6   | 0.30× (30%)  | 1.70×        | 0.33× (33%)     | 1.70×          | Gen sell: 3g / Spec sell: 3.3g / Buy: 17g |
| 8   | 0.40× (40%)  | 1.60×        | 0.44× (44%)     | 1.60×          | Gen sell: 4g / Spec sell: 4.4g / Buy: 16g |
| 10  | 0.50× (50%)  | 1.50×        | 0.55× (55%)     | 1.50×          | Gen sell: 5g / Spec sell: 5.5g / Buy: 15g |
| 12  | 0.60× (60%)  | 1.40×        | 0.66× (66%)     | 1.40×          | Gen sell: 6g / Spec sell: 6.6g / Buy: 14g |
| 14  | 0.70× (70%)  | 1.30×        | 0.77× (77%)     | 1.30×          | Gen sell: 7g / Spec sell: 7.7g / Buy: 13g |
| 16  | 0.80× (80%)  | 1.20×        | 0.88× (88%)     | 1.20×          | Gen sell: 8g / Spec sell: 8.8g / Buy: 12g |
| 18  | 0.90× (90%)  | 1.10×        | 0.99× (99%)     | 1.10×          | Gen sell: 9g / Spec sell: 9.9g / Buy: 11g |
| 19  | 0.95× (95%)  | 1.05×        | 1.00× (CAP)     | 1.05×          | Gen sell: 9.5g / Spec sell: 10g / Buy: 10.5g |
| 20  | 1.00× (CAP)  | 1.00× (CAP)  | 1.00× (CAP)     | 1.00× (CAP)    | Gen sell: 10g / Spec sell: 10g / Buy: 10g |

**Note**: CHA 20 is the maximum for player characters

## Implementation Notes

### Shop Types
1. **General Stores**: Buy ANY item, sell common goods, standard CHA pricing
2. **Specialized Shops**: ONLY buy/sell specialty items, pay 10% MORE for buying player items

### Why Use Specialized Shops?
- **Selling specialty items**: Get 10% more gold than general stores (up to 100% cap)
- **Example**: Longsword (15g base value)
  - General store at CHA 18: 13.5g (90%)
  - Weapon shop at CHA 18: 14.85g (99%, +1.35g bonus!)
  - General store at CHA 19: 14.25g (95%)
  - Weapon shop at CHA 19: 15g (100% capped, +0.75g bonus!)
  - General store at CHA 20: 15g (100% cap)
  - Weapon shop at CHA 20: 15g (100% cap, same price)
  - General store at CHA 10: 7.5g (50%)
  - Weapon shop at CHA 10: 8.25g (55%, +0.75g bonus!)

**Key Insight**: Specialists help you reach the 100% cap at lower CHA (around CHA 19 instead of CHA 20)

### Shop Constraints
Shops still have limitations:
1. **Merchant Gold**: Shops have limited gold to buy items from players
2. **Stock Limits**: Shops have limited quantities of items to sell
3. **Daily Regeneration**: Gold and stock regenerate daily
4. **Specialty Restrictions**: Specialized shops ONLY deal in their specialty items

### Shop Differentiation
- **Merchant Gold Pools**: Vary by location wealth (300g-900g)
- **Stock Limits**: Each item has quantity available
- **Specialty Categories**: What items they buy/sell
- **Specialist Bonus**: 10% extra when selling to specialists

## Examples

### Example 1: Selling a Longsword (base_value = 15g)
**To General Store:**
- Player with CHA 10: Receives 7.5g (50%)
- Player with CHA 15: Receives 11.25g (75%)
- Player with CHA 20: Receives 15g (100% cap)

**To Specialized Weapon Shop:**
- Player with CHA 10: Receives 8.25g (55%, +0.75g bonus)
- Player with CHA 15: Receives 12.375g (82.5%, +1.125g bonus)
- Player with CHA 19: Receives 15g (100% cap, +0.75g bonus)
- Player with CHA 20: Receives 15g (100% cap, same as general store)

### Example 2: Buying Leather Armor (base_value = 10g)
**From Any Shop:**
- Player with CHA 10: Pays 15g (150%)
- Player with CHA 15: Pays 12.5g (125%)
- Player with CHA 20: Pays 10g (100% cap)

### Example 3: Low CHA Penalty
- Player with CHA 6 selling a 100g item to general store: Receives 30g (30%)
- Player with CHA 6 selling a 100g item to specialist: Receives 33g (33%, +3g)
- Player with CHA 6 buying a 100g item: Must pay 170g (170%)

### Example 4: Max CHA Performance
- Player with CHA 20 selling 100g item to general store: Receives 100g (100% cap)
- Player with CHA 20 selling 100g item to specialist: Receives 100g (100% cap)
- Player with CHA 20 buying 100g item: Pays 100g (100% cap)
- **Perfect negotiation at CHA 20!**

## Backend Implementation

```javascript
// Pseudo-code for pricing calculations

function calculateSellPrice(baseValue, charisma, isSpecialist = false) {
  const baseSellMultiplier = 0.5 + ((charisma - 10) * 0.05);
  let sellMultiplier = baseSellMultiplier;

  if (isSpecialist) {
    sellMultiplier = baseSellMultiplier * 1.1;
  }

  // Cap at 100% for ALL shops
  sellMultiplier = Math.min(sellMultiplier, 1.0);

  return Math.floor(baseValue * sellMultiplier);
}

function calculateBuyPrice(baseValue, charisma) {
  const buyMultiplier = 1.5 - ((charisma - 10) * 0.05);
  const calculatedPrice = baseValue * buyMultiplier;

  // Cap at base value (can never buy for less than 100%)
  return Math.ceil(Math.max(calculatedPrice, baseValue));
}

// Example API endpoint
POST /api/shop/:npc_id/sell
{
  "item_id": "longsword",
  "quantity": 1
}

Response:
{
  "success": true,
  "item_base_value": 15,
  "player_charisma": 12,
  "sell_price": 9,  // (15 × 0.6)
  "gold_received": 9,
  "merchant_gold_remaining": 491
}
```

## Database Schema Update

```sql
-- Items table should include base_value
CREATE TABLE items (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  base_value INTEGER NOT NULL,  -- Base market value in gold
  weight REAL,
  description TEXT
);

-- Shop transactions log pricing details
CREATE TABLE shop_transactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  player_id TEXT NOT NULL,
  merchant_npc_id TEXT NOT NULL,
  item_id TEXT NOT NULL,
  transaction_type TEXT NOT NULL,  -- 'buy' or 'sell'
  quantity INTEGER NOT NULL,
  item_base_value INTEGER NOT NULL,
  player_charisma INTEGER NOT NULL,
  price_multiplier REAL NOT NULL,
  total_gold INTEGER NOT NULL,
  timestamp INTEGER NOT NULL
);
```

## Updated Shop Design Philosophy

Since pricing is now uniform across all shops, the differentiation comes from:

1. **Inventory Selection**: What items they stock
2. **Merchant Gold**: How much they can buy from players
3. **Stock Quantities**: How many of each item available
4. **Buy Acceptance**: Whether they buy items outside their specialty
5. **Thematic Flavor**: NPC personality and shop description

Specialized shops are now about **what** they sell, not **how much** they charge.

---

## Revised Specialized Shops

All shops now have:
- **No price multipliers** (removed)
- **Merchant Gold**: Based on location wealth (300g-900g)
- **Specialty Inventory**: What they focus on selling
- **Buy Policy**: General stores buy anything, specialists may be selective

See `specialized-shops-design.md` for shop list, but ignore all pricing multipliers mentioned there.
