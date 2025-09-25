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

// Get current character's npub from session manager
function getCurrentNpub() {
    if (window.sessionManager && window.sessionManager.isAuthenticated()) {
        return window.sessionManager.getNpub();
    }
    return null;
}

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

                    if (save && save.gameState) {
                        initializeFromSave(save.gameState);
                        showMessage('✅ Save file loaded successfully!', 'success');
                    } else {
                        throw new Error('Save file not found or corrupted');
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

        // Start game UI
        displayCurrentLocation();
        updateAllDisplays();

        // Start auto-save system
        if (typeof startAutoSave === 'function') {
            startAutoSave();
        }

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

        const result = await characterResponse.json();
        console.log('Generated character result:', result);

        const characterData = result.character;

        // Calculate derived stats based on class and race
        const maxHP = calculateMaxHP(characterData.class, characterData.stats.Constitution, characterData.race);
        const maxMana = calculateMaxMana(characterData.class, characterData.stats);
        const startingLocation = getStartingLocation(characterData.race);

        // Initialize complete character with all required fields
        const fullCharacter = {
            ...characterData,
            name: generateCharacterName(characterData.race, characterData.background),
            level: 1,
            experience: 0,
            hp: maxHP,
            max_hp: maxHP,
            mana: maxMana,
            max_mana: maxMana,
            fatigue: 0,
            gold: generateStartingGold(characterData.class, characterData.background),
            starting_location: startingLocation
        };

        // Initialize fresh game state
        const initialState = {
            character: fullCharacter,
            inventory: generateStartingInventory(characterData.class),
            spells: generateStartingSpells(characterData.class),
            location: {
                current: startingLocation,
                discovered: [startingLocation]
            },
            combat: null
        };

        updateGameState(initialState);
        console.log('Initialized game state:', initialState);

        // Show character intro before starting the game
        showCharacterIntro(fullCharacter);

    } catch (error) {
        console.error('Failed to create character:', error);
        throw error;
    }
}

// Generate starting inventory based on class
function generateStartingInventory(characterClass) {
    // Base survival items for all characters
    const baseItems = [
        { item: 'rations-1-day', quantity: 3 },
        { item: 'bedroll', quantity: 1 },
        { item: 'waterskin', quantity: 1 },
        { item: 'rope-hempen-50-feet', quantity: 1 },
        { item: 'backpack', quantity: 1 }
    ];

    // Add class-specific starting equipment
    switch (characterClass) {
        case 'Barbarian':
            baseItems.push(
                { item: 'greataxe', quantity: 1 },
                { item: 'handaxe', quantity: 2 },
                { item: 'hide', quantity: 1 }
            );
            break;
        case 'Bard':
            baseItems.push(
                { item: 'rapier', quantity: 1 },
                { item: 'leather', quantity: 1 },
                { item: 'lute', quantity: 1 },
                { item: 'dagger', quantity: 1 }
            );
            break;
        case 'Cleric':
            baseItems.push(
                { item: 'mace', quantity: 1 },
                { item: 'shield', quantity: 1 },
                { item: 'chain-mail', quantity: 1 },
                { item: 'reliquary', quantity: 1 }
            );
            break;
        case 'Druid':
            baseItems.push(
                { item: 'quarterstaff', quantity: 1 },
                { item: 'leather', quantity: 1 },
                { item: 'shield', quantity: 1 },
                { item: 'sprig-of-mistletoe', quantity: 1 }
            );
            break;
        case 'Fighter':
            baseItems.push(
                { item: 'longsword', quantity: 1 },
                { item: 'shield', quantity: 1 },
                { item: 'chain-mail', quantity: 1 },
                { item: 'handaxe', quantity: 2 }
            );
            break;
        case 'Monk':
            baseItems.push(
                { item: 'quarterstaff', quantity: 1 },
                { item: 'dagger', quantity: 10 },
                { item: 'padded', quantity: 1 }
            );
            break;
        case 'Paladin':
            baseItems.push(
                { item: 'longsword', quantity: 1 },
                { item: 'shield', quantity: 1 },
                { item: 'chain-mail', quantity: 1 },
                { item: 'emblem', quantity: 1 }
            );
            break;
        case 'Ranger':
            baseItems.push(
                { item: 'longbow', quantity: 1 },
                { item: 'arrows', quantity: 20 },
                { item: 'shortsword', quantity: 1 },
                { item: 'leather', quantity: 1 }
            );
            break;
        case 'Rogue':
            baseItems.push(
                { item: 'rapier', quantity: 1 },
                { item: 'dagger', quantity: 2 },
                { item: 'thieves-tools', quantity: 1 },
                { item: 'leather', quantity: 1 }
            );
            break;
        case 'Sorcerer':
            baseItems.push(
                { item: 'dagger', quantity: 2 },
                { item: 'component-pouch', quantity: 1 },
                { item: 'crystal', quantity: 1 },
                { item: 'padded', quantity: 1 }
            );
            break;
        case 'Warlock':
            baseItems.push(
                { item: 'dagger', quantity: 2 },
                { item: 'component-pouch', quantity: 1 },
                { item: 'leather', quantity: 1 },
                { item: 'cursed-bone-dust', quantity: 1 }
            );
            break;
        case 'Wizard':
            baseItems.push(
                { item: 'dagger', quantity: 1 },
                { item: 'spellbook', quantity: 1 },
                { item: 'component-pouch', quantity: 1 },
                { item: 'orb', quantity: 1 }
            );
            break;
        default:
            // Default adventurer kit
            baseItems.push(
                { item: 'club', quantity: 1 },
                { item: 'dagger', quantity: 1 }
            );
    }

    return baseItems;
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
        'Human': 'kingdom-center',
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

    return startingLocations[race] || 'kingdom-center';
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
function beginAdventure() {
    console.log('🎮 Beginning adventure...');

    // Show main game interface
    displayCurrentLocation();
    updateAllDisplays();

    showMessage('🎉 Your adventure begins!', 'success');
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

console.log('Game state management loaded');