/**
 * Display Coordinator UI Module
 *
 * Coordinates updates across all UI modules.
 *
 * @module ui/displayCoordinator
 */

import { logger } from '../lib/logger.js';
import { updateCharacterDisplay } from './characterDisplay.js';
import { updateSpellsDisplay } from './spellsDisplay.js';
import { displayCurrentLocation } from './locationDisplay.js';
import { updateTimeDisplay } from './timeDisplay.js';

/**
 * Update all UI displays
 * Main coordinator function that refreshes all UI components
 */
export async function updateAllDisplays() {
    await updateCharacterDisplay(); // Async - must await
    updateSpellsDisplay();
    await displayCurrentLocation(); // Async - must await (fetches NPCs from backend)
    updateTimeDisplay();
}

logger.debug('Display coordinator module loaded');
