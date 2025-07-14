// extension/background.js

// Create the context menu item when the extension is installed
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: "summarizeText",
    title: "Summarize Selected Text",
    contexts: ["selection"] // Only show when text is selected
  });
});

// Listener for when the context menu item is clicked
chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === "summarizeText" && info.selectionText) {
    // Store the selected text in local storage to be picked up by the popup
    chrome.storage.local.set({ selectedText: info.selectionText }, () => {
      console.log("Text saved.");
      // You can optionally open the popup here, but the user can also click the icon.
      // chrome.action.openPopup(); // This is a new API, might not work in all versions
    });
  }
});