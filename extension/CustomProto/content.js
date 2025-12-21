const PROTOCOL = "linkrouter-ext://";

document.addEventListener('click', (e) => {
  if (!e.altKey) return;

  const a = e.target.closest('a');
  if (!a || !/^https?:\/\//i.test(a.href)) return;

  e.preventDefault();
  e.stopPropagation();

  const routed = PROTOCOL + encodeURIComponent(a.href);
  window.location.href = routed;
}, true);
