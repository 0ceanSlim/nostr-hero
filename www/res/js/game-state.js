// Game State Management
// All game state is stored in hidden DOM elements and managed through these functions

// Get the current game state from DOM
function getGameState() {
    return {
        character: JSON.parse(document.getElementById('character-data').textContent || '{}'),
        inventory: JSON.parse(document.getElementById('inventory-data').textContent || '[]'),
        spells: JSON.parse(document.getElementById('spell-data').textContent || '[]'),
        location: JSON.parse(document.getElementById('location-data').textContent || '{}'),
        combat: JSON.parse(document.getElementById('combat-data').textContent || 'null')
    };
}

// Update the game state in DOM and trigger UI updates
function updateGameState(newState) {
    if (newState.character) {
        document.getElementById('character-data').textContent = JSON.stringify(newState.character);
    }
    if (newState.inventory) {
        document.getElementById('inventory-data').textContent = JSON.stringify(newState.inventory);
    }
    if (newState.spells) {
        document.getElementById('spell-data').textContent = JSON.stringify(newState.spells);
    }
    if (newState.location) {
        document.getElementById('location-data').textContent = JSON.stringify(newState.location);
    }
    if (newState.combat !== undefined) {
        document.getElementById('combat-data').textContent = JSON.stringify(newState.combat);
    }

    // Trigger UI updates
    document.dispatchEvent(new CustomEvent('gameStateChange', { detail: newState }));
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
    return allLocations.find(location => location.id === locationId);
}

function getMonsterById(monsterId) {
    const allMonsters = JSON.parse(document.getElementById('all-monsters').textContent || '[]');
    return allMonsters.find(monster => monster.id === monsterId);
}

function getPackById(packId) {
    const allPacks = JSON.parse(document.getElementById('all-packs').textContent || '[]');
    return allPacks.find(pack => pack.id === packId);
}

// Helper functions for inventory and spell management
function findInInventory(inventory, itemId) {
    return inventory.find(item => item.item === itemId);
}

function findInSpells(spells, spellId) {
    return spells.find(spell => spell.spell === spellId);
}

// Get current character's npub from grain session
async function getCurrentNpub() {
    try {
        const response = await fetch('/api/auth/session');
        if (response.ok) {
            const sessionData = await response.json();
            if (sessionData.success && sessionData.is_active) {
                return sessionData.npub || sessionData.session?.public_key;
            }
        }
        return null;
    } catch (error) {
        console.error('Failed to get current user:', error);
        return null;
    }
}

// Initialize the game with fresh state or from save
async function initializeGame() {
    console.log('Initializing game...');

    const npub = await getCurrentNpub();
    if (!npub) {
        showMessage('❌ No user logged in', 'error');
        // Redirect to login or show login interface
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

        console.log(`Loaded game data: ${gameData.items?.length || 0} items, ${gameData.spells?.length || 0} spells, ${gameData.monsters?.length || 0} monsters, ${gameData.locations?.length || 0} locations`);

        // Try to load existing save
        try {
            const saveResponse = await fetch(`/api/load-save/${npub}`);
            if (saveResponse.ok) {
                const saveData = await saveResponse.json();
                initializeFromSave(saveData.gameState);
                showMessage('✅ Game loaded successfully!', 'success');
            } else {
                // Create new character
                await createNewCharacter(npub);
                showMessage('🎉 New character created!', 'success');
            }
        } catch (saveError) {
            console.warn('Failed to load save, creating new character:', saveError);
            await createNewCharacter(npub);
            showMessage('🎉 New character created!', 'success');
        }

        // Start game UI
        displayCurrentLocation();
        updateAllDisplays();

    } catch (error) {
        console.error('Failed to initialize game:', error);
        showMessage('❌ Failed to load game: ' + error.message, 'error');
    }
}

// Initialize game state from a loaded save
function initializeFromSave(gameState) {
    console.log('Loading game from save:', gameState);
    updateGameState(gameState);
}

// Create a new character for a given npub
async function createNewCharacter(npub) {
    console.log('Creating new character for npub:', npub);

    try {
        const characterResponse = await fetch(`/api/character?npub=${npub}`);
        if (!characterResponse.ok) {
            throw new Error('Failed to generate character');
        }

        const character = await characterResponse.json();
        console.log('Generated character:', character);

        // Initialize fresh game state
        const initialState = {
            character: {
                ...character,
                hp: character.max_hp || 10,
                mana: character.max_mana || 0,
                fatigue: 0
            },
            inventory: generateStartingInventory(character.class),
            spells: generateStartingSpells(character.class),
            location: {
                current: character.starting_location || 'kingdom-center',
                discovered: [character.starting_location || 'kingdom-center']
            },
            combat: null
        };

        updateGameState(initialState);
        console.log('Initialized game state:', initialState);

    } catch (error) {
        console.error('Failed to create character:', error);
        throw error;
    }
}

// Generate starting inventory based on class
function generateStartingInventory(characterClass) {
    // This is a simplified starting inventory
    // You can expand this based on your character data
    const baseItems = [
        { item: 'rations', quantity: 5 },
        { item: 'bedroll', quantity: 1 },
        { item: 'waterskin', quantity: 1 }
    ];

    // Add class-specific items
    switch (characterClass?.toLowerCase()) {
        case 'wizard':
            baseItems.push({ item: 'spellbook', quantity: 1 });
            baseItems.push({ item: 'component-pouch', quantity: 1 });
            break;
        case 'fighter':
            baseItems.push({ item: 'longsword', quantity: 1 });
            baseItems.push({ item: 'shield', quantity: 1 });
            break;
        case 'rogue':
            baseItems.push({ item: 'dagger', quantity: 2 });
            baseItems.push({ item: 'thieves-tools', quantity: 1 });
            break;
        // Add more classes as needed
    }

    return baseItems;
}

// Generate starting spells based on class
function generateStartingSpells(characterClass) {
    const startingSpells = [];

    switch (characterClass?.toLowerCase()) {
        case 'wizard':
        case 'sorcerer':
            startingSpells.push(
                { spell: 'fire-bolt', prepared: true },
                { spell: 'magic-missile', prepared: true }
            );
            break;
        case 'cleric':
            startingSpells.push(
                { spell: 'cure-wounds', prepared: true },
                { spell: 'sacred-flame', prepared: true }
            );
            break;
        // Add more classes as needed
    }

    return startingSpells;
}

// Event listeners for game state changes
document.addEventListener('gameStateChange', function(event) {
    updateCharacterDisplay();
    updateInventoryDisplay();
    updateSpellsDisplay();

    // Update combat interface if in combat
    const state = getGameState();
    if (state.combat) {
        updateCombatInterface();
    }
});

console.log('Game state management loaded');