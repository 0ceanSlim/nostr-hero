/**
 * CODEX Validation Entry Point
 */

// Import styles
import './styles.css';

async function runValidation() {
    const results = document.getElementById('results');
    results.classList.add('show');

    // Clear cleanup results when running validation
    document.getElementById('cleanup-results').innerHTML = '';

    try {
        const response = await fetch('/api/validation/run', { method: 'POST' });
        const data = await response.json();

        document.getElementById('totalFiles').textContent = data.stats.total_files;
        document.getElementById('errors').textContent = data.stats.error_count;
        document.getElementById('warnings').textContent = data.stats.warning_count;
        document.getElementById('info').textContent = data.stats.info_count;

        const issuesDiv = document.getElementById('issues');
        issuesDiv.innerHTML = '';

        if (data.issues.length === 0) {
            issuesDiv.innerHTML = '<div style="text-align: center; color: #50fa7b; padding: 20px;">✓ No issues found!</div>';
        } else {
            data.issues.forEach(issue => {
                const div = document.createElement('div');
                div.className = 'issue ' + issue.type;
                div.innerHTML =
                    '<div class="issue-type">[' + issue.category.toUpperCase() + '] ' + issue.message + '</div>' +
                    '<div class="issue-file">' + issue.file + (issue.field ? ' → ' + issue.field : '') + '</div>';
                issuesDiv.appendChild(div);
            });
        }
    } catch (error) {
        alert('Error running validation: ' + error.message);
    }
}

async function runCleanup(dryRun) {
    const results = document.getElementById('results');
    results.classList.add('show');

    // Clear previous results
    document.getElementById('issues').innerHTML = '';
    const cleanupResultsDiv = document.getElementById('cleanup-results');
    cleanupResultsDiv.innerHTML = '<div style="text-align: center; padding: 20px;">Processing cleanup...</div>';

    try {
        const url = '/api/validation/cleanup' + (dryRun ? '?dry_run=true' : '');
        const response = await fetch(url, { method: 'POST' });
        const data = await response.json();

        // Update stats
        document.getElementById('totalFiles').textContent = data.files_processed;
        document.getElementById('errors').textContent = '0';
        document.getElementById('warnings').textContent = '0';
        document.getElementById('info').textContent = data.changes.length;

        cleanupResultsDiv.innerHTML = '';

        if (dryRun) {
            const header = document.createElement('div');
            header.style = 'background: #f1fa8c; color: #282a36; padding: 10px; margin-bottom: 10px; border-radius: 4px; font-weight: bold;';
            header.textContent = '👁️ PREVIEW MODE - No files were modified';
            cleanupResultsDiv.appendChild(header);
        } else {
            const header = document.createElement('div');
            header.style = 'background: #50fa7b; color: #282a36; padding: 10px; margin-bottom: 10px; border-radius: 4px; font-weight: bold;';
            header.textContent = `✅ Cleanup Complete - ${data.files_modified} files modified`;
            cleanupResultsDiv.appendChild(header);
        }

        if (data.changes.length === 0) {
            cleanupResultsDiv.innerHTML += '<div style="text-align: center; color: #50fa7b; padding: 20px;">✓ No changes needed!</div>';
        } else {
            // Group changes by file
            const changesByFile = {};
            data.changes.forEach(change => {
                if (!changesByFile[change.file]) {
                    changesByFile[change.file] = [];
                }
                changesByFile[change.file].push(change);
            });

            // Display changes grouped by file
            Object.keys(changesByFile).forEach(file => {
                const fileDiv = document.createElement('div');
                fileDiv.style = 'background: #44475a; padding: 15px; margin-bottom: 10px; border-radius: 4px;';

                const fileHeader = document.createElement('div');
                fileHeader.style = 'color: #8be9fd; font-weight: bold; margin-bottom: 8px;';
                fileHeader.textContent = file;
                fileDiv.appendChild(fileHeader);

                changesByFile[file].forEach(change => {
                    const changeDiv = document.createElement('div');
                    changeDiv.style = 'margin-left: 15px; margin-bottom: 5px; font-size: 14px;';

                    let color = '#f8f8f2';
                    let icon = '•';
                    if (change.type === 'added') {
                        color = '#50fa7b';
                        icon = '+';
                    } else if (change.type === 'removed') {
                        color = '#ff5555';
                        icon = '-';
                    } else if (change.type === 'fixed') {
                        color = '#f1fa8c';
                        icon = '!';
                    } else if (change.type === 'reordered') {
                        color = '#bd93f9';
                        icon = '↔';
                    }

                    changeDiv.innerHTML = `<span style="color: ${color}; font-weight: bold;">${icon}</span> ${change.message}`;
                    fileDiv.appendChild(changeDiv);
                });

                cleanupResultsDiv.appendChild(fileDiv);
            });
        }
    } catch (error) {
        cleanupResultsDiv.innerHTML = '<div style="color: #ff5555; padding: 20px;">Error running cleanup: ' + error.message + '</div>';
    }
}

// Expose functions to window for onclick handlers
window.runValidation = runValidation;
window.runCleanup = runCleanup;

console.log('🎯 CODEX Validation loaded');
