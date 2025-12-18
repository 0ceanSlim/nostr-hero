/**
 * Settings Page Entry Point
 *
 * Settings page bundle - imports and initializes settings systems.
 * This replaces the individual script tags in settings.html.
 */

// Core libraries
import { logger } from '../lib/logger.js';
import '../lib/session.js'; // Auto-initializes as window.sessionManager

// Systems
import { themeManager } from '../systems/themeManager.js';
import { relayManager } from '../systems/relayManager.js';
import '../systems/auth.js'; // Auto-initializes authentication

logger.info('⚙️ Settings page bundle loaded');
