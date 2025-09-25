// Authentication functions for Nostr Hero using grain client

// Redirect to login interface if no user session
function redirectToLogin() {
    const gameContainer = document.getElementById('game-app');
    if (gameContainer) {
        gameContainer.innerHTML = `
            <div class="text-center py-12">
                <h2 class="text-2xl font-bold mb-6 text-yellow-400">Welcome to Nostr Hero!</h2>
                <p class="text-gray-300 mb-8">Please log in with your Nostr identity to start playing.</p>
                <div class="space-y-4 max-w-md mx-auto">
                    <button onclick="loginWithExtension()"
                            class="w-full bg-purple-600 hover:bg-purple-700 text-white px-6 py-3 rounded-lg font-medium">
                        🔗 Login with Browser Extension
                    </button>
                    <button onclick="loginWithAmber()"
                            class="w-full bg-orange-600 hover:bg-orange-700 text-white px-6 py-3 rounded-lg font-medium">
                        📱 Login with Amber
                    </button>
                    <button onclick="showKeyLogin()"
                            class="w-full bg-gray-600 hover:bg-gray-700 text-white px-6 py-3 rounded-lg font-medium">
                        🗝️ Login with Private Key
                    </button>
                    <button onclick="generateNewKeys()"
                            class="w-full bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-lg font-medium">
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

// Login with browser extension (like Alby, nos2x)
async function loginWithExtension() {
    try {
        showMessage('🔗 Connecting to browser extension...', 'info');

        // Check if nostr extension is available
        if (!window.nostr) {
            showMessage('❌ No Nostr extension found. Please install Alby or nos2x.', 'error');
            return;
        }

        // Get public key from extension
        const publicKey = await window.nostr.getPublicKey();
        if (!publicKey) {
            showMessage('❌ Failed to get public key from extension', 'error');
            return;
        }

        // Create login request
        const loginRequest = {
            public_key: publicKey,
            signing_method: 'browser_extension',
            mode: 'write'
        };

        // Send login request to server
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(loginRequest)
        });

        const result = await response.json();

        if (result.success) {
            showMessage('✅ Successfully logged in with browser extension!', 'success');
            // Reload the game interface
            setTimeout(() => {
                window.location.reload();
            }, 1000);
        } else {
            showMessage('❌ Login failed: ' + (result.error || result.message), 'error');
        }

    } catch (error) {
        console.error('Extension login error:', error);
        showMessage('❌ Extension login failed: ' + error.message, 'error');
    }
}

// Login with Amber (Android app)
async function loginWithAmber() {
    try {
        showMessage('📱 Connecting to Amber...', 'info');

        const amberUrl = createAmberLoginURL();

        // Open Amber in a new window
        const amberWindow = window.open(amberUrl, 'amber_login', 'width=400,height=600');

        // Listen for the callback
        window.addEventListener('message', function(event) {
            if (event.data && event.data.type === 'amber_success') {
                showMessage('✅ Successfully logged in with Amber!', 'success');
                amberWindow.close();
                // Reload the game interface
                setTimeout(() => {
                    window.location.reload();
                }, 1000);
            } else if (event.data && event.data.type === 'amber_error') {
                showMessage('❌ Amber login failed: ' + event.data.error, 'error');
                amberWindow.close();
            }
        });

        // Check if window was closed without completing login
        const checkClosed = setInterval(() => {
            if (amberWindow.closed) {
                clearInterval(checkClosed);
                // Check localStorage for result
                const result = localStorage.getItem('amber_callback_result');
                if (result) {
                    const amberResult = JSON.parse(result);
                    if (amberResult.success) {
                        showMessage('✅ Successfully logged in with Amber!', 'success');
                        setTimeout(() => {
                            window.location.reload();
                        }, 1000);
                    }
                    localStorage.removeItem('amber_callback_result');
                }
            }
        }, 1000);

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
                    <input type="password" id="private-key-input"
                           class="w-full bg-gray-700 text-white px-3 py-2 rounded border border-gray-600 focus:border-yellow-400"
                           placeholder="nsec1... or hex">
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
                <p>⚠️ Your private key will be used for this session only and stored securely.</p>
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

// Login with private key
async function loginWithPrivateKey() {
    const privateKeyInput = document.getElementById('private-key-input');
    const privateKey = privateKeyInput.value.trim();

    if (!privateKey) {
        showMessage('❌ Please enter your private key', 'error');
        return;
    }

    try {
        showMessage('🗝️ Logging in with private key...', 'info');

        const loginRequest = {
            private_key: privateKey,
            signing_method: 'encrypted_key',
            mode: 'write'
        };

        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(loginRequest)
        });

        const result = await response.json();

        if (result.success) {
            showMessage('✅ Successfully logged in!', 'success');
            // Clear the input for security
            privateKeyInput.value = '';
            // Reload the game interface
            setTimeout(() => {
                window.location.reload();
            }, 1000);
        } else {
            showMessage('❌ Login failed: ' + (result.error || result.message), 'error');
        }

    } catch (error) {
        console.error('Private key login error:', error);
        showMessage('❌ Private key login failed: ' + error.message, 'error');
    }
}

// Generate new keys
async function generateNewKeys() {
    try {
        showMessage('✨ Generating new key pair...', 'info');

        const response = await fetch('/api/auth/generate-keys', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        if (result.success && result.key_pair) {
            showGeneratedKeys(result.key_pair);
        } else {
            showMessage('❌ Failed to generate keys: ' + (result.error || 'Unknown error'), 'error');
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

// Use the generated keys to login
async function useGeneratedKeys(nsec) {
    try {
        showMessage('🔐 Logging in with new keys...', 'info');

        const loginRequest = {
            private_key: nsec,
            signing_method: 'encrypted_key',
            mode: 'write'
        };

        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(loginRequest)
        });

        const result = await response.json();

        if (result.success) {
            showMessage('✅ Successfully logged in with new keys!', 'success');
            // Reload the game interface
            setTimeout(() => {
                window.location.reload();
            }, 1000);
        } else {
            showMessage('❌ Login failed: ' + (result.error || result.message), 'error');
        }

    } catch (error) {
        console.error('Generated key login error:', error);
        showMessage('❌ Login failed: ' + error.message, 'error');
    }
}

// Logout function
async function logout() {
    try {
        const response = await fetch('/api/auth/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        if (result.success) {
            showMessage('✅ Successfully logged out', 'success');
            // Redirect to login
            setTimeout(() => {
                redirectToLogin();
            }, 1000);
        } else {
            showMessage('❌ Logout failed: ' + (result.error || result.message), 'error');
        }

    } catch (error) {
        console.error('Logout error:', error);
        showMessage('❌ Logout failed: ' + error.message, 'error');
    }
}

console.log('Authentication system loaded');