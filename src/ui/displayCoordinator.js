/**
 * Display Coordinator UI Module
 *
 * Coordinates updates across all UI modules.
 * Provides main update functions and stub implementations for future features.
 *
 * @module ui/displayCoordinator
 */

import { logger } from '../lib/logger.js';
import { getGameStateSync } from '../state/gameState.js';
import { showMessage } from './messaging.js';
import { updateCharacterDisplay } from './characterDisplay.js';
import { updateSpellsDisplay } from './spellsDisplay.js';
import { displayCurrentLocation } from './locationDisplay.js';
import { updateTimeDisplay } from './timeDisplay.js';

/**
 * Update all UI displays
 * Main coordinator function that refreshes all UI components
 */
export async function updateAllDisplays() {
    await updateCharacterDisplay(); // Async - must await
    updateInventoryDisplay(); // Stub for compatibility
    updateSpellsDisplay();
    displayCurrentLocation();
    updateTimeDisplay();

    const state = getGameStateSync();
    if (state.combat) {
        updateCombatInterface();
    }
}

/**
 * Update inventory display (stub)
 * Inventory is now displayed in updateCharacterDisplay via general_slots and backpack_slots
 * This function is kept for compatibility but does nothing
 */
export function updateInventoryDisplay() {
    // Inventory rendering is handled by updateCharacterDisplay()
    // This function exists for backwards compatibility only
}

/**
 * Update combat interface (stub)
 * Combat system not implemented yet
 */
export function updateCombatInterface() {
    const combatInterface = document.getElementById('combat-interface');
    if (combatInterface) {
        combatInterface.classList.add('hidden');
    }
    // TODO: Implement combat system in Go backend
}

/**
 * Open shop interface (stub)
 * Shop system not fully implemented yet
 */
export function openShop() {
    showMessage('🛍️ Shop system not fully implemented yet', 'info');
    // TODO: Implement shop system
}

/**
 * Open tavern interface (stub)
 * Tavern system not fully implemented yet
 */
export function openTavern() {
    showMessage('🍺 Tavern system not fully implemented yet', 'info');
    // TODO: Implement tavern system
}

/**
 * Save game to Nostr relay (optional feature)
 * Sends game state to a Nostr relay for cloud backup
 */
export async function saveGameToRelay() {
    const state = getGameStateSync();

    // Try to get npub from session manager if available
    let npub = null;
    if (window.sessionManager && window.sessionManager.getCurrentNpub) {
        npub = window.sessionManager.getCurrentNpub();
    }

    if (!npub) {
        showMessage('❌ No user logged in', 'error');
        return;
    }

    const saveData = {
        npub: npub,
        timestamp: Date.now(),
        gameState: state,
        version: "1.0"
    };

    try {
        showMessage('💾 Saving game...', 'info');

        const response = await fetch('/api/save-game', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(saveData)
        });

        if (response.ok) {
            showMessage('✅ Game saved to Nostr relay!', 'success');
            const saveBtn = document.getElementById('save-btn');
            if (saveBtn) {
                saveBtn.textContent = '💾 Saved!';
                setTimeout(() => {
                    saveBtn.textContent = '💾 Save Game';
                }, 2000);
            }
        } else {
            showMessage('❌ Failed to save game', 'error');
        }
    } catch (error) {
        showMessage('❌ Error saving game: ' + error.message, 'error');
    }
}

logger.debug('Display coordinator module loaded');
