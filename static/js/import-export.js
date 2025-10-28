document.getElementById('import-form').addEventListener('submit', function(e) {
    e.preventDefault();
    
    const fileInput = document.getElementById('feed-file');
    const file = fileInput.files[0];
    
    if (!file) {
        alert('Please select a JSON file to import.');
        return;
    }
    
    const reader = new FileReader();
    reader.onload = function(e) {
        try {
            const feedData = JSON.parse(e.target.result);
            
            if (!feedData.feeds || !Array.isArray(feedData.feeds)) {
                alert('Invalid JSON format. Expected a "feeds" array.');
                return;
            }
            
            fetch('/feeds/import', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(feedData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Feeds imported successfully!');
                    location.reload();
                } else {
                    alert('Error importing feeds: ' + (data.error || 'Unknown error'));
                }
            })
            .catch(error => {
                alert('Error importing feeds: ' + error.message);
            });
            
        } catch (error) {
            alert('Invalid JSON file: ' + error.message);
        }
    };
    reader.readAsText(file);
});