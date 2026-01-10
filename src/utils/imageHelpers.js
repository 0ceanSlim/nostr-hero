/**
 * Image Helper Utilities
 * Provides reusable functions for handling image loading with fallbacks
 */

/**
 * Create an item image element with automatic fallback to unknown.png
 * @param {string} itemId - The item ID
 * @param {object} options - Optional configuration
 * @param {string} options.className - CSS classes for the image
 * @param {string} options.alt - Alt text for the image
 * @returns {HTMLImageElement} - The configured image element
 */
export function createItemImage(itemId, options = {}) {
    const img = document.createElement('img');
    img.src = `/res/img/items/${itemId}.png`;
    img.alt = options.alt || itemId;
    img.className = options.className || 'w-full h-full object-contain';
    img.style.imageRendering = 'pixelated';

    // Add fallback handler that only triggers once
    img.onerror = function() {
        if (!this.dataset.fallbackAttempted) {
            this.dataset.fallbackAttempted = 'true';
            this.src = '/res/img/items/unknown.png';
        }
    };

    return img;
}

/**
 * Set up image fallback on an existing image element
 * @param {HTMLImageElement} img - The image element
 */
export function setupImageFallback(img) {
    img.onerror = function() {
        if (!this.dataset.fallbackAttempted) {
            this.dataset.fallbackAttempted = 'true';
            this.src = '/res/img/items/unknown.png';
        }
    };
}
