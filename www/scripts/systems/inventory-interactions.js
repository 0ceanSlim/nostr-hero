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
    // General slots (quick access)
    const generalSlots = document.querySelectorAll('#general-slots [data-item-slot]');
    generalSlots.forEach((slot, index) => {
        // Skip if already bound (check for a marker attribute)
        if (slot.hasAttribute('data-events-bound')) {
            return;
        }
        slot.setAttribute('data-events-bound', 'true');
        bindSlotEvents(slot, 'general', index);
    });

    // Backpack slots
    const backpackSlots = document.querySelectorAll('#backpack-slots [data-item-slot]');
    backpackSlots.forEach((slot, index) => {
        // Skip if already bound
        if (slot.hasAttribute('data-events-bound')) {
            return;
        }
        slot.setAttribute('data-events-bound', 'true');
        bindSlotEvents(slot, 'inventory', index);
    });

    // Equipment slots
    const equipmentSlots = document.querySelectorAll('[data-slot]');
    equipmentSlots.forEach(slot => {
        // Skip if already bound
        if (slot.hasAttribute('data-events-bound')) {
            return;
        }
        slot.setAttribute('data-events-bound', 'true');
        const slotName = slot.getAttribute('data-slot');
        bindEquipmentSlotEvents(slot, slotName);
    });
}

/**
 * Bind events to an inventory slot
 */
function bindSlotEvents(slotElement, slotType, slotIndex) {
    const itemId = slotElement.getAttribute('data-item-id');

    if (!itemId) {
        // Empty slot - only allow dropping
        slotElement.addEventListener('dragover', handleDragOver);
        slotElement.addEventListener('drop', (e) => handleDrop(e, slotType, slotIndex));
        return;
    }

    // Make slot draggable
    slotElement.setAttribute('draggable', 'true');

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
    // Check if slot has an item (look for data-item-id on the slot or inside it)
    let itemId = slotElement.getAttribute('data-item-id');
    if (!itemId) {
        const itemData = slotElement.querySelector('[data-item-id]');
        itemId = itemData?.getAttribute('data-item-id');
    }

    if (itemId) {
        // Make slot draggable
        slotElement.setAttribute('draggable', 'true');

        // Drag events
        slotElement.addEventListener('dragstart', (e) => handleDragStart(e, itemId, 'equipment', slotName));
        slotElement.addEventListener('dragend', handleDragEnd);

        // Click events
        slotElement.addEventListener('click', (e) => handleLeftClick(e, itemId, 'equipment', slotName));
        slotElement.addEventListener('contextmenu', (e) => handleRightClick(e, itemId, 'equipment', slotName));

        // Hover events
        slotElement.addEventListener('mouseenter', (e) => showItemTooltip(e, itemId, 'equipment'));
        slotElement.addEventListener('mouseleave', hideItemTooltip);
    }

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

    // If dragging from equipment to inventory
    if (draggedFromType === 'equipment') {
        await performAction('unequip', draggedItem, draggedFromSlot);
    }
    // If moving within inventory (pass slot types to backend)
    else if (draggedFromType === toSlotType) {
        await performAction('move', draggedItem, draggedFromSlot, toSlotIndex, draggedFromType, toSlotType);
    }
    // If moving between inventory types (general to backpack or vice versa)
    else {
        await performAction('move', draggedItem, draggedFromSlot, toSlotIndex, draggedFromType, toSlotType);
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
        await performAction('equip', draggedItem, draggedFromSlot, equipSlotName);
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

    // Determine default action based on item type
    const action = getDefaultAction(itemData.type || itemData.item_type, slotType === 'equipment');

    console.log(`🖱️ Left click: ${itemId} -> ${action}`);

    // Perform the default action
    if (action) {
        await performAction(action, itemId, slotIndex);
    }
}

/**
 * Handle right click (context menu)
 */
function handleRightClick(e, itemId, slotType, slotIndex) {
    console.log(`🖱️ RIGHT CLICK EVENT FIRED: ${itemId} in ${slotType}[${slotIndex}]`);
    e.preventDefault();
    e.stopPropagation();

    // Get item data
    const itemData = getItemById(itemId);
    if (!itemData) {
        console.warn(`Item ${itemId} not found`);
        return;
    }

    // Get available actions
    const isEquipped = slotType === 'equipment';
    const actions = getItemActions(itemData.type || itemData.item_type, isEquipped);

    console.log(`📋 Showing context menu with ${actions.length} actions`);

    // Show context menu
    showContextMenu(e.clientX, e.clientY, itemId, slotIndex, slotType, actions);
}

/**
 * Get default action for an item type
 */
function getDefaultAction(itemType, isEquipped) {
    if (isEquipped) {
        return 'unequip';
    }

    const weaponTypes = ['Weapon', 'Melee Weapon', 'Ranged Weapon', 'Simple Weapon', 'Martial Weapon'];
    const armorTypes = ['Armor', 'Light Armor', 'Medium Armor', 'Heavy Armor', 'Shield'];
    const wearableTypes = ['Ring', 'Necklace', 'Amulet', 'Cloak', 'Boots', 'Gloves', 'Helmet', 'Hat'];
    const consumableTypes = ['Potion', 'Consumable', 'Food'];

    if (weaponTypes.includes(itemType) || armorTypes.includes(itemType) || wearableTypes.includes(itemType)) {
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
function getItemActions(itemType, isEquipped) {
    const actions = [];

    const weaponTypes = ['Weapon', 'Melee Weapon', 'Ranged Weapon', 'Simple Weapon', 'Martial Weapon'];
    const armorTypes = ['Armor', 'Light Armor', 'Medium Armor', 'Heavy Armor', 'Shield'];
    const wearableTypes = ['Ring', 'Necklace', 'Amulet', 'Cloak', 'Boots', 'Gloves', 'Helmet', 'Hat'];
    const consumableTypes = ['Potion', 'Consumable', 'Food'];

    if (isEquipped) {
        actions.push({ action: 'unequip', label: 'Unequip' });
    } else {
        if (weaponTypes.includes(itemType) || armorTypes.includes(itemType) || wearableTypes.includes(itemType)) {
            actions.push({ action: 'equip', label: 'Equip' });
        }

        if (consumableTypes.includes(itemType)) {
            actions.push({ action: 'use', label: 'Use' });
        }

        actions.push({ action: 'drop', label: 'Drop' });
    }

    actions.push({ action: 'examine', label: 'Examine' });

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
            await performAction(action, itemId, slotIndex, slotType);
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

    // Get current game state
    const npub = getCurrentNpub();
    const saveId = getSaveId();

    if (!npub || !saveId) {
        console.error('No active save');
        showMessage('No active save', 'error');
        return;
    }

    // Prepare request
    const request = {
        npub: npub,
        save_id: saveId,
        item_id: itemId,
        action: action,
        from_slot: fromSlot,
        to_slot: typeof toSlotOrType === 'number' ? toSlotOrType : -1,
        from_slot_type: fromSlotType || '',
        to_slot_type: toSlotType || '',
        from_equip: draggedFromType === 'equipment' ? fromSlot : '',
        to_equip: typeof toSlotOrType === 'string' ? toSlotOrType : '',
        quantity: 1
    };

    try {
        const response = await fetch('/api/inventory/action', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(request)
        });

        const result = await response.json();

        if (result.success) {
            console.log('✅ Action successful:', result.message);
            showMessage(result.message, 'success');

            // Update game state with new inventory
            if (result.newState) {
                const currentState = getGameState();
                currentState.character.inventory = result.newState;
                updateGameState(currentState);
            }
        } else {
            console.error('❌ Action failed:', result.error);
            showMessage(result.error || 'Action failed', 'error');
        }
    } catch (error) {
        console.error('❌ Error performing action:', error);
        showMessage('Failed to perform action', 'error');
    }
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

    // Create modal
    const modal = document.createElement('div');
    modal.className = 'fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50';
    modal.addEventListener('click', (e) => {
        if (e.target === modal) modal.remove();
    });

    const content = document.createElement('div');
    content.className = 'bg-gray-800 border-4 border-gray-600 p-6 max-w-md';
    content.style.clipPath = 'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))';

    content.innerHTML = `
        <h2 class="text-2xl font-bold text-yellow-400 mb-4">${itemData.name}</h2>
        ${itemData.image ? `<img src="${itemData.image}" alt="${itemData.name}" class="w-full mb-4" style="image-rendering: pixelated;">` : ''}
        <p class="text-gray-300 mb-4">${itemData.description || 'No description available.'}</p>
        <div class="grid grid-cols-2 gap-2 text-sm">
            <div><strong class="text-gray-400">Type:</strong> ${itemData.type}</div>
            <div><strong class="text-gray-400">Rarity:</strong> ${itemData.rarity || 'common'}</div>
            <div><strong class="text-gray-400">Weight:</strong> ${itemData.weight || 0} lbs</div>
            <div><strong class="text-gray-400">Value:</strong> ${itemData.price || 0} gp</div>
        </div>
        <button class="mt-4 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white w-full">Close</button>
    `;

    content.querySelector('button').addEventListener('click', () => modal.remove());

    modal.appendChild(content);
    document.body.appendChild(modal);
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
    const defaultAction = getDefaultAction(itemData.type || itemData.item_type, isEquipped);

    // Action display names (simple text)
    const actionNames = {
        'equip': 'Equip',
        'unequip': 'Unequip',
        'use': 'Use',
        'examine': 'Info',
        'drop': 'Drop'
    };

    const actionName = actionNames[defaultAction] || 'Info';

    // Update action text in bottom-right corner
    const actionText = document.getElementById('action-text');
    if (actionText) {
        actionText.textContent = actionName;
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
