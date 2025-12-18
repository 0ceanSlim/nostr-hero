/**
 * Game API Client - Go-First Architecture
 *
 * All game state lives in Go memory.
 * JavaScript only sends actions and renders responses.
 *
 * @module lib/api
 */

import { logger } from './logger.js';
import { API_BASE_URL } from '../config/constants.js';

class GameAPI {
    constructor() {
        this.npub = null;
        this.saveID = null;
        this.initialized = false;
    }

    /**
     * Initialize the API with user session
     * @param {string} npub - User's Nostr public key
     * @param {string} saveID - Save file ID
     */
    init(npub, saveID) {
        this.npub = npub;
        this.saveID = saveID;
        this.initialized = true;
        logger.info('Game API initialized:', { npub, saveID });
    }

    /**
     * Check if initialized
     * @throws {Error} If API not initialized
     */
    ensureInitialized() {
        if (!this.initialized) {
            throw new Error('Game API not initialized. Call init() first.');
        }
    }

    /**
     * Send a game action to the backend
     * @param {string} actionType - Type of action
     * @param {Object} params - Action parameters
     * @returns {Promise<Object>} Action result with updated state
     */
    async sendAction(actionType, params = {}) {
        this.ensureInitialized();

        logger.debug(`Sending action: ${actionType}`, params);

        try {
            const response = await fetch(`${API_BASE_URL}/game/action`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    npub: this.npub,
                    save_id: this.saveID,
                    action: {
                        type: actionType,
                        params: params
                    }
                })
            });

            if (!response.ok) {
                throw new Error(`Action failed: ${response.status}`);
            }

            const result = await response.json();

            if (!result.success) {
                throw new Error(result.error || 'Action failed');
            }

            logger.info(`Action completed: ${actionType}`, result.message);

            // Return the updated state
            return result;

        } catch (error) {
            logger.error(`Action failed: ${actionType}`, error);
            throw error;
        }
    }

    /**
     * Fetch current game state from backend
     * @returns {Promise<Object>} Current game state
     */
    async getState() {
        this.ensureInitialized();

        try {
            const response = await fetch(
                `${API_BASE_URL}/game/state?npub=${this.npub}&save_id=${this.saveID}`
            );

            if (!response.ok) {
                throw new Error(`Failed to fetch state: ${response.status}`);
            }

            const result = await response.json();

            if (!result.success) {
                throw new Error('Failed to fetch state');
            }

            return result.state;

        } catch (error) {
            logger.error('Failed to fetch game state:', error);
            throw error;
        }
    }

    /**
     * Save game to disk (manual save)
     * @returns {Promise<boolean>} Success status
     */
    async saveGame() {
        this.ensureInitialized();

        logger.info('Saving game to disk...');

        try {
            const response = await fetch(`${API_BASE_URL}/session/save`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    npub: this.npub,
                    save_id: this.saveID
                })
            });

            if (!response.ok) {
                throw new Error(`Save failed: ${response.status}`);
            }

            const result = await response.json();

            if (!result.success) {
                throw new Error(result.message || 'Save failed');
            }

            logger.info('Game saved to disk');
            return true;

        } catch (error) {
            logger.error('Save failed:', error);
            throw error;
        }
    }

    // ========================================================================
    // CONVENIENCE METHODS FOR COMMON ACTIONS
    // ========================================================================

    /**
     * Move to a new location
     * @param {string} location - Location ID
     * @param {string} district - District name (optional)
     * @param {string} building - Building name (optional)
     */
    async move(location, district = '', building = '') {
        return await this.sendAction('move', {
            location,
            district,
            building
        });
    }

    /**
     * Use a consumable item
     * @param {string} itemId - Item ID
     * @param {number} slot - Inventory slot (-1 for equipment)
     */
    async useItem(itemId, slot = -1) {
        return await this.sendAction('use_item', {
            item_id: itemId,
            slot: slot
        });
    }

    /**
     * Equip an item
     * @param {string} itemId - Item ID
     * @param {string} equipmentSlot - Equipment slot name
     * @param {number} fromSlot - Source inventory slot
     */
    async equipItem(itemId, equipmentSlot, fromSlot = -1) {
        return await this.sendAction('equip_item', {
            item_id: itemId,
            equipment_slot: equipmentSlot,
            from_slot: fromSlot
        });
    }

    /**
     * Unequip an item
     * @param {string} equipmentSlot - Equipment slot name
     */
    async unequipItem(equipmentSlot) {
        return await this.sendAction('unequip_item', {
            equipment_slot: equipmentSlot
        });
    }

    /**
     * Drop an item
     * @param {string} itemId - Item ID
     * @param {number} slot - Inventory slot
     */
    async dropItem(itemId, slot = -1) {
        return await this.sendAction('drop_item', {
            item_id: itemId,
            slot: slot
        });
    }

    /**
     * Pick up an item from ground
     * @param {string} itemId - Item ID
     */
    async pickupItem(itemId) {
        return await this.sendAction('pickup_item', {
            item_id: itemId
        });
    }

    /**
     * Cast a spell
     * @param {string} spellId - Spell ID
     * @param {Object} target - Target (optional)
     */
    async castSpell(spellId, target = null) {
        return await this.sendAction('cast_spell', {
            spell_id: spellId,
            target: target
        });
    }

    /**
     * Rest to restore HP/Mana
     */
    async rest() {
        return await this.sendAction('rest', {});
    }

    /**
     * Advance game time
     * @param {number} segments - Number of time segments
     */
    async advanceTime(segments = 1) {
        return await this.sendAction('advance_time', {
            segments: segments
        });
    }

    /**
     * Deposit item to vault
     * @param {string} itemId - Item ID
     * @param {number} quantity - Quantity to deposit
     */
    async depositToVault(itemId, quantity = 1) {
        return await this.sendAction('vault_deposit', {
            item_id: itemId,
            quantity: quantity
        });
    }

    /**
     * Withdraw item from vault
     * @param {string} itemId - Item ID
     * @param {number} quantity - Quantity to withdraw
     */
    async withdrawFromVault(itemId, quantity = 1) {
        return await this.sendAction('vault_withdraw', {
            item_id: itemId,
            quantity: quantity
        });
    }
}

// Export singleton instance
export const gameAPI = new GameAPI();

logger.debug('Game API client loaded');
