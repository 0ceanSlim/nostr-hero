/**
 * Theme Manager
 * Handles theme switching and persistence
 */

class ThemeManager {
  constructor() {
    this.themes = ['dark', 'light', 'midnight', 'lava', 'solar', 'matrix'];
    this.currentTheme = this.getStoredTheme() || 'dark';
    this.init();
  }

  /**
   * Initialize theme manager
   */
  init() {
    this.applyTheme(this.currentTheme);
    this.dispatchThemeChange();
  }

  /**
   * Get all available themes
   * @returns {Array<string>} List of theme names
   */
  getThemes() {
    return this.themes;
  }

  /**
   * Get current active theme
   * @returns {string} Current theme name
   */
  getCurrentTheme() {
    return this.currentTheme;
  }

  /**
   * Set theme
   * @param {string} themeName - Name of theme to apply
   */
  setTheme(themeName) {
    if (!this.themes.includes(themeName)) {
      console.error(`Theme "${themeName}" not found`);
      return;
    }

    this.currentTheme = themeName;
    this.applyTheme(themeName);
    this.storeTheme(themeName);
    this.dispatchThemeChange();
  }

  /**
   * Apply theme to document
   * @param {string} themeName - Theme name
   */
  applyTheme(themeName) {
    const root = document.documentElement;
    root.setAttribute('data-theme', themeName);
  }

  /**
   * Store theme preference in localStorage
   * @param {string} themeName - Theme name
   */
  storeTheme(themeName) {
    try {
      localStorage.setItem('preferred-theme', themeName);
    } catch (error) {
      console.error('Failed to store theme preference:', error);
    }
  }

  /**
   * Get stored theme from localStorage
   * @returns {string|null} Stored theme name or null
   */
  getStoredTheme() {
    try {
      return localStorage.getItem('preferred-theme');
    } catch (error) {
      console.error('Failed to get stored theme:', error);
      return null;
    }
  }

  /**
   * Dispatch theme change event
   */
  dispatchThemeChange() {
    window.dispatchEvent(new CustomEvent('theme-changed', {
      detail: {
        theme: this.currentTheme
      }
    }));
  }

  /**
   * Get theme preview colors
   * @param {string} themeName - Theme name
   * @returns {Object} Color preview object
   */
  getThemePreview(themeName) {
    const previews = {
      dark: {
        bg: 'rgb(16, 16, 16)',
        text: 'rgb(255, 255, 255)',
        highlight: 'rgb(149, 40, 244)'
      },
      light: {
        bg: 'rgb(230, 230, 230)',
        text: 'rgb(0, 0, 0)',
        highlight: 'rgb(149, 40, 244)'
      },
      midnight: {
        bg: 'rgb(25, 24, 48)',
        text: 'rgb(177, 174, 255)',
        highlight: 'rgb(250, 208, 0)'
      },
      lava: {
        bg: 'rgb(57, 0, 0)',
        text: 'rgb(255, 90, 90)',
        highlight: 'rgb(255, 114, 154)'
      },
      solar: {
        bg: 'rgb(0, 43, 54)',
        text: 'rgb(131, 148, 150)',
        highlight: 'rgb(38, 139, 210)'
      },
      matrix: {
        bg: 'rgb(0, 0, 0)',
        text: 'rgb(0, 255, 65)',
        highlight: 'rgb(0, 255, 100)'
      }
    };

    return previews[themeName] || previews.dark;
  }

  /**
   * Toggle between themes
   */
  toggleTheme() {
    const currentIndex = this.themes.indexOf(this.currentTheme);
    const nextIndex = (currentIndex + 1) % this.themes.length;
    this.setTheme(this.themes[nextIndex]);
  }
}

// Initialize global theme manager
if (typeof window !== 'undefined') {
  window.themeManager = new ThemeManager();
}
