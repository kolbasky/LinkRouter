document.addEventListener('click', (event) => {
  const anchor = event.target.closest('a');
  if (!anchor || !/^https?:\/\//i.test(anchor.href)) return;

  event.preventDefault();
  event.stopPropagation();

  chrome.runtime.sendMessage(
    { action: "shouldHandle", url: anchor.href },
    (response) => {
      if (response?.handled) {
        // console.log("Link was handled by LinkRouter")
        return
        // chrome.runtime.sendMessage({ url: anchor.href }); // if handled is true the app is already opened
      } else {
        // console.log("Link should be handled by browser")
        if (anchor.target) {
          window.open(anchor.href, anchor.target);
          return;
        }
        const openInNewTab =
          event.button === 1 ||
          event.ctrlKey ||
          event.metaKey ||
          event.shiftKey ||
          event.altKey;

        if (openInNewTab) {
          window.open(anchor.href, '_blank');
        } else {
          window.location.href = anchor.href;
        }
      }
    }
  );
}, true);
