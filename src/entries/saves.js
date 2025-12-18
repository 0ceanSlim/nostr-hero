/**
 * Saves Page Entry Point
 *
 * Loads dependencies for the save selection page.
 */

import { logger } from '../lib/logger.js';
import '../lib/session.js';
import { themeManager } from '../systems/themeManager.js';
import '../systems/auth.js';

// Simple message display function for saves page
window.showMessage = function showMessage(text, type = 'info', duration = 5000) {
  logger.info(`[${type.toUpperCase()}] ${text}`);

  // Could add a toast notification here in the future
  // For now, just log to console
};

// Optional loading modal functions (if they exist elsewhere)
window.hideLoadingModal = function hideLoadingModal() {
  const loadingModal = document.getElementById('loading-modal');
  if (loadingModal) {
    loadingModal.style.display = 'none';
  }
};

logger.info('Saves page initialized');
