// Game State Management
// All game state is stored in hidden DOM elements and managed through these functions

// Get the current game state from DOM
function getGameState() {
    return {
        character: JSON.parse(document.getElementById('character-data').textContent || '{}'),
        inventory: JSON.parse(document.getElementById('inventory-data').textContent || '[]'),
        equipment: JSON.parse(document.getElementById('equipment-data').textContent || '{}'),
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
    if (newState.equipment) {
        document.getElementById('equipment-data').textContent = JSON.stringify(newState.equipment);
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

        console.log(`Loaded game data: ${gameData.items?.length || 0} items, ${gameData.spells?.length || 0} spells, ${gameData.monsters?.length || 0} monsters, ${gameData.locations?.length || 0} locations`);

        // Check if we're loading a specific save or starting new game
        const session = window.sessionManager.getSession();
        const urlParams = new URLSearchParams(window.location.search);
        const saveID = urlParams.get('save');
        const newGame = urlParams.get('new') === 'true';

        if (saveID) {
            // Load specific save
            try {
                const saveResponse = await fetch(`/api/saves/${session.npub}`);
                if (saveResponse.ok) {
                    const saves = await saveResponse.json();
                    const save = saves.find(s => s.id === saveID);

                    if (save) {
                        // New save structure: character and location are at root level
                        await initializeFromSave(save);
                        showMessage('✅ Save file loaded successfully!', 'success');
                    } else {
                        throw new Error('Save file not found');
                    }
                } else {
                    throw new Error('Failed to load save files');
                }
            } catch (saveError) {
                console.error('Failed to load specific save:', saveError);
                showMessage('❌ Failed to load save: ' + saveError.message, 'error');
                // Redirect back to saves page
                setTimeout(() => window.location.href = '/saves', 2000);
                return;
            }
        } else if (newGame) {
            // Create new character
            await createNewCharacter(session.npub);
            showMessage('🎉 New adventure begins!', 'success');
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
            gold: saveData.gold || 0,
            stats: saveData.stats,
            inventory: saveData.inventory,
            vault: saveData.vault || {},
            spells: saveData.known_spells || [],
            spell_slots: saveData.spell_slots || {},
            current_day: saveData.current_day || 1,
            time_of_day: saveData.time_of_day || 'day'
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

// Global variable to store initial state during character creation
let pendingCharacterState = null;

// Create a new character for a given npub
async function createNewCharacter(npub) {
    console.log('Creating new character for npub:', npub);

    try {
        const characterResponse = await fetch(`/api/character?npub=${npub}`);
        if (!characterResponse.ok) {
            throw new Error('Failed to generate character');
        }

        const result = await characterResponse.json();
        console.log('Generated character result:', result);

        const characterData = result.character;

        // Calculate derived stats based on class and race
        const maxHP = calculateMaxHP(characterData.class, characterData.stats.Constitution, characterData.race);
        const maxMana = calculateMaxMana(characterData.class, characterData.stats);
        const startingLocation = getStartingLocation(characterData.race);

        // Initialize complete character with all required fields
        // DO NOT include vault, current_day, time_of_day here - they go in character object below
        const fullCharacter = {
            race: characterData.race,
            class: characterData.class,
            background: characterData.background,
            alignment: characterData.alignment,
            stats: characterData.stats,
            name: generateCharacterName(characterData.race, characterData.background),
            level: 1,
            experience: 0,
            hp: maxHP,
            max_hp: maxHP,
            mana: maxMana,
            max_mana: maxMana,
            fatigue: 0,
            gold: generateStartingGold(characterData.class, characterData.background),
            starting_location: startingLocation,
            music_tracks_unlocked: []
        };

        console.log('🔍 fullCharacter:', fullCharacter);

        // Generate starting equipment
        const startingEquipment = generateStartingEquipment(characterData.class);

        // Add slot numbers to inventory items
        const inventoryWithSlots = startingEquipment.inventory.map((item, index) => ({
            ...item,
            slot: index
        }));

        // Generate starting vault (40 slots, empty)
        const startingVault = generateStartingVault(startingLocation);
        console.log('🔍 Generated starting vault:', startingVault);

        // Extract district from location
        const district = getDistrictFromLocation(startingLocation);
        console.log('🔍 Starting location and district:', startingLocation, district);

        // Initialize fresh game state
        const characterObject = {
            ...fullCharacter,
            current_day: 1,
            time_of_day: 'day',
            vault: startingVault,
            spell_slots: generateSpellSlots(characterData.class)
        };

        console.log('🔍 characterObject with vault/time:', {
            vault: characterObject.vault,
            current_day: characterObject.current_day,
            time_of_day: characterObject.time_of_day
        });

        const initialState = {
            character: characterObject,
            inventory: inventoryWithSlots,
            equipment: startingEquipment.equipped,
            spells: generateStartingSpells(characterData.class),
            location: {
                current: startingLocation,
                district: district,
                building: null,  // Start outdoors in the center district
                discovered: [startingLocation]
            },
            combat: null
        };

        // Store state globally for beginAdventure to use
        pendingCharacterState = initialState;

        console.log('🔍 Initialized game state:', initialState);
        console.log('🔍 Character vault in state:', initialState.character.vault);

        // Show character intro before starting the game
        showCharacterIntro(fullCharacter);

    } catch (error) {
        console.error('Failed to create character:', error);
        throw error;
    }
}

// Generate starting equipment (equipped vs inventory) based on class
function generateStartingEquipment(characterClass) {
    // Define what equipment slots can be equipped
    const equipmentSlots = {
        mainHand: null,
        offHand: null,
        armor: null,
        helmet: null,
        boots: null,
        gloves: null,
        ring1: null,
        ring2: null,
        necklace: null,
        cloak: null
    };

    // Base survival items (always go to inventory)
    const inventoryItems = [
        { item: 'rations-1-day', quantity: 3 },
        { item: 'bedroll', quantity: 1 },
        { item: 'waterskin', quantity: 1 },
        { item: 'rope-hempen-50-feet', quantity: 1 },
        { item: 'backpack', quantity: 1 }
    ];

    // Add class-specific starting equipment
    switch (characterClass) {
        case 'Barbarian':
            equipmentSlots.mainHand = { item: 'greataxe', quantity: 1 };
            equipmentSlots.armor = { item: 'hide', quantity: 1 };
            inventoryItems.push({ item: 'handaxe', quantity: 2 });
            break;
        case 'Bard':
            equipmentSlots.mainHand = { item: 'rapier', quantity: 1 };
            equipmentSlots.armor = { item: 'leather', quantity: 1 };
            inventoryItems.push(
                { item: 'lute', quantity: 1 },
                { item: 'dagger', quantity: 1 }
            );
            break;
        case 'Cleric':
            equipmentSlots.mainHand = { item: 'mace', quantity: 1 };
            equipmentSlots.offHand = { item: 'shield', quantity: 1 };
            equipmentSlots.armor = { item: 'chain-mail', quantity: 1 };
            inventoryItems.push({ item: 'reliquary', quantity: 1 });
            break;
        case 'Druid':
            equipmentSlots.mainHand = { item: 'quarterstaff', quantity: 1 };
            equipmentSlots.armor = { item: 'leather', quantity: 1 };
            equipmentSlots.offHand = { item: 'shield', quantity: 1 };
            inventoryItems.push({ item: 'sprig-of-mistletoe', quantity: 1 });
            break;
        case 'Fighter':
            equipmentSlots.mainHand = { item: 'longsword', quantity: 1 };
            equipmentSlots.offHand = { item: 'shield', quantity: 1 };
            equipmentSlots.armor = { item: 'chain-mail', quantity: 1 };
            inventoryItems.push({ item: 'handaxe', quantity: 2 });
            break;
        case 'Monk':
            equipmentSlots.mainHand = { item: 'quarterstaff', quantity: 1 };
            equipmentSlots.armor = { item: 'padded', quantity: 1 };
            inventoryItems.push({ item: 'dagger', quantity: 10 });
            break;
        case 'Paladin':
            equipmentSlots.mainHand = { item: 'longsword', quantity: 1 };
            equipmentSlots.offHand = { item: 'shield', quantity: 1 };
            equipmentSlots.armor = { item: 'chain-mail', quantity: 1 };
            inventoryItems.push({ item: 'emblem', quantity: 1 });
            break;
        case 'Ranger':
            equipmentSlots.mainHand = { item: 'longbow', quantity: 1 };
            equipmentSlots.armor = { item: 'leather', quantity: 1 };
            inventoryItems.push(
                { item: 'arrows', quantity: 20 },
                { item: 'shortsword', quantity: 1 }
            );
            break;
        case 'Rogue':
            equipmentSlots.mainHand = { item: 'rapier', quantity: 1 };
            equipmentSlots.armor = { item: 'leather', quantity: 1 };
            inventoryItems.push(
                { item: 'dagger', quantity: 2 },
                { item: 'thieves-tools', quantity: 1 }
            );
            break;
        case 'Sorcerer':
            equipmentSlots.mainHand = { item: 'crystal', quantity: 1 };
            equipmentSlots.armor = { item: 'padded', quantity: 1 };
            inventoryItems.push(
                { item: 'dagger', quantity: 2 },
                { item: 'component-pouch', quantity: 1 }
            );
            break;
        case 'Warlock':
            equipmentSlots.armor = { item: 'leather', quantity: 1 };
            inventoryItems.push(
                { item: 'dagger', quantity: 2 },
                { item: 'component-pouch', quantity: 1 },
                { item: 'cursed-bone-dust', quantity: 1 }
            );
            break;
        case 'Wizard':
            equipmentSlots.mainHand = { item: 'orb', quantity: 1 };
            inventoryItems.push(
                { item: 'dagger', quantity: 1 },
                { item: 'spellbook', quantity: 1 },
                { item: 'component-pouch', quantity: 1 }
            );
            break;
        default:
            // Default adventurer kit
            equipmentSlots.mainHand = { item: 'club', quantity: 1 };
            inventoryItems.push({ item: 'dagger', quantity: 1 });
    }

    return {
        equipped: equipmentSlots,
        inventory: inventoryItems
    };
}

// Generate starting spells based on class
function generateStartingSpells(characterClass) {
    const startingSpells = [];

    switch (characterClass) {
        case 'Wizard':
            startingSpells.push(
                { spell: 'fire-bolt', prepared: true },
                { spell: 'magic-missile', prepared: true },
                { spell: 'shield', prepared: true }
            );
            break;
        case 'Sorcerer':
            startingSpells.push(
                { spell: 'fire-bolt', prepared: true },
                { spell: 'magic-missile', prepared: true }
            );
            break;
        case 'Warlock':
            startingSpells.push(
                { spell: 'eldritch-blast', prepared: true },
                { spell: 'burning-hands', prepared: true }
            );
            break;
        case 'Bard':
            startingSpells.push(
                { spell: 'vicious-mockery', prepared: true },
                { spell: 'healing-word', prepared: true }
            );
            break;
        case 'Cleric':
            startingSpells.push(
                { spell: 'cure-wounds', prepared: true },
                { spell: 'sacred-flame', prepared: true },
                { spell: 'bless', prepared: true }
            );
            break;
        case 'Druid':
            startingSpells.push(
                { spell: 'cure-wounds', prepared: true },
                { spell: 'thunderwave', prepared: true }
            );
            break;
        case 'Paladin':
            // Paladins get spells at level 2, but we'll give them one
            startingSpells.push(
                { spell: 'cure-wounds', prepared: true }
            );
            break;
        case 'Ranger':
            // Rangers get spells at level 2, but we'll give them one
            startingSpells.push(
                { spell: 'cure-wounds', prepared: true }
            );
            break;
        default:
            // Non-spellcasters get no spells
            break;
    }

    return startingSpells;
}

// Generate spell slots based on class
function generateSpellSlots(characterClass) {
    // Basic spell slot structure for level 1 spellcasters
    const spellcasterSlots = {
        'Wizard': { cantrip: 2, level1: 2 },
        'Sorcerer': { cantrip: 2, level1: 2 },
        'Warlock': { cantrip: 2, level1: 1 },
        'Bard': { cantrip: 2, level1: 2 },
        'Cleric': { cantrip: 3, level1: 2 },
        'Druid': { cantrip: 2, level1: 2 },
        'Paladin': { level1: 0 }, // Get slots at level 2
        'Ranger': { level1: 0 }   // Get slots at level 2
    };

    return spellcasterSlots[characterClass] || {};
}

// Event listeners for game state changes
document.addEventListener('gameStateChange', async function(event) {
    await updateCharacterDisplay();  // Wait for character display to finish
    updateInventoryDisplay();
    updateSpellsDisplay();

    // Rebind inventory interactions after UI is updated
    if (window.inventoryInteractions && window.inventoryInteractions.bindInventoryEvents) {
        window.inventoryInteractions.bindInventoryEvents();
    }

    // Update combat interface if in combat
    const state = getGameState();
    if (state.combat) {
        updateCombatInterface();
    }
});

// Helper functions for character generation
function calculateMaxHP(characterClass, constitution, race) {
    // Base HP by class (D&D 5e style)
    const classHP = {
        'Barbarian': 12,
        'Fighter': 10, 'Paladin': 10, 'Ranger': 10,
        'Bard': 8, 'Cleric': 8, 'Druid': 8, 'Monk': 8, 'Rogue': 8, 'Warlock': 8,
        'Sorcerer': 6, 'Wizard': 6
    };

    // Constitution modifier
    const conMod = Math.floor((constitution - 10) / 2);

    // Base HP = class hit die + con modifier (minimum 1)
    const baseHP = (classHP[characterClass] || 8) + conMod;
    return Math.max(1, baseHP);
}

function calculateMaxMana(characterClass, stats) {
    // Only spellcasters have mana
    const casterClasses = {
        'Wizard': 'Intelligence',
        'Sorcerer': 'Charisma',
        'Warlock': 'Charisma',
        'Bard': 'Charisma',
        'Cleric': 'Wisdom',
        'Druid': 'Wisdom',
        'Paladin': 'Charisma', // Half-caster
        'Ranger': 'Wisdom'     // Half-caster
    };

    if (!casterClasses[characterClass]) {
        return 0; // Non-casters have no mana
    }

    const castingStat = casterClasses[characterClass];
    const statMod = Math.floor((stats[castingStat] - 10) / 2);

    // Half-casters get less mana
    const isHalfCaster = ['Paladin', 'Ranger'].includes(characterClass);
    const baseMana = isHalfCaster ? 2 : 4;

    return Math.max(1, baseMana + statMod);
}

function getStartingLocation(race) {
    // Starting locations based on race
    const startingLocations = {
        'Human': 'kingdom',
        'Elf': 'forest-kingdom',
        'Dwarf': 'hill-kingdom',
        'Halfling': 'village-south',
        'Gnome': 'village-west',
        'Dragonborn': 'mountain-northeast',
        'Tiefling': 'city-east',
        'Half-Elf': 'town-north',
        'Half-Orc': 'town-northeast',
        'Orc': 'swamp-kingdom'
    };

    return startingLocations[race] || 'kingdom';
}

function generateCharacterName(race, background) {
    // Simple name generation based on race and background
    const raceNames = {
        'Human': ['Aiden', 'Emma', 'Gareth', 'Luna', 'Marcus', 'Sera'],
        'Elf': ['Aeliana', 'Caelynn', 'Erevan', 'Silvyr', 'Thalion', 'Vaelish'],
        'Dwarf': ['Balin', 'Daina', 'Gimli', 'Nala', 'Thorin', 'Vera'],
        'Halfling': ['Bilbo', 'Daisy', 'Frodo', 'Poppy', 'Sam', 'Rose'],
        'Gnome': ['Bimpkin', 'Dimble', 'Glim', 'Nimble', 'Pip', 'Zook'],
        'Dragonborn': ['Arjhan', 'Balasar', 'Bharash', 'Donaar', 'Ghesh', 'Heskan'],
        'Tiefling': ['Akmen', 'Amnon', 'Barakas', 'Damakos', 'Ekemon', 'Iados'],
        'Half-Elf': ['Aramil', 'Berris', 'Dayereth', 'Enna', 'Galinndan', 'Hadarai'],
        'Half-Orc': ['Dench', 'Feng', 'Gell', 'Henk', 'Holg', 'Imsh'],
        'Orc': ['Dench', 'Feng', 'Gell', 'Henk', 'Holg', 'Imsh']
    };

    const names = raceNames[race] || raceNames['Human'];
    return names[Math.floor(Math.random() * names.length)];
}

function generateStartingGold(characterClass, background) {
    // Starting gold by class (5e style)
    const classGold = {
        'Barbarian': 50, 'Fighter': 125, 'Paladin': 125, 'Ranger': 125,
        'Bard': 125, 'Cleric': 125, 'Druid': 50, 'Monk': 12, 'Rogue': 100, 'Warlock': 100,
        'Sorcerer': 75, 'Wizard': 100
    };

    // Background might modify starting gold
    const backgroundBonus = {
        'Noble': 50, 'Merchant': 25, 'Criminal': 15, 'Hermit': 5
    };

    const baseGold = classGold[characterClass] || 75;
    const bonus = backgroundBonus[background] || 0;

    return baseGold + bonus;
}

// Show character introduction with story and stats
function showCharacterIntro(character) {
    const gameContainer = document.getElementById('game-app');
    if (!gameContainer) return;

    // Generate character story based on background
    const backgroundStories = {
        'Acolyte': "You spent your youth in service to a temple, monastery, or other religious community. Your faith has prepared you for the adventures ahead.",
        'Criminal': "You have always thumbed your nose at the rules. A life of crime has taught you to rely on your wits and reflexes.",
        'Folk Hero': "You come from humble social rank, but you are destined for so much more. The common folk see you as their champion.",
        'Noble': "You were born to rule, with wealth, power, and prestige from birth. Noblesse oblige guides your actions.",
        'Sage': "You spent years learning the lore of the multiverse. You scoured manuscripts, studied scrolls, and listened to ancient wisdom.",
        'Soldier': "War has been your life for as long as you care to remember. You trained as a youth, studied strategy, and lived the life of a warrior.",
        'Hermit': "You lived in seclusion for years, seeking enlightenment and understanding. Now you emerge to face the world with newfound wisdom.",
        'Entertainer': "You thrive in front of an audience, whether telling stories, singing songs, or performing dramatic acts.",
        'Merchant': "You grew up in the world of commerce, learning to assess value and opportunity with a keen eye.",
        'Charlatan': "You have always had a way with people, using wit and charm to get what you want through trickery and misdirection."
    };

    const story = backgroundStories[character.background] || "Your past has shaped you in unique ways, preparing you for the adventures that lie ahead.";

    gameContainer.innerHTML = `
        <div class="max-w-4xl mx-auto p-6">
            <div class="text-center mb-8">
                <h1 class="text-4xl font-bold text-yellow-400 mb-4">🎭 Your Hero Awakens 🎭</h1>
                <p class="text-xl text-gray-300">Born from your Nostr identity</p>
            </div>

            <div class="grid md:grid-cols-2 gap-8">
                <!-- Character Portrait & Basic Info -->
                <div class="bg-gray-800 rounded-lg p-6 border border-yellow-400">
                    <div class="text-center mb-6">
                        <div class="text-6xl mb-4">${getCharacterIcon(character.race, character.class)}</div>
                        <h2 class="text-2xl font-bold text-green-400">${character.name}</h2>
                        <p class="text-lg text-gray-300">${character.race} ${character.class}</p>
                        <p class="text-md text-gray-400">${character.background} • ${character.alignment}</p>
                    </div>

                    <!-- Stats Display -->
                    <div class="grid grid-cols-2 gap-3 text-sm">
                        ${Object.entries(character.stats).map(([stat, value]) => `
                            <div class="flex justify-between bg-gray-700 p-2 rounded">
                                <span class="font-medium">${stat}:</span>
                                <span class="font-bold ${value >= 15 ? 'text-green-400' : value >= 12 ? 'text-yellow-400' : value <= 8 ? 'text-red-400' : 'text-gray-300'}">${value}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>

                <!-- Character Story & Details -->
                <div class="space-y-6">
                    <!-- Background Story -->
                    <div class="bg-gray-800 rounded-lg p-6 border border-purple-400">
                        <h3 class="text-xl font-bold text-purple-400 mb-3">📜 Your Past</h3>
                        <p class="text-gray-300 leading-relaxed">${story}</p>
                    </div>

                    <!-- Starting Resources -->
                    <div class="bg-gray-800 rounded-lg p-6 border border-green-400">
                        <h3 class="text-xl font-bold text-green-400 mb-3">⚔️ Starting Resources</h3>
                        <div class="grid grid-cols-2 gap-4 text-sm">
                            <div>
                                <span class="font-medium text-red-400">Health:</span>
                                <span class="text-white ml-2">${character.hp} HP</span>
                            </div>
                            <div>
                                <span class="font-medium text-blue-400">Mana:</span>
                                <span class="text-white ml-2">${character.mana} MP</span>
                            </div>
                            <div>
                                <span class="font-medium text-yellow-400">Gold:</span>
                                <span class="text-white ml-2">${character.gold} coins</span>
                            </div>
                            <div>
                                <span class="font-medium text-gray-400">Location:</span>
                                <span class="text-white ml-2">${character.starting_location.replace('-', ' ')}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Continue Button -->
            <div class="text-center mt-8">
                <button onclick="beginAdventure()"
                        class="bg-gradient-to-r from-green-600 to-blue-600 hover:from-green-700 hover:to-blue-700 text-white px-8 py-4 rounded-lg text-xl font-bold transition-all duration-300 shadow-lg hover:shadow-xl">
                    ⚔️ Begin Your Adventure ⚔️
                </button>
            </div>
        </div>
    `;
}

// Begin the actual adventure (show main game interface)
async function beginAdventure() {
    console.log('🎮 Beginning adventure...');

    // Use the stored state from character creation
    if (!pendingCharacterState) {
        console.error('No pending character state found!');
        showMessage('❌ Error: Character state not found', 'error');
        return;
    }

    const gameState = pendingCharacterState;
    console.log('🔍 PENDING STATE:', JSON.stringify({
        vault: gameState.character.vault,
        current_day: gameState.character.current_day,
        time_of_day: gameState.character.time_of_day,
        location: gameState.location
    }, null, 2));

    // Save the initial game state to create save file
    try {
        const session = window.sessionManager.getSession();

        // Structure inventory properly (combining general slots and gear slots)
        const inventoryStructure = {
            general_slots: gameState.inventory || [],
            gear_slots: gameState.equipment || {}
        };

        // Convert location IDs to display names for save file
        console.log('🔍 Converting location IDs:', {
            locationId: gameState.location?.current,
            districtKey: gameState.location?.district,
            buildingId: gameState.location?.building
        });

        const displayNames = await getDisplayNamesForLocation(
            gameState.location?.current || 'kingdom',
            gameState.location?.district || 'center',
            gameState.location?.building || ''
        );

        console.log('🔍 Display names:', displayNames);
        console.log('🔍 Vault data:', gameState.character.vault);

        // DEBUG: Check values before creating save data
        const debugInfo = {
            raw_vault: gameState.character.vault,
            raw_current_day: gameState.character.current_day,
            raw_time_of_day: gameState.character.time_of_day,
            raw_district: gameState.location?.district
        };

        // Prepare save data with all required fields (using display names)
        const saveData = {
            d: gameState.character.name || '',
            race: gameState.character.race || '',
            class: gameState.character.class || '',
            background: gameState.character.background || '',
            alignment: gameState.character.alignment || '',
            experience: gameState.character.experience || 0,
            hp: gameState.character.hp || 0,
            max_hp: gameState.character.max_hp || 0,
            mana: gameState.character.mana || 0,
            max_mana: gameState.character.max_mana || 0,
            fatigue: gameState.character.fatigue || 0,
            gold: gameState.character.gold || 0,
            stats: gameState.character.stats || {},
            location: displayNames.location,
            district: displayNames.district,
            building: displayNames.building || '',  // Empty string for outdoors
            inventory: inventoryStructure,
            vault: gameState.character.vault || null,
            known_spells: gameState.spells || [],
            spell_slots: gameState.character.spell_slots || {},
            locations_discovered: gameState.location?.discovered || [],
            music_tracks_unlocked: gameState.character.music_tracks_unlocked || [],
            current_day: gameState.character.current_day !== undefined ? gameState.character.current_day : 1,
            time_of_day: gameState.character.time_of_day || 'day',
            _debug: debugInfo  // Temporary debug field
        };

        console.log('💾 Creating initial save file:', saveData);

        // Create the save file
        const response = await fetch(`/api/saves/${session.npub}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(saveData)
        });

        if (!response.ok) {
            throw new Error(`Failed to create save file: ${response.status}`);
        }

        const result = await response.json();
        console.log('✅ Initial save file created:', result);

        // Update URL to include save ID
        const saveID = result.save_id;
        const newURL = `/game?save=${saveID}`;
        window.history.replaceState({}, '', newURL);

        showMessage('💾 Game saved! Your adventure begins!', 'success');

        // Now update the game state in DOM with the saved state
        updateGameState(gameState);

        // Clear pending state
        pendingCharacterState = null;

    } catch (error) {
        console.error('❌ Failed to create initial save:', error);
        showMessage('⚠️ Warning: Failed to create save file. Progress may not be saved.', 'warning');
    }

    // Show main game interface
    displayCurrentLocation();
    updateAllDisplays();
}

// Helper function to get character icon (matching the saves page)
function getCharacterIcon(race, characterClass) {
    const raceIcons = {
        'Human': '👤',
        'Elf': '🧝',
        'Dwarf': '🧔',
        'Halfling': '🧒',
        'Dragonborn': '🐲',
        'Gnome': '🧙',
        'Half-Elf': '👨‍🎤',
        'Half-Orc': '👹',
        'Tiefling': '😈',
        'Orc': '👹'
    };

    return raceIcons[race] || '⚔️';
}

// Generate starting vault (40 slots, empty, city-based)
function generateStartingVault(location) {
    // Create 40 empty slots
    const vaultSlots = [];
    for (let i = 0; i < 40; i++) {
        vaultSlots.push({
            slot: i,
            item: null,
            quantity: 0
        });
    }

    return {
        location: location,                          // City ID
        building: getVaultBuildingForLocation(location), // Building ID for house_of_keeping
        slots: vaultSlots
    };
}

// Extract district from location
function getDistrictFromLocation(location) {
    // All characters start in the center district of their starting city
    return 'center';
}

// Get vault building ID for a given location
function getVaultBuildingForLocation(location) {
    // Map cities to their house_of_keeping building IDs
    const vaultBuildings = {
        'kingdom': 'vault_of_crowns',
        'village-west': 'burrowlock',
        'village-south': 'halfling_burrows',
        'village-southeast': 'secure_cellars',
        'village-southwest': 'stone_vaults',
        'town-north': 'northwatch_vault',
        'town-northeast': 'stormhold_storage',
        'city-east': 'shadowhaven_vaults',
        'city-south': 'coastal_storage',
        'forest-kingdom': 'silverwood_treasury',
        'hill-kingdom': 'ironforge_vaults',
        'mountain-northeast': 'draconis_hoard',
        'swamp-kingdom': 'mire_keep_storage'
    };

    return vaultBuildings[location] || 'vault_of_crowns'; // Default to kingdom vault
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

console.log('Game state management loaded');