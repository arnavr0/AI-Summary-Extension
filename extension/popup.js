// extension/popup.js

document.addEventListener('DOMContentLoaded', () => {
  const summaryContainer = document.getElementById('summary-container');
  const statusDiv = document.getElementById('status');

  // Try to get the text from storage
  chrome.storage.local.get(['selectedText'], async (result) => {
    const text = result.selectedText;

    if (text) {
      statusDiv.textContent = 'Generating summary...';

      try {
        // Call our Go backend
        const response = await fetch('http://localhost:8080/summarize', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ text: text }),
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        
        // Display the summary and clear the status
        statusDiv.style.display = 'none';
        summaryContainer.textContent = data.summary;

      } catch (error) {
        console.error('Error fetching summary:', error);
        statusDiv.textContent = 'Error: Could not fetch summary. Is the backend server running?';
        statusDiv.style.color = 'red';
      } finally {
        // Clear the stored text after we've used it
        chrome.storage.local.remove('selectedText');
      }

    } else {
      // Keep the initial message if no text was found
      statusDiv.textContent = 'Select text on a page, right-click, and choose "Summarize Selected Text".';
    }
  });
});