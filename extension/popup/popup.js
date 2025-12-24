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
      if (!currentTab?.url || !/^https?:\/\//i.test(currentTab.url)) {
        window.close();
        return;
      }

      const originalUrl = currentTab.url;
      const routed = "linkrouter-ext://" + encodeURIComponent(originalUrl);

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
