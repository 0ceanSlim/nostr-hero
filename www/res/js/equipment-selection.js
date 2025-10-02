// ============================================================================
// EQUIPMENT SELECTION SYSTEM
// Scene-based equipment choice system matching character-generator logic
// ============================================================================

let currentChoiceIndex = 0;
let selectedChoices = {};
let equipmentChoices = [];
let itemStatsCache = {}; // Cache for item stats

/**
 * Start equipment selection flow
 */
async function startEquipmentSelection(equipment) {
  equipmentChoices = equipment.choices || [];
  currentChoiceIndex = 0;
  selectedChoices = {};

  if (equipmentChoices.length === 0) {
    return;
  }

  // Show each choice as a separate scene
  for (let i = 0; i < equipmentChoices.length; i++) {
    currentChoiceIndex = i;
    const choice = equipmentChoices[i];
    await showEquipmentChoiceScene(choice, i);
  }
}

/**
 * Show equipment choice scene
 */
async function showEquipmentChoiceScene(choice, choiceIndex) {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  // Turn off background for equipment selection
  background.style.backgroundImage = 'none';
  background.style.backgroundColor = '#111827'; // dark gray
  content.innerHTML = '';
  content.style.zIndex = '10'; // Ensure content is above everything

  // Title
  const title = document.createElement('div');
  title.className = 'scene-text text-xl md:text-2xl font-bold text-yellow-400 mb-4';
  title.textContent = `Choose Your Equipment (${choiceIndex + 1} of ${equipmentChoices.length})`;
  content.appendChild(title);

  // Check if any option is a complex choice
  const hasComplexChoice = choice.options.some(opt => opt.isComplexChoice);

  if (hasComplexChoice) {
    // Show complex choice with weapon slots
    await showComplexChoiceSelection(content, choice, choiceIndex);
  } else {
    // Show regular choice
    await showRegularChoiceSelection(content, choice, choiceIndex);
  }
}

/**
 * Show regular (non-complex) choice selection
 */
async function showRegularChoiceSelection(content, choice, choiceIndex) {
  const container = document.getElementById('scene-container');

  console.log('🎯 showRegularChoiceSelection received choice:', choice);
  console.log('🎯 First 3 options:', choice.options.slice(0, 3));

  // Preload all item stats for this choice
  const allItems = [];
  choice.options.forEach(option => {
    if (option.isBundle) {
      option.bundle.forEach(bundleItem => allItems.push(bundleItem[0]));
    } else if (!option.isComplexChoice) {
      allItems.push(option.item);
    }
  });
  await preloadItemStats(allItems);

  // Scrollable container
  const scrollContainer = document.createElement('div');
  scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
  scrollContainer.style.maxHeight = 'calc(100vh - 280px)'; // Fit between title and button

  // Container for all choice options
  const optionsContainer = document.createElement('div');

  let selectedOption = null;
  let selectedOptionIndex = null;

  // Separate bundles from simple items
  const bundles = choice.options.filter(opt => opt.isBundle);
  const simpleItems = choice.options.filter(opt => !opt.isBundle && !opt.isComplexChoice);
  const complexItems = choice.options.filter(opt => opt.isComplexChoice);

  console.log('📋 Choice analysis:', {
    totalOptions: choice.options.length,
    bundles: bundles.length,
    simpleItems: simpleItems.length,
    complexItems: complexItems.length
  });

  // Use vertical container with sections
  optionsContainer.className = 'flex flex-col gap-6';

  // Helper function to create clickable option
  const createClickableOption = (option, optionIndex, container) => {
    container.dataset.optionIndex = optionIndex;
    container.style.cursor = 'pointer';

    container.onclick = (e) => {
      console.log('🖱️ Click on option', optionIndex, '- target:', e.target);

      // Allow info button clicks to work
      if (e.target.closest('.info-btn')) {
        console.log('  Info button clicked, ignoring');
        return;
      }

      console.log('  Selecting option', optionIndex);

      // Deselect all containers
      const allContainers = optionsContainer.querySelectorAll('[data-option-index]');
      console.log('  Found', allContainers.length, 'option containers to deselect');
      allContainers.forEach(row => {
        row.classList.remove('selected');
        row.style.border = ''; // Clear custom border
        row.style.boxShadow = '';
        row.style.backgroundColor = '';
      });

      // Select this container with strong visual feedback
      container.classList.add('selected');
      container.style.border = '4px solid #10b981'; // Green border
      container.style.boxShadow = '0 0 30px rgba(16, 185, 129, 0.8)'; // Green glow
      container.style.backgroundColor = '#1f2937'; // Darker background
      console.log('  Applied green border and shadow to container');

      selectedOption = option;
      selectedOptionIndex = optionIndex;
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      console.log('  ✅ Selection complete');
    };
  };

  // Render bundles first (if any)
  if (bundles.length > 0) {
    console.log('🎁 Rendering', bundles.length, 'bundles');
    bundles.forEach((option, idx) => {
      const optionIndex = choice.options.indexOf(option);
      console.log('  Bundle', idx, ':', option);

      // Create outer container for the bundle (this is what gets clicked and highlighted)
      const bundleContainer = document.createElement('div');
      bundleContainer.className = 'p-4 bg-gray-800 rounded-lg border-2 border-gray-700';
      bundleContainer.style.cursor = 'pointer';

      // Add bundle label
      const bundleLabel = document.createElement('div');
      bundleLabel.className = 'text-center text-gray-400 font-semibold text-xs mb-3';
      bundleLabel.textContent = '📦 BUNDLE';
      bundleContainer.appendChild(bundleLabel);

      // Create inner flex container for bundle items
      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

      // Add each item in the bundle as a card
      option.bundle.forEach((bundleItem) => {
        const itemName = bundleItem[0];
        const quantity = bundleItem[1];
        const itemCard = createSimpleItemCard(itemName, quantity, true);

        // Prevent clicks on individual cards from bubbling to container
        // but allow info button clicks to work
        itemCard.addEventListener('click', (e) => {
          if (!e.target.closest('.info-btn')) {
            e.stopPropagation();
            // Trigger the container click instead
            bundleContainer.click();
          }
        });

        itemsRow.appendChild(itemCard);
      });

      bundleContainer.appendChild(itemsRow);
      createClickableOption(option, optionIndex, bundleContainer);
      optionsContainer.appendChild(bundleContainer);
    });

    // Add OR separator after bundles if there are simple items
    if (simpleItems.length > 0) {
      const separator = document.createElement('div');
      separator.className = 'text-center text-yellow-400 font-bold text-xl my-2';
      separator.textContent = '— OR —';
      optionsContainer.appendChild(separator);
    }
  }

  // Render simple items in a grid
  if (simpleItems.length > 0) {
    const gridContainer = document.createElement('div');
    gridContainer.className = 'grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3';

    simpleItems.forEach((option, idx) => {
      const optionIndex = choice.options.indexOf(option);
      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-2 flex-wrap p-2 bg-gray-800 rounded-lg border-2 border-gray-700';
      itemsRow.style.cursor = 'pointer';

      const itemCard = createSimpleItemCard(option.item, option.quantity, false);

      // Prevent clicks on card from bubbling, redirect to container
      // but allow info button clicks to work
      itemCard.addEventListener('click', (e) => {
        if (!e.target.closest('.info-btn')) {
          e.stopPropagation();
          // Trigger the container click instead
          itemsRow.click();
        }
      });

      itemsRow.appendChild(itemCard);

      createClickableOption(option, optionIndex, itemsRow);
      gridContainer.appendChild(itemsRow);
    });

    optionsContainer.appendChild(gridContainer);
  }

  scrollContainer.appendChild(optionsContainer);
  content.appendChild(scrollContainer);

  // Confirm button
  const confirmBtn = createContinueButton(0, 'Confirm Choice');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  // Show scene
  container.classList.remove('hidden', 'fade-out');
  container.classList.add('fade-in');

  // Wait for confirmation
  await waitForButtonClick(confirmBtn);

  // Store selection
  selectedChoices[choiceIndex] = selectedOption;

  // Animate out
  await animateSceneOut(content, container);
}

/**
 * Show complex choice selection (weapon slots)
 */
async function showComplexChoiceSelection(content, choice, choiceIndex) {
  const container = document.getElementById('scene-container');

  // Get the complex choice option
  const complexOption = choice.options.find(opt => opt.isComplexChoice);

  // Preload all item stats for this complex choice
  const allItems = [];
  complexOption.weaponSlots.forEach(slot => {
    if (slot.type === 'fixed_item') {
      allItems.push(slot.item[0]);
    } else if (slot.type === 'weapon_choice') {
      slot.options.forEach(weaponOption => allItems.push(weaponOption[0]));
    }
  });
  await preloadItemStats(allItems);

  // Scrollable container
  const scrollContainer = document.createElement('div');
  scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
  scrollContainer.style.maxHeight = 'calc(100vh - 280px)'; // Fit between title and button

  // Container for all weapon slots
  const slotsContainer = document.createElement('div');
  slotsContainer.className = 'flex flex-col gap-6';

  const slotSelections = {};

  // Render each weapon slot
  complexOption.weaponSlots.forEach((slot, slotIndex) => {
    const slotDiv = document.createElement('div');
    slotDiv.className = 'flex flex-col';

    // Slot title
    const slotTitle = document.createElement('div');
    slotTitle.className = 'text-xl font-semibold text-green-400 mb-3 text-center';
    slotTitle.textContent = `Slot ${slotIndex + 1}`;
    slotDiv.appendChild(slotTitle);

    // Items row for this slot
    const itemsRow = document.createElement('div');
    itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

    if (slot.type === 'fixed_item') {
      // Fixed item - show as pre-selected
      const itemCard = createSimpleItemCard(slot.item[0], slot.item[1]);
      itemCard.classList.add('selected');
      itemCard.style.pointerEvents = 'none';
      itemCard.style.opacity = '0.7';
      itemsRow.appendChild(itemCard);

      slotSelections[slotIndex] = slot.item;
    } else if (slot.type === 'weapon_choice') {
      // Show each weapon option as a card
      slot.options.forEach(weaponOption => {
        const weaponName = weaponOption[0];
        const quantity = weaponOption[1];
        const itemCard = createSimpleItemCard(weaponName, quantity);

        itemCard.onclick = () => {
          // Deselect all in this slot
          itemsRow.querySelectorAll('.item-card').forEach(card => {
            card.classList.remove('selected');
          });

          // Select this one
          itemCard.classList.add('selected');
          slotSelections[slotIndex] = weaponOption;

          // Check if all slots are filled
          checkAllSlotsSelected();
        };

        itemsRow.appendChild(itemCard);
      });
    }

    slotDiv.appendChild(itemsRow);
    slotsContainer.appendChild(slotDiv);
  });

  scrollContainer.appendChild(slotsContainer);
  content.appendChild(scrollContainer);

  // Confirm button
  const confirmBtn = createContinueButton(0, 'Confirm Selection');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  function checkAllSlotsSelected() {
    const allSelected = complexOption.weaponSlots.every((_, idx) => slotSelections[idx] !== undefined);
    if (allSelected) {
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    }
  }

  // Show scene
  container.classList.remove('hidden', 'fade-out');
  container.classList.add('fade-in');

  // Wait for confirmation
  await waitForButtonClick(confirmBtn);

  // Build the selected complex choice
  const selectedComplexChoice = {
    isComplexChoice: true,
    weapons: Object.values(slotSelections)
  };

  // Store selection
  selectedChoices[choiceIndex] = selectedComplexChoice;

  // Animate out
  await animateSceneOut(content, container);
}

/**
 * Create equipment card based on option type
 */
function createEquipmentCard(option) {
  const card = document.createElement('div');
  card.className = 'item-card bg-gray-800 rounded-lg p-3 cursor-pointer transition-all hover:scale-105 relative';
  card.style.width = '140px';
  card.style.minHeight = '170px';

  // Determine what to display
  if (option.isComplexChoice) {
    // Complex choice - show description
    card.appendChild(createComplexChoiceContent(option));
  } else if (option.isBundle) {
    // Bundle - show bundle items
    card.appendChild(createBundleContent(option));
  } else {
    // Simple item
    card.appendChild(createSimpleItemContent(option));
  }

  // Tooltip on hover (debounced)
  let tooltipTimeout;
  card.addEventListener('mouseenter', () => {
    clearTimeout(tooltipTimeout);
    tooltipTimeout = setTimeout(() => showItemTooltip(card, option), 300);
  });
  card.addEventListener('mouseleave', () => {
    clearTimeout(tooltipTimeout);
    hideItemTooltip();
  });

  return card;
}

/**
 * Create a simple item card (for use in weapon slots and simple items)
 */
function createSimpleItemCard(itemName, quantity, isInBundle = false) {
  const card = document.createElement('div');
  card.className = 'item-card bg-gray-800 rounded-lg p-2 transition-all hover:scale-105 relative';
  card.style.width = '100px';
  card.style.minHeight = '120px';

  // Only add cursor pointer if not in bundle (bundle items aren't individually clickable)
  if (!isInBundle) {
    card.style.cursor = 'pointer';
  }

  // Info button (question mark)
  const infoBtn = document.createElement('button');
  infoBtn.className = 'info-btn absolute top-1 right-1 w-5 h-5 bg-yellow-400 text-black rounded-full text-xs font-bold flex items-center justify-center hover:bg-yellow-300 transition-colors z-10';
  infoBtn.textContent = '?';
  infoBtn.style.fontSize = '10px';
  infoBtn.onclick = (e) => {
    e.stopPropagation(); // Don't trigger card/container selection
    showItemModal(itemName);
  };
  card.appendChild(infoBtn);

  // Item image
  const img = document.createElement('img');
  img.src = `/res/img/items/${getItemImageName(itemName)}.png`;
  img.alt = itemName;
  img.className = 'w-14 h-14 mx-auto mb-1 object-contain';
  img.onerror = () => {
    img.src = '/res/img/otherstuff.png';
  };
  card.appendChild(img);

  // Item name
  const name = document.createElement('div');
  name.className = 'text-center text-white font-semibold text-xs leading-tight mb-1 px-1';
  name.style.fontSize = '0.65rem';
  name.style.lineHeight = '0.9rem';
  name.textContent = itemName;
  card.appendChild(name);

  // Quantity
  if (quantity && quantity > 1) {
    const qtyDiv = document.createElement('div');
    qtyDiv.className = 'text-center text-gray-400';
    qtyDiv.style.fontSize = '0.6rem';
    qtyDiv.textContent = `x${quantity}`;
    card.appendChild(qtyDiv);
  }

  return card;
}

/**
 * Create content for simple item
 */
function createSimpleItemContent(option) {
  const container = document.createElement('div');

  // Item image
  const img = document.createElement('img');
  img.src = `/res/img/items/${getItemImageName(option.item)}.png`;
  img.alt = option.item;
  img.className = 'w-20 h-20 mx-auto mb-2 object-contain';
  img.onerror = () => {
    img.src = '/res/img/otherstuff.png';
  };
  container.appendChild(img);

  // Item name
  const name = document.createElement('div');
  name.className = 'text-center text-white font-semibold text-xs leading-tight mb-1';
  name.textContent = option.item;
  container.appendChild(name);

  // Quantity
  if (option.quantity > 1) {
    const qtyDiv = document.createElement('div');
    qtyDiv.className = 'text-center text-gray-400 text-xs';
    qtyDiv.textContent = `Quantity: ${option.quantity}`;
    container.appendChild(qtyDiv);
  }

  return container;
}

/**
 * Create content for bundle
 */
function createBundleContent(option) {
  const container = document.createElement('div');

  // Bundle icon
  const icon = document.createElement('div');
  icon.className = 'w-20 h-20 mx-auto mb-2 bg-gray-700 rounded flex items-center justify-center';
  icon.innerHTML = '<span class="text-3xl">📦</span>';
  container.appendChild(icon);

  // Bundle description
  const name = document.createElement('div');
  name.className = 'text-center text-white font-semibold text-xs leading-tight';
  name.textContent = 'Bundle';
  container.appendChild(name);

  // Bundle items list
  const itemsList = document.createElement('div');
  itemsList.className = 'text-center text-gray-400 text-xs mt-1';
  itemsList.textContent = option.bundle.map(item =>
    item[1] > 1 ? `${item[0]} (x${item[1]})` : item[0]
  ).slice(0, 2).join(', ');
  if (option.bundle.length > 2) {
    itemsList.textContent += '...';
  }
  container.appendChild(itemsList);

  return container;
}

/**
 * Create content for complex choice
 */
function createComplexChoiceContent(option) {
  const container = document.createElement('div');

  // Complex choice icon
  const icon = document.createElement('div');
  icon.className = 'w-20 h-20 mx-auto mb-2 bg-gray-700 rounded flex items-center justify-center';
  icon.innerHTML = '<span class="text-3xl">⚔️</span>';
  container.appendChild(icon);

  // Description
  const name = document.createElement('div');
  name.className = 'text-center text-white font-semibold text-xs leading-tight';
  name.textContent = option.item; // "Choose weapon + shield" etc
  container.appendChild(name);

  return container;
}

/**
 * Get item image name
 */
function getItemImageName(itemName) {
  return itemName.toLowerCase().replace(/[',]/g, '').replace(/\s+/g, '-');
}

/**
 * Show item info modal
 */
async function showItemModal(itemName) {
  console.log('🔍 showItemModal called with:', itemName);
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
  console.log('📄 Stats HTML for', itemName, '- Length:', statsHTML.length, '- Content:', statsHTML.substring(0, 100));

  modal.innerHTML = '';
  modal.appendChild(closeBtn);

  const contentDiv = document.createElement('div');
  contentDiv.className = 'text-white';
  contentDiv.innerHTML = statsHTML;
  modal.appendChild(contentDiv);

  console.log('✅ Modal updated with content for:', itemName);
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
 * Preload item stats for multiple items
 */
async function preloadItemStats(itemNames) {
  const uniqueNames = [...new Set(itemNames)];
  console.log('🔄 Preloading item stats for:', uniqueNames);

  // Load sequentially to avoid race conditions
  for (const name of uniqueNames) {
    await getItemStats(name);
  }

  console.log('✅ Preloaded cache:', Object.keys(itemStatsCache));
  // Verify cache contents
  for (const name of uniqueNames) {
    const cached = itemStatsCache[name];
    if (cached) {
      const match = cached.match(/>([^<]+)<\/div>/);
      console.log(`  ${name} -> ${match ? match[1] : 'unknown'}`);
    }
  }
}

/**
 * Get item stats from API
 */
async function getItemStats(itemName) {
  console.log('📊 Getting stats for:', itemName);

  if (itemStatsCache[itemName]) {
    console.log('✨ Using cached data for:', itemName, '- First 50 chars:', itemStatsCache[itemName].substring(0, 50));
    return itemStatsCache[itemName];
  }

  try {
    console.log('🌐 Fetching from API:', itemName);
    const response = await fetch(`/api/items?name=${encodeURIComponent(itemName)}`);
    if (!response.ok) {
      console.warn('❌ API response not ok for:', itemName);
      const fallback = `<div class="font-bold text-yellow-400 mb-1">${itemName}</div><div class="text-gray-400 text-xs">No details available</div>`;
      itemStatsCache[itemName] = fallback;
      return fallback;
    }

    const items = await response.json();
    console.log('📦 API returned for', itemName, ':', items);

    if (!items || items.length === 0) {
      console.warn('⚠️ No items found for:', itemName);
      const fallback = `<div class="font-bold text-yellow-400 mb-1">${itemName}</div><div class="text-gray-400 text-xs">No details available</div>`;
      itemStatsCache[itemName] = fallback;
      return fallback;
    }

    const itemData = items[0];
    console.log('📋 Item data received:', {
      requested: itemName,
      apiReturned: itemData.name,
      id: itemData.id,
      match: itemData.name === itemName,
      fullData: itemData
    });
    const props = itemData.properties || {};
    console.log('📦 Properties object:', props);

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

    console.log('💾 Caching data for:', itemName);
    itemStatsCache[itemName] = statsHTML;
    return statsHTML;

  } catch (error) {
    console.error('❌ Error fetching item:', itemName, error);
    const fallback = `<div class="font-bold text-yellow-400 mb-1">${itemName}</div><div class="text-gray-400 text-xs">Error loading details</div>`;
    itemStatsCache[itemName] = fallback;
    return fallback;
  }
}

/**
 * Animate scene out
 */
async function animateSceneOut(content, container) {
  const textElements = content.querySelectorAll('.scene-text, .item-card, .pixel-continue-btn');
  textElements.forEach(el => {
    el.style.animation = 'wipeOut 0.6s ease-in forwards';
  });

  await new Promise(resolve => setTimeout(resolve, 600));
  await new Promise(resolve => setTimeout(resolve, 1000));

  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));
  container.classList.add('hidden');
}

/**
 * Get all selected equipment
 */
function getSelectedEquipment() {
  return selectedChoices;
}
