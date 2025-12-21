const HOST_NAME = "com.kolbasky.linkrouter";

function checkConnection() {
  chrome.runtime.sendMessage(
    { action: "checkNativeStatus" },
    (response) => {
      const statusEl = document.getElementById("status");
      const pathEl = document.getElementById("path");

      if (chrome.runtime.lastError) {
        statusEl.textContent = "Error contacting background";
        statusEl.className = "red";
        pathEl.textContent = "N/A";
      } else if (response && response.status === "ok") {
        statusEl.textContent = "Connected";
        statusEl.className = "green";
        pathEl.textContent = response.exePath || "Unknown";
      } else {
        statusEl.textContent = "Disconnected (Not Registered)";
        statusEl.className = "red";
        pathEl.textContent = "N/A";
      }
    }
  );
}

document.addEventListener("DOMContentLoaded", checkConnection);
