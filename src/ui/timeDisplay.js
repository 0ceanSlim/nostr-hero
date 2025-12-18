/**
 * Time Display UI Module
 *
 * Handles the display of in-game time and day counter.
 * Shows time-of-day images and formatted time text.
 *
 * @module ui/timeDisplay
 */

import { logger } from '../lib/logger.js';
import { getGameStateSync } from '../state/gameState.js';

/**
 * Update time of day display
 */
export function updateTimeDisplay() {
    const state = getGameStateSync();
    const timeOfDay = state.character?.time_of_day !== undefined ? state.character.time_of_day : 12;
    const currentDay = state.character?.current_day || 1;

    // Map time (0-23) to 12 PNG filenames (each image covers 2 hours)
    const timeImages = [
        '00-midnight.png',     // 0-1 (12 AM - 1 AM)
        '01-twilight.png',     // 2-3
        '02-witching.png',     // 4-5
        '03-dawn.png',         // 6-7
        '04-morning.png',      // 8-9
        '05-latemorning.png',  // 10-11
        '06-highnoon.png',     // 12-13 (12 PM - 1 PM)
        '07-midday.png',       // 14-15
        '08-afternoon.png',    // 16-17
        '09-golden.png',       // 18-19
        '10-dusk.png',         // 20-21
        '11-evening.png'       // 22-23
    ];

    // Calculate which image to use (divide by 2 since each image covers 2 hours)
    const imageIndex = Math.floor(timeOfDay / 2);

    // Update time image
    const timeImage = document.getElementById('time-of-day-image');
    if (timeImage) {
        const imageName = timeImages[imageIndex] || timeImages[6]; // Default to noon if invalid
        timeImage.src = `/res/img/time/${imageName}`;
        timeImage.alt = `Time: ${formatTime(timeOfDay)}`;
    }

    // Update time text (AM/PM format)
    const timeText = document.getElementById('time-of-day-text');
    if (timeText) {
        timeText.textContent = formatTime(timeOfDay);
    }

    // Update day counter
    const dayCounter = document.getElementById('day-counter');
    if (dayCounter) {
        dayCounter.textContent = `Day ${currentDay}`;
    }
}

/**
 * Helper function to format time in AM/PM format
 * @param {number} hour - Hour in 24-hour format (0-23)
 * @returns {string} Formatted time string
 */
export function formatTime(hour) {
    if (hour === 0) return '12 AM';
    if (hour < 12) return `${hour} AM`;
    if (hour === 12) return '12 PM';
    return `${hour - 12} PM`;
}

logger.debug('Time display module loaded');
