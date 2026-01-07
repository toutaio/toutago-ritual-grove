// {{.wiki_name}} Wiki JavaScript
document.addEventListener('DOMContentLoaded', function() {
    // Auto-save draft functionality
    const contentArea = document.getElementById('content');
    if (contentArea) {
        const saveKey = 'wiki_draft_' + window.location.pathname;
        
        // Load draft on page load
        const draft = localStorage.getItem(saveKey);
        if (draft && !contentArea.value) {
            if (confirm('Restore unsaved draft?')) {
                contentArea.value = draft;
            }
        }
        
        // Save draft on input
        let saveTimeout;
        contentArea.addEventListener('input', function() {
            clearTimeout(saveTimeout);
            saveTimeout = setTimeout(function() {
                localStorage.setItem(saveKey, contentArea.value);
            }, 1000);
        });
        
        // Clear draft on successful save
        const form = contentArea.closest('form');
        if (form) {
            form.addEventListener('submit', function() {
                localStorage.removeItem(saveKey);
            });
        }
    }
    
    // Markdown preview (basic)
    const previewBtn = document.getElementById('preview-btn');
    if (previewBtn) {
        previewBtn.addEventListener('click', function(e) {
            e.preventDefault();
            const content = document.getElementById('content').value;
            const preview = document.getElementById('preview');
            if (preview) {
                // Basic markdown rendering (would need proper markdown parser in production)
                preview.innerHTML = content
                    .replace(/^### (.*$)/gim, '<h3>$1</h3>')
                    .replace(/^## (.*$)/gim, '<h2>$1</h2>')
                    .replace(/^# (.*$)/gim, '<h1>$1</h1>')
                    .replace(/\*\*(.*)\*\*/gim, '<strong>$1</strong>')
                    .replace(/\*(.*)\*/gim, '<em>$1</em>')
                    .replace(/\n/gim, '<br>');
            }
        });
    }
});
