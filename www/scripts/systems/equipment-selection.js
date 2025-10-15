// ============================================================================
// EQUIPMENT SELECTION SYSTEM
// Scene-based equipment choice system matching character-generator logic
// ============================================================================

let currentChoiceIndex = 0;
let selectedChoices = {};
let equipmentChoices = [];
let itemStatsCache = {}; // Cache for item stats
let selectionHistory = []; // Track history for back button
let currentBackButton = null; // Current back button element
let backButtonCallback = null; // Callback to go back
let itemsDatabaseCache = null; // Cache for items from database

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
    itemsDatabaseCache = data.items || [];
    return itemsDatabaseCache;
  } catch (error) {
    console.error('Error loading items from database:', error);
    return [];
  }
}

/**
 * Get item data from database cache by ID
 */
async function getItemById(itemId) {
  try {
    const items = await loadItemsFromDatabase();
    const item = items.find(i => i.id === itemId);
    return item || null;
  } catch (error) {
    console.error(`Error getting item ${itemId}:`, error);
    return null;
  }
}

/**
 * Start equipment selection flow (excluding pack)
 */
async function startEquipmentSelection(equipment) {
  equipmentChoices = equipment.choices || [];
  currentChoiceIndex = 0;
  selectedChoices = {};
  selectionHistory = []; // Clear history on new start

  if (equipmentChoices.length === 0) {
    return;
  }

  // Show each choice as a separate scene (but stop before pack)
  while (currentChoiceIndex < equipmentChoices.length) {
    const choice = equipmentChoices[currentChoiceIndex];
    const shouldGoBack = await showEquipmentChoiceScene(choice, currentChoiceIndex);

    if (shouldGoBack) {
      // Go back to previous choice
      currentChoiceIndex--;
      if (currentChoiceIndex >= 0) {
        delete selectedChoices[currentChoiceIndex];
      }
    } else {
      // Continue to next choice
      currentChoiceIndex++;
    }
  }

  // Show confirmation after all equipment choices are made
  const confirmed = await showFinalConfirmation();
  if (!confirmed) {
    // Go back to last choice
    currentChoiceIndex = equipmentChoices.length - 1;
    delete selectedChoices[currentChoiceIndex];
    return startEquipmentSelection({choices: equipmentChoices});
  }
}

/**
 * Show final confirmation before proceeding to given items
 */
async function showFinalConfirmation() {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  // Set dark background for confirmation screen
  background.style.backgroundImage = 'none';
  background.style.backgroundColor = '#111827';

  content.innerHTML = '';

  // Show confirmation dialog
  const confirmDialog = document.createElement('div');
  confirmDialog.className = 'flex flex-col items-center justify-center gap-6';

  const message = document.createElement('div');
  message.className = 'text-2xl font-bold text-yellow-400 text-center';
  message.textContent = 'Are you sure?';
  confirmDialog.appendChild(message);

  const subMessage = document.createElement('div');
  subMessage.className = 'text-lg text-gray-300 text-center max-w-md';
  subMessage.textContent = 'Once you proceed, you cannot change your equipment choices.';
  confirmDialog.appendChild(subMessage);

  const note = document.createElement('div');
  note.className = 'text-sm text-gray-400 text-center max-w-md italic';
  note.textContent = 'You can always find more equipment in game.';
  confirmDialog.appendChild(note);

  const buttonContainer = document.createElement('div');
  buttonContainer.className = 'flex flex-col gap-3 items-center';

  const confirmBtn = document.createElement('button');
  confirmBtn.className = 'pixel-continue-btn';
  confirmBtn.textContent = 'Yes, Continue';
  buttonContainer.appendChild(confirmBtn);

  const backBtn = document.createElement('button');
  backBtn.className = 'pixel-continue-btn';
  backBtn.style.background = '#6b7280'; // Gray background
  backBtn.textContent = 'Go Back';
  buttonContainer.appendChild(backBtn);

  confirmDialog.appendChild(buttonContainer);
  content.appendChild(confirmDialog);

  // Show container with fade-in (matching showScene pattern)
  container.classList.remove('hidden', 'fade-out');
  container.classList.remove('fade-in');
  void container.offsetHeight; // Force reflow
  container.classList.add('fade-in');

  // Wait for user choice
  const userChoice = await new Promise(resolve => {
    confirmBtn.onclick = () => resolve(true);
    backBtn.onclick = () => resolve(false);
  });

  // Animate text out first before fading container
  const allElements = content.querySelectorAll('*');
  allElements.forEach(el => {
    el.style.transition = 'opacity 0.3s ease-out';
    el.style.opacity = '0';
  });
  await new Promise(resolve => setTimeout(resolve, 300));

  // Clear content
  content.innerHTML = '';

  // Fade out the scene (matching showScene pattern)
  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));

  // Fully reset container for next scene (matching showScene pattern)
  container.classList.remove('fade-in', 'fade-out');
  container.classList.add('hidden');

  // Reset any leftover styles from equipment selection
  container.style.opacity = '';
  background.style.backgroundColor = '';

  return userChoice;
}

/**
 * Handle pack selection (called after scene 6)
 */
async function handlePackSelection(startingEquipment) {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  // Check if pack is a choice
  const packChoice = startingEquipment.pack_choice;

  // Check if pack is in given items
  const packGiven = startingEquipment.inventory?.find(item =>
    item.item && item.item.includes('-pack')
  );

  // Turn off background
  background.style.backgroundImage = 'none';
  background.style.backgroundColor = '#111827';
  content.innerHTML = '';
  content.style.zIndex = '10';

  if (packChoice) {
    // Player chooses pack
    const title = document.createElement('div');
    title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
    title.textContent = 'Choose Your Pack';
    content.appendChild(title);

    const description = document.createElement('div');
    description.className = 'text-lg mb-6';
    description.textContent = packChoice.description || 'Choose your adventuring pack';
    content.appendChild(description);

    // Preload all pack items for display
    const allPackItems = [];
    for (const option of packChoice.options) {
      const packData = await fetchPackData(option.item);
      const contents = packData?.properties?.contents || packData?.contents;
      if (contents) {
        contents.forEach(item => allPackItems.push(item[0]));
      }
    }
    await preloadItemStats(allPackItems);

    // Create scrollable container
    const scrollContainer = document.createElement('div');
    scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
    scrollContainer.style.maxHeight = 'calc(100vh - 280px)';

    // Create options container
    const optionsContainer = document.createElement('div');
    optionsContainer.className = 'flex flex-col gap-6';

    let selectedPackIndex = null;

    // Create pack options (similar to bundles)
    for (let i = 0; i < packChoice.options.length; i++) {
      const option = packChoice.options[i];
      const packName = option.item;

      // Fetch pack data to get contents
      const packData = await fetchPackData(packName);

      // Create outer container for the pack (like bundle)
      const packContainer = document.createElement('div');
      packContainer.className = 'bg-gray-800 rounded-lg';
      packContainer.style.padding = '0.75rem';
      packContainer.style.border = '3px solid #374151';
      packContainer.style.cursor = 'pointer';
      packContainer.style.boxSizing = 'border-box';
      packContainer.style.width = 'fit-content';
      packContainer.style.margin = '0 auto';
      packContainer.setAttribute('data-option-index', i);

      // Pack name as title
      const packTitle = document.createElement('div');
      packTitle.className = 'text-center text-yellow-400 font-bold text-xl mb-3';
      packTitle.textContent = packName.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
      packContainer.appendChild(packTitle);

      // Create inner flex container for pack items (like bundle items)
      const contents = packData?.properties?.contents || packData?.contents;
      if (contents) {
        const itemsRow = document.createElement('div');
        itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

        // Add each item in the pack as a card
        for (const packItem of contents) {
          const itemName = packItem[0];
          const quantity = packItem[1];
          const itemCard = await createSimpleItemCard(itemName, quantity, true);

          // Prevent clicks on individual cards from bubbling
          itemCard.addEventListener('click', (e) => {
            if (!e.target.closest('.info-btn')) {
              e.stopPropagation();
              packContainer.click();
            }
          });

          itemsRow.appendChild(itemCard);
        }

        packContainer.appendChild(itemsRow);
      }

      packContainer.onclick = (e) => {
        if (e.target.closest('.info-btn')) return;

        // Deselect all
        optionsContainer.querySelectorAll('[data-option-index]').forEach(opt => {
          opt.classList.remove('selected');
        });
        // Select this one
        packContainer.classList.add('selected');
        selectedPackIndex = i;
        confirmBtn.disabled = false;
        confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      };

      optionsContainer.appendChild(packContainer);
    }

    scrollContainer.appendChild(optionsContainer);
    content.appendChild(scrollContainer);

    // Confirm button
    const confirmBtn = document.createElement('button');
    confirmBtn.className = 'pixel-continue-btn';
    confirmBtn.textContent = 'Confirm Selection';
    confirmBtn.disabled = true;
    confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
    content.appendChild(confirmBtn);

    // Show container with fade-in (matching showScene pattern)
    container.classList.remove('hidden', 'fade-out');
    container.classList.remove('fade-in');
    void container.offsetHeight;
    container.classList.add('fade-in');

    // Wait for selection
    await new Promise(resolve => {
      confirmBtn.onclick = () => {
        if (selectedPackIndex !== null) {
          selectedChoices['pack'] = packChoice.options[selectedPackIndex].item;
          resolve();
        }
      };
    });

  } else if (packGiven) {
    // Show given pack
    const title = document.createElement('div');
    title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
    title.textContent = 'Your Pack';
    content.appendChild(title);

    const description = document.createElement('div');
    description.className = 'text-lg mb-4 text-center';
    description.textContent = 'You have been provided with this pack:';
    content.appendChild(description);

    const packName = packGiven.item;

    // Fetch pack data to get contents
    const packData = await fetchPackData(packName);

    // Preload pack items
    const contents = packData?.properties?.contents || packData?.contents;
    if (contents) {
      await preloadItemStats(contents.map(item => item[0]));
    }

    // Create pack container (like bundle)
    const packContainer = document.createElement('div');
    packContainer.className = 'bg-gray-800 rounded-lg';
    packContainer.style.padding = '0.75rem';
    packContainer.style.border = '3px solid #10b981'; // Green border for given
    packContainer.style.boxSizing = 'border-box';
    packContainer.style.width = 'fit-content';
    packContainer.style.margin = '0 auto';

    // Pack name as title
    const packTitle = document.createElement('div');
    packTitle.className = 'text-center text-green-400 font-bold text-xl mb-3';
    packTitle.textContent = packName.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    packContainer.appendChild(packTitle);

    // Create inner flex container for pack items (reuse contents from above)
    if (contents) {
      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

      // Add each item in the pack as a card
      for (const packItem of contents) {
        const itemName = packItem[0];
        const quantity = packItem[1];
        const itemCard = await createSimpleItemCard(itemName, quantity, true);
        itemsRow.appendChild(itemCard);
      }

      packContainer.appendChild(itemsRow);
    }

    content.appendChild(packContainer);

    selectedChoices['pack'] = packName;

    // Continue button
    const continueBtn = document.createElement('button');
    continueBtn.className = 'pixel-continue-btn mt-6';
    continueBtn.textContent = 'Continue';
    content.appendChild(continueBtn);

    // Show container with fade-in (matching showScene pattern)
    container.classList.remove('hidden', 'fade-out');
    container.classList.remove('fade-in');
    void container.offsetHeight;
    container.classList.add('fade-in');

    await new Promise(resolve => {
      continueBtn.onclick = resolve;
    });
  }

  // Clear content
  content.innerHTML = '';

  // Fade out the scene (matching showScene pattern)
  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));

  // Fully reset container for next scene (matching showScene pattern)
  container.classList.remove('fade-in', 'fade-out');
  container.classList.add('hidden');
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
  title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
  title.textContent = `Choose Your Equipment (${choiceIndex + 1} of ${equipmentChoices.length})`;
  content.appendChild(title);

  // Check if any option is a complex choice (multi_slot)
  const complexOptions = choice.options.filter(opt => opt.isComplexChoice);

  if (complexOptions.length > 0) {
    // Show complex choice with weapon slots (two-step process)
    return await showMultiSlotChoiceSelection(content, choice, choiceIndex, complexOptions);
  } else {
    // Show regular choice
    return await showRegularChoiceSelection(content, choice, choiceIndex);
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
  let userClickedBack = false;

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
      });

      // Select this container - let CSS handle the styling
      container.classList.add('selected');
      console.log('  Applied selected class to container');

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
    for (let idx = 0; idx < bundles.length; idx++) {
      const option = bundles[idx];
      const optionIndex = choice.options.indexOf(option);
      console.log('  Bundle', idx, ':', option);

      // Create outer container for the bundle (this is what gets clicked and highlighted)
      const bundleContainer = document.createElement('div');
      bundleContainer.className = 'bg-gray-800 rounded-lg';
      bundleContainer.style.padding = '0.75rem'; // Match gap between items (gap-3 = 0.75rem)
      bundleContainer.style.border = '3px solid #374151'; // Start with 3px border
      bundleContainer.style.cursor = 'pointer';
      bundleContainer.style.boxSizing = 'border-box';
      bundleContainer.style.width = 'fit-content'; // Only as wide as content
      bundleContainer.style.margin = '0 auto'; // Center it

      // Add bundle label
      const bundleLabel = document.createElement('div');
      bundleLabel.className = 'text-center text-gray-400 font-semibold text-xs mb-3';
      bundleLabel.textContent = '📦 BUNDLE';
      bundleContainer.appendChild(bundleLabel);

      // Create inner flex container for bundle items
      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

      // Add each item in the bundle as a card
      for (const bundleItem of option.bundle) {
        const itemName = bundleItem[0];
        const quantity = bundleItem[1];
        const itemCard = await createSimpleItemCard(itemName, quantity, true);

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
      }

      bundleContainer.appendChild(itemsRow);
      createClickableOption(option, optionIndex, bundleContainer);
      optionsContainer.appendChild(bundleContainer);
    }

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
    gridContainer.className = 'flex flex-wrap justify-center gap-3';

    // Use actual grid only if there are many items
    if (simpleItems.length > 4) {
      gridContainer.className = 'grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3';
    }

    for (let idx = 0; idx < simpleItems.length; idx++) {
      const option = simpleItems[idx];
      const optionIndex = choice.options.indexOf(option);
      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-2 flex-wrap p-2 bg-gray-800 rounded-lg';
      itemsRow.style.border = '3px solid #374151'; // Start with 3px border
      itemsRow.style.cursor = 'pointer';
      itemsRow.style.boxSizing = 'border-box';
      itemsRow.style.width = 'fit-content'; // Only as wide as content

      const itemCard = await createSimpleItemCard(option.item, option.quantity, false);

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
    }

    optionsContainer.appendChild(gridContainer);
  }

  scrollContainer.appendChild(optionsContainer);
  content.appendChild(scrollContainer);

  // Confirm button
  const confirmBtn = createContinueButton(0, 'Confirm Choice');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  // Add back button if this is not the first choice
  if (choiceIndex > 0) {
    const backBtn = createBackButton(() => {
      userClickedBack = true;
      // Enable confirm button so it can be clicked
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      confirmBtn.click(); // Trigger the waiting promise to resolve
    });
    document.body.appendChild(backBtn);
    currentBackButton = backBtn;
  }

  // Ensure container is visible but content is ready before animating
  container.classList.remove('hidden');
  container.style.opacity = '1';
  // Wait a frame to ensure DOM is ready
  await new Promise(resolve => requestAnimationFrame(resolve));
  content.style.animation = 'slideInFromRight 0.3s ease-out';

  // Wait for confirmation
  await waitForButtonClick(confirmBtn);

  // Remove back button
  if (currentBackButton) {
    currentBackButton.remove();
    currentBackButton = null;
  }

  // Animate out with wipe
  if (userClickedBack) {
    content.style.animation = 'wipeLeft 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    // Ensure container is fully hidden before returning
    container.style.opacity = '0';
    container.classList.add('hidden');
    return true; // Signal to go back
  } else {
    content.style.animation = 'wipeRight 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    // DON'T hide the container - let the next scene manage the fade in/out
    // Just reset the content and animation
    content.style.animation = '';
  }

  // Store selection
  selectedChoices[choiceIndex] = selectedOption;

  return false; // Continue forward
}

/**
 * Show multi-slot choice selection (two-step process)
 * Step 1: Choose configuration (weapon+shield OR 2 weapons)
 * Step 2: Choose specific weapons for each slot
 */
async function showMultiSlotChoiceSelection(content, choice, choiceIndex, complexOptions) {
  const container = document.getElementById('scene-container');

  // Step 1: Show configuration options
  const scrollContainer = document.createElement('div');
  scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
  scrollContainer.style.maxHeight = 'calc(100vh - 280px)';

  const optionsContainer = document.createElement('div');
  optionsContainer.className = 'flex flex-col gap-6';

  let selectedConfiguration = null;
  let selectedConfigIndex = null;
  let userClickedBack = false;

  // Render each multi-slot configuration
  complexOptions.forEach((option, idx) => {
    const optionIndex = choice.options.indexOf(option);

    // Create container for this configuration
    const configContainer = document.createElement('div');
    configContainer.className = 'p-6 bg-gray-800 rounded-lg';
    configContainer.style.border = '3px solid #374151'; // Start with 3px border
    configContainer.style.cursor = 'pointer';
    configContainer.style.boxSizing = 'border-box';
    configContainer.dataset.optionIndex = optionIndex;

    // Configuration title
    const configTitle = document.createElement('div');
    configTitle.className = 'text-center text-yellow-400 font-bold text-lg mb-4';
    configTitle.textContent = `Option ${idx + 1}`;
    configContainer.appendChild(configTitle);

    // Show slot descriptions
    const slotsDesc = document.createElement('div');
    slotsDesc.className = 'text-center text-gray-300 space-y-2';

    option.weaponSlots.forEach((slot, slotIdx) => {
      const slotDiv = document.createElement('div');
      if (slot.type === 'weapon_choice') {
        slotDiv.innerHTML = `<span class="text-green-400">⚔️ Choose a weapon</span> from ${slot.options.length} options`;
      } else if (slot.type === 'fixed_item') {
        slotDiv.innerHTML = `<span class="text-blue-400">🛡️ ${slot.item[0]}</span> (included)`;
      }
      slotsDesc.appendChild(slotDiv);
    });

    configContainer.appendChild(slotsDesc);

    // Click handler for configuration selection
    configContainer.onclick = () => {
      // Deselect all
      optionsContainer.querySelectorAll('[data-option-index]').forEach(row => {
        row.classList.remove('selected');
      });

      // Select this one
      configContainer.classList.add('selected');
      selectedConfiguration = option;
      selectedConfigIndex = optionIndex;
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    };

    optionsContainer.appendChild(configContainer);

    // Add OR separator
    if (idx < complexOptions.length - 1) {
      const separator = document.createElement('div');
      separator.className = 'text-center text-yellow-400 font-bold text-xl my-2';
      separator.textContent = '— OR —';
      optionsContainer.appendChild(separator);
    }
  });

  scrollContainer.appendChild(optionsContainer);
  content.appendChild(scrollContainer);

  // Confirm button
  const confirmBtn = createContinueButton(0, 'Confirm Configuration');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  // Add back button if this is not the first choice
  if (choiceIndex > 0) {
    const backBtn = createBackButton(() => {
      userClickedBack = true;
      // Enable confirm button so it can be clicked
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      confirmBtn.click();
    });
    document.body.appendChild(backBtn);
    currentBackButton = backBtn;
  }

  // Show scene with slide in animation
  container.classList.remove('hidden');
  container.style.opacity = '1';
  content.style.animation = 'slideInFromRight 0.3s ease-out';

  // Wait for configuration selection
  await waitForButtonClick(confirmBtn);

  // Remove back button
  if (currentBackButton) {
    currentBackButton.remove();
    currentBackButton = null;
  }

  // Animate out with wipe
  if (userClickedBack) {
    content.style.animation = 'wipeLeft 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    // Ensure container is fully hidden before returning
    container.style.opacity = '0';
    container.classList.add('hidden');
    return true; // Signal to go back
  } else {
    content.style.animation = 'wipeRight 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    // DON'T hide the container - let the next scene manage the fade in/out
    // Just reset the content and animation
    content.style.animation = '';
  }

  // Step 2: Now show weapon selection for each slot in the chosen configuration
  const weaponSelections = {};
  let currentSlotIdx = 0;
  let weaponChoiceIndices = []; // Track which slots are weapon choices (for back navigation)

  // Build list of weapon choice indices
  selectedConfiguration.weaponSlots.forEach((slot, idx) => {
    if (slot.type === 'weapon_choice') {
      weaponChoiceIndices.push(idx);
    }
  });

  let weaponChoicePosition = 0; // Position in weaponChoiceIndices array

  while (weaponChoicePosition < weaponChoiceIndices.length) {
    currentSlotIdx = weaponChoiceIndices[weaponChoicePosition];
    const slot = selectedConfiguration.weaponSlots[currentSlotIdx];

    // Show weapon selection scene
    const result = await showWeaponSlotSelection(slot, weaponChoicePosition, weaponChoiceIndices.length, choiceIndex);

    if (result.shouldGoBack) {
      // Go back to previous weapon choice, or back to configuration if first
      if (weaponChoicePosition > 0) {
        weaponChoicePosition--;
        const prevSlotIdx = weaponChoiceIndices[weaponChoicePosition];
        delete weaponSelections[prevSlotIdx];
      } else {
        // Go back to configuration selection - restart the whole multi-slot process
        return await showMultiSlotChoiceSelection(content, choice, choiceIndex, complexOptions);
      }
    } else {
      weaponSelections[currentSlotIdx] = result.selectedWeapon;
      weaponChoicePosition++;
    }
  }

  // Fill in fixed items
  selectedConfiguration.weaponSlots.forEach((slot, idx) => {
    if (slot.type === 'fixed_item') {
      weaponSelections[idx] = slot.item;
    }
  });

  // Build final selection
  const finalSelection = {
    isComplexChoice: true,
    weapons: Object.values(weaponSelections)
  };

  // Store selection
  selectedChoices[choiceIndex] = finalSelection;
  return false; // Continue forward
}

/**
 * Show weapon selection for a single slot
 */
async function showWeaponSlotSelection(slot, slotIndex, totalSlots, choiceIndex) {
  const container = document.getElementById('scene-container');
  const content = document.getElementById('scene-content');

  content.innerHTML = '';

  // Title
  const title = document.createElement('div');
  title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
  title.textContent = `Choose Weapon ${slotIndex + 1} of ${totalSlots}`;
  content.appendChild(title);

  // Preload weapon stats
  await preloadItemStats(slot.options.map(w => w[0]));

  // Scrollable container
  const scrollContainer = document.createElement('div');
  scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
  scrollContainer.style.maxHeight = 'calc(100vh - 280px)';

  // Grid for weapons
  const gridContainer = document.createElement('div');
  gridContainer.className = 'grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3';

  let selectedWeapon = null;
  let userClickedBack = false;

  for (let idx = 0; idx < slot.options.length; idx++) {
    const weaponOption = slot.options[idx];
    const weaponName = weaponOption[0];
    const quantity = weaponOption[1];

    const itemContainer = document.createElement('div');
    itemContainer.className = 'flex flex-row justify-center gap-2 flex-wrap p-2 bg-gray-800 rounded-lg';
    itemContainer.style.border = '3px solid #374151'; // Start with 3px border
    itemContainer.style.cursor = 'pointer';
    itemContainer.style.boxSizing = 'border-box';
    itemContainer.dataset.optionIndex = idx;

    const itemCard = await createSimpleItemCard(weaponName, quantity, false);

    // Prevent card clicks, redirect to container
    itemCard.addEventListener('click', (e) => {
      if (!e.target.closest('.info-btn')) {
        e.stopPropagation();
        itemContainer.click();
      }
    });

    itemContainer.onclick = (e) => {
      if (e.target.closest('.info-btn')) return;

      // Deselect all
      gridContainer.querySelectorAll('[data-option-index]').forEach(row => {
        row.classList.remove('selected');
      });

      // Select this one
      itemContainer.classList.add('selected');
      selectedWeapon = weaponOption;
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    };

    itemContainer.appendChild(itemCard);
    gridContainer.appendChild(itemContainer);
  }

  scrollContainer.appendChild(gridContainer);
  content.appendChild(scrollContainer);

  // Confirm button
  const confirmBtn = createContinueButton(0, 'Confirm Weapon');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  // Add back button (always show during weapon selection, since we can go back to config or previous weapon)
  const backBtn = createBackButton(() => {
    userClickedBack = true;
    // Enable confirm button so it can be clicked
    confirmBtn.disabled = false;
    confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    confirmBtn.click();
  });
  document.body.appendChild(backBtn);
  currentBackButton = backBtn;

  // Show scene with slide in animation
  container.classList.remove('hidden');
  container.style.opacity = '1';
  content.style.animation = 'slideInFromRight 0.3s ease-out';

  // Wait for weapon selection
  await waitForButtonClick(confirmBtn);

  // Remove back button
  if (currentBackButton) {
    currentBackButton.remove();
    currentBackButton = null;
  }

  // Animate out with wipe
  if (userClickedBack) {
    content.style.animation = 'wipeLeft 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    // Ensure container is fully hidden before returning
    container.style.opacity = '0';
    container.classList.add('hidden');
  } else {
    content.style.animation = 'wipeRight 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    // DON'T hide the container - let the next scene manage the fade in/out
    // Just reset the content and animation
    content.style.animation = '';
  }

  return {
    selectedWeapon: selectedWeapon,
    shouldGoBack: userClickedBack
  };
}

/**
 * DEPRECATED: Old complex choice selection (kept for reference)
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
  for (let slotIndex = 0; slotIndex < complexOption.weaponSlots.length; slotIndex++) {
    const slot = complexOption.weaponSlots[slotIndex];
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
      const itemCard = await createSimpleItemCard(slot.item[0], slot.item[1]);
      itemCard.classList.add('selected');
      itemCard.style.pointerEvents = 'none';
      itemCard.style.opacity = '0.7';
      itemsRow.appendChild(itemCard);

      slotSelections[slotIndex] = slot.item;
    } else if (slot.type === 'weapon_choice') {
      // Show each weapon option as a card
      for (const weaponOption of slot.options) {
        const weaponName = weaponOption[0];
        const quantity = weaponOption[1];
        const itemCard = await createSimpleItemCard(weaponName, quantity);

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
      }
    }

    slotDiv.appendChild(itemsRow);
    slotsContainer.appendChild(slotDiv);
  }

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
async function createEquipmentCard(option) {
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
    card.appendChild(await createSimpleItemContent(option));
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
 * Create a simple item card (for use in weapon slots and simple items)
 */
async function createSimpleItemCard(itemName, quantity, isInBundle = false) {
  const card = document.createElement('div');
  card.className = 'item-card bg-gray-800 rounded-lg relative overflow-hidden';
  card.style.width = '110px';
  card.style.height = '110px';
  card.style.aspectRatio = '1/1';

  // Add border for bundle items to differentiate them
  if (isInBundle) {
    card.style.border = '2px solid #4b5563'; // gray border for bundle items
  }

  // Only add cursor pointer if not in bundle (bundle items aren't individually clickable)
  if (!isInBundle) {
    card.style.cursor = 'pointer';
  }

  // Fetch item data from database
  const itemData = await getItemById(itemName);

  // Item image (fills 80% of container)
  const img = document.createElement('img');
  img.src = itemData?.image || `/res/img/items/${itemName}.png`;
  img.alt = itemName;
  img.className = 'absolute inset-0 w-full h-full object-contain p-3';
  img.style.imageRendering = 'pixelated';
  img.style.imageRendering = '-moz-crisp-edges';
  img.style.imageRendering = 'crisp-edges';
  card.appendChild(img);

  // Rarity dot (top right, larger)
  const rarityDot = document.createElement('div');
  rarityDot.className = 'absolute top-1.5 right-1.5 z-10';
  rarityDot.style.width = '10px';
  rarityDot.style.height = '10px';
  rarityDot.style.borderRadius = '50%';
  rarityDot.style.backgroundColor = getRarityColor(itemName);
  rarityDot.style.border = '1px solid rgba(0,0,0,0.3)';
  card.appendChild(rarityDot);

  // Quantity text (top left, purple text only, larger)
  if (quantity && quantity > 1) {
    const qtyText = document.createElement('div');
    qtyText.className = 'absolute top-1 left-1.5 z-10 font-bold';
    qtyText.style.color = '#a855f7'; // purple
    qtyText.style.fontSize = '0.85rem';
    qtyText.style.textShadow = '0 0 3px rgba(0,0,0,0.8), 1px 1px 2px rgba(0,0,0,0.9)';
    qtyText.textContent = `×${quantity}`;
    card.appendChild(qtyText);
  }

  // Item name overlay at bottom (will be populated with actual name from item data)
  const name = document.createElement('div');
  name.className = 'item-name absolute bottom-0 left-0 right-0 text-center text-white font-semibold px-1 py-1 z-10';
  name.style.fontSize = '0.65rem';
  name.style.lineHeight = '0.75rem';
  name.style.backgroundColor = 'rgba(0,0,0,0.6)';
  name.style.textShadow = '0 1px 2px rgba(0,0,0,0.8)';
  name.dataset.itemId = itemName; // Store item ID for lookup
  name.textContent = itemName; // Temporary - will be replaced by actual name
  card.appendChild(name);

  // Fetch and update with actual item name
  getItemStats(itemName).then(() => {
    const cachedData = itemStatsCache[itemName];
    if (cachedData) {
      const nameMatch = cachedData.match(/<div class="font-bold text-yellow-400[^>]*>([^<]+)<\/div>/);
      if (nameMatch && nameMatch[1]) {
        name.textContent = nameMatch[1];
      }
    }
  });

  // Info button (bottom right, yellow ?, larger, hover effect)
  const infoBtn = document.createElement('button');
  infoBtn.className = 'info-btn absolute bottom-1 right-1.5 text-yellow-400 font-bold hover:text-green-400 transition-colors z-20';
  infoBtn.textContent = '?';
  infoBtn.style.fontSize = '16px';
  infoBtn.style.textShadow = '0 0 3px rgba(0,0,0,0.8), 1px 1px 2px rgba(0,0,0,0.9)';
  infoBtn.onclick = (e) => {
    e.stopPropagation(); // Don't trigger card/container selection
    showItemModal(itemName);
  };
  card.appendChild(infoBtn);

  return card;
}

/**
 * Create content for simple item
 */
async function createSimpleItemContent(option) {
  const container = document.createElement('div');

  // Fetch item data from database
  const itemData = await getItemById(option.item);

  // Item image
  const img = document.createElement('img');
  img.src = itemData?.image || `/res/img/items/${option.item}.png`;
  img.alt = option.item;
  img.className = 'w-20 h-20 mx-auto mb-2 object-contain';
  img.style.imageRendering = 'pixelated';
  img.style.imageRendering = '-moz-crisp-edges';
  img.style.imageRendering = 'crisp-edges';
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
 * Fetch pack data from API to get contents
 */
async function fetchPackData(packName) {
  try {
    console.log('📦 Fetching pack data for:', packName);
    const response = await fetch(`/api/items?name=${encodeURIComponent(packName)}`);
    if (!response.ok) {
      console.warn('❌ Failed to fetch pack data for:', packName);
      return null;
    }

    const items = await response.json();
    console.log('📊 API returned for', packName, ':', items);

    if (!items || items.length === 0) {
      console.warn('⚠️ No pack data found for:', packName);
      return null;
    }

    const packData = items[0];
    const contents = packData?.properties?.contents || packData?.contents;
    console.log('📋 Pack data structure:', {
      id: packData.id,
      name: packData.name,
      hasContents: !!contents,
      contentsCount: contents ? contents.length : 0,
      contentsLocation: packData.contents ? 'root' : (packData.properties?.contents ? 'properties.contents' : 'not found'),
      fullData: packData
    });

    return packData; // Return the pack item data with contents
  } catch (error) {
    console.error('❌ Error fetching pack data:', packName, error);
    return null;
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
