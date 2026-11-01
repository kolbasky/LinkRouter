const MODIFIERS_KEY = 'modifiers';
const DEFAULT_MODIFIERS = { alt: true, ctrl: true, shift: false };

document.addEventListener('DOMContentLoaded', () => {
  chrome.storage.sync.get([MODIFIERS_KEY], (result) => {
    let saved = result[MODIFIERS_KEY];

    if (!saved) {
      saved = DEFAULT_MODIFIERS;
      chrome.storage.sync.set({ [MODIFIERS_KEY]: saved });
    }

    document.getElementById('alt').checked = !!saved.alt;
    document.getElementById('ctrl').checked = !!saved.ctrl;
    document.getElementById('shift').checked = !!saved.shift;

    notifyAllTabs(saved);
  });

  document.querySelectorAll('input[type="checkbox"]').forEach(cb => {
    cb.addEventListener('change', saveAndNotify);
  });

  document.getElementById('open-current').addEventListener('click', () => {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      const currentTab = tabs[0];
      const targetUrl = (currentTab?.pendingUrl)
        ? currentTab.pendingUrl
        : currentTab?.url;

      if (!targetUrl) {
        return;
      }

      if (!targetUrl || /^(chrome|brave|edge|about|moz-extension|opera|vivaldi|yandex|browser|view-source|resource|moz-src):/i.test(targetUrl)) {
        return;
      }

      const originalUrl = targetUrl;
      const routed = "linkrouter-ext://" + encodeURIComponent(targetUrl);

      chrome.tabs.update(currentTab.id, { url: routed }, () => {
        if (chrome.runtime.lastError) {
          chrome.tabs.update(currentTab.id, { url: originalUrl });
        }
      });

      window.close();
    });
  });
});

function saveAndNotify() {
  const settings = {
    alt: document.getElementById('alt').checked,
    ctrl: document.getElementById('ctrl').checked,
    shift: document.getElementById('shift').checked
  };

  chrome.storage.sync.set({ [MODIFIERS_KEY]: settings }, () => {
    notifyAllTabs(settings);
  });
}

function notifyAllTabs(settings) {
  chrome.tabs.query({}, (tabs) => {
    tabs.forEach(tab => {
      chrome.tabs.sendMessage(tab.id, { action: 'updateModifiers', modifiers: settings })
        .catch(() => {}); // ignore closed/unloaded tabs
    });
  });
}
