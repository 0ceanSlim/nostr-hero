// Session Manager for Nostr Hero
// Handles session initialization, persistence, and recovery similar to gnostream

class SessionManager {
    constructor() {
        this.sessionData = null;
        this.isInitialized = false;
        this.initPromise = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 3;
        this.sessionCheckInterval = null;
        this.eventListeners = new Map();

        // Session status enum
        this.SessionStatus = {
            INITIALIZING: 'initializing',
            ACTIVE: 'active',
            EXPIRED: 'expired',
            ERROR: 'error',
            UNAUTHENTICATED: 'unauthenticated'
        };

        this.currentStatus = this.SessionStatus.INITIALIZING;

        // Initialize session on construction
        this.init();
    }

    // Initialize session manager
    async init() {
        if (this.initPromise) {
            return this.initPromise;
        }

        this.initPromise = this._performInit();
        return this.initPromise;
    }

    async _performInit() {
        console.log('🎮 SessionManager: Initializing Nostr Hero session...');

        try {
            // Check for existing session
            await this.checkExistingSession();

            if (this.currentStatus === this.SessionStatus.ACTIVE) {
                console.log('✅ SessionManager: Found active session');
                this.startSessionMonitoring();
                this.isInitialized = true;
                this.emit('sessionReady', this.sessionData);
                return true;
            } else {
                console.log('🔐 SessionManager: No active session, authentication required');
                this.currentStatus = this.SessionStatus.UNAUTHENTICATED;
                this.emit('authenticationRequired');
                return false;
            }
        } catch (error) {
            console.error('❌ SessionManager: Initialization failed:', error);
            this.currentStatus = this.SessionStatus.ERROR;
            this.emit('sessionError', error);
            return false;
        }
    }

    // Check for existing session
    async checkExistingSession() {
        try {
            const response = await fetch('/api/auth/session', {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`Session check failed: ${response.status}`);
            }

            const result = await response.json();

            if (result.success && result.is_active && result.session) {
                this.sessionData = {
                    publicKey: result.session.public_key,
                    npub: result.npub,
                    signingMethod: result.session.signing_method,
                    mode: result.session.mode,
                    isActive: true,
                    lastCheck: Date.now()
                };
                this.currentStatus = this.SessionStatus.ACTIVE;
                return true;
            } else {
                this.sessionData = null;
                this.currentStatus = this.SessionStatus.UNAUTHENTICATED;
                return false;
            }
        } catch (error) {
            console.error('Session check error:', error);
            this.currentStatus = this.SessionStatus.ERROR;
            throw error;
        }
    }

    // Start session monitoring to detect expiration
    startSessionMonitoring() {
        if (this.sessionCheckInterval) {
            clearInterval(this.sessionCheckInterval);
        }

        // Check session every 30 seconds
        this.sessionCheckInterval = setInterval(async () => {
            try {
                const isValid = await this.validateSession();
                if (!isValid) {
                    console.warn('⚠️ SessionManager: Session expired or invalid');
                    this.handleSessionExpiry();
                }
            } catch (error) {
                console.error('SessionManager: Session validation error:', error);
            }
        }, 30000);
    }

    // Validate current session
    async validateSession() {
        if (!this.sessionData) return false;

        try {
            const response = await fetch('/api/auth/session');
            if (!response.ok) return false;

            const result = await response.json();
            return result.success && result.is_active;
        } catch (error) {
            console.error('Session validation error:', error);
            return false;
        }
    }

    // Handle session expiry
    handleSessionExpiry() {
        this.currentStatus = this.SessionStatus.EXPIRED;
        this.sessionData = null;

        if (this.sessionCheckInterval) {
            clearInterval(this.sessionCheckInterval);
            this.sessionCheckInterval = null;
        }

        this.emit('sessionExpired');

        // Attempt to re-authenticate if possible
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`🔄 SessionManager: Attempting reconnection (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            setTimeout(() => this.attemptReconnection(), 2000);
        } else {
            console.log('❌ SessionManager: Max reconnection attempts reached');
            this.emit('authenticationRequired');
        }
    }

    // Attempt to reconnect session
    async attemptReconnection() {
        try {
            // Try to restore session from localStorage or other persistent storage
            const restored = await this.restoreSessionFromStorage();
            if (restored) {
                console.log('✅ SessionManager: Session restored from storage');
                this.reconnectAttempts = 0;
                this.startSessionMonitoring();
                this.emit('sessionRestored', this.sessionData);
                return true;
            }
        } catch (error) {
            console.error('SessionManager: Reconnection failed:', error);
        }

        return false;
    }

    // Restore session from localStorage (fallback)
    async restoreSessionFromStorage() {
        try {
            // For security, we don't store private keys in localStorage
            // We only store session metadata and re-validate with server
            const storedSession = localStorage.getItem('nostr_hero_session_meta');
            if (!storedSession) return false;

            const sessionMeta = JSON.parse(storedSession);

            // Validate the stored session is recent (within 1 hour)
            if (Date.now() - sessionMeta.timestamp > 3600000) {
                localStorage.removeItem('nostr_hero_session_meta');
                return false;
            }

            // Try to restore session with server
            const isValid = await this.checkExistingSession();
            if (isValid && this.sessionData.publicKey === sessionMeta.publicKey) {
                this.currentStatus = this.SessionStatus.ACTIVE;
                return true;
            }

            return false;
        } catch (error) {
            console.error('Session restore error:', error);
            return false;
        }
    }

    // Login with different methods
    async loginWithExtension() {
        try {
            if (!window.nostr) {
                throw new Error('No Nostr extension found. Please install Alby or nos2x.');
            }

            this.emit('authenticationStarted', 'extension');

            // Add timeout for extension prompt (30 seconds)
            const publicKeyPromise = window.nostr.getPublicKey();
            const timeoutPromise = new Promise((_, reject) => {
                setTimeout(() => reject(new Error('Extension request timed out. Please try again.')), 30000);
            });

            const publicKey = await Promise.race([publicKeyPromise, timeoutPromise]);
            if (!publicKey) {
                throw new Error('Failed to get public key from extension');
            }

            const loginRequest = {
                public_key: publicKey,
                signing_method: 'browser_extension',
                mode: 'write'
            };

            return await this.performLogin(loginRequest);
        } catch (error) {
            this.emit('authenticationFailed', { method: 'extension', error: error.message });
            throw error;
        }
    }

    async loginWithPrivateKey(privateKey) {
        try {
            if (!privateKey) {
                throw new Error('Private key is required');
            }

            this.emit('authenticationStarted', 'private_key');

            const loginRequest = {
                private_key: privateKey,
                signing_method: 'encrypted_key',
                mode: 'write'
            };

            return await this.performLogin(loginRequest);
        } catch (error) {
            this.emit('authenticationFailed', { method: 'private_key', error: error.message });
            throw error;
        }
    }

    async loginWithAmber() {
        try {
            this.emit('authenticationStarted', 'amber');

            // Set up callback listener BEFORE opening Amber
            this.setupAmberCallbackListener();

            // Use proper NIP-55 nostrsigner URL format
            const amberUrl = this.createAmberLoginURL();

            console.log('Opening Amber with URL:', amberUrl);

            // Try multiple approaches for opening the nostrsigner protocol (mobile-first)
            let protocolOpened = false;

            // Method 1: Create anchor element and click it (most reliable on mobile)
            try {
                const anchor = document.createElement('a');
                anchor.href = amberUrl;
                anchor.target = '_blank';
                anchor.style.display = 'none';
                document.body.appendChild(anchor);

                // Trigger click to open Amber
                anchor.click();
                protocolOpened = true;

                // Clean up anchor element
                setTimeout(() => {
                    if (document.body.contains(anchor)) {
                        document.body.removeChild(anchor);
                    }
                }, 100);

                console.log('Amber protocol opened via anchor click');
            } catch (anchorError) {
                console.warn('Anchor method failed:', anchorError);
            }

            // Method 2: Fallback to window.location.href if anchor didn't work
            if (!protocolOpened) {
                try {
                    window.location.href = amberUrl;
                    protocolOpened = true;
                    console.log('Amber protocol opened via window.location.href');
                } catch (locationError) {
                    console.warn('Window location method failed:', locationError);
                }
            }

            if (!protocolOpened) {
                throw new Error('Unable to open Amber protocol');
            }

            return new Promise((resolve, reject) => {
                // Store resolve/reject for callback handler
                window._amberLoginPromise = { resolve, reject };

                // Set timeout in case user doesn't complete the flow
                setTimeout(() => {
                    if (window._amberLoginPromise) {
                        window._amberLoginPromise = null;
                        reject(new Error('Amber connection timed out. Make sure Amber is installed and try again.'));
                    }
                }, 60000); // 60 seconds timeout
            });
        } catch (error) {
            this.emit('authenticationFailed', { method: 'amber', error: error.message });
            throw error;
        }
    }

    // Set up Amber callback listener
    setupAmberCallbackListener() {
        const handleVisibilityChange = () => {
            if (!document.hidden) {
                setTimeout(() => this.checkForAmberCallback(), 500);
            }
        };

        const handleFocus = () => {
            setTimeout(() => this.checkForAmberCallback(), 500);
        };

        // Add listeners
        document.addEventListener('visibilitychange', handleVisibilityChange);
        window.addEventListener('focus', handleFocus);

        // Check immediately
        setTimeout(() => this.checkForAmberCallback(), 1000);

        // Clean up listeners after timeout
        setTimeout(() => {
            document.removeEventListener('visibilitychange', handleVisibilityChange);
            window.removeEventListener('focus', handleFocus);
        }, 65000);
    }

    // Check for Amber callback
    async checkForAmberCallback() {
        const currentUrl = new URL(window.location.href);

        // Check if this is the amber-callback page or has event parameter
        if (currentUrl.pathname === '/api/auth/amber-callback' || currentUrl.searchParams.has('event')) {
            try {
                // Session should be created by backend, check it
                const isActive = await this.checkExistingSession();
                if (isActive && window._amberLoginPromise) {
                    this.currentStatus = this.SessionStatus.ACTIVE;
                    this.startSessionMonitoring();
                    this.emit('authenticationSuccess', 'amber');
                    window._amberLoginPromise.resolve(this.sessionData);
                    window._amberLoginPromise = null;
                }
            } catch (error) {
                if (window._amberLoginPromise) {
                    window._amberLoginPromise.reject(error);
                    window._amberLoginPromise = null;
                }
            }
        }

        // Check localStorage for callback result
        const amberResult = localStorage.getItem('amber_callback_result');
        if (amberResult && window._amberLoginPromise) {
            try {
                localStorage.removeItem('amber_callback_result');
                const data = JSON.parse(amberResult);

                if (data.error) {
                    throw new Error(data.error);
                }

                // Check session
                const isActive = await this.checkExistingSession();
                if (isActive) {
                    this.currentStatus = this.SessionStatus.ACTIVE;
                    this.startSessionMonitoring();
                    this.emit('authenticationSuccess', 'amber');
                    window._amberLoginPromise.resolve(this.sessionData);
                    window._amberLoginPromise = null;
                } else {
                    throw new Error('Amber login succeeded but session not found');
                }
            } catch (error) {
                if (window._amberLoginPromise) {
                    window._amberLoginPromise.reject(error);
                    window._amberLoginPromise = null;
                }
            }
        }
    }

    // Perform login request to server
    async performLogin(loginRequest) {
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(loginRequest)
        });

        if (!response.ok) {
            throw new Error(`Login failed: ${response.status}`);
        }

        const result = await response.json();

        if (!result.success) {
            throw new Error(result.error || result.message || 'Login failed');
        }

        // Store session data
        this.sessionData = {
            publicKey: result.session.public_key,
            npub: result.npub,
            signingMethod: result.session.signing_method,
            mode: result.session.mode,
            isActive: true,
            lastCheck: Date.now()
        };

        this.currentStatus = this.SessionStatus.ACTIVE;
        this.reconnectAttempts = 0;

        // Store session metadata for recovery (no private keys)
        this.storeSessionMetadata();

        // Start monitoring
        this.startSessionMonitoring();

        this.emit('authenticationSuccess', loginRequest.signing_method);
        return this.sessionData;
    }

    // Store session metadata for recovery
    storeSessionMetadata() {
        try {
            const sessionMeta = {
                publicKey: this.sessionData.publicKey,
                npub: this.sessionData.npub,
                signingMethod: this.sessionData.signingMethod,
                timestamp: Date.now()
            };
            localStorage.setItem('nostr_hero_session_meta', JSON.stringify(sessionMeta));
        } catch (error) {
            console.warn('Failed to store session metadata:', error);
        }
    }

    // Logout
    async logout() {
        try {
            await fetch('/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });
        } catch (error) {
            console.warn('Logout request failed:', error);
        }

        // Clean up local state
        this.sessionData = null;
        this.currentStatus = this.SessionStatus.UNAUTHENTICATED;
        this.reconnectAttempts = 0;

        if (this.sessionCheckInterval) {
            clearInterval(this.sessionCheckInterval);
            this.sessionCheckInterval = null;
        }

        // Clear stored session metadata
        localStorage.removeItem('nostr_hero_session_meta');

        this.emit('loggedOut');
    }

    // Generate new key pair
    async generateKeys() {
        try {
            const response = await fetch('/api/auth/generate-keys', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`Key generation failed: ${response.status}`);
            }

            const result = await response.json();

            if (!result.success) {
                throw new Error(result.error || 'Key generation failed');
            }

            return result.key_pair;
        } catch (error) {
            console.error('Key generation error:', error);
            throw error;
        }
    }

    // Create Amber login URL
    createAmberLoginURL() {
        const callbackUrl = encodeURIComponent(`${window.location.origin}/api/auth/amber-callback`);
        return `intent://get_public_key?callback_url=${callbackUrl}#Intent;scheme=nostrsigner;package=com.greenart7c3.nostrsigner;end`;
    }

    // Event system
    on(eventName, callback) {
        if (!this.eventListeners.has(eventName)) {
            this.eventListeners.set(eventName, []);
        }
        this.eventListeners.get(eventName).push(callback);
    }

    off(eventName, callback) {
        if (this.eventListeners.has(eventName)) {
            const callbacks = this.eventListeners.get(eventName);
            const index = callbacks.indexOf(callback);
            if (index > -1) {
                callbacks.splice(index, 1);
            }
        }
    }

    emit(eventName, data) {
        console.log(`🎮 SessionManager: ${eventName}`, data);
        if (this.eventListeners.has(eventName)) {
            this.eventListeners.get(eventName).forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error(`Error in ${eventName} event handler:`, error);
                }
            });
        }
    }

    // Getters
    getSession() {
        return this.sessionData;
    }

    getStatus() {
        return this.currentStatus;
    }

    isAuthenticated() {
        return this.currentStatus === this.SessionStatus.ACTIVE && this.sessionData;
    }

    getPublicKey() {
        return this.sessionData?.publicKey;
    }

    getNpub() {
        return this.sessionData?.npub;
    }

    getSigningMethod() {
        return this.sessionData?.signingMethod;
    }
}

// Create global session manager instance
window.sessionManager = new SessionManager();

console.log('🎮 SessionManager loaded and initialized');

// Global login functions for use in dropdown and other components
window.loginWithExtension = async function() {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    try {
        showMessage('🔗 Connecting to browser extension...', 'info');
        await window.sessionManager.loginWithExtension();
        // Success handling is done by event listeners
    } catch (error) {
        console.error('Extension login error:', error);
        showMessage('❌ Extension login failed: ' + error.message, 'error');
    }
};

window.loginWithAmber = async function() {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    try {
        showMessage('📱 Connecting to Amber...', 'info');
        await window.sessionManager.loginWithAmber();
        // Success handling is done by event listeners
    } catch (error) {
        console.error('Amber login error:', error);
        showMessage('❌ Amber login failed: ' + error.message, 'error');
    }
};

window.loginWithPrivateKey = async function(privateKeyParam) {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    const privateKey = privateKeyParam || document.getElementById('private-key-input')?.value?.trim();

    if (!privateKey) {
        showMessage('❌ Please enter your private key', 'error');
        return;
    }

    try {
        showMessage('🗝️ Logging in with private key...', 'info');
        await window.sessionManager.loginWithPrivateKey(privateKey);

        // Clear the input for security if it exists
        const input = document.getElementById('private-key-input');
        if (input && !privateKeyParam) {
            input.value = '';
        }

        // Success handling is done by event listeners
    } catch (error) {
        console.error('Private key login error:', error);
        showMessage('❌ Private key login failed: ' + error.message, 'error');
    }
};

window.generateNewKeys = async function() {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    try {
        showMessage('✨ Generating new key pair...', 'info');
        const keyPair = await window.sessionManager.generateKeys();

        if (keyPair) {
            // Check if we're on a page with the dropdown display
            if (typeof showGeneratedKeys === 'function') {
                showGeneratedKeys(keyPair);
            } else {
                // Show keys in a simple alert for now
                alert(`New Keys Generated!\n\nPublic Key: ${keyPair.npub}\n\nPrivate Key: ${keyPair.nsec}\n\nSAVE YOUR PRIVATE KEY SECURELY!`);

                // Auto-login with new keys
                await window.loginWithPrivateKey(keyPair.nsec);
            }
            showMessage('✅ Keys generated successfully!', 'success');
        } else {
            showMessage('❌ Failed to generate keys', 'error');
        }
    } catch (error) {
        console.error('Key generation error:', error);
        showMessage('❌ Key generation failed: ' + error.message, 'error');
    }
};