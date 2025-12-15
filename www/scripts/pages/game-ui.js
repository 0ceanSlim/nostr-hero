// Game UI Functions
// Handles all user interface updates and interactions

// Advancement data cache
let advancementDataCache = null;

/**
 * Add a message to the game text log
 * Messages are appended to preserve history
 */
function addGameLog(message) {
    const gameText = document.getElementById('game-text');
    if (!gameText) return;

    // Create new log entry with styled border
    const logEntry = document.createElement('p');
    logEntry.className = 'border-l-2 border-gray-600 pl-2';
    logEntry.textContent = message;

    // Append to log
    gameText.appendChild(logEntry);

    // Auto-scroll to bottom to show latest message
    const textContainer = gameText.parentElement;
    if (textContainer) {
        textContainer.scrollTop = textContainer.scrollHeight;
    }
}

/**
 * Load advancement data from JSON
 */
async function loadAdvancementData() {
  if (advancementDataCache) {
    return advancementDataCache;
  }

  try {
    const response = await fetch('/data/systems/advancement.json');
    if (!response.ok) {
      throw new Error('Failed to fetch advancement data');
    }
    advancementDataCache = await response.json();
    return advancementDataCache;
  } catch (error) {
    console.error('Error loading advancement data:', error);
    return [];
  }
}

/**
 * Get item data from database cache by ID
 */
async function getItemByIdAsync(itemId) {
  try {
    const items = await loadItemsFromDatabase();
    const item = items.find(i => i.id === itemId);
    return item || null;
  } catch (error) {
    console.error(`Error getting item ${itemId}:`, error);
    return null;
  }
}

// Message system - DISABLED for work-in-progress UI
/**
 * Show action text in the game log with color coding
 * @param {string} text - The message to display
 * @param {string} color - Color: 'purple', 'white', 'red', 'green', 'yellow', 'blue'
 * @param {number} duration - Ignored (kept for API compatibility)
 */
window.showActionText = function showActionText(text, color = 'white', duration = 0) {
    const gameText = document.getElementById('game-text');
    if (!gameText) return;

    // Color mapping
    const colors = {
        'purple': '#a78bfa',   // Welcome messages
        'white': '#ffffff',    // Descriptions, neutral info
        'red': '#ef4444',      // Errors
        'green': '#22c55e',    // Success
        'yellow': '#eab308',   // Warnings
        'blue': '#3b82f6'      // Info
    };

    // Create new log entry with styled border and color
    const logEntry = document.createElement('p');
    logEntry.className = 'border-l-2 pl-2';
    logEntry.style.borderColor = colors[color] || colors['white'];
    logEntry.style.color = colors[color] || colors['white'];
    logEntry.textContent = text;

    // Append to log
    gameText.appendChild(logEntry);

    // Auto-scroll to bottom to show latest message
    const textContainer = gameText.parentElement;
    if (textContainer) {
        textContainer.scrollTop = textContainer.scrollHeight;
    }
}

// Legacy showMessage function - now uses showActionText
window.showMessage = function showMessage(text, type = 'info', duration = 5000) {
    const colorMap = {
        'error': 'red',
        'success': 'green',
        'warning': 'yellow',
        'info': 'blue'
    };
    showActionText(text, colorMap[type] || 'white', duration);
}

/**
 * Calculate max weight capacity based on strength and equipped items
 */
async function calculateMaxCapacity(character) {
    const strength = character.stats?.strength || 10;
    let baseCapacity = strength * 5;

    // Add weight_increase from all equipped items
    const items = await loadItemsFromDatabase();
    let capacityBonus = 0;

    if (character.inventory?.gear_slots) {
        for (const slotName in character.inventory.gear_slots) {
            const slot = character.inventory.gear_slots[slotName];
            if (slot && slot.item) {
                const itemData = items.find(i => i.id === slot.item);
                if (itemData) {
                    const weightIncrease = itemData.weight_increase || itemData.properties?.weight_increase || 0;
                    if (weightIncrease > 0) {
                        capacityBonus += weightIncrease;
                    }
                }
            }
        }
    }

    const totalCapacity = baseCapacity + capacityBonus;
    return totalCapacity;
}

/**
 * Calculate total weight of all items in inventory
 */
async function calculateAndDisplayWeight(character) {
    if (!character.inventory) {
        return 0;
    }

    let totalWeight = 0;
    const items = await loadItemsFromDatabase();

    // Function to get item weight by ID
    const getItemWeight = (itemId) => {
        const item = items.find(i => i.id === itemId);
        if (!item) {
            return 0;
        }
        // Weight can be at top level or in properties object
        const weight = item.weight || item.properties?.weight || 0;
        return weight;
    };

    // Calculate gear slots weight
    if (character.inventory.gear_slots) {
        const gearSlots = character.inventory.gear_slots;

        // All gear slots except bag contents
        for (const slotName in gearSlots) {
            if (slotName === 'bag') continue; // Handle bag separately

            const slot = gearSlots[slotName];
            if (slot && slot.item) {
                const weight = getItemWeight(slot.item);
                const itemWeight = weight * (slot.quantity || 1);
                totalWeight += itemWeight;
            }
        }

        // Bag itself and its contents
        if (gearSlots.bag) {
            // Add bag weight
            if (gearSlots.bag.item) {
                const bagWeight = getItemWeight(gearSlots.bag.item);
                totalWeight += bagWeight;
            }

            // Add contents weight
            if (gearSlots.bag.contents && Array.isArray(gearSlots.bag.contents)) {
                gearSlots.bag.contents.forEach(slot => {
                    if (slot && slot.item) {
                        const weight = getItemWeight(slot.item);
                        const itemWeight = weight * (slot.quantity || 1);
                        totalWeight += itemWeight;
                    }
                });
            }
        }
    }

    // Calculate general slots weight
    if (character.inventory.general_slots && Array.isArray(character.inventory.general_slots)) {
        character.inventory.general_slots.forEach(slot => {
            if (slot && slot.item) {
                const weight = getItemWeight(slot.item);
                const itemWeight = weight * (slot.quantity || 1);
                totalWeight += itemWeight;
            }
        });
    }

    // Add gold weight (50 gold = 1 lb)
    const goldWeight = Math.floor((character.gold || 0) / 50);
    totalWeight += goldWeight;

    return Math.round(totalWeight);
}

// Update full character display from save data
async function updateCharacterDisplay() {
    const state = getGameStateSync();
    const character = state.character;

    if (!character) {
        console.error('❌ No character data found!');
        return;
    }

    // Update character info
    if (document.getElementById('char-name')) {
        document.getElementById('char-name').textContent = character.name || '-';
    }
    if (document.getElementById('char-race')) {
        document.getElementById('char-race').textContent = character.race || '-';
    }
    if (document.getElementById('char-class')) {
        document.getElementById('char-class').textContent = character.class || '-';
    }
    if (document.getElementById('char-background')) {
        document.getElementById('char-background').textContent = character.background || '-';
    }
    if (document.getElementById('char-alignment')) {
        document.getElementById('char-alignment').textContent = character.alignment || '-';
    }
    if (document.getElementById('char-level')) {
        document.getElementById('char-level').textContent = character.level || 1;
    }
    if (document.getElementById('char-gold')) {
        document.getElementById('char-gold').textContent = character.gold || 0;
    }

    // Update XP bar
    const currentExpEl = document.getElementById('current-exp');
    const expToNextEl = document.getElementById('exp-to-next');
    const expBarEl = document.getElementById('exp-bar');

    const advancementData = await loadAdvancementData();
    const currentXP = character.experience || 0;
    const currentLevel = character.level || 1;

    // Find current and next level data
    const currentLevelData = advancementData.find(l => l.Level === currentLevel);
    const nextLevelData = advancementData.find(l => l.Level === currentLevel + 1);

    if (currentExpEl && expToNextEl && expBarEl && currentLevelData) {
        const currentLevelXP = currentLevelData.ExperiencePoints;
        const nextLevelXP = nextLevelData ? nextLevelData.ExperiencePoints : currentLevelXP;
        const xpIntoLevel = currentXP - currentLevelXP;
        const xpNeededForLevel = nextLevelXP - currentLevelXP;

        currentExpEl.textContent = xpIntoLevel.toLocaleString();
        expToNextEl.textContent = xpNeededForLevel.toLocaleString();

        const expPercentage = xpNeededForLevel > 0 ? (xpIntoLevel / xpNeededForLevel * 100) : 0;
        expBarEl.style.width = Math.min(100, Math.max(0, expPercentage)) + '%';
    }

    // Update HP display
    const currentHpEl = document.getElementById('current-hp');
    const maxHpEl = document.getElementById('max-hp');
    const hpBarEl = document.getElementById('hp-bar');
    if (currentHpEl) currentHpEl.textContent = character.hp || 0;
    if (maxHpEl) maxHpEl.textContent = character.max_hp || 0;
    if (hpBarEl) {
        const hpPercentage = (character.max_hp > 0) ? (character.hp / character.max_hp * 100) : 0;
        hpBarEl.style.width = hpPercentage + '%';
    }

    // Update mana display
    const currentManaEl = document.getElementById('current-mana');
    const maxManaEl = document.getElementById('max-mana');
    const manaBarEl = document.getElementById('mana-bar');
    if (currentManaEl) currentManaEl.textContent = character.mana || 0;
    if (maxManaEl) maxManaEl.textContent = character.max_mana || 0;
    if (manaBarEl) {
        const manaPercentage = (character.max_mana > 0) ? (character.mana / character.max_mana * 100) : 0;
        manaBarEl.style.width = manaPercentage + '%';
    }

    // Update quick status (main bar - numbers + emojis)
    const fatigue = Math.min(character.fatigue || 0, 10);
    const hunger = Math.max(0, Math.min(character.hunger !== undefined ? character.hunger : 1, 3));

    // Fatigue number and emoji
    const fatigueLevelEl = document.getElementById('fatigue-level');
    const fatigueEmojiEl = document.getElementById('fatigue-emoji');

    if (fatigueLevelEl) fatigueLevelEl.textContent = fatigue;
    if (fatigueEmojiEl) {
        if (fatigue <= 2) {
            fatigueEmojiEl.textContent = '😊'; // Fresh
        } else if (fatigue <= 5) {
            fatigueEmojiEl.textContent = '😐'; // Tired
        } else if (fatigue <= 8) {
            fatigueEmojiEl.textContent = '😓'; // Weary
        } else {
            fatigueEmojiEl.textContent = '😵'; // Exhausted
        }
    }

    // Hunger number and emoji
    const hungerLevelEl = document.getElementById('hunger-level');
    const hungerEmojiEl = document.getElementById('hunger-emoji');

    if (hungerLevelEl) hungerLevelEl.textContent = hunger;
    if (hungerEmojiEl) {
        if (hunger === 0) {
            hungerEmojiEl.textContent = '😵'; // Famished
        } else if (hunger === 1) {
            hungerEmojiEl.textContent = '😋'; // Hungry
        } else if (hunger === 2) {
            hungerEmojiEl.textContent = '🙂'; // Satisfied
        } else {
            hungerEmojiEl.textContent = '😊'; // Full
        }
    }

    // Weight numbers and emoji
    const weightEl = document.getElementById('char-weight');
    const maxWeightEl = document.getElementById('max-weight');
    const weightEmojiEl = document.getElementById('weight-emoji');

    if (weightEl || maxWeightEl || weightEmojiEl) {
        Promise.all([
            calculateAndDisplayWeight(character),
            calculateMaxCapacity(character)
        ]).then(([weight, maxCapacity]) => {
            if (weightEl) weightEl.textContent = weight;
            if (maxWeightEl) maxWeightEl.textContent = maxCapacity;

            const weightPercentage = (weight / maxCapacity) * 100;

            if (weightEmojiEl) {
                if (weightPercentage <= 50) {
                    weightEmojiEl.textContent = '🪶'; // Light
                } else if (weightPercentage <= 100) {
                    weightEmojiEl.textContent = '✓'; // OK
                } else if (weightPercentage <= 150) {
                    weightEmojiEl.textContent = '📦'; // Heavy
                } else {
                    weightEmojiEl.textContent = '🐌'; // Overloaded
                }
            }
        });
    }

    // Update detailed stats tab (if visible)
    updateStatsTab(character);

    // Update stats
    if (character.stats) {
        const statStrEl = document.getElementById('stat-str');
        const statDexEl = document.getElementById('stat-dex');
        const statConEl = document.getElementById('stat-con');
        const statIntEl = document.getElementById('stat-int');
        const statWisEl = document.getElementById('stat-wis');
        const statChaEl = document.getElementById('stat-cha');

        if (statStrEl) statStrEl.textContent = character.stats.strength || 10;
        if (statDexEl) statDexEl.textContent = character.stats.dexterity || 10;
        if (statConEl) statConEl.textContent = character.stats.constitution || 10;
        if (statIntEl) statIntEl.textContent = character.stats.intelligence || 10;
        if (statWisEl) statWisEl.textContent = character.stats.wisdom || 10;
        if (statChaEl) statChaEl.textContent = character.stats.charisma || 10;
    }

    // Update equipment slots
    if (character.inventory && character.inventory.gear_slots) {
        const gear = character.inventory.gear_slots;
        const slots = ['left_arm', 'right_arm', 'armor', 'bag', 'necklace', 'ring', 'ammunition', 'clothes'];

        // Use for...of instead of forEach to properly handle async
        for (const slotName of slots) {
            const slotEl = document.querySelector(`[data-slot="${slotName}"]`);
            if (slotEl) {
                const itemId = gear[slotName]?.item;
                const quantity = gear[slotName]?.quantity || 1;

                if (itemId) {
                    // Add data-item-id attribute to the slot for interaction system
                    slotEl.setAttribute('data-item-id', itemId);

                    // Fetch item data
                    const itemData = await getItemByIdAsync(itemId);

                    if (itemData) {
                        // Replace placeholder with item image
                        const imageContainer = slotEl.querySelector('.w-10.h-10');
                        if (imageContainer) {
                            imageContainer.innerHTML = `<img src="/res/img/items/${itemId}.png" alt="${itemData.name}" class="w-full h-full object-contain" style="image-rendering: pixelated;">`;
                        }

                        // Add quantity label if > 1 (for ammunition, potions, etc.)
                        // First remove any existing quantity label
                        const existingLabel = slotEl.querySelector('.equipment-quantity-label');
                        if (existingLabel) {
                            existingLabel.remove();
                        }

                        if (quantity > 1) {
                            const quantityLabel = document.createElement('div');
                            quantityLabel.className = 'equipment-quantity-label absolute bottom-0 right-0 text-white';
                            quantityLabel.style.fontSize = '10px';
                            quantityLabel.textContent = `${quantity}`;
                            slotEl.appendChild(quantityLabel);
                        }
                    }
                } else {
                    // Remove data-item-id attribute if slot is empty
                    slotEl.removeAttribute('data-item-id');

                    // Remove quantity label if present
                    const existingLabel = slotEl.querySelector('.equipment-quantity-label');
                    if (existingLabel) {
                        existingLabel.remove();
                    }

                    // Reset to placeholder if empty
                    const imageContainer = slotEl.querySelector('.w-10.h-10');
                    if (imageContainer) {
                        // Check if placeholder exists
                        let placeholderIcon = slotEl.querySelector('.placeholder-icon');
                        if (placeholderIcon) {
                            // Placeholder exists, make sure it's visible
                            placeholderIcon.style.display = 'block';
                            // Remove any item image
                            const itemImg = imageContainer.querySelector('img');
                            if (itemImg) {
                                itemImg.remove();
                            }
                        } else {
                            // No placeholder found, just clear the image
                            const itemImg = imageContainer.querySelector('img');
                            if (itemImg) {
                                itemImg.remove();
                            }
                        }
                    }
                }
            }
        }
    }

    // Update general slots (4x1 grid) - ALWAYS create slots even if empty
    const generalSlotsDiv = document.getElementById('general-slots');
    if (generalSlotsDiv) {
        generalSlotsDiv.innerHTML = '';

        // Ensure inventory structure exists
        if (!character.inventory) {
            character.inventory = {};
        }
        if (!character.inventory.general_slots) {
            character.inventory.general_slots = [];
        }

        // Create a map of slot index to item data (respecting the "slot" field)
        const slotMap = {};
        character.inventory.general_slots.forEach(item => {
            if (item && item.item) {
                const slotIndex = item.slot;
                // Only use valid slot indices (0-3)
                if (slotIndex >= 0 && slotIndex < 4) {
                    slotMap[slotIndex] = item;
                }
            }
        });

        // Create all 4 general slots
        for (let i = 0; i < 4; i++) {
            const slot = slotMap[i];
            const slotDiv = document.createElement('div');
            slotDiv.className = 'relative cursor-pointer hover:bg-gray-600 flex items-center justify-center';
            slotDiv.style.cssText = `aspect-ratio: 1; background: #2a2a2a; border-top: 2px solid #1a1a1a; border-left: 2px solid #1a1a1a; border-right: 2px solid #4a4a4a; border-bottom: 2px solid #4a4a4a; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);`;

            // Add data attributes for interaction system
            slotDiv.setAttribute('data-item-slot', i);

            if (slot && slot.item) {
                slotDiv.setAttribute('data-item-id', slot.item);

                // Create image container
                const imgDiv = document.createElement('div');
                imgDiv.className = 'w-full h-full flex items-center justify-center p-1';
                const img = document.createElement('img');
                img.src = `/res/img/items/${slot.item}.png`;
                img.alt = slot.item;
                img.className = 'w-full h-full object-contain';
                img.style.imageRendering = 'pixelated';
                imgDiv.appendChild(img);
                slotDiv.appendChild(imgDiv);

                // Add quantity label if > 1
                if (slot.quantity > 1) {
                    const quantityLabel = document.createElement('div');
                    quantityLabel.className = 'absolute bottom-0 right-0 text-white';
                    quantityLabel.style.fontSize = '10px';
                    quantityLabel.textContent = `${slot.quantity}`;
                    slotDiv.appendChild(quantityLabel);
                }
            }

            generalSlotsDiv.appendChild(slotDiv);
        }
    }

    // Update backpack items (4x5 grid = 20 slots) - ONLY show if bag is equipped
    const backpackDiv = document.getElementById('backpack-slots');
    if (backpackDiv) {
        backpackDiv.innerHTML = '';

        // Check if a bag is actually equipped
        const bagEquipped = character.inventory?.gear_slots?.bag?.item;

        if (!bagEquipped) {
            // No bag equipped - hide the backpack div
            if (backpackDiv.parentElement) {
                backpackDiv.parentElement.style.display = 'none';
            }
            return; // Exit early, don't render any slots
        }

        // Bag is equipped - show the backpack div
        if (backpackDiv.parentElement) {
            backpackDiv.parentElement.style.display = 'grid';
        }

        // Get or initialize contents
        if (!character.inventory.gear_slots.bag.contents) {
            character.inventory.gear_slots.bag.contents = [];
        }

        const contents = character.inventory.gear_slots.bag.contents;

        // Create a map of slot index to item data (respecting the "slot" field)
        const bagSlotMap = {};
        contents.forEach(item => {
            if (item && item.item) {
                const slotIndex = item.slot;
                // Only use valid slot indices (0-19 for 20-slot backpack)
                if (slotIndex >= 0 && slotIndex < 20) {
                    bagSlotMap[slotIndex] = item;
                }
            }
        });

        let itemCount = 0;

        // Create all 20 backpack slots
        for (let i = 0; i < 20; i++) {
            const slot = bagSlotMap[i];
            const slotDiv = document.createElement('div');
            slotDiv.className = 'relative cursor-pointer hover:bg-gray-800 flex items-center justify-center';
            slotDiv.style.cssText = `aspect-ratio: 1; background: #1a1a1a; border-top: 2px solid #000000; border-left: 2px solid #000000; border-right: 2px solid #3a3a3a; border-bottom: 2px solid #3a3a3a; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);`;

            // Add data attributes for interaction system
            slotDiv.setAttribute('data-item-slot', i);

            if (slot && slot.item) {
                itemCount++;
                slotDiv.setAttribute('data-item-id', slot.item);

                // Create image container
                const imgDiv = document.createElement('div');
                imgDiv.className = 'w-full h-full flex items-center justify-center p-1';
                const img = document.createElement('img');
                img.src = `/res/img/items/${slot.item}.png`;
                img.alt = slot.item;
                img.className = 'w-full h-full object-contain';
                img.style.imageRendering = 'pixelated';
                imgDiv.appendChild(img);
                slotDiv.appendChild(imgDiv);

                // Add quantity label if > 1
                if (slot.quantity > 1) {
                    const quantityLabel = document.createElement('div');
                    quantityLabel.className = 'absolute bottom-0 right-0 text-white';
                    quantityLabel.style.fontSize = '10px';
                    quantityLabel.textContent = `${slot.quantity}`;
                    slotDiv.appendChild(quantityLabel);
                }
            }

            backpackDiv.appendChild(slotDiv);
        }

        const bagCountEl = document.getElementById('bag-count');
        if (bagCountEl) {
            bagCountEl.textContent = itemCount;
        }
    }

    // Rebind inventory interactions after rendering slots
    // This ensures events are always attached, regardless of where updateCharacterDisplay() is called from
    if (window.inventoryInteractions && window.inventoryInteractions.bindInventoryEvents) {
        window.inventoryInteractions.bindInventoryEvents();
    }
}

// Update detailed stats tab
async function updateStatsTab(character) {
    // Character info
    const raceEl = document.getElementById('stats-char-race');
    const classEl = document.getElementById('stats-char-class');
    const backgroundEl = document.getElementById('stats-char-background');
    const alignmentEl = document.getElementById('stats-char-alignment');

    if (raceEl) raceEl.textContent = character.race || '-';
    if (classEl) classEl.textContent = character.class || '-';
    if (backgroundEl) backgroundEl.textContent = character.background || '-';
    if (alignmentEl) alignmentEl.textContent = character.alignment || '-';

    // Ability scores with modifiers
    if (character.stats) {
        const stats = ['str', 'dex', 'con', 'int', 'wis', 'cha'];
        const statNames = ['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma'];

        stats.forEach((stat, index) => {
            const valueEl = document.getElementById(`stats-${stat}`);
            const modEl = document.getElementById(`stats-${stat}-mod`);

            const value = character.stats[statNames[index]] || 10;
            const modifier = Math.floor((value - 10) / 2);

            if (valueEl) valueEl.textContent = value;
            if (modEl) modEl.textContent = modifier >= 0 ? `+${modifier}` : modifier;
        });
    }

    // Fatigue details
    const fatigue = Math.min(character.fatigue || 0, 10);
    const fatigueLevelEl = document.getElementById('stats-fatigue-level');
    const fatigueStatusEl = document.getElementById('stats-fatigue-status');
    const fatigueEmojiEl = document.getElementById('stats-fatigue-emoji');
    const fatigueDescEl = document.getElementById('stats-fatigue-desc');

    if (fatigueLevelEl) fatigueLevelEl.textContent = fatigue;

    let fatigueStatus, fatigueEmoji, fatigueDesc, fatigueColor;
    if (fatigue <= 2) {
        fatigueStatus = 'FRESH';
        fatigueEmoji = '😊';
        fatigueDesc = 'You feel energetic and ready for adventure';
        fatigueColor = 'text-green-400';
    } else if (fatigue <= 5) {
        fatigueStatus = 'TIRED';
        fatigueEmoji = '😐';
        fatigueDesc = "You're starting to feel the strain of travel";
        fatigueColor = 'text-yellow-400';
    } else if (fatigue <= 8) {
        fatigueStatus = 'WEARY';
        fatigueEmoji = '😓';
        fatigueDesc = 'Your steps are heavy and reactions slower';
        fatigueColor = 'text-orange-400';
    } else {
        fatigueStatus = 'EXHAUSTED';
        fatigueEmoji = '😵';
        fatigueDesc = 'You can barely function - rest immediately!';
        fatigueColor = 'text-red-400';
    }

    if (fatigueStatusEl) {
        fatigueStatusEl.textContent = fatigueStatus;
        fatigueStatusEl.className = `${fatigueColor} font-bold text-xs`;
    }
    if (fatigueEmojiEl) fatigueEmojiEl.textContent = fatigueEmoji;
    if (fatigueDescEl) fatigueDescEl.textContent = fatigueDesc;

    // Hunger details
    const hunger = Math.max(0, Math.min(character.hunger !== undefined ? character.hunger : 1, 3));
    const hungerLevelEl = document.getElementById('stats-hunger-level');
    const hungerStatusEl = document.getElementById('stats-hunger-status');
    const hungerEmojiEl = document.getElementById('stats-hunger-emoji');
    const hungerDescEl = document.getElementById('stats-hunger-desc');

    if (hungerLevelEl) hungerLevelEl.textContent = hunger;

    let hungerStatus, hungerEmoji, hungerDesc, hungerColor;
    if (hunger === 0) {
        hungerStatus = 'FAMISHED';
        hungerEmoji = '😵';
        hungerDesc = 'You are starving and weak';
        hungerColor = 'text-red-400';
    } else if (hunger === 1) {
        hungerStatus = 'HUNGRY';
        hungerEmoji = '😋';
        hungerDesc = 'You could use a meal';
        hungerColor = 'text-yellow-400';
    } else if (hunger === 2) {
        hungerStatus = 'SATISFIED';
        hungerEmoji = '🙂';
        hungerDesc = 'Your belly is content';
        hungerColor = 'text-green-400';
    } else {
        hungerStatus = 'FULL';
        hungerEmoji = '😊';
        hungerDesc = "You're well-fed and energized";
        hungerColor = 'text-green-400';
    }

    if (hungerStatusEl) {
        hungerStatusEl.textContent = hungerStatus;
        hungerStatusEl.className = `${hungerColor} font-bold text-xs`;
    }
    if (hungerEmojiEl) hungerEmojiEl.textContent = hungerEmoji;
    if (hungerDescEl) hungerDescEl.textContent = hungerDesc;

    // Weight details
    const weightEl = document.getElementById('stats-weight');
    const maxWeightEl = document.getElementById('stats-max-weight');
    const weightStatusEl = document.getElementById('stats-weight-status');
    const weightEmojiEl = document.getElementById('stats-weight-emoji');
    const weightDescEl = document.getElementById('stats-weight-desc');

    try {
        const [weight, maxCapacity] = await Promise.all([
            calculateAndDisplayWeight(character),
            calculateMaxCapacity(character)
        ]);

        if (weightEl) weightEl.textContent = weight;
        if (maxWeightEl) maxWeightEl.textContent = maxCapacity;

        const weightPercentage = (weight / maxCapacity) * 100;
        let weightStatus, weightEmoji, weightDesc, weightColor;

        if (weightPercentage <= 50) {
            weightStatus = 'LIGHT';
            weightEmoji = '🪶';
            weightDesc = 'You move at full speed';
            weightColor = 'text-green-400';
        } else if (weightPercentage <= 100) {
            weightStatus = 'NORMAL';
            weightEmoji = '✓';
            weightDesc = 'Carrying a comfortable load';
            weightColor = 'text-green-400';
        } else if (weightPercentage <= 150) {
            weightStatus = 'HEAVY';
            weightEmoji = '📦';
            weightDesc = 'Movement slightly hindered';
            weightColor = 'text-yellow-400';
        } else if (weightPercentage <= 200) {
            weightStatus = 'OVERLOADED';
            weightEmoji = '🐌';
            weightDesc = 'Severely slowed, drop items!';
            weightColor = 'text-orange-400';
        } else {
            weightStatus = 'IMMOBILE';
            weightEmoji = '🛑';
            weightDesc = 'Cannot move! Drop items immediately!';
            weightColor = 'text-red-400';
        }

        if (weightStatusEl) {
            weightStatusEl.textContent = weightStatus;
            weightStatusEl.className = `${weightColor} font-bold text-xs`;
        }
        if (weightEmojiEl) weightEmojiEl.textContent = weightEmoji;
        if (weightDescEl) weightDescEl.textContent = weightDesc;
    } catch (error) {
        console.error('Error calculating weight:', error);
    }
}

// Update inventory display
function updateInventoryDisplay() {
    // Inventory is now displayed in updateCharacterDisplay via general_slots and backpack_slots
    // This function is kept for compatibility but does nothing
}

// Update spells display
function updateSpellsDisplay() {
    const state = getGameStateSync();
    const character = state.character;

    // Update known spells
    const knownSpellsEl = document.getElementById('known-spells');
    if (knownSpellsEl && character?.spells) {
        knownSpellsEl.innerHTML = '';

        // Character.spells is an array of spell IDs
        const spellsArray = Array.isArray(character.spells) ? character.spells : [];

        spellsArray.forEach(spellId => {
            const spell = getSpellById(spellId);

            if (!spell) {
                console.error(`❌ Spell not found: ${spellId}`);
                return; // Skip this spell
            }

            const spellDiv = document.createElement('div');
            spellDiv.className = 'relative cursor-pointer hover:opacity-80 flex flex-col items-center justify-center overflow-hidden';
            spellDiv.style.cssText = 'width: 56px; height: 56px; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);';

            // Background image from school
            const bgImg = document.createElement('img');
            bgImg.src = `/res/img/spells/${spell.school}.png`;
            bgImg.className = 'absolute inset-0 w-full h-full object-cover';
            bgImg.style.imageRendering = 'pixelated';
            bgImg.style.opacity = '0.6';
            spellDiv.appendChild(bgImg);

            // Spell name
            const nameDiv = document.createElement('div');
            nameDiv.className = 'relative z-10 text-center px-1';
            nameDiv.innerHTML = `<span class="text-white font-bold" style="font-size: 7px; line-height: 1.1; text-shadow: 1px 1px 2px black; word-break: break-word;">${spellId.replace(/-/g, ' ').toUpperCase()}</span>`;
            spellDiv.appendChild(nameDiv);

            // Damage info (if applicable)
            if (spell.damage) {
                const damageDiv = document.createElement('div');
                damageDiv.className = 'absolute bottom-0 right-0 bg-black bg-opacity-70 px-1 text-white font-bold z-10';
                damageDiv.style.fontSize = '8px';
                damageDiv.textContent = `${spell.damage} ${spell.emoji}`;
                spellDiv.appendChild(damageDiv);
            }

            spellDiv.addEventListener('click', () => {
                if (typeof castSpell === 'function') {
                    castSpell(spellId);
                }
            });

            knownSpellsEl.appendChild(spellDiv);
        });
    }

    // Update spell slots - new format with array of slot objects
    const spellSlotsEl = document.getElementById('spell-slots-container');
    if (spellSlotsEl && character?.spell_slots) {
        spellSlotsEl.innerHTML = '';

        // Sort spell levels (cantrips first, then level_1, level_2, etc.)
        const sortedLevels = Object.keys(character.spell_slots).sort((a, b) => {
            if (a === 'cantrips') return -1;
            if (b === 'cantrips') return 1;
            const aNum = parseInt(a.split('_')[1]) || 0;
            const bNum = parseInt(b.split('_')[1]) || 0;
            return aNum - bNum;
        });

        sortedLevels.forEach(level => {
            const slots = character.spell_slots[level];
            if (slots && Array.isArray(slots) && slots.length > 0) {
                const levelDiv = document.createElement('div');
                levelDiv.className = 'mb-3';

                // Create level header
                const levelLabel = level === 'cantrips' ? 'CANTRIPS' : `LEVEL ${level.split('_')[1]}`;
                const header = document.createElement('div');
                header.className = 'text-xs text-gray-400 mb-1';
                header.textContent = levelLabel;
                levelDiv.appendChild(header);

                // Create slot grid
                const slotGrid = document.createElement('div');
                slotGrid.className = 'grid grid-cols-4 gap-1';

                slots.forEach((slotData, index) => {
                    const slotDiv = document.createElement('div');
                    slotDiv.className = 'bg-gray-700 relative cursor-pointer hover:bg-gray-600 flex flex-col items-center justify-center';
                    slotDiv.style.cssText = 'width: 56px; height: 56px; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);';

                    if (slotData.spell) {
                        // Has a spell prepared
                        const spellDiv = document.createElement('div');
                        spellDiv.className = 'w-full h-full flex items-center justify-center p-1 text-center';
                        spellDiv.innerHTML = `<span class="text-purple-400 text-xs" style="font-size: 8px; line-height: 1.2;">${slotData.spell}</span>`;
                        slotDiv.appendChild(spellDiv);
                    } else {
                        // Empty slot
                        const emptyDiv = document.createElement('div');
                        emptyDiv.className = 'w-full h-full flex items-center justify-center text-gray-600';
                        emptyDiv.innerHTML = '<svg class="w-8 h-8" viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg"><circle cx="32" cy="32" r="16" fill="none" stroke="#4B5563" stroke-width="2"/><circle cx="32" cy="32" r="4" fill="#6B7280"/></svg>';
                        slotDiv.appendChild(emptyDiv);
                    }

                    slotGrid.appendChild(slotDiv);
                });

                levelDiv.appendChild(slotGrid);
                spellSlotsEl.appendChild(levelDiv);
            }
        });
    }
}

// Current location music player
let currentLocationMusic = null;

// Track last displayed location to prevent duplicate descriptions
let lastDisplayedLocation = null;

/**
 * Play location-specific music
 */
function playLocationMusic(musicPath) {
    // Stop current music if playing
    if (currentLocationMusic) {
        currentLocationMusic.pause();
        currentLocationMusic.currentTime = 0;
    }

    // Create new audio element
    currentLocationMusic = new Audio(musicPath);
    currentLocationMusic.loop = true;
    currentLocationMusic.volume = 0.5; // Adjust volume as needed

    // Play the music
    currentLocationMusic.play().catch(err => {
        console.log('Music autoplay prevented:', err);
        // User interaction may be required to start audio
    });

    console.log('Playing location music:', musicPath);
}

// Display current location
function displayCurrentLocation() {
    const state = getGameStateSync();
    const cityId = state.location?.current;
    const districtKey = state.location?.district || 'center';

    if (!cityId) return;

    // Construct full district ID from city + district (e.g., "village-west-east")
    const districtId = `${cityId}-${districtKey}`;
    const currentLocationId = districtId;  // For compatibility with rest of function

    console.log('📍 Display location:', { cityId, districtKey, districtId });

    // Get the city data (for image, music)
    const cityData = getLocationById(cityId);
    if (!cityData) {
        console.error('❌ City not found:', cityId);
        return;
    }

    // Get the district data (for description, buildings, connections)
    const locationData = getLocationById(districtId);
    if (!locationData) {
        console.error('❌ District not found:', districtId);
        return;
    }

    console.log('✅ City data:', cityData);
    console.log('✅ District data:', locationData);

    // Update scene image (use city's image for all districts)
    const sceneImage = document.getElementById('scene-image');
    if (sceneImage && cityData.image) {
        sceneImage.src = cityData.image;
        sceneImage.alt = cityData.name;
    }

    // Update music (use city's music for all districts)
    if (cityData.music) {
        playLocationMusic(cityData.music);
    }

    // Update city name (top of scene)
    const cityName = document.getElementById('city-name');
    if (cityName) {
        cityName.textContent = cityData.name;
    }

    // Update district name (bottom of scene)
    const districtName = document.getElementById('district-name');
    if (districtName) {
        districtName.textContent = locationData.name;
    }

    // Update time of day display
    updateTimeDisplay();

    // Show location description in action text (white color) - only if location changed
    // Include building ID in location key to detect entering/exiting buildings
    const currentBuildingId = state.location?.building || '';
    const locationKey = `${cityId}-${districtKey}-${currentBuildingId}`;
    if (locationData.description && lastDisplayedLocation !== locationKey) {
        showActionText(locationData.description, 'white');
        lastDisplayedLocation = locationKey;
    }

    // Generate location actions based on city district structure
    const navContainer = document.getElementById('navigation-buttons');
    const buildingContainer = document.getElementById('building-buttons');
    const npcContainer = document.getElementById('npc-buttons');

    // Clear building and NPC containers (not navigation - that's handled per-slot)
    if (buildingContainer) {
        const buildingButtonContainer = buildingContainer.querySelector('div');
        if (buildingButtonContainer) buildingButtonContainer.innerHTML = '';
    }
    if (npcContainer) {
        const npcButtonContainer = npcContainer.querySelector('div');
        if (npcButtonContainer) npcButtonContainer.innerHTML = '';
    }

    // Get district data from location
    let districtData = null;
    if (locationData.properties?.districts) {
        // Find the current district
        for (const district of Object.values(locationData.properties.districts)) {
            if (district.id === currentLocationId) {
                districtData = district;
                break;
            }
        }
    }

    // If we have district data, use it; otherwise fall back to location data
    const currentData = districtData || locationData;
    console.log('Current data for buttons:', currentData);

    // Get connections - check both direct and properties.connections
    const connections = currentData.connections || currentData.properties?.connections;
    let buildings = currentData.buildings || currentData.properties?.buildings;
    let npcs = currentData.npcs || currentData.properties?.npcs;

    // Check if we're inside a building (currentBuildingId already declared above)
    if (currentBuildingId && buildings) {
        // Find the current building data
        const currentBuilding = buildings.find(b => b.id === currentBuildingId);

        if (currentBuilding) {
            console.log('📍 Inside building:', currentBuilding);

            // Get NPCs from the building
            if (currentBuilding.npcs) {
                npcs = currentBuilding.npcs;  // Array of NPC IDs
            } else if (currentBuilding.npc) {
                npcs = [currentBuilding.npc];  // Single NPC
            } else {
                npcs = [];  // No NPCs in building
            }

            // Override buildings to show only "Exit Building" button
            buildings = [{ id: '__exit__', name: 'Exit Building', isExit: true }];
        }
    }

    // 1. NAVIGATION BUTTONS (D-pad style with cardinal directions)
    console.log('Navigation connections:', connections);
    if (connections) {
        // Clear all D-pad slots first
        ['travel-n', 'travel-s', 'travel-e', 'travel-w', 'travel-center'].forEach(slotId => {
            const slot = document.getElementById(slotId);
            if (slot) slot.innerHTML = '';
        });

        Object.entries(connections).forEach(([direction, connectionId]) => {
            console.log(`Processing connection: ${direction} -> ${connectionId}`);
            const connectedLocation = getLocationById(connectionId);
            console.log(`Found location:`, connectedLocation);

            if (connectedLocation) {
                // Map direction to D-pad slot
                const slotMap = {
                    'north': 'travel-n',
                    'south': 'travel-s',
                    'east': 'travel-e',
                    'west': 'travel-w',
                    'center': 'travel-center'
                };

                const slotId = slotMap[direction.toLowerCase()];
                const slot = document.getElementById(slotId);

                if (slot) {
                    // Determine button type based on location_type
                    const buttonType = connectedLocation.location_type === 'environment' ? 'environment' : 'navigation';

                    const button = createLocationButton(
                        connectedLocation.name || direction.toUpperCase(),
                        () => moveToLocation(connectionId),
                        buttonType
                    );
                    button.className += ' w-full h-full';
                    slot.appendChild(button);
                } else {
                    console.warn(`⚠️ No slot found for ${direction} (${slotId})`);
                }
            } else {
                console.warn(`⚠️ No location found for ${direction} -> ${connectionId}`);
            }
        });
    } else {
        console.log('No connections found for this location');
    }

    // 2. BUILDING BUTTONS
    if (buildingContainer) {
        const buildingButtonContainer = buildingContainer.querySelector('div');

        if (buildings && buildings.length > 0) {
            // Get current time of day from game state
            const currentTime = state.character?.time_of_day !== undefined ? state.character.time_of_day : 12;

            buildings.forEach(building => {
                // Check if this is the special "Exit Building" button
                if (building.isExit) {
                    const button = createLocationButton(
                        building.name,
                        () => exitBuilding(),
                        'building'  // Use green for exit button
                    );
                    buildingButtonContainer.appendChild(button);
                    return;
                }

                // Check if building is currently open
                const isOpen = isBuildingOpen(building, currentTime);

                if (isOpen) {
                    // Open building - normal styling
                    const button = createLocationButton(
                        building.name,
                        () => enterBuilding(building.id),
                        'building'
                    );
                    buildingButtonContainer.appendChild(button);
                } else {
                    // Closed building - grey styling with different click handler
                    const button = createLocationButton(
                        building.name,
                        () => showBuildingClosedMessage(building),
                        'building-closed'
                    );
                    buildingButtonContainer.appendChild(button);
                }
            });
        } else {
            // No buildings in this district
            const emptyMessage = document.createElement('div');
            emptyMessage.className = 'text-gray-400 text-xs p-2 text-center italic';
            emptyMessage.textContent = 'No buildings here.';
            buildingButtonContainer.appendChild(emptyMessage);
        }
    }

    // 3. NPC BUTTONS (only district-level NPCs, not building NPCs)
    if (npcContainer) {
        const npcButtonContainer = npcContainer.querySelector('div');

        if (npcs && npcs.length > 0) {
            npcs.forEach(npcId => {
                const npcData = getNPCById(npcId);
                const displayName = npcData ? npcData.name : npcId.replace(/_/g, ' ');
                const button = createLocationButton(
                    displayName,
                    () => talkToNPC(npcId),
                    'npc'
                );
                npcButtonContainer.appendChild(button);
            });
        } else {
            // Show message when no NPCs in district (they're all in buildings)
            const emptyMessage = document.createElement('div');
            emptyMessage.className = 'text-gray-400 text-xs p-2 text-center italic';
            emptyMessage.textContent = 'No one here. Check buildings.';
            npcButtonContainer.appendChild(emptyMessage);
        }
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

// Create a location button with consistent styling
function createLocationButton(text, onClick, type = 'navigation') {
    const button = document.createElement('button');

    // Different colors for different types
    const typeStyles = {
        navigation: 'bg-blue-700 hover:bg-blue-600 border-blue-900',      // City districts - blue
        environment: 'bg-red-700 hover:bg-red-600 border-red-900',        // Outside city - red
        building: 'bg-green-700 hover:bg-green-600 border-green-900',     // Open buildings - green
        'building-closed': 'bg-gray-500 hover:bg-gray-600 border-gray-700 text-black',  // Closed buildings - grey with black text
        npc: 'bg-purple-700 hover:bg-purple-600 border-purple-900'
    };

    const colorClass = typeStyles[type] || typeStyles.navigation;

    // For closed buildings, text color is already in colorClass, otherwise default to white
    const textColor = type === 'building-closed' ? '' : 'text-white';

    button.className = `${colorClass} ${textColor} border-2 px-2 py-1 font-bold transition-all leading-tight text-center overflow-hidden`;
    button.style.fontSize = '7px';
    button.style.imageRendering = 'pixelated';
    button.style.clipPath = 'polygon(0 2px, 2px 2px, 2px 0, calc(100% - 2px) 0, calc(100% - 2px) 2px, 100% 2px, 100% calc(100% - 2px), calc(100% - 2px) calc(100% - 2px), calc(100% - 2px) 100%, 2px 100%, 2px calc(100% - 2px), 0 calc(100% - 2px))';
    button.style.overflowWrap = 'break-word';
    button.style.hyphens = 'none';
    button.style.display = 'flex';
    button.style.alignItems = 'center';
    button.style.justifyContent = 'center';
    button.textContent = text;
    button.addEventListener('click', onClick);
    return button;
}

// Check if a building is currently open based on time of day
function isBuildingOpen(building, currentTime) {
    // Always open buildings
    if (building.open === 'always') {
        return true;
    }

    const openTime = building.open;
    const closeTime = building.close;

    // No hours specified - assume always open
    if (openTime === undefined) {
        return true;
    }

    // Convert old time values (0-11) to new 24-hour format if needed
    // Old buildings may still use 0-11 values, so we need to handle both
    // In 24-hour system: buildings should use actual hours (e.g., 6 = 6 AM, 14 = 2 PM)

    // Open rest of day (close is null)
    if (closeTime === null) {
        return currentTime >= openTime;
    }

    // Check if open hours wrap around midnight
    if (openTime < closeTime) {
        // Normal hours (e.g., 6-18: 6 AM to 6 PM)
        return currentTime >= openTime && currentTime < closeTime;
    } else {
        // Overnight hours (e.g., 20-6: 8 PM through night to 6 AM)
        return currentTime >= openTime || currentTime < closeTime;
    }
}

// Show message when clicking a closed building
function showBuildingClosedMessage(building) {
    const openTimeName = building.open === 'always' ? 'always' : formatTime(building.open);

    showMessage(`🔒 ${building.name} is closed. Opens at ${openTimeName}.`, 'error');
}

// Enter a building
async function enterBuilding(buildingId) {
    console.log('Entering building:', buildingId);

    try {
        await window.gameAPI.sendAction('enter_building', { building_id: buildingId });
        await refreshGameState();
        displayCurrentLocation();
    } catch (error) {
        console.error('Failed to enter building:', error);
        showMessage('❌ Failed to enter building', 'error');
    }
}

// Exit a building
async function exitBuilding() {
    console.log('Exiting building');

    try {
        await window.gameAPI.sendAction('exit_building', {});
        await refreshGameState();
        displayCurrentLocation();
    } catch (error) {
        console.error('Failed to exit building:', error);
        showMessage('❌ Failed to exit building', 'error');
    }
}

async function talkToNPC(npcId) {
    console.log('💬 Initiating dialogue with NPC:', npcId);

    try {
        const result = await window.gameAPI.sendAction('talk_to_npc', { npc_id: npcId });

        if (result.success && result.delta?.npc_dialogue) {
            showNPCDialogue(result.delta.npc_dialogue, result.message);
        } else {
            showMessage('❌ Failed to talk to NPC', 'error');
        }
    } catch (error) {
        console.error('Error talking to NPC:', error);
        showMessage('❌ Failed to talk to NPC', 'error');
    }
}

// Show NPC dialogue UI - replaces bottom UI with dialogue options
function showNPCDialogue(dialogueData, npcMessage) {
    console.log('📜 Showing NPC dialogue:', dialogueData);

    // Show NPC message in yellow
    if (npcMessage) {
        showMessage(npcMessage, 'warning'); // warning = yellow
    }

    // Get the action-buttons container (parent of all three columns)
    const actionButtonsArea = document.getElementById('action-buttons');
    if (!actionButtonsArea) {
        console.error('❌ action-buttons container not found!');
        return;
    }

    // Hide the normal action buttons
    actionButtonsArea.style.display = 'none';

    // Create or get dialogue overlay container
    let dialogueOverlay = document.getElementById('npc-dialogue-overlay');
    if (!dialogueOverlay) {
        dialogueOverlay = document.createElement('div');
        dialogueOverlay.id = 'npc-dialogue-overlay';
        dialogueOverlay.className = 'p-4 bg-gray-800 border-t-4 border-yellow-500';
        dialogueOverlay.style.height = '125px'; // Match action-buttons height

        // Insert right after action-buttons
        actionButtonsArea.parentNode.insertBefore(dialogueOverlay, actionButtonsArea.nextSibling);
    }

    // Clear previous content
    dialogueOverlay.innerHTML = '';

    // Create dialogue options grid
    const optionsGrid = document.createElement('div');
    optionsGrid.className = 'grid grid-cols-3 gap-1';

    // Add dialogue option buttons
    if (dialogueData.options && dialogueData.options.length > 0) {
        dialogueData.options.forEach(optionKey => {
            const button = document.createElement('button');
            button.className = 'bg-yellow-700 hover:bg-yellow-600 border-2 border-yellow-900 text-white px-2 py-1 font-bold transition-all';
            button.style.fontSize = '10px';
            button.style.clipPath = 'polygon(0 2px, 2px 2px, 2px 0, calc(100% - 2px) 0, calc(100% - 2px) 2px, 100% 2px, 100% calc(100% - 2px), calc(100% - 2px) calc(100% - 2px), calc(100% - 2px) 100%, 2px 100%, 2px calc(100% - 2px), 0 calc(100% - 2px))';

            // Format option text (convert snake_case to readable)
            const optionText = formatDialogueOption(optionKey);
            button.textContent = optionText;

            button.addEventListener('click', () => selectDialogueOption(dialogueData.npc_id, optionKey));
            optionsGrid.appendChild(button);
        });
    } else {
        console.warn('⚠️ No dialogue options provided!');
    }

    dialogueOverlay.appendChild(optionsGrid);
    dialogueOverlay.style.display = 'block';

    console.log('✅ Dialogue overlay created with', dialogueData.options?.length || 0, 'options');
}

// Format dialogue option key to readable text
function formatDialogueOption(optionKey) {
    const optionNames = {
        'ask_about_fee': 'Ask about fee',
        'ask_about_tribute': 'Ask about tribute',
        'pay_fee': 'Pay the fee',
        'pay_tribute': 'Pay tribute',
        'use_storage': 'Use storage',
        'maybe_later': 'Maybe later',
        'goodbye': 'Goodbye'
    };

    return optionNames[optionKey] || optionKey.replace(/_/g, ' ');
}

// Handle dialogue option selection
async function selectDialogueOption(npcId, choice) {
    console.log('💬 Selected dialogue option:', choice);

    try {
        const result = await window.gameAPI.sendAction('npc_dialogue_choice', {
            npc_id: npcId,
            choice: choice
        });

        if (result.success) {
            // Show NPC response in yellow
            if (result.message) {
                showMessage(result.message, 'warning');
            }

            // Check if vault should open (check this first before close action)
            if (result.delta?.open_vault) {
                console.log('🏦 Opening vault with data:', result.delta.open_vault);
                closeNPCDialogue();
                await refreshGameState();
                showVaultUI(result.delta.open_vault);
            }
            // Check if dialogue should close
            else if (result.delta?.npc_dialogue?.action === 'close') {
                closeNPCDialogue();
                await refreshGameState();
            }
            // Continue dialogue with new options
            else if (result.delta?.npc_dialogue) {
                showNPCDialogue(result.delta.npc_dialogue, result.message);
            }
        } else {
            console.error('❌ Dialogue option failed:', result.error);
            showMessage(result.error || 'Dialogue option failed', 'error');
        }
    } catch (error) {
        console.error('Error selecting dialogue option:', error);
        showMessage('❌ Failed to process dialogue', 'error');
    }
}

// Close NPC dialogue and restore normal UI
function closeNPCDialogue() {
    console.log('🚪 Closing NPC dialogue');

    // Hide dialogue overlay
    const dialogueOverlay = document.getElementById('npc-dialogue-overlay');
    if (dialogueOverlay) {
        dialogueOverlay.style.display = 'none';
    }

    // Restore action buttons
    const actionButtonsArea = document.getElementById('action-buttons');
    if (actionButtonsArea) {
        actionButtonsArea.style.display = 'grid'; // Restore grid display
    }

    console.log('✅ Dialogue closed, action buttons restored');
}

// Show vault UI overlay (40 slots over main scene)
function showVaultUI(vaultData) {
    // Get scene container to overlay on top of it
    const sceneContainer = document.getElementById('scene-container');
    if (!sceneContainer) return;

    // Create or get vault overlay
    let vaultOverlay = document.getElementById('vault-overlay');
    if (!vaultOverlay) {
        vaultOverlay = document.createElement('div');
        vaultOverlay.id = 'vault-overlay';
        vaultOverlay.className = 'absolute inset-0 bg-black bg-opacity-90 flex items-center justify-center';
        vaultOverlay.style.zIndex = '100';
        sceneContainer.appendChild(vaultOverlay);
    }

    // Clear previous content
    vaultOverlay.innerHTML = '';

    // Create vault container
    const vaultContainer = document.createElement('div');
    vaultContainer.className = 'p-2 w-full h-full flex flex-col';

    // Header
    const header = document.createElement('div');
    header.className = 'flex justify-between items-center mb-2';

    const title = document.createElement('h2');
    title.className = 'text-yellow-400 font-bold';
    title.style.fontSize = '12px';
    title.textContent = '🏦 Vault Storage';

    const closeButton = document.createElement('button');
    closeButton.className = 'text-white px-2 py-1 font-bold';
    closeButton.style.cssText = 'background: #dc2626; border-top: 2px solid #ef4444; border-left: 2px solid #ef4444; border-right: 2px solid #991b1b; border-bottom: 2px solid #991b1b; font-size: 10px;';
    closeButton.textContent = 'Close';
    closeButton.addEventListener('click', closeVaultUI);

    header.appendChild(title);
    header.appendChild(closeButton);
    vaultContainer.appendChild(header);

    // Vault slots grid (40 slots in 8x5 grid)
    const slotsGrid = document.createElement('div');
    slotsGrid.className = 'grid grid-cols-8 gap-1 flex-1';
    slotsGrid.id = 'vault-slots-grid';
    slotsGrid.style.gridAutoRows = '1fr';

    const slots = vaultData.slots || [];
    for (let i = 0; i < 40; i++) {
        const slotData = slots[i] || { slot: i, item: null, quantity: 0 };
        const slotElement = createVaultSlot(slotData, i, vaultData.building || vaultData.location);
        slotsGrid.appendChild(slotElement);
    }

    vaultContainer.appendChild(slotsGrid);

    // Instructions
    const instructions = document.createElement('div');
    instructions.className = 'text-gray-400 text-center mt-1';
    instructions.style.fontSize = '8px';
    instructions.textContent = 'Click inventory items to store. Click vault items to withdraw.';
    vaultContainer.appendChild(instructions);

    vaultOverlay.appendChild(vaultContainer);
    vaultOverlay.style.display = 'flex';

    // Mark vault as open
    window.vaultOpen = true;

    // Force DOM to update before binding events
    requestAnimationFrame(() => {
        if (window.inventoryInteractions && window.inventoryInteractions.bindInventoryEvents) {
            window.inventoryInteractions.bindInventoryEvents();
        }
    });
}

// Create a single vault slot element (styled like backpack slots)
function createVaultSlot(slotData, slotIndex, buildingId) {
    const slot = document.createElement('div');
    slot.className = 'vault-slot relative cursor-pointer hover:bg-gray-800 flex items-center justify-center';
    // Match backpack slot styling exactly
    slot.style.cssText = `aspect-ratio: 1; background: #1a1a1a; border-top: 2px solid #000000; border-left: 2px solid #000000; border-right: 2px solid #3a3a3a; border-bottom: 2px solid #3a3a3a; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);`;

    // Data attributes for drag-and-drop
    slot.setAttribute('data-vault-slot', slotIndex);
    slot.setAttribute('data-vault-building', buildingId);
    slot.setAttribute('data-slot-type', 'vault');

    if (slotData.item && slotData.quantity > 0) {
        slot.setAttribute('data-item-id', slotData.item);

        // Create image container
        const imgDiv = document.createElement('div');
        imgDiv.className = 'w-full h-full flex items-center justify-center p-1';
        const img = document.createElement('img');
        img.src = `/res/img/items/${slotData.item}.png`;
        img.alt = slotData.item;
        img.className = 'w-full h-full object-contain';
        img.style.imageRendering = 'pixelated';
        imgDiv.appendChild(img);
        slot.appendChild(imgDiv);

        // Add quantity label if > 1
        if (slotData.quantity > 1) {
            const quantityLabel = document.createElement('div');
            quantityLabel.className = 'absolute bottom-0 right-0 text-white';
            quantityLabel.style.fontSize = '10px';
            quantityLabel.textContent = `${slotData.quantity}`;
            slot.appendChild(quantityLabel);
        }
    }

    return slot;
}

// Close vault UI
function closeVaultUI() {
    const vaultOverlay = document.getElementById('vault-overlay');
    if (vaultOverlay) {
        vaultOverlay.style.display = 'none';
    }

    // Mark vault as closed
    window.vaultOpen = false;

    // Refresh inventory display
    updateCharacterDisplay();
}

// Update combat interface (combat system not implemented yet)
function updateCombatInterface() {
    const combatInterface = document.getElementById('combat-interface');
    if (combatInterface) {
        combatInterface.classList.add('hidden');
    }
    // TODO: Implement combat system in Go backend
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

// Update time of day display
function updateTimeDisplay() {
    const state = getGameStateSync();
    const timeOfDay = state.character?.time_of_day !== undefined ? state.character.time_of_day : 12;
    const currentDay = state.character?.current_day || 1;

    // Map time (0-23) to 12 PNG filenames (each image covers 2 hours)
    const timeImages = [
        '00-midnight.png',     // 0-1 (12 AM - 1 AM)
        '01-twilight.png',     // 2-3
        '02-witching.png',     // 4-5
        '03-dawn.png',         // 6-7
        '04-morning.png',      // 8-9
        '05-latemorning.png',  // 10-11
        '06-highnoon.png',     // 12-13 (12 PM - 1 PM)
        '07-midday.png',       // 14-15
        '08-afternoon.png',    // 16-17
        '09-golden.png',       // 18-19
        '10-dusk.png',         // 20-21
        '11-evening.png'       // 22-23
    ];

    // Calculate which image to use (divide by 2 since each image covers 2 hours)
    const imageIndex = Math.floor(timeOfDay / 2);

    // Update time image
    const timeImage = document.getElementById('time-of-day-image');
    if (timeImage) {
        const imageName = timeImages[imageIndex] || timeImages[6]; // Default to noon if invalid
        timeImage.src = `/res/img/time/${imageName}`;
        timeImage.alt = `Time: ${formatTime(timeOfDay)}`;
    }

    // Update time text (AM/PM format)
    const timeText = document.getElementById('time-of-day-text');
    if (timeText) {
        timeText.textContent = formatTime(timeOfDay);
    }

    // Update day counter
    const dayCounter = document.getElementById('day-counter');
    if (dayCounter) {
        dayCounter.textContent = `Day ${currentDay}`;
    }
}

// Helper function to format time in AM/PM format
function formatTime(hour) {
    if (hour === 0) return '12 AM';
    if (hour < 12) return `${hour} AM`;
    if (hour === 12) return '12 PM';
    return `${hour - 12} PM`;
}

// Update all displays
function updateAllDisplays() {
    updateCharacterDisplay();
    updateInventoryDisplay();
    updateSpellsDisplay();
    displayCurrentLocation();
    updateTimeDisplay();

    const state = getGameStateSync();
    if (state.combat) {
        updateCombatInterface();
    }
}

// Save game to Nostr relay
async function saveGameToRelay() {
    const state = getGameStateSync();
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

// Export game log function globally for use in other modules
window.addGameLog = addGameLog;

// ========================================
// Ground Items Modal
// ========================================

/**
 * Open ground items modal
 */
function openGroundModal() {
    // Get ground items at current location
    const groundItems = window.getGroundItems();

    // Create modal backdrop (only over scene area)
    const modal = document.createElement('div');
    modal.id = 'ground-modal';
    modal.style.position = 'absolute';
    modal.style.top = '0';
    modal.style.left = '0';
    modal.style.width = '100%';
    modal.style.height = '100%';
    modal.style.zIndex = '50';
    modal.style.background = 'rgba(0, 0, 0, 0.95)';
    modal.style.border = '2px solid #4a4a4a';
    modal.style.boxShadow = 'inset 1px 1px 0 #3a3a3a, inset -1px -1px 0 #000000';

    // Header with title and close button
    const header = document.createElement('div');
    header.className = 'flex items-center justify-between p-2';
    header.style.background = '#2a2a2a';
    header.style.borderBottom = '2px solid #4a4a4a';
    header.innerHTML = `
        <span class="text-white font-bold" style="font-size: 10px;">ITEMS ON GROUND</span>
        <button onclick="closeGroundModal()" class="text-white font-bold px-2 hover:bg-gray-700" style="font-size: 12px;">✕</button>
    `;

    // Content wrapper with centered grid
    const contentWrapper = document.createElement('div');
    contentWrapper.style.height = 'calc(100% - 36px)';
    contentWrapper.style.display = 'flex';
    contentWrapper.style.alignItems = 'flex-start';
    contentWrapper.style.justifyContent = 'center';
    contentWrapper.style.padding = '20px';
    contentWrapper.style.overflowY = 'auto';
    contentWrapper.style.boxSizing = 'border-box';

    // Grid container (4 columns, max width to match inventory sizing)
    const gridContainer = document.createElement('div');
    gridContainer.style.display = 'grid';
    gridContainer.style.gridTemplateColumns = 'repeat(4, 1fr)';
    gridContainer.style.gap = '4px';
    gridContainer.style.maxWidth = '200px';
    gridContainer.style.width = '100%';

    if (groundItems.length === 0) {
        const emptyMsg = document.createElement('div');
        emptyMsg.className = 'text-center text-gray-500 mt-8';
        emptyMsg.style.fontSize = '10px';
        emptyMsg.style.gridColumn = '1 / -1';
        emptyMsg.textContent = 'Nothing on the ground';
        gridContainer.appendChild(emptyMsg);
    } else {
        // Render each ground item (matching inventory slot style)
        groundItems.forEach(ground => {
            const itemSlot = document.createElement('div');
            itemSlot.className = 'relative cursor-pointer hover:bg-gray-600 flex items-center justify-center';
            itemSlot.style.cssText = `aspect-ratio: 1; background: #2a2a2a; border-top: 2px solid #1a1a1a; border-left: 2px solid #1a1a1a; border-right: 2px solid #4a4a4a; border-bottom: 2px solid #4a4a4a; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);`;
            itemSlot.onclick = () => pickupGroundItem(ground.item);

            // Create image container
            const imgDiv = document.createElement('div');
            imgDiv.className = 'w-full h-full flex items-center justify-center p-1';
            const img = document.createElement('img');
            img.src = `/res/img/items/${ground.item}.png`;
            img.alt = ground.item;
            img.className = 'w-full h-full object-contain';
            img.style.imageRendering = 'pixelated';
            imgDiv.appendChild(img);
            itemSlot.appendChild(imgDiv);

            // Add quantity label if > 1
            if (ground.quantity > 1) {
                const quantityLabel = document.createElement('div');
                quantityLabel.className = 'absolute bottom-0 right-0 text-white';
                quantityLabel.style.fontSize = '10px';
                quantityLabel.textContent = `${ground.quantity}`;
                itemSlot.appendChild(quantityLabel);
            }

            gridContainer.appendChild(itemSlot);
        });
    }

    // Assemble modal structure
    contentWrapper.appendChild(gridContainer);
    modal.appendChild(header);
    modal.appendChild(contentWrapper);

    // Add to scene container (not body)
    const sceneContainer = document.querySelector('#game-window .flex.flex-1 > div[style*="width: 556px"] > div[style*="height: 347px"] > div[style*="width: 347px"]');
    if (sceneContainer) {
        sceneContainer.style.position = 'relative';
        sceneContainer.appendChild(modal);
    } else {
        // Fallback: add to body if scene container not found
        document.body.appendChild(modal);
    }

    console.log('📦 Ground modal opened:', groundItems.length, 'items');
}

/**
 * Close ground items modal
 */
function closeGroundModal() {
    const modal = document.getElementById('ground-modal');
    if (modal) {
        modal.remove();
    }
}

/**
 * Refresh ground modal if it's open
 */
function refreshGroundModal() {
    const modal = document.getElementById('ground-modal');
    if (modal) {
        // Modal is open, refresh it
        closeGroundModal();
        openGroundModal();
    }
}

/**
 * Pick up an item from the ground
 */
async function pickupGroundItem(itemId) {
    const groundItem = window.removeItemFromGround(itemId);

    if (!groundItem) {
        showMessage('Item not found on ground', 'error');
        return;
    }

    const itemData = getItemById(itemId);

    // Check if Game API is initialized
    if (!window.gameAPI || !window.gameAPI.initialized) {
        console.error('Game API not initialized');
        // Put back on ground
        window.addItemToGround(itemId, groundItem.quantity);
        showMessage('❌ Game not initialized', 'error');
        return;
    }

    try {
        // Use new game API
        const result = await window.gameAPI.sendAction('add_item', {
            item_id: itemId,
            quantity: groundItem.quantity
        });

        if (result.success) {
            if (window.showActionText) {
                window.showActionText(`Picked up ${itemData?.name || itemId}`, 'green');
            }

            // Refresh game state from Go memory
            await refreshGameState();

            // Refresh modal
            closeGroundModal();
            openGroundModal();
        } else {
            // Failed to add - put back on ground
            window.addItemToGround(itemId, groundItem.quantity);
            showMessage(result.error || 'Failed to pick up item', 'error');
        }
    } catch (error) {
        // Failed to add - put back on ground
        window.addItemToGround(itemId, groundItem.quantity);
        showMessage('Error picking up item: ' + error.message, 'error');
    }
}

// Export ground functions
window.openGroundModal = openGroundModal;
window.closeGroundModal = closeGroundModal;
window.refreshGroundModal = refreshGroundModal;
window.pickupGroundItem = pickupGroundItem;

// Export vault functions
window.showVaultUI = showVaultUI;
window.closeVaultUI = closeVaultUI;

console.log('Game UI functions loaded');