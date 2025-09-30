# Game Intro Scene Flow - Frontend Implementation Notes

## Complete Scene Flow Order

### 1. Name Input Screen
- Typewriter effect: "What is your name?"
- Input field and "Begin Your Story" button
- **Transition**: Fade out → Start scene sequence

### 2. Scene 1 - Rainy Night
- **Image**: `scene1.png` (fullscreen background)
- **Text**: Rainy night introduction
- **Animation**: Wipe-down fade-in (text appears with opacity + translateY)
- **Duration**: ~3-4 seconds
- **Transition**: Fade out (both image and text)

### 3. Scene 2 - Old Caretaker's Home
- **Image**: `scene2.png` (fullscreen background)
- **Text**: Caretaker description + "Their final words still echo..."
- **Animation**: Wipe-down fade-in
- **Duration**: ~3-4 seconds
- **Transition**: Fade out → Fade to black

### 4. Final Words (BLACK SCREEN)
- **Image**: None (black background)
- **Text**: "Your destiny awaits beyond these village borders. You have a hero's heart - don't let it remain hidden here."
- **Styling**: Large font (text-2xl or text-3xl), centered, prominent color (yellow-400)
- **Animation**: Fade-in
- **Duration**: ~4-5 seconds
- **Transition**: Fade out

### 5. Scene 3 - Background-Specific Scene
- **Image**: `scene3.png` (placeholder, will be background-specific later)
- **Text**: Character's background story (e.g., Acolyte faith scene, Criminal shadows scene, etc.)
- **Data Source**: `background_intros[character.background].text`
- **Animation**: Wipe-down fade-in
- **Duration**: ~3-4 seconds
- **Transition**: Fade out

### 6. Letter Intro
- **Image**: `scene4.png`
- **Text**: "Before passing, they left you a letter..."
- **Data Source**: `letter_intro.text`
- **Animation**: Wipe-down fade-in
- **Duration**: ~3 seconds
- **Transition**: Fade out

### 7. Letter Reading (Scene 4a)
- **Image**: `scene4letter.png` (letter background)
- **Text**: Background-specific letter content
- **Data Source**: `background_letters[character.background].text`
- **Styling**: Special letter styling - serif font, parchment-colored overlay, italic
- **Animation**: Fade-in (slower, to emphasize importance)
- **Duration**: ~5-6 seconds (longer for reading)
- **Transition**: Fade out

### 8. Equipment Intro - Class-Based
- **Image**: `scene4.png`
- **Text**: Class-specific equipment scene (Warrior/Faithful/Wilderness/Arcane/Clever)
- **Data Source**: Determine class category from `character.class`, lookup in `equipment_intros[category]`
  - Fighter/Barbarian/Paladin → "warrior"
  - Cleric/Paladin/Monk → "faithful"
  - Ranger/Druid/Barbarian → "wilderness"
  - Wizard/Sorcerer/Warlock → "arcane"
  - Rogue/Bard/Ranger → "clever"
- **Text includes quote**: Equipment description + caretaker's quote (both in same text field)
- **Animation**: Wipe-down fade-in
- **Duration**: ~4-5 seconds
- **Transition**: Fade out

### 9. Scene 5 - Equipment Ready
- **Image**: `scene4.png`
- **Text**: "Whatever your calling, the equipment before you seems chosen..."
- **Data Source**: `scene5.text`
- **Animation**: Wipe-down fade-in
- **Duration**: ~3 seconds
- **Transition**: Fade out → Equipment selection interface

### 10. EQUIPMENT SELECTION (Interactive)
- **Display**: Individual item selection scenes
- **Note**: One scene per item choice/given item (as specified in intro.txt line 85)
- **Current Implementation**: Bulk selection screen (needs to be changed)
- **Styling**: Keep existing equipment selection UI but show items one at a time
- **Transition**: After all items selected → Fade out

### 11. Scene 5a (BLACK SCREEN)
- **Image**: None (black background)
- **Text**: "As dawn approaches, you prepare to leave the only home you've known. The items you select will help define the hero you become."
- **Data Source**: `scene5a.text`
- **Styling**: Large font, centered
- **Animation**: Fade-in
- **Duration**: ~4 seconds
- **Transition**: Fade out

### 12. Scene 6 - Pack Selection
- **Image**: `scene2.png`
- **Text**: Pack description + "The weight you carry shapes the journey. Choose wisely."
- **Data Source**: `scene6.text` (includes quote)
- **Animation**: Wipe-down fade-in
- **Duration**: ~3-4 seconds
- **Transition**: Fade out → Pack selection interface

### 13. PACK SELECTION/GIVEN (Interactive)
- **Display**: Show pack choice OR display given pack
- **Logic**: If character has pack choice → show options; else → show assigned pack
- **Transition**: After pack selected/acknowledged → Fade out

### 14. Departure Scene
- **Image**: `scene3.png` (placeholder for village gate scene)
- **Text**: "With the letter carefully folded and tucked away..."
- **Data Source**: `departure.text`
- **Animation**: Wipe-down fade-in
- **Duration**: ~3-4 seconds
- **Transition**: Fade out → Fade to black

### 15. Final Text (BLACK SCREEN) + Button
- **Image**: None (black background)
- **Text**: "Where fate will lead, only time will tell. But one thing is certain - you won't find your destiny by staying here."
- **Data Source**: `final_text.text`
- **Button**: "Begin Journey" (or current placeholder behavior)
- **Styling**: Large font, centered
- **Animation**: Fade-in
- **Action**: Saves character and redirects to game/load screen

---

## Key Animation Requirements

### Wipe-Down Fade Effect
- **NOT typewriter**: No character-by-character typing
- **Effect**: Text fades in with slight downward slide (translateY)
- **CSS**:
  ```css
  opacity: 0 → 1
  transform: translateY(-20px) → translateY(0)
  transition: all 0.8s ease-out
  ```

### Fade-to-Black Scenes
- Scenes without images (final_words, scene5a, final_text)
- **Background**: Pure black (`bg-black`)
- **Text**: Larger, centered, prominent color
- **Font**: Same pixel font (Pixelify Sans from layout)

### Scene Transitions
- **Fade out duration**: ~0.5-0.8 seconds
- **Pause between scenes**: ~0.3 seconds (black screen)
- **Fade in duration**: ~0.8-1.0 seconds

### Background Images
- **Display**: Fullscreen, cover, centered
- **Z-index**: Behind text
- **Overlay**: Optional dark overlay (rgba(0,0,0,0.3)) for text readability

---

## Data Structure Reference

```javascript
// Example data access
const intro = await fetch('/data/character/introductions.json');

// Scene with image
intro.scene1.text → "The rain falls..."
intro.scene1.image → "scene1.png"

// Black screen scene (no image)
intro.final_words.text → "Your destiny awaits..."

// Background-specific
intro.background_intros[character.background].text
intro.background_intros[character.background].image

// Class-based equipment
const classCategory = getEquipmentCategory(character.class);
intro.equipment_intros[classCategory].text
intro.equipment_intros[classCategory].image
intro.equipment_intros[classCategory].classes // Array of applicable classes
```

---

## CSS Classes Needed

```css
/* Scene container - fullscreen */
.intro-scene {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Background image */
.scene-background {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
  z-index: -1;
}

/* Text container */
.scene-text {
  max-width: 56rem; /* max-w-4xl */
  padding: 2rem;
  text-align: center;
  font-size: 1.125rem; /* text-lg */
  line-height: 1.75;
}

/* Large quote text (black screens) */
.scene-quote {
  font-size: 1.875rem; /* text-3xl */
  font-weight: 700;
  color: #facc15; /* yellow-400 */
  max-width: 48rem;
  padding: 2rem;
}

/* Letter styling */
.letter-text {
  font-family: serif;
  font-style: italic;
  background: rgba(255, 248, 220, 0.9); /* cornsilk overlay */
  color: black;
  padding: 2rem;
  border-radius: 0.5rem;
  max-width: 48rem;
}

/* Wipe-down fade animation */
@keyframes wipeDownFade {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-wipe-down {
  animation: wipeDownFade 0.8s ease-out forwards;
}
```

---

## Implementation Checklist

- [ ] Remove typewriter effect from all scenes
- [ ] Implement wipe-down fade animation
- [ ] Add background image support to scene divs
- [ ] Create fade-to-black scene styling (large centered text)
- [ ] Update scene flow to match order above
- [ ] Add letter styling for scene 4a
- [ ] Refactor equipment selection to individual item scenes
- [ ] Add pack selection after scene6
- [ ] Ensure all transitions are fade-based
- [ ] Test with different character backgrounds and classes
- [ ] Verify all scene images load correctly
- [ ] Check timing/duration of each scene

---

## Notes for Developer

1. **Scene images are placeholders**: scene3.png is currently a black placeholder. Real background-specific images will be added later.

2. **Font is global**: The Pixelify Sans font is already loaded in layout.html, no need to specify it per scene.

3. **Equipment category mapping**: Some classes appear in multiple categories (e.g., Paladin in both warrior and faithful, Barbarian in warrior and wilderness, Ranger in wilderness and clever). Choose primary category based on class focus.

4. **Timing is flexible**: The durations listed are suggestions. Adjust based on text length and user testing.

5. **Letter intro vs letter reading**: Don't confuse `letter_intro` (scene 6) with `background_letters` (scene 7/4a). One introduces the letter, the other shows its content.

6. **All text fields are final**: No [Fade to:] prefixes remain in the JSON - those were just notes in the draft.
