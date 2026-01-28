/**
 * Damage Types Shared Utility
 *
 * Provides damage type emoji mapping and helper functions
 * used across spell displays, ability displays, and combat UI.
 *
 * @module lib/damageTypes
 */

export const damageEmojis = {
    fire: '🔥',
    cold: '❄️',
    lightning: '⚡',
    thunder: '⚡',
    acid: '🧪',
    poison: '☠️',
    necrotic: '☠️',
    radiant: '✨',
    psychic: '🧠',
    force: '🌀',
    slashing: '🗡️',
    piercing: '🏹',
    bludgeoning: '🔨',
    healing: '💚',
};

/**
 * Get emoji for a damage type
 * @param {string} damageType - The damage type (e.g., "fire", "cold")
 * @param {boolean} [isHealing=false] - Whether this is a healing effect
 * @returns {string} Emoji character
 */
export function getDamageEmoji(damageType, isHealing = false) {
    return damageEmojis[damageType] || (isHealing ? '💚' : '⚔️');
}
