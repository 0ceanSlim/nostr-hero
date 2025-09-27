# Dynamic Character Generation System - Implementation Plan

## Problem Statement
The current character generation system has critical ID inconsistencies that prevent reliable dynamic character creation:

- **Item files**: Use filename as ID (`component-pouch.json` → ID: `component-pouch`)
- **Database**: Uses filename as ID (`component-pouch` in database)
- **Starting gear**: Uses display names (`"crossbow, light"`, `"component-pouch"`)
- **Result**: Lookups fail because `"crossbow, light"` ≠ `crossbow-light` ≠ `"Crossbow light"`

This prevents scalable character generation and requires manual mapping for every possible choice.

## Current Architecture Issues

### Data Flow Problem
```
starting-gear.json → Frontend → Database lookup by NAME → ❌ FAILS
```

### Specific Examples of Failures
- Starting gear: `"component-pouch"` → Database: `"Component Pouch"` (hyphen vs space)
- Starting gear: `"crossbow, light"` → Database: `"Crossbow light"` (comma vs no comma)
- Result: Items return `null`, treated as non-containers, placed incorrectly

## Solution: Universal ID-Based System

### Phase 1: Establish ID Standards ⭐ PRIORITY

#### 1. Item ID Format
- Use kebab-case: `crossbow-light`, `component-pouch`, `explorers-pack`
- IDs derived from filenames (already working in database)
- Consistent across all systems

#### 2. Add Explicit ID Fields to ALL Item JSONs
```json
{
  "id": "crossbow-light",           // ← NEW: Explicit ID (required)
  "name": "Crossbow, Light",        // ← Display name (can vary)
  "description": "...",
  "gear_slot": "hands",
  "tags": ["equipment", "ranged"],
  "slots": "1",
  ...
}
```

**Files to update**: All files in `docs/data/equipment/items/*.json` (~191 files)

### Phase 2: Convert Starting Gear System

#### 3. Update starting-gear.json to Use IDs
```json
{
  "class": "Druid",
  "starting_gear": [
    {
      "given": [
        ["leather", 1],                // ← Use IDs, not display names
        ["crossbow-light", 1],         // ← ID matches filename
        ["component-pouch", 1],        // ← ID matches filename
        ["explorers-pack", 1]          // ← ID matches filename
      ]
    },
    {
      "option": [
        ["scimitar", 1],               // ← All options use IDs
        ["shortsword", 1]
      ]
    }
  ]
}
```

**Files to update**: `docs/data/character/starting-gear.json`

### Phase 3: Update Frontend System

#### 4. Replace Name-Based Lookup with ID-Based
```javascript
// OLD: getItemData(itemName) - fragile name matching
// NEW: getItemById(itemId) - reliable ID lookup

async function getItemById(itemId) {
  const items = await loadItemsFromDatabase();
  return items.find(item => item.id === itemId);
}
```

#### 5. Update Inventory Creation Logic
```javascript
// Process starting gear items by ID
allItems.forEach(async item => {
  const [itemId, quantity] = item; // ["crossbow-light", 1]
  const itemData = await getItemById(itemId);
  // Now itemData is guaranteed to work with proper tags/properties!
});
```

**Files to update**: `www/res/js/new-game.js`

### Phase 4: Database Optimization

#### 6. Ensure Database Uses IDs Consistently
- Primary key: `id` (already working)
- Lookups by ID only (fast, reliable)
- Display names in `name` field

**Current migration in `src/db/migration.go` already extracts filename as ID - NO CHANGES NEEDED**

### Phase 5: Character Generation Pipeline

#### 7. Complete Flow
```
1. Backend generates character class/background
2. Lookup starting gear by class → Gets item IDs
3. Frontend receives item IDs
4. Frontend looks up each ID in database → Gets item data
5. Frontend processes items by tags/properties (container, equipment, etc.)
6. Frontend creates proper inventory structure
7. ✅ ANY character choice works dynamically
```

## Implementation Priority

### IMMEDIATE (Phase 1 & 2) - THIS SESSION
1. Add `id` fields to all item JSONs (batch operation)
2. Convert starting-gear.json to use IDs
3. Update frontend to use ID-based lookups
4. Test basic character generation

### NEXT (Phase 3) - NEXT SESSION
1. Test with all character classes
2. Verify container logic works with IDs
3. Test equipment placement system
4. Handle edge cases and validation

### FUTURE (Phase 4 & 5) - OPTIMIZATION
1. Performance optimization
2. Add validation for missing items
3. Create item ID registry/validation
4. Full cinematic character generation

## Current Status

### Completed ✅
- Analysis of ID inconsistency problems
- Identification of all affected systems
- Architecture design for ID-based system

### In Progress 🔄
- Creating comprehensive implementation plan
- Documentation for future sessions

### Next Steps 📋
1. **Add ID fields to item JSONs** - Start with critical items (component-pouch, crossbow-light, etc.)
2. **Update starting-gear.json** - Convert display names to IDs
3. **Update frontend lookup** - Replace getItemData with getItemById
4. **Test character generation** - Verify containers work properly

## Key Files Affected

### Data Files
- `docs/data/equipment/items/*.json` (~191 files) - Add `id` fields
- `docs/data/character/starting-gear.json` - Convert to ID references

### Frontend Files
- `www/res/js/new-game.js` - Update item lookup system
- `www/res/js/character-generator.js` - May need updates for equipment processing

### Backend Files
- `src/db/migration.go` - No changes needed (already uses filename as ID)
- `src/api/gamedata.go` - No changes needed (already serves by ID)

## Success Criteria

### Phase 1 Success
- [ ] All item JSONs have explicit `id` fields
- [ ] Starting gear uses IDs instead of display names
- [ ] Frontend looks up items by ID
- [ ] Component pouch correctly identified as container with 4 slots
- [ ] Explorer's pack correctly unpacked
- [ ] Leather armor correctly equipped

### Final Success
- [ ] Any character class/background combination works
- [ ] Any equipment choice combination works
- [ ] Containers never placed inside other containers
- [ ] Equipment automatically placed in correct gear slots
- [ ] No manual name mapping required
- [ ] System ready for dynamic cinematic generation

## Risk Mitigation

### Backup Strategy
- Test changes on single character class first
- Keep original files as backups
- Implement ID validation to catch missing items

### Rollback Plan
- Revert to name-based lookups with manual mapping
- Use git to restore original starting-gear.json
- Database migration remains unchanged (safe)

---

**Goal**: Transform the current fragile name-based system into a robust ID-based foundation that supports unlimited dynamic character generation scenarios for the cinematic experience.