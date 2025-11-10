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
   * @param {boolean} skipRelayFetch - Skip relay fetching (for new accounts)
   */
  async init(pubkey, npub, skipRelayFetch = false) {
    this.pubkey = pubkey;
    this.npub = npub;

    // Check cache first
    const cached = this.getCachedProfile(pubkey);
    if (cached) {
      this.profile = cached;
      this.dispatchProfileUpdate();
      return cached;
    }

    // Skip relay fetch for new accounts (they have no events yet)
    if (skipRelayFetch) {
      console.log('⚡ Skipping relay fetch for new account');
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
   * Fetch profile from backend API (with caching)
   * @param {string} pubkey - Hex public key
   * @returns {Promise<Object>} Profile metadata
   */
  async fetchProfile(pubkey) {
    try {
      // Use backend API which has caching
      const response = await fetch(`/api/profile?npub=${this.npub}`);

      if (!response.ok) {
        console.warn('Failed to fetch profile from backend:', response.status);
        return null;
      }

      const data = await response.json();

      if (data.profile) {
        console.log(`${data.cached ? '✅ Profile loaded from cache' : '⏳ Profile fetched from relays'}`);
        return {
          display_name: data.profile.display_name || data.profile.DisplayName,
          name: data.profile.name || data.profile.Name,
          picture: data.profile.picture || data.profile.Picture,
          about: data.profile.about || data.profile.About,
          nip05: data.profile.nip05 || data.profile.Nip05,
          lud16: data.profile.lud16 || data.profile.Lud16
        };
      }

      return null;
    } catch (error) {
      console.error('Error fetching profile from backend:', error);
      return null;
    }
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
      const skipRelayFetch = event.detail.isNewAccount || false;
      await window.profileManager.init(event.detail.pubkey, event.detail.npub, skipRelayFetch);
      // Update login button after profile loads
      if (typeof updateLoginButton === 'function') {
        updateLoginButton();
      }
    }
  });

  // Listen for authentication success
  window.addEventListener('authenticationSuccess', async (event) => {
    if (event.detail && event.detail.npub && event.detail.pubkey) {
      const skipRelayFetch = event.detail.isNewAccount || false;
      await window.profileManager.init(event.detail.pubkey, event.detail.npub, skipRelayFetch);
      // Update login button after profile loads
      if (typeof updateLoginButton === 'function') {
        updateLoginButton();
      }
    }
  });

  // Clear profile on logout
  window.addEventListener('loggedOut', () => {
    window.profileManager.clear();
    // Update login button after logout
    if (typeof updateLoginButton === 'function') {
      updateLoginButton();
    }
  });
}
