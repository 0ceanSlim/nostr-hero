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

        // Convert display names to IDs for game logic
        console.log('🔄 Converting location display names to IDs...');
        const locationIds = await window.getIdsFromDisplayNames(
            saveData.location || 'The Royal Kingdom',
            saveData.district || 'Kingdom Center',
            saveData.building || ''
        );
        console.log('✅ Converted to IDs:', locationIds);

        // The save data is flat, so we need to map it to the DOM structure
        // Character data includes vault, spell_slots, time tracking
        const characterData = {
            name: saveData.d || 'Unknown',
            race: saveData.race,
            class: saveData.class,
            background: saveData.background,
            alignment: saveData.alignment,
            level: saveData.level || 1,
            experience: saveData.experience || 0,
            hp: saveData.hp,
            max_hp: saveData.max_hp,
            mana: saveData.mana,
            max_mana: saveData.max_mana,
            fatigue: saveData.fatigue || 0,
            gold: saveData.gold || 0,
            stats: saveData.stats,
            vault: saveData.vault || {},
            spell_slots: saveData.spell_slots || {},
            current_day: saveData.current_day || 1,
            time_of_day: saveData.time_of_day || 'day'
        };

        document.getElementById('character-data').textContent = JSON.stringify(characterData);

        // Location uses IDs (converted from display names)
        const locationData = {
            current: locationIds.locationId,
            district: locationIds.districtKey,
            building: locationIds.buildingId || '',
            discovered: saveData.locations_discovered || []
        };
        console.log('✅ Location data for state:', locationData);
        document.getElementById('location-data').textContent = JSON.stringify(locationData);

        // Inventory general slots
        if (saveData.inventory && saveData.inventory.general_slots) {
            document.getElementById('inventory-data').textContent = JSON.stringify(saveData.inventory.general_slots);
        } else {
            document.getElementById('inventory-data').textContent = '[]';
        }

        // Spells (known_spells)
        if (saveData.known_spells) {
            document.getElementById('spell-data').textContent = JSON.stringify(saveData.known_spells);
        } else {
            document.getElementById('spell-data').textContent = '[]';
        }

        // Equipment (gear_slots)
        if (saveData.inventory && saveData.inventory.gear_slots) {
            document.getElementById('equipment-data').textContent = JSON.stringify(saveData.inventory.gear_slots);
        } else {
            document.getElementById('equipment-data').textContent = '{}';
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
        const character = gameState.character;

        // Convert location IDs to display names for save file
        const displayNames = await window.getDisplayNamesForLocation(
            gameState.location?.current || 'kingdom',
            gameState.location?.district || 'center',
            gameState.location?.building || ''
        );

        // Prepare save data in the flat structure the backend expects (with display names)
        const saveData = {
            d: character.name || '',
            race: character.race || '',
            class: character.class || '',
            background: character.background || '',
            alignment: character.alignment || '',
            experience: character.experience || 0,
            hp: character.hp || 0,
            max_hp: character.max_hp || 0,
            mana: character.mana || 0,
            max_mana: character.max_mana || 0,
            fatigue: character.fatigue || 0,
            gold: character.gold || 0,
            stats: character.stats || {},
            location: displayNames.location,
            district: displayNames.district,
            building: displayNames.building,
            inventory: character.inventory || {},
            vault: character.vault || {},
            known_spells: character.spells || [],
            spell_slots: character.spell_slots || {},
            locations_discovered: gameState.location?.discovered || [],
            music_tracks_unlocked: character.music_tracks_unlocked || [],
            current_day: character.current_day || 1,
            time_of_day: character.time_of_day || 'day'
        };

        // ALWAYS get the save ID from URL - we only overwrite, never create new saves
        const urlParams = new URLSearchParams(window.location.search);
        const currentSaveID = urlParams.get('save');

        if (!currentSaveID) {
            throw new Error('No save file loaded. Cannot save without an active save.');
        }

        // Always use the existing save ID (overwrite mode)
        saveData.id = currentSaveID;

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

// DISABLED: Auto-save on page unload/hide
// These were creating blank save files because the page might not be fully loaded
// Users can use Ctrl+S or the manual save button in settings instead

// // Save before leaving page
// window.addEventListener('beforeunload', async (event) => {
//     if (saveDataLoaded && window.sessionManager && window.sessionManager.isAuthenticated()) {
//         // Attempt quick save (though it may not complete due to page unload)
//         saveGameToLocal();
//     }
// });

// // Save on page visibility change (when tab becomes hidden)
// document.addEventListener('visibilitychange', async () => {
//     if (saveDataLoaded && document.hidden && window.sessionManager && window.sessionManager.isAuthenticated()) {
//         console.log('👁️ Page hidden, saving game...');
//         await saveGameToLocal();
//     }
// });

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

// Alias for manual save button
window.saveGame = saveGameToLocal;
window.saveGameToLocal = saveGameToLocal;  // Explicit alias for other modules

console.log('💾 Save system loaded');