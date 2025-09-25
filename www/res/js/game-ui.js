// Game UI Functions
// Handles all user interface updates and interactions

// Message system
function showMessage(text, type = 'info', duration = 5000) {
    const messageArea = document.getElementById('message-area');

    const messageDiv = document.createElement('div');
    messageDiv.className = `p-3 rounded-lg shadow-lg transform transition-all duration-300 translate-x-full`;

    // Set color based on type
    switch (type) {
        case 'error':
            messageDiv.className += ' bg-red-600 text-white';
            break;
        case 'success':
            messageDiv.className += ' bg-green-600 text-white';
            break;
        case 'warning':
            messageDiv.className += ' bg-yellow-600 text-gray-900';
            break;
        case 'info':
        default:
            messageDiv.className += ' bg-blue-600 text-white';
            break;
    }

    messageDiv.textContent = text;
    messageArea.appendChild(messageDiv);

    // Animate in
    setTimeout(() => {
        messageDiv.classList.remove('translate-x-full');
    }, 10);

    // Auto-remove after duration
    setTimeout(() => {
        messageDiv.classList.add('translate-x-full');
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.parentNode.removeChild(messageDiv);
            }
        }, 300);
    }, duration);
}

// Update character display
function updateCharacterDisplay() {
    const state = getGameState();
    const character = state.character;

    if (!character) return;

    // Update HP display
    document.getElementById('current-hp').textContent = character.hp || 0;
    document.getElementById('max-hp').textContent = character.max_hp || 0;

    // Update HP bar
    const hpPercentage = (character.max_hp > 0) ? (character.hp / character.max_hp * 100) : 0;
    document.getElementById('hp-bar').style.width = hpPercentage + '%';

    // Update mana display
    document.getElementById('current-mana').textContent = character.mana || 0;
    document.getElementById('max-mana').textContent = character.max_mana || 0;

    // Update mana bar
    const manaPercentage = (character.max_mana > 0) ? (character.mana / character.max_mana * 100) : 0;
    document.getElementById('mana-bar').style.width = manaPercentage + '%';

    // Update fatigue
    document.getElementById('fatigue-level').textContent = character.fatigue || 0;
}

// Update inventory display
function updateInventoryDisplay() {
    const state = getGameState();
    const inventory = state.inventory || [];
    const inventoryList = document.getElementById('inventory-list');

    if (!inventoryList) return;

    inventoryList.innerHTML = '';

    if (inventory.length === 0) {
        inventoryList.innerHTML = '<div class="text-gray-500 text-sm">No items</div>';
        return;
    }

    inventory.forEach(item => {
        const itemData = getItemById(item.item);
        const itemDiv = document.createElement('div');
        itemDiv.className = 'flex justify-between items-center p-2 bg-gray-700 rounded text-sm hover:bg-gray-600 cursor-pointer';

        const itemName = itemData ? itemData.name : item.item;
        itemDiv.innerHTML = `
            <span>${itemName}</span>
            <span class="text-yellow-400">${item.quantity}</span>
        `;

        // Add click handler to use item
        itemDiv.addEventListener('click', () => {
            useItem(item.item);
        });

        inventoryList.appendChild(itemDiv);
    });
}

// Update spells display
function updateSpellsDisplay() {
    const state = getGameState();
    const spells = state.spells || [];
    const spellsList = document.getElementById('spells-list');

    if (!spellsList) return;

    spellsList.innerHTML = '';

    const preparedSpells = spells.filter(spell => spell.prepared);

    if (preparedSpells.length === 0) {
        spellsList.innerHTML = '<div class="text-gray-500 text-sm">No spells prepared</div>';
        return;
    }

    preparedSpells.forEach(spell => {
        const spellData = getSpellById(spell.spell);
        const spellDiv = document.createElement('div');
        spellDiv.className = 'flex justify-between items-center p-2 bg-gray-700 rounded text-sm hover:bg-gray-600 cursor-pointer';

        const spellName = spellData ? spellData.name : spell.spell;
        const manaCost = spellData ? spellData.mana_cost || 0 : 0;

        spellDiv.innerHTML = `
            <span>${spellName}</span>
            <span class="text-blue-400">${manaCost} MP</span>
        `;

        // Add click handler to cast spell
        spellDiv.addEventListener('click', () => {
            castSpell(spell.spell);
        });

        spellsList.appendChild(spellDiv);
    });
}

// Display current location
function displayCurrentLocation() {
    const state = getGameState();
    const currentLocationId = state.location?.current;

    if (!currentLocationId) return;

    const locationData = getLocationById(currentLocationId);
    if (!locationData) return;

    // Update location display
    const locationName = document.getElementById('location-name');
    const locationDescription = document.getElementById('location-description');
    const locationActions = document.getElementById('location-actions');

    if (locationName) locationName.textContent = locationData.name;
    if (locationDescription) locationDescription.textContent = locationData.description || 'A mysterious place...';

    // Generate location actions
    if (locationActions) {
        locationActions.innerHTML = '';

        // Add movement options based on connections
        if (locationData.properties?.connections) {
            locationData.properties.connections.forEach(connectionId => {
                const connectedLocation = getLocationById(connectionId);
                if (connectedLocation) {
                    const actionButton = createActionButton(
                        `→ ${connectedLocation.name}`,
                        () => moveToLocation(connectionId),
                        'bg-blue-600 hover:bg-blue-700'
                    );
                    locationActions.appendChild(actionButton);
                }
            });
        }

        // Add location-specific actions
        if (locationData.location_type === 'city' || locationData.location_type === 'town') {
            // City actions
            locationActions.appendChild(
                createActionButton('🛍️ Visit Shop', () => openShop(), 'bg-green-600 hover:bg-green-700')
            );
            locationActions.appendChild(
                createActionButton('🍺 Enter Tavern', () => openTavern(), 'bg-amber-600 hover:bg-amber-700')
            );
        }

        // Universal actions
        locationActions.appendChild(
            createActionButton('😴 Rest', () => restAtLocation(), 'bg-purple-600 hover:bg-purple-700')
        );

        locationActions.appendChild(
            createActionButton('👹 Random Encounter', () => initiateRandomEncounter(), 'bg-red-600 hover:bg-red-700')
        );
    }
}

// Create an action button
function createActionButton(text, onClick, classes = 'bg-gray-600 hover:bg-gray-700') {
    const button = document.createElement('button');
    button.className = `${classes} text-white px-4 py-2 rounded text-sm font-medium transition-colors`;
    button.textContent = text;
    button.addEventListener('click', onClick);
    return button;
}

// Update combat interface
function updateCombatInterface() {
    const state = getGameState();
    const combatInterface = document.getElementById('combat-interface');
    const combatStatus = document.getElementById('combat-status');
    const combatActions = document.getElementById('combat-actions');

    if (!state.combat) {
        if (combatInterface) combatInterface.classList.add('hidden');
        return;
    }

    if (combatInterface) combatInterface.classList.remove('hidden');

    // Update combat status
    if (combatStatus) {
        combatStatus.innerHTML = `
            <div class="flex justify-between items-center">
                <div>
                    <h4 class="text-lg font-bold">${state.combat.monster.name}</h4>
                    <div class="flex items-center space-x-2">
                        <span>HP: ${state.combat.monster.hp}/${state.combat.monster.max_hp}</span>
                        <div class="w-32 bg-gray-600 rounded-full h-2">
                            <div class="bg-red-500 h-2 rounded-full" style="width: ${(state.combat.monster.hp / state.combat.monster.max_hp) * 100}%"></div>
                        </div>
                    </div>
                </div>
                <div class="text-right">
                    <div>Distance: <span class="font-bold">${state.combat.distance}</span></div>
                    <div class="text-sm ${state.combat.player_turn ? 'text-green-400' : 'text-red-400'}">
                        ${state.combat.player_turn ? 'Your Turn' : 'Enemy Turn'}
                    </div>
                </div>
            </div>
        `;
    }

    // Update combat actions
    if (combatActions && state.combat.player_turn) {
        combatActions.innerHTML = '';

        // Movement actions
        if (state.combat.distance > 0) {
            combatActions.appendChild(
                createActionButton('→ Advance', () => combatAdvance(), 'bg-blue-600 hover:bg-blue-700')
            );
        }
        if (state.combat.distance < 6) {
            combatActions.appendChild(
                createActionButton('← Retreat', () => combatRetreat(), 'bg-yellow-600 hover:bg-yellow-700')
            );
        }

        // Attack actions
        combatActions.appendChild(
            createActionButton('⚔️ Attack', () => combatAttack(), 'bg-red-600 hover:bg-red-700')
        );

        // Spell actions
        const preparedSpells = (state.spells || []).filter(spell => spell.prepared);
        preparedSpells.slice(0, 2).forEach(spell => { // Limit to 2 spells for space
            const spellData = getSpellById(spell.spell);
            if (spellData && state.character.mana >= (spellData.mana_cost || 0)) {
                combatActions.appendChild(
                    createActionButton(`✨ ${spellData.name}`, () => castSpell(spell.spell), 'bg-purple-600 hover:bg-purple-700')
                );
            }
        });

        // Other actions
        combatActions.appendChild(
            createActionButton('🛡️ Defend', () => combatDefend(), 'bg-gray-600 hover:bg-gray-700')
        );

        combatActions.appendChild(
            createActionButton('🏃 Flee', () => combatFlee(), 'bg-orange-600 hover:bg-orange-700')
        );
    }
}

// Basic combat actions
function combatAdvance() {
    const state = getGameState();
    if (!state.combat) return;

    const newCombat = { ...state.combat };
    newCombat.distance = Math.max(0, newCombat.distance - 1);
    newCombat.player_turn = false;

    updateGameState({ combat: newCombat });
    showMessage('→ You move closer', 'info');

    // Simple AI turn
    setTimeout(enemyTurn, 1000);
}

function combatRetreat() {
    const state = getGameState();
    if (!state.combat) return;

    const newCombat = { ...state.combat };
    newCombat.distance = Math.min(6, newCombat.distance + 1);
    newCombat.player_turn = false;

    updateGameState({ combat: newCombat });
    showMessage('← You back away', 'info');

    setTimeout(enemyTurn, 1000);
}

function combatAttack() {
    const state = getGameState();
    if (!state.combat) return;

    // Simple attack calculation
    const damage = Math.floor(Math.random() * 6) + 1; // 1d6 damage
    const newCombat = { ...state.combat };
    newCombat.monster.hp = Math.max(0, newCombat.monster.hp - damage);
    newCombat.player_turn = false;

    showMessage(`⚔️ You deal ${damage} damage!`, 'success');

    if (newCombat.monster.hp <= 0) {
        // Combat over - victory
        showMessage('🏆 Victory!', 'success');
        updateGameState({ combat: null });
        return;
    }

    updateGameState({ combat: newCombat });
    setTimeout(enemyTurn, 1000);
}

function combatDefend() {
    const state = getGameState();
    if (!state.combat) return;

    const newCombat = { ...state.combat };
    newCombat.player_turn = false;

    updateGameState({ combat: newCombat });
    showMessage('🛡️ You take a defensive stance', 'info');

    setTimeout(enemyTurn, 1000);
}

function combatFlee() {
    const state = getGameState();
    if (!state.combat) return;

    // Simple flee chance
    if (Math.random() < 0.7) {
        showMessage('🏃 You escape!', 'success');

        // Add fatigue from fleeing
        const newCharacter = { ...state.character };
        newCharacter.fatigue = Math.min(10, newCharacter.fatigue + 1);

        updateGameState({
            combat: null,
            character: newCharacter
        });
    } else {
        showMessage('❌ Cannot escape!', 'error');
        const newCombat = { ...state.combat };
        newCombat.player_turn = false;
        updateGameState({ combat: newCombat });
        setTimeout(enemyTurn, 1000);
    }
}

// Simple enemy AI
function enemyTurn() {
    const state = getGameState();
    if (!state.combat || state.combat.player_turn) return;

    // Simple enemy attack
    const damage = Math.floor(Math.random() * 4) + 1; // 1d4 damage
    const newCharacter = { ...state.character };
    newCharacter.hp = Math.max(0, newCharacter.hp - damage);

    const newCombat = { ...state.combat };
    newCombat.player_turn = true;

    showMessage(`💥 ${state.combat.monster.name} deals ${damage} damage!`, 'warning');

    if (newCharacter.hp <= 0) {
        // Game over
        showMessage('💀 You have been defeated!', 'error');
        // Could implement respawn system here
        return;
    }

    updateGameState({
        character: newCharacter,
        combat: newCombat
    });
}

// Shop interface (basic)
function openShop() {
    showMessage('🛍️ Shop system not fully implemented yet', 'info');
    // This would open a shop interface
}

// Tavern interface (basic)
function openTavern() {
    showMessage('🍺 Tavern system not fully implemented yet', 'info');
    // This would open a tavern interface
}

// Update all displays
function updateAllDisplays() {
    updateCharacterDisplay();
    updateInventoryDisplay();
    updateSpellsDisplay();
    displayCurrentLocation();

    const state = getGameState();
    if (state.combat) {
        updateCombatInterface();
    }
}

// Save game to Nostr relay
async function saveGameToRelay() {
    const state = getGameState();
    const npub = getCurrentNpub();

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
            document.getElementById('save-btn').textContent = '💾 Saved!';
            setTimeout(() => {
                document.getElementById('save-btn').textContent = '💾 Save Game';
            }, 2000);
        } else {
            showMessage('❌ Failed to save game', 'error');
        }
    } catch (error) {
        showMessage('❌ Error saving game: ' + error.message, 'error');
    }
}

console.log('Game UI functions loaded');