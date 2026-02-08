/**
 * CODEX Home Page Entry Point
 */

// Import styles
import './styles.css';

// Detect and apply theme from main game
const savedTheme = localStorage.getItem('theme') || 'dark';
document.documentElement.setAttribute('data-theme', savedTheme);
console.log(`🎯 CODEX Home loaded (Theme: ${savedTheme})`);
