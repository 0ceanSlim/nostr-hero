# Shop NPC JSON Audit Report

**STATUS: ✅ CLEANUP COMPLETED**

All shop NPC JSON files have been cleaned and validated.

## Issues Found

### 1. **~~Deprecated Fields (ALL NPCs)~~** ✅ FIXED

**Solution Applied**: Removed deprecated fields from all shop configs:
- `buy_price_multiplier` - Removed (pricing now driven by shop-pricing.json in database)
- `sell_price_multiplier` - Removed (pricing now driven by shop-pricing.json in database)
- `max_gold` - Removed (merchants can accumulate unlimited gold)

---

### 2. **~~Missing Top-Level Restock Interval (ALL NPCs)~~** ✅ FIXED

**Solution Applied**: Added separate restock intervals:
- `"item_restock_interval": 10` (minutes) - restocks items to max_stock
- `"gold_restock_interval": 30` (minutes) - restocks gold to starting_gold

These can be customized per shop as needed.

---

### 3. **~~Wrong Location IDs (4 NPCs)~~** ✅ RESOLVED

**UPDATE**: Location IDs were CORRECT. They represent the districts NPCs are in at certain times, not the city ID. No changes needed.

---

### 4. **~~Redundant Item-Level Restock Intervals (ALL NPCs)~~** ✅ FIXED

**Solution Applied**: Removed `restock_interval` and `restock_rate` from all individual items. Shop-level `item_restock_interval` now controls restocking.

---

### 5. **~~Missing Restock Rates (ALL NPCs)~~** ✅ FIXED

**Solution Applied**: Removed `restock_rate` from all items. Items now restock directly to `max_stock` when the interval triggers.

---

## Cleaned Shop Config Template

```json
"shop_config": {
  "shop_type": "general",
  "buys_items": true,
  "starting_gold": 500,
  "gold_regen_rate": 100,
  "gold_regen_interval": "daily",
  "gold_restock_interval": 30,
  "item_restock_interval": 10,
  "inventory": [
    {
      "item_id": "rope-hemp",
      "stock": 10,
      "max_stock": 10
    }
  ]
}
```

**Field Definitions**:
- `gold_regen_interval`: Gradual gold regeneration (e.g., "daily" = 10min game time)
- `gold_restock_interval`: Full gold restock to starting_gold (default: 30 minutes real time)
- `item_restock_interval`: Item restock to max_stock (default: 10 minutes real time)

---

## Summary by NPC

### ✅ general-merchant (Kingdom)
- **Location**: ✅ Correct (`kingdom`)
- **Issues**: ✅ Fixed - removed deprecated fields, added separate restock intervals

### ✅ trade-merchant (Goldenhaven)
- **Location**: ✅ Correct (`city-east` - district ID)
- **Issues**: ✅ Fixed - removed deprecated fields, added separate restock intervals

### ✅ mining-supplier (Ironpeak)
- **Location**: ✅ Correct (`town-northeast` - district ID)
- **Issues**: ✅ Fixed - removed deprecated fields, added separate restock intervals

### ✅ swamp-trader (Marshlight)
- **Location**: ✅ Correct (`village-southwest` - district ID)
- **Issues**: ✅ Fixed - removed deprecated fields, added separate restock intervals

### ✅ nature-supplier (Verdant)
- **Location**: ✅ Correct (`city-south` - district ID)
- **Issues**: ✅ Fixed - removed deprecated fields, added separate restock intervals

### ✅ village-shopkeeper (Millhaven)
- **Location**: ✅ Correct (`village-west` - district ID)
- **Issues**: ✅ Fixed - removed deprecated fields, added separate restock intervals

---

## ✅ Completed Actions

1. ✅ **Removed deprecated fields**: `buy_price_multiplier`, `sell_price_multiplier`, `max_gold`
2. ✅ **Verified location IDs**: District IDs are correct (represent NPC schedule locations)
3. ✅ **Added separate restock intervals**:
   - `item_restock_interval`: 10 minutes (default)
   - `gold_restock_interval`: 30 minutes (default)
4. ✅ **Removed item-level fields**: Removed `restock_interval` and `restock_rate` from all items
5. ✅ **Simplified item entries**: Now only contain `item_id`, `stock`, `max_stock`
