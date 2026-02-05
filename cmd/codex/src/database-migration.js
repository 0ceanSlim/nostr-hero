/**
 * CODEX Database Migration Entry Point
 */

// Import styles
import './styles.css';

async function startMigration() {
    const btn = document.getElementById('migrateBtn');
    const status = document.getElementById('status');
    const message = document.getElementById('statusMessage');

    btn.disabled = true;
    status.classList.add('show');
    message.innerHTML = '<span class="progress">Starting migration...</span>';

    try {
        const response = await fetch('/api/migrate/start', { method: 'POST' });
        if (!response.ok) throw new Error('Failed to start migration');

        // Poll for status
        const interval = setInterval(async () => {
            const statusResponse = await fetch('/api/migrate/status');
            const statusData = await statusResponse.json();

            if (statusData.step === 'complete') {
                clearInterval(interval);
                message.innerHTML = '<span class="success">✓ ' + statusData.message + '</span>';
                btn.disabled = false;
            } else if (statusData.step === 'error') {
                clearInterval(interval);
                message.innerHTML = '<span class="error">✗ Error: ' + statusData.error + '</span>';
                btn.disabled = false;
            } else {
                let msg = '<span class="progress">' + statusData.message + '</span>';
                if (statusData.progress > 0) {
                    msg += '<br><span style="color: #6272a4;">Progress: ' + statusData.progress + '/' + statusData.total + '</span>';
                }
                message.innerHTML = msg;
            }
        }, 500);
    } catch (error) {
        message.innerHTML = '<span class="error">✗ Error: ' + error.message + '</span>';
        btn.disabled = false;
    }
}

// Expose function to window for onclick handler
window.startMigration = startMigration;

console.log('🎯 CODEX Database Migration loaded');
