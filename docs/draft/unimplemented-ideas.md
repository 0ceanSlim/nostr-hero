# Unimplemented Ideas Archive

Original ideas and features not yet implemented in the game.

**Purpose**: Preserve original design concepts for future consideration
**Source**: Extracted from `planning.txt` and early design documents
**Last Updated**: 2025-11-03

---

## Original Character Creation Concept

### Orphan Origin Story
*From planning.txt*

**Concept**:
- Player is an orphan who reaches age 18
- Someone from the city helped raise you but has passed away
- Their death prompts your adventure
- They left you items but you can't take everything
- Choose what to bring based on class/background

**Current Status**: ❌ Not implemented
- Current system: Direct character generation from npub
- No orphan backstory
- No item selection choice at start

**Notes**: Could be interesting narrative hook but conflicts with deterministic character gen. Maybe add as flavor text?

---

## Skills System (Cosmetic)

### Ability-Based Skill Checks
*From planning.txt*

**Concept**:
- Skills available if ability score ≥ 12
- Skills are "cosmetic" in demo (no mechanical effect)
- Show skill icons: ✨💪🏻💡🤸🏻💫🩸

**Current Status**: ⚠️ Partially considered
- No skill system implemented
- Could add for flavor/roleplay

**Notes**: Low priority, but could add depth to character sheets

---

## Carry Weight System

### Strength-Based Encumbrance
*From planning.txt*

**Formula**: `carry_weight = 15 × STR`

**Current Status**: ⚠️ System exists in code
- See: `docs/data/systems/encumbrance-system.json`
- May not be actively enforced in gameplay
- Check if implemented

**Notes**: Verify if this is actually working in the current build

---

## Background Integration

### Starting Gold from Background
*From planning.txt*

**Concept**:
> "From your time as a <background> you have <gp saved>"

**Current Status**: ❌ Not implemented
- Backgrounds exist
- Starting gold is fixed per class, not background-influenced

**Notes**: Could add variety to starting resources

---

## Level Up System

### HP Gain Formula
*From planning.txt*

**Rules**:
- Level 1: Hit die max + CON modifier
- Level 2+: Previous HP + (half max hit die) + CON modifier

**Example**:
- Ranger with CON 12 (+1)
- Level 1: 10 + 1 = 11 HP
- Level 2: 11 + 5 + 1 = 17 HP

**Current Status**: ⚠️ Unknown
- Character progression not yet implemented
- Level system exists in data structure

**Notes**: Need to verify formula when implementing leveling

---

## Magic Items

### Post-Launch Feature
*From planning.txt*

**Concept**:
> "magic items to be added one at a time after base game is done"

**Source**: https://www.dndbeyond.com/magic-items

**Current Status**: ❌ Not implemented
- Only mundane items
- No magic item system

**Notes**:
- Planned for post-1.0
- Would need enchantment system
- Attunement mechanics
- Rarity system already exists (can extend)

---

## Feats System

### Character Customization
*From planning.txt*

**Source**: https://www.dndbeyond.com/feats (D&D core rules)

**Current Status**: ❌ Not implemented
- No feat system
- No ASI (Ability Score Improvement) system

**Notes**: Would require level-up system first

---

## Monster Lairs

### Location-Specific Bosses
*From planning.txt*

**Concept**:
> "add special monsters with lairs at specific locations later"

**Current Status**: ❌ Not implemented
- Monsters exist (303 JSON files)
- No lair mechanic
- No location-specific boss encounters

**Notes**:
- Could add endgame content
- Unique loot tables
- Story significance

---

## CR Limitations (Demo)

### Early Access Scope
*From planning.txt*

**Original Plan**:
> "demo to only include monsters up to CR 5"

**Current Status**: ⚠️ Check implementation
- All 303 monsters imported from D&D
- May include monsters above CR 5
- Need to verify if CR restrictions are enforced

---

## Equipment Packs

### Starting Gear Bundles
*From planning.txt*

**Note**:
> "packs still needs work"

**Current Status**: ✅ Implemented
- See: `docs/data/equipment/packs.json`
- Used in character generation

**Notes**: ✅ This was completed

---

## Future Scraped Content

### D&D Beyond Integration
*From planning.txt*

**Planned Sources**:
- ✅ Equipment - Completed
- ✅ Spells - Completed (84 spells)
- ✅ Monsters - Completed (303 monsters)
- ❌ Feats - Not implemented
- ❌ Backgrounds - Partial (exist but limited integration)
- ❌ Magic Items - Not implemented

**Notes**: May not need to scrape more - game diverging from strict D&D 5e

---

## Ideas to Revisit

### High Priority
1. **Carry weight enforcement** - System exists, just verify it works
2. **Level up mechanics** - Needed for game progression
3. **Monster lairs** - Good endgame content

### Medium Priority
4. **Background starting gold variance** - Easy flavor addition
5. **Orphan narrative** - Could add as flavor text
6. **Skills (cosmetic)** - Character sheet flavor

### Low Priority / Post-1.0
7. **Magic items** - Major system, save for later
8. **Feats** - Requires level system
9. **CR restrictions** - May not matter if combat is different

---

## Design Philosophy Notes

From reviewing planning.txt:
- Originally planned as faithful D&D 5e adaptation
- Game has evolved to have unique mechanics (hunger, fatigue, Nostr auth)
- Some D&D systems may not fit current vision
- Focus on what makes this game unique vs. being "D&D but web-based"

**Question for Future**:
- Keep D&D combat/rules strictly?
- Or adapt/simplify for web-based single-player experience?

---

**Preservation Note**:
This document preserves original ideas that may be revisited. Not all ideas need to be implemented - some may not fit the evolved vision of the game.
