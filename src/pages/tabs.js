/**
 * Tab System Module
 *
 * Handles tab switching and URL hash routing for the homepage.
 * Manages tab navigation with hash-based routing.
 *
 * @module pages/tabs
 */

import { logger } from '../lib/logger.js';

/**
 * Tab Manager Class
 * Manages tab switching and URL hash routing
 */
class TabManager {
    constructor() {
        this.tabs = ['home', 'updates', 'roadmap', 'about', 'contribute', 'wiki'];
        this.currentTab = 'home';
        this.init();
    }

    /**
     * Initialize tab manager
     */
    init() {
        // Load tab from URL hash if present
        const hash = window.location.hash.slice(1);
        if (hash && this.tabs.includes(hash)) {
            this.currentTab = hash;
        }

        // Show initial tab
        this.showTab(this.currentTab);

        // Listen for hash changes
        window.addEventListener('hashchange', () => {
            const hash = window.location.hash.slice(1);
            if (hash && this.tabs.includes(hash)) {
                this.showTab(hash);
            }
        });

        // Add click handlers to tab buttons
        this.tabs.forEach(tabName => {
            const button = document.getElementById(`tab-btn-${tabName}`);
            if (button) {
                button.addEventListener('click', () => {
                    this.switchTab(tabName);
                });
            }
        });
    }

    /**
     * Switch to a specific tab
     * @param {string} tabName - Name of tab to show
     */
    switchTab(tabName) {
        if (!this.tabs.includes(tabName)) {
            logger.error(`Tab "${tabName}" not found`);
            return;
        }

        // Update URL hash
        window.location.hash = tabName;

        // Show tab
        this.showTab(tabName);
    }

    /**
     * Show a specific tab
     * @param {string} tabName - Name of tab to show
     */
    showTab(tabName) {
        this.currentTab = tabName;

        // Hide all tab content
        this.tabs.forEach(tab => {
            const content = document.getElementById(`tab-${tab}`);
            const button = document.getElementById(`tab-btn-${tab}`);

            if (content) {
                content.classList.add('hidden');
            }

            if (button) {
                button.classList.remove('tab-active');
                button.classList.add('tab-inactive');
            }
        });

        // Show active tab content
        const activeContent = document.getElementById(`tab-${tabName}`);
        const activeButton = document.getElementById(`tab-btn-${tabName}`);

        if (activeContent) {
            activeContent.classList.remove('hidden');
        }

        if (activeButton) {
            activeButton.classList.remove('tab-inactive');
            activeButton.classList.add('tab-active');
        }

        // Dispatch tab change event
        window.dispatchEvent(new CustomEvent('tab-changed', {
            detail: { tab: tabName }
        }));

        // Scroll to top of page
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }

    /**
     * Get current active tab
     * @returns {string} Current tab name
     */
    getCurrentTab() {
        return this.currentTab;
    }
}

// Create and export tab manager instance
export const tabManager = new TabManager();

// Initialize tab manager when DOM is ready
if (typeof window !== 'undefined') {
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            // Already initialized in constructor
        });
    }

    // Make globally available for compatibility
    window.tabManager = tabManager;
}

logger.debug('Tab system loaded');
