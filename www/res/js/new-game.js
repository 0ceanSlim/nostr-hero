// New Game Character Generation JavaScript
// Moved from inline to avoid template parsing issues

let generatedCharacter = null;
let characterIntroduction = null;
let startingEquipment = null;
let selectedEquipment = {};

// Initialize character generation
document.addEventListener('DOMContentLoaded', async function() {
  console.log('🎮 Initializing new game...');

  if (!window.sessionManager) {
    showMessage('❌ Session manager not available', 'error');
    goToSaves();
    return;
  }

  try {
    await window.sessionManager.init();

    if (!window.sessionManager.isAuthenticated()) {
      console.log('❌ User not authenticated, redirecting to home');
      window.location.href = '/';
      return;
    }

    const session = window.sessionManager.getSession();
    console.log('✅ User authenticated:', session.npub);

    await generateCharacterFromNpub(session.npub);

  } catch (error) {
    console.error('❌ Failed to initialize new game:', error);
    showMessage('❌ Failed to create character: ' + error.message, 'error');
  }
});

async function generateCharacterFromNpub(npub) {
  try {
    // Initialize character generator
    await window.characterGenerator.initialize();

    // Generate character from npub
    generatedCharacter = await window.characterGenerator.generateCharacter(npub);
    characterIntroduction = window.characterGenerator.generateIntroduction(generatedCharacter);
    startingEquipment = window.characterGenerator.generateStartingEquipment(generatedCharacter);

    console.log('Generated character:', generatedCharacter);
    console.log('Character introduction:', characterIntroduction);
    console.log('Starting equipment:', startingEquipment);

    // Display the character
    displayGeneratedCharacter();

    // Hide loading, show character
    document.getElementById('generation-loading').classList.add('hidden');
    document.getElementById('character-display').classList.remove('hidden');

  } catch (error) {
    console.error('❌ Character generation failed:', error);
    showMessage('❌ Character generation failed: ' + error.message, 'error');

    // Show error state
    const loadingEl = document.getElementById('generation-loading');
    loadingEl.innerHTML = '';

    const errorDiv = document.createElement('div');
    errorDiv.className = 'text-center py-8';

    const errorIcon = document.createElement('div');
    errorIcon.className = 'text-red-400 text-xl mb-2';
    errorIcon.textContent = '❌';
    errorDiv.appendChild(errorIcon);

    const errorTitle = document.createElement('p');
    errorTitle.className = 'text-red-400';
    errorTitle.textContent = 'Character generation failed';
    errorDiv.appendChild(errorTitle);

    const errorMessage = document.createElement('p');
    errorMessage.className = 'text-gray-500 text-sm';
    errorMessage.textContent = error.message;
    errorDiv.appendChild(errorMessage);

    const retryButton = document.createElement('button');
    retryButton.className = 'mt-4 bg-yellow-600 hover:bg-yellow-700 text-gray-900 px-4 py-2 rounded';
    retryButton.textContent = '🔄 Try Again';
    retryButton.onclick = function() { window.location.reload(); };
    errorDiv.appendChild(retryButton);

    loadingEl.appendChild(errorDiv);
  }
}

function displayGeneratedCharacter() {
  if (!generatedCharacter) return;

  // Character icon
  const raceIcons = {
    'Human': '👤', 'Elf': '🧝', 'Dwarf': '🧔', 'Halfling': '🧒',
    'Dragonborn': '🐲', 'Gnome': '🧙', 'Half-Elf': '👨‍🎤',
    'Half-Orc': '👹', 'Tiefling': '😈', 'Orc': '👹'
  };

  document.getElementById('character-icon').textContent = raceIcons[generatedCharacter.race] || '⚔️';
  document.getElementById('character-name').textContent = generatedCharacter.name;
  document.getElementById('character-title').textContent = generatedCharacter.race + ' ' + generatedCharacter.class;

  // Character details
  document.getElementById('char-race').textContent = generatedCharacter.race;
  document.getElementById('char-class').textContent = generatedCharacter.class;
  document.getElementById('char-background').textContent = generatedCharacter.background;
  document.getElementById('char-alignment').textContent = generatedCharacter.alignment;
  document.getElementById('char-level').textContent = generatedCharacter.level;

  // Ability scores
  document.getElementById('stat-strength').textContent = generatedCharacter.stats.strength;
  document.getElementById('stat-dexterity').textContent = generatedCharacter.stats.dexterity;
  document.getElementById('stat-constitution').textContent = generatedCharacter.stats.constitution;
  document.getElementById('stat-intelligence').textContent = generatedCharacter.stats.intelligence;
  document.getElementById('stat-wisdom').textContent = generatedCharacter.stats.wisdom;
  document.getElementById('stat-charisma').textContent = generatedCharacter.stats.charisma;

  // Character stats
  document.getElementById('char-hp').textContent = generatedCharacter.hp;
  document.getElementById('char-mana').textContent = generatedCharacter.mana;
  document.getElementById('char-gold').textContent = generatedCharacter.gold;
}

function showIntroduction() {
  if (!characterIntroduction) return;

  const introContent = document.getElementById('intro-content');
  const introSection = document.getElementById('introduction-section');

  // Clear any existing content
  introContent.innerHTML = '';

  // Base introduction
  const introScene = document.createElement('div');
  introScene.className = 'intro-scene mb-6';

  const scene1 = document.createElement('p');
  scene1.className = 'mb-3';
  scene1.textContent = characterIntroduction.baseIntro.scene1;
  introScene.appendChild(scene1);

  const scene2 = document.createElement('p');
  scene2.className = 'mb-3';
  scene2.textContent = characterIntroduction.baseIntro.scene2;
  introScene.appendChild(scene2);

  const scene3 = document.createElement('p');
  scene3.className = 'mb-3';
  scene3.textContent = characterIntroduction.baseIntro.scene3;
  introScene.appendChild(scene3);

  const quote = document.createElement('blockquote');
  quote.className = 'border-l-4 border-yellow-400 pl-4 italic text-yellow-200 mb-3';
  quote.textContent = '"' + characterIntroduction.baseIntro.final_words + '"';
  introScene.appendChild(quote);

  const letterIntro = document.createElement('p');
  letterIntro.textContent = characterIntroduction.baseIntro.letter_intro;
  introScene.appendChild(letterIntro);

  introContent.appendChild(introScene);

  // Background-specific content with race elements
  if (characterIntroduction.backgroundIntro) {
    const backgroundDiv = document.createElement('div');
    backgroundDiv.className = 'background-intro mb-6';

    const title = document.createElement('h4');
    title.className = 'text-lg font-medium text-cyan-400 mb-3';
    title.textContent = 'Your ' + generatedCharacter.race + ' Heritage and ' + generatedCharacter.background + ' Past';
    backgroundDiv.appendChild(title);

    const scenePara = document.createElement('p');
    scenePara.className = 'mb-3';
    scenePara.textContent = addRaceFlavorToScene(characterIntroduction.backgroundIntro.scene, generatedCharacter.race);
    backgroundDiv.appendChild(scenePara);

    const letterDiv = document.createElement('div');
    letterDiv.className = 'bg-gray-700 rounded p-4 mt-4';

    const letterTitle = document.createElement('h5');
    letterTitle.className = 'text-sm font-medium text-gray-300 mb-2';
    letterTitle.textContent = 'The Letter:';
    letterDiv.appendChild(letterTitle);

    const letterText = document.createElement('p');
    letterText.className = 'text-sm text-gray-200 italic';
    letterText.textContent = characterIntroduction.backgroundIntro.letter;
    letterDiv.appendChild(letterText);

    backgroundDiv.appendChild(letterDiv);

    introContent.appendChild(backgroundDiv);
  }

  // Equipment introduction
  if (characterIntroduction.equipmentIntro) {
    const equipDiv = document.createElement('div');
    equipDiv.className = 'equipment-intro mb-6';

    const equipTitle = document.createElement('h4');
    equipTitle.className = 'text-lg font-medium text-purple-400 mb-3';
    equipTitle.textContent = 'Your Training';
    equipDiv.appendChild(equipTitle);

    const equipScene = document.createElement('p');
    equipScene.className = 'mb-3';
    equipScene.textContent = characterIntroduction.equipmentIntro.scene;
    equipDiv.appendChild(equipScene);

    const equipQuote = document.createElement('blockquote');
    equipQuote.className = 'border-l-4 border-purple-400 pl-4 italic text-purple-200';
    equipQuote.textContent = '"' + characterIntroduction.equipmentIntro.quote + '"';
    equipDiv.appendChild(equipQuote);

    introContent.appendChild(equipDiv);
  }

  // Final note
  const finalDiv = document.createElement('div');
  finalDiv.className = 'final-note mb-6';

  const finalText = document.createElement('p');
  finalText.className = 'mb-2';
  finalText.textContent = characterIntroduction.finalNote.text;
  finalDiv.appendChild(finalText);

  const finalQuote = document.createElement('blockquote');
  finalQuote.className = 'border-l-4 border-green-400 pl-4 italic text-green-200';
  finalQuote.textContent = '"' + characterIntroduction.finalNote.quote + '"';
  finalDiv.appendChild(finalQuote);

  introContent.appendChild(finalDiv);

  // Departure
  const departureDiv = document.createElement('div');
  departureDiv.className = 'departure text-center bg-gray-700 rounded p-4';
  const departurePara = document.createElement('p');
  departurePara.className = 'text-gray-300';
  departurePara.textContent = characterIntroduction.departure.text;
  departureDiv.appendChild(departurePara);
  introContent.appendChild(departureDiv);
  introSection.classList.remove('hidden');

  // Update buttons
  document.getElementById('show-intro-btn').style.display = 'none';
  document.getElementById('show-equipment-btn').style.display = 'block';
}

function showEquipmentSelection() {
  const equipmentSection = document.getElementById('equipment-selection');
  const equipmentChoices = document.getElementById('equipment-choices');
  const startingInventory = document.getElementById('starting-inventory');

  // Show automatic equipment
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
      quantity.textContent = ' (' + item.quantity + ')';
      itemDiv.appendChild(quantity);
    }

    startingInventory.appendChild(itemDiv);
  });

  // Show equipment choices
  equipmentChoices.innerHTML = '';
  if (startingEquipment.choices && startingEquipment.choices.length > 0) {
    startingEquipment.choices.forEach((choice, index) => {
      const choiceGroup = document.createElement('div');
      choiceGroup.className = 'choice-group bg-gray-700 rounded p-4';

      const choiceTitle = document.createElement('h5');
      choiceTitle.className = 'text-md font-medium text-yellow-400 mb-3';
      choiceTitle.textContent = 'Choose one:';
      choiceGroup.appendChild(choiceTitle);

      const optionsDiv = document.createElement('div');
      optionsDiv.className = 'space-y-2';

      choice.options.forEach((option, optionIndex) => {
        const label = document.createElement('label');
        label.className = 'flex items-center space-x-3 cursor-pointer hover:bg-gray-600 rounded p-2';

        const input = document.createElement('input');
        input.type = 'radio';
        input.name = 'choice_' + index;
        input.value = option.item; // Use option.item for the value
        input.id = 'choice_' + index + '_' + optionIndex;
        input.onchange = function() { selectEquipment(index, option); };
        label.appendChild(input);

        const span = document.createElement('span');
        span.className = 'text-white';
        span.textContent = option.item; // Display option.item as the label
        label.appendChild(span);

        optionsDiv.appendChild(label);
      });

      choiceGroup.appendChild(optionsDiv);
      equipmentChoices.appendChild(choiceGroup);
    });
  } else {
    const noChoices = document.createElement('p');
    noChoices.className = 'text-gray-400';
    noChoices.textContent = 'No equipment choices needed - your starting gear is ready!';
    equipmentChoices.appendChild(noChoices);
  }
  equipmentSection.classList.remove('hidden');

  // Update buttons
  document.getElementById('show-equipment-btn').style.display = 'none';
  document.getElementById('start-adventure-btn').style.display = 'block';
}

function selectEquipment(choiceIndex, option) {
  selectedEquipment[choiceIndex] = option;
  console.log('Selected equipment:', selectedEquipment);
}

async function startAdventure() {
  try {
    showMessage('🎮 Starting your adventure...', 'info');

    // Prepare final character data
    const finalCharacter = {
      ...generatedCharacter,
      equipment: startingEquipment.equipment,
      selectedEquipment: selectedEquipment
    };

    // Prepare final inventory (starting items + selected equipment)
    let finalInventory = [...startingEquipment.inventory];

    // Add selected equipment to inventory
    Object.values(selectedEquipment).forEach(option => {
        if (option.isBundle) {
            option.bundle.forEach(bundleItem => {
                const existingItem = finalInventory.find(i => i.item === bundleItem[0]);
                if (existingItem) {
                    existingItem.quantity += bundleItem[1];
                } else {
                    finalInventory.push({ item: bundleItem[0], quantity: bundleItem[1] });
                }
            });
        } else {
            const existingItem = finalInventory.find(i => i.item === option.item);
            if (existingItem) {
                existingItem.quantity += option.quantity;
            } else {
                finalInventory.push({ item: option.item, quantity: option.quantity });
            }
        }
    });

    finalCharacter.inventory = finalInventory;

    console.log('Final character:', finalCharacter);

    // Create new save file
    const session = window.sessionManager.getSession();
    // Create complete game state matching our game-state.js structure
    const gameState = {
      character: finalCharacter,
      inventory: finalCharacter.inventory || [],
      equipment: finalCharacter.equipment || {},
      spells: finalCharacter.spells || [],
      location: {
        current: finalCharacter.city || 'kingdom',
        discovered: [finalCharacter.city || 'kingdom']
      },
      combat: null
    };

    const saveData = {
      character: finalCharacter,
      gameState: gameState,
      location: finalCharacter.city || 'kingdom', // Use character's starting city
      npub: session.npub,
      created_at: new Date().toISOString()
    };

    const response = await fetch(`/api/saves/${session.npub}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(saveData)
    });

    if (response.ok) {
      const result = await response.json();
      showMessage('✅ Adventure begins!', 'success');

      // Redirect to game with the new save
      setTimeout(() => {
        window.location.href = '/game?save=' + result.id;
      }, 1500);
    } else {
      throw new Error('Failed to create save file');
    }

  } catch (error) {
    console.error('❌ Failed to start adventure:', error);
    showMessage('❌ Failed to start adventure: ' + error.message, 'error');
  }
}

function regenerateCharacter() {
  // For development - regenerate with random seed
  const session = window.sessionManager.getSession();
  const randomSeed = session.npub + Date.now().toString();
  window.location.href = '/new-game?seed=' + encodeURIComponent(randomSeed);
}

function goToSaves() {
  window.location.href = '/saves';
}

// Add race-specific flavor to background scenes
function addRaceFlavorToScene(scene, race) {
  const raceModifiers = {
    'Human': 'With the adaptability common to your human lineage, ',
    'Elf': 'Your elven grace and centuries-old wisdom guided you as ',
    'Dwarf': 'The sturdy resilience of your dwarven ancestors strengthened you while ',
    'Halfling': 'Your halfling\'s natural curiosity and comfort with simple pleasures meant ',
    'Dragonborn': 'The draconic heritage flowing in your veins instilled pride as ',
    'Gnome': 'Your gnomish ingenuity and love of tinkering made you naturally suited as ',
    'Half-Elf': 'Caught between two worlds, your half-elf nature gave you unique perspective as ',
    'Half-Orc': 'Despite the suspicious looks your orcish heritage often earned, you proved yourself as ',
    'Tiefling': 'Though others feared your infernal bloodline, you channeled that strength into becoming ',
    'Orc': 'Your orcish strength and warrior culture shaped you into '
  };

  const modifier = raceModifiers[race] || '';

  // Insert race modifier at the beginning or find a natural place to insert it
  if (scene.includes('The old caretaker')) {
    return scene.replace('The old caretaker', modifier + 'the old caretaker');
  } else if (scene.includes('They')) {
    return scene.replace('They', modifier + 'they');
  } else {
    return modifier + scene.charAt(0).toLowerCase() + scene.slice(1);
  }
}

console.log('🎮 New game scripts loaded');