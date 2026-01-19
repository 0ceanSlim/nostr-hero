/**
 * 144x Real-Time Clock System (Delta Architecture)
 *
 * This is a compatibility layer that integrates the new delta-based systems:
 * - smoothClock.js: 60fps interpolated clock display
 * - tickManager.js: 417ms tick backend sync
 * - deltaApplier.js: Surgical DOM updates
 *
 * The old 5-second full-state sync is replaced with 417ms delta updates.
 * Time display runs at 60fps for smooth visuals.
 */

import { logger } from '../lib/logger.js';
import { getGameStateSync } from '../state/gameState.js';
import { updateTimeDisplay } from '../ui/timeDisplay.js';
import { updateCharacterDisplay } from '../ui/characterDisplay.js';
import { gameAPI } from '../lib/api.js';
import { eventBus } from '../lib/events.js';

// Import new delta architecture systems
import { smoothClock } from './smoothClock.js';
import { tickManager } from './tickManager.js';
import { deltaApplier } from './deltaApplier.js';

// Constants (kept for backwards compatibility)
const TICK_INTERVAL_MS = 417; // Now using 417ms ticks (1 in-game minute)
const TIME_MULTIPLIER = 144.0;  // 144x real-time speed
const GAME_MINUTES_PER_TICK = 1; // 1 in-game minute per tick at 417ms
const MINUTES_PER_DAY = 1440; // 24 hours * 60 minutes

// Use new delta system
const USE_DELTA_SYSTEM = true; // Set to false to revert to old behavior

// Legacy state (for backwards compatibility)
let tickIntervalId = null;
let displayIntervalId = null;
let isPaused = true;
let accumulatedMinutes = 0;
let displayAccumulatedMinutes = 0;
let backendSyncCounter = 0;

/**
 * Initialize the clock system
 */
export function initTimeClock() {
    logger.info('Initializing 144x time clock');

    if (USE_DELTA_SYSTEM) {
        // Initialize new delta-based systems
        initDeltaSystem();
    } else {
        // Legacy initialization
        startTicking();
        startDisplayUpdates();
    }

    // Update button state
    updatePlayPauseButton();

    logger.debug('Time clock initialized (paused)');
}

/**
 * Initialize the new delta-based clock system
 */
function initDeltaSystem() {
    logger.info('Using delta architecture for time updates');

    // Get initial time from game state (may not be loaded yet)
    syncClockToGameState();

    // Start smooth clock animation (60fps) - starts paused
    smoothClock.start();

    // Start tick manager (417ms ticks) - respects pause state
    tickManager.start();

    // Listen for character stat updates - update cache only, NO full re-renders
    eventBus.on('character:statsUpdated', (data) => {
        // Update cached game state silently (delta applier handles DOM)
        const state = getGameStateSync();
        if (state.character) {
            if (data.fatigue !== undefined) state.character.fatigue = data.fatigue;
            if (data.hunger !== undefined) state.character.hunger = data.hunger;
            if (data.hp !== undefined) state.character.hp = data.hp;
            if (data.active_effects) state.character.active_effects = data.active_effects;
        }
        // DO NOT emit gameStateChange - delta applier already updated DOM surgically
    });

    // Listen for game state loaded event to re-sync clock with correct time
    eventBus.on('gameStateLoaded', (loadedState) => {
        logger.info('📥 Game state loaded, re-syncing clock');
        // Pass the loaded state directly to avoid any timing issues
        syncClockToGameState(loadedState);
        // Clear smoothClock's DOM cache in case elements were recreated
        smoothClock.clearCache();
    });

    logger.info('Delta architecture initialized');
}

/**
 * Sync the clock to the current game state
 * @param {Object} stateOverride - Optional state to use instead of calling getGameStateSync
 */
function syncClockToGameState(stateOverride = null) {
    const state = stateOverride || getGameStateSync();
    logger.debug('syncClockToGameState called with state:', state ? 'exists' : 'null');

    // Check if we have actual time data - don't sync with defaults
    const hasCharacterTime = state?.character?.time_of_day !== undefined && state?.character?.time_of_day !== null;
    const hasTopLevelTime = state?.time_of_day !== undefined && state?.time_of_day !== null;

    if (hasCharacterTime) {
        let timeOfDay = state.character.time_of_day;
        logger.debug('Raw time_of_day from state:', timeOfDay);

        // Convert old hour format if needed
        if (timeOfDay < 24) {
            timeOfDay = timeOfDay * 60;
        }
        const currentDay = state.character.current_day || 1;

        // Sync smooth clock to state (force sync on initial load)
        smoothClock.syncFromBackend(timeOfDay, currentDay, true);
        logger.info(`Clock synced to game state: Day ${currentDay}, time ${timeOfDay} mins`);
    } else if (hasTopLevelTime) {
        // Alternative: state might have time_of_day at top level
        let timeOfDay = state.time_of_day;
        if (timeOfDay < 24) {
            timeOfDay = timeOfDay * 60;
        }
        const currentDay = state.current_day || 1;
        smoothClock.syncFromBackend(timeOfDay, currentDay, true);
        logger.info(`Clock synced from top-level state: Day ${currentDay}, time ${timeOfDay} mins`);
    } else {
        // No state available yet - DON'T sync with default values
        // This keeps hasInitialSync = false so clock displays "--:-- --" until real data
        logger.debug('No game state available yet, waiting for real data before clock sync');
    }
}

/**
 * Start the tick interval
 */
function startTicking() {
    if (tickIntervalId) {
        return;
    }

    accumulatedMinutes = 0;

    tickIntervalId = setInterval(() => {
        if (!isPaused) {
            tick();
        }
    }, TICK_INTERVAL_MS);

    logger.debug('Tick interval started (every 1 second, advancing 2.4 game minutes per tick)');
}

/**
 * Start the display update interval (smooth minute-by-minute UI updates)
 */
function startDisplayUpdates() {
    if (displayIntervalId) {
        return;
    }

    // Update display frequently for smooth visual updates (10 times per second)
    const DISPLAY_INTERVAL_MS = 100;

    displayIntervalId = setInterval(() => {
        if (!isPaused) {
            // Accumulate fractional minutes for display (2.4 minutes per second)
            displayAccumulatedMinutes += (GAME_MINUTES_PER_TICK * DISPLAY_INTERVAL_MS / 1000);

            // Update time display with interpolated time
            updateTimeDisplay();
        }
    }, DISPLAY_INTERVAL_MS);

    logger.debug(`Display update interval started (every ${DISPLAY_INTERVAL_MS}ms for smooth display)`);
}

/**
 * Main tick function - called every 1 second if not paused
 */
async function tick() {
    const state = getGameStateSync();

    if (!state.character) {
        return;
    }

    // Accumulate minutes (2.4 minutes per real second)
    accumulatedMinutes += GAME_MINUTES_PER_TICK;

    // Advance time when we've accumulated at least 1 full minute
    if (accumulatedMinutes >= 1.0) {
        const minutesToAdvance = Math.floor(accumulatedMinutes);
        accumulatedMinutes = accumulatedMinutes - minutesToAdvance;

        // Update local cache immediately
        advanceGameTime(state.character, minutesToAdvance);

        // Reset display accumulator to prevent drift
        displayAccumulatedMinutes = 0;
    }

    // Sync to backend every 5 seconds (not every second)
    backendSyncCounter++;
    if (backendSyncCounter >= 5) {
        backendSyncCounter = 0;
        await sendTimeUpdateToBackend(state.character);
    }
}

/**
 * Advance game time by specified minutes (frontend display only)
 * Backend calculates actual fatigue/hunger via effects system
 */
function advanceGameTime(character, minutes) {
    // Initialize time if not present (backwards compatibility)
    if (character.time_of_day === undefined) {
        character.time_of_day = 720; // Default to noon (720 minutes)
    }

    // Convert old hour-based values to minutes on first load
    if (character.time_of_day < 24) {
        // Old save: time_of_day was in hours (0-23)
        character.time_of_day = character.time_of_day * 60;
        logger.info(`Converted time_of_day from hours to minutes: ${character.time_of_day}`);
    }

    // Advance time of day (in minutes, 0-1439)
    character.time_of_day += minutes;

    // Handle day rollover
    const daysAdvanced = Math.floor(character.time_of_day / MINUTES_PER_DAY);
    if (daysAdvanced > 0) {
        character.current_day += daysAdvanced;
        character.time_of_day = character.time_of_day % MINUTES_PER_DAY;
        logger.info(`Day advanced to ${character.current_day}`);
    }

    // Fatigue and hunger are now calculated by backend effects system (not here!)
    // Remove old fatigue_counter and hunger_counter from character object
    delete character.fatigue_counter;
    delete character.hunger_counter;
}

/**
 * Send time update to backend (LEGACY - only used when USE_DELTA_SYSTEM is false)
 * Syncs the frontend time state to backend Go memory
 * Backend processes effects and returns updated state
 */
async function sendTimeUpdateToBackend(character) {
    // Don't run legacy sync if delta system is active
    if (USE_DELTA_SYSTEM) {
        logger.debug('Skipping legacy sendTimeUpdateToBackend - delta system active');
        return;
    }

    if (!gameAPI.initialized) {
        return;
    }

    try {
        // Send update via game action API
        const response = await gameAPI.sendAction('update_time', {
            time_of_day: character.time_of_day,
            current_day: character.current_day
        });

        // Update local state with backend's calculated values
        if (response && response.data) {
            let statsChanged = false;

            if (response.data.fatigue !== undefined && response.data.fatigue !== character.fatigue) {
                character.fatigue = response.data.fatigue;
                statsChanged = true;
                logger.debug(`Fatigue updated to ${character.fatigue}`);
            }
            if (response.data.hunger !== undefined && response.data.hunger !== character.hunger) {
                character.hunger = response.data.hunger;
                statsChanged = true;
                logger.debug(`Hunger updated to ${character.hunger}`);
            }
            if (response.data.hp !== undefined && response.data.hp !== character.hp) {
                character.hp = response.data.hp;
                statsChanged = true;
                logger.debug(`HP updated to ${character.hp}`);
            }
            if (response.data.active_effects) {
                character.active_effects = response.data.active_effects;
            }

            // Emit event on every time update (not just when stats change)
            // This ensures building open/close status updates even when fatigue/hunger don't change
            logger.debug('📡 Emitting gameStateChange event for time update...');
            eventBus.emit('gameStateChange', { character });
            document.dispatchEvent(new CustomEvent('gameStateChange', { detail: { character } }));

            // Update display when stats change
            if (statsChanged) {
                logger.info('⚡ Stats changed! Fatigue:', character.fatigue, 'Hunger:', character.hunger);
                logger.info('🔄 Calling updateCharacterDisplay directly...');
                await updateCharacterDisplay();
                logger.info('✅ Display update complete');
            }
        }
    } catch (error) {
        logger.error('Failed to sync time to backend:', error);
    }
}

/**
 * Toggle play/pause
 */
export function togglePause() {
    if (USE_DELTA_SYSTEM) {
        isPaused = smoothClock.togglePause();
    } else {
        isPaused = !isPaused;

        if (!isPaused) {
            // Reset accumulators when resuming
            accumulatedMinutes = 0;
            backendSyncCounter = 0;
        }

        // Reset display accumulator when toggling
        displayAccumulatedMinutes = 0;
    }

    updatePlayPauseButton();
    logger.info(isPaused ? 'Time paused' : 'Time playing');
}

/**
 * Get current pause state
 */
export function isPausedState() {
    if (USE_DELTA_SYSTEM) {
        return smoothClock.isPausedState();
    }
    return isPaused;
}

/**
 * Get current time including minutes (uses interpolated display time for smooth UI)
 * @returns {{hour: number, minute: number, synced: boolean}} Current game time
 */
export function getCurrentTime() {
    if (USE_DELTA_SYSTEM) {
        const time = smoothClock.getCurrentTime();
        return { hour: time.hours, minute: time.minutes, synced: time.synced };
    }

    const state = getGameStateSync();

    // Get actual time in minutes
    let actualTimeMinutes = state.character?.time_of_day || 720;

    // Convert old hour format if needed
    if (actualTimeMinutes < 24) {
        actualTimeMinutes = actualTimeMinutes * 60;
    }

    // Add fractional minutes for smooth display
    let displayTimeMinutes = actualTimeMinutes + displayAccumulatedMinutes;

    // Wrap at 1440 (midnight)
    if (displayTimeMinutes >= MINUTES_PER_DAY) {
        displayTimeMinutes = displayTimeMinutes % MINUTES_PER_DAY;
    }

    // Derive hour and minute
    const hour = Math.floor(displayTimeMinutes / 60) % 24;
    const minute = Math.floor(displayTimeMinutes % 60);

    return { hour, minute };
}

/**
 * Force pause (called when loading game, etc.)
 */
export function pause() {
    if (USE_DELTA_SYSTEM) {
        smoothClock.pause();
        isPaused = true;
    } else {
        if (!isPaused) {
            isPaused = true;
            displayAccumulatedMinutes = 0;
        }
    }
    updatePlayPauseButton();
    logger.info('Time force-paused');
}

/**
 * Force play (optional - for auto-play on actions)
 */
export function play() {
    if (USE_DELTA_SYSTEM) {
        smoothClock.unpause();
        isPaused = false;
    } else {
        if (isPaused) {
            isPaused = false;
            accumulatedMinutes = 0;
            backendSyncCounter = 0;
            displayAccumulatedMinutes = 0;
        }
    }
    updatePlayPauseButton();
    logger.info('Time force-played');
}

/**
 * Update the play/pause button appearance
 */
function updatePlayPauseButton() {
    const playButton = document.getElementById('time-play-button');
    const pauseButton = document.getElementById('time-pause-button');

    if (!playButton || !pauseButton) {
        return;
    }

    if (isPaused) {
        playButton.style.display = '';
        pauseButton.style.display = 'none';
    } else {
        playButton.style.display = 'none';
        pauseButton.style.display = '';
    }
}

/**
 * Cleanup on page unload
 */
export function cleanupTimeClock() {
    if (USE_DELTA_SYSTEM) {
        smoothClock.stop();
        tickManager.stop();
        deltaApplier.clearCache();
        logger.debug('Delta time systems stopped');
    } else {
        if (tickIntervalId) {
            clearInterval(tickIntervalId);
            tickIntervalId = null;
        }
        if (displayIntervalId) {
            clearInterval(displayIntervalId);
            displayIntervalId = null;
        }
        logger.debug('Legacy time clock stopped');
    }
}

// Export for global access (onclick in HTML)
window.timeClock = {
    togglePause,
    pause,
    play,
    getCurrentTime
};

logger.debug('Time clock module loaded');
