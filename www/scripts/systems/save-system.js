// Save System for Nostr Hero
// Handles saving game state from Go memory to disk

// Save game to local JSON file (writes from memory to disk)
// State is already in Go memory, this just persists it to disk
async function saveGameToLocal() {
    if (!window.sessionManager || !window.sessionManager.isAuthenticated()) {
        showMessage('❌ Must be logged in to save', 'error');
        return false;
    }

    if (!window.gameAPI || !window.gameAPI.initialized) {
        showMessage('❌ Game not initialized', 'error');
        return false;
    }

    try {
        showMessage('💾 Saving game...', 'info');

        // Save from Go memory to disk
        await window.gameAPI.saveGame();

        showMessage('✅ Game saved successfully!', 'success');
        return true;

    } catch (error) {
        console.error('❌ Save failed:', error);
        showMessage('❌ Failed to save game: ' + error.message, 'error');
        return false;
    }
}

// Quick save hotkey (Ctrl+S)
document.addEventListener('keydown', (event) => {
    if (event.ctrlKey && event.key === 's') {
        event.preventDefault();
        saveGameToLocal();
    }
});

// Alias for manual save button
window.saveGame = saveGameToLocal;
window.saveGameToLocal = saveGameToLocal;  // Explicit alias for other modules

console.log('💾 Save system loaded');
