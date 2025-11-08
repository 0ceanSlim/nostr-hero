// Game API Client - Go-First Architecture
// All game state lives in Go memory
// JavaScript only sends actions and renders responses

class GameAPI {
    constructor() {
        this.npub = null;
        this.saveID = null;
        this.initialized = false;
    }

    // Initialize the API with user session
    init(npub, saveID) {
        this.npub = npub;
        this.saveID = saveID;
        this.initialized = true;
        console.log('🎮 Game API initialized:', { npub, saveID });
    }

    // Check if initialized
    ensureInitialized() {
        if (!this.initialized) {
            throw new Error('Game API not initialized. Call init() first.');
        }
    }

    // Send a game action to the backend
    async sendAction(actionType, params = {}) {
        this.ensureInitialized();

        console.log(`📤 Sending action: ${actionType}`, params);

        try {
            const response = await fetch('/api/game/action', {
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

            console.log(`✅ Action completed: ${actionType}`, result.message);

            // Return the updated state
            return result;

        } catch (error) {
            console.error(`❌ Action failed: ${actionType}`, error);
            throw error;
        }
    }

    // Fetch current game state from backend
    async getState() {
        this.ensureInitialized();

        try {
            const response = await fetch(`/api/game/state?npub=${this.npub}&save_id=${this.saveID}`);

            if (!response.ok) {
                throw new Error(`Failed to fetch state: ${response.status}`);
            }

            const result = await response.json();

            if (!result.success) {
                throw new Error('Failed to fetch state');
            }

            return result.state;

        } catch (error) {
            console.error('❌ Failed to fetch game state:', error);
            throw error;
        }
    }

    // Save game to disk (manual save)
    async saveGame() {
        this.ensureInitialized();

        console.log('💾 Saving game to disk...');

        try {
            const response = await fetch('/api/session/save', {
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

            console.log('✅ Game saved to disk');
            return true;

        } catch (error) {
            console.error('❌ Save failed:', error);
            throw error;
        }
    }

    // ========================================================================
    // CONVENIENCE METHODS FOR COMMON ACTIONS
    // ========================================================================

    // Move to a new location
    async move(location, district = '', building = '') {
        return await this.sendAction('move', {
            location,
            district,
            building
        });
    }

    // Use a consumable item
    async useItem(itemId, slot = -1) {
        return await this.sendAction('use_item', {
            item_id: itemId,
            slot: slot
        });
    }

    // Equip an item
    async equipItem(itemId, equipmentSlot, fromSlot = -1) {
        return await this.sendAction('equip_item', {
            item_id: itemId,
            equipment_slot: equipmentSlot,
            from_slot: fromSlot
        });
    }

    // Unequip an item
    async unequipItem(equipmentSlot) {
        return await this.sendAction('unequip_item', {
            equipment_slot: equipmentSlot
        });
    }

    // Drop an item
    async dropItem(itemId, slot = -1) {
        return await this.sendAction('drop_item', {
            item_id: itemId,
            slot: slot
        });
    }

    // Pick up an item from ground
    async pickupItem(itemId) {
        return await this.sendAction('pickup_item', {
            item_id: itemId
        });
    }

    // Cast a spell
    async castSpell(spellId, target = null) {
        return await this.sendAction('cast_spell', {
            spell_id: spellId,
            target: target
        });
    }

    // Rest to restore HP/Mana
    async rest() {
        return await this.sendAction('rest', {});
    }

    // Advance game time
    async advanceTime(segments = 1) {
        return await this.sendAction('advance_time', {
            segments: segments
        });
    }

    // Deposit item to vault
    async depositToVault(itemId, quantity = 1) {
        return await this.sendAction('vault_deposit', {
            item_id: itemId,
            quantity: quantity
        });
    }

    // Withdraw item from vault
    async withdrawFromVault(itemId, quantity = 1) {
        return await this.sendAction('vault_withdraw', {
            item_id: itemId,
            quantity: quantity
        });
    }
}

// Create global instance
window.gameAPI = new GameAPI();

console.log('🎮 Game API client loaded');
