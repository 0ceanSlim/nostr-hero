/**
 * 144x Real-Time Clock System
 *
 * Ticks game time at 144x real-time speed.
 * Processes every 1 second: time advances 2.4 game minutes per real second.
 * Tracks time in minutes (0-1439 per day) for smooth updates.
 */

import { logger } from '../lib/logger.js';
import { getGameStateSync } from '../state/gameState.js';
import { updateTimeDisplay } from '../ui/timeDisplay.js';
import { updateCharacterDisplay } from '../ui/characterDisplay.js';
import { gameAPI } from '../lib/api.js';
import { eventBus } from '../lib/events.js';

// Constants
const TICK_INTERVAL_MS = 1000; // Process every 1 real second
const TIME_MULTIPLIER = 144.0;  // 144x real-time speed
const GAME_MINUTES_PER_TICK = (TICK_INTERVAL_MS / 1000) * TIME_MULTIPLIER / 60; // = 2.4 minutes
const MINUTES_PER_DAY = 1440; // 24 hours * 60 minutes

// Thresholds in minutes
const FATIGUE_THRESHOLD_MINUTES = 240; // 4 hours
const HUNGER_THRESHOLD_NORMAL_MINUTES = 360; // 6 hours
const HUNGER_THRESHOLD_HUNGRY_MINUTES = 720; // 12 hours

// State
let tickIntervalId = null;
let displayIntervalId = null; // Separate interval for smooth display updates
let isPaused = true; // Start paused
let accumulatedMinutes = 0; // Fractional minutes not yet applied (for actual game time)
let displayAccumulatedMinutes = 0; // Fractional minutes for smooth display
let backendSyncCounter = 0; // Count seconds until next backend sync (every 5 seconds)

/**
 * Initialize the clock system
 */
export function initTimeClock() {
    logger.info('Initializing 144x time clock');

    // Start the tick loop
    startTicking();

    // Start the display update loop (updates UI smoothly)
    startDisplayUpdates();

    // Update button state
    updatePlayPauseButton();

    logger.debug('Time clock initialized (paused)');
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
 * Send time update to backend
 * Syncs the frontend time state to backend Go memory
 * Backend processes effects and returns updated state
 */
async function sendTimeUpdateToBackend(character) {
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
    isPaused = !isPaused;

    if (!isPaused) {
        // Reset accumulators when resuming
        accumulatedMinutes = 0;
        backendSyncCounter = 0;
    }

    // Reset display accumulator when toggling
    displayAccumulatedMinutes = 0;

    updatePlayPauseButton();

    logger.info(isPaused ? 'Time paused' : 'Time playing');
}

/**
 * Get current pause state
 */
export function isPausedState() {
    return isPaused;
}

/**
 * Get current time including minutes (uses interpolated display time for smooth UI)
 * @returns {{hour: number, minute: number}} Current game time
 */
export function getCurrentTime() {
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
    if (!isPaused) {
        isPaused = true;
        // Reset display accumulator when pausing
        displayAccumulatedMinutes = 0;
        updatePlayPauseButton();
        logger.info('Time force-paused');
    }
}

/**
 * Force play (optional - for auto-play on actions)
 */
export function play() {
    if (isPaused) {
        isPaused = false;
        accumulatedMinutes = 0;
        backendSyncCounter = 0;
        displayAccumulatedMinutes = 0;
        updatePlayPauseButton();
        logger.info('Time force-played');
    }
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
    if (tickIntervalId) {
        clearInterval(tickIntervalId);
        tickIntervalId = null;
    }
    if (displayIntervalId) {
        clearInterval(displayIntervalId);
        displayIntervalId = null;
    }
    logger.debug('Time clock stopped');
}

// Export for global access (onclick in HTML)
window.timeClock = {
    togglePause,
    pause,
    play,
    getCurrentTime
};

logger.debug('Time clock module loaded');
