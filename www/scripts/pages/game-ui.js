// Game UI Functions
// Handles all user interface updates and interactions

// Item database cache
let itemsDatabaseCache = null;

// Advancement data cache
let advancementDataCache = null;

// Spell data mapping
const spellData = {
    'fire-bolt': { school: 'evocation', damage: '1d10', type: 'fire', emoji: '🔥' },
    'ray-of-frost': { school: 'evocation', damage: '1d8', type: 'cold', emoji: '❄️' },
    'shocking-grasp': { school: 'evocation', damage: '1d8', type: 'lightning', emoji: '⚡' },
    'mage-hand': { school: 'conjuration', damage: null, type: null, emoji: '' },
    'light': { school: 'evocation', damage: null, type: null, emoji: '' },
    'acid-splash': { school: 'conjuration', damage: '1d6', type: 'acid', emoji: '🧪' },
    'magic-missile': { school: 'evocation', damage: '1d4+1', type: 'force', emoji: '✨' },
    'shield': { school: 'abjuration', damage: null, type: null, emoji: '' },
    'burning-hands': { school: 'evocation', damage: '3d6', type: 'fire', emoji: '🔥' },
    'thunderwave': { school: 'evocation', damage: '2d8', type: 'thunder', emoji: '💥' },
    'mage-armor': { school: 'abjuration', damage: null, type: null, emoji: '' },
    'sleep': { school: 'enchantment', damage: null, type: null, emoji: '' },
    'detect-magic': { school: 'divination', damage: null, type: null, emoji: '' },
    'identify': { school: 'divination', damage: null, type: null, emoji: '' },
    'sacred-flame': { school: 'evocation', damage: '1d8', type: 'radiant', emoji: '✨' },
    'guidance': { school: 'divination', damage: null, type: null, emoji: '' },
    'spare-the-dying': { school: 'necromancy', damage: null, type: null, emoji: '' },
    'cure-wounds': { school: 'evocation', damage: '1d8', type: 'healing', emoji: '💚' },
    'healing-word': { school: 'evocation', damage: '1d4', type: 'healing', emoji: '💚' },
    'bless': { school: 'enchantment', damage: null, type: null, emoji: '' },
    'guiding-bolt': { school: 'evocation', damage: '4d6', type: 'radiant', emoji: '✨' },
};

/**
 * Load items from database cache
 */
async function loadItemsFromDatabase() {
  if (itemsDatabaseCache) {
    return itemsDatabaseCache;
  }

  try {
    const response = await fetch('/api/items');
    if (!response.ok) {
      throw new Error('Failed to fetch items from database');
    }
    const data = await response.json();
    console.log('Raw API response:', data);
    console.log('Type of data:', typeof data, 'Is array:', Array.isArray(data));

    // The API returns an array directly, not wrapped in an object
    itemsDatabaseCache = Array.isArray(data) ? data : (data.items || []);
    console.log('Cached items:', itemsDatabaseCache.length);
    return itemsDatabaseCache;
  } catch (error) {
    console.error('Error loading items from database:', error);
    return [];
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
    const response = await fetch('/data/character/advancement.json');
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
    console.log(`Looking for item '${itemId}' in ${items.length} items`);
    const item = items.find(i => i.id === itemId);
    if (!item) {
      console.warn(`Item '${itemId}' not found. Available IDs:`, items.slice(0, 5).map(i => i.id));
    }
    return item || null;
  } catch (error) {
    console.error(`Error getting item ${itemId}:`, error);
    return null;
  }
}

// Message system - DISABLED for work-in-progress UI
window.showMessage = function showMessage(text, type = 'info', duration = 5000) {
    // Only log to console, don't show UI notifications
    console.log(`📝 ${type.toUpperCase()}: ${text}`);
    // All notifications are disabled during work-in-progress phase
    return;
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
                        console.log(`Item ${slot.item} adds ${weightIncrease} lbs capacity`);
                        capacityBonus += weightIncrease;
                    }
                }
            }
        }
    }

    const totalCapacity = baseCapacity + capacityBonus;
    console.log(`Max capacity: ${baseCapacity} (base) + ${capacityBonus} (items) = ${totalCapacity} lbs`);
    return totalCapacity;
}

/**
 * Calculate total weight of all items in inventory
 */
async function calculateAndDisplayWeight(character) {
    if (!character.inventory) {
        console.log('No inventory found');
        return 0;
    }

    let totalWeight = 0;
    const items = await loadItemsFromDatabase();
    console.log('Calculating weight with', items.length, 'items in database');

    // Function to get item weight by ID
    const getItemWeight = (itemId) => {
        const item = items.find(i => i.id === itemId);
        if (!item) {
            console.log(`Item ${itemId}: not found`);
            return 0;
        }
        // Weight can be at top level or in properties object
        const weight = item.weight || item.properties?.weight || 0;
        console.log(`Item ${itemId}: weight = ${weight}`);
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
                console.log(`Gear slot ${slotName}: ${slot.item} x${slot.quantity || 1} = ${itemWeight} lb`);
                totalWeight += itemWeight;
            }
        }

        // Bag itself and its contents
        if (gearSlots.bag) {
            // Add bag weight
            if (gearSlots.bag.item) {
                const bagWeight = getItemWeight(gearSlots.bag.item);
                console.log(`Bag itself: ${gearSlots.bag.item} = ${bagWeight} lb`);
                totalWeight += bagWeight;
            }

            // Add contents weight
            if (gearSlots.bag.contents && Array.isArray(gearSlots.bag.contents)) {
                console.log(`Bag has ${gearSlots.bag.contents.length} slots`);
                gearSlots.bag.contents.forEach(slot => {
                    if (slot && slot.item) {
                        const weight = getItemWeight(slot.item);
                        const itemWeight = weight * (slot.quantity || 1);
                        console.log(`Bag slot: ${slot.item} x${slot.quantity || 1} = ${itemWeight} lb`);
                        totalWeight += itemWeight;
                    }
                });
            }
        }
    }

    // Calculate general slots weight
    if (character.inventory.general_slots && Array.isArray(character.inventory.general_slots)) {
        console.log(`General slots: ${character.inventory.general_slots.length}`);
        character.inventory.general_slots.forEach(slot => {
            if (slot && slot.item) {
                const weight = getItemWeight(slot.item);
                const itemWeight = weight * (slot.quantity || 1);
                console.log(`General slot: ${slot.item} x${slot.quantity || 1} = ${itemWeight} lb`);
                totalWeight += itemWeight;
            }
        });
    }

    // Add gold weight (50 gold = 1 lb)
    const goldWeight = Math.floor((character.gold || 0) / 50);
    console.log(`Gold weight: ${character.gold || 0} gp = ${goldWeight} lb`);
    totalWeight += goldWeight;

    console.log(`Total weight: ${totalWeight} lb`);
    return Math.round(totalWeight);
}

// Update full character display from save data
async function updateCharacterDisplay() {
    const state = getGameState();
    const character = state.character;

    if (!character) return;

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

    // Update fatigue (scale 0-10)
    const fatigueLevelEl = document.getElementById('fatigue-level');
    const fatigueBarEl = document.getElementById('fatigue-bar');
    const fatigueStatusEl = document.getElementById('fatigue-status');
    const fatigue = Math.min(character.fatigue || 0, 10);

    if (fatigueLevelEl) fatigueLevelEl.textContent = fatigue;
    if (fatigueBarEl) {
        const fatiguePercentage = (fatigue / 10) * 100;
        fatigueBarEl.style.width = fatiguePercentage + '%';
    }
    if (fatigueStatusEl) {
        // Update status text based on fatigue level
        if (fatigue <= 2) {
            fatigueStatusEl.textContent = 'FRESH';
        } else if (fatigue <= 5) {
            fatigueStatusEl.textContent = 'TIRED';
        } else if (fatigue <= 8) {
            fatigueStatusEl.textContent = 'WEARY';
        } else {
            fatigueStatusEl.textContent = 'EXHAUSTED';
        }
    }

    // Update weight and encumbrance status
    const weightEl = document.getElementById('char-weight');
    const maxWeightEl = document.getElementById('max-weight');
    const weightStatusEl = document.getElementById('weight-status');

    if (weightEl) {
        Promise.all([
            calculateAndDisplayWeight(character),
            calculateMaxCapacity(character)
        ]).then(([weight, maxCapacity]) => {
            weightEl.textContent = weight;

            if (maxWeightEl) {
                maxWeightEl.textContent = maxCapacity;
            }

            // Calculate encumbrance status with emoji
            const weightPercentage = (weight / maxCapacity) * 100;
            let status = '✓';
            let statusColor = '#10b981'; // green

            if (weightPercentage <= 50) {
                status = '🪶'; // feather - light
                statusColor = '#10b981'; // green
            } else if (weightPercentage <= 100) {
                status = '✓'; // checkmark - ok
                statusColor = '#ffffff'; // white
            } else if (weightPercentage <= 150) {
                status = '📦'; // box - heavy
                statusColor = '#eab308'; // yellow
            } else if (weightPercentage <= 200) {
                status = '🐌'; // snail - slow
                statusColor = '#f97316'; // orange
            } else {
                status = '🛑'; // stop sign - max
                statusColor = '#ef4444'; // red
            }

            if (weightStatusEl) {
                weightStatusEl.textContent = status;
                weightStatusEl.style.color = statusColor;
            }
        });
    }

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
                console.log(`Equipment slot ${slotName}:`, itemId);

                if (itemId) {
                    // Fetch item data
                    const itemData = await getItemByIdAsync(itemId);
                    console.log(`Item data for ${itemId}:`, itemData);

                    if (itemData) {
                        // Replace placeholder with item image
                        const imageContainer = slotEl.querySelector('.w-10.h-10');
                        if (imageContainer) {
                            imageContainer.innerHTML = `<img src="/res/img/items/${itemId}.png" alt="${itemData.name}" class="w-full h-full object-contain" style="image-rendering: pixelated;">`;
                            console.log(`✅ Loaded image for ${slotName}: ${itemId}`);
                        } else {
                            console.warn(`⚠️ No image container found for ${slotName}`);
                        }
                    }
                } else {
                    // Reset to placeholder if empty
                    const imageContainer = slotEl.querySelector('.w-10.h-10');
                    const placeholderIcon = slotEl.querySelector('.placeholder-icon');
                    if (imageContainer && placeholderIcon) {
                        // Placeholder is already there, just make sure it's visible
                        placeholderIcon.style.display = 'block';
                    }
                }
            }
        }
    }

    // Update general slots (4x1 grid)
    if (character.inventory && character.inventory.general_slots) {
        const generalSlotsDiv = document.getElementById('general-slots');
        if (generalSlotsDiv) {
            generalSlotsDiv.innerHTML = '';

            // Create all 4 general slots
            for (let i = 0; i < 4; i++) {
                const slot = character.inventory.general_slots[i];
                const slotDiv = document.createElement('div');
                slotDiv.className = 'relative cursor-pointer hover:bg-gray-600 flex items-center justify-center';
                slotDiv.style.cssText = `aspect-ratio: 1; background: #2a2a2a; border-top: 2px solid #1a1a1a; border-left: 2px solid #1a1a1a; border-right: 2px solid #4a4a4a; border-bottom: 2px solid #4a4a4a; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);`;

                if (slot && slot.item) {
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
                        quantityLabel.className = 'absolute bottom-0 right-0 text-blue-400 font-bold';
                        quantityLabel.style.fontSize = '10px';
                        quantityLabel.textContent = `x${slot.quantity}`;
                        slotDiv.appendChild(quantityLabel);
                    }
                }

                generalSlotsDiv.appendChild(slotDiv);
            }
        }
    }

    // Update backpack items (4x5 grid = 20 slots)
    if (character.inventory && character.inventory.gear_slots?.bag?.contents) {
        const backpackDiv = document.getElementById('backpack-slots');
        if (backpackDiv) {
            const contents = character.inventory.gear_slots.bag.contents;
            backpackDiv.innerHTML = '';

            let itemCount = 0;

            // Create all 20 backpack slots
            for (let i = 0; i < 20; i++) {
                const slot = contents[i];
                const slotDiv = document.createElement('div');
                slotDiv.className = 'relative cursor-pointer hover:bg-gray-800 flex items-center justify-center';
                slotDiv.style.cssText = `aspect-ratio: 1; background: #1a1a1a; border-top: 2px solid #000000; border-left: 2px solid #000000; border-right: 2px solid #3a3a3a; border-bottom: 2px solid #3a3a3a; clip-path: polygon(3px 0, calc(100% - 3px) 0, 100% 3px, 100% calc(100% - 3px), calc(100% - 3px) 100%, 3px 100%, 0 calc(100% - 3px), 0 3px);`;

                if (slot && slot.item) {
                    itemCount++;

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
                        quantityLabel.className = 'absolute bottom-0 right-0 text-blue-400 font-bold';
                        quantityLabel.style.fontSize = '10px';
                        quantityLabel.textContent = `x${slot.quantity}`;
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
    }
}

// Update inventory display
function updateInventoryDisplay() {
    // Inventory is now displayed in updateCharacterDisplay via general_slots and backpack_slots
    // This function is kept for compatibility but does nothing
}

// Update spells display
function updateSpellsDisplay() {
    const state = getGameState();
    const character = state.character;

    // Update known spells
    const knownSpellsEl = document.getElementById('known-spells');
    if (knownSpellsEl && character?.spells) {
        knownSpellsEl.innerHTML = '';

        // Character.spells is an array of spell IDs
        const spellsArray = Array.isArray(character.spells) ? character.spells : [];

        spellsArray.forEach(spellId => {
            const spell = spellData[spellId] || { school: 'evocation', damage: null, type: null, emoji: '' };

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
    const state = getGameState();
    let currentLocationId = state.location?.current;

    if (!currentLocationId) return;

    // If location is just a city name (e.g., "kingdom"), default to center district
    let locationData = getLocationById(currentLocationId);
    if (!locationData) {
        // Try appending -center to get the district
        const centerLocationId = currentLocationId + '-center';
        locationData = getLocationById(centerLocationId);
        if (locationData) {
            currentLocationId = centerLocationId;
            // Update the game state to use the correct location
            updateGameState({ location: { current: currentLocationId, discovered: [currentLocationId] } });
        }
    }

    console.log('Current location ID:', currentLocationId);
    console.log('Location data:', locationData);
    if (!locationData) return;

    // Get the parent city data to use its image and music
    let cityData = locationData;

    // If this is a district, find the parent city
    if (currentLocationId.includes('-')) {
        const cityId = currentLocationId.split('-')[0]; // e.g., "kingdom" from "kingdom-north"
        const parentCity = getLocationById(cityId);
        if (parentCity) {
            cityData = parentCity;
        }
    }

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

    // Update scene title
    const sceneTitle = document.getElementById('scene-title');
    if (sceneTitle) {
        sceneTitle.textContent = locationData.name;
    }

    // Update game text with location description
    const gameText = document.getElementById('game-text');
    if (gameText) {
        gameText.innerHTML = `<p>${locationData.description || 'A mysterious place...'}</p>`;
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
    const buildings = currentData.buildings || currentData.properties?.buildings;
    const npcs = currentData.npcs || currentData.properties?.npcs;

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
    if (buildings && buildingContainer) {
        const buildingButtonContainer = buildingContainer.querySelector('div');

        buildings.forEach(building => {
            if (building.accessible) {
                const button = createLocationButton(
                    building.name,
                    () => enterBuilding(building.id),
                    'building'
                );
                buildingButtonContainer.appendChild(button);
            }
        });
    }

    // 3. NPC BUTTONS
    if (npcs && npcContainer) {
        const npcButtonContainer = npcContainer.querySelector('div');

        npcs.forEach(npcId => {
            const button = createLocationButton(
                `${npcId.replace(/_/g, ' ')}`,
                () => talkToNPC(npcId),
                'npc'
            );
            npcButtonContainer.appendChild(button);
        });
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
        building: 'bg-green-700 hover:bg-green-600 border-green-900',
        npc: 'bg-purple-700 hover:bg-purple-600 border-purple-900'
    };

    const colorClass = typeStyles[type] || typeStyles.navigation;

    button.className = `${colorClass} text-white border-2 px-2 py-1 font-bold transition-all leading-tight text-center overflow-hidden`;
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

// Placeholder functions for building and NPC interactions
function enterBuilding(buildingId) {
    console.log('Entering building:', buildingId);
    showMessage(`🏛️ Entering ${buildingId}...`, 'info');
    // TODO: Implement building interaction
}

function talkToNPC(npcId) {
    console.log('Talking to NPC:', npcId);
    showMessage(`💬 Talking to ${npcId.replace(/_/g, ' ')}...`, 'info');
    // TODO: Implement NPC dialogue
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