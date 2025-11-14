// Authentication functions for Nostr Hero using grain client

// Initialize authentication with session manager integration
function initializeAuthentication() {
    if (!window.sessionManager) {
        console.error('❌ SessionManager not available for authentication');
        return;
    }

    // Set up session manager event listeners
    window.sessionManager.on('sessionReady', (sessionData) => {
        console.log('✅ Session ready');
        // Only redirect to saves if on home page (not saves, game, new-game, settings, or discover)
        const allowedPaths = ['/saves', '/game', '/new-game', '/settings', '/discover'];
        if (!allowedPaths.includes(window.location.pathname)) {
            console.log('Redirecting to saves page');
            window.location.href = '/saves';
        }
    });

    window.sessionManager.on('authenticationRequired', () => {
        console.log('🔐 Authentication required');
        // Only show login interface on game pages, not on home page
        if (window.location.pathname === '/game' || window.location.pathname === '/new-game') {
            console.log('Showing login interface');
            showLoginInterface();
        }
    });

    window.sessionManager.on('sessionExpired', () => {
        console.log('⏰ Session expired, showing login interface');
        showMessage('⏰ Your session has expired. Please log in again.', 'warning');
        showLoginInterface();
    });

    window.sessionManager.on('authenticationSuccess', (data) => {
        // Handle both old format (string) and new format (object)
        const method = typeof data === 'string' ? data : data.method;
        const isNewAccount = typeof data === 'object' ? data.isNewAccount : false;

        console.log(`✅ Authentication successful via ${method}${isNewAccount ? ' (new account)' : ''}`);

        // Update loading modal message if showing
        if (typeof showLoadingModal === 'function') {
            showLoadingModal('Redirecting to saves...');
        }

        setTimeout(() => {
            // Only redirect to saves if on home page (not saves, game, new-game, settings, or discover)
            const allowedPaths = ['/saves', '/game', '/new-game', '/settings', '/discover'];
            if (!allowedPaths.includes(window.location.pathname)) {
                console.log('Redirecting to saves page after login');
                // Loading modal will disappear on page navigation
                window.location.href = '/saves';
            } else {
                // Hide loading modal if we're staying on same page
                if (typeof hideLoadingModal === 'function') {
                    hideLoadingModal();
                }
            }
        }, 1000);
    });

    window.sessionManager.on('authenticationFailed', ({ method, error }) => {
        console.error(`❌ Authentication failed via ${method}:`, error);
        showMessage(`❌ Login failed via ${method}: ${error}`, 'error');
    });

    window.sessionManager.on('sessionError', (error) => {
        console.error('❌ Session error:', error);
        showMessage('❌ Session error: ' + error.message, 'error');
        showLoginInterface();
    });
}

// Show login interface
function showLoginInterface() {
    const gameContainer = document.getElementById('game-app');
    if (gameContainer) {
        gameContainer.innerHTML = `
            <div class="text-center py-12">
                <h2 class="text-3xl font-bold mb-6 text-yellow-400 flex items-center justify-center gap-2">
                    <img src="/res/img/static/logo.png" alt="Nostr Hero" class="inline-block" style="height: 1.5em; width: auto; image-rendering: pixelated;">
                    Nostr Hero
                    <img src="/res/img/static/logo.png" alt="Nostr Hero" class="inline-block" style="height: 1.5em; width: auto; image-rendering: pixelated;">
                </h2>
                <p class="text-gray-300 mb-8">A text-based RPG powered by Nostr</p>
                <div class="space-y-4 max-w-md mx-auto">
                    <button onclick="loginWithExtension()"
                            class="w-full bg-purple-600 hover:bg-purple-700 text-white px-6 py-3 rounded-lg font-medium transition-colors">
                        🔗 Login with Browser Extension
                    </button>
                    <button onclick="loginWithAmber()"
                            class="w-full bg-orange-600 hover:bg-orange-700 text-white px-6 py-3 rounded-lg font-medium transition-colors">
                        📱 Login with Amber
                    </button>
                    <button onclick="showKeyLogin()"
                            class="w-full bg-gray-600 hover:bg-gray-700 text-white px-6 py-3 rounded-lg font-medium transition-colors">
                        🗝️ Login with Private Key
                    </button>
                    <button onclick="generateNewKeys()"
                            class="w-full bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-lg font-medium transition-colors">
                        ✨ Generate New Keys
                    </button>
                </div>
            </div>
            <div id="login-details" class="mt-8 hidden">
                <!-- Detailed login forms will appear here -->
            </div>
        `;
    }
}

// Hide login interface and show game
function hideLoginInterface() {
    const gameContainer = document.getElementById('game-app');
    if (gameContainer) {
        // The game interface will be populated by initializeGame()
        gameContainer.innerHTML = `
            <div class="text-center py-12">
                <div class="spinner-border animate-spin inline-block w-8 h-8 border-4 rounded-full" role="status">
                    <span class="visually-hidden"></span>
                </div>
                <p class="text-gray-300 mt-4">🎮 Loading game...</p>
            </div>
        `;
    }
}

// Legacy function for backward compatibility
function redirectToLogin() {
    showLoginInterface();
}

// Login with browser extension using SessionManager (Grain-style)
async function loginWithExtension() {
    try {
        // Show loading modal if not already shown (handles both direct calls and nav-play flow)
        const loadingModal = document.getElementById('loading-modal');
        if (loadingModal && loadingModal.classList.contains('hidden')) {
            if (typeof showLoadingModal === 'function') {
                showLoadingModal('Requesting access from extension...');
            }
        } else if (typeof showLoadingModal === 'function') {
            // Update text if already shown
            showLoadingModal('Requesting access from extension...');
        }

        if (typeof showAuthResult === 'function') {
            showAuthResult('loading', 'Requesting access from extension...');
        }

        if (!window.nostr) {
            throw new Error('Nostr extension not found');
        }

        const publicKey = await window.nostr.getPublicKey();

        if (!publicKey || publicKey.length !== 64) {
            throw new Error('Invalid public key received from extension');
        }

        console.log('Extension returned public key:', publicKey);

        if (typeof showLoadingModal === 'function') {
            showLoadingModal('Creating session...');
        }

        if (typeof showAuthResult === 'function') {
            showAuthResult('loading', 'Creating session with extension signing...');
        }

        const sessionRequest = {
            public_key: publicKey,
            signing_method: 'browser_extension',
            mode: 'write'
        };

        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(sessionRequest)
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => null);
            const errorMsg = errorData?.message || `HTTP ${response.status}`;
            throw new Error(`Login failed: ${errorMsg}`);
        }

        const result = await response.json();

        if (!result.success) {
            throw new Error(result.message || 'Login failed');
        }

        console.log('Extension login successful');

        // Hide loading modal
        if (typeof hideLoadingModal === 'function') {
            hideLoadingModal();
        }

        if (typeof showAuthResult === 'function') {
            showAuthResult('success', 'Connected via browser extension!');
        }

        window.nostrExtensionConnected = true;

        setTimeout(() => {
            hideLoginModal();
            window.location.href = '/saves';
        }, 1000);
    } catch (error) {
        console.error('Extension login error:', error);

        // Hide loading modal on error
        if (typeof hideLoadingModal === 'function') {
            hideLoadingModal();
        }

        if (typeof showAuthResult === 'function') {
            showAuthResult('error', `Extension error: ${error.message}`);
        } else {
            showMessage('❌ Extension login failed: ' + error.message, 'error');
        }
    }
}

// Amber callback state
let amberCallbackReceived = false;

// Login with Amber using NIP-55 protocol (Grain-style)
function loginWithAmber() {
    // Check if running on localhost (won't work with Amber from mobile)
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        if (typeof showAuthResult === 'function') {
            showAuthResult('error', '⚠️ Amber login requires a local IP address (like 192.168.x.x), not localhost. Access the site from your computer\'s IP address on your local network.');
        } else {
            showMessage('❌ Amber login requires a local IP address, not localhost', 'error', 10000);
        }
        return;
    }

    // Show loading modal with gif
    if (typeof showLoadingModal === 'function') {
        showLoadingModal('Opening Amber app...');
    }

    if (typeof showAuthResult === 'function') {
        showAuthResult('loading', 'Opening Amber app...');
    }

    // Set up callback listener BEFORE opening Amber
    setupAmberCallbackListener();

    // Generate proper callback URL (EXACTLY like Grain - note the ?event= at the end)
    const callbackUrl = `${window.location.origin}/api/auth/amber-callback?event=`;

    // Use NIP-55 nostrsigner URL format (EXACTLY like Grain)
    const amberUrl = `nostrsigner:?compressionType=none&returnType=signature&type=get_public_key&callbackUrl=${encodeURIComponent(callbackUrl)}&appName=${encodeURIComponent('Nostr Hero')}`;

    console.log('=== AMBER DEBUG ===');
    console.log('window.location.hostname:', window.location.hostname);
    console.log('window.location.origin:', window.location.origin);
    console.log('Callback URL:', callbackUrl);
    console.log('Full Amber URL:', amberUrl);
    console.log('===================');

    try {
        // Try multiple approaches for opening the nostrsigner protocol
        let protocolOpened = false;

        // Method 1: Create anchor element and click it (most reliable on mobile)
        try {
            const anchor = document.createElement('a');
            anchor.href = amberUrl;
            anchor.target = '_blank';
            anchor.style.display = 'none';
            document.body.appendChild(anchor);

            anchor.click();
            protocolOpened = true;

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

        // Method 3: Last resort - try window.open
        if (!protocolOpened) {
            try {
                const newWindow = window.open(amberUrl, '_blank');
                if (newWindow) {
                    newWindow.close(); // Close immediately, we just want to trigger the protocol
                    protocolOpened = true;
                    console.log('Amber protocol opened via window.open');
                }
            } catch (openError) {
                console.warn('Window open method failed:', openError);
            }
        }

        if (!protocolOpened) {
            throw new Error('Unable to open Amber protocol - no method worked');
        }

        // Show additional guidance for mobile users
        if (typeof showAuthResult === 'function') {
            showAuthResult('loading', 'Opening Amber app... If nothing happens, make sure Amber is installed and try again.');
        }

        // Set timeout in case user doesn't complete the flow
        setTimeout(() => {
            if (!amberCallbackReceived) {
                // Hide loading modal on timeout
                if (typeof hideLoadingModal === 'function') {
                    hideLoadingModal();
                }

                if (typeof showAuthResult === 'function') {
                    showAuthResult('error', 'Amber connection timed out. Make sure Amber is installed and try again. If the app opened but didn\'t return, check your Amber app permissions.');
                } else {
                    showMessage('❌ Amber connection timed out', 'error');
                }
            }
        }, 60000); // 60 seconds timeout
    } catch (error) {
        console.error('Error opening Amber:', error);

        // Hide loading modal on error
        if (typeof hideLoadingModal === 'function') {
            hideLoadingModal();
        }

        if (typeof showAuthResult === 'function') {
            showAuthResult('error', 'Failed to open Amber app. Please ensure Amber is installed and your browser supports the nostrsigner protocol.');
        } else {
            showMessage('❌ Amber login failed: ' + error.message, 'error');
        }
    }
}

// Set up Amber callback listener (Grain-style)
function setupAmberCallbackListener() {
    // Listen for when user returns to the page
    const handleVisibilityChange = () => {
        if (!document.hidden && !amberCallbackReceived) {
            // Check if we're on the callback URL
            setTimeout(checkForAmberCallback, 500);
        }
    };

    const handleFocus = () => {
        if (!amberCallbackReceived) {
            setTimeout(checkForAmberCallback, 500);
        }
    };

    // Add multiple listeners to catch the return
    document.addEventListener('visibilitychange', handleVisibilityChange);
    window.addEventListener('focus', handleFocus);

    // Also check immediately
    setTimeout(checkForAmberCallback, 1000);

    // Clean up listeners after timeout
    setTimeout(() => {
        document.removeEventListener('visibilitychange', handleVisibilityChange);
        window.removeEventListener('focus', handleFocus);
    }, 65000);
}

// Check for Amber callback (Grain-style)
function checkForAmberCallback() {
    console.log('Checking for Amber callback...');

    // Check localStorage for callback result (set by the callback page)
    const amberResult = localStorage.getItem('amber_callback_result');
    if (amberResult) {
        try {
            console.log('Found Amber result in localStorage:', amberResult);
            localStorage.removeItem('amber_callback_result');
            const data = JSON.parse(amberResult);
            amberCallbackReceived = true;
            handleAmberCallbackData(data);
        } catch (error) {
            console.error('Failed to parse stored Amber result:', error);
        }
    }
}

// Handle Amber callback data (Grain-style)
function handleAmberCallbackData(data) {
    try {
        // If there's an error in the stored data
        if (data.error) {
            throw new Error(data.error);
        }

        // Session was already created by the callback handler
        // Just show success and update UI like extension login does
        console.log('Amber login completed successfully');

        // Hide loading modal
        if (typeof hideLoadingModal === 'function') {
            hideLoadingModal();
        }

        if (typeof showAuthResult === 'function') {
            showAuthResult('success', 'Connected via Amber!');
        }

        // Store Amber connection info
        window.amberConnected = true;

        // Hide modal and redirect
        setTimeout(() => {
            hideLoginModal();
            window.location.href = '/saves';
        }, 1000);
    } catch (error) {
        console.error('Error processing Amber callback data:', error);

        // Hide loading modal on error
        if (typeof hideLoadingModal === 'function') {
            hideLoadingModal();
        }

        if (typeof showAuthResult === 'function') {
            showAuthResult('error', `Amber login failed: ${error.message}`);
        } else {
            showMessage('❌ Amber login failed: ' + error.message, 'error');
        }
    }
}

// Create Amber login URL
function createAmberLoginURL() {
    const callbackUrl = encodeURIComponent(`${window.location.origin}/api/auth/amber-callback`);
    return `intent://get_public_key?callback_url=${callbackUrl}#Intent;scheme=nostrsigner;package=com.greenart7c3.nostrsigner;end`;
}

// Show private key login form
function showKeyLogin() {
    const loginDetails = document.getElementById('login-details');
    loginDetails.innerHTML = `
        <div class="max-w-md mx-auto bg-gray-800 p-6 rounded-lg">
            <h3 class="text-xl font-bold mb-4 text-yellow-400">Login with Private Key</h3>
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Private Key (nsec or hex)</label>
                    <input type="password" id="auth-private-key-input"
                           class="w-full bg-gray-700 text-white px-3 py-2 rounded border border-gray-600 focus:border-yellow-400"
                           placeholder="nsec1... or hex">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Encryption Password</label>
                    <input type="password" id="auth-encryption-password-input"
                           class="w-full bg-gray-700 text-white px-3 py-2 rounded border border-gray-600 focus:border-yellow-400"
                           placeholder="Password to encrypt your key">
                </div>
                <div class="flex space-x-3">
                    <button onclick="loginWithPrivateKey()"
                            class="flex-1 bg-yellow-600 hover:bg-yellow-700 text-gray-900 px-4 py-2 rounded font-medium">
                        Login
                    </button>
                    <button onclick="hideKeyLogin()"
                            class="flex-1 bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded">
                        Cancel
                    </button>
                </div>
            </div>
            <div class="mt-4 text-sm text-gray-400">
                <p>⚠️ Your private key will be encrypted with your password and stored securely for this session only.</p>
            </div>
        </div>
    `;
    loginDetails.classList.remove('hidden');
}

// Hide private key login form
function hideKeyLogin() {
    const loginDetails = document.getElementById('login-details');
    loginDetails.classList.add('hidden');
    loginDetails.innerHTML = '';
}

// Login with private key using SessionManager
async function loginWithPrivateKey() {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    const privateKeyInput = document.getElementById('auth-private-key-input');
    const passwordInput = document.getElementById('auth-encryption-password-input');
    const privateKey = privateKeyInput?.value?.trim();
    const password = passwordInput?.value?.trim();

    if (!privateKey) {
        showMessage('❌ Please enter your private key', 'error');
        return;
    }

    if (!password) {
        showMessage('❌ Please enter an encryption password', 'error');
        return;
    }

    try {
        showMessage('🗝️ Logging in with private key...', 'info');
        await window.sessionManager.loginWithPrivateKey(privateKey, { password });

        // Clear the inputs for security
        privateKeyInput.value = '';
        passwordInput.value = '';
        // Success handling is done by event listeners
    } catch (error) {
        console.error('Private key login error:', error);
        showMessage('❌ Private key login failed: ' + error.message, 'error');
    }
}

// Generate new keys using SessionManager
async function generateNewKeys() {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    try {
        const keyPair = await window.sessionManager.generateKeys();

        if (keyPair) {
            showGeneratedKeys(keyPair);
        } else {
            showMessage('❌ Failed to generate keys', 'error');
        }

    } catch (error) {
        console.error('Key generation error:', error);
        showMessage('❌ Key generation failed: ' + error.message, 'error');
    }
}

// Show generated keys to user
function showGeneratedKeys(keyPair) {
    // Store globally for use by useGeneratedKeys
    window.generatedKeyPair = keyPair;

    // Populate the modal fields
    const npubEl = document.getElementById('gen-npub');
    const nsecEl = document.getElementById('gen-nsec');

    if (npubEl) npubEl.textContent = keyPair.npub;
    if (nsecEl) nsecEl.textContent = keyPair.nsec;

    // Show the modal
    const modal = document.getElementById('generated-keys-modal');
    if (modal) {
        modal.classList.remove('hidden');
    }
}

// Use the generated keys to login
async function useGeneratedKeys() {
    // Check both possible variable locations
    const privateKey = window.generatedPrivateKey || (window.generatedKeyPair && window.generatedKeyPair.nsec);

    if (!privateKey) {
        showMessage('❌ No keys available', 'error');
        return;
    }

    // Store in global variable for encryption modal to access
    window.generatedPrivateKey = privateKey;

    // Hide generated keys modal and show encryption password modal
    hideGeneratedKeys();

    // Show encryption password modal (function defined in nav-play.html)
    if (typeof showEncryptionPasswordModal === 'function') {
        showEncryptionPasswordModal();
    } else {
        console.error('showEncryptionPasswordModal function not found');
        showMessage('❌ Error: Encryption modal not available', 'error');
    }
}

// Logout function using SessionManager
async function logout() {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    try {
        showMessage('🚪 Logging out...', 'info');
        await window.sessionManager.logout();
        showMessage('✅ Successfully logged out', 'success');
        setTimeout(() => {
            showLoginInterface();
        }, 1000);
    } catch (error) {
        console.error('Logout error:', error);
        showMessage('❌ Logout failed: ' + error.message, 'error');
    }
}

// Initialize authentication system when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('🎮 Authentication system loading...');

    // Wait for session manager to be available
    const checkSessionManager = () => {
        if (window.sessionManager) {
            console.log('✅ SessionManager found, initializing authentication');
            initializeAuthentication();
        } else {
            console.log('⏳ Waiting for SessionManager...');
            setTimeout(checkSessionManager, 100);
        }
    };

    checkSessionManager();
});

console.log('🔐 Authentication system loaded');