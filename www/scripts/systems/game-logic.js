// Game Logic Functions
// Core gameplay mechanics following the DOM state management approach

// Movement and exploration functions
async function moveToLocation(locationId) {
    console.log('Moving to location:', locationId);

    const state = getGameState();
    const locationData = getLocationById(locationId);

    if (!locationData) {
        showMessage('❌ Unknown location: ' + locationId, 'error');
        return;
    }

    // Check if this is an environment (outside city) - block it
    if (locationData.location_type === 'environment') {
        showTravelDisabledPopup(locationId);
        return;
    }

    // Parse the locationId to extract city and district
    // Format: "city-id-districtKey" (e.g., "village-west-center")
    const { cityId, districtKey } = parseCityAndDistrict(locationId);

    console.log('📍 Parsed location:', { locationId, cityId, districtKey });

    // Allow travel within city (districts)
    // Update location state with parsed city and district
    const newLocationState = {
        current: cityId,         // Store just the city ID
        district: districtKey,   // Store just the district key
        building: '',            // Default to outdoors
        discovered: [...state.location.discovered]
    };

    if (!newLocationState.discovered.includes(locationId)) {
        newLocationState.discovered.push(locationId);
        showMessage('🗺️ Discovered: ' + locationData.name, 'success');
    }

    // Handle travel effects (fatigue, time passage, etc.)
    const newCharacterState = { ...state.character };

    // Update game state
    updateGameState({
        location: newLocationState,
        character: newCharacterState
    });

    // Update display
    displayCurrentLocation();
    showMessage('📍 Moved to ' + locationData.name, 'info');

    // Save the location change to backend
    if (window.saveGameToLocal) {
        await window.saveGameToLocal();
    }
}

// Parse location ID to extract city ID and district key
// Format: "city-id-districtKey" → {cityId: "city-id", districtKey: "districtKey"}
function parseCityAndDistrict(locationId) {
    // Handle simple city IDs (no district)
    if (!locationId.includes('-')) {
        return { cityId: locationId, districtKey: 'center' };
    }

    // Find the last hyphen to separate district key
    const lastHyphenIndex = locationId.lastIndexOf('-');
    const cityId = locationId.substring(0, lastHyphenIndex);
    const districtKey = locationId.substring(lastHyphenIndex + 1);

    return { cityId, districtKey };
}

// Show popup when trying to travel (work-in-progress feature)
function showTravelDisabledPopup(locationId) {
    const locationData = getLocationById(locationId);
    const locationName = locationData ? locationData.name : locationId;

    // Create modal backdrop
    const backdrop = document.createElement('div');
    backdrop.id = 'travel-disabled-backdrop';
    backdrop.className = 'fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-[9999]';
    backdrop.style.fontFamily = '"Dogica", monospace';

    // Create modal content
    const modal = document.createElement('div');
    modal.className = 'bg-gray-800 rounded-lg p-6 max-w-md mx-4 relative';
    modal.style.border = '3px solid #ef4444';
    modal.style.boxShadow = '0 0 20px rgba(239, 68, 68, 0.5)';

    modal.innerHTML = `
        <h2 class="text-xl font-bold text-red-400 mb-4 text-center">🚧 Travel Not Available</h2>

        <div class="text-gray-300 space-y-3 text-sm leading-relaxed">
            <p>
                You can't travel to <span class="text-yellow-400 font-bold">${locationName}</span> yet!
            </p>

            <div class="bg-gray-900 border border-red-600 rounded p-3 text-xs">
                <p class="text-red-300 mb-2">⚠️ This feature is coming soon</p>
                <p class="text-gray-400">The game UI is still in development. Travel and exploration mechanics will be added in future updates.</p>
            </div>

            <p class="text-center text-xs text-gray-500 mt-4">
                For now, you can only view your character's stats and inventory from the intro.
            </p>
        </div>

        <div class="mt-4 text-center">
            <button
                id="travel-close-btn"
                class="px-6 py-2 bg-red-600 hover:bg-red-500 text-white font-bold rounded-lg transition-colors"
                style="font-size: 0.875rem;">
                Okay
            </button>
        </div>
    `;

    backdrop.appendChild(modal);
    document.body.appendChild(backdrop);

    // Close button handler
    document.getElementById('travel-close-btn').onclick = () => {
        backdrop.remove();
    };

    // Close on backdrop click
    backdrop.onclick = (e) => {
        if (e.target === backdrop) {
            backdrop.remove();
        }
    };
}

// Item usage functions
function useItem(itemId) {
    console.log('Using item:', itemId);

    const state = getGameState();
    const item = findInInventory(state.inventory, itemId);
    const itemData = getItemById(itemId);

    if (!item || item.quantity <= 0) {
        showMessage("❌ You don't have that item!", 'error');
        return;
    }

    if (!itemData) {
        showMessage("❌ Unknown item: " + itemId, 'error');
        return;
    }

    const newCharacterState = { ...state.character };
    const newInventory = [...state.inventory];

    // Apply item effects based on item properties
    if (itemData.properties) {
        if (itemData.properties.heal) {
            const healAmount = parseInt(itemData.properties.heal);
            newCharacterState.hp = Math.min(
                newCharacterState.max_hp,
                newCharacterState.hp + healAmount
            );
            showMessage(`💚 Healed ${healAmount} HP!`, 'success');
        }

        if (itemData.properties.mana_restore) {
            const manaAmount = parseInt(itemData.properties.mana_restore);
            newCharacterState.mana = Math.min(
                newCharacterState.max_mana || 0,
                newCharacterState.mana + manaAmount
            );
            showMessage(`💙 Restored ${manaAmount} mana!`, 'success');
        }

        if (itemData.properties.reduce_fatigue) {
            const fatigueReduction = parseInt(itemData.properties.reduce_fatigue);
            newCharacterState.fatigue = Math.max(0, newCharacterState.fatigue - fatigueReduction);
            showMessage('😌 Feeling less tired', 'success');
        }
    }

    // Handle rations specially (reduce fatigue and provide sustenance)
    if (itemId === 'rations') {
        newCharacterState.fatigue = Math.max(0, newCharacterState.fatigue - 1);
        showMessage('🍖 You feel nourished and rested', 'success');
    }

    // Remove/reduce item from inventory
    const itemIndex = newInventory.findIndex(i => i.item === itemId);
    if (itemIndex !== -1) {
        newInventory[itemIndex].quantity -= 1;
        if (newInventory[itemIndex].quantity === 0) {
            newInventory.splice(itemIndex, 1);
        }
    }

    // Update game state
    updateGameState({
        character: newCharacterState,
        inventory: newInventory
    });

    showMessage(`Used ${itemData.name}!`, 'info');
}

// Spell casting functions
function castSpell(spellId) {
    console.log('Casting spell:', spellId);

    const state = getGameState();
    const spell = findInSpells(state.spells, spellId);
    const spellData = getSpellById(spellId);

    if (!spell || !spell.prepared) {
        showMessage("❌ Spell not prepared!", 'error');
        return;
    }

    if (!spellData) {
        showMessage("❌ Unknown spell: " + spellId, 'error');
        return;
    }

    const manaCost = spellData.mana_cost || 0;
    if (state.character.mana < manaCost) {
        showMessage("❌ Not enough mana!", 'error');
        return;
    }

    const newCharacterState = { ...state.character };
    newCharacterState.mana -= manaCost;

    // Apply spell effects
    if (state.combat) {
        handleCombatSpell(spellId, spellData);
    } else {
        handleNonCombatSpell(spellId, spellData, newCharacterState);
    }

    // Update game state
    updateGameState({
        character: newCharacterState
    });

    showMessage(`✨ Cast ${spellData.name}!`, 'success');
}

// Handle non-combat spell effects
function handleNonCombatSpell(spellId, spellData, characterState) {
    switch (spellId) {
        case 'cure-wounds':
        case 'healing-word':
            const healAmount = rollDice(spellData.damage || '1d8') + 3; // Simplified healing
            characterState.hp = Math.min(characterState.max_hp, characterState.hp + healAmount);
            showMessage(`💚 Healed ${healAmount} HP!`, 'success');
            break;

        case 'mage-armor':
            // Temporary AC boost (would need to implement buff system)
            showMessage('🛡️ Magical armor surrounds you!', 'success');
            break;

        default:
            showMessage(`✨ ${spellData.name} effect applied!`, 'info');
    }
}

// Combat spell handling (basic implementation)
function handleCombatSpell(spellId, spellData) {
    const state = getGameState();
    if (!state.combat) return;

    // This would integrate with your combat system
    console.log('Combat spell:', spellId, spellData);
    showMessage(`⚔️ ${spellData.name} used in combat!`, 'success');
}

// Rest system
function restAtLocation() {
    const state = getGameState();
    const locationData = getLocationById(state.location.current);

    if (!locationData) return;

    // Check if resting is possible
    const hasRations = findInInventory(state.inventory, 'rations');
    const hasBedroll = findInInventory(state.inventory, 'bedroll');

    if (!hasRations || hasRations.quantity <= 0) {
        showMessage("❌ You need rations to rest!", 'error');
        return;
    }

    const newCharacterState = { ...state.character };
    const newInventory = [...state.inventory];

    // Rest effects
    if (hasBedroll && hasBedroll.quantity > 0) {
        // Full rest with bedroll
        newCharacterState.fatigue = Math.max(0, newCharacterState.fatigue - 2);
        newCharacterState.hp = Math.min(newCharacterState.max_hp, newCharacterState.hp + Math.floor(newCharacterState.max_hp / 2));
        showMessage('😴 You rest comfortably and feel refreshed!', 'success');
    } else {
        // Poor rest without bedroll
        newCharacterState.fatigue = Math.max(0, newCharacterState.fatigue - 1);
        newCharacterState.hp = Math.min(newCharacterState.max_hp, newCharacterState.hp + Math.floor(newCharacterState.max_hp / 4));
        showMessage('😪 You rest uncomfortably but recover slightly', 'warning');
    }

    // Consume rations
    const rationIndex = newInventory.findIndex(i => i.item === 'rations');
    if (rationIndex !== -1) {
        newInventory[rationIndex].quantity -= 1;
        if (newInventory[rationIndex].quantity === 0) {
            newInventory.splice(rationIndex, 1);
        }
    }

    // Restore some mana
    newCharacterState.mana = Math.min(
        newCharacterState.max_mana || 0,
        newCharacterState.mana + Math.floor((newCharacterState.max_mana || 0) / 2)
    );

    updateGameState({
        character: newCharacterState,
        inventory: newInventory
    });
}

// Simple combat initiation
function initiateRandomEncounter() {
    const state = getGameState();
    const locationData = getLocationById(state.location.current);

    if (state.combat) {
        showMessage("❌ Already in combat!", 'error');
        return;
    }

    // Simple random encounter based on location
    const monsters = ['goblin', 'wolf', 'bandit']; // Simplified monster list
    const randomMonster = monsters[Math.floor(Math.random() * monsters.length)];
    const monsterData = getMonsterById(randomMonster);

    if (!monsterData) {
        showMessage("🌙 The area is quiet...", 'info');
        return;
    }

    // Initialize combat
    const combatState = {
        monster: {
            id: randomMonster,
            name: monsterData.name,
            hp: monsterData.stats?.hit_points || 10,
            max_hp: monsterData.stats?.hit_points || 10,
            ac: monsterData.stats?.armor_class || 12
        },
        distance: 3, // Start at medium range
        player_turn: true
    };

    updateGameState({ combat: combatState });
    showMessage(`⚔️ A ${monsterData.name} appears!`, 'warning');
    updateCombatInterface();
}

// Simple dice rolling utility
function rollDice(diceString) {
    // Parse strings like "2d6+3", "1d8", etc.
    const match = diceString.match(/(\d+)d(\d+)(?:\+(\d+))?/);
    if (!match) return 1;

    const numDice = parseInt(match[1]);
    const dieSize = parseInt(match[2]);
    const modifier = parseInt(match[3]) || 0;

    let total = 0;
    for (let i = 0; i < numDice; i++) {
        total += Math.floor(Math.random() * dieSize) + 1;
    }

    return total + modifier;
}

// Equipment management
function equipItem(itemId) {
    const state = getGameState();
    const itemData = getItemById(itemId);
    const inventoryItem = findInInventory(state.inventory, itemId);

    if (!inventoryItem || !itemData) {
        showMessage("❌ Cannot equip that item!", 'error');
        return;
    }

    // This would integrate with your equipment system
    showMessage(`🎒 Equipped ${itemData.name}!`, 'success');
}

// Shop interactions (basic)
function buyItem(itemId, cost) {
    const state = getGameState();

    if ((state.character.gold || 0) < cost) {
        showMessage("❌ Not enough gold!", 'error');
        return;
    }

    const itemData = getItemById(itemId);
    if (!itemData) {
        showMessage("❌ Unknown item!", 'error');
        return;
    }

    const newCharacterState = { ...state.character };
    newCharacterState.gold = (newCharacterState.gold || 0) - cost;

    const newInventory = [...state.inventory];
    const existingItem = newInventory.find(i => i.item === itemId);

    if (existingItem) {
        existingItem.quantity += 1;
    } else {
        newInventory.push({ item: itemId, quantity: 1 });
    }

    updateGameState({
        character: newCharacterState,
        inventory: newInventory
    });

    showMessage(`💰 Bought ${itemData.name} for ${cost} gold!`, 'success');
}

console.log('Game logic functions loaded');