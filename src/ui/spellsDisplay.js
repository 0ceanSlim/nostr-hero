/**
 * Spells Display UI Module
 *
 * Handles display of known spells and spell slots.
 * Shows spell cards with school backgrounds and spell slot management.
 *
 * @module ui/spellsDisplay
 */

import { logger } from '../lib/logger.js';
import { getGameStateSync } from '../state/gameState.js';
import { getSpellById } from '../state/staticData.js';

/**
 * Update the spells display tab
 * Shows known spells and spell slots with prepared spells
 */
export function updateSpellsDisplay() {
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
                logger.error(`Spell not found: ${spellId}`);
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

logger.debug('Spells display module loaded');
