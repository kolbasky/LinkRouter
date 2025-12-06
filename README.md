# LinkRouter

A **portable, zero-install**, Windows-only app that routes web links to specific applications based on **regex rules**.

- ✅ Open links like `music.yandex.ru`, `store.steampowered.com` directly in the apps
- ✅ Open links to different domains in different browsers or different profiles of single browser.
- ✅ Add custom rules for **any app** (Spotify, Discord, etc.). Supports regex and capture-groups for advanced cases.
- ✅ Use default browser of your choice if clicked URL doesn't match any rule
- ✅ **No installer**, **no telemetry**, **no network access**
- ✅ **Single EXE** (~3 MB), config stored next to it

Perfect for power users who want **total control** over how links open on their machine.

---

## 🚀 Quick Start

1. **Download** [`LinkRouter.exe`](https://github.com/kolbasky/link-router/releases/latest)
2. **Open PowerShell or Command Prompt** in the download folder
3. Run:
   ```powershell
   .\LinkRouter.exe --register
   ```
   this will create registry keys, necessary for setting LinkRouter as a browser. Use `--unregister` to remove them.
4. Go to Windows Settings → Apps → Default apps → Web browser → Choose LinkRouter
5. Edit config.json (auto-created on double-click) to add your rules. See config.json.example in repo.


## ⚙️ Configuration
The app creates config.json automatically when launched without parameters (or double-clicked). Key fields:

global.defaultBrowserPath: Fallback browser (e.g., Brave, Chrome)
global.defaultBrowserArgs: Browser flags (e.g., "--new-tab {URL}")
rules: Array of routing rules
  regex: Pattern to match (full URL)
  program: path to executable
  arguments: Command-line args (use {URL} to pass URL or $1, $2 for capture groups used in `regex` field)

Supports Windows environment variables like `"%LOCALAPPDATA%\\Yandex\\YandexMusic\\YandexMusic.exe"`

## 🔒 Privacy & Security
No network access — ever
No data collection
Fully open source — inspect every line

## 📦 Download
See the [Releases page](https://github.com/kolbasky/link-router/releases/latest) for the latest LinkRouter.exe.
