/**
 * Simple Event Bus for decoupled communication between modules
 *
 * Allows modules to communicate without direct dependencies.
 * Useful for breaking circular dependencies.
 *
 * @module lib/events
 */

import { logger } from './logger.js';

class EventBus {
    constructor() {
        this.listeners = new Map();
    }

    /**
     * Subscribe to an event
     * @param {string} eventName - Event name
     * @param {Function} callback - Callback function
     * @returns {Function} Unsubscribe function
     */
    on(eventName, callback) {
        if (!this.listeners.has(eventName)) {
            this.listeners.set(eventName, []);
        }

        this.listeners.get(eventName).push(callback);

        logger.debug(`Subscribed to event: ${eventName}`);

        // Return unsubscribe function
        return () => this.off(eventName, callback);
    }

    /**
     * Unsubscribe from an event
     * @param {string} eventName - Event name
     * @param {Function} callback - Callback function to remove
     */
    off(eventName, callback) {
        if (this.listeners.has(eventName)) {
            const callbacks = this.listeners.get(eventName);
            const index = callbacks.indexOf(callback);
            if (index > -1) {
                callbacks.splice(index, 1);
                logger.debug(`Unsubscribed from event: ${eventName}`);
            }

            // Clean up empty listener arrays
            if (callbacks.length === 0) {
                this.listeners.delete(eventName);
            }
        }
    }

    /**
     * Emit an event
     * @param {string} eventName - Event name
     * @param {*} data - Data to pass to listeners
     */
    emit(eventName, data) {
        logger.debug(`Event emitted: ${eventName}`, data);

        if (this.listeners.has(eventName)) {
            // Create a copy of listeners array to avoid issues if listeners modify the array
            const callbacks = [...this.listeners.get(eventName)];

            callbacks.forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    logger.error(`Error in ${eventName} event handler:`, error);
                }
            });
        }
    }

}

// Export singleton instance
export const eventBus = new EventBus();

logger.debug('Event bus initialized');
