// Save System for Nostr Hero
// Handles saving game state to local JSON files

// Save game to local JSON file
async function saveGameToLocal() {
    if (!window.sessionManager || !window.sessionManager.isAuthenticated()) {
        showMessage('❌ Must be logged in to save', 'error');
        return false;
    }

    try {
        showMessage('💾 Saving game...', 'info');

        const session = window.sessionManager.getSession();
        const gameState = getGameState();

        // Prepare save data
        const saveData = {
            character: gameState.character,
            gameState: gameState,
            location: gameState.location.current,
        };

        // Get current save ID from URL if loading existing save
        const urlParams = new URLSearchParams(window.location.search);
        const currentSaveID = urlParams.get('save');

        if (currentSaveID) {
            saveData.id = currentSaveID;
        }

        // Send save request
        const response = await fetch(`/api/saves/${session.npub}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(saveData)
        });

        if (!response.ok) {
            throw new Error(`Save failed: ${response.status}`);
        }

        const result = await response.json();

        if (result.success) {
            showMessage('✅ Game saved successfully!', 'success');

            // Update URL with save ID if this was a new save
            if (!currentSaveID && result.save_id) {
                const newUrl = new URL(window.location);
                newUrl.searchParams.set('save', result.save_id);
                newUrl.searchParams.delete('new');
                window.history.replaceState({}, '', newUrl);
            }

            return true;
        } else {
            throw new Error(result.message || 'Save failed');
        }

    } catch (error) {
        console.error('❌ Save failed:', error);
        showMessage('❌ Failed to save game: ' + error.message, 'error');
        return false;
    }
}

// Auto-save functionality
let autoSaveInterval = null;
const AUTO_SAVE_INTERVAL = 5 * 60 * 1000; // 5 minutes

function startAutoSave() {
    if (autoSaveInterval) {
        clearInterval(autoSaveInterval);
    }

    autoSaveInterval = setInterval(async () => {
        console.log('🔄 Auto-saving game...');
        const success = await saveGameToLocal();
        if (success) {
            console.log('✅ Auto-save complete');
        }
    }, AUTO_SAVE_INTERVAL);

    console.log('⏰ Auto-save enabled (every 5 minutes)');
}

function stopAutoSave() {
    if (autoSaveInterval) {
        clearInterval(autoSaveInterval);
        autoSaveInterval = null;
        console.log('⏹️ Auto-save disabled');
    }
}

// Save before leaving page
window.addEventListener('beforeunload', async (event) => {
    if (window.sessionManager && window.sessionManager.isAuthenticated()) {
        // Attempt quick save (though it may not complete due to page unload)
        saveGameToLocal();
    }
});

// Save on page visibility change (when tab becomes hidden)
document.addEventListener('visibilitychange', async () => {
    if (document.hidden && window.sessionManager && window.sessionManager.isAuthenticated()) {
        console.log('👁️ Page hidden, saving game...');
        await saveGameToLocal();
    }
});

// Export save functionality
async function exportSave() {
    if (!window.sessionManager || !window.sessionManager.isAuthenticated()) {
        showMessage('❌ Must be logged in to export', 'error');
        return;
    }

    try {
        const session = window.sessionManager.getSession();
        const gameState = getGameState();

        const exportData = {
            version: "1.0",
            npub: session.npub,
            exported_at: new Date().toISOString(),
            character: gameState.character,
            gameState: gameState
        };

        const dataStr = JSON.stringify(exportData, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });

        const link = document.createElement('a');
        link.href = URL.createObjectURL(dataBlob);
        link.download = `nostr-hero-save-${new Date().toISOString().split('T')[0]}.json`;
        link.click();

        showMessage('✅ Save file exported!', 'success');

    } catch (error) {
        console.error('❌ Export failed:', error);
        showMessage('❌ Failed to export save: ' + error.message, 'error');
    }
}

// Quick save hotkey (Ctrl+S)
document.addEventListener('keydown', (event) => {
    if (event.ctrlKey && event.key === 's') {
        event.preventDefault();
        saveGameToLocal();
    }
});

console.log('💾 Save system loaded');