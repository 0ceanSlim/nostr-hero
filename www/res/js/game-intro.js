// ============================================================================
// NOSTR HERO - GAME INTRO
// Handles the character introduction sequence and equipment selection
// ============================================================================

// ============================================================================
// GLOBAL STATE
// ============================================================================
let generatedCharacter = null;
let introData = null;
let startingEquipment = null;
let selectedEquipment = {};
let playerName = '';

// ============================================================================
// MUSIC PLAYLIST SYSTEM
// ============================================================================
let currentTrack = 0;
const tracks = ['intro-music', 'game-music'];

function setupMusicPlaylist() {
  tracks.forEach((trackId, index) => {
    const audio = document.getElementById(trackId);
    audio.volume = 0.3;

    audio.addEventListener('ended', () => {
      // Move to next track
      currentTrack = (currentTrack + 1) % tracks.length;
      const nextAudio = document.getElementById(tracks[currentTrack]);
      nextAudio.currentTime = 0;
      nextAudio.play().catch(e => console.log('Music autoplay blocked:', e));
    });
  });
}

function startMusic() {
  const firstTrack = document.getElementById(tracks[0]);
  firstTrack.play().catch(e => console.log('Music autoplay blocked:', e));
}

// ============================================================================
// SCENE DISPLAY SYSTEM
// ============================================================================

/**
 * Display a scene with background image and text, with Continue button
 * @param {Object} config - Scene configuration
 * @param {string} config.text - Text to display
 * @param {string} config.image - Background image filename (optional)
 * @param {boolean} config.isQuote - Large centered quote style
 * @param {boolean} config.isLetter - Letter styling
 * @param {number} config.buttonDelay - Delay before showing Continue button (ms, default 7000)
 */
async function showScene(config) {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  // Set up background
  if (config.image) {
    background.style.backgroundImage = `url(/res/img/scene/${config.image})`;
  } else {
    background.style.backgroundImage = '';
  }

  // Clear and set up content
  content.innerHTML = '';

  if (config.isLetter) {
    // Letter scene - special styling
    const letterDiv = document.createElement('div');
    letterDiv.className = 'letter-container';
    letterDiv.innerHTML = `
      <div class="text-sm opacity-70 mb-4 text-center">The Letter:</div>
      <div class="leading-relaxed">${config.text}</div>
    `;
    content.appendChild(letterDiv);
  } else {
    // Regular scene
    const textElement = document.createElement('div');
    textElement.className = 'scene-text';

    if (config.isQuote) {
      // Quote styling - large, centered, yellow
      textElement.className += ' text-3xl md:text-5xl font-bold text-yellow-400 leading-relaxed';
    } else {
      // Normal scene text - slightly larger size
      textElement.className += ' text-2xl md:text-3xl leading-relaxed';
    }

    textElement.textContent = config.text;
    content.appendChild(textElement);
  }

  // Add Continue button using component
  const buttonDelay = config.buttonDelay !== undefined ? config.buttonDelay : 7000;
  const continueBtn = createContinueButton(buttonDelay);
  content.appendChild(continueBtn);

  // Show container with fade-in
  container.classList.remove('hidden', 'fade-out');
  container.classList.add('fade-in');

  // Wait for user to click Continue
  await waitForButtonClick(continueBtn);

  // Animate text out first (wipe down)
  const textElements = content.querySelectorAll('.scene-text, .letter-container, .pixel-continue-btn');
  textElements.forEach(el => {
    el.style.animation = 'wipeOut 0.6s ease-in forwards';
  });

  // Wait for text animation to complete
  await new Promise(resolve => setTimeout(resolve, 600));

  // Then fade out the scene
  await new Promise(resolve => setTimeout(resolve, 1000));

  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));
  container.classList.add('hidden');
}

/**
 * Show final scene with button
 */
async function showFinalScene(text, buttonText, buttonAction) {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  background.style.backgroundImage = '';

  content.innerHTML = `
    <div class="scene-text text-2xl md:text-4xl font-bold text-yellow-400 leading-relaxed mb-8">
      ${text}
    </div>
    <button
      onclick="${buttonAction}"
      class="bg-green-600 hover:bg-green-700 text-black px-8 py-4 rounded-lg font-bold text-xl transition-colors mt-8"
    >
      ${buttonText}
    </button>
  `;

  container.classList.remove('hidden', 'fade-out');
  container.classList.add('fade-in');
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Get equipment category from character class
 */
function getEquipmentCategory(className) {
  const categories = {
    'Fighter': 'warrior',
    'Barbarian': 'warrior',
    'Paladin': 'warrior',
    'Cleric': 'faithful',
    'Monk': 'faithful',
    'Ranger': 'wilderness',
    'Druid': 'wilderness',
    'Wizard': 'arcane',
    'Sorcerer': 'arcane',
    'Warlock': 'arcane',
    'Rogue': 'clever',
    'Bard': 'clever'
  };
  return categories[className] || 'warrior';
}

/**
 * Get item image name (convert item name to image filename)
 */
function getItemImageName(itemName) {
  return itemName.toLowerCase().replace(/[',]/g, '').replace(/\s+/g, '-');
}

// ============================================================================
// INTRO SEQUENCE
// ============================================================================

/**
 * Main intro sequence - shows all story scenes
 */
async function startIntroSequence() {
  playerName = document.getElementById('character-name').value.trim();

  if (!playerName) {
    alert('Please enter your name');
    return;
  }

  // Hide name screen
  document.getElementById('name-screen').classList.add('hidden');

  // Load intro data
  const response = await fetch('/data/character/introductions.json');
  introData = await response.json();

  // 1. Scene 1 - Rainy Night
  await showScene({
    text: introData.scene1.text,
    image: introData.scene1.image
  });

  // 2. Scene 2 - Caretaker's Home
  await showScene({
    text: introData.scene2.text,
    image: introData.scene2.image
  });

  // 3. Final Words (black screen quote)
  console.log('🎬 Step 3: Final Words');
  await showScene({
    text: introData.final_words.text,
    isQuote: true
  });

  // 4. Background Intro - MOVED BEFORE letter intro
  console.log('🎬 Step 4: Background Intro for:', generatedCharacter.background);
  const bgIntro = introData.background_intros.find(entry =>
    entry.backgrounds.includes(generatedCharacter.background)
  );
  console.log('🎬 Background intro data:', bgIntro);
  if (bgIntro) {
    console.log('🎬 Showing background intro scene');
    await showScene({
      text: bgIntro.text,
      image: bgIntro.image
    });
  } else {
    console.warn('⚠️ No background intro found for:', generatedCharacter.background);
  }

  // 5. Letter Intro - MOVED AFTER background intro
  console.log('🎬 Step 5: Letter Intro');
  await showScene({
    text: introData.letter_intro.text,
    image: introData.letter_intro.image
  });

  // 6. Letter Reading (scene 4a)
  console.log('🎬 Step 6: Letter Reading for:', generatedCharacter.background);
  const bgLetter = introData.background_letters.find(entry =>
    entry.backgrounds.includes(generatedCharacter.background)
  );
  console.log('🎬 Background letter data:', bgLetter);
  if (bgLetter) {
    console.log('🎬 Showing letter reading scene');
    await showScene({
      text: bgLetter.text,
      image: bgLetter.image,
      isLetter: true
    });
  } else {
    console.warn('⚠️ No background letter found for:', generatedCharacter.background);
  }

  // 7. Equipment Intro (class-based) - narrative + quote
  const equipCategory = getEquipmentCategory(generatedCharacter.class);
  const equipIntro = introData.equipment_intros[equipCategory];
  if (equipIntro) {
    // Show narrative
    await showScene({
      text: equipIntro.text,
      image: equipIntro.image
    });

    // Show quote if exists
    if (equipIntro.quote) {
      await showScene({
        text: equipIntro.quote,
        isQuote: true
      });
    }
  }

  // 8. Scene 5 - Equipment ready
  await showScene({
    text: introData.scene5.text,
    image: introData.scene5.image
  });

  // 9. Scene 5a - Preparation
  await showScene({
    text: introData.scene5a.text,
    image: introData.scene5a.image
  });

  // 10. Show equipment selection
  await startEquipmentSelection(startingEquipment);

  // Get selected equipment
  selectedEquipment = getSelectedEquipment();

  // 11. Show items given in addition to choices
  await showGivenItemsScene(startingEquipment.inventory);

  // Continue with remaining scenes
  await continueAfterEquipment();
}

/**
 * DEBUG: Skip directly to equipment selection
 */
async function skipToEquipment() {
  playerName = document.getElementById('character-name').value.trim() || 'Debug Hero';

  // Hide name screen
  document.getElementById('name-screen').classList.add('hidden');

  // Show equipment selection directly
  await startEquipmentSelection(startingEquipment);

  // Get selected equipment
  selectedEquipment = getSelectedEquipment();

  console.log('🐛 DEBUG: Equipment selection completed:', selectedEquipment);
  alert('Equipment selection complete! Check console for results.');
}

/**
 * Continue after equipment selection
 */
async function continueAfterEquipment() {

  // 12. Scene 6 - Pack note (moved before pack selection)
  await showScene({
    text: introData.scene6.text,
    image: introData.scene6.image
  });

  // 13. Pack selection or display
  await handlePackSelection(startingEquipment);

  // 14. Departure
  await showScene({
    text: introData.departure.text,
    image: introData.departure.image
  });

  // 15. Final Text + Begin Journey button
  await showFinalScene(
    introData.final_text.text,
    'Begin Journey',
    'startAdventure()'
  );
}

// ============================================================================
// GIVEN ITEMS DISPLAY
// ============================================================================

/**
 * Get rarity color based on item rarity
 */
function getRarityColor(itemName) {
  // TODO: This should come from item data
  // For now, defaulting to common (grey)
  const rarityColors = {
    'common': '#9ca3af',      // grey
    'uncommon': '#10b981',    // green
    'rare': '#3b82f6',        // blue
    'legendary': '#a855f7',   // purple
    'mythic': '#f97316'       // orange
  };
  return rarityColors.common; // Default to common for now
}

/**
 * Create a styled item card for given items display
 */
function createGivenItemCard(itemName, quantity) {
  const card = document.createElement('div');
  card.className = 'item-card bg-gray-800 rounded-lg relative overflow-hidden';
  card.style.width = '110px';
  card.style.height = '110px';
  card.style.aspectRatio = '1/1';

  // Item image (fills container with padding)
  const img = document.createElement('img');
  img.src = `/res/img/items/${getItemImageName(itemName)}.png`;
  img.alt = itemName;
  img.className = 'absolute inset-0 w-full h-full object-contain p-3';
  img.onerror = () => {
    img.src = '/res/img/otherstuff.png';
  };
  card.appendChild(img);

  // Rarity dot (top right)
  const rarityDot = document.createElement('div');
  rarityDot.className = 'absolute top-1.5 right-1.5 z-10';
  rarityDot.style.width = '10px';
  rarityDot.style.height = '10px';
  rarityDot.style.borderRadius = '50%';
  rarityDot.style.backgroundColor = getRarityColor(itemName);
  rarityDot.style.border = '1px solid rgba(0,0,0,0.3)';
  card.appendChild(rarityDot);

  // Quantity text (top left, purple)
  if (quantity && quantity > 1) {
    const qtyText = document.createElement('div');
    qtyText.className = 'absolute top-1 left-1.5 z-10 font-bold';
    qtyText.style.color = '#a855f7'; // purple
    qtyText.style.fontSize = '0.85rem';
    qtyText.style.textShadow = '0 0 3px rgba(0,0,0,0.8), 1px 1px 2px rgba(0,0,0,0.9)';
    qtyText.textContent = `×${quantity}`;
    card.appendChild(qtyText);
  }

  // Item name overlay at bottom
  const name = document.createElement('div');
  name.className = 'item-name absolute bottom-0 left-0 right-0 text-center text-white font-semibold px-1 py-1 z-10';
  name.style.fontSize = '0.65rem';
  name.style.lineHeight = '0.75rem';
  name.style.backgroundColor = 'rgba(0,0,0,0.6)';
  name.style.textShadow = '0 1px 2px rgba(0,0,0,0.8)';
  name.textContent = itemName;
  card.appendChild(name);

  // Info button (bottom right, yellow ?, hover effect)
  const infoBtn = document.createElement('button');
  infoBtn.className = 'info-btn absolute bottom-1 right-1.5 text-yellow-400 font-bold hover:text-green-400 transition-colors z-20';
  infoBtn.textContent = '?';
  infoBtn.style.fontSize = '16px';
  infoBtn.style.textShadow = '0 0 3px rgba(0,0,0,0.8), 1px 1px 2px rgba(0,0,0,0.9)';
  infoBtn.onclick = (e) => {
    e.stopPropagation();
    showItemModal(itemName);
  };
  card.appendChild(infoBtn);

  return card;
}

/**
 * Show item info modal
 */
async function showItemModal(itemName) {
  // Remove any existing modal
  hideItemModal();

  // Create modal backdrop
  const backdrop = document.createElement('div');
  backdrop.id = 'item-modal-backdrop';
  backdrop.className = 'fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50';
  backdrop.onclick = (e) => {
    if (e.target === backdrop) {
      hideItemModal();
    }
  };

  // Create modal content
  const modal = document.createElement('div');
  modal.className = 'bg-gray-800 border-2 border-yellow-400 rounded-lg p-6 max-w-md w-full mx-4 relative';
  modal.onclick = (e) => e.stopPropagation();

  // Close button
  const closeBtn = document.createElement('button');
  closeBtn.className = 'absolute top-2 right-2 w-8 h-8 bg-red-600 hover:bg-red-700 text-white rounded-full font-bold transition-colors';
  closeBtn.textContent = '✕';
  closeBtn.onclick = hideItemModal;
  modal.appendChild(closeBtn);

  // Loading state
  modal.innerHTML += '<div class="text-center text-white">Loading...</div>';

  backdrop.appendChild(modal);
  document.body.appendChild(backdrop);

  // Fetch and display item stats
  const statsHTML = await getItemStats(itemName);

  modal.innerHTML = '';
  modal.appendChild(closeBtn);

  const contentDiv = document.createElement('div');
  contentDiv.className = 'text-white';
  contentDiv.innerHTML = statsHTML;
  modal.appendChild(contentDiv);
}

/**
 * Hide item modal
 */
function hideItemModal() {
  const backdrop = document.getElementById('item-modal-backdrop');
  if (backdrop) {
    backdrop.remove();
  }
}

/**
 * Get item stats from API
 */
async function getItemStats(itemName) {
  try {
    const response = await fetch(`/api/items?name=${encodeURIComponent(itemName)}`);
    if (!response.ok) {
      return `<div class="font-bold text-yellow-400 mb-1">${itemName}</div><div class="text-gray-400 text-xs">No details available</div>`;
    }

    const items = await response.json();

    if (!items || items.length === 0) {
      return `<div class="font-bold text-yellow-400 mb-1">${itemName}</div><div class="text-gray-400 text-xs">No details available</div>`;
    }

    const itemData = items[0];
    const props = itemData.properties || {};

    let statsHTML = `<div class="font-bold text-yellow-400 mb-2 text-xl">${itemData.name}</div>`;

    if (itemData.item_type) {
      statsHTML += `<div class="text-green-400 text-sm font-semibold mb-3">${itemData.item_type}</div>`;
    }

    // Stats section
    statsHTML += `<div class="space-y-1 mb-3">`;

    if (props.damage) {
      statsHTML += `<div class="text-gray-300 text-sm">⚔️ Damage: ${props.damage} ${props['damage-type'] || ''}</div>`;
    }

    if (props.ac) {
      statsHTML += `<div class="text-gray-300 text-sm">🛡️ AC: ${props.ac}</div>`;
    }

    if (props.weight) {
      statsHTML += `<div class="text-gray-300 text-sm">⚖️ Weight: ${props.weight} lb</div>`;
    }

    statsHTML += `</div>`;

    // Add tags
    if (itemData.tags && itemData.tags.length > 0) {
      statsHTML += `<div class="flex flex-wrap gap-1 mb-3">`;
      statsHTML += itemData.tags.map(tag => `<span class="bg-gray-700 px-2 py-1 rounded text-xs text-gray-300">${tag}</span>`).join('');
      statsHTML += `</div>`;
    }

    // Add full description
    if (itemData.description) {
      statsHTML += `<div class="text-gray-300 text-sm mt-3 leading-relaxed border-t border-gray-600 pt-3">${itemData.description}</div>`;
    }

    return statsHTML;

  } catch (error) {
    console.error('❌ Error fetching item:', itemName, error);
    return `<div class="font-bold text-yellow-400 mb-1">${itemName}</div><div class="text-gray-400 text-xs">Error loading details</div>`;
  }
}

/**
 * Show scene displaying items given in addition to choices
 */
async function showGivenItemsScene(givenItems) {
  if (!givenItems || givenItems.length === 0) {
    return; // No items to show
  }

  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  // Turn off background
  background.style.backgroundImage = 'none';
  background.style.backgroundColor = '#111827';
  content.innerHTML = '';
  content.style.zIndex = '10';

  // Title
  const title = document.createElement('div');
  title.className = 'scene-text text-xl md:text-2xl font-bold text-yellow-400 mb-6';
  title.textContent = 'Items Provided';
  content.appendChild(title);

  // Description
  const description = document.createElement('div');
  description.className = 'scene-text text-lg mb-6 text-center';
  description.textContent = 'In addition to your choices, you have been provided with these items:';
  content.appendChild(description);

  // Items container
  const itemsContainer = document.createElement('div');
  itemsContainer.className = 'flex flex-wrap justify-center gap-3 mb-6 max-w-4xl mx-auto';

  // Display each given item
  for (const givenItem of givenItems) {
    const itemCard = createGivenItemCard(givenItem.item, givenItem.quantity);
    itemsContainer.appendChild(itemCard);
  }

  content.appendChild(itemsContainer);

  // Continue button
  const continueBtn = createContinueButton(2000);
  content.appendChild(continueBtn);

  // Show container with fade-in
  container.classList.remove('hidden', 'fade-out');
  container.classList.add('fade-in');

  // Wait for user to click Continue
  await waitForButtonClick(continueBtn);

  // Animate text out first (wipe down)
  const textElements = content.querySelectorAll('.scene-text, .item-card, .pixel-continue-btn');
  textElements.forEach(el => {
    el.style.animation = 'wipeOut 0.6s ease-in forwards';
  });

  // Wait for text animation to complete
  await new Promise(resolve => setTimeout(resolve, 600));

  // Clear content
  content.innerHTML = '';

  // Then fade out the scene
  await new Promise(resolve => setTimeout(resolve, 1000));

  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));

  // Fully reset container for next scene
  container.classList.remove('fade-out');
  container.classList.add('hidden');
}

// ============================================================================
// EQUIPMENT SELECTION
// ============================================================================

/**
 * Show equipment selection screen
 */
function showEquipmentSelection() {
  const equipmentSection = document.getElementById('equipment-selection');
  const equipmentChoices = document.getElementById('equipment-choices');
  const startingInventory = document.getElementById('starting-inventory');
  const continueBtn = document.getElementById('continue-adventure-btn');

  equipmentSection.className = 'flex flex-col items-center justify-center h-full p-8 overflow-y-auto fixed inset-0 bg-black';

  // Show starting inventory
  startingInventory.innerHTML = '';
  startingEquipment.inventory.forEach(item => {
    const itemDiv = document.createElement('div');
    itemDiv.className = 'bg-gray-700 rounded p-2 text-sm';

    const itemName = document.createElement('span');
    itemName.className = 'text-white';
    itemName.textContent = item.item;
    itemDiv.appendChild(itemName);

    if (item.quantity > 1) {
      const quantity = document.createElement('span');
      quantity.className = 'text-gray-400';
      quantity.textContent = ` (${item.quantity})`;
      itemDiv.appendChild(quantity);
    }

    startingInventory.appendChild(itemDiv);
  });

  // Show equipment choices
  equipmentChoices.innerHTML = '';

  if (startingEquipment.choices && startingEquipment.choices.length > 0) {
    startingEquipment.choices.forEach((choice, index) => {
      const choiceGroup = document.createElement('div');
      choiceGroup.className = 'bg-gray-700 rounded-lg p-6';

      const choiceTitle = document.createElement('h3');
      choiceTitle.className = 'text-xl font-semibold text-yellow-400 mb-4';
      choiceTitle.textContent = `Choice ${index + 1}: ${choice.description || 'Choose your equipment'}`;
      choiceGroup.appendChild(choiceTitle);

      // Handle complex weapon choices
      if (choice.type === 'complex_weapon_choice') {
        renderComplexWeaponChoice(choiceGroup, choice, index);
      } else {
        // Regular options
        choice.options.forEach((option, optionIndex) => {
          const optionDiv = document.createElement('div');
          optionDiv.className = 'bg-gray-600 hover:bg-gray-500 rounded p-4 cursor-pointer transition-colors mb-2';
          optionDiv.onclick = () => selectEquipment(index, option);

          const optionName = document.createElement('div');
          optionName.className = 'font-semibold text-white';
          optionName.textContent = option.name;

          if (option.bundle && option.bundle.length > 0) {
            const bundleList = document.createElement('div');
            bundleList.className = 'text-sm text-gray-300 mt-1';
            bundleList.textContent = option.bundle.map(item =>
              item[1] > 1 ? `${item[0]} (${item[1]})` : item[0]
            ).join(', ');
            optionDiv.appendChild(optionName);
            optionDiv.appendChild(bundleList);
          } else {
            optionDiv.appendChild(optionName);
          }

          choiceGroup.appendChild(optionDiv);
        });
      }

      equipmentChoices.appendChild(choiceGroup);
    });

    // Show continue button when all choices made
    checkEquipmentComplete();
  } else {
    // No choices, show continue immediately
    continueBtn.style.display = 'block';
  }
}

/**
 * Handle complex weapon choices (martial/simple weapon selection)
 */
function renderComplexWeaponChoice(container, choice, choiceIndex) {
  const weaponData = {};

  choice.slots.forEach((slot, slotIndex) => {
    const slotDiv = document.createElement('div');
    slotDiv.className = 'mb-4 p-3 bg-gray-800 rounded';

    const slotTitle = document.createElement('div');
    slotTitle.className = 'font-medium text-green-400 mb-2';
    slotTitle.textContent = `Slot ${slotIndex + 1}`;
    slotDiv.appendChild(slotTitle);

    if (slot.type === 'choice') {
      const select = document.createElement('select');
      select.className = 'bg-gray-700 text-white rounded px-3 py-2 w-full';
      select.id = `weapon-slot-${choiceIndex}-${slotIndex}`;

      const defaultOption = document.createElement('option');
      defaultOption.value = '';
      defaultOption.textContent = '-- Select Weapon --';
      select.appendChild(defaultOption);

      slot.options.forEach(weapon => {
        const option = document.createElement('option');
        option.value = weapon;
        option.textContent = weapon;
        select.appendChild(option);
      });

      slotDiv.appendChild(select);
      weaponData[slotIndex] = { type: 'choice', element: select };
    } else if (slot.type === 'fixed_item') {
      const fixedText = document.createElement('div');
      fixedText.className = 'text-gray-300';
      fixedText.textContent = slot.item;
      slotDiv.appendChild(fixedText);
      weaponData[slotIndex] = { type: 'fixed', item: slot.item };
    }

    container.appendChild(slotDiv);
  });

  // Store complex choice data
  selectedEquipment[choiceIndex] = {
    isComplexChoice: true,
    slots: weaponData,
    getSelectedWeapons: function() {
      const weapons = [];
      Object.values(this.slots).forEach(slot => {
        if (slot.type === 'choice' && slot.element.value) {
          weapons.push([slot.element.value, 1]);
        } else if (slot.type === 'fixed') {
          weapons.push([slot.item, 1]);
        }
      });
      return weapons;
    }
  };
}

/**
 * Select equipment option
 */
function selectEquipment(choiceIndex, option) {
  if (option.bundle) {
    selectedEquipment[choiceIndex] = {
      name: option.name,
      bundle: option.bundle,
      isBundle: true
    };
  } else {
    selectedEquipment[choiceIndex] = {
      item: option.item,
      quantity: option.quantity,
      name: option.name
    };
  }

  // Visual feedback
  const choiceGroups = document.querySelectorAll('#equipment-choices > div');
  if (choiceGroups[choiceIndex]) {
    const options = choiceGroups[choiceIndex].querySelectorAll('.bg-gray-600, .bg-gray-500, .bg-green-700');
    options.forEach(opt => {
      opt.classList.remove('bg-green-700');
      opt.classList.add('bg-gray-600');
    });

    event.currentTarget.classList.remove('bg-gray-600');
    event.currentTarget.classList.add('bg-green-700');
  }

  checkEquipmentComplete();
}

/**
 * Check if all equipment choices are complete
 */
function checkEquipmentComplete() {
  const continueBtn = document.getElementById('continue-adventure-btn');
  const totalChoices = startingEquipment.choices ? startingEquipment.choices.length : 0;
  const selectedCount = Object.keys(selectedEquipment).length;

  if (selectedCount >= totalChoices) {
    continueBtn.style.display = 'block';
  }
}

// ============================================================================
// SAVE AND START ADVENTURE
// ============================================================================

/**
 * Create final inventory from all items with stacking
 */
async function addItemWithStacking(inventory, itemName, quantity) {
  // Load item database
  const itemDbResponse = await fetch('/data/systems/item-database.json');
  const itemDatabase = await itemDbResponse.json();
  const itemData = itemDatabase.items[itemName];

  if (!itemData) {
    console.warn(`Item not found in database: ${itemName}`);
    return;
  }

  // Find existing stack or add new
  const existingItem = inventory.find(inv => inv.item === itemName);
  if (existingItem) {
    existingItem.quantity += quantity;
  } else {
    inventory.push({
      item: itemName,
      quantity: quantity,
      stackable: itemData.stackable || false
    });
  }
}

/**
 * Create proper inventory structure with equipment slots
 */
async function createInventoryFromItems(allItems) {
  const inventoryResponse = await fetch('/data/systems/inventory-system.json');
  const inventorySystem = await inventoryResponse.json();

  const inventory = {
    equipped: {},
    backpack: {
      slots: inventorySystem.containers.leather_backpack.capacity,
      weight_capacity: inventorySystem.containers.leather_backpack.weight_capacity,
      items: []
    }
  };

  // Load item database for slot info
  const itemDbResponse = await fetch('/data/systems/item-database.json');
  const itemDatabase = await itemDbResponse.json();

  // Auto-equip items that can be equipped
  for (const inventoryItem of allItems) {
    const itemData = itemDatabase.items[inventoryItem.item];

    if (itemData && itemData.slot) {
      // Try to equip
      const slot = itemData.slot;
      if (!inventory.equipped[slot]) {
        inventory.equipped[slot] = {
          item: inventoryItem.item,
          quantity: 1
        };
        inventoryItem.quantity -= 1;
      }
    }

    // Put remaining in backpack
    if (inventoryItem.quantity > 0) {
      inventory.backpack.items.push({
        item: inventoryItem.item,
        quantity: inventoryItem.quantity
      });
    }
  }

  return inventory;
}

/**
 * Start the adventure - save character and redirect
 */
async function startAdventure() {
  try {
    console.log('🎮 Starting adventure...');

    if (!generatedCharacter || !startingEquipment) {
      throw new Error('Missing character or equipment data');
    }

    const finalName = playerName || generatedCharacter.name;

    // Collect all items
    let allItems = [];

    // Add starting equipment
    for (const startingItem of startingEquipment.inventory) {
      await addItemWithStacking(allItems, startingItem.item, startingItem.quantity);
    }

    // Add selected equipment
    for (const option of Object.values(selectedEquipment)) {
      if (option.isComplexChoice && option.weapons) {
        // Complex choice with weapons array
        for (const weapon of option.weapons) {
          await addItemWithStacking(allItems, weapon[0], weapon[1]);
        }
      } else if (option.bundle) {
        // Bundle option
        for (const bundleItem of option.bundle) {
          await addItemWithStacking(allItems, bundleItem[0], bundleItem[1]);
        }
      } else if (option.item) {
        // Simple item option
        await addItemWithStacking(allItems, option.item, option.quantity || 1);
      }
    }

    // Create inventory structure
    const inventory = await createInventoryFromItems(allItems);

    // Final character data
    const finalCharacter = {
      name: finalName,
      race: generatedCharacter.race,
      class: generatedCharacter.class,
      background: generatedCharacter.background,
      alignment: generatedCharacter.alignment,
      level: generatedCharacter.level,
      experience: generatedCharacter.experience,
      stats: generatedCharacter.stats,
      hp: generatedCharacter.hp,
      max_hp: generatedCharacter.max_hp,
      mana: generatedCharacter.mana,
      max_mana: generatedCharacter.max_mana,
      fatigue: generatedCharacter.fatigue,
      gold: generatedCharacter.gold,
      inventory: inventory,
      spells: generatedCharacter.spells || {}
    };

    // Create save file
    const session = window.sessionManager.getSession();
    const saveData = {
      npub: session.npub,
      created_at: new Date().toISOString(),
      last_played: new Date().toISOString(),
      character: finalCharacter,
      gameState: {
        location: {
          current: generatedCharacter.city || 'kingdom',
          discovered: [generatedCharacter.city || 'kingdom'],
          music_tracks_unlocked: ['character-creation', 'kingdom-theme']
        }
      }
    };

    const response = await fetch(`/api/saves/${session.npub}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(saveData)
    });

    if (response.ok) {
      const result = await response.json();
      console.log('✅ Save created:', result);

      // Switch music
      const introMusic = document.getElementById('intro-music');
      const gameMusic = document.getElementById('game-music');
      introMusic.pause();
      gameMusic.volume = 0.3;
      gameMusic.play().catch(e => console.log('Music blocked:', e));

      // Redirect to game
      setTimeout(() => {
        window.location.href = '/game?save=' + result.save_id;
      }, 2000);
    } else {
      throw new Error('Failed to create save file');
    }

  } catch (error) {
    console.error('❌ Failed to start adventure:', error);
    alert('Failed to start adventure: ' + error.message);
  }
}

// ============================================================================
// INITIALIZATION
// ============================================================================

document.addEventListener('DOMContentLoaded', async function() {
  console.log('🎮 Initializing game intro...');

  if (!window.sessionManager) {
    console.error('❌ Session manager not available');
    window.location.href = '/';
    return;
  }

  try {
    await window.sessionManager.init();

    if (!window.sessionManager.isAuthenticated()) {
      console.log('❌ Not authenticated, redirecting');
      window.location.href = '/';
      return;
    }

    const session = window.sessionManager.getSession();
    console.log('✅ Authenticated:', session.npub);

    // Initialize character generator
    console.log('🎲 Initializing character generator...');
    await window.characterGenerator.initialize();

    // Generate character
    console.log('🎲 Generating character...');
    generatedCharacter = await window.characterGenerator.generateCharacter(session.npub);
    console.log('✅ Character generated:', generatedCharacter);

    // Generate starting equipment
    startingEquipment = window.characterGenerator.generateStartingEquipment(generatedCharacter);
    console.log('✅ Starting equipment loaded:', startingEquipment);

    // Set up music playlist
    setupMusicPlaylist();

    // Start music when name input is interacted with
    const nameInput = document.getElementById('character-name');
    nameInput.addEventListener('focus', () => {
      startMusic();
    }, { once: true });

  } catch (error) {
    console.error('❌ Initialization failed:', error);
    alert('Failed to initialize: ' + error.message);
  }
});
