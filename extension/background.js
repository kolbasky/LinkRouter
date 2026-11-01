const PROTOCOL = "linkrouter-ext://";

const pendingUrls = new Map();

chrome.webNavigation.onBeforeNavigate.addListener((details) => {
  if (details.frameId === 0) {
    pendingUrls.set(details.tabId, details.url);
  }
});

chrome.webNavigation.onCompleted.addListener((details) => {
  if (details.frameId === 0) {
    pendingUrls.delete(details.tabId);
  }
});

chrome.webNavigation.onErrorOccurred.addListener((details) => {
  if (details.frameId === 0) {
    pendingUrls.delete(details.tabId);
  }
});

chrome.tabs.onRemoved.addListener((tabId) => {
  pendingUrls.delete(tabId);
});


chrome.runtime.onInstalled.addListener(() => {
  // right-click menu for links
  chrome.contextMenus.create({
    id: "open-in-linkrouter",
    title: "Open with LinkRouter",
    contexts: ["link"]
  });

  // right-click menu for ext
  chrome.contextMenus.create({
    id: "separator-1",
    type: "separator",
    contexts: ["action"]
  });

  chrome.contextMenus.create({
    id: "visit-releases",
    title: "Download LinkRouter App",
    contexts: ["action"]
  });

  chrome.contextMenus.create({
    id: "about-linkrouter",
    title: "About LinkRouter",
    contexts: ["action"]
  });
});

chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === "open-in-linkrouter" && info.linkUrl && tab?.id) {
    const originalUrl = tab.url || '';  // fallback
    const routed = PROTOCOL + encodeURIComponent(info.linkUrl);

    chrome.tabs.update(tab.id, { url: routed }, (updatedTab) => {
      if (chrome.runtime.lastError || !updatedTab) {
        chrome.tabs.update(tab.id, { url: originalUrl });
      }
    });
  }
  else if (info.menuItemId === "visit-releases") {
    chrome.tabs.create({ url: "https://github.com/kolbasky/LinkRouter/releases/latest" });
  }
  else if (info.menuItemId === "about-linkrouter") {
    chrome.tabs.create({
      url: "https://github.com/kolbasky/LinkRouter/blob/main/README.md#-linkrouter"
    });
  }
});

// Click on the extension icon (toolbar button)
chrome.action.onClicked.addListener((tab) => {
  if (!tab?.id) return;

  let targetUrl = pendingUrls.get(tab.id);
  if (!targetUrl) {
    targetUrl = tab.url;
  }

  if (!targetUrl || /^(chrome|brave|edge|about|moz-extension|opera|vivaldi|yandex|browser|view-source|resource|moz-src):/i.test(targetUrl)) {
    return;
  }

  const originalUrl = tab.url || targetUrl;
  const routed = PROTOCOL + encodeURIComponent(targetUrl);
  
  chrome.tabs.update(tab.id, { url: routed }, (updatedTab) => {
    if (chrome.runtime.lastError || !updatedTab) {
      chrome.tabs.update(tab.id, { url: originalUrl });
    }
  });
});
