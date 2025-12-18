/**
 * Equipment Selection System
 *
 * Handles the scene-based equipment choice flow, including complex weapon
 * selections and pack choices, matching the character-generator logic.
 *
 * @module systems/equipmentSelection
 */

import {
  createContinueButton,
  waitForButtonClick
} from '../components/continueButton.js';
import {
  createBackButton
} from '../components/backButton.js';
import {
  getItemById
} from '../state/staticData.js'; // For basic item data lookup
import {
  getItemStats
} from '../data/items.js'; // For fetching detailed item stats HTML
import {
  logger
} from '../lib/logger.js'; // Assuming logger is available

let currentChoiceIndex = 0;
let selectedChoices = {};
let equipmentChoices = [];
let itemStatsCache = {}; // Cache for item stats
let selectionHistory = []; // Track history for back button
let currentBackButton = null; // Current back button element
let backButtonCallback = null; // Callback to go back

// ============================================================================
// PUBLIC API
// ============================================================================

/**
 * Start equipment selection flow (excluding pack)
 * @param {Object} equipment - The equipment choices provided by character generator.
 * @param {Array<Object>} equipment.choices - Array of choice objects.
 */
export async function startEquipmentSelection(equipment) {
  logger.debug("⚙️ Starting equipment selection with:", equipment);
  equipmentChoices = equipment.choices || [];
  currentChoiceIndex = 0;
  selectedChoices = {};
  selectionHistory = []; // Clear history on new start

  if (equipmentChoices.length === 0) {
    logger.debug("⚙️ No equipment choices, skipping selection.");
    return;
  }

  // Show each choice as a separate scene (but stop before pack)
  while (currentChoiceIndex < equipmentChoices.length) {
    logger.debug(`⚙️ Showing choice ${currentChoiceIndex + 1}/${equipmentChoices.length}`);
    const choice = equipmentChoices[currentChoiceIndex];
    const shouldGoBack = await showEquipmentChoiceScene(choice, currentChoiceIndex);

    if (shouldGoBack) {
      // Go back to previous choice
      currentChoiceIndex--;
      if (currentChoiceIndex >= 0) {
        delete selectedChoices[currentChoiceIndex];
        logger.debug(`⚙️ Went back to choice ${currentChoiceIndex + 1}. Selected choices:`, selectedChoices);
      } else {
        // If already at first choice and going back, essentially cancel selection
        logger.warn("⚙️ Attempted to go back before first choice. Resetting selection.");
        return; // Or handle as a cancellation
      }
    } else {
      // Continue to next choice
      currentChoiceIndex++;
      logger.debug(`⚙️ Proceeded to next choice. Selected choices:`, selectedChoices);
    }
  }

  // Show confirmation after all equipment choices are made
  logger.debug("⚙️ All equipment choices made. Showing final confirmation.");
  const confirmed = await showFinalConfirmation();
  if (!confirmed) {
    logger.debug("⚙️ Final confirmation rejected. Going back to last choice.");
    // Go back to last choice
    currentChoiceIndex = equipmentChoices.length - 1;
    delete selectedChoices[currentChoiceIndex];
    // Restart the selection process (might be better to have a dedicated "back to first choice" mechanism)
    // For now, re-init with the same choices.
    return startEquipmentSelection({
      choices: equipmentChoices
    });
  }
  logger.debug("⚙️ Final confirmation accepted.");
}

/**
 * Handle pack selection (called after other equipment choices)
 * @param {Object} startingEquipment - The initial equipment data from character generator.
 */
export async function handlePackSelection(startingEquipment) {
  logger.debug("📦 Handling pack selection with:", startingEquipment);
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  const packChoice = startingEquipment.pack_choice;
  const packGiven = startingEquipment.inventory?.find(item =>
    item.item && item.item.includes('-pack')
  );

  background.style.backgroundImage = 'none';
  background.style.backgroundColor = '#111827';
  content.innerHTML = '';
  content.style.zIndex = '10';

  if (packChoice) {
    logger.debug("📦 Player needs to choose a pack.");
    const title = document.createElement('div');
    title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
    title.textContent = 'Choose Your Pack';
    content.appendChild(title);

    const description = document.createElement('div');
    description.className = 'text-lg mb-6';
    description.textContent = packChoice.description || 'Choose your adventuring pack';
    content.appendChild(description);

    const allPackItems = [];
    for (const option of packChoice.options) {
      const packData = await fetchPackData(option.item);
      const contents = packData?.properties?.contents || packData?.contents;
      if (contents) {
        contents.forEach(item => allPackItems.push(item[0]));
      }
    }
    await preloadItemStats(allPackItems);

    const scrollContainer = document.createElement('div');
    scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
    scrollContainer.style.maxHeight = 'calc(100vh - 280px)';

    const optionsContainer = document.createElement('div');
    optionsContainer.className = 'flex flex-col gap-6';

    let selectedPackIndex = null;

    for (let i = 0; i < packChoice.options.length; i++) {
      const option = packChoice.options[i];
      const packName = option.item;
      const packData = await fetchPackData(packName);

      const packContainer = document.createElement('div');
      packContainer.className = 'bg-gray-800 rounded-lg';
      packContainer.style.padding = '0.75rem';
      packContainer.style.border = '3px solid #374151';
      packContainer.style.cursor = 'pointer';
      packContainer.style.boxSizing = 'border-box';
      packContainer.style.width = 'fit-content';
      packContainer.style.margin = '0 auto';
      packContainer.setAttribute('data-option-index', i);

      const packTitle = document.createElement('div');
      packTitle.className = 'text-center text-yellow-400 font-bold text-xl mb-3';
      packTitle.textContent = packName.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
      packContainer.appendChild(packTitle);

      const contents = packData?.properties?.contents || packData?.contents;
      if (contents) {
        const itemsRow = document.createElement('div');
        itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

        for (const packItem of contents) {
          const itemName = packItem[0];
          const quantity = packItem[1];
          const itemCard = await createSimpleItemCard(itemName, quantity, true);

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

        optionsContainer.querySelectorAll('[data-option-index]').forEach(opt => {
          opt.classList.remove('selected');
        });
        packContainer.classList.add('selected');
        selectedPackIndex = i;
        confirmBtn.disabled = false;
        confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      };
      optionsContainer.appendChild(packContainer);
    }

    scrollContainer.appendChild(optionsContainer);
    content.appendChild(scrollContainer);

    const confirmBtn = createContinueButton(0, 'Confirm Selection');
    confirmBtn.disabled = true;
    confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
    content.appendChild(confirmBtn);

    container.classList.remove('hidden', 'fade-out');
    container.classList.remove('fade-in');
    void container.offsetHeight;
    container.classList.add('fade-in');

    await new Promise(resolve => {
      confirmBtn.onclick = () => {
        if (selectedPackIndex !== null) {
          selectedChoices['pack'] = packChoice.options[selectedPackIndex].item;
          resolve();
        }
      };
    });

  } else if (packGiven) {
    logger.debug("📦 Player has a pack already provided:", packGiven.item);
    const title = document.createElement('div');
    title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
    title.textContent = 'Your Pack';
    content.appendChild(title);

    const description = document.createElement('div');
    description.className = 'text-lg mb-4 text-center';
    description.textContent = 'You have been provided with this pack:';
    content.appendChild(description);

    const packName = packGiven.item;
    const packData = await fetchPackData(packName);

    const contents = packData?.properties?.contents || packData?.contents;
    if (contents) {
      await preloadItemStats(contents.map(item => item[0]));
    }

    const packContainer = document.createElement('div');
    packContainer.className = 'bg-gray-800 rounded-lg';
    packContainer.style.padding = '0.75rem';
    packContainer.style.border = '3px solid #10b981';
    packContainer.style.boxSizing = 'border-box';
    packContainer.style.width = 'fit-content';
    packContainer.style.margin = '0 auto';

    const packTitle = document.createElement('div');
    packTitle.className = 'text-center text-green-400 font-bold text-xl mb-3';
    packTitle.textContent = packName.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    packContainer.appendChild(packTitle);

    if (contents) {
      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

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

    const continueBtn = createContinueButton(0, 'Continue');
    continueBtn.className += ' mt-6';
    content.appendChild(continueBtn);

    container.classList.remove('hidden', 'fade-out');
    container.classList.remove('fade-in');
    void container.offsetHeight;
    container.classList.add('fade-in');

    await new Promise(resolve => {
      continueBtn.onclick = resolve;
    });
  } else {
    logger.debug("📦 No pack choice and no pack given. Skipping pack selection.");
  }

  content.innerHTML = '';

  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));

  container.classList.remove('fade-in', 'fade-out');
  container.classList.add('hidden');
}

/**
 * Get all selected equipment.
 * @returns {Object} An object mapping choice index/key to selected equipment.
 */
export function getSelectedEquipment() {
  logger.debug("⚙️ Getting selected equipment:", selectedChoices);
  return selectedChoices;
}

// ============================================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================================

/**
 * Show final confirmation before proceeding to given items.
 * @returns {Promise<boolean>} True if confirmed, false if back.
 */
async function showFinalConfirmation() {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  background.style.backgroundImage = 'none';
  background.style.backgroundColor = '#111827';
  content.innerHTML = '';

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

  const confirmBtn = createContinueButton(0, 'Yes, Continue');
  buttonContainer.appendChild(confirmBtn);

  const backBtn = createContinueButton(0, 'Go Back'); // Using createContinueButton for styling consistency
  backBtn.style.background = '#6b7280';
  buttonContainer.appendChild(backBtn);

  confirmDialog.appendChild(buttonContainer);
  content.appendChild(confirmDialog);

  container.classList.remove('hidden', 'fade-out');
  container.classList.remove('fade-in');
  void container.offsetHeight;
  container.classList.add('fade-in');

  const userChoice = await new Promise(resolve => {
    confirmBtn.onclick = () => resolve(true);
    backBtn.onclick = () => resolve(false);
  });

  const allElements = content.querySelectorAll('*');
  allElements.forEach(el => {
    el.style.transition = 'opacity 0.3s ease-out';
    el.style.opacity = '0';
  });
  await new Promise(resolve => setTimeout(resolve, 300));

  content.innerHTML = '';

  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));

  container.classList.remove('fade-in', 'fade-out');
  container.classList.add('hidden');

  container.style.opacity = '';
  background.style.backgroundColor = '';

  return userChoice;
}


/**
 * Show equipment choice scene.
 * @param {Object} choice - The current equipment choice object.
 * @param {number} choiceIndex - The index of the current choice.
 * @returns {Promise<boolean>} True if the user went back, false if they continued.
 */
async function showEquipmentChoiceScene(choice, choiceIndex) {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  background.style.backgroundImage = 'none';
  background.style.backgroundColor = '#111827';
  content.innerHTML = '';
  content.style.zIndex = '10';

  const title = document.createElement('div');
  title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
  title.textContent = `Choose Your Equipment (${choiceIndex + 1} of ${equipmentChoices.length})`;
  content.appendChild(title);

  const complexOptions = choice.options.filter(opt => opt.isComplexChoice);

  if (complexOptions.length > 0) {
    return await showMultiSlotChoiceSelection(content, choice, choiceIndex, complexOptions);
  } else {
    return await showRegularChoiceSelection(content, choice, choiceIndex);
  }
}

/**
 * Show regular (non-complex) choice selection.
 * @returns {Promise<boolean>} True if the user went back, false if they continued.
 */
async function showRegularChoiceSelection(content, choice, choiceIndex) {
  const container = document.getElementById('scene-container');

  const allItems = [];
  choice.options.forEach(option => {
    if (option.isBundle) {
      option.bundle.forEach(bundleItem => allItems.push(bundleItem[0]));
    } else if (!option.isComplexChoice) {
      allItems.push(option.item);
    }
  });
  await preloadItemStats(allItems);

  const scrollContainer = document.createElement('div');
  scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
  scrollContainer.style.maxHeight = 'calc(100vh - 280px)';

  const optionsContainer = document.createElement('div');

  let selectedOption = null;
  let selectedOptionIndex = null;
  let userClickedBack = false;

  const bundles = choice.options.filter(opt => opt.isBundle);
  const simpleItems = choice.options.filter(opt => !opt.isBundle && !opt.isComplexChoice);

  optionsContainer.className = 'flex flex-col gap-6';

  const createClickableOption = (option, optionIndex, containerElement) => {
    containerElement.dataset.optionIndex = optionIndex;
    containerElement.style.cursor = 'pointer';

    containerElement.onclick = (e) => {
      if (e.target.closest('.info-btn')) {
        return;
      }

      const allContainers = optionsContainer.querySelectorAll('[data-option-index]');
      allContainers.forEach(row => {
        row.classList.remove('selected');
      });

      containerElement.classList.add('selected');

      selectedOption = option;
      selectedOptionIndex = optionIndex;
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    };
  };

  if (bundles.length > 0) {
    for (let idx = 0; idx < bundles.length; idx++) {
      const option = bundles[idx];
      const optionIndex = choice.options.indexOf(option);

      const bundleContainer = document.createElement('div');
      bundleContainer.className = 'bg-gray-800 rounded-lg';
      bundleContainer.style.padding = '0.75rem';
      bundleContainer.style.border = '3px solid #374151';
      bundleContainer.style.cursor = 'pointer';
      bundleContainer.style.boxSizing = 'border-box';
      bundleContainer.style.width = 'fit-content';
      bundleContainer.style.margin = '0 auto';

      const bundleLabel = document.createElement('div');
      bundleLabel.className = 'text-center text-gray-400 font-semibold text-xs mb-3';
      bundleLabel.textContent = '📦 BUNDLE';
      bundleContainer.appendChild(bundleLabel);

      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-3 flex-wrap';

      for (const bundleItem of option.bundle) {
        const itemName = bundleItem[0];
        const quantity = bundleItem[1];
        const itemCard = await createSimpleItemCard(itemName, quantity, true);

        itemCard.addEventListener('click', (e) => {
          if (!e.target.closest('.info-btn')) {
            e.stopPropagation();
            bundleContainer.click();
          }
        });
        itemsRow.appendChild(itemCard);
      }

      bundleContainer.appendChild(itemsRow);
      createClickableOption(option, optionIndex, bundleContainer);
      optionsContainer.appendChild(bundleContainer);
    }

    if (simpleItems.length > 0) {
      const separator = document.createElement('div');
      separator.className = 'text-center text-yellow-400 font-bold text-xl my-2';
      separator.textContent = '— OR —';
      optionsContainer.appendChild(separator);
    }
  }

  if (simpleItems.length > 0) {
    const gridContainer = document.createElement('div');
    gridContainer.className = 'flex flex-wrap justify-center gap-3';

    if (simpleItems.length > 4) {
      gridContainer.className = 'grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3';
    }

    for (let idx = 0; idx < simpleItems.length; idx++) {
      const option = simpleItems[idx];
      const optionIndex = choice.options.indexOf(option);
      const itemsRow = document.createElement('div');
      itemsRow.className = 'flex flex-row justify-center gap-2 flex-wrap p-2 bg-gray-800 rounded-lg';
      itemsRow.style.border = '3px solid #374151';
      itemsRow.style.cursor = 'pointer';
      itemsRow.style.boxSizing = 'border-box';
      itemsRow.style.width = 'fit-content';

      const itemCard = await createSimpleItemCard(option.item, option.quantity, false);

      itemCard.addEventListener('click', (e) => {
        if (!e.target.closest('.info-btn')) {
          e.stopPropagation();
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

  const confirmBtn = createContinueButton(0, 'Confirm Choice');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  if (choiceIndex > 0) {
    const backBtn = createBackButton(() => {
      userClickedBack = true;
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      confirmBtn.click();
    });
    document.body.appendChild(backBtn);
    currentBackButton = backBtn;
  }

  container.classList.remove('hidden');
  container.style.opacity = '1';
  await new Promise(resolve => requestAnimationFrame(resolve));
  content.style.animation = 'slideInFromRight 0.3s ease-out';

  await waitForButtonClick(confirmBtn);

  if (currentBackButton) {
    currentBackButton.remove();
    currentBackButton = null;
  }

  if (userClickedBack) {
    content.style.animation = 'wipeLeft 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    container.style.opacity = '0';
    container.classList.add('hidden');
    return true;
  } else {
    content.style.animation = 'wipeRight 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';
    content.style.animation = '';
  }

  selectedChoices[choiceIndex] = selectedOption;

  return false;
}

/**
 * Show multi-slot choice selection (two-step process).
 * @returns {Promise<boolean>} True if the user went back, false if they continued.
 */
async function showMultiSlotChoiceSelection(content, choice, choiceIndex, complexOptions) {
  logger.debug('🎮 showMultiSlotChoiceSelection called!');

  const container = document.getElementById('scene-container');

  const scrollContainer = document.createElement('div');
  scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
  scrollContainer.style.maxHeight = 'calc(100vh - 280px)';

  const optionsContainer = document.createElement('div');
  optionsContainer.className = 'flex flex-col gap-6';

  let selectedConfiguration = null;
  let selectedConfigIndex = null;
  let userClickedBack = false;

  complexOptions.forEach((option, idx) => {
    const optionIndex = choice.options.indexOf(option);

    const configContainer = document.createElement('div');
    configContainer.className = 'p-6 bg-gray-800 rounded-lg';
    configContainer.style.border = '3px solid #374151';
    configContainer.style.cursor = 'pointer';
    configContainer.style.boxSizing = 'border-box';
    configContainer.dataset.optionIndex = optionIndex;

    const configTitle = document.createElement('div');
    configTitle.className = 'text-center text-yellow-400 font-bold text-lg mb-4';
    configTitle.textContent = `Option ${idx + 1}`;
    configContainer.appendChild(configTitle);

    const slotsDesc = document.createElement('div');
    slotsDesc.className = 'text-center text-gray-300 space-y-2';

    if (!option.weaponSlots || option.weaponSlots.length === 0) {
      logger.error('❌ No weaponSlots found for complex option!');
      return;
    }

    option.weaponSlots.forEach((slot) => {
      const slotDiv = document.createElement('div');
      if (slot.type === 'weapon_choice') {
        slotDiv.innerHTML = `<span class="text-green-400">⚔️ Choose a weapon</span> from ${slot.options.length} options`;
      } else if (slot.type === 'fixed_item') {
        slotDiv.innerHTML = `<span class="text-blue-400">🛡️ ${slot.item[0]}</span> (included)`;
      }
      slotsDesc.appendChild(slotDiv);
    });

    configContainer.appendChild(slotsDesc);

    configContainer.onclick = () => {
      optionsContainer.querySelectorAll('[data-option-index]').forEach(row => {
        row.classList.remove('selected');
      });

      configContainer.classList.add('selected');
      selectedConfiguration = option;
      selectedConfigIndex = optionIndex;
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    };

    optionsContainer.appendChild(configContainer);

    if (idx < complexOptions.length - 1) {
      const separator = document.createElement('div');
      separator.className = 'text-center text-yellow-400 font-bold text-xl my-2';
      separator.textContent = '— OR —';
      optionsContainer.appendChild(separator);
    }
  });

  scrollContainer.appendChild(optionsContainer);
  content.appendChild(scrollContainer);

  const confirmBtn = createContinueButton(0, 'Confirm Configuration');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  if (choiceIndex > 0) {
    const backBtn = createBackButton(() => {
      userClickedBack = true;
      confirmBtn.disabled = false;
      confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
      confirmBtn.click();
    });
    document.body.appendChild(backBtn);
    currentBackButton = backBtn;
  }

  container.classList.remove('hidden');
  container.style.opacity = '1';
  content.style.animation = 'slideInFromRight 0.3s ease-out';

  await waitForButtonClick(confirmBtn);

  if (currentBackButton) {
    currentBackButton.remove();
    currentBackButton = null;
  }

  if (userClickedBack) {
    content.style.animation = 'wipeLeft 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    container.style.opacity = '0';
    container.classList.add('hidden');
    return true;
  } else {
    content.style.animation = 'wipeRight 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';
    content.style.animation = '';
  }

  const weaponSelections = {};
  let currentSlotIdx = 0;
  let weaponChoiceIndices = [];

  selectedConfiguration.weaponSlots.forEach((slot, idx) => {
    if (slot.type === 'weapon_choice') {
      weaponChoiceIndices.push(idx);
    }
  });

  let weaponChoicePosition = 0;

  while (weaponChoicePosition < weaponChoiceIndices.length) {
    currentSlotIdx = weaponChoiceIndices[weaponChoicePosition];
    const slot = selectedConfiguration.weaponSlots[currentSlotIdx];

    const result = await showWeaponSlotSelection(slot, weaponChoicePosition, weaponChoiceIndices.length, choiceIndex);

    if (result.shouldGoBack) {
      if (weaponChoicePosition > 0) {
        weaponChoicePosition--;
        const prevSlotIdx = weaponChoiceIndices[weaponChoicePosition];
        delete weaponSelections[prevSlotIdx];
      } else {
        return await showMultiSlotChoiceSelection(content, choice, choiceIndex, complexOptions);
      }
    } else {
      weaponSelections[currentSlotIdx] = result.selectedWeapon;
      weaponChoicePosition++;
    }
  }

  selectedConfiguration.weaponSlots.forEach((slot, idx) => {
    if (slot.type === 'fixed_item') {
      weaponSelections[idx] = slot.item;
    }
  });

  const finalSelection = {
    isComplexChoice: true,
    weapons: Object.values(weaponSelections)
  };

  selectedChoices[choiceIndex] = finalSelection;
  return false;
}

/**
 * Show weapon selection for a single slot.
 * @returns {Promise<Object>} Object with selectedWeapon and shouldGoBack.
 */
async function showWeaponSlotSelection(slot, slotIndex, totalSlots, choiceIndex) {
  const container = document.getElementById('scene-container');
  const content = document.getElementById('scene-content');

  content.innerHTML = '';

  const title = document.createElement('div');
  title.className = 'text-xl md:text-2xl font-bold text-yellow-400 mb-4';
  title.textContent = `Choose Weapon ${slotIndex + 1} of ${totalSlots}`;
  content.appendChild(title);

  await preloadItemStats(slot.options.map(w => w[0]));

  const scrollContainer = document.createElement('div');
  scrollContainer.className = 'w-full max-w-6xl mx-auto overflow-y-auto px-4 mb-4';
  scrollContainer.style.maxHeight = 'calc(100vh - 280px)';

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
    itemContainer.style.border = '3px solid #374151';
    itemContainer.style.cursor = 'pointer';
    itemContainer.style.boxSizing = 'border-box';
    itemContainer.dataset.optionIndex = idx;

    const itemCard = await createSimpleItemCard(weaponName, quantity, false);

    itemCard.addEventListener('click', (e) => {
      if (!e.target.closest('.info-btn')) {
        e.stopPropagation();
        itemContainer.click();
      }
    });

    itemContainer.onclick = (e) => {
      if (e.target.closest('.info-btn')) return;

      gridContainer.querySelectorAll('[data-option-index]').forEach(row => {
        row.classList.remove('selected');
      });

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

  const confirmBtn = createContinueButton(0, 'Confirm Weapon');
  confirmBtn.disabled = true;
  confirmBtn.classList.add('opacity-50', 'cursor-not-allowed');
  content.appendChild(confirmBtn);

  const existingBackButtons = document.querySelectorAll('.pixel-back-btn');
  existingBackButtons.forEach((btn) => {
    btn.remove();
  });

  const backBtn = createBackButton(() => {
    userClickedBack = true;
    confirmBtn.disabled = false;
    confirmBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    confirmBtn.click();
  });
  document.body.appendChild(backBtn);
  currentBackButton = backBtn;

  container.classList.remove('hidden');
  container.style.opacity = '1';
  content.style.animation = 'slideInFromRight 0.3s ease-out';

  await waitForButtonClick(confirmBtn);

  if (currentBackButton) {
    currentBackButton.remove();
    currentBackButton = null;
  }

  if (userClickedBack) {
    content.style.animation = 'wipeLeft 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';

    container.style.opacity = '0';
    container.classList.add('hidden');
  } else {
    content.style.animation = 'wipeRight 0.3s ease-in';
    await new Promise(resolve => setTimeout(resolve, 300));
    content.innerHTML = '';
    content.style.animation = '';
  }

  return {
    selectedWeapon: selectedWeapon,
    shouldGoBack: userClickedBack
  };
}

/**
 * Create a simple item card (for use in weapon slots and simple items)
 * @param {string} itemName - The ID of the item.
 * @param {number} quantity - The quantity of the item.
 * @param {boolean} isInBundle - True if the item is part of a bundle.
 * @returns {Promise<HTMLDivElement>} The item card element.
 */
async function createSimpleItemCard(itemName, quantity, isInBundle = false) {
  const card = document.createElement('div');
  card.className = 'item-card bg-gray-800 rounded-lg relative overflow-hidden';
  card.style.width = '110px';
  card.style.height = '110px';
  card.style.aspectRatio = '1/1';

  if (isInBundle) {
    card.style.border = '2px solid #4b5563';
  }

  if (!isInBundle) {
    card.style.cursor = 'pointer';
  }

  const itemData = getItemById(itemName); // Use getItemById from staticData

  const img = document.createElement('img');
  img.src = itemData?.image || `/res/img/items/${itemData?.id || itemName}.png`;
  img.alt = itemData?.name || itemName;
  img.className = 'absolute inset-0 w-full h-full object-contain p-3';
  img.style.imageRendering = 'pixelated';
  img.style.imageRendering = '-moz-crisp-edges';
  img.style.imageRendering = 'crisp-edges';
  card.appendChild(img);

  const rarityDot = document.createElement('div');
  rarityDot.className = 'absolute top-1.5 right-1.5 z-10';
  rarityDot.style.width = '10px';
  rarityDot.style.height = '10px';
  rarityDot.style.borderRadius = '50%';
  rarityDot.style.backgroundColor = getRarityColor(itemName);
  rarityDot.style.border = '1px solid rgba(0,0,0,0.3)';
  card.appendChild(rarityDot);

  if (quantity && quantity > 1) {
    const qtyText = document.createElement('div');
    qtyText.className = 'absolute top-1 left-1.5 z-10 font-bold';
    qtyText.style.color = '#a855f7';
    qtyText.style.fontSize = '0.85rem';
    qtyText.style.textShadow = '0 0 3px rgba(0,0,0,0.8), 1px 1px 2px rgba(0,0,0,0.9)';
    qtyText.textContent = `×${quantity}`;
    card.appendChild(qtyText);
  }

  const name = document.createElement('div');
  name.className = 'item-name absolute bottom-0 left-0 right-0 text-center text-white font-semibold px-1 py-1 z-10';
  name.style.fontSize = '0.65rem';
  name.style.lineHeight = '0.75rem';
  name.style.backgroundColor = 'rgba(0,0,0,0.6)';
  name.style.textShadow = '0 1px 2px rgba(0,0,0,0.8)';
  name.dataset.itemId = itemName;
  name.textContent = itemName;
  card.appendChild(name);

  // Fetch and update with actual item name using the imported getItemStats
  // Note: getItemStats returns HTML, so we need to parse it for the name.
  const statsHtml = await getItemStats(itemName);
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = statsHtml;
  const nameElement = tempDiv.querySelector('.font-bold.text-yellow-400');
  if (nameElement) {
    name.textContent = nameElement.textContent;
  }


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
 * Get rarity color based on item rarity (placeholder)
 * @param {string} itemName - The ID of the item.
 * @returns {string} The CSS color string.
 */
function getRarityColor(itemName) {
  // TODO: This should come from item data
  const rarityColors = {
    'common': '#9ca3af', // grey
    'uncommon': '#10b981', // green
    'rare': '#3b82f6', // blue
    'legendary': '#a855f7', // purple
    'mythic': '#f97316' // orange
  };
  return rarityColors.common; // Default to common for now
}

/**
 * Show item info modal.
 * @param {string} itemName - The ID of the item to show info for.
 */
async function showItemModal(itemName) {
  hideItemModal();

  const backdrop = document.createElement('div');
  backdrop.id = 'item-modal-backdrop';
  backdrop.className = 'fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50';
  backdrop.onclick = (e) => {
    if (e.target === backdrop) {
      hideItemModal();
    }
  };

  const modal = document.createElement('div');
  modal.className = 'bg-gray-800 border-2 border-yellow-400 rounded-lg p-6 max-w-md w-full mx-4 relative';
  modal.onclick = (e) => e.stopPropagation();

  const closeBtn = document.createElement('button');
  closeBtn.className = 'absolute top-2 right-2 w-8 h-8 bg-red-600 hover:bg-red-700 text-white rounded-full font-bold transition-colors';
  closeBtn.textContent = '✕';
  closeBtn.onclick = hideItemModal;
  modal.appendChild(closeBtn);

  modal.innerHTML += '<div class="text-center text-white">Loading...</div>';

  backdrop.appendChild(modal);
  document.body.appendChild(backdrop);

  const statsHTML = await getItemStats(itemName);

  modal.innerHTML = '';
  modal.appendChild(closeBtn);

  const contentDiv = document.createElement('div');
  contentDiv.className = 'text-white';
  contentDiv.innerHTML = statsHTML;
  modal.appendChild(contentDiv);
}

/**
 * Hide item modal.
 */
function hideItemModal() {
  const backdrop = document.getElementById('item-modal-backdrop');
  if (backdrop) {
    backdrop.remove();
  }
}

/**
 * Preload item stats for multiple items into cache.
 * @param {Array<string>} itemNames - An array of item IDs to preload.
 */
async function preloadItemStats(itemNames) {
  const uniqueNames = [...new Set(itemNames)];
  logger.debug('🔄 Preloading item stats for:', uniqueNames);

  for (const name of uniqueNames) {
    await getItemStats(name); // Calling getItemStats will internally cache the result
  }
  logger.debug('✅ Preloaded cache for items.');
}

/**
 * Fetch pack data from API to get contents.
 * @param {string} packName - The ID of the pack.
 * @returns {Promise<Object|null>} The pack data object, or null if not found/error.
 */
async function fetchPackData(packName) {
  try {
    logger.debug('📦 Fetching pack data for:', packName);
    const response = await fetch(`/api/items?name=${encodeURIComponent(packName)}`);
    if (!response.ok) {
      logger.warn('❌ Failed to fetch pack data for:', packName);
      return null;
    }

    const items = await response.json();

    if (!items || items.length === 0) {
      logger.warn('⚠️ No pack data found for:', packName);
      return null;
    }

    const packData = items[0];
    return packData;
  } catch (error) {
    logger.error('❌ Error fetching pack data:', packName, error);
    return null;
  }
}

/**
 * Animate scene out (used by equipment selection scenes).
 * @param {HTMLElement} content - The content container element.
 * @param {HTMLElement} container - The main scene container element.
 */
async function animateSceneOut(content, container) {
  const textElements = content.querySelectorAll('.scene-text, .item-card, .pixel-continue-btn');
  textElements.forEach(el => {
    el.style.animation = 'wipeOut 0.6s ease-in forwards';
  });

  await new Promise(resolve => setTimeout(resolve, 600)); // Wait for wipe animation
  await new Promise(resolve => setTimeout(resolve, 1000)); // Additional delay if needed

  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800)); // Wait for fade out
  container.classList.add('hidden');
}