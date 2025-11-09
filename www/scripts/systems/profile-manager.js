/**
 * Profile Manager
 * Fetches and manages Nostr profile metadata (kind 0 events)
 */

class ProfileManager {
  constructor() {
    this.profile = null;
    this.pubkey = null;
    this.npub = null;
    this.relays = [
      'wss://relay.damus.io',
      'wss://nos.lol',
      'wss://relay.nostr.band',
      'wss://nostr.wine'
    ];
  }

  /**
   * Initialize profile manager with user's pubkey
   * @param {string} pubkey - Hex public key
   * @param {string} npub - Bech32 npub
   */
  async init(pubkey, npub) {
    this.pubkey = pubkey;
    this.npub = npub;

    // Check cache first
    const cached = this.getCachedProfile(pubkey);
    if (cached) {
      this.profile = cached;
      this.dispatchProfileUpdate();
      return cached;
    }

    // Fetch from relays
    try {
      const profile = await this.fetchProfile(pubkey);
      if (profile) {
        this.profile = { ...profile, npub, pubkey };
        this.cacheProfile(pubkey, this.profile);
        this.dispatchProfileUpdate();
        return this.profile;
      }
    } catch (error) {
      console.error('Failed to fetch profile:', error);
    }

    // Return minimal profile if fetch fails
    this.profile = {
      npub,
      pubkey,
      display_name: npub.slice(0, 10) + '...',
      name: null,
      picture: null,
      about: null
    };
    this.dispatchProfileUpdate();
    return this.profile;
  }

  /**
   * Fetch kind 0 profile event from Nostr relays
   * @param {string} pubkey - Hex public key
   * @returns {Promise<Object>} Profile metadata
   */
  async fetchProfile(pubkey) {
    return new Promise((resolve) => {
      const timeout = setTimeout(() => {
        resolve(null);
      }, 5000); // 5 second timeout

      let resolved = false;
      let connectedRelays = 0;
      const targetRelays = this.relays.length;

      this.relays.forEach(relayUrl => {
        try {
          const ws = new WebSocket(relayUrl);

          ws.onopen = () => {
            connectedRelays++;
            // Subscribe to kind 0 events for this pubkey
            const subscription = {
              id: `profile-${pubkey.slice(0, 8)}`,
              kinds: [0],
              authors: [pubkey],
              limit: 1
            };
            ws.send(JSON.stringify(['REQ', subscription.id, {
              kinds: [0],
              authors: [pubkey],
              limit: 1
            }]));
          };

          ws.onmessage = (event) => {
            try {
              const data = JSON.parse(event.data);
              if (data[0] === 'EVENT' && data[2]?.kind === 0) {
                const content = JSON.parse(data[2].content);
                if (!resolved) {
                  resolved = true;
                  clearTimeout(timeout);
                  ws.close();
                  resolve(content);
                }
              }
            } catch (error) {
              console.error('Error parsing relay message:', error);
            }
          };

          ws.onerror = () => {
            ws.close();
          };

          ws.onclose = () => {
            connectedRelays--;
            if (connectedRelays === 0 && !resolved) {
              clearTimeout(timeout);
              resolve(null);
            }
          };

          // Close connection after 5 seconds if still open
          setTimeout(() => {
            if (ws.readyState === WebSocket.OPEN) {
              ws.close();
            }
          }, 5000);

        } catch (error) {
          console.error(`Error connecting to ${relayUrl}:`, error);
        }
      });
    });
  }

  /**
   * Cache profile data in sessionStorage
   * @param {string} pubkey - Hex public key
   * @param {Object} profile - Profile data
   */
  cacheProfile(pubkey, profile) {
    try {
      sessionStorage.setItem(`profile_${pubkey}`, JSON.stringify({
        data: profile,
        timestamp: Date.now()
      }));
    } catch (error) {
      console.error('Failed to cache profile:', error);
    }
  }

  /**
   * Get cached profile from sessionStorage
   * @param {string} pubkey - Hex public key
   * @returns {Object|null} Cached profile or null
   */
  getCachedProfile(pubkey) {
    try {
      const cached = sessionStorage.getItem(`profile_${pubkey}`);
      if (cached) {
        const { data, timestamp } = JSON.parse(cached);
        // Cache valid for 1 hour
        if (Date.now() - timestamp < 3600000) {
          return data;
        }
      }
    } catch (error) {
      console.error('Failed to get cached profile:', error);
    }
    return null;
  }

  /**
   * Dispatch profile update event
   */
  dispatchProfileUpdate() {
    window.dispatchEvent(new CustomEvent('profile-updated', {
      detail: this.profile
    }));
  }

  /**
   * Clear profile data
   */
  clear() {
    this.profile = null;
    this.pubkey = null;
    this.npub = null;
  }

  /**
   * Get current profile
   * @returns {Object|null} Current profile data
   */
  getProfile() {
    return this.profile;
  }
}

// Initialize global profile manager
if (typeof window !== 'undefined') {
  window.profileManager = new ProfileManager();

  // Listen for session ready event to load profile
  window.addEventListener('sessionReady', async (event) => {
    if (event.detail && event.detail.npub && event.detail.pubkey) {
      await window.profileManager.init(event.detail.pubkey, event.detail.npub);
    }
  });

  // Listen for authentication success
  window.addEventListener('authenticationSuccess', async (event) => {
    if (event.detail && event.detail.npub && event.detail.pubkey) {
      await window.profileManager.init(event.detail.pubkey, event.detail.npub);
    }
  });

  // Clear profile on logout
  window.addEventListener('loggedOut', () => {
    window.profileManager.clear();
  });
}
