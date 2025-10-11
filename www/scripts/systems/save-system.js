// Save System for Nostr Hero
// Handles saving game state to local JSON files

// Load save data when game starts
async function loadSaveData() {
    const urlParams = new URLSearchParams(window.location.search);
    const saveID = urlParams.get('save');

    if (!saveID) {
        console.log('No save ID in URL, using default empty state');
        return;
    }

    if (!window.sessionManager || !window.sessionManager.isAuthenticated()) {
        console.log('Not authenticated, cannot load save');
        return;
    }

    try {
        console.log('🔄 Loading save data for:', saveID);
        const session = window.sessionManager.getSession();

        const response = await fetch(`/api/saves/${session.npub}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch saves: ${response.status}`);
        }

        const saves = await response.json();
        const saveData = saves.find(save => save.id === saveID);

        if (!saveData) {
            console.warn('⚠️ Save not found:', saveID);
            return;
        }

        console.log('✅ Loaded save data:', saveData);

        // Update DOM with loaded data (new structure)
        if (saveData.character) {
            document.getElementById('character-data').textContent = JSON.stringify(saveData.character);
        }

        // Location from new structure
        if (saveData.location) {
            document.getElementById('location-data').textContent = JSON.stringify(saveData.location);
        }

        // Inventory is now inside character object
        if (saveData.character && saveData.character.inventory) {
            document.getElementById('inventory-data').textContent = JSON.stringify(saveData.character.inventory);
        }

        // Spells are now inside character object
        if (saveData.character && saveData.character.spells) {
            document.getElementById('spell-data').textContent = JSON.stringify(saveData.character.spells);
        }

        // Equipment is now inside character.inventory.gear_slots
        if (saveData.character && saveData.character.inventory && saveData.character.inventory.gear_slots) {
            document.getElementById('equipment-data').textContent = JSON.stringify(saveData.character.inventory.gear_slots);
        }

        // Combat data (default to null if not present)
        document.getElementById('combat-data').textContent = 'null';

        console.log('✅ Save data loaded into DOM successfully');

        // Trigger UI update
        if (typeof updateCharacterDisplay === 'function') {
            updateCharacterDisplay();
        }

        // Mark save data as loaded to enable auto-saves
        saveDataLoaded = true;

    } catch (error) {
        console.error('❌ Failed to load save data:', error);
        // Even if loading fails, allow saves after delay
        setTimeout(() => {
            saveDataLoaded = true;
        }, 5000);
    }
}

// Load save data when the page loads
document.addEventListener('DOMContentLoaded', async () => {
    console.log('🎮 Save system: DOM loaded, attempting to load save data...');
    await loadSaveData();
});

// Save game to local JSON file
async function saveGameToLocal() {
    if (!window.sessionManager || !window.sessionManager.isAuthenticated()) {
        showMessage('❌ Must be logged in to save', 'error');
        return false;
    }

    try {
        showMessage('💾 Saving game...', 'info');

        const session = window.sessionManager.getSession();
        const gameState = getGameState();

        // Prepare save data
        const saveData = {
            character: gameState.character,
            gameState: gameState,
            location: gameState.location.current,
        };

        // Get current save ID from URL if loading existing save
        const urlParams = new URLSearchParams(window.location.search);
        const currentSaveID = urlParams.get('save');

        if (currentSaveID) {
            saveData.id = currentSaveID;
        }

        // Send save request
        const response = await fetch(`/api/saves/${session.npub}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(saveData)
        });

        if (!response.ok) {
            throw new Error(`Save failed: ${response.status}`);
        }

        const result = await response.json();

        if (result.success) {
            showMessage('✅ Game saved successfully!', 'success');

            // Update URL with save ID if this was a new save
            if (!currentSaveID && result.save_id) {
                const newUrl = new URL(window.location);
                newUrl.searchParams.set('save', result.save_id);
                newUrl.searchParams.delete('new');
                window.history.replaceState({}, '', newUrl);
            }

            return true;
        } else {
            throw new Error(result.message || 'Save failed');
        }

    } catch (error) {
        console.error('❌ Save failed:', error);
        showMessage('❌ Failed to save game: ' + error.message, 'error');
        return false;
    }
}

// Auto-save functionality
let autoSaveInterval = null;
const AUTO_SAVE_INTERVAL = 5 * 60 * 1000; // 5 minutes

function startAutoSave() {
    if (autoSaveInterval) {
        clearInterval(autoSaveInterval);
    }

    // Delay auto-save start to allow save data to load first
    setTimeout(() => {
        autoSaveInterval = setInterval(async () => {
            if (saveDataLoaded) {
                console.log('🔄 Auto-saving game...');
                const success = await saveGameToLocal();
                if (success) {
                    console.log('✅ Auto-save complete');
                }
            }
        }, AUTO_SAVE_INTERVAL);

        console.log('⏰ Auto-save enabled (every 5 minutes)');
    }, 3000); // Reduced delay since we're using saveDataLoaded flag
}

function stopAutoSave() {
    if (autoSaveInterval) {
        clearInterval(autoSaveInterval);
        autoSaveInterval = null;
        console.log('⏹️ Auto-save disabled');
    }
}

// Track if save data has been loaded to prevent early auto-saves
let saveDataLoaded = false;

// Save before leaving page
window.addEventListener('beforeunload', async (event) => {
    if (saveDataLoaded && window.sessionManager && window.sessionManager.isAuthenticated()) {
        // Attempt quick save (though it may not complete due to page unload)
        saveGameToLocal();
    }
});

// Save on page visibility change (when tab becomes hidden)
document.addEventListener('visibilitychange', async () => {
    if (saveDataLoaded && document.hidden && window.sessionManager && window.sessionManager.isAuthenticated()) {
        console.log('👁️ Page hidden, saving game...');
        await saveGameToLocal();
    }
});

// Export save functionality
async function exportSave() {
    if (!window.sessionManager || !window.sessionManager.isAuthenticated()) {
        showMessage('❌ Must be logged in to export', 'error');
        return;
    }

    try {
        const session = window.sessionManager.getSession();
        const gameState = getGameState();

        const exportData = {
            version: "1.0",
            npub: session.npub,
            exported_at: new Date().toISOString(),
            character: gameState.character,
            gameState: gameState
        };

        const dataStr = JSON.stringify(exportData, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });

        const link = document.createElement('a');
        link.href = URL.createObjectURL(dataBlob);
        link.download = `nostr-hero-save-${new Date().toISOString().split('T')[0]}.json`;
        link.click();

        showMessage('✅ Save file exported!', 'success');

    } catch (error) {
        console.error('❌ Export failed:', error);
        showMessage('❌ Failed to export save: ' + error.message, 'error');
    }
}

// Quick save hotkey (Ctrl+S)
document.addEventListener('keydown', (event) => {
    if (event.ctrlKey && event.key === 's') {
        event.preventDefault();
        saveGameToLocal();
    }
});

console.log('💾 Save system loaded');