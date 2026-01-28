/**
 * Spells & Abilities Display UI Module
 *
 * Dual-view display:
 * - Casters: Spell slots (row layout) + known spells list with columns
 * - Martials: Class-specific ability layouts (Fighter/Barbarian/Monk/Rogue)
 *
 * @module ui/spellsDisplay
 */

import { logger } from '../lib/logger.js';
import { getGameStateSync } from '../state/gameState.js';
import { getSpellById } from '../state/staticData.js';
import { getClassResourceSync } from '../data/classResources.js';
import { getDamageEmoji } from '../lib/damageTypes.js';

// Cache fetched abilities
let abilitiesCache = {};

/**
 * Update the spells/abilities display tab
 * Detects caster vs martial and shows appropriate view
 */
export function updateSpellsDisplay() {
    const state = getGameStateSync();
    const character = state.character;
    if (!character) return;

    const resourceConfig = getClassResourceSync(character.class);
    const isCaster = resourceConfig.type === 'mana';

    const casterView = document.getElementById('caster-view');
    const martialView = document.getElementById('martial-view');

    if (!casterView || !martialView) return;

    if (isCaster) {
        casterView.classList.remove('hidden');
        martialView.classList.add('hidden');
        renderCasterView(character);
    } else {
        casterView.classList.add('hidden');
        martialView.classList.remove('hidden');
        renderMartialView(character, resourceConfig);
    }
}

// ============================================================
// CASTER VIEW
// ============================================================

function renderCasterView(character) {
    renderSpellSlots(character);
    renderKnownSpellsList(character);
}

/**
 * Render spell slots as rows (level label + slot cells)
 */
function renderSpellSlots(character) {
    const container = document.getElementById('spell-slots-container');
    if (!container || !character?.spell_slots) return;
    container.innerHTML = '';

    const sortedLevels = Object.keys(character.spell_slots).sort((a, b) => {
        if (a === 'cantrips') return -1;
        if (b === 'cantrips') return 1;
        return (parseInt(a.split('_')[1]) || 0) - (parseInt(b.split('_')[1]) || 0);
    });

    sortedLevels.forEach(level => {
        const slots = character.spell_slots[level];
        if (!slots || !Array.isArray(slots) || slots.length === 0) return;

        const row = document.createElement('div');
        row.className = 'flex items-center gap-1';

        // Level label
        const label = document.createElement('div');
        label.className = 'text-gray-400 font-bold flex-shrink-0';
        label.style.cssText = 'width: 52px; font-size: 6px;';
        label.textContent = level === 'cantrips' ? 'Cantrips' : `Level ${level.split('_')[1]}`;
        row.appendChild(label);

        // Slot cells
        const slotsContainer = document.createElement('div');
        slotsContainer.className = 'flex gap-1 flex-wrap';

        slots.forEach(slotData => {
            const cell = document.createElement('div');
            cell.className = 'flex items-center justify-center cursor-pointer hover:brightness-125';
            cell.style.cssText = 'width: 60px; height: 20px; font-size: 6px; background: #1a1a2e; border: 1px solid #333; border-radius: 2px;';

            if (slotData.spell) {
                const spell = getSpellById(slotData.spell);
                const name = spell?.name || slotData.spell.replace(/-/g, ' ');
                cell.innerHTML = `<span class="text-purple-300 truncate px-1">${name}</span>`;
                cell.style.borderColor = '#7c3aed';
                cell.onclick = () => window.openSpellModal && window.openSpellModal(slotData.spell);
            } else {
                cell.innerHTML = '<span class="text-gray-600">[ empty ]</span>';
            }

            slotsContainer.appendChild(cell);
        });

        row.appendChild(slotsContainer);
        container.appendChild(row);
    });
}

/**
 * Render known spells as a list with columns:
 * school icon | name | damage emoji + roll | component icons
 */
function renderKnownSpellsList(character) {
    const container = document.getElementById('known-spells-list');
    if (!container || !character?.spells) return;
    container.innerHTML = '';

    const spellsArray = Array.isArray(character.spells) ? character.spells : [];

    if (spellsArray.length === 0) {
        container.innerHTML = '<div class="text-gray-500 italic" style="font-size: 7px;">No spells known</div>';
        return;
    }

    spellsArray.forEach(spellId => {
        const spell = getSpellById(spellId);
        if (!spell) return;

        const row = document.createElement('div');
        row.className = 'flex items-center gap-1 px-1 py-0.5 cursor-pointer hover:bg-white hover:bg-opacity-5';
        row.style.cssText = 'border-bottom: 1px solid #222; min-height: 18px;';
        row.onclick = () => window.openSpellModal && window.openSpellModal(spellId);

        // School icon
        const schoolIcon = document.createElement('img');
        schoolIcon.src = `/res/img/spells/${spell.school}.png`;
        schoolIcon.alt = spell.school;
        schoolIcon.style.cssText = 'width: 14px; height: 14px; image-rendering: pixelated; flex-shrink: 0;';
        schoolIcon.onerror = () => { schoolIcon.style.display = 'none'; };
        row.appendChild(schoolIcon);

        // Spell name
        const nameSpan = document.createElement('span');
        nameSpan.className = 'text-white flex-1 truncate';
        nameSpan.style.fontSize = '7px';
        nameSpan.textContent = spell.name;
        row.appendChild(nameSpan);

        // Damage/heal emoji + roll
        if (spell.damage || spell.heal) {
            const dmgSpan = document.createElement('span');
            dmgSpan.style.cssText = 'font-size: 7px; flex-shrink: 0; white-space: nowrap;';
            const isHealing = !!spell.heal;
            const emoji = getDamageEmoji(spell.damage_type, isHealing);
            const roll = spell.damage || spell.heal;
            dmgSpan.className = isHealing ? 'text-green-400' : 'text-red-400';
            dmgSpan.textContent = `${emoji} ${roll}`;
            row.appendChild(dmgSpan);
        }

        // Material component icons
        if (spell.material_component?.required) {
            const compContainer = document.createElement('div');
            compContainer.className = 'flex gap-0.5 flex-shrink-0';
            spell.material_component.required.forEach(comp => {
                const compIcon = document.createElement('img');
                compIcon.src = `/res/img/items/${comp.component}.png`;
                compIcon.alt = comp.component;
                compIcon.title = `${comp.component} x${comp.quantity}`;
                compIcon.style.cssText = 'width: 12px; height: 12px; image-rendering: pixelated;';
                compIcon.onerror = () => { compIcon.style.display = 'none'; };
                compContainer.appendChild(compIcon);
            });
            row.appendChild(compContainer);
        }

        container.appendChild(row);
    });
}

// ============================================================
// MARTIAL VIEW
// ============================================================

async function renderMartialView(character, resourceConfig) {
    const className = character.class.toLowerCase();
    const level = character.level || 1;

    // Hide all layouts, show the right one
    ['fighter', 'barbarian', 'monk', 'rogue'].forEach(c => {
        const el = document.getElementById(`${c}-layout`);
        if (el) el.classList.toggle('hidden', c !== className);
    });

    // Fetch abilities from API
    const abilities = await fetchAbilities(className, level);
    if (!abilities) return;

    switch (className) {
        case 'fighter':
            renderFighterLayout(abilities, level, resourceConfig);
            break;
        case 'barbarian':
            renderBarbarianLayout(abilities, level, resourceConfig);
            break;
        case 'monk':
            renderMonkLayout(abilities, level, resourceConfig);
            break;
        case 'rogue':
            renderRogueLayout(abilities, level, resourceConfig);
            break;
    }
}

async function fetchAbilities(className, level) {
    const cacheKey = `${className}-${level}`;
    if (abilitiesCache[cacheKey]) return abilitiesCache[cacheKey];

    try {
        const response = await fetch(`/api/abilities?class=${className}&level=${level}`);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        const data = await response.json();
        if (data.success) {
            abilitiesCache[cacheKey] = data.abilities;
            return data.abilities;
        }
    } catch (err) {
        logger.error('Failed to fetch abilities:', err);
    }
    return null;
}

/**
 * Create an ability card element (shared across layouts)
 */
function createAbilityCard(ability, level, resourceConfig, options = {}) {
    const { width = '100%', compact = false } = options;
    const card = document.createElement('div');
    const isUnlocked = ability.is_unlocked;
    const isPassive = ability.cooldown === 'passive';

    card.className = 'cursor-pointer';
    card.style.cssText = `
        width: ${width};
        padding: ${compact ? '4px 6px' : '6px 8px'};
        background: ${isUnlocked ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.3)'};
        border: 1px solid ${isUnlocked ? resourceConfig.color + '60' : '#333'};
        border-radius: 3px;
        opacity: ${isUnlocked ? '1' : '0.5'};
        font-size: 7px;
        box-sizing: border-box;
    `;

    if (isUnlocked) {
        card.addEventListener('mouseenter', () => { card.style.borderColor = resourceConfig.color; });
        card.addEventListener('mouseleave', () => { card.style.borderColor = resourceConfig.color + '60'; });
    }

    // Name row
    const nameRow = document.createElement('div');
    nameRow.className = 'flex items-center justify-between';

    const nameSpan = document.createElement('span');
    nameSpan.className = 'font-bold';
    nameSpan.style.color = isUnlocked ? '#fff' : '#666';
    nameSpan.textContent = isUnlocked ? ability.name : `🔒 ${ability.name}`;
    nameRow.appendChild(nameSpan);

    if (!isUnlocked) {
        const lvlSpan = document.createElement('span');
        lvlSpan.className = 'text-gray-600 flex-shrink-0';
        lvlSpan.style.fontSize = '6px';
        lvlSpan.textContent = `Lv ${ability.unlock_level}`;
        nameRow.appendChild(lvlSpan);
    }

    card.appendChild(nameRow);

    // Cost + summary
    if (isUnlocked && !compact) {
        const infoRow = document.createElement('div');
        infoRow.style.cssText = 'margin-top: 2px; font-size: 6px;';

        const costText = isPassive ? 'Passive' : `${ability.resource_cost} ${resourceConfig.short_label}`;
        const tierSummary = ability.current_tier?.summary || '';

        infoRow.innerHTML = `
            <span style="color: ${resourceConfig.color};">${costText}</span>
            <span class="text-gray-400"> · </span>
            <span class="text-gray-300">${tierSummary}</span>
        `;
        card.appendChild(infoRow);
    }

    // Click handler
    card.onclick = () => window.openAbilityModal && window.openAbilityModal(ability);

    return card;
}

// ============================================================
// FIGHTER: Battle Formation (2-col grid, 3 rows)
// ============================================================

function renderFighterLayout(abilities, level, resourceConfig) {
    const grid = document.getElementById('fighter-abilities-grid');
    if (!grid) return;
    grid.innerHTML = '';

    abilities.forEach(ability => {
        grid.appendChild(createAbilityCard(ability, level, resourceConfig));
    });
}

// ============================================================
// BARBARIAN: Rage Escalation (tiered sections, bottom to top)
// ============================================================

function renderBarbarianLayout(abilities, level, resourceConfig) {
    const container = document.getElementById('barbarian-abilities-tiers');
    if (!container) return;
    container.innerHTML = '';

    // Group abilities into themed tiers
    const tierNames = ['RAGE', 'WRATH', 'FURY', 'PRIMAL'];
    const tierGroups = [
        abilities.filter(a => a.unlock_level <= 3),
        abilities.filter(a => a.unlock_level > 3 && a.unlock_level <= 7),
        abilities.filter(a => a.unlock_level > 7 && a.unlock_level <= 10),
        abilities.filter(a => a.unlock_level > 10),
    ];

    tierGroups.forEach((group, i) => {
        if (group.length === 0) return;

        const section = document.createElement('div');
        section.style.cssText = `
            display: block;
            border: 1px solid ${resourceConfig.color}30;
            border-radius: 3px;
            padding: 4px;
            background: rgba(0,0,0,0.2);
            width: 100%;
            box-sizing: border-box;
        `;

        // Tier label
        const label = document.createElement('div');
        label.className = 'text-center font-bold mb-1';
        label.style.cssText = `font-size: 6px; color: ${resourceConfig.color}; letter-spacing: 2px;`;
        label.textContent = `── ${tierNames[i]} ──`;
        section.appendChild(label);

        // Abilities in this tier
        const abilitiesDiv = document.createElement('div');
        abilitiesDiv.className = 'space-y-1';
        group.forEach(ability => {
            abilitiesDiv.appendChild(createAbilityCard(ability, level, resourceConfig, { compact: true }));
        });
        section.appendChild(abilitiesDiv);

        container.appendChild(section);
    });
}

// ============================================================
// MONK: Chakra (diamond/centered pattern)
// ============================================================

function renderMonkLayout(abilities, level, resourceConfig) {
    const container = document.getElementById('monk-abilities-diamond');
    if (!container) return;
    container.innerHTML = '';

    // Diamond pattern: alternate between centered (1) and side-by-side (2)
    // Sort by unlock level, then arrange in diamond
    const sorted = [...abilities].sort((a, b) => a.unlock_level - b.unlock_level);

    // Pattern: [2 side], [1 center], [2 side], [1 center], etc.
    // Bottom = early abilities, top = late abilities (reversed for visual)
    const rows = [];
    let idx = 0;
    let isSide = true;
    while (idx < sorted.length) {
        if (isSide && idx + 1 < sorted.length) {
            rows.push([sorted[idx], sorted[idx + 1]]);
            idx += 2;
        } else {
            rows.push([sorted[idx]]);
            idx += 1;
        }
        isSide = !isSide;
    }

    // Reverse so strongest is at top
    rows.reverse();

    rows.forEach(rowAbilities => {
        const rowDiv = document.createElement('div');
        rowDiv.className = 'flex gap-2 justify-center';
        rowDiv.style.width = '100%';

        const cardWidth = rowAbilities.length === 1 ? '70%' : '48%';
        rowAbilities.forEach(ability => {
            rowDiv.appendChild(createAbilityCard(ability, level, resourceConfig, { width: cardWidth }));
        });

        container.appendChild(rowDiv);
    });
}

// ============================================================
// ROGUE: Shadow Web (web/network pattern with connecting lines)
// ============================================================

function renderRogueLayout(abilities, level, resourceConfig) {
    const container = document.getElementById('rogue-abilities-web');
    if (!container) return;
    container.innerHTML = '';

    // Web pattern: center, diverge, converge, diverge, center
    const sorted = [...abilities].sort((a, b) => a.unlock_level - b.unlock_level);

    // Pattern: [2 side], [1 center], [2 side], [1 center], etc.
    const rows = [];
    let idx = 0;
    let isSide = true;
    while (idx < sorted.length) {
        if (isSide && idx + 1 < sorted.length) {
            rows.push([sorted[idx], sorted[idx + 1]]);
            idx += 2;
        } else {
            rows.push([sorted[idx]]);
            idx += 1;
        }
        isSide = !isSide;
    }

    // Reverse so strongest at top
    rows.reverse();

    rows.forEach((rowAbilities, rowIdx) => {
        // Add connecting lines between rows
        if (rowIdx > 0) {
            const connector = document.createElement('div');
            connector.className = 'flex justify-center';
            connector.style.cssText = `color: ${resourceConfig.color}40; font-size: 8px; line-height: 1;`;
            const prevLen = rows[rowIdx - 1].length;
            const curLen = rowAbilities.length;
            if (prevLen === 1 && curLen === 2) {
                connector.textContent = '╱     ╲';
            } else if (prevLen === 2 && curLen === 1) {
                connector.textContent = '╲     ╱';
            } else {
                connector.textContent = '│';
            }
            container.appendChild(connector);
        }

        const rowDiv = document.createElement('div');
        rowDiv.className = 'flex gap-2 justify-center';
        rowDiv.style.width = '100%';

        const cardWidth = rowAbilities.length === 1 ? '65%' : '48%';
        rowAbilities.forEach(ability => {
            rowDiv.appendChild(createAbilityCard(ability, level, resourceConfig, { width: cardWidth }));
        });

        container.appendChild(rowDiv);
    });
}

logger.debug('Spells & abilities display module loaded');
