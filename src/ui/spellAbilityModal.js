/**
 * Spell/Ability Detail Modal Controller
 *
 * Opens detail modals for spells and abilities in the scene window overlay.
 * - Spell modal: school, description, stats, damage+emoji, component item icons, "Prepare" stub
 * - Ability modal: cost badge, description, tier tabs, effects per tier, "Use" stub
 *
 * @module ui/spellAbilityModal
 */

import { logger } from '../lib/logger.js';
import { getSpellById } from '../state/staticData.js';
import { getDamageEmoji } from '../lib/damageTypes.js';

// Currently displayed data
let currentSpellId = null;
let currentAbility = null;
let currentTierIndex = 0;

/**
 * Open the modal for a spell
 */
export function openSpellModal(spellId) {
    const spell = getSpellById(spellId);
    if (!spell) {
        logger.error(`Spell not found for modal: ${spellId}`);
        return;
    }

    currentSpellId = spellId;
    currentAbility = null;

    const modal = document.getElementById('spell-ability-modal');
    if (!modal) return;

    // Title bar
    const icon = document.getElementById('modal-icon');
    if (icon) {
        icon.src = `/res/img/spells/${spell.school}.png`;
        icon.classList.remove('hidden');
    }
    const title = document.getElementById('modal-title');
    if (title) title.textContent = spell.name;

    // Hide tier tabs (spells don't have tiers)
    const tierTabs = document.getElementById('modal-tier-tabs');
    if (tierTabs) tierTabs.classList.add('hidden');

    // Badges: school + level
    const badges = document.getElementById('modal-badges');
    if (badges) {
        badges.innerHTML = '';
        badges.appendChild(createBadge(capitalize(spell.school), '#4a5568'));
        const levelText = spell.level === 0 ? 'Cantrip' : `Level ${spell.level}`;
        badges.appendChild(createBadge(levelText, '#2563eb'));
        if (spell.mana_cost) {
            badges.appendChild(createBadge(`${spell.mana_cost} MP`, '#1d4ed8'));
        }
    }

    // Description
    const desc = document.getElementById('modal-description');
    if (desc) desc.textContent = spell.description || '';

    // Stats grid
    const statsGrid = document.getElementById('modal-stats-grid');
    if (statsGrid) {
        statsGrid.innerHTML = '';
        const stats = [
            ['Range', spell.range || '—'],
            ['Cast Time', spell.casting_time || '—'],
            ['Duration', spell.duration || '—'],
            ['Save', spell.save_type ? capitalize(spell.save_type) : '—'],
        ];
        stats.forEach(([label, value]) => {
            const statEl = document.createElement('div');
            statEl.innerHTML = `<span class="text-gray-500">${label}:</span> <span class="text-white">${value}</span>`;
            statsGrid.appendChild(statEl);
        });
        statsGrid.classList.remove('hidden');
    }

    // Damage section
    const dmgSection = document.getElementById('modal-damage-section');
    if (dmgSection) {
        if (spell.damage || spell.heal) {
            const isHealing = !!spell.heal;
            const emoji = getDamageEmoji(spell.damage_type, isHealing);
            const roll = spell.damage || spell.heal;
            const typeLabel = isHealing ? 'Healing' : capitalize(spell.damage_type || 'damage');
            const color = isHealing ? '#4ade80' : '#f87171';
            dmgSection.innerHTML = `
                <div class="font-bold mb-1" style="color: ${color};">${emoji} ${typeLabel}</div>
                <div class="text-white">${roll}</div>
            `;
            dmgSection.classList.remove('hidden');
        } else {
            dmgSection.classList.add('hidden');
        }
    }

    // Effects section (hide for spells)
    const effectsSection = document.getElementById('modal-effects-section');
    if (effectsSection) effectsSection.classList.add('hidden');

    // Components section
    const compSection = document.getElementById('modal-components-section');
    const compList = document.getElementById('modal-components-list');
    if (compSection && compList) {
        if (spell.material_component?.required && spell.material_component.required.length > 0) {
            compList.innerHTML = '';
            spell.material_component.required.forEach(comp => {
                const item = document.createElement('div');
                item.className = 'flex items-center gap-1';
                item.innerHTML = `
                    <img src="/res/img/items/${comp.component}.png" alt="${comp.component}"
                         style="width: 16px; height: 16px; image-rendering: pixelated;"
                         onerror="this.style.display='none'">
                    <span class="text-gray-300">${comp.component.replace(/-/g, ' ')} x${comp.quantity}</span>
                `;
                compList.appendChild(item);
            });
            compSection.classList.remove('hidden');
        } else {
            compSection.classList.add('hidden');
        }
    }

    // Notes
    const notesEl = document.getElementById('modal-notes');
    if (notesEl) {
        if (spell.notes && spell.notes.length > 0) {
            notesEl.textContent = spell.notes.join(' · ');
            notesEl.classList.remove('hidden');
        } else {
            notesEl.classList.add('hidden');
        }
    }

    // Action button (Prepare - stubbed)
    const actionContainer = document.getElementById('modal-action-container');
    const actionButton = document.getElementById('modal-action-button');
    if (actionContainer && actionButton) {
        actionButton.textContent = 'PREPARE';
        actionContainer.classList.remove('hidden');
    }

    modal.classList.remove('hidden');
}

/**
 * Open the modal for a martial ability
 */
export function openAbilityModal(ability) {
    if (!ability) return;

    currentAbility = ability;
    currentSpellId = null;
    currentTierIndex = 0;

    const modal = document.getElementById('spell-ability-modal');
    if (!modal) return;

    // Title bar
    const icon = document.getElementById('modal-icon');
    if (icon) icon.classList.add('hidden');
    const title = document.getElementById('modal-title');
    if (title) title.textContent = ability.name;

    // Tier tabs
    const tierTabs = document.getElementById('modal-tier-tabs');
    if (tierTabs && ability.all_tiers && ability.all_tiers.length > 1) {
        tierTabs.innerHTML = '';

        // Find which tier is current
        let activeTierIdx = 0;
        if (ability.current_tier) {
            activeTierIdx = ability.all_tiers.findIndex(
                t => t.min_level === ability.current_tier.min_level
            );
            if (activeTierIdx < 0) activeTierIdx = 0;
        }
        currentTierIndex = activeTierIdx;

        ability.all_tiers.forEach((tier, i) => {
            const tab = document.createElement('button');
            const isActive = i === activeTierIdx;
            const isCurrentCharTier = i === activeTierIdx;
            const label = tier.max_level >= 99
                ? `Lv ${tier.min_level}+`
                : `Lv ${tier.min_level}-${tier.max_level}`;

            tab.className = 'px-2 py-1 font-bold';
            tab.style.cssText = `
                font-size: 6px;
                background: ${isActive ? '#2563eb' : '#1a1a2e'};
                color: ${isActive ? '#fff' : '#888'};
                border: 1px solid ${isCurrentCharTier ? '#3b82f6' : '#333'};
                border-radius: 2px;
                cursor: pointer;
            `;
            tab.textContent = label;
            tab.onclick = () => selectAbilityTier(i);

            tierTabs.appendChild(tab);
        });

        tierTabs.classList.remove('hidden');
    } else if (tierTabs) {
        tierTabs.classList.add('hidden');
    }

    // Render the rest based on active tier
    renderAbilityTierContent(ability, currentTierIndex);

    // Hide spell-specific sections
    const statsGrid = document.getElementById('modal-stats-grid');
    if (statsGrid) statsGrid.classList.add('hidden');

    const compSection = document.getElementById('modal-components-section');
    if (compSection) compSection.classList.add('hidden');

    const notesEl = document.getElementById('modal-notes');
    if (notesEl) notesEl.classList.add('hidden');

    // Damage section (hide for abilities)
    const dmgSection = document.getElementById('modal-damage-section');
    if (dmgSection) dmgSection.classList.add('hidden');

    // Action button (Use Ability - stubbed)
    const actionContainer = document.getElementById('modal-action-container');
    const actionButton = document.getElementById('modal-action-button');
    if (actionContainer && actionButton) {
        if (ability.is_unlocked && ability.cooldown !== 'passive') {
            actionButton.textContent = 'USE ABILITY';
            actionContainer.classList.remove('hidden');
        } else {
            actionContainer.classList.add('hidden');
        }
    }

    modal.classList.remove('hidden');
}

/**
 * Select a tier tab in the ability modal
 */
function selectAbilityTier(tierIndex) {
    if (!currentAbility) return;
    currentTierIndex = tierIndex;

    // Update tab styles
    const tierTabs = document.getElementById('modal-tier-tabs');
    if (tierTabs) {
        Array.from(tierTabs.children).forEach((tab, i) => {
            const isActive = i === tierIndex;
            tab.style.background = isActive ? '#2563eb' : '#1a1a2e';
            tab.style.color = isActive ? '#fff' : '#888';
        });
    }

    renderAbilityTierContent(currentAbility, tierIndex);
}

/**
 * Render the ability modal content for a specific tier
 */
function renderAbilityTierContent(ability, tierIndex) {
    const tier = ability.all_tiers?.[tierIndex];

    // Badges: cost + cooldown
    const badges = document.getElementById('modal-badges');
    if (badges) {
        badges.innerHTML = '';

        const isPassive = ability.cooldown === 'passive';
        const cost = tier?.override_cost ?? ability.resource_cost;
        const costLabel = isPassive ? 'Passive' : `${cost} ${ability.resource_type.toUpperCase()}`;
        badges.appendChild(createBadge(costLabel, isPassive ? '#059669' : '#d97706'));

        if (ability.cooldown && ability.cooldown !== 'none' && ability.cooldown !== 'passive') {
            const cooldownLabel = ability.cooldown.replace(/_/g, ' ');
            const uses = tier?.override_cooldown;
            const usesText = uses ? `${cooldownLabel} (${uses}x)` : cooldownLabel;
            badges.appendChild(createBadge(capitalize(usesText), '#6b7280'));
        }

        if (!ability.is_unlocked) {
            badges.appendChild(createBadge(`Unlocks Lv ${ability.unlock_level}`, '#991b1b'));
        }
    }

    // Description
    const desc = document.getElementById('modal-description');
    if (desc) desc.textContent = ability.description || '';

    // Effects section
    const effectsSection = document.getElementById('modal-effects-section');
    if (effectsSection && tier) {
        const levelRange = tier.max_level >= 99
            ? `Level ${tier.min_level}+`
            : `Level ${tier.min_level}-${tier.max_level}`;

        effectsSection.innerHTML = `
            <div class="font-bold mb-1 text-yellow-400">EFFECTS (${levelRange})</div>
            <div class="text-gray-200">${tier.summary || 'No effects defined'}</div>
        `;

        // Show effects_applied IDs for reference
        if (tier.effects_applied && tier.effects_applied.length > 0) {
            const effectsRef = document.createElement('div');
            effectsRef.className = 'mt-1';
            effectsRef.style.cssText = 'font-size: 6px; color: #555;';
            effectsRef.textContent = `Effects: ${tier.effects_applied.join(', ')}`;
            effectsSection.appendChild(effectsRef);
        }

        effectsSection.classList.remove('hidden');
    } else if (effectsSection) {
        effectsSection.classList.add('hidden');
    }
}

/**
 * Close the spell/ability modal
 */
export function closeSpellAbilityModal() {
    const modal = document.getElementById('spell-ability-modal');
    if (modal) modal.classList.add('hidden');
    currentSpellId = null;
    currentAbility = null;
}

/**
 * Handle the modal action button (Prepare spell / Use ability)
 * Currently stubbed - will be implemented with combat system
 */
export function handleModalAction() {
    if (currentSpellId) {
        logger.info(`[STUB] Prepare spell: ${currentSpellId}`);
        // TODO: Show slot selection for spell preparation
        closeSpellAbilityModal();
    } else if (currentAbility) {
        logger.info(`[STUB] Use ability: ${currentAbility.id}`);
        // TODO: Execute ability through game action API
        closeSpellAbilityModal();
    }
}

// ============================================================
// Helpers
// ============================================================

function createBadge(text, color) {
    const badge = document.createElement('span');
    badge.className = 'inline-block px-2 py-0.5 font-bold rounded';
    badge.style.cssText = `font-size: 6px; background: ${color}; color: #fff;`;
    badge.textContent = text;
    return badge;
}

function capitalize(str) {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
}

// ============================================================
// Window bindings + keyboard handler
// ============================================================

function handleEscapeKey(e) {
    if (e.key === 'Escape') {
        const modal = document.getElementById('spell-ability-modal');
        if (modal && !modal.classList.contains('hidden')) {
            closeSpellAbilityModal();
            e.stopPropagation();
        }
    }
}

// Bind to window for onclick handlers in HTML
window.openSpellModal = openSpellModal;
window.openAbilityModal = openAbilityModal;
window.closeSpellAbilityModal = closeSpellAbilityModal;
window.handleModalAction = handleModalAction;

// Listen for Escape key
document.addEventListener('keydown', handleEscapeKey);

logger.debug('Spell/ability modal controller loaded');
