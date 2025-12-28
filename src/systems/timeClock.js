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
 * Advance game time by specified minutes
 */
function advanceGameTime(character, minutes) {
    // Initialize counters if not present (backwards compatibility)
    if (character.time_of_day === undefined) {
        character.time_of_day = 720; // Default to noon (720 minutes)
    }
    if (character.fatigue_counter === undefined) {
        character.fatigue_counter = 0;
    }
    if (character.hunger_counter === undefined) {
        character.hunger_counter = 0;
    }

    // Convert old hour-based values to minutes on first load
    if (character.time_of_day < 24) {
        // Old save: time_of_day was in hours (0-23)
        character.time_of_day = character.time_of_day * 60;
        logger.info(`Converted time_of_day from hours to minutes: ${character.time_of_day}`);
    }
    if (character.fatigue_counter < 24 && character.fatigue_counter !== 0) {
        // Old save: fatigue_counter was in hours
        character.fatigue_counter = character.fatigue_counter * 60;
        logger.info(`Converted fatigue_counter from hours to minutes: ${character.fatigue_counter}`);
    }
    if (character.hunger_counter < 24 && character.hunger_counter !== 0) {
        // Old save: hunger_counter was in hours
        character.hunger_counter = character.hunger_counter * 60;
        logger.info(`Converted hunger_counter from hours to minutes: ${character.hunger_counter}`);
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

    // Update fatigue counter (every 240 minutes = 4 hours)
    character.fatigue_counter += minutes;
    if (character.fatigue_counter >= FATIGUE_THRESHOLD_MINUTES) {
        const fatigueIncreases = Math.floor(character.fatigue_counter / FATIGUE_THRESHOLD_MINUTES);
        character.fatigue = Math.min(10, character.fatigue + fatigueIncreases);
        character.fatigue_counter = character.fatigue_counter % FATIGUE_THRESHOLD_MINUTES;

        // Update character display when stats change
        updateCharacterDisplay();
        logger.debug(`Fatigue increased to ${character.fatigue}`);
    }

    // Update hunger counter (every 360 minutes normally, 720 if hungry)
    const hungerThreshold = character.hunger <= 1 ? HUNGER_THRESHOLD_HUNGRY_MINUTES : HUNGER_THRESHOLD_NORMAL_MINUTES;
    character.hunger_counter += minutes;

    if (character.hunger_counter >= hungerThreshold) {
        const hungerDecreases = Math.floor(character.hunger_counter / hungerThreshold);
        character.hunger = Math.max(0, character.hunger - hungerDecreases);
        character.hunger_counter = character.hunger_counter % hungerThreshold;

        // Update character display when stats change
        updateCharacterDisplay();
        logger.debug(`Hunger decreased to ${character.hunger}`);
    }
}

/**
 * Send time update to backend
 * Syncs the frontend time state to backend Go memory
 */
async function sendTimeUpdateToBackend(character) {
    if (!gameAPI.initialized) {
        return;
    }

    try {
        // Send update via game action API
        await gameAPI.sendAction('update_time', {
            time_of_day: character.time_of_day,
            fatigue_counter: character.fatigue_counter,
            hunger_counter: character.hunger_counter,
            fatigue: character.fatigue,
            hunger: character.hunger,
            current_day: character.current_day
        });
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
