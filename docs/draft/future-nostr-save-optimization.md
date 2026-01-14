# Future: Nostr Save File Optimization

**Status:** Future consideration (not part of delta architecture rewrite)
**Priority:** Low (current saves work, optimize before final release)

---

## Overview

This document captures ideas for optimizing save files for Nostr relay storage. These optimizations are **separate from the delta architecture** and should be implemented later when preparing for production Nostr integration.

## Current Save Format Issues

The current save format stores full item/spell data inline:

```json
{
  "inventory": {
    "general_slots": [
      {
        "item": {
          "id": "longsword",
          "name": "Longsword",
          "description": "A versatile martial weapon...",
          "damage": "1d8",
          "damage_type": "slashing",
          "weight": 3,
          "value": 15,
          "rarity": "common",
          "properties": ["versatile"],
          "image": "longsword.png"
        },
        "quantity": 1
      }
    ]
  }
}
```

**Problem:** ~500 bytes per item x 20 slots = **10KB just for inventory**

---

## Proposed Optimized Format

Store only IDs and quantities, hydrate from game-data on load:

```json
{
  "hp": 15,
  "f": 3,
  "h": 2,
  "g": 250,
  "xp": 100,
  "loc": "kingdom",
  "dist": "center",
  "bld": "",
  "inv": [
    { "id": "longsword", "q": 1 },
    { "id": "health-potion", "q": 3 }
  ],
  "eq": {
    "mainHand": "longsword",
    "armor": "leather-armor"
  },
  "sp": ["fire-bolt", "magic-missile"],
  "eff": [{ "id": "fatigue-accumulation", "d": -1 }],
  "t": 720,
  "d": 1
}
```

**Result:** ~500 bytes total for entire save

### Field Name Shortening

| Full Name | Short | Type |
|-----------|-------|------|
| fatigue | f | int |
| hunger | h | int |
| gold | g | int |
| location | loc | string |
| district | dist | string |
| building | bld | string |
| inventory | inv | array |
| equipment | eq | map |
| spells | sp | array |
| effects | eff | array |
| time | t | int |
| day | d | int |

---

## Go Types for Nostr Storage

```go
// Optimized for Nostr storage
type NostrSaveFile struct {
    HP         int                    `json:"hp"`
    Fatigue    int                    `json:"f"`
    Hunger     int                    `json:"h"`
    Gold       int                    `json:"g"`
    XP         int                    `json:"xp"`

    Location   string                 `json:"loc"`
    District   string                 `json:"dist"`
    Building   string                 `json:"bld"`

    Items      []InventorySlotMinimal `json:"inv"`
    Equipment  map[string]string      `json:"eq"`
    Spells     []string               `json:"sp"`
    Effects    []ActiveEffectMinimal  `json:"eff"`

    Time       int                    `json:"t"`
    Day        int                    `json:"d"`
}

type InventorySlotMinimal struct {
    ID  string `json:"id"`
    Q   int    `json:"q"`
}

type ActiveEffectMinimal struct {
    ID  string `json:"id"`
    D   int    `json:"d"`    // duration (-1 = permanent)
}
```

---

## Estimated Sizes

| Format | Size |
|--------|------|
| Current (full data) | 20-50KB |
| Optimized (IDs only) | 5-10KB |
| Compressed (gzip) | 2-5KB |
| Encrypted (NIP-17) | 3-6KB |
| **Nostr limit** | **64KB** |

All formats well within Nostr 64KB event limit.

---

## Hydration Flow

When loading a save from Nostr:

```
Fetch minimal save from Nostr relay (~5KB)
  ↓
Parse JSON into NostrSaveFile struct
  ↓
Hydrate full data:
  1. Look up item IDs → full item data from game-data/items/
  2. Look up spell IDs → full spell data from game-data/magic/spells/
  3. Look up location → full location data from game-data/locations/
  ↓
Create rich SessionState in memory (~100KB+)
  ↓
Game runs using SessionState (not NostrSaveFile)
```

---

## Quick Wins (Before Full Optimization)

These can be done without restructuring:

1. **Remove empty inventory slots** - Don't serialize null/empty slots
2. **Compress JSON** - gzip before publishing to relay
3. **Remove redundant data** - Don't store computed values

---

## Implementation Notes

- This optimization is **independent** of the delta architecture
- Current saves work fine for local development
- Prioritize this before production Nostr deployment
- Consider NIP-17 encryption for privacy

---

**Last Updated:** 2026-01-14
**Status:** Draft / Future Work
