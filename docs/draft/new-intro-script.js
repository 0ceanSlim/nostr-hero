// NEW INTRO SCRIPT FOR GAME-INTRO.HTML
// This replaces the typewriter-based intro with fade-based scene transitions

// Global variables
let generatedCharacter = null;
let introData = null;
let startingEquipment = null;
let selectedEquipment = {};
let playerName = '';

// Scene display function
async function showScene(sceneConfig) {
  const container = document.getElementById('scene-container');
  const background = document.getElementById('scene-background');
  const content = document.getElementById('scene-content');

  // Set up scene
  if (sceneConfig.image) {
    background.style.backgroundImage = `url(/res/img/scene/${sceneConfig.image})`;
    background.style.opacity = '1';
  } else {
    background.style.backgroundImage = '';
    background.style.opacity = '0';
  }

  // Set up content
  content.innerHTML = '';
  const textElement = document.createElement('div');
  textElement.className = 'scene-text';

  // Style based on scene type
  if (sceneConfig.isQuote) {
    textElement.className += ' text-3xl md:text-4xl font-bold text-yellow-400 leading-relaxed';
  } else if (sceneConfig.isLetter) {
    const letterDiv = document.createElement('div');
    letterDiv.className = 'letter-container';
    letterDiv.innerHTML = `<div class="text-sm opacity-70 mb-4 text-center">The Letter:</div><div>${sceneConfig.text}</div>`;
    content.appendChild(letterDiv);

    // Show scene
    container.classList.remove('hidden');
    container.classList.add('fade-in');

    await new Promise(resolve => setTimeout(resolve, sceneConfig.duration || 5000));

    // Hide scene
    container.classList.remove('fade-in');
    container.classList.add('fade-out');
    await new Promise(resolve => setTimeout(resolve, 800));
    container.classList.add('hidden');
    container.classList.remove('fade-out');
    return;
  } else {
    textElement.className += ' text-lg md:text-xl leading-relaxed';
  }

  textElement.textContent = sceneConfig.text;
  content.appendChild(textElement);

  // Show scene
  container.classList.remove('hidden');
  container.classList.add('fade-in');

  await new Promise(resolve => setTimeout(resolve, sceneConfig.duration || 4000));

  // Hide scene
  container.classList.remove('fade-in');
  container.classList.add('fade-out');
  await new Promise(resolve => setTimeout(resolve, 800));
  container.classList.add('hidden');
  container.classList.remove('fade-out');
}

// Get equipment category from class
function getEquipmentCategory(className) {
  const categories = {
    'Fighter': 'warrior',
    'Barbarian': 'warrior',
    'Paladin': 'warrior', // Primary category
    'Cleric': 'faithful',
    'Monk': 'faithful',
    'Ranger': 'wilderness', // Primary category
    'Druid': 'wilderness',
    'Wizard': 'arcane',
    'Sorcerer': 'arcane',
    'Warlock': 'arcane',
    'Rogue': 'clever',
    'Bard': 'clever'
  };
  return categories[className] || 'warrior';
}

// Main intro sequence
async function startIntroSequence() {
  playerName = document.getElementById('character-name').value.trim();
  document.getElementById('name-screen').classList.add('hidden');

  // Load intro data
  const response = await fetch('/data/character/introductions.json');
  introData = await response.json();

  // 1. Scene 1
  await showScene({
    text: introData.scene1.text,
    image: introData.scene1.image,
    duration: 4000
  });

  // 2. Scene 2
  await showScene({
    text: introData.scene2.text,
    image: introData.scene2.image,
    duration: 4000
  });

  // 3. Final Words (black screen)
  await showScene({
    text: introData.final_words.text,
    isQuote: true,
    duration: 5000
  });

  // 4. Background Scene
  const bgIntro = introData.background_intros[generatedCharacter.background];
  if (bgIntro) {
    await showScene({
      text: bgIntro.text,
      image: bgIntro.image,
      duration: 4000
    });
  }

  // 5. Letter Intro
  await showScene({
    text: introData.letter_intro.text,
    image: introData.letter_intro.image,
    duration: 3500
  });

  // 6. Letter Reading
  const bgLetter = introData.background_letters[generatedCharacter.background];
  if (bgLetter) {
    await showScene({
      text: bgLetter.text,
      image: bgLetter.image,
      isLetter: true,
      duration: 6000
    });
  }

  // 7. Equipment Intro (class-based)
  const equipCategory = getEquipmentCategory(generatedCharacter.class);
  const equipIntro = introData.equipment_intros[equipCategory];
  if (equipIntro) {
    await showScene({
      text: equipIntro.text,
      image: equipIntro.image,
      duration: 5000
    });
  }

  // 8. Scene 5
  await showScene({
    text: introData.scene5.text,
    image: introData.scene5.image,
    duration: 3500
  });

  // 9. Equipment Selection
  document.getElementById('scene-container').classList.add('hidden');
  showEquipmentSelection();
}

// Continue after equipment selection
async function continueAfterEquipment() {
  document.getElementById('equipment-selection').classList.add('hidden');

  // 10. Scene 5a (black screen)
  await showScene({
    text: introData.scene5a.text,
    isQuote: true,
    duration: 4500
  });

  // 11. Scene 6
  await showScene({
    text: introData.scene6.text,
    image: introData.scene6.image,
    duration: 4000
  });

  // 12. Pack selection would go here (for now skip to departure)

  // 13. Departure
  await showScene({
    text: introData.departure.text,
    image: introData.departure.image,
    duration: 4000
  });

  // 14. Final Text (black screen) + button
  const container = document.getElementById('scene-container');
  const content = document.getElementById('scene-content');
  content.innerHTML = `
    <div class="scene-text text-3xl md:text-4xl font-bold text-yellow-400 leading-relaxed mb-8">
      ${introData.final_text.text}
    </div>
    <button
      onclick="startAdventure()"
      class="bg-green-600 hover:bg-green-700 text-black px-8 py-4 rounded-lg font-bold text-xl transition-colors mt-8"
    >
      Begin Journey
    </button>
  `;
  container.classList.remove('hidden');
  container.classList.add('fade-in');
}
