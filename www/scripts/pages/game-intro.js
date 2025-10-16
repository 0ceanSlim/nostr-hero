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
let playerName = "";

// ============================================================================
// MUSIC PLAYLIST SYSTEM
// ============================================================================
let currentTrack = 0;
const tracks = ["intro-music", "game-music"];

function setupMusicPlaylist() {
  tracks.forEach((trackId, index) => {
    const audio = document.getElementById(trackId);
    audio.volume = 0.3;

    audio.addEventListener("ended", () => {
      // Move to next track
      currentTrack = (currentTrack + 1) % tracks.length;
      const nextAudio = document.getElementById(tracks[currentTrack]);
      nextAudio.currentTime = 0;
      nextAudio.play().catch((e) => console.log("Music autoplay blocked:", e));
    });
  });
}

function startMusic() {
  const firstTrack = document.getElementById(tracks[0]);
  firstTrack.play().catch((e) => console.log("Music autoplay blocked:", e));
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
  const container = document.getElementById("scene-container");
  const background = document.getElementById("scene-background");
  const content = document.getElementById("scene-content");

  // Set up background
  if (config.image) {
    background.style.backgroundImage = `url(/res/img/scene/${config.image})`;
  } else {
    background.style.backgroundImage = "";
  }

  // Clear and set up content
  content.innerHTML = "";

  // Add skip button (top right)
  const skipBtn = document.createElement("button");
  skipBtn.className = "pixel-skip-btn";
  skipBtn.textContent = "Skip Scene";
  skipBtn.onclick = () => {
    continueBtn.click();
  };
  container.appendChild(skipBtn);

  if (config.isLetter) {
    // Letter scene - special styling positioned on background letter
    const letterDiv = document.createElement("div");
    letterDiv.className = "letter-container";
    letterDiv.innerHTML = `
      <div>${config.text}</div>
    `;
    content.appendChild(letterDiv);
  } else {
    // Regular scene
    const textElement = document.createElement("div");
    textElement.className = "scene-text";

    if (config.isQuote) {
      // Quote styling - large, centered, light purple with responsive sizing
      textElement.className +=
        " text-2xl md:text-3xl lg:text-4xl font-bold text-purple-500 leading-loose";
      textElement.style.textShadow =
        "3px 3px 6px rgba(0, 0, 0, 1), 0 0 12px rgba(0, 0, 0, 0.95), 1px 1px 3px rgba(0, 0, 0, 1)";
      textElement.style.lineHeight = "1.8";
      textElement.style.maxWidth = "90vw";
      textElement.style.maxHeight = "80vh";
      textElement.style.overflow = "visible";
      textElement.style.display = "flex";
      textElement.style.flexDirection = "column";
      textElement.style.alignItems = "center";
      textElement.style.justifyContent = "center";

      // Split text into sentences and animate each one
      const sentences = config.text.match(/[^.!?]+[.!?]+/g) || [config.text];
      textElement.innerHTML = ""; // Clear content

      // Calculate timing: divide buttonDelay by (number of sentences - 1) so last sentence appears when button shows
      const totalDuration = (config.buttonDelay || 7000) / 1000; // Convert ms to seconds
      const delayPerSentence =
        sentences.length > 1 ? totalDuration / (sentences.length - 1) : 0;

      sentences.forEach((sentence, index) => {
        const sentenceSpan = document.createElement("span");
        sentenceSpan.textContent = sentence.trim() + " ";
        sentenceSpan.style.opacity = "0";
        sentenceSpan.style.display = "block";
        sentenceSpan.style.marginBottom = "0.5em";
        sentenceSpan.style.animation = `fadeInSentence 0.8s ease-out forwards`;
        sentenceSpan.style.animationDelay = `${index * delayPerSentence}s`;
        textElement.appendChild(sentenceSpan);
      });
    } else {
      // Normal scene text - slightly larger size with dark shadow
      textElement.className += " text-2xl md:text-3xl leading-relaxed";
      textElement.style.textShadow =
        "3px 3px 6px rgba(0, 0, 0, 1), 0 0 12px rgba(0, 0, 0, 0.95), 1px 1px 3px rgba(0, 0, 0, 1)";
      textElement.textContent = config.text;
    }

    if (!config.isQuote) {
      textElement.textContent = config.text;
    }
    content.appendChild(textElement);
  }

  // Add Continue button using component
  const buttonDelay =
    config.buttonDelay !== undefined ? config.buttonDelay : 7000;
  const continueBtn = createContinueButton(buttonDelay);

  // For letter scenes, position button below the letter area
  if (config.isLetter) {
    continueBtn.style.cssText = `
      position: fixed !important;
      bottom: 10vh !important;
      left: 50% !important;
      right: auto !important;
      transform: translateX(-50%) !important;
      z-index: 100 !important;
      margin: 0 !important;
    `;
  }

  content.appendChild(continueBtn);

  // Show container with fade-in
  container.classList.remove("hidden", "fade-out");
  // First ensure we're not in fade-in state
  container.classList.remove("fade-in");
  // Force reflow to ensure classes are applied
  void container.offsetHeight;
  // Now add fade-in to trigger transition
  container.classList.add("fade-in");

  // Wait for user to click Continue
  await waitForButtonClick(continueBtn);

  // Remove skip button
  skipBtn.remove();

  // For letter scenes, just fade out without wipe animation
  if (config.isLetter) {
    // Fade out the letter container and button
    const textElements = content.querySelectorAll(
      ".letter-container, .pixel-continue-btn"
    );
    textElements.forEach((el) => {
      el.style.transition = "opacity 0.6s ease-out";
      el.style.opacity = "0";
    });

    // Wait for fade animation to complete
    await new Promise((resolve) => setTimeout(resolve, 600));
  } else {
    // Animate text out first (wipe down) for regular scenes
    const textElements = content.querySelectorAll(
      ".scene-text, .pixel-continue-btn"
    );
    textElements.forEach((el) => {
      el.style.animation = "wipeOut 0.6s ease-in forwards";
    });

    // Wait for text animation to complete
    await new Promise((resolve) => setTimeout(resolve, 600));
  }

  // Clear content
  content.innerHTML = "";

  // Then fade out the scene
  container.classList.remove("fade-in");
  container.classList.add("fade-out");
  await new Promise((resolve) => setTimeout(resolve, 800));

  // Fully reset container for next scene
  container.classList.remove("fade-in", "fade-out");
  container.classList.add("hidden");
}

/**
 * Show final scene with button
 */
async function showFinalScene(text, buttonText, buttonAction) {
  const container = document.getElementById("scene-container");
  const background = document.getElementById("scene-background");
  const content = document.getElementById("scene-content");

  background.style.backgroundImage = "";
  background.style.backgroundColor = "#000000";

  // Create text element
  const textElement = document.createElement("div");
  textElement.className = "scene-text text-2xl md:text-4xl font-bold text-yellow-400 leading-relaxed mb-8";
  textElement.textContent = text;

  // Create button with delay using the continue button component
  const button = createContinueButton(7000, buttonText);
  button.onclick = () => {
    eval(buttonAction);
  };

  content.innerHTML = "";
  content.appendChild(textElement);
  content.appendChild(button);

  container.classList.remove("hidden", "fade-out");
  // Force reflow to ensure hidden is removed before fade-in
  void container.offsetHeight;
  container.classList.add("fade-in");
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Get equipment category from character class
 */
function getEquipmentCategory(className) {
  const categories = {
    Fighter: "warrior",
    Barbarian: "warrior",
    Paladin: "warrior",
    Cleric: "faithful",
    Monk: "faithful",
    Ranger: "wilderness",
    Druid: "wilderness",
    Wizard: "arcane",
    Sorcerer: "arcane",
    Warlock: "arcane",
    Rogue: "clever",
    Bard: "clever",
  };
  return categories[className] || "warrior";
}

/**
 * Get item image name (convert item name to image filename)
 */
function getItemImageName(itemName) {
  return itemName.toLowerCase().replace(/[',]/g, "").replace(/\s+/g, "-");
}

// ============================================================================
// INTRO SEQUENCE
// ============================================================================

/**
 * Main intro sequence - shows all story scenes
 */
async function startIntroSequence() {
  playerName = document.getElementById("character-name").value.trim();

  if (!playerName) {
    alert("Please enter your name");
    return;
  }

  // Hide name screen
  document.getElementById("name-screen").classList.add("hidden");

  // Load intro data
  const response = await fetch("/data/character/introductions.json");
  introData = await response.json();

  // 1. Scene 1 - Rainy Night
  await showScene({
    text: introData.scene1.text,
    image: introData.scene1.image,
    buttonDelay: introData.scene1.duration,
  });

  // 2. Scene 2 - Caretaker's Home
  await showScene({
    text: introData.scene2.text,
    image: introData.scene2.image,
    buttonDelay: introData.scene2.duration,
  });

  // 3. Final Words (black screen quote)
  console.log("🎬 Step 3: Final Words");
  await showScene({
    text: introData.final_words.text,
    isQuote: true,
    buttonDelay: introData.final_words.duration,
  });

  // 4. Background Intro - MOVED BEFORE letter intro
  console.log(
    "🎬 Step 4: Background Intro for:",
    generatedCharacter.background
  );
  const bgIntro = introData.background_intros.find((entry) =>
    entry.backgrounds.includes(generatedCharacter.background)
  );
  console.log("🎬 Background intro data:", bgIntro);
  if (bgIntro) {
    console.log("🎬 Showing background intro scene");
    await showScene({
      text: bgIntro.text,
      image: bgIntro.image,
      buttonDelay: bgIntro.duration,
    });
  } else {
    console.warn(
      "⚠️ No background intro found for:",
      generatedCharacter.background
    );
  }

  // 5. Letter Intro - MOVED AFTER background intro
  console.log("🎬 Step 5: Letter Intro");
  await showScene({
    text: introData.letter_intro.text,
    image: introData.letter_intro.image,
    buttonDelay: introData.letter_intro.duration,
  });

  // 6. Letter Reading (scene 4a)
  console.log("🎬 Step 6: Letter Reading for:", generatedCharacter.background);
  const bgLetter = introData.background_letters.find((entry) =>
    entry.backgrounds.includes(generatedCharacter.background)
  );
  console.log("🎬 Background letter data:", bgLetter);
  if (bgLetter) {
    console.log("🎬 Showing letter reading scene");
    await showScene({
      text: bgLetter.text,
      image: bgLetter.image,
      isLetter: true,
      buttonDelay: bgLetter.duration,
    });
  } else {
    console.warn(
      "⚠️ No background letter found for:",
      generatedCharacter.background
    );
  }

  // 7. Spell knowledge intro (if spellcaster)
  if (generatedCharacter.spells && generatedCharacter.spells.length > 0) {
    await showScene({
      text: introData.spell_knowledge.text,
      image: introData.spell_knowledge.image,
      buttonDelay: introData.spell_knowledge.duration,
    });
  }

  // 8. Show starting spells (if spellcaster)
  await showStartingSpells(generatedCharacter);

  // 9. Scene 5 - Equipment ready (AFTER spell scenes)
  await showScene({
    text: introData.scene5.text,
    image: introData.scene5.image,
    buttonDelay: introData.scene5.duration,
  });

  // 10. Equipment Intro (class-based) - narrative + quote (AFTER scene 5)
  const equipCategory = getEquipmentCategory(generatedCharacter.class);
  const equipIntro = introData.equipment_intros[equipCategory];
  if (equipIntro) {
    // Show narrative
    await showScene({
      text: equipIntro.text,
      image: equipIntro.image,
      buttonDelay: equipIntro.duration,
    });

    // Show quote if exists
    if (equipIntro.quote) {
      await showScene({
        text: equipIntro.quote,
        isQuote: true,
        buttonDelay: equipIntro.quote_duration,
      });
    }
  }

  // 11. Show items provided first
  await showGivenItemsScene(startingEquipment.inventory);

  // 12. Show equipment selection
  await startEquipmentSelection(startingEquipment);

  // Get selected equipment
  selectedEquipment = getSelectedEquipment();

  // Continue with remaining scenes
  await continueAfterEquipment();
}

/**
 * Continue after equipment selection
 */
async function continueAfterEquipment() {
  // Add delay to ensure equipment screen has fully faded out before starting next scene
  await new Promise((resolve) => setTimeout(resolve, 400));

  // 12. Scene 5a - Preparation (as dawn approaches) - AFTER equipment selection
  await showScene({
    text: introData.scene5a.text,
    image: introData.scene5a.image,
    buttonDelay: introData.scene5a.duration,
  });

  // 13. Scene 6 - Pack note (moved before pack selection)
  await showScene({
    text: introData.scene6.text,
    image: introData.scene6.image,
    buttonDelay: introData.scene6.duration,
  });

  // 14. Pack selection or display
  await handlePackSelection(startingEquipment);

  // 15. Departure
  await showScene({
    text: introData.departure.text,
    image: introData.departure.image,
    buttonDelay: introData.departure.duration,
  });

  // 16. Final Text + Begin Journey button
  await showFinalScene(
    introData.final_text.text,
    "Begin Journey",
    "startAdventure()"
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
    common: "#9ca3af", // grey
    uncommon: "#10b981", // green
    rare: "#3b82f6", // blue
    legendary: "#a855f7", // purple
    mythic: "#f97316", // orange
  };
  return rarityColors.common; // Default to common for now
}

/**
 * Create a styled item card for given items display
 */
async function createGivenItemCard(itemName, quantity) {
  const card = document.createElement("div");
  card.className = "relative overflow-hidden bg-gray-800 rounded-lg item-card";
  card.style.width = "110px";
  card.style.height = "110px";
  card.style.aspectRatio = "1/1";

  // Fetch item data to get the image path
  const itemData = await getItemById(itemName);

  // Item image (fills container with padding)
  const img = document.createElement("img");
  img.src = itemData?.image || `/res/img/items/${itemName}.png`;
  img.alt = itemName;
  img.className = "absolute inset-0 object-contain w-full h-full p-3";
  img.style.imageRendering = "pixelated";
  img.style.imageRendering = "-moz-crisp-edges";
  img.style.imageRendering = "crisp-edges";
  card.appendChild(img);

  // Rarity dot (top right)
  const rarityDot = document.createElement("div");
  rarityDot.className = "absolute top-1.5 right-1.5 z-10";
  rarityDot.style.width = "10px";
  rarityDot.style.height = "10px";
  rarityDot.style.borderRadius = "50%";
  rarityDot.style.backgroundColor = getRarityColor(itemName);
  rarityDot.style.border = "1px solid rgba(0,0,0,0.3)";
  card.appendChild(rarityDot);

  // Quantity text (top left, purple)
  if (quantity && quantity > 1) {
    const qtyText = document.createElement("div");
    qtyText.className = "absolute top-1 left-1.5 z-10 font-bold";
    qtyText.style.color = "#a855f7"; // purple
    qtyText.style.fontSize = "0.85rem";
    qtyText.style.textShadow =
      "0 0 3px rgba(0,0,0,0.8), 1px 1px 2px rgba(0,0,0,0.9)";
    qtyText.textContent = `×${quantity}`;
    card.appendChild(qtyText);
  }

  // Item name overlay at bottom
  const name = document.createElement("div");
  name.className =
    "absolute bottom-0 left-0 right-0 z-10 px-1 py-1 font-semibold text-center text-white item-name";
  name.style.fontSize = "0.65rem";
  name.style.lineHeight = "0.75rem";
  name.style.backgroundColor = "rgba(0,0,0,0.6)";
  name.style.textShadow = "0 1px 2px rgba(0,0,0,0.8)";
  name.textContent = itemName;
  card.appendChild(name);

  // Info button (bottom right, yellow ?, hover effect)
  const infoBtn = document.createElement("button");
  infoBtn.className =
    "info-btn absolute bottom-1 right-1.5 text-yellow-400 font-bold hover:text-green-400 transition-colors z-20";
  infoBtn.textContent = "?";
  infoBtn.style.fontSize = "16px";
  infoBtn.style.textShadow =
    "0 0 3px rgba(0,0,0,0.8), 1px 1px 2px rgba(0,0,0,0.9)";
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
  const backdrop = document.createElement("div");
  backdrop.id = "item-modal-backdrop";
  backdrop.className =
    "fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-70";
  backdrop.onclick = (e) => {
    if (e.target === backdrop) {
      hideItemModal();
    }
  };

  // Create modal content
  const modal = document.createElement("div");
  modal.className =
    "relative w-full max-w-md p-6 mx-4 bg-gray-800 border-2 border-yellow-400 rounded-lg";
  modal.onclick = (e) => e.stopPropagation();

  // Close button
  const closeBtn = document.createElement("button");
  closeBtn.className =
    "absolute w-8 h-8 font-bold text-white transition-colors bg-red-600 rounded-full top-2 right-2 hover:bg-red-700";
  closeBtn.textContent = "✕";
  closeBtn.onclick = hideItemModal;
  modal.appendChild(closeBtn);

  // Loading state
  modal.innerHTML += '<div class="text-center text-white">Loading...</div>';

  backdrop.appendChild(modal);
  document.body.appendChild(backdrop);

  // Fetch and display item stats
  const statsHTML = await getItemStats(itemName);

  modal.innerHTML = "";
  modal.appendChild(closeBtn);

  const contentDiv = document.createElement("div");
  contentDiv.className = "text-white";
  contentDiv.innerHTML = statsHTML;
  modal.appendChild(contentDiv);
}

/**
 * Hide item modal
 */
function hideItemModal() {
  const backdrop = document.getElementById("item-modal-backdrop");
  if (backdrop) {
    backdrop.remove();
  }
}

/**
 * Get item stats from API
 */
async function getItemStats(itemName) {
  try {
    const response = await fetch(
      `/api/items?name=${encodeURIComponent(itemName)}`
    );
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

    // Check if this is a focus item
    const isFocus =
      itemData.item_type === "Arcane Focus" ||
      itemData.item_type === "Druidic Focus" ||
      itemData.item_type === "Holy Symbol";

    // If it's a focus, show the component it provides prominently
    if (isFocus && props.provides) {
      statsHTML += `<div class="bg-purple-900 bg-opacity-40 border-2 border-purple-500 rounded-lg p-3 mb-3">`;
      statsHTML += `<div class="text-purple-300 text-xs font-semibold mb-2">✨ Provides Unlimited:</div>`;
      statsHTML += `<div class="flex items-center gap-2">`;
      statsHTML += `<img src="/res/img/items/${props.provides}.png" class="w-8 h-8 object-contain" style="image-rendering: pixelated; image-rendering: -moz-crisp-edges; image-rendering: crisp-edges;" onerror="this.style.display='none'">`;
      statsHTML += `<div class="text-white font-semibold">${props.provides
        .split("-")
        .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
        .join(" ")}</div>`;
      statsHTML += `</div>`;
      statsHTML += `</div>`;
    }

    // Stats section
    statsHTML += `<div class="space-y-1 mb-3">`;

    if (props.damage) {
      statsHTML += `<div class="text-gray-300 text-sm">⚔️ Damage: ${
        props.damage
      } ${props["damage-type"] || ""}</div>`;
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
      statsHTML += itemData.tags
        .map(
          (tag) =>
            `<span class="bg-gray-700 px-2 py-1 rounded text-xs text-gray-300">${tag}</span>`
        )
        .join("");
      statsHTML += `</div>`;
    }

    // Add full description
    if (itemData.description) {
      statsHTML += `<div class="text-gray-300 text-sm mt-3 leading-relaxed border-t border-gray-600 pt-3">${itemData.description}</div>`;
    }

    // Add notes for focuses
    if (isFocus && props.notes && props.notes.length > 0) {
      statsHTML += `<div class="text-purple-300 text-xs mt-3 leading-relaxed border-t border-purple-600 pt-3">`;
      props.notes.forEach((note) => {
        statsHTML += `<div class="mb-1">• ${note}</div>`;
      });
      statsHTML += `</div>`;
    }

    return statsHTML;
  } catch (error) {
    console.error("❌ Error fetching item:", itemName, error);
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

  // Filter out packs (they will be shown in the pack screen)
  const nonPackItems = givenItems.filter(
    (item) => !item.item.includes("-pack")
  );

  if (nonPackItems.length === 0) {
    return; // No non-pack items to show
  }

  const container = document.getElementById("scene-container");
  const background = document.getElementById("scene-background");
  const content = document.getElementById("scene-content");

  // Turn off background
  background.style.backgroundImage = "none";
  background.style.backgroundColor = "#111827";
  content.innerHTML = "";
  content.style.zIndex = "10";

  // Title
  const title = document.createElement("div");
  title.className = "mb-6 text-xl font-bold text-yellow-400 md:text-2xl";
  title.textContent = "Items Provided";
  content.appendChild(title);

  // Description
  const description = document.createElement("div");
  description.className = "mb-6 text-lg text-center";
  description.textContent =
    "In addition to your choices, you have been provided with these items:";
  content.appendChild(description);

  // Items container
  const itemsContainer = document.createElement("div");
  itemsContainer.className =
    "flex flex-wrap justify-center max-w-4xl gap-3 mx-auto mb-6";

  // Display each given item (excluding packs) with green selection border
  for (const givenItem of nonPackItems) {
    const itemCard = await createGivenItemCard(
      givenItem.item,
      givenItem.quantity
    );
    // Add green selection border
    itemCard.style.border = "3px solid #10b981";
    itemCard.style.boxShadow = "0 0 20px rgba(16, 185, 129, 0.6)";
    itemCard.style.backgroundColor = "rgba(16, 185, 129, 0.05)";
    itemsContainer.appendChild(itemCard);
  }

  content.appendChild(itemsContainer);

  // Continue button (no delay for equipment screens)
  const continueBtn = document.createElement("button");
  continueBtn.className = "pixel-continue-btn";
  continueBtn.textContent = "Continue →";
  content.appendChild(continueBtn);

  // Show container with fade-in
  container.classList.remove("hidden", "fade-out");
  // First ensure we're not in fade-in state
  container.classList.remove("fade-in");
  // Force reflow to ensure classes are applied
  void container.offsetHeight;
  // Now add fade-in to trigger transition
  container.classList.add("fade-in");

  // Wait for user to click Continue
  await waitForButtonClick(continueBtn);

  // Fade out the scene
  container.classList.remove("fade-in");
  container.classList.add("fade-out");
  await new Promise((resolve) => setTimeout(resolve, 800));

  // Clear content
  content.innerHTML = "";

  // Reset container for next scene (fade transitions)
  container.classList.remove("fade-in", "fade-out");
  container.classList.add("hidden");
}

/**
 * Show starting spells scene (if character is a spellcaster)
 */
async function showStartingSpells(character) {
  // Skip if not a spellcaster or no spells
  if (!character.spells || character.spells.length === 0) {
    console.log(
      `${character.class} has no starting spells, skipping spell scene`
    );
    return;
  }

  const container = document.getElementById("scene-container");
  const background = document.getElementById("scene-background");
  const content = document.getElementById("scene-content");

  // Turn off background
  background.style.backgroundImage = "none";
  background.style.backgroundColor = "#111827";
  content.innerHTML = "";
  content.style.zIndex = "10";

  // Make content fit viewport with flex layout
  content.style.display = "flex";
  content.style.flexDirection = "column";
  content.style.height = "100vh";
  content.style.padding = "2rem 1rem";
  content.style.boxSizing = "border-box";

  // Title - smaller
  const title = document.createElement("h2");
  title.className = "mb-2 text-2xl text-center";
  title.textContent = "Your Magical Arsenal";
  content.appendChild(title);

  // Description - smaller and more compact
  const description = document.createElement("div");
  description.className = "max-w-3xl mx-auto mb-3 text-sm text-center";
  description.innerHTML = `As a ${character.class}, you begin your journey with knowledge of these spells.`;
  content.appendChild(description);

  // Spell Slots Info - more compact
  if (character.spell_slots) {
    const slotsInfo = document.createElement("div");
    slotsInfo.className =
      "max-w-2xl p-2 mx-auto mb-3 bg-gray-800 border-2 border-purple-600 rounded-lg";

    const slotsTitle = document.createElement("div");
    slotsTitle.className = "mb-1 text-sm font-bold text-center text-purple-400";
    slotsTitle.textContent = "Spell Slots Available";
    slotsInfo.appendChild(slotsTitle);

    const slotsGrid = document.createElement("div");
    slotsGrid.className = "flex justify-center gap-3 text-xs";

    // spell_slots is now an object of arrays, not numbers
    Object.entries(character.spell_slots).forEach(([level, slots]) => {
      if (Array.isArray(slots) && slots.length > 0) {
        const slotDiv = document.createElement("div");
        slotDiv.className = "text-center";
        const levelName =
          level === "cantrips"
            ? "Cantrips"
            : `Level ${level.replace("level_", "")}`;
        slotDiv.innerHTML = `<div class="text-gray-400 text-xs">${levelName}</div><div class="text-lg font-bold text-purple-300">${slots.length}</div>`;
        slotsGrid.appendChild(slotDiv);
      }
    });

    slotsInfo.appendChild(slotsGrid);
    content.appendChild(slotsInfo);
  }

  // Spells container - flexible height to fill remaining space
  const spellsContainer = document.createElement("div");
  spellsContainer.className =
    "flex flex-wrap justify-center max-w-5xl gap-3 mx-auto mb-3 overflow-y-auto";
  spellsContainer.style.flex = "1";
  spellsContainer.style.minHeight = "0";

  // Load and display each spell
  console.log(`Loading ${character.spells.length} spells for display`);
  for (const spellId of character.spells) {
    console.log(`Creating card for spell: ${spellId}`);
    try {
      const spellCard = await createSpellCard(spellId);
      if (spellCard) {
        spellsContainer.appendChild(spellCard);
      }
    } catch (error) {
      console.error(`Failed to create card for spell ${spellId}:`, error);
    }
  }

  if (spellsContainer.children.length === 0) {
    console.warn("No spell cards were created");
  }

  content.appendChild(spellsContainer);

  // Continue button
  const continueBtn = document.createElement("button");
  continueBtn.className = "pixel-continue-btn";
  continueBtn.textContent = "Continue →";
  content.appendChild(continueBtn);

  // Show container with fade-in
  container.classList.remove("hidden", "fade-out");
  // First ensure we're not in fade-in state
  container.classList.remove("fade-in");
  // Force reflow to ensure classes are applied
  void container.offsetHeight;
  // Now add fade-in to trigger transition
  container.classList.add("fade-in");

  // Wait for user to click Continue
  await waitForButtonClick(continueBtn);

  // Fade out
  container.classList.remove("fade-in");
  container.classList.add("fade-out");
  await new Promise((resolve) => setTimeout(resolve, 800));

  // Clear content
  content.innerHTML = "";

  // Reset content styles
  content.style.display = "";
  content.style.flexDirection = "";
  content.style.height = "";
  content.style.padding = "";
  content.style.boxSizing = "";

  // Reset container
  container.classList.remove("fade-in", "fade-out");
  container.classList.add("hidden");
}

/**
 * Create a spell card with spell information
 */
async function createSpellCard(spellId) {
  // Fetch spell data from API
  console.log(`Fetching spell data for: ${spellId}`);
  const response = await fetch(`/api/spells/${spellId}`);

  if (!response.ok) {
    console.error(
      `Failed to fetch spell ${spellId}: ${response.status} ${response.statusText}`
    );
    return null;
  }

  const spell = await response.json();
  console.log(`Spell data received for ${spellId}:`, spell);

  // Outer container for card and tooltip
  const container = document.createElement("div");
  container.className = "relative";
  container.style.width = "120px";
  container.style.height = "150px";

  // Card element
  const card = document.createElement("div");
  card.className =
    "relative overflow-hidden transition-all bg-gray-900 border-2 border-purple-600 rounded-lg cursor-pointer hover:border-purple-400";
  card.style.width = "100%";
  card.style.height = "100%";
  card.onclick = () => {
    showSpellTooltip(spell, container);
  };

  // Spell school image
  const school = spell.school || "evocation";
  const img = document.createElement("img");
  img.src = `/res/img/spells/${school}.png`;
  img.className = "object-cover w-full h-full";
  img.style.imageRendering = "pixelated";
  img.style.imageRendering = "-moz-crisp-edges";
  img.style.imageRendering = "crisp-edges";
  img.onerror = () => {
    img.src = "/res/img/spells/evocation.png"; // Fallback
  };
  card.appendChild(img);

  // Overlay gradient for better text visibility
  const overlay = document.createElement("div");
  overlay.className =
    "absolute inset-0 bg-gradient-to-t from-black via-transparent to-transparent";
  card.appendChild(overlay);

  // Top left: Level indicator
  const levelBadge = document.createElement("div");
  levelBadge.className =
    "absolute px-2 py-1 text-xs font-bold text-purple-100 bg-purple-900 rounded top-1 left-1";
  levelBadge.textContent = spell.level === 0 ? "C" : `L${spell.level}`;
  card.appendChild(levelBadge);

  // Top right: Damage indicator (number only)
  if (spell.damage) {
    const damageType = (
      spell.properties?.damage_type ||
      spell.damage_type ||
      ""
    ).toLowerCase();
    const isHealing = damageType.includes("heal");

    const damageBadge = document.createElement("div");
    damageBadge.className = `absolute top-1 right-1 ${
      isHealing ? "bg-green-700 text-green-100" : "bg-red-700 text-red-100"
    } font-bold px-1 py-0.5 rounded`;
    damageBadge.style.fontSize = "9px";
    damageBadge.style.lineHeight = "1";
    damageBadge.textContent = spell.damage;
    card.appendChild(damageBadge);

    // Damage type emoji - position above spell name
    const damageEmojis = {
      fire: "🔥",
      cold: "❄️",
      lightning: "⚡",
      thunder: "⚡",
      acid: "🧪",
      poison: "☠️",
      necrotic: "☠️",
      radiant: "✨",
      psychic: "🧠",
      force: "🌀",
      slashing: "🗡️",
      piercing: "🏹",
      bludgeoning: "🔨",
      healing: "💚",
    };
    const emoji = damageEmojis[damageType] || (isHealing ? "💚" : "⚔️");

    const emojiDiv = document.createElement("div");
    emojiDiv.className = "absolute right-1";
    emojiDiv.style.bottom = "22px"; // Position above the name bar
    emojiDiv.style.fontSize = "18px";
    emojiDiv.style.textShadow =
      "0 0 4px rgba(0,0,0,0.9), 1px 1px 3px rgba(0,0,0,1)";
    emojiDiv.style.zIndex = "20";
    emojiDiv.textContent = emoji;
    card.appendChild(emojiDiv);
  }

  // Bottom left: Materials indicator with images
  const materialComponent =
    spell.properties?.material_component || spell.material_component;
  if (
    materialComponent &&
    materialComponent.required &&
    materialComponent.required.length > 0
  ) {
    // Container for material components
    const materialsContainer = document.createElement("div");
    materialsContainer.className =
      "absolute bottom-8 left-1 flex flex-col gap-0.5";

    // Show each material component
    materialComponent.required.forEach((mat) => {
      const matDiv = document.createElement("div");
      matDiv.className = "relative";
      matDiv.style.width = "24px";
      matDiv.style.height = "24px";

      // Material image
      const matImg = document.createElement("img");
      matImg.src = `/res/img/items/${mat.component}.png`;
      matImg.className = "object-contain w-full h-full";
      matImg.style.imageRendering = "pixelated";
      matImg.style.imageRendering = "-moz-crisp-edges";
      matImg.style.imageRendering = "crisp-edges";
      matImg.onerror = () => {
        // Fallback to a generic component icon
        matImg.style.display = "none";
      };

      // Quantity badge (top-right of material icon)
      const qtyBadge = document.createElement("div");
      qtyBadge.className =
        "absolute flex items-center justify-center font-bold text-white bg-purple-700 rounded-full -top-1 -right-1";
      qtyBadge.style.fontSize = "8px";
      qtyBadge.style.width = "12px";
      qtyBadge.style.height = "12px";
      qtyBadge.textContent = mat.quantity;

      matDiv.appendChild(matImg);
      matDiv.appendChild(qtyBadge);
      materialsContainer.appendChild(matDiv);
    });

    card.appendChild(materialsContainer);
  }

  // Bottom: Spell name
  const nameDiv = document.createElement("div");
  nameDiv.className =
    "absolute bottom-0 left-0 right-0 px-1 py-1 font-bold text-center text-purple-200 truncate bg-black bg-opacity-80";
  nameDiv.style.fontSize = "9px";
  nameDiv.style.lineHeight = "1.1";
  nameDiv.textContent = spell.name || spellId;
  nameDiv.title = spell.name || spellId;
  card.appendChild(nameDiv);

  container.appendChild(card);
  return container;
}

/**
 * Show spell tooltip with detailed information
 */
function showSpellTooltip(spell, containerElement) {
  // Remove any existing tooltips
  const existing = document.querySelector(".spell-tooltip");
  if (existing) {
    existing.remove();
  }

  // Create tooltip
  const tooltip = document.createElement("div");
  tooltip.className =
    "fixed z-50 p-4 bg-gray-800 border-2 border-purple-500 rounded-lg shadow-xl spell-tooltip";
  tooltip.style.maxWidth = "400px";
  tooltip.style.minWidth = "300px";

  // Position near the card
  const rect = containerElement.getBoundingClientRect();
  tooltip.style.left = `${rect.right + 10}px`;
  tooltip.style.top = `${rect.top}px`;

  // Spell name and level
  const header = document.createElement("div");
  header.className =
    "flex items-start justify-between pb-2 mb-2 border-b border-purple-600";

  const nameDiv = document.createElement("div");
  nameDiv.className = "text-xl font-bold text-purple-300";
  nameDiv.textContent = spell.name;

  const levelDiv = document.createElement("div");
  levelDiv.className = "text-sm text-purple-400";
  levelDiv.textContent = spell.level === 0 ? "Cantrip" : `Level ${spell.level}`;

  header.appendChild(nameDiv);
  header.appendChild(levelDiv);
  tooltip.appendChild(header);

  // School and casting time
  const meta = document.createElement("div");
  meta.className = "mb-3 text-sm text-gray-400";
  const castingTime =
    spell.properties?.casting_time || spell.casting_time || "Action";
  const range = spell.properties?.range || spell.range || "Touch";
  const duration =
    spell.properties?.duration || spell.duration || "Instantaneous";
  meta.innerHTML = `<div><strong>School:</strong> ${spell.school}</div>
                    <div><strong>Casting Time:</strong> ${castingTime}</div>
                    <div><strong>Range:</strong> ${range}</div>
                    <div><strong>Duration:</strong> ${duration}</div>`;
  tooltip.appendChild(meta);

  // Description
  const desc = document.createElement("div");
  desc.className = "pt-2 mb-3 text-sm text-gray-300 border-t border-gray-700";
  desc.textContent = spell.description;
  tooltip.appendChild(desc);

  // Damage and effects
  if (spell.damage || spell.properties?.heal) {
    const effects = document.createElement("div");
    effects.className = "flex flex-wrap gap-2 mb-2 text-xs";

    if (spell.damage) {
      const dmg = document.createElement("span");
      dmg.className = "px-2 py-1 text-red-200 bg-red-900 rounded";
      const damageType =
        spell.properties?.damage_type || spell.damage_type || "damage";
      dmg.textContent = `${spell.damage} ${damageType}`;
      effects.appendChild(dmg);
    }

    if (spell.properties?.heal) {
      const heal = document.createElement("span");
      heal.className = "px-2 py-1 text-green-200 bg-green-900 rounded";
      heal.textContent = `Heal: ${spell.properties.heal}`;
      effects.appendChild(heal);
    }

    tooltip.appendChild(effects);
  }

  // Materials
  const materialComponent =
    spell.properties?.material_component || spell.material_component;
  if (materialComponent && materialComponent.required) {
    const materials = document.createElement("div");
    materials.className =
      "p-2 mb-2 text-xs text-yellow-300 bg-yellow-900 rounded bg-opacity-30";
    materials.innerHTML = `<strong>Materials Required:</strong><br>${materialComponent.required
      .map((m) => `${m.component} (${m.quantity})`)
      .join(", ")}`;
    tooltip.appendChild(materials);
  }

  // Close button
  const closeBtn = document.createElement("button");
  closeBtn.className =
    "w-full px-4 py-2 mt-2 font-bold text-white bg-purple-700 rounded hover:bg-purple-600";
  closeBtn.textContent = "Close";
  closeBtn.onclick = () => tooltip.remove();
  tooltip.appendChild(closeBtn);

  document.body.appendChild(tooltip);

  // Close on click outside
  const closeOnClickOutside = (e) => {
    if (!tooltip.contains(e.target) && !containerElement.contains(e.target)) {
      tooltip.remove();
      document.removeEventListener("click", closeOnClickOutside);
    }
  };
  setTimeout(() => {
    document.addEventListener("click", closeOnClickOutside);
  }, 100);

  // Adjust position if tooltip goes off screen
  setTimeout(() => {
    const tooltipRect = tooltip.getBoundingClientRect();
    if (tooltipRect.right > window.innerWidth) {
      tooltip.style.left = `${rect.left - tooltipRect.width - 10}px`;
    }
    if (tooltipRect.bottom > window.innerHeight) {
      tooltip.style.top = `${window.innerHeight - tooltipRect.height - 10}px`;
    }
  }, 0);
}

// ============================================================================
// EQUIPMENT SELECTION
// ============================================================================

/**
 * Show equipment selection screen
 */
function showEquipmentSelection() {
  const equipmentSection = document.getElementById("equipment-selection");
  const equipmentChoices = document.getElementById("equipment-choices");
  const startingInventory = document.getElementById("starting-inventory");
  const continueBtn = document.getElementById("continue-adventure-btn");

  equipmentSection.className =
    "fixed inset-0 flex flex-col items-center justify-center h-full p-8 overflow-y-auto bg-black";

  // Show starting inventory
  startingInventory.innerHTML = "";
  startingEquipment.inventory.forEach((item) => {
    const itemDiv = document.createElement("div");
    itemDiv.className = "p-2 text-sm bg-gray-700 rounded";

    const itemName = document.createElement("span");
    itemName.className = "text-white";
    itemName.textContent = item.item;
    itemDiv.appendChild(itemName);

    if (item.quantity > 1) {
      const quantity = document.createElement("span");
      quantity.className = "text-gray-400";
      quantity.textContent = ` (${item.quantity})`;
      itemDiv.appendChild(quantity);
    }

    startingInventory.appendChild(itemDiv);
  });

  // Show equipment choices
  equipmentChoices.innerHTML = "";

  if (startingEquipment.choices && startingEquipment.choices.length > 0) {
    startingEquipment.choices.forEach((choice, index) => {
      const choiceGroup = document.createElement("div");
      choiceGroup.className = "p-6 bg-gray-700 rounded-lg";

      const choiceTitle = document.createElement("h3");
      choiceTitle.className = "mb-4 text-xl font-semibold text-yellow-400";
      choiceTitle.textContent = `Choice ${index + 1}: ${
        choice.description || "Choose your equipment"
      }`;
      choiceGroup.appendChild(choiceTitle);

      // Handle complex weapon choices
      if (choice.type === "complex_weapon_choice") {
        renderComplexWeaponChoice(choiceGroup, choice, index);
      } else {
        // Regular options
        choice.options.forEach((option, optionIndex) => {
          const optionDiv = document.createElement("div");
          optionDiv.className =
            "p-4 mb-2 transition-colors bg-gray-600 rounded cursor-pointer hover:bg-gray-500";
          optionDiv.onclick = () => selectEquipment(index, option);

          const optionName = document.createElement("div");
          optionName.className = "font-semibold text-white";
          optionName.textContent = option.name;

          if (option.bundle && option.bundle.length > 0) {
            const bundleList = document.createElement("div");
            bundleList.className = "mt-1 text-sm text-gray-300";
            bundleList.textContent = option.bundle
              .map((item) =>
                item[1] > 1 ? `${item[0]} (${item[1]})` : item[0]
              )
              .join(", ");
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
    continueBtn.style.display = "block";
  }
}

/**
 * Handle complex weapon choices (martial/simple weapon selection)
 */
function renderComplexWeaponChoice(container, choice, choiceIndex) {
  const weaponData = {};

  choice.slots.forEach((slot, slotIndex) => {
    const slotDiv = document.createElement("div");
    slotDiv.className = "p-3 mb-4 bg-gray-800 rounded";

    const slotTitle = document.createElement("div");
    slotTitle.className = "mb-2 font-medium text-green-400";
    slotTitle.textContent = `Slot ${slotIndex + 1}`;
    slotDiv.appendChild(slotTitle);

    if (slot.type === "choice") {
      const select = document.createElement("select");
      select.className = "w-full px-3 py-2 text-white bg-gray-700 rounded";
      select.id = `weapon-slot-${choiceIndex}-${slotIndex}`;

      const defaultOption = document.createElement("option");
      defaultOption.value = "";
      defaultOption.textContent = "-- Select Weapon --";
      select.appendChild(defaultOption);

      slot.options.forEach((weapon) => {
        const option = document.createElement("option");
        option.value = weapon;
        option.textContent = weapon;
        select.appendChild(option);
      });

      slotDiv.appendChild(select);
      weaponData[slotIndex] = { type: "choice", element: select };
    } else if (slot.type === "fixed_item") {
      const fixedText = document.createElement("div");
      fixedText.className = "text-gray-300";
      fixedText.textContent = slot.item;
      slotDiv.appendChild(fixedText);
      weaponData[slotIndex] = { type: "fixed", item: slot.item };
    }

    container.appendChild(slotDiv);
  });

  // Store complex choice data
  selectedEquipment[choiceIndex] = {
    isComplexChoice: true,
    slots: weaponData,
    getSelectedWeapons: function () {
      const weapons = [];
      Object.values(this.slots).forEach((slot) => {
        if (slot.type === "choice" && slot.element.value) {
          weapons.push([slot.element.value, 1]);
        } else if (slot.type === "fixed") {
          weapons.push([slot.item, 1]);
        }
      });
      return weapons;
    },
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
      isBundle: true,
    };
  } else {
    selectedEquipment[choiceIndex] = {
      item: option.item,
      quantity: option.quantity,
      name: option.name,
    };
  }

  // Visual feedback
  const choiceGroups = document.querySelectorAll("#equipment-choices > div");
  if (choiceGroups[choiceIndex]) {
    const options = choiceGroups[choiceIndex].querySelectorAll(
      ".bg-gray-600, .bg-gray-500, .bg-green-700"
    );
    options.forEach((opt) => {
      opt.classList.remove("bg-green-700");
      opt.classList.add("bg-gray-600");
    });

    event.currentTarget.classList.remove("bg-gray-600");
    event.currentTarget.classList.add("bg-green-700");
  }

  checkEquipmentComplete();
}

/**
 * Check if all equipment choices are complete
 */
function checkEquipmentComplete() {
  const continueBtn = document.getElementById("continue-adventure-btn");
  const totalChoices = startingEquipment.choices
    ? startingEquipment.choices.length
    : 0;
  const selectedCount = Object.keys(selectedEquipment).length;

  if (selectedCount >= totalChoices) {
    continueBtn.style.display = "block";
  }
}

// ============================================================================
// SAVE AND START ADVENTURE
// ============================================================================

// Cache for item data from database
let itemsCache = null;

/**
 * Load all items from database once
 */
async function loadItemsFromDatabase() {
  if (itemsCache) {
    return itemsCache;
  }

  try {
    const response = await fetch("/api/items");
    if (response.ok) {
      itemsCache = await response.json();
      console.log(`📦 Loaded ${itemsCache.length} items from database`);
      return itemsCache;
    }
  } catch (error) {
    console.warn("Could not load items from database:", error);
  }
  return [];
}

/**
 * Get item data from database cache by ID
 */
async function getItemById(itemId) {
  try {
    const items = await loadItemsFromDatabase();

    // Find item by ID (exact match)
    const item = items.find((i) => i.id === itemId);

    if (item) {
      // Convert database format to expected frontend format
      return {
        id: item.id,
        name: item.name,
        description: item.description,
        type: item.item_type,
        tags: item.tags || [],
        rarity: item.rarity,
        gear_slot: item.properties?.gear_slot,
        slots: item.properties?.slots,
        contents: item.properties?.contents,
        stack: item.properties?.stack,
        ...item.properties, // Spread all other properties
      };
    } else {
      console.warn(`❌ Item ID "${itemId}" not found in database`);
    }
  } catch (error) {
    console.warn(`Could not load item data for ID: ${itemId}`, error);
  }
  return null;
}

/**
 * Add items with proper stacking logic
 */
async function addItemWithStacking(allItems, itemId, quantity) {
  // Get item data to check stack limit
  const itemData = await getItemById(itemId);
  const stackLimit = itemData ? parseInt(itemData.stack) || 1 : 1;

  let remainingQuantity = quantity;

  // Try to add to existing stacks first
  for (let existingItem of allItems) {
    if (existingItem.item === itemId && existingItem.quantity < stackLimit) {
      const canAdd = Math.min(
        remainingQuantity,
        stackLimit - existingItem.quantity
      );
      existingItem.quantity += canAdd;
      remainingQuantity -= canAdd;

      if (remainingQuantity <= 0) break;
    }
  }

  // Create new stacks for remaining quantity
  while (remainingQuantity > 0) {
    const stackSize = Math.min(remainingQuantity, stackLimit);
    allItems.push({ item: itemId, quantity: stackSize });
    remainingQuantity -= stackSize;
  }
}

/**
 * Unpack packs and return contents
 */
async function unpackItem(packId) {
  try {
    console.log(`🎒 Attempting to unpack: "${packId}"`);
    const packData = await getItemById(packId);
    if (packData) {
      console.log(`🎒 Loaded pack data for ${packId}:`, packData);
      if (packData.contents) {
        // Parse contents string if it's a string, or use directly if array
        const contents =
          typeof packData.contents === "string"
            ? JSON.parse(packData.contents)
            : packData.contents;

        console.log(`🎒 Pack contents:`, contents);
        // Convert to proper slot format
        const slots = [];
        let slotIndex = 0;
        contents.forEach((item) => {
          const itemId = item[0];
          const itemName = itemId.toLowerCase();
          // Don't include the backpack itself or any pack items
          if (
            itemName !== "backpack" &&
            !itemName.includes("-pack") &&
            !itemName.includes("pack-")
          ) {
            slots.push({
              slot: slotIndex,
              item: itemId,
              quantity: item[1],
            });
            slotIndex++;
          } else {
            console.log(`🎒 Excluding from pack contents: ${itemId}`);
          }
        });

        // Fill remaining slots with null
        const totalSlots = 20; // Backpack has 20 slots
        while (slots.length < totalSlots) {
          slots.push({
            slot: slots.length,
            item: null,
            quantity: 0,
          });
        }

        console.log(
          `🎒 Successfully unpacked ${packId} into ${
            slots.filter((s) => s.item).length
          } items`
        );
        return slots;
      } else {
        console.warn(`🎒 Pack ${packId} has no contents field`);
      }
    } else {
      console.warn(`🎒 Pack data not found for: ${packId}`);
    }
  } catch (error) {
    console.warn(`🎒 Could not unpack: ${packId}`, error);
  }
  return null;
}

/**
 * Add items to general slots or bag
 */
function addToGeneralSlotOrBag(inventory, item, slotIndex, itemData = null) {
  // Check if item is a container (has 'container' tag)
  const isContainer =
    itemData?.tags?.includes("container") ||
    item.item.toLowerCase().includes("pouch") ||
    item.item.toLowerCase().includes("pack");

  // Containers should NOT go in the bag (no containers inside containers)
  if (isContainer) {
    console.log(`📦 ${item.item} is a container, placing in general slot`);
    if (slotIndex < 4 && inventory.general_slots[slotIndex].item === null) {
      inventory.general_slots[slotIndex] = {
        slot: slotIndex,
        item: item.item,
        quantity: item.quantity,
      };
      return;
    }
  } else {
    // Try to add to backpack first if it exists (only non-containers)
    if (
      inventory.gear_slots.bag.item !== null &&
      inventory.gear_slots.bag.contents
    ) {
      const emptyBagSlot = inventory.gear_slots.bag.contents.find(
        (slot) => slot.item === null
      );
      if (emptyBagSlot) {
        console.log(
          `📦 Adding ${item.item} to backpack slot ${emptyBagSlot.slot}`
        );
        emptyBagSlot.item = item.item;
        emptyBagSlot.quantity = item.quantity;
        return;
      }
    }
  }

  // If backpack is full or item is a container, use general slots
  if (slotIndex < 4 && inventory.general_slots[slotIndex].item === null) {
    console.log(`📦 Adding ${item.item} to general slot ${slotIndex}`);
    inventory.general_slots[slotIndex] = {
      slot: slotIndex,
      item: item.item,
      quantity: item.quantity,
    };
    return;
  }

  console.warn(`⚠️ Could not place item ${item.item} - inventory full`);
}

/**
 * Dynamic inventory creation with proper equipment placement
 */
async function createInventoryFromItems(allItems) {
  // Initialize empty inventory structure
  const inventory = {
    general_slots: [
      { slot: 0, item: null, quantity: 0 },
      { slot: 1, item: null, quantity: 0 },
      { slot: 2, item: null, quantity: 0 },
      { slot: 3, item: null, quantity: 0 },
    ],
    gear_slots: {
      bag: { item: null, quantity: 0 },
      left_arm: { item: null, quantity: 0 },
      right_arm: { item: null, quantity: 0 },
      armor: { item: null, quantity: 0 },
      necklace: { item: null, quantity: 0 },
      ring: { item: null, quantity: 0 },
      ammunition: { item: null, quantity: 0 },
      clothes: { item: null, quantity: 0 },
    },
  };

  let remainingItems = [...allItems];
  let currentGeneralSlot = 0;
  let twoHandedEquipped = false;

  // 1. First pass - Handle packs (automatically unpack to bag slot)
  for (let i = remainingItems.length - 1; i >= 0; i--) {
    const item = remainingItems[i];
    const itemName = item.item.toLowerCase();

    if (itemName.includes("pack")) {
      console.log(`🎒 Found pack: ${item.item}`);
      // This is a pack - unpack it to bag slot
      const packContents = await unpackItem(item.item);
      if (packContents) {
        console.log(`🎒 Successfully unpacked ${item.item}:`, packContents);
        // Equip the pack itself as a backpack to bag slot (the pack becomes the backpack)
        inventory.gear_slots.bag = {
          item: "backpack", // All packs become backpacks when equipped
          quantity: 1,
          contents: packContents,
        };
        remainingItems.splice(i, 1);
      }
    }
  }

  // 2. Second pass - Handle all equipment items based on gear_slot
  for (let i = remainingItems.length - 1; i >= 0; i--) {
    const item = remainingItems[i];
    const itemData = await getItemById(item.item);

    console.log(`🔍 Checking item: ${item.item}`, {
      hasData: !!itemData,
      tags: itemData?.tags,
      gearSlot: itemData?.gear_slot,
    });

    // Check if item has equipment tag and gear_slot
    if (
      itemData &&
      itemData.tags &&
      itemData.tags.includes("equipment") &&
      itemData.gear_slot
    ) {
      const gearSlot = itemData.gear_slot;
      console.log(`⚔️ Found equipment: ${item.item} → ${gearSlot}`);

      // Handle different gear slots
      if (gearSlot === "armor" && inventory.gear_slots.armor.item === null) {
        console.log(`Equipping armor: ${item.item}`);
        inventory.gear_slots.armor = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "necklace" &&
        inventory.gear_slots.necklace.item === null
      ) {
        console.log(`Equipping necklace: ${item.item}`);
        inventory.gear_slots.necklace = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "ring" &&
        inventory.gear_slots.ring.item === null
      ) {
        console.log(`Equipping ring: ${item.item}`);
        inventory.gear_slots.ring = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "ammunition" &&
        inventory.gear_slots.ammunition.item === null
      ) {
        console.log(`Equipping ammunition: ${item.item}`);
        inventory.gear_slots.ammunition = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (
        gearSlot === "clothes" &&
        inventory.gear_slots.clothes.item === null
      ) {
        console.log(`Equipping clothes: ${item.item}`);
        inventory.gear_slots.clothes = {
          item: item.item,
          quantity: item.quantity,
        };
        remainingItems.splice(i, 1);
      } else if (gearSlot === "hands") {
        // Handle weapons - check if two-handed
        const isTwoHanded =
          itemData.tags && itemData.tags.includes("two-handed");

        if (isTwoHanded) {
          if (inventory.gear_slots.right_arm.item === null) {
            console.log(`Equipping two-handed weapon: ${item.item}`);
            inventory.gear_slots.right_arm = {
              item: item.item,
              quantity: item.quantity,
            };
            twoHandedEquipped = true;
            remainingItems.splice(i, 1);
          }
        } else {
          // One-handed weapon
          if (inventory.gear_slots.right_arm.item === null) {
            console.log(`Equipping weapon in right hand: ${item.item}`);
            inventory.gear_slots.right_arm = {
              item: item.item,
              quantity: item.quantity,
            };
            remainingItems.splice(i, 1);
          } else if (
            inventory.gear_slots.left_arm.item === null &&
            !twoHandedEquipped
          ) {
            console.log(`Equipping weapon in left hand: ${item.item}`);
            inventory.gear_slots.left_arm = {
              item: item.item,
              quantity: item.quantity,
            };
            remainingItems.splice(i, 1);
          }
        }
      }
    }
  }

  // 3. Put remaining items in general slots or bag
  for (let item of remainingItems) {
    const itemData = await getItemById(item.item);
    addToGeneralSlotOrBag(inventory, item, currentGeneralSlot++, itemData);
  }

  return inventory;
}

/**
 * Start the adventure - save character and redirect
 */
async function startAdventure() {
  try {
    console.log("🎮 Starting adventure...");
    console.log("generatedCharacter:", generatedCharacter);
    console.log("startingEquipment:", startingEquipment);
    console.log("selectedEquipment:", selectedEquipment);
    console.log("playerName:", playerName);

    // Validate we have the required data
    if (!generatedCharacter) {
      throw new Error("No character data available");
    }
    if (!startingEquipment) {
      throw new Error("No starting equipment data available");
    }

    // Get the final character name (either edited or generated)
    const finalName = playerName || generatedCharacter.name;

    // Collect all items (starting + selected)
    let allItems = [];

    // Add starting equipment with proper stacking
    for (const startingItem of startingEquipment.inventory) {
      await addItemWithStacking(
        allItems,
        startingItem.item,
        startingItem.quantity
      );
    }

    // Add selected equipment to items list with proper stacking
    for (const [key, option] of Object.entries(selectedEquipment)) {
      // Skip pack - it will be handled separately in inventory creation
      if (key === "pack") {
        // Pack is just a string ID, add it as an item
        await addItemWithStacking(allItems, option, 1);
        continue;
      }

      if (option.isComplexChoice) {
        // Handle complex weapon choices
        const selectedWeapons = option.weapons || [];
        for (const weapon of selectedWeapons) {
          await addItemWithStacking(allItems, weapon[0], weapon[1]);
        }
      } else if (option.isBundle) {
        for (const bundleItem of option.bundle) {
          await addItemWithStacking(allItems, bundleItem[0], bundleItem[1]);
        }
      } else if (option.item) {
        await addItemWithStacking(allItems, option.item, option.quantity || 1);
      }
    }

    // Create proper inventory structure with dynamic equipment placement
    const inventory = await createInventoryFromItems(allItems);

    // Prepare final character data (remove choices and other unnecessary fields)
    const finalCharacter = {
      name: finalName,
      race: generatedCharacter.race,
      class: generatedCharacter.class,
      background: generatedCharacter.background,
      alignment: generatedCharacter.alignment,
      experience: generatedCharacter.experience,
      stats: generatedCharacter.stats,
      hp: generatedCharacter.hp,
      max_hp: generatedCharacter.max_hp,
      mana: generatedCharacter.mana,
      max_mana: generatedCharacter.max_mana,
      fatigue: generatedCharacter.fatigue,
      gold: generatedCharacter.gold,
      inventory: inventory,
      known_spells: generatedCharacter.spells || [],
      spell_slots: generatedCharacter.spell_slots || {},
    };

    console.log("Final character:", finalCharacter);

    // Create new save file
    const session = window.sessionManager.getSession();
    const saveData = {
      d: finalName,
      created_at: new Date().toISOString(),
      race: finalCharacter.race,
      class: finalCharacter.class,
      background: finalCharacter.background,
      alignment: finalCharacter.alignment,
      experience: finalCharacter.experience,
      hp: finalCharacter.hp,
      max_hp: finalCharacter.max_hp,
      mana: finalCharacter.mana,
      max_mana: finalCharacter.max_mana,
      fatigue: finalCharacter.fatigue,
      gold: finalCharacter.gold,
      stats: finalCharacter.stats,
      location: generatedCharacter.city || "kingdom",
      inventory: finalCharacter.inventory,
      known_spells: finalCharacter.known_spells,
      spell_slots: finalCharacter.spell_slots,
      locations_discovered: [generatedCharacter.city || "kingdom"],
      music_tracks_unlocked: ["character-creation", "kingdom-theme"],
    };

    const response = await fetch(`/api/saves/${session.npub}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(saveData),
    });

    if (response.ok) {
      const result = await response.json();
      console.log("✅ Save created successfully:", result);

      // Play departure music and transition
      const introMusic = document.getElementById("intro-music");
      const gameMusic = document.getElementById("game-music");
      introMusic.pause();
      gameMusic.volume = 0.3;
      gameMusic
        .play()
        .catch((e) => console.log("Game music autoplay blocked:", e));

      // Redirect to game using the save_id from backend response
      setTimeout(() => {
        window.location.href = "/game?save=" + result.save_id;
      }, 3000);
    } else {
      throw new Error("Failed to create save file");
    }
  } catch (error) {
    console.error("❌ Failed to start adventure:", error);
    alert("Failed to start adventure: " + error.message);
  }
}

// ============================================================================
// INITIALIZATION
// ============================================================================

/**
 * Request fullscreen mode (mobile optimization)
 */
function requestFullscreen() {
  const elem = document.documentElement;

  if (elem.requestFullscreen) {
    elem.requestFullscreen().catch(err => {
      console.log("Fullscreen request failed:", err);
    });
  } else if (elem.webkitRequestFullscreen) { // Safari
    elem.webkitRequestFullscreen();
  } else if (elem.mozRequestFullScreen) { // Firefox
    elem.mozRequestFullScreen();
  } else if (elem.msRequestFullscreen) { // IE11
    elem.msRequestFullscreen();
  }
}

/**
 * Check if device is mobile
 */
function isMobileDevice() {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) ||
         (window.innerWidth <= 768);
}

document.addEventListener("DOMContentLoaded", async function () {
  console.log("🎮 Initializing game intro...");

  if (!window.sessionManager) {
    console.error("❌ Session manager not available");
    window.location.href = "/";
    return;
  }

  try {
    await window.sessionManager.init();

    if (!window.sessionManager.isAuthenticated()) {
      console.log("❌ Not authenticated, redirecting");
      window.location.href = "/";
      return;
    }

    const session = window.sessionManager.getSession();
    console.log("✅ Authenticated:", session.npub);

    // Initialize character generator
    console.log("🎲 Initializing character generator...");
    await window.characterGenerator.initialize();

    // Generate character
    console.log("🎲 Generating character...");
    generatedCharacter = await window.characterGenerator.generateCharacter(
      session.npub
    );
    console.log("✅ Character generated:", generatedCharacter);

    // Generate starting equipment
    startingEquipment =
      window.characterGenerator.generateStartingEquipment(generatedCharacter);
    console.log("✅ Starting equipment loaded:", startingEquipment);

    // Set up music playlist
    setupMusicPlaylist();

    // Start music when name input is interacted with
    const nameInput = document.getElementById("character-name");
    nameInput.addEventListener(
      "focus",
      () => {
        startMusic();
      },
      { once: true }
    );

    // Request fullscreen on mobile when user interacts with name input
    if (isMobileDevice()) {
      nameInput.addEventListener(
        "click",
        () => {
          requestFullscreen();
        },
        { once: true }
      );

      // Also try on Begin Your Story button
      const continueBtn = document.getElementById("continue-btn");
      continueBtn.addEventListener("click", () => {
        requestFullscreen();
      });
    }
  } catch (error) {
    console.error("❌ Initialization failed:", error);
    alert("Failed to initialize: " + error.message);
  }
});
