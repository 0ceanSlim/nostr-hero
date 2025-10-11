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
        // Only redirect to saves if on home page (not saves, game, or new-game)
        const allowedPaths = ['/saves', '/game', '/new-game'];
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

    window.sessionManager.on('authenticationSuccess', (method) => {
        console.log(`✅ Authentication successful via ${method}`);
        showMessage(`✅ Successfully logged in via ${method}!`, 'success');
        setTimeout(() => {
            // Only redirect to saves if on home page (not saves, game, or new-game)
            const allowedPaths = ['/saves', '/game', '/new-game'];
            if (!allowedPaths.includes(window.location.pathname)) {
                console.log('Redirecting to saves page after login');
                window.location.href = '/saves';
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
                <h2 class="text-3xl font-bold mb-6 text-yellow-400">⚔️ Nostr Hero ⚔️</h2>
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

// Login with browser extension using SessionManager
async function loginWithExtension() {
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
}

// Login with Amber using SessionManager
async function loginWithAmber() {
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
        showMessage('✨ Generating new key pair...', 'info');
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
    const loginDetails = document.getElementById('login-details');
    loginDetails.innerHTML = `
        <div class="max-w-md mx-auto bg-gray-800 p-6 rounded-lg">
            <h3 class="text-xl font-bold mb-4 text-green-400">New Keys Generated!</h3>
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Public Key (npub)</label>
                    <div class="bg-gray-700 p-3 rounded text-xs font-mono text-green-400 break-all">
                        ${keyPair.npub}
                    </div>
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Private Key (nsec)</label>
                    <div class="bg-gray-700 p-3 rounded text-xs font-mono text-red-400 break-all">
                        ${keyPair.nsec}
                    </div>
                </div>
                <div class="text-sm text-yellow-400 bg-yellow-900 bg-opacity-20 p-3 rounded">
                    <p><strong>⚠️ IMPORTANT:</strong> Save your private key (nsec) securely! You'll need it to access your account. Anyone with this key can control your account.</p>
                </div>
                <div class="flex space-x-3">
                    <button onclick="useGeneratedKeys('${keyPair.nsec}')"
                            class="flex-1 bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded font-medium">
                        Use These Keys
                    </button>
                    <button onclick="hideKeyLogin()"
                            class="flex-1 bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded">
                        Cancel
                    </button>
                </div>
            </div>
        </div>
    `;
    loginDetails.classList.remove('hidden');
}

// Use the generated keys to login using SessionManager
async function useGeneratedKeys(nsec) {
    if (!window.sessionManager) {
        showMessage('❌ Session manager not available', 'error');
        return;
    }

    try {
        showMessage('🔐 Logging in with new keys...', 'info');
        await window.sessionManager.loginWithPrivateKey(nsec);
        // Success handling is done by event listeners
    } catch (error) {
        console.error('Generated key login error:', error);
        showMessage('❌ Login failed: ' + error.message, 'error');
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