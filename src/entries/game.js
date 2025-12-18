/**
 * Game Page Entry Point
 *
 * Main game page bundle - imports and initializes all game systems.
 * This replaces the individual script tags in game.html.
 */

// Core libraries
import { logger } from '../lib/logger.js';
import '../lib/session.js'; // Auto-initializes as window.sessionManager
import '../lib/api.js'; // Auto-initializes as window.gameAPI

// State management
import { getGameState, updateGameState, getGameStateSync, refreshGameState, initializeGame } from '../state/gameState.js';
import { getItemById, getSpellById, getLocationById } from '../state/staticData.js';

// Systems
import { saveGameToLocal } from '../systems/saveSystem.js';
import * as inventoryInteractions from '../systems/inventoryInteractions.js';
import { openContainer, closeContainer } from '../systems/containers.js';

// UI modules
import { updateTimeDisplay } from '../ui/timeDisplay.js';
import { updateCharacterDisplay } from '../ui/characterDisplay.js';
import { displayCurrentLocation } from '../ui/locationDisplay.js';
import { updateSpellsDisplay } from '../ui/spellsDisplay.js';
import { updateAllDisplays } from '../ui/displayCoordinator.js';
import { showMessage, showActionText, addGameLog } from '../ui/messaging.js';
import { openGroundModal } from '../ui/groundItems.js';

// Page initialization
import { nostrHeroStartup } from '../pages/startup.js';

// Logic
import * as mechanics from '../logic/mechanics.js';
import { NostrCharacterGenerator } from '../logic/characterGenerator.js';

// TEST: This should appear first in console
console.log('🚀 BUNDLE LOADING - game.js entry point reached');
console.log('📋 Document ready state:', document.readyState);
console.log('📋 nostrHeroStartup exists:', typeof nostrHeroStartup);

// Make critical functions globally available for templates and inline scripts
window.getGameState = getGameState;
window.updateGameState = updateGameState;
window.getGameStateSync = getGameStateSync;
window.refreshGameState = refreshGameState;
window.initializeGame = initializeGame;
window.getItemById = getItemById;
window.getSpellById = getSpellById;
window.getLocationById = getLocationById;
window.saveGameToLocal = saveGameToLocal;
window.saveGame = saveGameToLocal; // Alias for template compatibility
window.inventoryInteractions = inventoryInteractions;
window.updateTimeDisplay = updateTimeDisplay;
window.updateCharacterDisplay = updateCharacterDisplay;
window.displayCurrentLocation = displayCurrentLocation;
window.updateSpellsDisplay = updateSpellsDisplay;
window.updateAllDisplays = updateAllDisplays;
window.showMessage = showMessage;
window.showActionText = showActionText;
window.addGameLog = addGameLog;
window.openGroundModal = openGroundModal;
window.moveToLocation = mechanics.moveToLocation;
window.characterGenerator = new NostrCharacterGenerator();
window.openContainer = openContainer;
window.closeContainer = closeContainer;

// Initialize inventory interactions on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        inventoryInteractions.bindInventoryEvents();
    });
} else {
    inventoryInteractions.bindInventoryEvents();
}

logger.info('🎮 Game page bundle loaded');
