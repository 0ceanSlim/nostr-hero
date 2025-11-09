/**
 * Relay Manager
 * Manages Nostr relay connections and persistence
 */

class RelayManager {
  constructor() {
    this.relays = this.loadRelays();
    this.init();
  }

  /**
   * Initialize relay manager with default relays
   */
  init() {
    // If no relays stored, add defaults
    if (this.relays.length === 0) {
      this.relays = [
        { url: 'wss://relay.damus.io', connected: false },
        { url: 'wss://nos.lol', connected: false },
        { url: 'wss://relay.nostr.band', connected: false },
        { url: 'wss://nostr.wine', connected: false }
      ];
      this.saveRelays();
    }
  }

  /**
   * Get all relays
   * @returns {Array} List of relay objects
   */
  getRelays() {
    return this.relays;
  }

  /**
   * Add a new relay
   * @param {string} url - Relay WebSocket URL
   * @returns {boolean} Success status
   */
  addRelay(url) {
    // Check if relay already exists
    if (this.relays.some(r => r.url === url)) {
      return false;
    }

    this.relays.push({
      url: url,
      connected: false
    });

    this.saveRelays();
    this.dispatchRelayChange();
    return true;
  }

  /**
   * Remove a relay
   * @param {string} url - Relay WebSocket URL
   */
  removeRelay(url) {
    this.relays = this.relays.filter(r => r.url !== url);
    this.saveRelays();
    this.dispatchRelayChange();
  }

  /**
   * Test relay connection
   * @param {string} url - Relay WebSocket URL
   * @returns {Promise<boolean>} Connection success
   */
  async testRelay(url) {
    return new Promise((resolve) => {
      const timeout = setTimeout(() => {
        ws.close();
        this.updateRelayStatus(url, false);
        resolve(false);
      }, 5000); // 5 second timeout

      try {
        const ws = new WebSocket(url);

        ws.onopen = () => {
          clearTimeout(timeout);
          this.updateRelayStatus(url, true);
          ws.close();
          resolve(true);
        };

        ws.onerror = () => {
          clearTimeout(timeout);
          this.updateRelayStatus(url, false);
          ws.close();
          resolve(false);
        };

      } catch (error) {
        clearTimeout(timeout);
        this.updateRelayStatus(url, false);
        resolve(false);
      }
    });
  }

  /**
   * Update relay connection status
   * @param {string} url - Relay URL
   * @param {boolean} connected - Connection status
   */
  updateRelayStatus(url, connected) {
    const relay = this.relays.find(r => r.url === url);
    if (relay) {
      relay.connected = connected;
      this.saveRelays();
      this.dispatchRelayChange();
    }
  }

  /**
   * Save relays to localStorage
   */
  saveRelays() {
    try {
      localStorage.setItem('nostr-relays', JSON.stringify(this.relays));
    } catch (error) {
      console.error('Failed to save relays:', error);
    }
  }

  /**
   * Load relays from localStorage
   * @returns {Array} Relay list
   */
  loadRelays() {
    try {
      const stored = localStorage.getItem('nostr-relays');
      if (stored) {
        return JSON.parse(stored);
      }
    } catch (error) {
      console.error('Failed to load relays:', error);
    }
    return [];
  }

  /**
   * Dispatch relay change event
   */
  dispatchRelayChange() {
    window.dispatchEvent(new CustomEvent('relays-changed', {
      detail: {
        relays: this.relays
      }
    }));
  }

  /**
   * Get connected relays only
   * @returns {Array} List of connected relay URLs
   */
  getConnectedRelays() {
    return this.relays
      .filter(r => r.connected)
      .map(r => r.url);
  }

  /**
   * Test all relays
   * @returns {Promise<void>}
   */
  async testAllRelays() {
    const promises = this.relays.map(relay => this.testRelay(relay.url));
    await Promise.all(promises);
  }

  /**
   * Reset to default relays
   */
  resetToDefaults() {
    this.relays = [
      { url: 'wss://relay.damus.io', connected: false },
      { url: 'wss://nos.lol', connected: false },
      { url: 'wss://relay.nostr.band', connected: false },
      { url: 'wss://nostr.wine', connected: false }
    ];
    this.saveRelays();
    this.dispatchRelayChange();
  }
}

// Initialize global relay manager
if (typeof window !== 'undefined') {
  window.relayManager = new RelayManager();
}
