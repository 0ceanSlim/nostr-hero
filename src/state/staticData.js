/**
 * Static Data Lookup Module
 *
 * Provides SYNCHRONOUS lookup functions for game data cached in the DOM.
 *
 * ⚠️ TRANSITIONAL CODE - This will eventually be replaced
 *
 * Use these functions when you need SYNCHRONOUS access to cached data (e.g., in render loops).
 * For async database queries, use the modules in data/ instead (e.g., data/items.js).
 *
 * @module state/staticData
 */

import { logger } from '../lib/logger.js';

/**
 * Get item by ID from cached data
 * @param {string} itemId - Item ID to look up
 * @returns {Object|undefined} Item object or undefined if not found
 */
export function getItemById(itemId) {
    const element = document.getElementById('all-items');
    if (!element) {
        logger.warn('all-items element not found in DOM');
        return undefined;
    }

    const allItems = JSON.parse(element.textContent || '[]');
    return allItems.find(item => item.id === itemId);
}

/**
 * Get spell by ID from cached data
 * @param {string} spellId - Spell ID to look up
 * @returns {Object|undefined} Spell object or undefined if not found
 */
export function getSpellById(spellId) {
    const element = document.getElementById('all-spells');
    if (!element) {
        logger.warn('all-spells element not found in DOM');
        return undefined;
    }

    const allSpells = JSON.parse(element.textContent || '[]');
    return allSpells.find(spell => spell.id === spellId);
}

/**
 * Get location by ID from cached data
 * @param {string} locationId - Location ID to look up
 * @returns {Object|undefined} Location object or undefined if not found
 */
export function getLocationById(locationId) {
    const element = document.getElementById('all-locations');
    if (!element) {
        logger.warn('all-locations element not found in DOM');
        return undefined;
    }

    const allLocations = JSON.parse(element.textContent || '[]');

    // First, try to find in top-level locations
    const topLevel = allLocations.find(location => location.id === locationId);
    if (topLevel) return topLevel;

    // If not found, search within city districts
    for (const location of allLocations) {
        if (location.properties?.districts) {
            for (const district of Object.values(location.properties.districts)) {
                if (district.id === locationId) {
                    return district;
                }
            }
        }
    }

    return undefined;
}

/**
 * Get monster by ID from cached data
 * @param {string} monsterId - Monster ID to look up
 * @returns {Object|undefined} Monster object or undefined if not found
 */
export function getMonsterById(monsterId) {
    const element = document.getElementById('all-monsters');
    if (!element) {
        logger.warn('all-monsters element not found in DOM');
        return undefined;
    }

    const allMonsters = JSON.parse(element.textContent || '[]');
    return allMonsters.find(monster => monster.id === monsterId);
}

/**
 * Get equipment pack by ID from cached data
 * @param {string} packId - Pack ID to look up
 * @returns {Object|undefined} Pack object or undefined if not found
 */
export function getPackById(packId) {
    const element = document.getElementById('all-packs');
    if (!element) {
        logger.warn('all-packs element not found in DOM');
        return undefined;
    }

    const allPacks = JSON.parse(element.textContent || '[]');
    return allPacks.find(pack => pack.id === packId);
}

/**
 * Get NPC by ID from cached data
 * @param {string} npcId - NPC ID to look up
 * @returns {Object|undefined} NPC object or undefined if not found
 */
export function getNPCById(npcId) {
    const element = document.getElementById('all-npcs');
    if (!element) {
        logger.warn('all-npcs element not found in DOM');
        return undefined;
    }

    const allNPCs = JSON.parse(element.textContent || '[]');
    return allNPCs.find(npc => npc.id === npcId);
}

logger.debug('Static data lookup module loaded');
