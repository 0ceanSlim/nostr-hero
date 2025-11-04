# Excel Spreadsheet Archive Notes

This document explains the purpose and status of the `.xlsx` files in the draft folder.

---

## monsters.xlsx

**Purpose**: Original D&D 5e monster database used for planning

**Content**: Monster stat blocks scraped from D&D Beyond
- Monster names
- CR (Challenge Rating)
- HP, AC, abilities
- Special actions and traits
- Lore and descriptions

**Current Status**: ✅ Data migrated to JSON
- **Location**: `docs/data/content/monsters/`
- **Count**: 303 monster JSON files
- **Completeness**: All monsters from xlsx have been migrated

**Keep or Delete?**
- ✅ **Safe to delete** - All data preserved in JSON format
- JSON files are more flexible and easier to work with
- Migration is complete

**Recommendation**: Archive or delete xlsx file

---

## spells.xlsx

**Purpose**: D&D 5e spell database used for planning

**Content**: Spell data scraped from D&D Beyond
- Spell names
- Level (cantrip through 9th)
- School of magic
- Casting time, range, components
- Descriptions and effects
- Classes that can use them

**Current Status**: ⚠️ Partially migrated to JSON
- **Location**: `docs/data/content/spells/`
- **Count**: 84 spell JSON files
- **Completeness**: Unknown - xlsx may have more spells than JSON

**Keep or Delete?**
- ⚠️ **Keep for now** - May contain spells not yet migrated
- Need to verify all xlsx spells are in JSON format
- Once verified complete, can delete

**Recommendation**:
1. Compare xlsx spell count vs JSON count
2. Check if any spells in xlsx are missing from JSON
3. Migrate any missing spells
4. Then safe to delete

**Next Steps**:
```bash
# Count spells in xlsx (need to open file)
# Count spells in JSON
ls docs/data/content/spells/*.json | wc -l  # Returns 84

# If xlsx has more than 84, migrate the rest
```

---

## Summary

| File | Status | Can Delete? | Notes |
|------|--------|-------------|-------|
| `monsters.xlsx` | ✅ Complete | Yes | All 303 monsters in JSON |
| `spells.xlsx` | ⚠️ Maybe incomplete | Not yet | Verify first, may have spells not in JSON |

---

**Last Updated**: 2025-11-03
