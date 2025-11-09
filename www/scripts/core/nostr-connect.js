/**
 * NIP-46 Nostr Connect Implementation
 * Allows desktop login by scanning QR code with Amber on mobile
 */

/**
 * Show the unified Amber modal with both Direct and QR options
 */
function showAmberOptions() {
  console.log('📱 Showing Amber options modal...');

  // Hide login modal, show Amber modal
  hideLoginModal();

  const qrModal = document.getElementById('nostr-connect-qr-modal');
  if (qrModal) {
    qrModal.classList.remove('hidden');
  }

  // Reset QR state
  const qrContainer = document.getElementById('qr-code-container');
  const connectionStringContainer = document.getElementById('connection-string-container');
  const showQRBtn = document.getElementById('show-qr-btn');

  if (qrContainer) qrContainer.classList.add('hidden');
  if (connectionStringContainer) connectionStringContainer.classList.add('hidden');
  if (showQRBtn) showQRBtn.classList.remove('hidden');

  updateNostrConnectStatus('⏳ Click "Show QR Code" to start', 'info');
}

/**
 * Generate QR code when button is clicked (from within the modal)
 */
async function generateAmberQRCode() {
  console.log('📷 Generating Amber QR code...');

  // Hide the button, show QR container
  const showQRBtn = document.getElementById('show-qr-btn');
  const qrContainer = document.getElementById('qr-code-container');
  const connectionStringContainer = document.getElementById('connection-string-container');

  if (showQRBtn) showQRBtn.classList.add('hidden');
  if (qrContainer) qrContainer.classList.remove('hidden');
  if (connectionStringContainer) connectionStringContainer.classList.remove('hidden');

  // Trigger the full NIP-46 flow
  await initiateNostrConnect();
}

/**
 * Initiate full Nostr Connect flow (called by generateAmberQRCode)
 */
async function initiateNostrConnect() {
  try {
    console.log('🔗 Initiating Nostr Connect (NIP-46)...');
    updateNostrConnectStatus('⏳ Generating connection...', 'loading');

    // Generate client keypair for this session
    if (!window.NostrTools) {
      throw new Error('nostr-tools library not loaded');
    }

    const { generateSecretKey, getPublicKey } = window.NostrTools;

    const clientSecretKey = generateSecretKey();
    const clientPubkey = getPublicKey(clientSecretKey);

    clientKeypair = {
      secretKey: clientSecretKey,
      publicKey: clientPubkey
    };

    console.log('🔑 Generated client keypair:', clientPubkey.substring(0, 16) + '...');

    // Generate random secret for connection verification
    const secret = Array.from(crypto.getRandomValues(new Uint8Array(16)))
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');

    // Create nostrconnect:// URI
    const metadata = encodeURIComponent(JSON.stringify({
      name: 'Nostr Hero',
      url: window.location.origin,
      description: 'D&D-inspired RPG powered by Nostr'
    }));

    const relayParams = DEFAULT_RELAYS.map(r => `relay=${encodeURIComponent(r)}`).join('&');
    const nostrConnectURI = `nostrconnect://${clientPubkey}?${relayParams}&secret=${secret}&metadata=${metadata}`;

    console.log('📋 Nostr Connect URI:', nostrConnectURI);

    // Display URI in input field
    const uriInput = document.getElementById('nostr-connect-uri');
    if (uriInput) {
      uriInput.value = nostrConnectURI;
    }

    // Generate and display QR code
    generateQRCode(nostrConnectURI);

    // Update status
    updateNostrConnectStatus('📱 Scan QR code with Amber', 'ready');

    // Start listening for connection on relays
    await listenForNostrConnect(clientKeypair, secret);

    // Set timeout (2 minutes)
    connectionTimeout = setTimeout(() => {
      updateNostrConnectStatus('⏰ Connection timeout - please try again', 'error');
      cleanupNostrConnect();
    }, 120000);

  } catch (error) {
    console.error('❌ Nostr Connect error:', error);
    updateNostrConnectStatus('❌ Error: ' + error.message, 'error');
  }
}

/**
 * Hide the Nostr Connect/Amber modal and cleanup
 */
function hideNostrConnectQR() {
  const qrModal = document.getElementById('nostr-connect-qr-modal');
  if (qrModal) {
    qrModal.classList.add('hidden');
  }
  cleanupNostrConnect();
}



// Default relays for NIP-46 communication
const DEFAULT_RELAYS = [
  'wss://relay.damus.io',
  'wss://relay.primal.net',
  'wss://relay.nostr.band',
  'wss://nos.lol'
];

function generateQRCode(uri) {
  try {
    // QRCodeJS uses a container div, not a canvas directly
    let container = document.getElementById('nostr-connect-qr-container');

    if (!container) {
      // Create container if it doesn't exist
      const canvas = document.getElementById('nostr-connect-qr-canvas');
      if (canvas) {
        // Replace canvas with div
        container = document.createElement('div');
        container.id = 'nostr-connect-qr-container';
        canvas.parentNode.replaceChild(container, canvas);
      } else {
        throw new Error('QR container element not found');
      }
    }

    // Clear any existing QR code
    container.innerHTML = '';

    // Check for QRCode library
    if (!window.QRCode) {
      console.error('Available on window:', Object.keys(window).filter(k => k.toLowerCase().includes('qr')));
      throw new Error('QRCode library not loaded');
    }

    console.log('Using QRCode:', typeof window.QRCode);

    // Generate QR code using qrcodejs API
    new window.QRCode(container, {
      text: uri,
      width: 280,
      height: 280,
      colorDark: '#000000',
      colorLight: '#ffffff',
      correctLevel: window.QRCode.CorrectLevel.M
    });

    console.log('✅ QR code generated successfully');
  } catch (error) {
    console.error('❌ QR code generation error:', error);
    throw error;
  }
}

/**
 * Copy Nostr Connect URI to clipboard
 */
async function copyNostrConnectURI() {
  const uriInput = document.getElementById('nostr-connect-uri');
  if (!uriInput || !uriInput.value) return;

  try {
    await navigator.clipboard.writeText(uriInput.value);
    if (typeof showMessage === 'function') {
      showMessage('✅ Connection string copied!', 'success');
    }
  } catch (error) {
    // Fallback for older browsers
    uriInput.select();
    document.execCommand('copy');
    if (typeof showMessage === 'function') {
      showMessage('✅ Connection string copied!', 'success');
    }
  }
}

/**
 * Listen for NIP-46 connect response on relays
 */
async function listenForNostrConnect(keypair, expectedSecret) {
  try {
    console.log('👂 Listening for Nostr Connect response on relays...');

    if (!window.NostrTools) {
      throw new Error('nostr-tools library not loaded');
    }

    const { SimplePool, nip04, nip44 } = window.NostrTools;

    // Create relay pool
    nostrConnectPool = new SimplePool();

    // Subscribe to kind 24133 events directed at our client pubkey
    const filters = [{
      kinds: [24133],
      '#p': [keypair.publicKey],
      since: Math.floor(Date.now() / 1000)
    }];

    console.log('🔍 Subscribing with filters:', filters);

    const sub = nostrConnectPool.subscribeMany(
      DEFAULT_RELAYS,
      filters,
      {
        onevent(event) {
          handleNostrConnectEvent(event, keypair, expectedSecret);
        },
        oneose() {
          console.log('📡 Relay subscription established (EOSE received)');
        }
      }
    );

    console.log('✅ Subscribed to relays for NIP-46 messages');

  } catch (error) {
    console.error('❌ Listen for Nostr Connect error:', error);
    updateNostrConnectStatus('❌ Connection error: ' + error.message, 'error');
  }
}

/**
 * Handle incoming NIP-46 event (kind 24133)
 */
async function handleNostrConnectEvent(event, clientKeypair, expectedSecret) {
  try {
    console.log('📨 Received NIP-46 event:', event);

    if (!window.NostrTools) {
      throw new Error('nostr-tools library not loaded');
    }

    const { nip44, getPublicKey } = window.NostrTools;

    // Decrypt the event content using NIP-44
    const decrypted = await nip44.decrypt(
      clientKeypair.secretKey,
      event.pubkey,
      event.content
    );

    console.log('🔓 Decrypted content:', decrypted);

    const message = JSON.parse(decrypted);

    // Check if it's a connect response
    if (message.method === 'connect') {
      // Validate secret
      if (message.result !== expectedSecret) {
        console.warn('⚠️ Invalid secret in connect response');
        return;
      }

      console.log('✅ Valid connect response received!');
      const remotePubkey = event.pubkey;

      updateNostrConnectStatus('✅ Connected! Logging in...', 'success');

      // Now request the public key from the remote signer
      await requestPublicKey(clientKeypair, remotePubkey);
    }
    else if (message.method === 'get_public_key' && message.result) {
      // Received public key response
      const publicKey = message.result;
      console.log('🔑 Received public key:', publicKey);

      // Login with this public key using Amber signing method
      await loginWithNostrConnect(publicKey);
    }

  } catch (error) {
    console.error('❌ Error handling NIP-46 event:', error);
  }
}

/**
 * Request public key from remote signer
 */
async function requestPublicKey(clientKeypair, remotePubkey) {
  try {
    console.log('📤 Requesting public key from remote signer...');

    if (!window.NostrTools) {
      throw new Error('nostr-tools library not loaded');
    }

    const { nip44, finalizeEvent } = window.NostrTools;

    // Create request message
    const request = {
      id: crypto.randomUUID(),
      method: 'get_public_key',
      params: []
    };

    // Encrypt request
    const encrypted = await nip44.encrypt(
      clientKeypair.secretKey,
      remotePubkey,
      JSON.stringify(request)
    );

    // Create kind 24133 event
    const event = {
      kind: 24133,
      created_at: Math.floor(Date.now() / 1000),
      tags: [['p', remotePubkey]],
      content: encrypted
    };

    // Sign and publish
    const signedEvent = finalizeEvent(event, clientKeypair.secretKey);

    await nostrConnectPool.publish(DEFAULT_RELAYS, signedEvent);
    console.log('✅ Published get_public_key request');

  } catch (error) {
    console.error('❌ Request public key error:', error);
    updateNostrConnectStatus('❌ Request error: ' + error.message, 'error');
  }
}

/**
 * Complete login after receiving public key via Nostr Connect
 */
async function loginWithNostrConnect(publicKey) {
  try {
    console.log('🎮 Logging in with Nostr Connect public key:', publicKey);

    const sessionRequest = {
      public_key: publicKey,
      signing_method: 'amber_signer',  // Use Amber signing method
      mode: 'write'
    };

    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(sessionRequest)
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => null);
      throw new Error(`Login failed: ${errorData?.message || response.status}`);
    }

    const result = await response.json();

    updateNostrConnectStatus('✅ Login successful!', 'success');

    // Clean up
    cleanupNostrConnect();

    // Redirect to saves page
    setTimeout(() => {
      hideNostrConnectQR();
      window.location.href = '/saves';
    }, 1000);

  } catch (error) {
    console.error('❌ Nostr Connect login error:', error);
    updateNostrConnectStatus('❌ Login failed: ' + error.message, 'error');
  }
}

/**
 * Update Nostr Connect status display
 */
function updateNostrConnectStatus(message, type) {
  const statusText = document.getElementById('nostr-connect-status-text');
  const statusDiv = document.getElementById('nostr-connect-status');

  if (statusText) {
    statusText.textContent = message;
  }

  if (statusDiv) {
    // Update styling based on type
    statusDiv.className = 'mb-3 p-2 win95-inset text-center';

    if (type === 'success') {
      statusDiv.style.borderColor = 'var(--color-success, #00ff41)';
    } else if (type === 'error') {
      statusDiv.style.borderColor = 'var(--color-error, #ff4444)';
    } else {
      statusDiv.style.borderColor = '';
    }
  }

  console.log(`[NostrConnect] ${message}`);
}

/**
 * Cleanup Nostr Connect resources
 */
function cleanupNostrConnect() {
  console.log('🧹 Cleaning up Nostr Connect resources...');

  if (connectionTimeout) {
    clearTimeout(connectionTimeout);
    connectionTimeout = null;
  }

  if (nostrConnectPool) {
    nostrConnectPool.close(DEFAULT_RELAYS);
    nostrConnectPool = null;
  }

  nostrConnectSigner = null;
  clientKeypair = null;
}

console.log('✅ Nostr Connect module loaded');

// Export functions to window for HTML onclick handlers
window.showAmberOptions = showAmberOptions;
window.generateAmberQRCode = generateAmberQRCode;

console.log('✅ Amber options and QR functions exported');

// Backwards compatibility alias
window.showNostrConnectQR = showAmberOptions;

