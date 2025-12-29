async function runValidation() {
    const results = document.getElementById('results');
    results.classList.add('show');

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
