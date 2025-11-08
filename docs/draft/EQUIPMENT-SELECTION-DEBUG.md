# Equipment Selection System - Debug Documentation

**Status**: IN PROGRESS - Bundle (choice-0) not appearing in save file
**Date**: 2025-10-25
**Purpose**: Document equipment selection flow to aid debugging

## System Overview

The equipment selection system spans JavaScript (UI) and Go (logic):

```
Character Generator → Equipment Selection UI → Game Intro → Go Backend → Save File
(character-generator.js) (equipment-selection.js) (game-intro.js) (character-creation-helpers.go) (JSON)
```

### Architecture Principles (Go-First)

- **JavaScript**: Only UI rendering and user input collection
- **Go**: All game logic, validation, inventory building, auto-equipping
- **Data Flow**: UI → JavaScript collects choices → Send to Go API → Go processes → Save file

## Data Flow Diagram

```
1. CHARACTER GENERATION (character-generator.js)
   ↓
   Fetches character from /api/character?npub={npub}
   ↓
   Calls generateStartingEquipment(character)
   ↓
   Creates choice structure:
   {
     inventory: [...],          // Given items (auto-added)
     equipment: {...},
     choices: [                 // User must choose
       {
         id: "choice-0",
         options: [
           { isBundle: true, bundle: [["leather",1],["longbow",1],["arrows",20]] },
           { item: "chain-mail", quantity: 1 }
         ]
       },
       {
         id: "choice-1",
         options: [
           { isComplexChoice: true, weaponSlots: [...] }
         ]
       }
     ],
     pack_choice: {              // Separate pack selection
       id: "pack-choice",
       options: [...]
     }
   }

2. EQUIPMENT SELECTION UI (equipment-selection.js)
   ↓
   User clicks through choices
   ↓
   Stores selections in selectedChoices object:
   {
     "choice-0": { isBundle: true, bundle: [["leather",1],["longbow",1],["arrows",20]] },
     "choice-1": { isComplexChoice: true, weapons: [["rapier",1],["scimitar",1]] },
     "choice-2": { item: "handaxe", quantity: 2 }
   }
   ↓
   Exposes via window.getSelectedEquipment()

3. GAME INTRO (game-intro.js)
   ↓
   Calls window.getSelectedEquipment()
   ↓
   Converts to API format:
   {
     "choice-0": "[["leather",1],["longbow",1],["arrows",20]]",  // JSON string
     "choice-1": "[["rapier",1],["scimitar",1]]",                 // JSON string
     "choice-2": "handaxe",                                       // Simple ID
     "pack-choice": "dungeoneers-pack"                           // Pack ID
   }
   ↓
   POST to /api/character/create-save with equipmentChoices

4. GO BACKEND (character-creation-helpers.go)
   ↓
   buildInventoryFromChoices() parses each choice:
   - If starts with '[': JSON.Unmarshal to [][]interface{}
   - Otherwise: Treat as single item ID
   ↓
   Creates inventory items with quantities
   ↓
   createInventoryStructure() auto-equips based on gear_slot
   ↓
   Saves to disk
```

## Data Formats

### Simple Choice (Single Item)
**JavaScript Storage**:
```javascript
{
  item: "chain-mail",
  quantity: 1,
  type: "single"
}
```

**Sent to Go**:
```json
"choice-0": "chain-mail"
```

**Go Processing**:
```go
// Single item - add directly
allItems = append(allItems, ItemWithQty{Item: "chain-mail", Quantity: 1})
```

### Bundle Choice (Multiple Items)
**JavaScript Storage**:
```javascript
{
  item: "Leather Armor (x1) + Longbow (x1) + Arrows (x20)",
  quantity: 1,
  isBundle: true,
  bundle: [
    ["leather", 1],
    ["longbow", 1],
    ["arrows", 20]
  ],
  type: "bundle"
}
```

**Sent to Go**:
```json
"choice-0": "[["leather",1],["longbow",1],["arrows",20]]"
```

**Go Processing**:
```go
// Check if it's a JSON array
if len(selectedID) > 0 && selectedID[0] == '[' {
    var itemList [][]interface{}
    if err := json.Unmarshal([]byte(selectedID), &itemList); err == nil {
        for _, itemPair := range itemList {
            itemID := itemPair[0].(string)
            qty := int(itemPair[1].(float64))
            allItems = append(allItems, ItemWithQty{Item: itemID, Quantity: qty})
        }
    }
}
```

### Complex Weapon Choice (Multi-Slot)
**JavaScript Storage**:
```javascript
{
  item: "Choose weapon + Choose weapon",
  quantity: 1,
  isComplexChoice: true,
  type: "multi_slot",
  weaponSlots: [...],
  weapons: [               // Added after user selection
    ["rapier", 1],
    ["scimitar", 1]
  ]
}
```

**Sent to Go**:
```json
"choice-1": "[["rapier",1],["scimitar",1]]"
```

**Go Processing**: Same as bundle (JSON array)

### Pack Choice
**JavaScript Storage**:
```javascript
{
  item: "dungeoneers-pack",
  quantity: 1,
  type: "single"
}
```

**Sent to Go**:
```json
"pack-choice": "dungeoneers-pack"
```

**Go Processing**:
```go
// Expand pack into individual items
packItems := GetPackItems("dungeoneers-pack")
allItems = append(allItems, packItems...)
```

## Auto-Equipping Logic

**Location**: `src/api/character-creation-helpers.go` - `createInventoryStructure()`

**Process**:
1. For each item in `allItems`, fetch item data from database
2. Check `item.GearSlot` field:
   - `"right_arm"` or `"left_arm"` → Equip to hands
   - `"torso"` → Equip to armor slot
   - `"neck"` → Equip to necklace slot
   - `"finger"` → Equip to ring1 or ring2
   - `"ammo"` → Equip to ammunition slot
   - `"legs"` → Equip to clothes (if no armor)
3. If hands slot full or no gear slot, add to general_slots
4. If general_slots full (4 slots), add to backpack

**Hands Slot Logic**:
```go
if item.GearSlot == "right_arm" || item.GearSlot == "left_arm" {
    if item.Properties != nil && item.Properties.TwoHanded {
        // Two-handed: Takes both hands
        gearSlots.Hands.RightArm = &itemID
        gearSlots.Hands.LeftArm = &itemID
    } else {
        // One-handed: Fill right first, then left
        if gearSlots.Hands.RightArm == nil {
            gearSlots.Hands.RightArm = &itemID
        } else if gearSlots.Hands.LeftArm == nil {
            gearSlots.Hands.LeftArm = &itemID
        }
    }
}
```

## Bug Fixed: Bundle Not Appearing

### Symptom
When creating a Fighter character:
- ✅ **Choice-1** (weapons): Rapier + Scimitar → Works
- ✅ **Choice-2** (handaxes): Handaxe x2 → Works
- ✅ **Pack-choice**: Dungeoneer's Pack → Works
- ❌ **Choice-0** (bundle): Leather + Longbow + Arrows → Missing

### Root Cause (FOUND)

**Logic Error in `game-intro.js`** - Order of if/else checks was wrong!

Bundle objects have BOTH properties:
```javascript
{
  item: "Leather Armor (x1) + Longbow (x1) + Arrows (x20)",  // Display string
  isBundle: true,
  bundle: [["leather",1],["longbow",1],["arrows",20]]        // Actual items
}
```

**OLD CODE (BROKEN)** - Checked `option.item` FIRST:
```javascript
if (option.item) {
  equipmentChoices[choiceKey] = option.item;  // ❌ Sent display string!
} else if (option.isBundle && option.bundle) {
  equipmentChoices[choiceKey] = JSON.stringify(option.bundle);  // Never reached!
}
```

Result: Sent `"choice-0": "Leather Armor (x1) + Longbow (x1) + Arrows (x20)"` to Go backend, which doesn't match any item ID.

**NEW CODE (FIXED)** - Check `isBundle` FIRST:
```javascript
if (option.isBundle && option.bundle && option.bundle.length > 0) {
  console.log('📦 Bundle choice:', option.bundle);
  equipmentChoices[choiceKey] = JSON.stringify(option.bundle);  // ✅ Sends array!
} else if (option.isComplexChoice && option.weapons) {
  equipmentChoices[choiceKey] = JSON.stringify(option.weapons);
} else if (option.item) {
  equipmentChoices[choiceKey] = option.item;  // Only for simple items
}
```

Result: Sends `"choice-0": "[["leather",1],["longbow",1],["arrows",20]]"` which Go can parse correctly.

### Fix Location
**File**: `www/scripts/pages/game-intro.js`
**Line**: 1885-1908
**Change**: Reordered if/else conditions to check `isBundle` before `item`

### Testing
Test with Fighter character creation:
1. Select first choice (bundle: Leather + Longbow + Arrows)
2. Select weapons (Rapier + Scimitar)
3. Select handaxes
4. Select pack
5. Check save file has all items:
   - Leather armor auto-equipped to torso
   - Longbow in inventory (or hands if space)
   - Arrows auto-equipped to ammo slot
   - Rapier and scimitar in hands
   - Handaxes in inventory
   - Pack items in backpack

## Code Locations

### JavaScript Files

**equipment-selection.js** (Line 1-500+):
- `showEquipmentChoices()` - Main entry point
- `showRegularChoiceSelection()` - Simple/bundle choices
- `showMultiSlotChoiceSelection()` - Complex weapon choices
- `selectedChoices` object - Stores user selections
- `window.getSelectedEquipment()` - Exports selections (Line ~490)

**game-intro.js** (Line 1800-2000):
- `beginAdventure()` - Final step before save creation
- Lines 1880-1940: Converts `selectedEquipment` to `equipmentChoices`
- Line 1910: Bundle handling
- Line 1920: Complex weapon handling
- Line 1945: POST to `/api/character/create-save`

**character-generator.js** (Line 183-296):
- `generateStartingEquipment()` - Creates choice structure
- Reads from `this.startingGear` (loaded from `/data/character/starting-gear.json`)

### Go Files

**character-creation-helpers.go**:
- Line 118-200: `buildInventoryFromChoices()` - Parses equipment choices
- Line 224-283: `createInventoryStructure()` - Auto-equips items
- Line 285-325: `autoEquipItem()` - Logic for each slot type

**dnd.go**:
- Line ~300: `/api/character/create-save` endpoint
- Calls `CreateSaveFromCharacterData()`

## Test Cases

### Test 1: Fighter with Bundle
**Steps**:
1. Generate Fighter character
2. Select first choice (Leather + Longbow + Arrows)
3. Select weapons (Rapier + Scimitar)
4. Select handaxes
5. Select pack
6. Click Begin Adventure

**Expected Inventory**:
```json
{
  "gear_slots": {
    "hands": {
      "right_arm": "rapier",
      "left_arm": "scimitar"
    },
    "torso": "leather",
    "ammunition": "arrows"
  },
  "general_slots": [
    { "item": "longbow", "quantity": 1 },
    { "item": "handaxe", "quantity": 2 }
  ],
  "backpack": {
    "capacity": 20,
    "contents": [
      // Pack items...
    ]
  }
}
```

**Actual Result**: Missing leather, longbow, arrows

### Test 2: Fighter with Chain Mail
**Steps**: Same as Test 1, but select chain mail instead of bundle

**Expected**: Chain mail auto-equipped to torso
**Actual**: (Untested)

## Workarounds

None currently. Bug must be fixed for proper character creation.

## Related Issues

- **Equipment auto-equipping to wrong slots**: Separate issue (wand going to gear_slots instead of hands)
- **Item stacking logic**: Not yet implemented
- **Weight/encumbrance**: Not yet implemented

---

**Last Updated**: 2025-10-25
**Status**: Debugging in progress - waiting for server logs
