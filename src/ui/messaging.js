/**
 * Messaging UI Module
 *
 * Handles game log messages and action text display.
 * Provides color-coded messages and auto-scrolling.
 *
 * @module ui/messaging
 */

import { logger } from '../lib/logger.js';

/**
 * Add a message to the game text log
 * Messages are appended to preserve history
 * @param {string} message - Message text to display
 */
export function addGameLog(message) {
    const gameText = document.getElementById('game-text');
    if (!gameText) return;

    // Create new log entry with styled border
    const logEntry = document.createElement('p');
    logEntry.className = 'border-l-2 border-gray-600 pl-2';
    logEntry.textContent = message;

    // Append to log
    gameText.appendChild(logEntry);

    // Auto-scroll to bottom to show latest message
    const textContainer = gameText.parentElement;
    if (textContainer) {
        textContainer.scrollTop = textContainer.scrollHeight;
    }
}

/**
 * Show action text in the game log with color coding
 * @param {string} text - The message to display
 * @param {string} color - Color: 'purple', 'white', 'red', 'green', 'yellow', 'blue'
 * @param {number} duration - Ignored (kept for API compatibility)
 */
export function showActionText(text, color = 'white', duration = 0) {
    const gameText = document.getElementById('game-text');
    if (!gameText) return;

    // Color mapping
    const colors = {
        'purple': '#a78bfa',   // Welcome messages
        'white': '#ffffff',    // Descriptions, neutral info
        'red': '#ef4444',      // Errors
        'green': '#22c55e',    // Success
        'yellow': '#eab308',   // Warnings
        'blue': '#3b82f6'      // Info
    };

    // Create new log entry with styled border and color
    const logEntry = document.createElement('p');
    logEntry.className = 'border-l-2 pl-2';
    logEntry.style.borderColor = colors[color] || colors['white'];
    logEntry.style.color = colors[color] || colors['white'];
    logEntry.textContent = text;

    // Append to log
    gameText.appendChild(logEntry);

    // Auto-scroll to bottom to show latest message
    const textContainer = gameText.parentElement;
    if (textContainer) {
        textContainer.scrollTop = textContainer.scrollHeight;
    }
}

/**
 * Legacy showMessage function - now uses showActionText
 * @param {string} text - Message text
 * @param {string} type - Message type: 'error', 'success', 'warning', 'info'
 * @param {number} duration - Duration in ms (ignored)
 */
export function showMessage(text, type = 'info', duration = 5000) {
    const colorMap = {
        'error': 'red',
        'success': 'green',
        'warning': 'yellow',
        'info': 'blue'
    };
    showActionText(text, colorMap[type] || 'white', duration);
}

logger.debug('Messaging UI module loaded');
