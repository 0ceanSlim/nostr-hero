/**
 * Discover Page Entry Point
 *
 * Character preview page bundle - imports and initializes preview systems.
 * This replaces the individual script tags in discover.html.
 */

// Core libraries
import { logger } from '../lib/logger.js';
import '../lib/session.js'; // Auto-initializes as window.sessionManager

// Systems
import { themeManager } from '../systems/themeManager.js';
import '../systems/auth.js'; // Auto-initializes authentication

logger.info('🔮 Discover page bundle loaded');
