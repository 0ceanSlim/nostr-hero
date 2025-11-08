// Inventory Interaction System
// Handles drag-and-drop, context menus, and item actions

console.log('📦 Inventory interactions script loaded');

// State for drag-and-drop
let draggedItem = null;
let draggedFromSlot = null;
let draggedFromType = null; // 'inventory', 'equipment', 'general'

// Context menu state
let activeContextMenu = null;

/**
 * Initialize inventory interaction system
 */
function initializeInventoryInteractions() {
    console.log('🎮 Initializing inventory interactions');

    // NOTE: We don't bind on gameStateChange here anymore
    // The binding now happens in game-state.js AFTER updateCharacterDisplay() completes
    // This prevents race conditions where we try to bind before items are rendered

    // Close context menu when clicking elsewhere
    document.addEventListener('click', (e) => {
        if (activeContextMenu && !e.target.closest('.context-menu')) {
            closeContextMenu();
        }
    });

    // Prevent default context menu
    document.addEventListener('contextmenu', (e) => {
        if (e.target.closest('[data-item-slot]') || e.target.closest('[data-slot]')) {
            e.preventDefault();
        }
    });
}

/**
 * Bind drag-drop and click events to inventory slots
 * Note: Since slots are recreated via innerHTML='', old listeners are automatically removed
 */
function bindInventoryEvents() {
    console.log('🔧 bindInventoryEvents() called');

    // General slots (quick access)
    const generalSlots = document.querySelectorAll('#general-slots [data-item-slot]');
    console.log(`📦 Found ${generalSlots.length} general slots`);
    generalSlots.forEach((slot) => {
        // Skip if already bound (check for a marker attribute)
        if (slot.hasAttribute('data-events-bound')) {
            return;
        }
        slot.setAttribute('data-events-bound', 'true');
        // Get actual slot index from attribute, not forEach index
        const slotIndex = parseInt(slot.getAttribute('data-item-slot'), 10);
        bindSlotEvents(slot, 'general', slotIndex);
    });

    // Backpack slots
    const backpackSlots = document.querySelectorAll('#backpack-slots [data-item-slot]');
    console.log(`🎒 Found ${backpackSlots.length} backpack slots`);
    backpackSlots.forEach((slot) => {
        // Skip if already bound
        if (slot.hasAttribute('data-events-bound')) {
            return;
        }
        slot.setAttribute('data-events-bound', 'true');
        // Get actual slot index from attribute, not forEach index
        const slotIndex = parseInt(slot.getAttribute('data-item-slot'), 10);
        bindSlotEvents(slot, 'inventory', slotIndex);
    });

    // Equipment slots
    const equipmentSlots = document.querySelectorAll('[data-slot]');
    console.log(`⚔️ Found ${equipmentSlots.length} equipment slots`);
    equipmentSlots.forEach(slot => {
        // Skip if already bound
        if (slot.hasAttribute('data-events-bound')) {
            return;
        }
        slot.setAttribute('data-events-bound', 'true');
        const slotName = slot.getAttribute('data-slot');
        bindEquipmentSlotEvents(slot, slotName);
    });

    console.log('✅ bindInventoryEvents() completed');
}

/**
 * Bind events to an inventory slot
 */
function bindSlotEvents(slotElement, slotType, slotIndex) {
    const itemId = slotElement.getAttribute('data-item-id');

    console.log(`🔗 Binding events to ${slotType}[${slotIndex}], itemId: ${itemId || 'empty'}`);

    if (!itemId) {
        // Empty slot - only allow dropping
        slotElement.addEventListener('dragover', handleDragOver);
        slotElement.addEventListener('drop', (e) => handleDrop(e, slotType, slotIndex));
        console.log(`  ✓ Bound drop event to empty slot`);
        return;
    }

    // Make slot draggable
    slotElement.setAttribute('draggable', 'true');
    console.log(`  ✓ Set draggable=true`);

    // Drag events
    slotElement.addEventListener('dragstart', (e) => handleDragStart(e, itemId, slotType, slotIndex));
    slotElement.addEventListener('dragend', handleDragEnd);
    slotElement.addEventListener('dragover', handleDragOver);
    slotElement.addEventListener('drop', (e) => handleDrop(e, slotType, slotIndex));

    // Click events
    slotElement.addEventListener('click', (e) => handleLeftClick(e, itemId, slotType, slotIndex));
    slotElement.addEventListener('contextmenu', (e) => handleRightClick(e, itemId, slotType, slotIndex));

    // Hover events for tooltip
    slotElement.addEventListener('mouseenter', (e) => showItemTooltip(e, itemId, slotType));
    slotElement.addEventListener('mouseleave', hideItemTooltip);
}

/**
 * Bind events to an equipment slot
 */
function bindEquipmentSlotEvents(slotElement, slotName) {
    // Always bind click events to equipment slots (even when empty)
    // The handlers will dynamically check if there's an item at click time
    slotElement.addEventListener('click', (e) => {
        const itemId = slotElement.getAttribute('data-item-id');
        if (itemId) {
            handleLeftClick(e, itemId, 'equipment', slotName);
        }
    });

    slotElement.addEventListener('contextmenu', (e) => {
        const itemId = slotElement.getAttribute('data-item-id');
        if (itemId) {
            handleRightClick(e, itemId, 'equipment', slotName);
        }
    });

    // Bind drag events dynamically
    slotElement.addEventListener('dragstart', (e) => {
        const itemId = slotElement.getAttribute('data-item-id');
        if (itemId) {
            slotElement.setAttribute('draggable', 'true');
            handleDragStart(e, itemId, 'equipment', slotName);
        } else {
            e.preventDefault();
        }
    });
    slotElement.addEventListener('dragend', handleDragEnd);

    // Hover events (dynamic check)
    slotElement.addEventListener('mouseenter', (e) => {
        const itemId = slotElement.getAttribute('data-item-id');
        if (itemId) {
            showItemTooltip(e, itemId, 'equipment');
        }
    });
    slotElement.addEventListener('mouseleave', hideItemTooltip);

    // Always allow dropping onto equipment slots
    slotElement.addEventListener('dragover', handleDragOver);
    slotElement.addEventListener('drop', (e) => handleDropOnEquipment(e, slotName));
}

/**
 * Handle drag start
 */
function handleDragStart(e, itemId, slotType, slotIndex) {
    draggedItem = itemId;
    draggedFromSlot = slotIndex;
    draggedFromType = slotType;

    e.target.style.opacity = '0.5';
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', itemId);

}

/**
 * Handle drag end
 */
function handleDragEnd(e) {
    e.target.style.opacity = '1';
    draggedItem = null;
    draggedFromSlot = null;
    draggedFromType = null;
}

/**
 * Handle drag over (allow drop)
 */
function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
}

/**
 * Handle drop on inventory slot
 */
async function handleDrop(e, toSlotType, toSlotIndex) {
    e.preventDefault();

    if (!draggedItem) return;

    // Check if dropping on the same slot (do nothing)
    if (draggedFromType === toSlotType && draggedFromSlot === toSlotIndex) {
        console.log('🚫 Cannot drop item on itself');
        return;
    }

    // If dragging from equipment to inventory
    if (draggedFromType === 'equipment') {
        await performAction('unequip', draggedItem, draggedFromSlot, undefined, 'equipment');
    }
    // If dropping on an inventory slot
    else {
        // Check if destination slot has an item
        const state = getGameStateSync();
        let destItem = null;

        // Get destination slot item
        if (toSlotType === 'general') {
            if (state.character.inventory?.general_slots?.[toSlotIndex]) {
                destItem = state.character.inventory.general_slots[toSlotIndex];
            }
        } else if (toSlotType === 'inventory') {
            if (state.character.inventory?.gear_slots?.bag?.contents?.[toSlotIndex]) {
                destItem = state.character.inventory.gear_slots.bag.contents[toSlotIndex];
            }
        }

        // If destination has an item and it's the same type, try to stack
        if (destItem && destItem.item === draggedItem) {
            console.log(`📦 Attempting to stack ${draggedItem}`);
            await performAction('stack', draggedItem, draggedFromSlot, toSlotIndex, draggedFromType, toSlotType);
        }
        // Otherwise, move/swap as normal
        else {
            await performAction('move', draggedItem, draggedFromSlot, toSlotIndex, draggedFromType, toSlotType);
        }
    }
}

/**
 * Handle drop on equipment slot
 */
async function handleDropOnEquipment(e, equipSlotName) {
    e.preventDefault();

    if (!draggedItem) return;

    console.log(`⚔️ Equip: ${draggedItem} to ${equipSlotName}`);

    // Only allow equipping from inventory
    if (draggedFromType === 'inventory' || draggedFromType === 'general') {
        await performAction('equip', draggedItem, draggedFromSlot, equipSlotName, draggedFromType);
    }
}

/**
 * Handle left click (default action)
 */
async function handleLeftClick(e, itemId, slotType, slotIndex) {
    console.log(`🖱️ LEFT CLICK EVENT FIRED: ${itemId} in ${slotType}[${slotIndex}]`);
    e.stopPropagation();

    // Get item data to determine default action
    const itemData = getItemById(itemId);
    if (!itemData) {
        console.warn(`Item ${itemId} not found`);
        return;
    }

    // Determine default action based on item data
    const action = getDefaultAction(itemData, slotType === 'equipment');

    console.log(`🖱️ Left click: ${itemId} -> ${action}`);

    // Perform the default action
    if (action === 'equip' && itemData.gear_slot) {
        // For equip action, pass the gear_slot as the target
        await performAction(action, itemId, slotIndex, itemData.gear_slot, slotType);
    } else if (action) {
        await performAction(action, itemId, slotIndex, undefined, slotType);
    }
}

/**
 * Handle right click (context menu)
 */
function handleRightClick(e, itemId, slotType, slotIndex) {
    console.log(`🖱️ RIGHT CLICK EVENT FIRED: ${itemId} in ${slotType}[${slotIndex}]`);
    e.preventDefault();
    e.stopPropagation();

    // Get item data template
    const itemData = getItemById(itemId);
    if (!itemData) {
        console.warn(`Item ${itemId} not found`);
        return;
    }

    // Get actual inventory item to check current quantity
    const state = getGameStateSync();
    let inventoryItem = null;

    if (slotType === 'general' && state.character.inventory?.general_slots) {
        inventoryItem = state.character.inventory.general_slots.find(slot =>
            slot && slot.slot === slotIndex && slot.item === itemId
        );
    } else if (slotType === 'inventory' && state.character.inventory?.gear_slots?.bag?.contents) {
        inventoryItem = state.character.inventory.gear_slots.bag.contents.find(slot =>
            slot && slot.slot === slotIndex && slot.item === itemId
        );
    }

    // Merge template data with actual inventory quantity
    const itemWithQuantity = {
        ...itemData,
        quantity: inventoryItem?.quantity || 1
    };

    // Get available actions
    const isEquipped = slotType === 'equipment';
    const actions = getItemActions(itemWithQuantity, isEquipped);

    console.log(`📋 Showing context menu with ${actions.length} actions`);

    // Show context menu
    showContextMenu(e.clientX, e.clientY, itemId, slotIndex, slotType, actions);
}

/**
 * Get default action for an item
 */
function getDefaultAction(itemData, isEquipped) {
    if (isEquipped) {
        return 'unequip';
    }

    // Check if item has "equipment" tag - this is the primary indicator
    if (itemData.tags && itemData.tags.includes('equipment')) {
        return 'equip';
    }

    // Fallback: Check item type for backwards compatibility
    const itemType = itemData.type || itemData.item_type;
    const weaponTypes = ['Weapon', 'Melee Weapon', 'Ranged Weapon', 'Simple Weapon', 'Martial Weapon'];
    const armorTypes = ['Armor', 'Light Armor', 'Medium Armor', 'Heavy Armor', 'Shield'];
    const wearableTypes = ['Ring', 'Necklace', 'Amulet', 'Cloak', 'Boots', 'Gloves', 'Helmet', 'Hat'];
    const ammunitionTypes = ['Ammunition', 'Ammo'];
    const consumableTypes = ['Potion', 'Consumable', 'Food'];

    if (weaponTypes.includes(itemType) || armorTypes.includes(itemType) || wearableTypes.includes(itemType) || ammunitionTypes.includes(itemType)) {
        return 'equip';
    }

    if (consumableTypes.includes(itemType)) {
        return 'use';
    }

    return 'examine';
}

/**
 * Get available actions for an item
 */
function getItemActions(itemData, isEquipped) {
    const actions = [];

    if (isEquipped) {
        actions.push({ action: 'unequip', label: 'Unequip' });
    } else {
        // Check for equipment tag first
        if (itemData.tags && itemData.tags.includes('equipment')) {
            actions.push({ action: 'equip', label: 'Equip' });
        } else {
            // Fallback to type checking
            const itemType = itemData.type || itemData.item_type;
            const weaponTypes = ['Weapon', 'Melee Weapon', 'Ranged Weapon', 'Simple Weapon', 'Martial Weapon'];
            const armorTypes = ['Armor', 'Light Armor', 'Medium Armor', 'Heavy Armor', 'Shield'];
            const wearableTypes = ['Ring', 'Necklace', 'Amulet', 'Cloak', 'Boots', 'Gloves', 'Helmet', 'Hat'];
            const ammunitionTypes = ['Ammunition', 'Ammo'];

            if (weaponTypes.includes(itemType) || armorTypes.includes(itemType) || wearableTypes.includes(itemType) || ammunitionTypes.includes(itemType)) {
                actions.push({ action: 'equip', label: 'Equip' });
            }
        }

        // Check for consumables
        const consumableTypes = ['Potion', 'Consumable', 'Food'];
        const itemType = itemData.type || itemData.item_type;
        if (consumableTypes.includes(itemType)) {
            actions.push({ action: 'use', label: 'Use' });
        }

        // Add split action for stackable items (quantity > 1)
        if (itemData.quantity && itemData.quantity > 1) {
            actions.push({ action: 'split', label: 'Split' });
        }
    }

    // Examine is always available
    actions.push({ action: 'examine', label: 'Examine' });

    // Drop is always last
    actions.push({ action: 'drop', label: 'Drop' });

    return actions;
}

/**
 * Show context menu
 */
function showContextMenu(x, y, itemId, slotIndex, slotType, actions) {
    // Close existing menu
    closeContextMenu();

    // Create context menu
    const menu = document.createElement('div');
    menu.className = 'context-menu fixed bg-gray-800 border-2 border-gray-600 shadow-lg z-50';
    menu.style.left = `${x}px`;
    menu.style.top = `${y}px`;
    menu.style.minWidth = '120px';

    // Add actions
    actions.forEach(({ action, label }) => {
        const item = document.createElement('div');
        item.className = 'context-menu-item px-3 py-2 hover:bg-gray-700 cursor-pointer text-sm text-white';
        item.textContent = label;
        item.addEventListener('click', async () => {
            // Pass parameters correctly: action, itemId, fromSlot, toSlotOrType, fromSlotType, toSlotType
            // For equip action, we need to pass the gear_slot from item data
            let toSlotOrType = undefined;
            if (action === 'equip') {
                const itemData = getItemById(itemId);
                toSlotOrType = itemData?.gear_slot || undefined;
            }

            await performAction(action, itemId, slotIndex, toSlotOrType, slotType, undefined);
            closeContextMenu();
        });
        menu.appendChild(item);
    });

    document.body.appendChild(menu);
    activeContextMenu = menu;

    // Adjust position if menu goes off screen
    const rect = menu.getBoundingClientRect();
    if (rect.right > window.innerWidth) {
        menu.style.left = `${window.innerWidth - rect.width - 5}px`;
    }
    if (rect.bottom > window.innerHeight) {
        menu.style.top = `${window.innerHeight - rect.height - 5}px`;
    }
}

/**
 * Close context menu
 */
function closeContextMenu() {
    if (activeContextMenu) {
        activeContextMenu.remove();
        activeContextMenu = null;
    }
}

/**
 * Perform an item action
 */
async function performAction(action, itemId, fromSlot, toSlotOrType, fromSlotType, toSlotType) {

    // Special case: examine (no backend call needed)
    if (action === 'examine') {
        showItemDetails(itemId);
        return;
    }

    // Special case: split stack (handle client-side for now)
    if (action === 'split') {
        await handleSplitStack(itemId, fromSlot, fromSlotType);
        return;
    }

    // Check if Game API is initialized
    if (!window.gameAPI || !window.gameAPI.initialized) {
        console.error('Game API not initialized');
        showMessage('❌ Game not initialized', 'error');
        return;
    }

    // Prepare parameters for the new Game API
    const params = {
        item_id: itemId,
        from_slot: fromSlotType === 'equipment' ? -1 : (typeof fromSlot === 'number' ? fromSlot : -1),
        to_slot: typeof toSlotOrType === 'number' ? toSlotOrType : -1,
        from_slot_type: fromSlotType || '',
        to_slot_type: toSlotType || '',
        from_equip: fromSlotType === 'equipment' ? fromSlot : '',
        to_equip: typeof toSlotOrType === 'string' ? toSlotOrType : '',
        equipment_slot: typeof toSlotOrType === 'string' ? toSlotOrType : '',
        quantity: 1
    };

    // Special case: drop - prompt for quantity, but only add to ground after API success
    let dropInfo = null;  // Store drop details for later
    if (action === 'drop') {
        // Get item data from inventory to check current quantity
        const state = getGameStateSync();
        let inventoryItem = null;

        // Find the SPECIFIC item at the clicked slot (use fromSlot, not find())
        if (fromSlotType === 'general' && state.character.inventory?.general_slots) {
            // Find by matching the slot index
            inventoryItem = state.character.inventory.general_slots.find(slot => slot && slot.slot === fromSlot && slot.item === itemId);
        } else if (fromSlotType === 'inventory' && state.character.inventory?.gear_slots?.bag?.contents) {
            // Find by matching the slot index
            inventoryItem = state.character.inventory.gear_slots.bag.contents.find(slot => slot && slot.slot === fromSlot && slot.item === itemId);
        }

        const currentQuantity = inventoryItem?.quantity || 1;
        const itemData = getItemById(itemId);

        // If quantity > 1, show prompt for how many to drop
        if (currentQuantity > 1) {
            const dropQuantity = await promptDropQuantity(itemData?.name || itemId, currentQuantity);

            if (dropQuantity === null || dropQuantity <= 0) {
                // User cancelled or entered invalid amount
                return;
            }

            // Set the drop quantity in the params
            params.quantity = dropQuantity;

            // Store drop info for after successful API call
            dropInfo = {
                itemId: itemId,
                quantity: dropQuantity,
                itemName: itemData?.name || itemId
            };
        } else {
            // Single item
            params.quantity = 1;

            // Store drop info for after successful API call
            dropInfo = {
                itemId: itemId,
                quantity: 1,
                itemName: itemData?.name || itemId
            };
        }
    }

    try {
        // Map old action names to new game action types
        const actionMap = {
            'equip': 'equip_item',
            'unequip': 'unequip_item',
            'use': 'use_item',
            'drop': 'drop_item',
            'move': 'move_item',
            'stack': 'stack_item',
            'add': 'add_item'
        };

        const gameAction = actionMap[action] || action;

        console.log(`📤 Sending action: ${gameAction}`, params);

        // Send action to Go backend
        const result = await window.gameAPI.sendAction(gameAction, params);

        if (result.success) {
            console.log('✅ Action successful:', result.message);
            showMessage(result.message, 'success');

            // Refresh game state from Go memory
            await refreshGameState();

            // If this was a drop action, NOW add to ground (after successful API call)
            if (action === 'drop' && dropInfo) {
                window.addItemToGround(dropInfo.itemId, dropInfo.quantity);
                window.addGameLog(`Dropped ${dropInfo.quantity}x ${dropInfo.itemName} on the ground.`);

                // Refresh ground modal if it's open
                if (window.refreshGroundModal) {
                    window.refreshGroundModal();
                }
            }
        } else {
            console.error('❌ Action failed:', result.error);
            showMessage(result.error || 'Action failed', 'error');
        }
    } catch (error) {
        console.error('❌ Error performing action:', error);
        showMessage(`❌ Failed to perform action: ${error.message}`, 'error');
    }
}

/**
 * Prompt user for quantity to drop
 */
async function promptDropQuantity(itemName, maxQuantity) {
    return new Promise((resolve) => {
        // Create modal backdrop
        const modal = document.createElement('div');
        modal.style.position = 'fixed';
        modal.style.top = '0';
        modal.style.left = '0';
        modal.style.width = '100%';
        modal.style.height = '100%';
        modal.style.background = 'rgba(0, 0, 0, 0.8)';
        modal.style.zIndex = '100';
        modal.style.display = 'flex';
        modal.style.alignItems = 'center';
        modal.style.justifyContent = 'center';

        // Create dialog box
        const dialog = document.createElement('div');
        dialog.style.background = '#2a2a2a';
        dialog.style.border = '2px solid #4a4a4a';
        dialog.style.padding = '20px';
        dialog.style.minWidth = '300px';
        dialog.style.boxShadow = 'inset 1px 1px 0 #3a3a3a, inset -1px -1px 0 #000000';

        dialog.innerHTML = `
            <div style="color: white; font-size: 12px;">
                <h3 style="margin: 0 0 15px 0; font-weight: bold;">Drop ${itemName}</h3>
                <p style="margin: 0 0 10px 0; color: #ccc;">How many do you want to drop? (Max: ${maxQuantity})</p>
                <input type="number" id="drop-quantity-input" min="1" max="${maxQuantity}" value="${maxQuantity}"
                    style="width: 100%; padding: 5px; background: #1a1a1a; color: white; border: 2px solid #4a4a4a; font-size: 12px;" />
                <div style="margin-top: 15px; display: flex; gap: 10px; justify-content: flex-end;">
                    <button id="drop-cancel-btn" style="padding: 5px 15px; background: #3a3a3a; color: white; border: 2px solid #4a4a4a; cursor: pointer; font-size: 11px;">Cancel</button>
                    <button id="drop-confirm-btn" style="padding: 5px 15px; background: #4a4a4a; color: white; border: 2px solid #6a6a6a; cursor: pointer; font-size: 11px;">Drop</button>
                </div>
            </div>
        `;

        modal.appendChild(dialog);
        document.body.appendChild(modal);

        const input = document.getElementById('drop-quantity-input');
        const cancelBtn = document.getElementById('drop-cancel-btn');
        const confirmBtn = document.getElementById('drop-confirm-btn');

        // Focus and select input
        input.focus();
        input.select();

        // Event handlers
        const cleanup = () => {
            modal.remove();
        };

        cancelBtn.onclick = () => {
            cleanup();
            resolve(null);
        };

        confirmBtn.onclick = () => {
            const quantity = parseInt(input.value);
            if (quantity > 0 && quantity <= maxQuantity) {
                cleanup();
                resolve(quantity);
            } else {
                input.style.borderColor = '#ff0000';
            }
        };

        input.onkeydown = (e) => {
            if (e.key === 'Enter') {
                confirmBtn.click();
            } else if (e.key === 'Escape') {
                cancelBtn.click();
            }
        };

        // Click outside to cancel
        modal.onclick = (e) => {
            if (e.target === modal) {
                cancelBtn.click();
            }
        };
    });
}

/**
 * Handle splitting a stack into two stacks
 */
async function handleSplitStack(itemId, fromSlot, fromSlotType) {
    // Get item data from inventory to check current quantity
    const state = getGameStateSync();
    let inventoryItem = null;

    // Find the item in inventory to get actual quantity
    if (fromSlotType === 'general' && state.character.inventory?.general_slots) {
        inventoryItem = state.character.inventory.general_slots[fromSlot];
    } else if (fromSlotType === 'inventory' && state.character.inventory?.gear_slots?.bag?.contents) {
        inventoryItem = state.character.inventory.gear_slots.bag.contents[fromSlot];
    }

    if (!inventoryItem || inventoryItem.quantity <= 1) {
        window.addGameLog('Cannot split a stack of 1 item.');
        return;
    }

    const currentQuantity = inventoryItem.quantity;
    const itemData = getItemById(itemId);

    // Show prompt for how many to split
    const splitQuantity = await promptSplitQuantity(itemData?.name || itemId, currentQuantity);

    if (splitQuantity === null || splitQuantity <= 0 || splitQuantity >= currentQuantity) {
        // User cancelled or entered invalid amount
        return;
    }

    // Find an empty slot in inventory
    let emptySlotIndex = -1;
    let emptySlotType = '';

    // Check backpack first (more space)
    if (state.character.inventory?.gear_slots?.bag?.contents) {
        const backpackSlots = state.character.inventory.gear_slots.bag.contents;

        // Build a set of used slot numbers by checking the 'slot' field of each item
        const usedSlots = new Set();
        backpackSlots.forEach(slot => {
            if (slot && slot.slot !== undefined && slot.item !== null && slot.item !== '') {
                usedSlots.add(slot.slot);
            }
        });

        // Find first unused slot number (0-19)
        for (let i = 0; i < 20; i++) {
            if (!usedSlots.has(i)) {
                emptySlotIndex = i;
                emptySlotType = 'inventory';
                break;
            }
        }
    }

    // If no empty backpack slot, check general slots
    if (emptySlotIndex === -1 && state.character.inventory?.general_slots) {
        const generalSlots = state.character.inventory.general_slots;

        // Build a set of used slot numbers
        const usedSlots = new Set();
        generalSlots.forEach(slot => {
            if (slot && slot.slot !== undefined && slot.item !== null && slot.item !== '') {
                usedSlots.add(slot.slot);
            }
        });

        // Find first unused slot number (0-3)
        for (let i = 0; i < 4; i++) {
            if (!usedSlots.has(i)) {
                emptySlotIndex = i;
                emptySlotType = 'general';
                break;
            }
        }
    }

    // If no empty slot, show error
    if (emptySlotIndex === -1) {
        window.addGameLog('❌ Inventory full - cannot split stack');
        window.showMessage('Inventory full - cannot split stack', 'error');
        return;
    }

    // Call new in-memory game action system
    if (!window.gameAPI || !window.gameAPI.initialized) {
        console.error('Game API not initialized');
        showMessage('❌ Game not initialized', 'error');
        return;
    }

    try {
        // Send split action to Go backend (in-memory)
        const result = await window.gameAPI.sendAction('split_item', {
            item_id: itemId,
            from_slot: fromSlot,
            to_slot: emptySlotIndex,
            from_slot_type: fromSlotType,
            to_slot_type: emptySlotType,
            quantity: splitQuantity
        });

        // Refresh UI from Go memory
        await refreshGameState();

        window.addGameLog(`Split ${splitQuantity} from stack of ${itemData?.name || itemId}`);
        showMessage(`✂️ Split ${splitQuantity} items into new stack`, 'success');

    } catch (error) {
        console.error('Failed to split stack:', error);
        showMessage('❌ Error splitting stack: ' + error.message, 'error');
    }
}

/**
 * Prompt user for quantity to split from stack
 */
async function promptSplitQuantity(itemName, maxQuantity) {
    const maxSplit = maxQuantity - 1; // Can't split all items
    const defaultSplit = Math.floor(maxQuantity / 2);

    return new Promise((resolve) => {
        // Create modal backdrop
        const modal = document.createElement('div');
        modal.style.position = 'fixed';
        modal.style.top = '0';
        modal.style.left = '0';
        modal.style.width = '100%';
        modal.style.height = '100%';
        modal.style.background = 'rgba(0, 0, 0, 0.8)';
        modal.style.zIndex = '100';
        modal.style.display = 'flex';
        modal.style.alignItems = 'center';
        modal.style.justifyContent = 'center';

        // Create dialog box
        const dialog = document.createElement('div');
        dialog.style.background = '#2a2a2a';
        dialog.style.border = '2px solid #4a4a4a';
        dialog.style.padding = '20px';
        dialog.style.minWidth = '300px';
        dialog.style.boxShadow = 'inset 1px 1px 0 #3a3a3a, inset -1px -1px 0 #000000';

        dialog.innerHTML = `
            <div style="color: white; font-size: 12px;">
                <h3 style="margin: 0 0 15px 0; font-weight: bold;">Split ${itemName}</h3>
                <p style="margin: 0 0 10px 0; color: #ccc;">How many to split into new stack? (Max: ${maxSplit})</p>
                <input type="number" id="split-quantity-input" min="1" max="${maxSplit}" value="${defaultSplit}"
                    style="width: 100%; padding: 5px; background: #1a1a1a; color: white; border: 2px solid #4a4a4a; font-size: 12px;" />
                <div style="margin-top: 15px; display: flex; gap: 10px; justify-content: flex-end;">
                    <button id="split-cancel-btn" style="padding: 5px 15px; background: #3a3a3a; color: white; border: 2px solid #4a4a4a; cursor: pointer; font-size: 11px;">Cancel</button>
                    <button id="split-confirm-btn" style="padding: 5px 15px; background: #4a4a4a; color: white; border: 2px solid #6a6a6a; cursor: pointer; font-size: 11px;">Split</button>
                </div>
            </div>
        `;

        modal.appendChild(dialog);
        document.body.appendChild(modal);

        const input = document.getElementById('split-quantity-input');
        const cancelBtn = document.getElementById('split-cancel-btn');
        const confirmBtn = document.getElementById('split-confirm-btn');

        // Focus and select input
        input.focus();
        input.select();

        // Event handlers
        const cleanup = () => {
            modal.remove();
        };

        cancelBtn.onclick = () => {
            cleanup();
            resolve(null);
        };

        confirmBtn.onclick = () => {
            const quantity = parseInt(input.value);
            if (quantity > 0 && quantity <= maxSplit) {
                cleanup();
                resolve(quantity);
            } else {
                input.style.borderColor = '#ff0000';
            }
        };

        input.onkeydown = (e) => {
            if (e.key === 'Enter') {
                confirmBtn.click();
            } else if (e.key === 'Escape') {
                cancelBtn.click();
            }
        };

        // Click outside to cancel
        modal.onclick = (e) => {
            if (e.target === modal) {
                cancelBtn.click();
            }
        };
    });
}

/**
 * Show item details modal
 */
function showItemDetails(itemId) {
    const itemData = getItemById(itemId);
    if (!itemData) {
        console.warn(`Item ${itemId} not found`);
        return;
    }

    // Find the scene container
    const sceneImage = document.getElementById('scene-image');
    const sceneContainer = sceneImage ? sceneImage.parentElement : null;

    if (!sceneContainer) {
        console.warn('Scene container not found');
        return;
    }

    // Create modal overlay within scene
    const modal = document.createElement('div');
    modal.className = 'absolute inset-0 bg-black bg-opacity-80 flex items-center justify-center z-50';
    modal.addEventListener('click', (e) => {
        if (e.target === modal) modal.remove();
    });

    const content = document.createElement('div');
    content.className = 'bg-gray-800 border-4 border-gray-600 p-3 max-w-xs w-full mx-4';
    content.style.clipPath = 'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))';

    // Build properties display - check both top-level and properties object
    const properties = [];
    const props = itemData.properties || {};

    // Helper to get property value
    const getProp = (name) => itemData[name] ?? props[name];

    // Always show basic info
    properties.push(`<div><strong class="text-gray-400">Type:</strong> ${itemData.type || itemData.item_type || 'Unknown'}</div>`);
    properties.push(`<div><strong class="text-gray-400">Rarity:</strong> <span class="capitalize">${itemData.rarity || 'common'}</span></div>`);

    const weight = getProp('weight');
    if (weight && weight > 0) {
        properties.push(`<div><strong class="text-gray-400">⚖️</strong> ${weight} lb</div>`);
    }

    const price = getProp('price');
    if (price && price > 0) {
        properties.push(`<div><strong class="text-yellow-400">💰</strong> ${price} gp</div>`);
    }

    // Weapon properties
    const damage = getProp('damage');
    if (damage && damage !== 'null' && damage !== null) {
        properties.push(`<div><strong class="text-red-400">Damage:</strong> ${damage}</div>`);
    }

    const damageType = getProp('damage-type');
    if (damageType && damageType !== 'null') {
        properties.push(`<div><strong class="text-red-400">Damage Type:</strong> <span class="capitalize">${damageType}</span></div>`);
    }

    const range = getProp('range');
    if (range && range !== 'null' && range !== null) {
        properties.push(`<div><strong class="text-blue-400">Range:</strong> ${range} ft</div>`);
    }

    const rangeLong = getProp('range-long');
    if (rangeLong && rangeLong !== 'null' && rangeLong !== null) {
        properties.push(`<div><strong class="text-blue-400">Long Range:</strong> ${rangeLong} ft</div>`);
    }

    // Armor properties
    const ac = getProp('ac');
    if (ac && ac !== 'null' && ac !== null) {
        properties.push(`<div><strong class="text-blue-400">AC:</strong> ${ac}</div>`);
    }

    // Healing properties
    const heal = getProp('heal');
    if (heal && heal !== 'null' && heal !== null) {
        properties.push(`<div><strong class="text-green-400">Healing:</strong> ${heal}</div>`);
    }

    // Effects (for consumables)
    const effects = props.effects;
    if (effects && Array.isArray(effects) && effects.length > 0) {
        const effectsList = effects.map(effect => {
            const sign = effect.value > 0 ? '+' : '';
            return `${effect.type} ${sign}${effect.value}`;
        }).join(', ');
        properties.push(`<div class="col-span-2"><strong class="text-green-400">Effects:</strong> ${effectsList}</div>`);
    }

    // Tags
    const tags = itemData.tags || props.tags;
    if (tags && Array.isArray(tags) && tags.length > 0) {
        const tagsList = tags
            .filter(tag => tag !== 'equipment') // Hide generic equipment tag
            .map(tag => `<span class="px-2 py-0.5 bg-gray-700 rounded text-xs capitalize">${tag.replace(/-/g, ' ')}</span>`)
            .join(' ');
        if (tagsList) {
            properties.push(`<div class="col-span-2 mt-2"><strong class="text-gray-400">Tags:</strong> ${tagsList}</div>`);
        }
    }

    // Gear slot
    const gearSlot = getProp('gear_slot');
    if (gearSlot) {
        properties.push(`<div class="col-span-2"><strong class="text-purple-400">Equip Slot:</strong> <span class="capitalize">${gearSlot.replace(/_/g, ' ')}</span></div>`);
    }

    content.innerHTML = `
        <h2 class="text-base font-bold text-yellow-400 mb-1">${itemData.name}</h2>
        ${itemData.image ? `<img src="${itemData.image}" alt="${itemData.name}" class="w-16 mx-auto mb-1" style="image-rendering: pixelated;">` : ''}
        <p class="text-gray-300 mb-2 italic text-xs">${itemData.description || itemData.ai_description || 'No description available.'}</p>
        <div class="grid grid-cols-2 gap-1 text-xs mb-2">
            ${properties.join('\n            ')}
        </div>
        <button class="mt-1 px-2 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs w-full">Close</button>
    `;

    content.querySelector('button').addEventListener('click', () => modal.remove());

    modal.appendChild(content);
    sceneContainer.appendChild(modal);
}

/**
 * Show action text on hover (in bottom-right of game scene)
 */
function showItemTooltip(e, itemId, slotType) {

    // Get item data
    const itemData = getItemById(itemId);
    if (!itemData) {
        console.warn(`Item ${itemId} not found for tooltip`);
        return;
    }

    // Determine default action
    const isEquipped = slotType === 'equipment';
    const defaultAction = getDefaultAction(itemData, isEquipped);

    // Action display names (simple text)
    const actionNames = {
        'equip': 'Equip',
        'unequip': 'Unequip',
        'use': 'Use',
        'examine': 'Info',
        'drop': 'Drop'
    };

    // Action colors
    const actionColors = {
        'equip': '#4a9eff',      // Blue for equip
        'unequip': '#ff8c00',    // Orange for unequip
        'use': '#00ff00',        // Green for use
        'examine': '#ff8c00',    // Orange for info
        'drop': '#ff0000'        // Red for drop
    };

    const actionName = actionNames[defaultAction] || 'Info';
    const actionColor = actionColors[defaultAction] || '#ff8c00';

    // Update action text in bottom-right corner
    const actionText = document.getElementById('action-text');
    if (actionText) {
        actionText.textContent = actionName;
        actionText.style.color = actionColor;
        actionText.style.display = 'block';
    } else {
        console.warn('⚠️ action-text element not found!');
    }
}

/**
 * Hide action text
 */
function hideItemTooltip() {
    const actionText = document.getElementById('action-text');
    if (actionText) {
        actionText.style.display = 'none';
    }
}

/**
 * Get save ID from URL or session
 */
function getSaveId() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('save');
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeInventoryInteractions);
} else {
    initializeInventoryInteractions();
}

// Export functions for use in other modules
window.inventoryInteractions = {
    initializeInventoryInteractions,
    bindInventoryEvents,
    performAction,
    showItemDetails
};
