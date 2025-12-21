const HOST_NAME = "com.kolbasky.linkrouter";

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "checkNativeStatus") {
    chrome.runtime.sendNativeMessage(
      HOST_NAME,
      { action: "ping" },
      (response) => {
        if (chrome.runtime.lastError) {
          console.error("Failed to contact LinkRouter:", chrome.runtime.lastError.message);
          sendResponse({ status: "error" });
        } else {
          sendResponse(response);
        }
      }
    );
    return true;
  }

  if (request.action === "shouldHandle" && request.url) {
    chrome.runtime.sendNativeMessage(
      HOST_NAME,
      { url: request.url, action: "shouldHandle" },
      (response) => {
        sendResponse(response || { handled: false });
      }
    );
    return true;
  }

  if (request.url && !request.action) {
    chrome.runtime.sendNativeMessage(HOST_NAME, { url: request.url });
    return false;
  }

  return false;
});
