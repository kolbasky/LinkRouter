const PROTOCOL = "linkrouter-ext://";

let modifiers = { alt: true, ctrl: false, shift: false };

if (typeof chrome !== 'undefined' && chrome.storage && chrome.storage.sync) {
  chrome.storage.sync.get(['modifiers'], (result) => {
    if (result.modifiers) modifiers = result.modifiers;
  });
  chrome.runtime.onMessage.addListener((msg) => {
    if (msg.action === 'updateModifiers') modifiers = msg.modifiers;
  });
}

const isFirefox = typeof navigator !== 'undefined' && navigator.userAgent.includes('Firefox');

// Firefox somtimes lets ctrl+click through in case of 'mousedown'
// Chrome sometimes doesn't work on first modifier+click with 'click'
const eventType = isFirefox ? 'click' : 'mousedown';

document.addEventListener(eventType, (e) => {
  const required = modifiers;
  const pressed = {
    alt: e.altKey,
    ctrl: e.ctrlKey,
    shift: e.shiftKey
  };

  // Modifiers unset
  if (!required.alt && !required.ctrl && !required.shift) return;

  // Modifiers don't match
  if (required.alt !== pressed.alt ||
      required.ctrl !== pressed.ctrl ||
      required.shift !== pressed.shift) return;

  if (e.button !== 0) return;

  const a = e.target.closest('a');
  if (!a || !/^https?:\/\//i.test(a.href)) return;

  e.preventDefault();
  e.stopPropagation();
  if (e.stopImmediatePropagation) e.stopImmediatePropagation();

  const routed = PROTOCOL + encodeURIComponent(a.href);
  window.location.href = routed;
}, true);
