const PROTOCOL = "linkrouter-ext://";

let modifiers = { alt: true, ctrl: true, shift: false };

// Load saved modifiers only if storage API is available
if (typeof chrome !== 'undefined' && chrome.storage && chrome.storage.sync) {
  chrome.storage.sync.get(['modifiers'], (result) => {
    if (result.modifiers) {
      modifiers = result.modifiers;
    }
  });

  // Also listen for updates
  chrome.runtime.onMessage.addListener((message) => {
    if (message.action === 'updateModifiers') {
      modifiers = message.modifiers;
    }
  });
}

// Fallback: use default if on restricted page (no storage access)

document.addEventListener('mousedown', (e) => {
  const required = modifiers;
  const pressed = {
    alt: e.altKey,
    ctrl: e.ctrlKey,
    shift: e.shiftKey
  };

  if (!required.alt && !required.ctrl && !required.shift) {
    return;
  }

  if (required.alt !== pressed.alt ||
      required.ctrl !== pressed.ctrl ||
      required.shift !== pressed.shift) {
    return;
  }

  if (e.button !== 0) {
    return;
  }

  const a = e.target.closest('a');
  if (!a || !/^https?:\/\//i.test(a.href)) return;

  e.preventDefault();
  e.stopPropagation();

  const routed = PROTOCOL + encodeURIComponent(a.href);
  window.location.href = routed;
}, true);
