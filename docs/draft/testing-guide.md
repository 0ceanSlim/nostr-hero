# Testing Guide - Go-First Migration

## Quick Fixes Applied

### ✅ Fixed Issues:
1. **Stats normalization** - Converted capital letter keys to lowercase
2. **Location/district mapping** - Mapped display names to IDs
3. **Added drag-drop logging** - See what's being bound

## Testing Steps

### 1. Refresh Browser
**Ctrl+F5** to clear cache

### 2. Check Console Output

You should see this sequence:
```
✅ Session loaded into Go memory
🔄 Refreshing game state from Go...
🔄 Transforming save data to UI state: {
  savedLocation: "Verdant City",
  savedDistrict: "Garden Plaza",
  mappedLocationId: "verdant",
  mappedDistrictKey: "center",
  statsNormalized: {strength: 8, dexterity: 16, ...}
}
📦 Fetched state: {hasCharacter: true, hasInventory: true, ...}
📢 Dispatching gameStateChange event...
🎨 gameStateChange event fired, updating displays...
🎨 updateCharacterDisplay() starting...
🔄 Rebinding inventory events...
🔧 bindInventoryEvents() called
📦 Found 4 general slots
🎒 Found 20 backpack slots
⚔️ Found X equipment slots
🔗 Binding events to general[0], itemId: empty
🔗 Binding events to inventory[0], itemId: robes
  ✓ Set draggable=true
... (more binding logs)
✅ bindInventoryEvents() completed
✅ All displays updated
```

### 3. Manual Tests in Console

#### Test 1: Check State
```javascript
// Get current cached state
const state = getGameStateSync();
console.log('Character:', state.character.name);
console.log('Stats:', state.character.stats);
console.log('Location:', state.location.current);
console.log('District:', state.location.district);
console.log('Inventory:', state.inventory);
console.log('Equipment:', state.equipment);
```

**Expected:**
- Character: "ocean"
- Stats: `{strength: 8, dexterity: 16, constitution: 8, ...}`
- Location: "verdant"
- District: "center"
- Inventory: Array of 4 general slot objects
- Equipment: Object with gear_slots

#### Test 2: Check if Game API is Ready
```javascript
console.log('API Initialized:', window.gameAPI.initialized);
console.log('API npub:', window.gameAPI.npub);
console.log('API saveID:', window.gameAPI.saveID);
```

**Expected:**
- All should be set

#### Test 3: Check Drag-Drop Binding
```javascript
// Find a slot with an item
const itemSlot = document.querySelector('[data-item-id="robes"]');
console.log('Found robes slot:', !!itemSlot);
console.log('Is draggable:', itemSlot?.draggable);
console.log('Has events bound:', itemSlot?.hasAttribute('data-events-bound'));
```

**Expected:**
- Found: true
- Draggable: true
- Has events bound: true

#### Test 4: Manually Trigger Rebind
```javascript
// Force rebind events
window.inventoryInteractions.bindInventoryEvents();
```

#### Test 5: Test an Action
```javascript
// Try moving an item from slot 0 to slot 1 in backpack
await window.gameAPI.sendAction('move_item', {
    item_id: 'robes',
    from_slot: 0,
    to_slot: 1,
    from_slot_type: 'inventory',
    to_slot_type: 'inventory'
});

// Refresh UI
await refreshGameState();
```

**Expected:**
- Should see: `📤 Sending action: move_item`
- Should see: `✅ Action completed: Item moved`
- Robes should move from slot 0 to slot 1

## Common Issues & Solutions

### Issue: Stats not showing
**Check:** Console log should show `statsNormalized: {strength: 8, ...}`
**Fix:** Already applied - stats keys normalized to lowercase

### Issue: Location not loading
**Check:** Console log should show `mappedLocationId: "verdant", mappedDistrictKey: "center"`
**Fix:** Already applied - location names mapped to IDs

### Issue: Drag-drop doesn't work
**Possible Causes:**

1. **Events not binding**
   - Check console for "bindInventoryEvents() called"
   - Check if you see binding logs for each slot
   - Try manual rebind: `window.inventoryInteractions.bindInventoryEvents()`

2. **Slots not rendering**
   - Check if slots exist: `document.querySelectorAll('[data-item-slot]').length`
   - Should be 24 total (4 general + 20 backpack)

3. **Items not set as draggable**
   - Check a slot: `document.querySelector('[data-item-id="robes"]').draggable`
   - Should be `true`

### Issue: Drop/Split not working

**For Drop:**
```javascript
// Test drop action directly
await window.gameAPI.sendAction('drop_item', {
    item_id: 'robes',
    from_slot: 0,
    from_slot_type: 'inventory'
});
await refreshGameState();
```

**For Split:**
- Split is still handled client-side (not migrated yet)
- Check console for errors when right-clicking item

## What Should Work Now

✅ **Working:**
- Load game from Go memory
- Display character stats (STR, DEX, CON, etc.)
- Display location and district
- Display inventory items
- Display equipped items
- Game API initialized

⚠️ **Needs Testing:**
- Drag and drop items
- Equip/unequip
- Use items
- Drop items
- Split stacks

❌ **Not Yet Migrated:**
- Location movement (still uses old system)
- Time advancement (still uses old system)
- Combat actions
- Spell casting

## Next Steps

If drag-drop still doesn't work after refresh:

1. Open browser console (F12)
2. Look for binding logs
3. Run manual tests above
4. Report what you see - specifically:
   - Are events being bound? (look for "🔗 Binding events" logs)
   - How many slots were found?
   - Is draggable set to true?
   - Any errors when trying to drag?

## Debug Commands

```javascript
// Full state dump
console.log(JSON.stringify(getGameStateSync(), null, 2));

// Check API
console.log('API:', window.gameAPI);

// Check inventory system
console.log('Inv System:', window.inventoryInteractions);

// Force UI refresh
await refreshGameState();

// Force rebind
window.inventoryInteractions.bindInventoryEvents();
```
