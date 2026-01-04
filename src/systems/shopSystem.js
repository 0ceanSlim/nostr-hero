/**
 * Shop System Module
 *
 * Handles shop UI and buy/sell transactions with merchants.
 *
 * @module systems/shopSystem
 */

import { logger } from '../lib/logger.js';
import { gameAPI } from '../lib/api.js';
import { getGameStateSync, refreshGameState } from '../state/gameState.js';
import { showMessage } from '../ui/messaging.js';
import { updateAllDisplays } from '../ui/displayCoordinator.js';
import { sessionManager } from '../lib/session.js';

// Module state
let currentMerchantID = null;
let currentShopData = null;
let currentTab = 'buy';

// Staging state
let buyStaging = [];   // {itemID, quantity, price, name}
let sellStaging = [];  // {itemID, quantity, value, slotIndex, slotType}

/**
 * Open shop interface for a merchant
 * @param {string} merchantID - Merchant NPC ID
 */
export async function openShop(merchantID) {
    logger.debug('Opening shop for merchant:', merchantID);
    currentMerchantID = merchantID;
    currentTab = 'buy';

    // Fetch shop data from backend
    try {
        // Get npub from session
        const npub = sessionManager.getNpub();
        if (!npub) {
            throw new Error('No npub found in session');
        }

        const response = await fetch(`/api/shop/${merchantID}?npub=${encodeURIComponent(npub)}`);
        if (!response.ok) {
            throw new Error('Failed to load shop data');
        }

        currentShopData = await response.json();
        logger.debug('Shop data loaded:', currentShopData);

        // Show modal
        const modal = document.getElementById('shop-modal');
        modal.classList.remove('hidden');

        // Update header
        document.getElementById('shop-merchant-name').textContent = currentShopData.merchant_name || 'Shop';

        // Render buy tab (default)
        switchShopTab('buy');

    } catch (error) {
        logger.error('Error loading shop:', error);
        showMessage('Failed to load shop', 'error');
    }
}

/**
 * Close shop interface
 */
export function closeShop() {
    logger.debug('Closing shop');
    const modal = document.getElementById('shop-modal');
    modal.classList.add('hidden');
    currentMerchantID = null;
    currentShopData = null;
}

/**
 * Open buy quantity modal for custom amount
 * @param {object} shopItem - Shop item data
 */
function openBuyQuantityModal(shopItem) {
    logger.debug('Opening buy quantity modal for:', shopItem.item_id);

    // Create modal backdrop
    const backdrop = document.createElement('div');
    backdrop.className = 'fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-[100]';

    // Create modal
    const modal = document.createElement('div');
    modal.style.cssText = `
        background: #1a1a1a;
        color: white;
        padding: 20px;
        min-width: 300px;
        border-top: 2px solid #4a4a4a;
        border-left: 2px solid #4a4a4a;
        border-right: 2px solid #0a0a0a;
        border-bottom: 2px solid #0a0a0a;
    `;

    // Title
    const title = document.createElement('h3');
    title.textContent = `Buy ${shopItem.name}`;
    title.className = 'text-lg font-bold mb-4 text-green-400';
    modal.appendChild(title);

    // Stock info
    const stockInfo = document.createElement('p');
    stockInfo.textContent = `Available: ${shopItem.stock}`;
    stockInfo.className = 'text-sm text-gray-400 mb-3';
    modal.appendChild(stockInfo);

    // Quantity input
    const inputLabel = document.createElement('label');
    inputLabel.textContent = 'Quantity:';
    inputLabel.className = 'block text-sm mb-1';
    modal.appendChild(inputLabel);

    const input = document.createElement('input');
    input.type = 'number';
    input.min = '1';
    input.max = shopItem.stock.toString();
    input.value = '1';
    input.className = 'w-full px-3 py-2 mb-3 text-base';
    input.style.cssText = `
        background: #0a0a0a;
        color: white;
        border-top: 2px solid #0a0a0a;
        border-left: 2px solid #0a0a0a;
        border-right: 2px solid #3a3a3a;
        border-bottom: 2px solid #3a3a3a;
    `;
    modal.appendChild(input);

    // Total cost preview
    const costPreview = document.createElement('div');
    costPreview.className = 'mb-4 p-2';
    costPreview.style.cssText = `
        background: #0a0a0a;
        border: 1px solid #4a4a4a;
    `;
    const updateCostPreview = () => {
        const qty = Math.min(Math.max(parseInt(input.value) || 1, 1), shopItem.stock);
        const total = qty * shopItem.buy_price;
        costPreview.innerHTML = `
            <div class="text-sm">
                <span class="text-gray-400">${qty}x ${shopItem.name}</span>
            </div>
            <div class="text-lg font-bold text-yellow-400 mt-1">
                Total: ${total}g
            </div>
        `;
    };
    updateCostPreview();
    input.addEventListener('input', updateCostPreview);
    modal.appendChild(costPreview);

    // Buttons container
    const buttonsContainer = document.createElement('div');
    buttonsContainer.className = 'flex gap-2';

    // Confirm button
    const confirmBtn = document.createElement('button');
    confirmBtn.textContent = 'Buy';
    confirmBtn.className = 'flex-1 px-4 py-2 font-medium';
    confirmBtn.style.cssText = `
        background: #15803d;
        color: white;
        border-top: 2px solid #22c55e;
        border-left: 2px solid #22c55e;
        border-right: 2px solid #166534;
        border-bottom: 2px solid #166534;
    `;
    confirmBtn.onclick = async () => {
        const qty = Math.min(Math.max(parseInt(input.value) || 1, 1), shopItem.stock);
        document.body.removeChild(backdrop);
        await buyItemNow(shopItem.item_id, shopItem.name, qty, shopItem.buy_price, shopItem.stock);
    };
    buttonsContainer.appendChild(confirmBtn);

    // Cancel button
    const cancelBtn = document.createElement('button');
    cancelBtn.textContent = 'Cancel';
    cancelBtn.className = 'flex-1 px-4 py-2 font-medium';
    cancelBtn.style.cssText = `
        background: #7f1d1d;
        color: white;
        border-top: 2px solid #dc2626;
        border-left: 2px solid #dc2626;
        border-right: 2px solid #991b1b;
        border-bottom: 2px solid #991b1b;
    `;
    cancelBtn.onclick = () => {
        document.body.removeChild(backdrop);
    };
    buttonsContainer.appendChild(cancelBtn);

    modal.appendChild(buttonsContainer);
    backdrop.appendChild(modal);
    document.body.appendChild(backdrop);

    // Focus input
    input.focus();
    input.select();

    // Enter key to confirm
    input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            confirmBtn.click();
        } else if (e.key === 'Escape') {
            cancelBtn.click();
        }
    });
}

/**
 * Switch between buy and sell tabs
 * @param {string} tab - 'buy' or 'sell'
 */
export function switchShopTab(tab) {
    logger.debug('Switching to shop tab:', tab);
    currentTab = tab;

    // Update tab button styles
    const buyTab = document.getElementById('shop-buy-tab');
    const sellTab = document.getElementById('shop-sell-tab');
    const buyContent = document.getElementById('shop-buy-content');
    const sellContent = document.getElementById('shop-sell-content');

    if (tab === 'buy') {
        // Active buy tab
        buyTab.style.background = '#15803d';
        buyTab.style.borderTop = '2px solid #22c55e';
        buyTab.style.borderLeft = '2px solid #22c55e';
        buyTab.style.borderRight = '2px solid #166534';
        buyTab.style.borderBottom = '2px solid #166534';

        // Inactive sell tab
        sellTab.style.background = '#2a2a2a';
        sellTab.style.borderTop = '2px solid #4a4a4a';
        sellTab.style.borderLeft = '2px solid #4a4a4a';
        sellTab.style.borderRight = '2px solid #1a1a1a';
        sellTab.style.borderBottom = '2px solid #1a1a1a';

        buyContent.classList.remove('hidden');
        sellContent.classList.add('hidden');

        renderBuyTab();
    } else {
        // Active sell tab
        sellTab.style.background = '#15803d';
        sellTab.style.borderTop = '2px solid #22c55e';
        sellTab.style.borderLeft = '2px solid #22c55e';
        sellTab.style.borderRight = '2px solid #166534';
        sellTab.style.borderBottom = '2px solid #166534';

        // Inactive buy tab
        buyTab.style.background = '#2a2a2a';
        buyTab.style.borderTop = '2px solid #4a4a4a';
        buyTab.style.borderLeft = '2px solid #4a4a4a';
        buyTab.style.borderRight = '2px solid #1a1a1a';
        buyTab.style.borderBottom = '2px solid #1a1a1a';

        sellContent.classList.remove('hidden');
        buyContent.classList.add('hidden');

        renderSellTab();
    }
}

/**
 * Get total gold from inventory (counts "gold-piece" items in slots)
 */
function getPlayerGold(state) {
    let totalGold = 0;

    // Debug: Log the entire state to see structure
    console.log('🪙 DEBUG getPlayerGold - Full state:', state);

    // State structure is: state.inventory = array (general slots), state.equipment = object (gear slots)
    const generalSlots = Array.isArray(state.inventory) ? state.inventory : [];
    const equipment = state.equipment || {};

    console.log('🪙 DEBUG General slots (state.inventory):', generalSlots);
    console.log('🪙 DEBUG Equipment (state.equipment):', equipment);

    // Check general slots (state.inventory is the array directly)
    for (const slot of generalSlots) {
        console.log('🪙 Checking general slot:', slot);
        if (slot && slot.item === 'gold-piece' && slot.quantity) {
            const amount = parseInt(slot.quantity) || 0;
            console.log('🪙 Found gold in general slot:', amount);
            totalGold += amount;
        }
    }

    // Check equipment/bag
    if (equipment.bag && equipment.bag.contents) {
        const contents = Array.isArray(equipment.bag.contents) ? equipment.bag.contents : [];
        console.log('🪙 DEBUG Bag contents:', contents);
        for (const slot of contents) {
            console.log('🪙 Checking bag slot:', slot);
            if (slot && slot.item === 'gold-piece' && slot.quantity) {
                const amount = parseInt(slot.quantity) || 0;
                console.log('🪙 Found gold in bag:', amount);
                totalGold += amount;
            }
        }
    }

    // Check if gold is the bag item itself
    if (equipment.bag && equipment.bag.item === 'gold-piece' && equipment.bag.quantity) {
        const amount = parseInt(equipment.bag.quantity) || 0;
        console.log('🪙 Found gold as bag item:', amount);
        totalGold += amount;
    }

    // Check all equipment slots in case gold is in one
    for (const slotName in equipment) {
        const slot = equipment[slotName];
        if (slot && typeof slot === 'object' && slot.item === 'gold-piece' && slot.quantity) {
            const amount = parseInt(slot.quantity) || 0;
            console.log('🪙 Found gold in equipment slot', slotName, ':', amount);
            totalGold += amount;
        }
    }

    console.log('🪙 FINAL Total gold calculated:', totalGold);
    return totalGold;
}

/**
 * Render buy tab (merchant's inventory) - Grid Layout
 */
function renderBuyTab() {
    const state = getGameStateSync();
    const playerGold = getPlayerGold(state);
    const merchantGold = currentShopData.current_gold || 0;

    // Update gold displays
    document.getElementById('shop-player-gold').textContent = playerGold;
    document.getElementById('shop-merchant-gold').textContent = merchantGold;

    const grid = document.getElementById('shop-item-grid');
    grid.innerHTML = '';

    if (!currentShopData.inventory || currentShopData.inventory.length === 0) {
        grid.innerHTML = '<p style="color: #9ca3af;">This merchant has no items for sale.</p>';
        return;
    }

    // Render each item as grid slot (5 columns, like container system)
    currentShopData.inventory.forEach(shopItem => {
        // Create slot
        const slot = document.createElement('div');
        slot.className = 'relative cursor-pointer hover:bg-gray-800 transition-colors';
        slot.style.cssText = `
            aspect-ratio: 1;
            background: #1a1a1a;
            border-top: 2px solid #3a3a3a;
            border-left: 2px solid #3a3a3a;
            border-right: 2px solid #0a0a0a;
            border-bottom: 2px solid #0a0a0a;
        `;

        // Item image
        const img = document.createElement('img');
        img.src = `/res/img/items/${shopItem.item_id}.png`;
        img.className = 'w-full h-full object-contain p-1';
        img.style.imageRendering = 'pixelated';
        img.onerror = () => { img.src = '/res/img/items/default.png'; };
        slot.appendChild(img);

        // Stock quantity (bottom-right)
        const stockBadge = document.createElement('div');
        stockBadge.className = 'absolute bottom-0 right-0 px-1 text-white font-bold';
        stockBadge.style.cssText = `
            font-size: 9px;
            background: rgba(0, 0, 0, 0.7);
            text-shadow: 1px 1px 2px #000;
        `;
        stockBadge.textContent = shopItem.stock;
        slot.appendChild(stockBadge);

        // Left-click: Show value in message screen
        slot.addEventListener('click', (e) => {
            e.stopPropagation();
            showItemValue(shopItem);
        });

        // Right-click: Context menu
        slot.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            e.stopPropagation();
            showBuyContextMenu(e, shopItem);
        });

        grid.appendChild(slot);
    });
}

/**
 * Show item value in message screen
 */
function showItemValue(shopItem) {
    const message = `${shopItem.name}: ${shopItem.buy_price}g`;
    showMessage(message, 'info');
}

/**
 * Show context menu for buying items
 */
function showBuyContextMenu(e, shopItem) {
    closeContextMenu();

    const menu = document.createElement('div');
    menu.id = 'shop-context-menu';
    menu.className = 'absolute z-50';
    menu.style.cssText = `
        left: ${e.pageX}px;
        top: ${e.pageY}px;
        background: #2a2a2a;
        border: 2px solid #4a4a4a;
        min-width: 150px;
    `;

    // Value button
    const valueBtn = createMenuButton('Value', () => {
        showItemValue(shopItem);
        closeContextMenu();
    });
    menu.appendChild(valueBtn);

    // Buy 1 button
    const buy1Btn = createMenuButton('Buy 1', async () => {
        closeContextMenu();
        await buyItemNow(shopItem.item_id, shopItem.name, 1, shopItem.buy_price, shopItem.stock);
    });
    menu.appendChild(buy1Btn);

    // Buy X button (opens modal)
    const buyXBtn = createMenuButton('Buy X', () => {
        closeContextMenu();
        openBuyQuantityModal(shopItem);
    });
    menu.appendChild(buyXBtn);

    document.body.appendChild(menu);

    // Click outside to close
    setTimeout(() => {
        document.addEventListener('click', handleContextMenuClose, { once: true });
    }, 10);
}

/**
 * Create a context menu button
 */
function createMenuButton(text, onClick) {
    const btn = document.createElement('button');
    btn.textContent = text;
    btn.className = 'w-full px-3 py-2 text-left text-sm hover:bg-gray-700';
    btn.style.color = '#fbbf24';
    btn.onclick = onClick;
    return btn;
}

/**
 * Close context menu
 */
function closeContextMenu() {
    const existing = document.getElementById('shop-context-menu');
    if (existing) existing.remove();
}

/**
 * Handle context menu close event
 */
function handleContextMenuClose() {
    closeContextMenu();
}

/**
 * Buy item immediately (no staging)
 */
async function buyItemNow(itemID, name, quantity, price, maxStock) {
    console.log('💰 buyItemNow called:', { itemID, name, quantity, price, maxStock });

    // Validate stock
    if (quantity > maxStock) {
        showMessage('Not enough stock', 'error');
        return;
    }

    // Validate player gold
    const state = getGameStateSync();
    const playerGold = getPlayerGold(state);
    const totalCost = price * quantity;

    if (playerGold < totalCost) {
        showMessage('Not enough gold', 'error');
        return;
    }

    // Get session info from gameAPI (already initialized at game start)
    const npub = gameAPI.npub;
    const saveID = gameAPI.saveID;

    console.log('💰 Session info from gameAPI:', { npub, saveID });

    if (!npub || !saveID) {
        showMessage('Session error', 'error');
        console.error('❌ GameAPI not initialized properly:', { npub, saveID });
        return;
    }

    try {
        const response = await fetch('/api/shop/buy', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                npub: npub,
                save_id: saveID,
                merchant_id: currentMerchantID,
                item_id: itemID,
                quantity: quantity,
                action: 'buy'
            })
        });

        const result = await response.json();
        console.log('💰 Buy response:', result);

        if (!response.ok || !result.success) {
            showMessage(result.error || result.message || 'Purchase failed', 'error');
            return;
        }

        showMessage(`Bought ${quantity}x ${name}!`, 'success');

        // Refresh shop and game state
        await refreshGameState();
        await openShop(currentMerchantID);
        updateAllDisplays();
    } catch (error) {
        logger.error('Buy transaction error:', error);
        console.error('❌ Buy error:', error);
        showMessage('Failed to purchase items', 'error');
    }
}

/**
 * Render buy staging area
 */
function renderBuyStaging() {
    const container = document.getElementById('buy-staged-items');
    container.innerHTML = '';

    let totalCost = 0;
    buyStaging.forEach((item, index) => {
        totalCost += item.price * item.quantity;

        // Create staged item badge
        const badge = document.createElement('div');
        badge.className = 'relative';
        badge.style.cssText = `
            width: 36px;
            height: 36px;
            background: #1a1a1a;
            border: 1px solid #4a4a4a;
        `;

        const img = document.createElement('img');
        img.src = `/res/img/items/${item.itemID}.png`;
        img.className = 'w-full h-full object-contain';
        img.style.imageRendering = 'pixelated';
        img.onerror = () => { img.src = '/res/img/items/default.png'; };
        badge.appendChild(img);

        // Quantity
        const qty = document.createElement('div');
        qty.className = 'absolute bottom-0 right-0 text-white font-bold';
        qty.style.cssText = 'font-size: 8px; text-shadow: 1px 1px 2px #000;';
        qty.textContent = item.quantity;
        badge.appendChild(qty);

        // Remove button
        const removeBtn = document.createElement('button');
        removeBtn.className = 'absolute -top-1 -right-1 text-white';
        removeBtn.style.cssText = `
            width: 14px;
            height: 14px;
            background: #dc2626;
            font-size: 10px;
            line-height: 1;
        `;
        removeBtn.textContent = '✕';
        removeBtn.onclick = () => removeStagedBuy(index);
        badge.appendChild(removeBtn);

        container.appendChild(badge);
    });

    document.getElementById('buy-total-cost').textContent = totalCost;
}

/**
 * Show staging area
 */
function showStagingArea(type) {
    const buyStaging = document.getElementById('buy-staging');
    const sellStaging = document.getElementById('sell-staging');

    console.log('🛒 showStagingArea called for type:', type);
    console.log('🛒 Staging elements:', { buyStaging, sellStaging });

    if (type === 'buy') {
        buyStaging.classList.remove('hidden');
        console.log('🛒 Buy staging shown');
    } else {
        sellStaging.classList.remove('hidden');
        console.log('🛒 Sell staging shown');
    }
}

/**
 * Remove staged buy item
 */
function removeStagedBuy(index) {
    buyStaging.splice(index, 1);
    renderBuyStaging();
    if (buyStaging.length === 0) {
        document.getElementById('buy-staging').classList.add('hidden');
    }
}

/**
 * Clear buy staging
 */
function clearBuyStaging() {
    buyStaging = [];
    document.getElementById('buy-staging').classList.add('hidden');
}

/**
 * Confirm buy transaction
 */
async function confirmBuyTransaction() {
    if (buyStaging.length === 0) return;

    const state = getGameStateSync();
    const saveID = state.save_id || state.character?.save_id;
    const npub = sessionStorage.getItem('npub');

    if (!saveID || !npub) {
        showMessage('Session error', 'error');
        return;
    }

    // Process each staged item
    for (const item of buyStaging) {
        try {
            const response = await fetch('/api/shop/buy', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    npub: npub,
                    save_id: saveID,
                    merchant_id: currentMerchantID,
                    item_id: item.itemID,
                    quantity: item.quantity,
                    action: 'buy'
                })
            });

            const result = await response.json();

            if (!response.ok || !result.success) {
                showMessage(result.error || result.message || 'Purchase failed', 'error');
                break;  // Stop on first error
            }
        } catch (error) {
            logger.error('Buy transaction error:', error);
            showMessage('Failed to purchase items', 'error');
            break;
        }
    }

    // Clear staging and refresh
    buyStaging = [];
    clearBuyStaging();
    await refreshGameState();
    await openShop(currentMerchantID);  // Refresh shop data
    updateAllDisplays();
    showMessage('Purchase complete!', 'success');
}

/**
 * Render sell tab (player's inventory) - Grid Layout
 */
function renderSellTab() {
    const state = getGameStateSync();
    const playerGold = getPlayerGold(state);
    const merchantGold = currentShopData.current_gold || 0;

    // Update gold displays
    document.getElementById('shop-player-gold').textContent = playerGold;
    document.getElementById('shop-merchant-gold').textContent = merchantGold;

    const grid = document.getElementById('shop-sell-grid');
    grid.innerHTML = '';

    if (!currentShopData.buys_items) {
        grid.innerHTML = '<p style="color: #9ca3af;">This merchant doesn\'t buy items.</p>';
        return;
    }

    // Use correct state structure: state.inventory = array, state.equipment = object
    const generalSlots = Array.isArray(state.inventory) ? state.inventory : [];
    const equipment = state.equipment || {};
    const backpack = equipment.bag?.contents || [];

    console.log('🛒 Rendering sell tab - general slots:', generalSlots);
    console.log('🛒 Rendering sell tab - backpack:', backpack);

    // Render general slots (skip gold-piece)
    generalSlots.forEach((slot, index) => {
        if (!slot || !slot.item || slot.item === 'gold-piece') return;
        renderSellSlot(grid, slot, index, 'general');
    });

    // Render backpack slots (skip gold-piece)
    backpack.forEach((slot, index) => {
        if (!slot || !slot.item || slot.item === 'gold-piece') return;
        renderSellSlot(grid, slot, index, 'backpack');
    });
}

/**
 * Render a single sell slot
 */
function renderSellSlot(grid, slot, slotIndex, slotType) {
    console.log('🛒 renderSellSlot called for:', slot.item, 'at index', slotIndex);

    // Get item data from static data using imported function
    const itemData = getItemById(slot.item);
    if (!itemData) {
        console.log('🛒 No item data found for:', slot.item);
        return;
    }

    console.log('🛒 Item data found:', itemData);

    // Calculate sell value
    const sellValue = Math.floor(itemData.value * currentShopData.buy_price_multiplier);
    console.log('🛒 Sell value calculated:', sellValue, '(base:', itemData.value, 'x', currentShopData.buy_price_multiplier, ')');

    // Create slot
    const slotDiv = document.createElement('div');
    slotDiv.className = 'relative cursor-pointer hover:bg-gray-700 transition-colors';
    slotDiv.style.cssText = `
        aspect-ratio: 1;
        background: #1a1a1a;
        border-top: 2px solid #3a3a3a;
        border-left: 2px solid #3a3a3a;
        border-right: 2px solid #0a0a0a;
        border-bottom: 2px solid #0a0a0a;
    `;

    // Item image
    const img = document.createElement('img');
    img.src = `/res/img/items/${slot.item}.png`;
    img.className = 'w-full h-full object-contain p-1';
    img.style.imageRendering = 'pixelated';
    img.onerror = () => { img.src = '/res/img/items/default.png'; };
    slotDiv.appendChild(img);

    // Quantity (if > 1)
    if (slot.quantity > 1) {
        const qtyBadge = document.createElement('div');
        qtyBadge.className = 'absolute bottom-0 right-0 px-1 text-white font-bold';
        qtyBadge.style.cssText = `
            font-size: 9px;
            background: rgba(0, 0, 0, 0.7);
            text-shadow: 1px 1px 2px #000;
        `;
        qtyBadge.textContent = slot.quantity;
        slotDiv.appendChild(qtyBadge);
    }

    // Left-click: add to staging
    console.log('🛒 Adding click listener to slot for', slot.item);
    slotDiv.addEventListener('click', (e) => {
        console.log('🛒 CLICK EVENT FIRED for:', slot.item);
        e.preventDefault();
        e.stopPropagation();
        addToSellStaging(slot.item, itemData.name, 1, sellValue, slotIndex, slotType);
    });

    // Also try onclick as a fallback
    slotDiv.onclick = (e) => {
        console.log('🛒 ONCLICK EVENT FIRED for:', slot.item);
        e.preventDefault();
        e.stopPropagation();
        addToSellStaging(slot.item, itemData.name, 1, sellValue, slotIndex, slotType);
    };

    console.log('🛒 Appending slot to grid');
    grid.appendChild(slotDiv);
}

/**
 * Add item to sell staging
 */
function addToSellStaging(itemID, name, quantity, value, slotIndex, slotType) {
    console.log('🛒 addToSellStaging called:', { itemID, name, quantity, value, slotIndex, slotType });

    // Validate merchant has gold
    const totalValue = value * quantity;
    const merchantGold = currentShopData.current_gold || currentShopData.starting_gold || 0;

    console.log('🛒 Merchant gold check:', merchantGold, 'vs needed:', totalValue);

    if (merchantGold < totalValue) {
        showMessage("Merchant doesn't have enough gold", 'error');
        return;
    }

    sellStaging.push({ itemID, name, quantity, value, slotIndex, slotType });
    console.log('🛒 Sell staging now has', sellStaging.length, 'items');
    renderSellStaging();
    showStagingArea('sell');
}

/**
 * Render sell staging area
 */
function renderSellStaging() {
    const container = document.getElementById('sell-staged-items');
    container.innerHTML = '';

    let totalValue = 0;
    sellStaging.forEach((item, index) => {
        totalValue += item.value * item.quantity;

        // Create staged item badge (same as buy staging)
        const badge = document.createElement('div');
        badge.className = 'relative';
        badge.style.cssText = `
            width: 36px;
            height: 36px;
            background: #1a1a1a;
            border: 1px solid #4a4a4a;
        `;

        const img = document.createElement('img');
        img.src = `/res/img/items/${item.itemID}.png`;
        img.className = 'w-full h-full object-contain';
        img.style.imageRendering = 'pixelated';
        img.onerror = () => { img.src = '/res/img/items/default.png'; };
        badge.appendChild(img);

        // Quantity
        const qty = document.createElement('div');
        qty.className = 'absolute bottom-0 right-0 text-white font-bold';
        qty.style.cssText = 'font-size: 8px; text-shadow: 1px 1px 2px #000;';
        qty.textContent = item.quantity;
        badge.appendChild(qty);

        // Remove button
        const removeBtn = document.createElement('button');
        removeBtn.className = 'absolute -top-1 -right-1 text-white';
        removeBtn.style.cssText = `
            width: 14px;
            height: 14px;
            background: #dc2626;
            font-size: 10px;
            line-height: 1;
        `;
        removeBtn.textContent = '✕';
        removeBtn.onclick = () => removeStagedSell(index);
        badge.appendChild(removeBtn);

        container.appendChild(badge);
    });

    document.getElementById('sell-total-value').textContent = totalValue;
}

/**
 * Remove staged sell item
 */
function removeStagedSell(index) {
    sellStaging.splice(index, 1);
    renderSellStaging();
    if (sellStaging.length === 0) {
        document.getElementById('sell-staging').classList.add('hidden');
    }
}

/**
 * Clear sell staging
 */
function clearSellStaging() {
    sellStaging = [];
    document.getElementById('sell-staging').classList.add('hidden');
}

/**
 * Confirm sell transaction
 */
async function confirmSellTransaction() {
    if (sellStaging.length === 0) return;

    // Get session info from gameAPI (already initialized at game start)
    const npub = gameAPI.npub;
    const saveID = gameAPI.saveID;

    if (!npub || !saveID) {
        showMessage('Session error', 'error');
        console.error('❌ GameAPI not initialized properly:', { npub, saveID });
        return;
    }

    // Process each staged item
    for (const item of sellStaging) {
        try {
            const response = await fetch('/api/shop/sell', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    npub: npub,
                    save_id: saveID,
                    merchant_id: currentMerchantID,
                    item_id: item.itemID,
                    quantity: item.quantity,
                    action: 'sell'
                })
            });

            const result = await response.json();

            if (!response.ok || !result.success) {
                showMessage(result.error || result.message || 'Sale failed', 'error');
                break;
            }
        } catch (error) {
            logger.error('Sell transaction error:', error);
            showMessage('Failed to sell items', 'error');
            break;
        }
    }

    // Clear staging and refresh
    sellStaging = [];
    clearSellStaging();
    await refreshGameState();
    await openShop(currentMerchantID);
    updateAllDisplays();
    showMessage('Sale complete!', 'success');
}

// Expose functions to window for onclick handlers
window.openShop = openShop;
window.closeShop = closeShop;
window.switchShopTab = switchShopTab;
window.confirmBuyTransaction = confirmBuyTransaction;
window.clearBuyStaging = clearBuyStaging;
window.confirmSellTransaction = confirmSellTransaction;
window.clearSellStaging = clearSellStaging;
