/**
 * Wait Modal System
 * Handles the wait functionality with time slider
 */

import { logger } from '../lib/logger.js';
import { gameAPI } from '../lib/api.js';
import { getGameStateSync, refreshGameState } from '../state/gameState.js';
import { updateAllDisplays } from '../ui/displayCoordinator.js';
import { showMessage } from '../ui/messaging.js';

let currentWaitHours = 1;
let updateIntervalId = null;

/**
 * Open the wait modal
 */
export function openWaitModal() {
    logger.debug('Opening wait modal');

    const modal = document.getElementById('wait-modal');
    if (!modal) {
        logger.error('Wait modal element not found');
        return;
    }

    // Reset to default 1 hour
    currentWaitHours = 1;
    const slider = document.getElementById('wait-slider');
    if (slider) {
        slider.value = 1;
    }

    // Update display with initial values
    updateWaitDisplay(1);

    // Start updating the display every 100ms to sync with live time
    updateIntervalId = setInterval(() => {
        updateWaitDisplay(currentWaitHours);
    }, 100);

    // Show modal
    modal.classList.remove('hidden');
}

/**
 * Close the wait modal
 */
export function closeWaitModal() {
    logger.debug('Closing wait modal');

    const modal = document.getElementById('wait-modal');
    if (modal) {
        modal.classList.add('hidden');
    }

    // Stop the update interval
    if (updateIntervalId) {
        clearInterval(updateIntervalId);
        updateIntervalId = null;
    }
}

/**
 * Update the wait display based on slider value
 */
export function updateWaitDisplay(hours) {
    currentWaitHours = parseInt(hours);

    // Update hours display
    const hoursDisplay = document.getElementById('wait-hours-display');
    const hoursPlural = document.getElementById('wait-hours-plural');
    if (hoursDisplay) {
        hoursDisplay.textContent = currentWaitHours;
    }
    if (hoursPlural) {
        hoursPlural.textContent = currentWaitHours === 1 ? '' : 's';
    }

    // Calculate target time - use same source as time display for consistency
    let currentTimeMinutes;
    if (window.timeClock && window.timeClock.getCurrentTime) {
        const currentTime = window.timeClock.getCurrentTime();
        currentTimeMinutes = (currentTime.hour * 60) + currentTime.minute;
    } else {
        // Fallback to game state if time clock not available
        const state = getGameStateSync();
        currentTimeMinutes = state.time_of_day || state.character?.time_of_day || 720;
    }

    const targetTimeMinutes = (currentTimeMinutes + (currentWaitHours * 60)) % 1440;

    const targetHours = Math.floor(targetTimeMinutes / 60);
    const targetMins = targetTimeMinutes % 60;
    const targetPeriod = targetHours >= 12 ? 'PM' : 'AM';
    const targetHours12 = targetHours % 12 || 12;

    const targetTimeDisplay = document.getElementById('wait-target-time');
    if (targetTimeDisplay) {
        targetTimeDisplay.textContent = `${targetHours12}:${targetMins.toString().padStart(2, '0')} ${targetPeriod}`;
    }

    // Calculate fatigue and hunger changes
    // Fatigue increases by 1 per hour waited
    const fatigueIncrease = currentWaitHours;

    // Hunger increases by 1 every 3 hours (so 1-2 hours = 0, 3-5 hours = 1, 6 hours = 2)
    const hungerIncrease = Math.floor(currentWaitHours / 3);

    const fatigueChange = document.getElementById('wait-fatigue-change');
    const hungerChange = document.getElementById('wait-hunger-change');

    if (fatigueChange) {
        fatigueChange.textContent = `+${fatigueIncrease}`;
        fatigueChange.className = fatigueIncrease > 0 ? 'text-red-400' : 'text-gray-400';
    }

    if (hungerChange) {
        hungerChange.textContent = hungerIncrease > 0 ? `+${hungerIncrease}` : '-';
        hungerChange.className = hungerIncrease > 0 ? 'text-orange-400' : 'text-gray-400';
    }
}

/**
 * Confirm and execute the wait action
 */
export async function confirmWait() {
    logger.debug(`Confirming wait for ${currentWaitHours} hours`);

    try {
        const result = await gameAPI.sendAction('wait', {
            hours: currentWaitHours
        });

        if (result.success) {
            showMessage(result.message || `You waited ${currentWaitHours} hour${currentWaitHours === 1 ? '' : 's'}.`, 'success');
            await refreshGameState();
            await updateAllDisplays();
            closeWaitModal();
        } else {
            showMessage(result.error || 'Failed to wait', 'error');
        }
    } catch (error) {
        logger.error('Failed to execute wait:', error);
        showMessage('❌ Failed to wait', 'error');
    }
}

// Export functions to window for onclick handlers
if (typeof window !== 'undefined') {
    window.openWaitModal = openWaitModal;
    window.closeWaitModal = closeWaitModal;
    window.updateWaitDisplay = updateWaitDisplay;
    window.confirmWait = confirmWait;
}
