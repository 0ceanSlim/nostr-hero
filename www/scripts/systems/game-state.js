// Game State Management - Go-First Architecture
// All game state lives in Go memory
// JavaScript fetches state from Go and triggers UI updates

// Cache the last fetched state for immediate UI access
let cachedGameState = null;

// Get the current game state (fetches from Go if needed)
async function getGameState(forceRefresh = false) {
    if (!forceRefresh && cachedGameState) {
        return cachedGameState;
    }

    try {
        // Fetch from Go backend
        const state = await window.gameAPI.getState();

        // Transform Go SaveFile format to UI format
        const uiState = transformSaveDataToUIState(state);

        // Cache it
        cachedGameState = uiState;

        return uiState;
    } catch (error) {
        console.error('❌ Failed to fetch game state:', error);
        // Return cached state if fetch fails
        return cachedGameState || getEmptyGameState();
    }
}

// Synchronous version for immediate access (uses cached state)
function getGameStateSync() {
    if (!cachedGameState) {
        console.warn('⚠️ No cached game state, returning empty state');
        return getEmptyGameState();
    }
    return cachedGameState;
}

// Transform SaveFile (Go format) to UI state format
function transformSaveDataToUIState(saveData) {
    // Map display names to location/district IDs
    const locationNameMap = {
        // Cities
        'Verdant City': 'verdant',
        'Golden Haven': 'goldenhaven',
        'Goldenhaven': 'goldenhaven',

        // Towns
        'Ironpeak': 'ironpeak',
        'Iron Peak': 'ironpeak',
        'Frosthold': 'frosthold',

        // Villages
        'Millhaven Village': 'millhaven',
        'Millhaven': 'millhaven',
        'Saltwind Village': 'saltwind',
        'Saltwind': 'saltwind',
        'Marshlight Village': 'marshlight',
        'Marshlight': 'marshlight',
        'Dusthaven Village': 'dusthaven',
        'Dusthaven': 'dusthaven',

        // Kingdoms/General
        'Kingdom': 'kingdom',
        'The Royal Kingdom': 'kingdom',
        'Village': 'village',
        'Dwarven Stronghold': 'dwarven',
        'Desert Oasis': 'desert',
        'Port City': 'port'
    };

    const districtNameMap = {
        // Generic district names
        'Garden Plaza': 'center',
        'Town Square': 'center',
        'Village Center': 'center',
        'City Center': 'center',
        'Town Center': 'center',
        'Kingdom Center': 'center',

        // Directional districts
        'Market District': 'market',
        'Northern Quarter': 'north',
        'North District': 'north',
        'Southern Quarter': 'south',
        'South District': 'south',
        'Western Quarter': 'west',
        'West District': 'west',
        'Eastern Quarter': 'east',
        'East District': 'east',

        // Special districts
        'Harbor': 'harbor',
        'Residential': 'residential'
    };

    // Get location ID (handle both display names and IDs)
    let locationId = saveData.location || 'kingdom';
    if (locationNameMap[locationId]) {
        locationId = locationNameMap[locationId];
    }

    // Get district key (handle both display names and IDs)
    let districtKey = saveData.district || 'center';
    // If it's a full district ID like "verdant-center", extract the key
    if (districtKey.includes('-')) {
        districtKey = districtKey.split('-').pop();
    }
    // If it's a display name like "Garden Plaza", map it
    if (districtNameMap[districtKey]) {
        districtKey = districtNameMap[districtKey];
    }

    // Normalize stats keys to lowercase (Go sends "Strength", UI expects "strength")
    const normalizedStats = {};
    if (saveData.stats) {
        for (const [key, value] of Object.entries(saveData.stats)) {
            normalizedStats[key.toLowerCase()] = value;
        }
    }

    console.log('🔄 Transforming save data to UI state:', {
        savedLocation: saveData.location,
        savedDistrict: saveData.district,
        mappedLocationId: locationId,
        mappedDistrictKey: districtKey,
        hasInventory: !!saveData.inventory,
        hasGearSlots: !!saveData.inventory?.gear_slots,
        hasGeneralSlots: !!saveData.inventory?.general_slots,
        statsNormalized: normalizedStats
    });

    return {
        character: {
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
            fatigue_counter: saveData.fatigue_counter || 0,
            hunger: saveData.hunger !== undefined ? saveData.hunger : 1,
            hunger_counter: saveData.hunger_counter || 0,
            gold: saveData.gold || 0,
            stats: normalizedStats,
            inventory: saveData.inventory || {},
            vaults: saveData.vaults || [],
            spells: saveData.known_spells || [],
            spell_slots: saveData.spell_slots || {},
            music_tracks_unlocked: saveData.music_tracks_unlocked || [],
            current_day: saveData.current_day || 1,
            time_of_day: saveData.time_of_day !== undefined ? saveData.time_of_day : 12
        },
        location: {
            current: locationId,
            district: districtKey,
            building: saveData.building || null,
            discovered: saveData.locations_discovered || []
        },
        inventory: saveData.inventory?.general_slots || [],
        equipment: saveData.inventory?.gear_slots || {},
        spells: saveData.known_spells || [],
        combat: null // Combat state not stored in save file
    };
}

// Get empty game state (fallback)
function getEmptyGameState() {
    return {
        character: {},
        inventory: [],
        equipment: {},
        spells: [],
        location: {},
        combat: null
    };
}

// Update the game state (triggers UI refresh)
// This is now just for UI updates - game actions should use gameAPI.sendAction()
async function refreshGameState() {
    console.log('🔄 Refreshing game state from Go...');

    // Fetch fresh state from Go
    const newState = await getGameState(true);

    console.log('📦 Fetched state:', {
        hasCharacter: !!newState.character,
        hasInventory: !!newState.inventory,
        hasEquipment: !!newState.equipment,
        inventorySlots: newState.inventory?.length,
        characterInventory: !!newState.character?.inventory
    });

    // Trigger UI updates
    console.log('📢 Dispatching gameStateChange event...');
    document.dispatchEvent(new CustomEvent('gameStateChange', { detail: newState }));

    return newState;
}

// Legacy compatibility: Update state locally (only for UI, doesn't persist)
// DEPRECATED: Use gameAPI.sendAction() instead
function updateGameState(newState) {
    console.warn('⚠️ updateGameState() is deprecated. Use gameAPI.sendAction() instead.');

    // Update cache
    if (cachedGameState) {
        if (newState.character) cachedGameState.character = {...cachedGameState.character, ...newState.character};
        if (newState.inventory) cachedGameState.inventory = newState.inventory;
        if (newState.equipment) cachedGameState.equipment = newState.equipment;
        if (newState.spells) cachedGameState.spells = newState.spells;
        if (newState.location) cachedGameState.location = {...cachedGameState.location, ...newState.location};
        if (newState.combat !== undefined) cachedGameState.combat = newState.combat;
    } else {
        cachedGameState = newState;
    }

    // Trigger UI updates
    document.dispatchEvent(new CustomEvent('gameStateChange', { detail: cachedGameState }));
}

// Get static game data helpers
function getItemById(itemId) {
    const allItems = JSON.parse(document.getElementById('all-items').textContent || '[]');
    return allItems.find(item => item.id === itemId);
}

function getSpellById(spellId) {
    const allSpells = JSON.parse(document.getElementById('all-spells').textContent || '[]');
    return allSpells.find(spell => spell.id === spellId);
}

function getLocationById(locationId) {
    const allLocations = JSON.parse(document.getElementById('all-locations').textContent || '[]');

    // First, try to find in top-level locations
    const topLevel = allLocations.find(location => location.id === locationId);
    if (topLevel) return topLevel;

    // If not found, search within city districts
    for (const location of allLocations) {
        if (location.properties?.districts) {
            for (const district of Object.values(location.properties.districts)) {
                if (district.id === locationId) {
                    return district;
                }
            }
        }
    }

    return undefined;
}

function getMonsterById(monsterId) {
    const allMonsters = JSON.parse(document.getElementById('all-monsters').textContent || '[]');
    return allMonsters.find(monster => monster.id === monsterId);
}

function getPackById(packId) {
    const allPacks = JSON.parse(document.getElementById('all-packs').textContent || '[]');
    return allPacks.find(pack => pack.id === packId);
}

function getNPCById(npcId) {
    const allNPCs = JSON.parse(document.getElementById('all-npcs').textContent || '[]');
    return allNPCs.find(npc => npc.id === npcId);
}

// Helper functions for inventory and spell management
function findInInventory(inventory, itemId) {
    return inventory.find(item => item.item === itemId);
}

function findInSpells(spells, spellId) {
    return spells.find(spell => spell.spell === spellId);
}

// Get current character's npub from session manager
function getCurrentNpub() {
    if (window.sessionManager && window.sessionManager.isAuthenticated()) {
        return window.sessionManager.getNpub();
    }
    return null;
}

// Listen for startup completion and initialize game
window.addEventListener('nostrHeroReady', () => {
    console.log('🎮 Startup complete, initializing game...');
    initializeGame();
});

// Initialize the game with fresh state or from save
async function initializeGame() {
    console.log('🎮 Initializing Nostr Hero...');

    // Wait for session manager to be ready
    if (!window.sessionManager) {
        console.error('❌ SessionManager not available');
        showMessage('❌ Session manager not loaded', 'error');
        return;
    }

    // Initialize session manager and wait for result
    try {
        await window.sessionManager.init();

        if (!window.sessionManager.isAuthenticated()) {
            console.log('🔐 User not authenticated, showing login interface');
            redirectToLogin();
            return;
        }

        console.log('✅ User authenticated, loading game...');
        const session = window.sessionManager.getSession();
        console.log(`🎮 Starting game for user: ${session.npub}`);

        // Get save ID from URL
        const urlParams = new URLSearchParams(window.location.search);
        const saveID = urlParams.get('save');

        // Initialize Game API
        if (saveID) {
            window.gameAPI.init(session.npub, saveID);
            console.log('🎮 Game API initialized with save:', saveID);
        }

    } catch (error) {
        console.error('❌ Session initialization failed:', error);
        showMessage('❌ Failed to initialize session: ' + error.message, 'error');
        redirectToLogin();
        return;
    }

    showMessage('🎮 Loading game data...', 'info');

    try {
        // Load static game data
        const gameDataResponse = await fetch('/api/game-data');
        if (!gameDataResponse.ok) {
            throw new Error('Failed to load game data');
        }

        const gameData = await gameDataResponse.json();

        // Store static data in DOM
        document.getElementById('all-items').textContent = JSON.stringify(gameData.items || []);
        document.getElementById('all-spells').textContent = JSON.stringify(gameData.spells || []);
        document.getElementById('all-monsters').textContent = JSON.stringify(gameData.monsters || []);
        document.getElementById('all-locations').textContent = JSON.stringify(gameData.locations || []);
        document.getElementById('all-packs').textContent = JSON.stringify(gameData.packs || []);

        // Load NPCs separately
        const npcsResponse = await fetch('/api/npcs');
        if (npcsResponse.ok) {
            const npcs = await npcsResponse.json();
            document.getElementById('all-npcs').textContent = JSON.stringify(npcs || []);
            console.log(`Loaded ${npcs?.length || 0} NPCs`);
        }

        console.log(`Loaded game data: ${gameData.items?.length || 0} items, ${gameData.spells?.length || 0} spells, ${gameData.monsters?.length || 0} monsters, ${gameData.locations?.length || 0} locations`);

        // Check if we're loading a specific save
        const session = window.sessionManager.getSession();
        const urlParams = new URLSearchParams(window.location.search);
        const saveID = urlParams.get('save');

        if (saveID) {
            // Load specific save into Go memory
            try {
                console.log('🔄 Initializing session in Go memory...');

                // Initialize session in backend memory
                const initResponse = await fetch('/api/session/init', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        npub: session.npub,
                        save_id: saveID
                    })
                });

                if (!initResponse.ok) {
                    throw new Error(`Failed to initialize session: ${initResponse.status}`);
                }

                const initResult = await initResponse.json();
                console.log('✅ Session loaded into Go memory:', initResult);

                // Fetch initial game state from Go
                await refreshGameState();

                showMessage('✅ Save file loaded successfully!', 'success');

            } catch (saveError) {
                console.error('Failed to load specific save:', saveError);
                showMessage('❌ Failed to load save: ' + saveError.message, 'error');
                // Redirect back to saves page
                setTimeout(() => window.location.href = '/saves', 2000);
                return;
            }
        } else {
            // No save specified, redirect to saves page
            console.log('🔄 No save specified, redirecting to save selection');
            window.location.href = '/saves';
            return;
        }

        // Hide loading screen and show game UI
        const gameApp = document.getElementById('game-app');
        if (gameApp) {
            // Clear the loading screen content
            gameApp.style.display = 'block';
        }

        // Start game UI
        displayCurrentLocation();
        updateAllDisplays();

        console.log('✅ Game initialized and ready to play!');

        // DEBUG: Disable auto-save system to prevent undefined saves
        // if (typeof startAutoSave === 'function') {
        //     startAutoSave();
        // }
        console.log('🚫 Auto-save disabled for debugging');

    } catch (error) {
        console.error('Failed to initialize game:', error);
        showMessage('❌ Failed to load game: ' + error.message, 'error');
    }
}

// Initialize game state from a loaded save
async function initializeFromSave(saveData) {
    console.log('Loading game from save:', saveData);

    // Convert display names back to IDs for internal game logic
    const ids = await getIdsFromDisplayNames(
        saveData.location || 'The Royal Kingdom',
        saveData.district || 'Kingdom Center',
        saveData.building || ''
    );

    // New save structure is flat with all fields at root level
    // Map it to the game state structure expected by the UI (using IDs internally)
    const gameState = {
        character: {
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
            fatigue_counter: saveData.fatigue_counter || 0,
            hunger: saveData.hunger !== undefined ? saveData.hunger : 1,  // Default to Hungry (0-3 scale)
            hunger_counter: saveData.hunger_counter || 0,
            gold: saveData.gold || 0,
            stats: saveData.stats,
            inventory: saveData.inventory,
            vault: saveData.vault || {},
            spells: saveData.known_spells || [],
            spell_slots: saveData.spell_slots || {},
            current_day: saveData.current_day || 1,
            time_of_day: saveData.time_of_day !== undefined ? saveData.time_of_day : 12  // Default to noon (0-23 hours)
        },
        location: {
            current: ids.locationId,
            district: ids.districtKey,
            building: ids.buildingId || null,
            discovered: saveData.locations_discovered || [ids.locationId]
        },
        inventory: saveData.inventory?.general_slots || [],
        equipment: saveData.inventory?.gear_slots || {},
        spells: saveData.known_spells || [],
        combat: null
    };

    console.log('Mapped game state:', gameState);
    updateGameState(gameState);
}
// Convert location/district/building IDs to display names for saving
async function getDisplayNamesForLocation(locationId, districtKey, buildingId) {
    try {
        // Fetch location data directly from API instead of DOM (DOM gets wiped by intro screen)
        const response = await fetch('/api/locations');
        if (!response.ok) {
            console.warn('Failed to fetch locations from API');
            return { location: locationId, district: districtKey, building: buildingId };
        }

        const allLocations = await response.json();
        console.log('🔍 Fetched', allLocations.length, 'locations from API');
        console.log('🔍 Looking for locationId:', locationId);

        // Find the location
        const location = allLocations.find(loc => loc.id === locationId);
        if (!location) {
            console.warn('❌ Location not found:', locationId);
            return { location: locationId, district: districtKey, building: buildingId };
        }

        const locationName = location.name || locationId;
        console.log('✅ Found location name:', locationName);

        // Find the district (check both location.districts and location.properties.districts)
        const districts = location.districts || location.properties?.districts;
        let districtName = districtKey;
        if (districts && districts[districtKey]) {
            districtName = districts[districtKey].name || districtKey;
            console.log('✅ Found district name:', districtName);
        } else {
            console.warn('❌ District not found:', districtKey, 'Available districts:', districts ? Object.keys(districts) : 'none');
        }

        // Find the building
        let buildingName = buildingId || '';
        if (buildingId && districts && districts[districtKey]) {
            const district = districts[districtKey];
            if (district.buildings) {
                const building = district.buildings.find(b => b.id === buildingId);
                if (building) {
                    buildingName = building.name || buildingId;
                }
            }
        }

        const result = {
            location: locationName,
            district: districtName,
            building: buildingName
        };
        console.log('🔍 Final display names:', result);
        return result;
    } catch (error) {
        console.error('❌ Error getting display names:', error);
        return { location: locationId, district: districtKey, building: buildingId };
    }
}

// Convert display names back to IDs for game logic
async function getIdsFromDisplayNames(locationName, districtName, buildingName) {
    try {
        // Fetch location data directly from API
        const response = await fetch('/api/locations');
        if (!response.ok) {
            console.warn('Failed to fetch locations from API');
            return { locationId: locationName, districtKey: districtName, buildingId: buildingName };
        }

        const allLocations = await response.json();

        // Find location by name
        const location = allLocations.find(loc => loc.name === locationName);
        if (!location) {
            console.warn('Location not found by name:', locationName);
            return { locationId: locationName, districtKey: districtName, buildingId: buildingName };
        }

        const locationId = location.id;

        // Find district by name (check both location.districts and location.properties.districts)
        const districts = location.districts || location.properties?.districts;
        let districtKey = districtName;
        if (districts) {
            for (const [key, district] of Object.entries(districts)) {
                if (district.name === districtName) {
                    districtKey = key;
                    break;
                }
            }
        }

        // Find building by name
        let buildingId = buildingName || '';
        if (buildingName && districts && districts[districtKey]) {
            const district = districts[districtKey];
            if (district.buildings) {
                const building = district.buildings.find(b => b.name === buildingName);
                if (building) {
                    buildingId = building.id;
                }
            }
        }

        return {
            locationId: locationId,
            districtKey: districtKey,
            buildingId: buildingId
        };
    } catch (error) {
        console.error('Error getting IDs from display names:', error);
        return { locationId: locationName, districtKey: districtName, buildingId: buildingName };
    }
}

// Export helper functions globally for use in other modules
window.getDisplayNamesForLocation = getDisplayNamesForLocation;
window.getIdsFromDisplayNames = getIdsFromDisplayNames;

// ========================================
// Ground Items System (Session-only storage)
// ========================================

// Ground items storage: { "locationId-districtKey": [{item, droppedAt, droppedDay}, ...] }
const groundItems = {};

/**
 * Get location key for ground storage
 */
function getGroundLocationKey() {
    const state = getGameStateSync();
    const cityId = state.location?.current || 'unknown';
    const districtKey = state.location?.district || 'center';
    return `${cityId}-${districtKey}`;
}

/**
 * Add item to ground at current location
 */
function addItemToGround(itemId, quantity = 1) {
    const locationKey = getGroundLocationKey();
    const state = getGameStateSync();
    const currentDay = state.character?.current_day || 1;

    if (!groundItems[locationKey]) {
        groundItems[locationKey] = [];
    }

    groundItems[locationKey].push({
        item: itemId,
        quantity: quantity,
        droppedAt: Date.now(),
        droppedDay: currentDay
    });

    console.log(`📍 Item ${itemId} dropped at ${locationKey}`);
    cleanupOldGroundItems();
}

/**
 * Remove item from ground at current location
 */
function removeItemFromGround(itemId) {
    const locationKey = getGroundLocationKey();

    if (!groundItems[locationKey]) {
        return null;
    }

    const index = groundItems[locationKey].findIndex(ground => ground.item === itemId);
    if (index === -1) {
        return null;
    }

    const removed = groundItems[locationKey].splice(index, 1)[0];
    console.log(`✅ Picked up ${itemId} from ${locationKey}`);
    return removed;
}

/**
 * Get all items on ground at current location
 */
function getGroundItems() {
    const locationKey = getGroundLocationKey();
    cleanupOldGroundItems();
    return groundItems[locationKey] || [];
}

/**
 * Clean up items older than 1 game day
 */
function cleanupOldGroundItems() {
    const state = getGameStateSync();
    const currentDay = state.character?.current_day || 1;

    for (const locationKey in groundItems) {
        groundItems[locationKey] = groundItems[locationKey].filter(ground => {
            const daysPassed = currentDay - ground.droppedDay;
            return daysPassed < 1; // Keep items for less than 1 day
        });

        // Remove empty locations
        if (groundItems[locationKey].length === 0) {
            delete groundItems[locationKey];
        }
    }
}

// Export ground functions globally
window.addItemToGround = addItemToGround;
window.removeItemFromGround = removeItemFromGround;
window.getGroundItems = getGroundItems;

// Event listener for game state changes - triggers UI updates
document.addEventListener('gameStateChange', async function(event) {
    // The event.detail contains the updated state
    const state = event.detail || getGameStateSync();

    console.log('🎨 gameStateChange event fired, updating displays...');

    // Update all displays
    await updateCharacterDisplay();

    // Update location display
    if (typeof displayCurrentLocation === 'function') {
        displayCurrentLocation();
    }

    // Give the DOM time to render
    await new Promise(resolve => setTimeout(resolve, 100));

    // Now rebind inventory interactions after UI is updated
    if (window.inventoryInteractions && window.inventoryInteractions.bindInventoryEvents) {
        console.log('🔄 Rebinding inventory events...');
        window.inventoryInteractions.bindInventoryEvents();
    }
});

console.log('Game state management loaded');