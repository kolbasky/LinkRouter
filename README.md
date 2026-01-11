# ![LinkRouter icon](resources/icon/32.png?raw=true) LinkRouter

A lightweight portable Windows app that routes links to specific applications based on **regex rules**.

Windows lets you choose a program to handle specific protocols, but there is no way to choose an app based on link contents. This app aims to fill that gap and suits power users who want **total control** over how links open on their machine.
- ✅ Open https-links to Steam, Discord, Spotify, Yandex Music etc in their native apps
- ✅ Route work domains to one browser, personal sites to another, and sensitive links to private/incognito mode
- ✅ Handle mailto: or ssh:// links from different companies with different apps or command-line arguments
- ✅ Add custom regex rules for **any app** and **any protocol**
- ✅ Use regex capture groups to reformat URLs before launching (e.g., extract host, user, path)
- ✅ All unhandled links fall back to your default browser - no broken workflows

Key features:
- ✅ Zero telemetry, zero network access
- ✅ Portable (single .exe) or installer with GUI config editor
- ✅ Companion browser extension for [Chrome, Edge, Opera](https://chromewebstore.google.com/detail/linkrouter/bglhinejoiniefbgkpkppjgckedilcki) and [Firefox](https://addons.mozilla.org/en-US/firefox/addon/linkrouter/). Also available on [Releases page](https://github.com/kolbasky/LinkRouter/releases/latest) for loading in developer mode.
- ✅ Fire-and-forget: launches target app and exits — no background processes, no additional memory or cpu usage

Use it in two ways:
- System: Set as default handler for HTTP/HTTPS or any other protocol to catch links from Word, PDF readers, messengers, etc.
- Browser: Use the extension to send links from inside your browser

> [!WARNING]
> There are some false-positives on VirusTotal for this program: [4 out of 72 AVs mark this file suspicious](https://www.virustotal.com/gui/file/adf68c106da238d73d5d644a32a3da8f04325c7a9c97012e246e26e884409b3d). We can do nothing about it at the moment without sacrificing functionality. If the project lives we will try code signing and contacting AV vendors.

## 🚀 Quick Start

### Installer (recommended)
1. **Download** [`LinkRouter-Installer.exe`](https://github.com/kolbasky/LinkRouter/releases/latest) and follow the instructions
2. Use GUI editor to create/edit rules. See [configuration section](#%EF%B8%8F-configuration) to get you started.
3. **optionally**, install the browser extension from [mozilla addons](https://addons.mozilla.org/en-US/firefox/addon/linkrouter/) or [chorme store](https://chromewebstore.google.com/detail/linkrouter/bglhinejoiniefbgkpkppjgckedilcki). If there are any problems with extension avaiability in web-stores, there is an according optional task during isntallation.

### Portable/custom
1. Download [`linkrouter.exe`](https://github.com/kolbasky/LinkRouter/releases/latest)
2. Place `linkrouter.exe` where you want and **run it** by double-clicking.
3. Select "Yes" in dialog to register the app in the system.
4. `Windows Settings` → `Apps` → `Default apps` dialog should pop-up automatically. If not press `Win+I` and start typing "default".
5. **optionally select LinkRouter as the default handler for protocols** you want it to handle. `linkrouter-ext` protocol is created by default and is mandatory for browser-extension to work.
6. **optionally**, install the browser extension from [mozilla addons](https://addons.mozilla.org/en-US/firefox/addon/linkrouter/) or [chorme store](https://chromewebstore.google.com/detail/linkrouter/bglhinejoiniefbgkpkppjgckedilcki). Also extension are available on [Releases page](https://github.com/kolbasky/LinkRouter/releases/latest) and may be loaded in develoer mode.
7. **Edit the config** by double-clicking `linkrouter.exe` again. Or copy [`linkrouter-gui.exe`](https://github.com/kolbasky/LinkRouter/releases/latest) to the same folder as `linkrouter.exe` and launch it.
> [!NOTE]
> Additional right-click menu entries are available on `linkrouter.exe` for you convenience after registration (may be hidden inside "show more options"):
> - Register LinkRouter (or re-register)
> - Unregister LinkRouter
> - Edit LinkRouter config
> - Help with LinkRouter

## 💻 Command line usage
```
linkrouter.exe
  no parameters - asks to register if not registered. If registered - runs --edit
  --register - register app in system (also available via right-click menu)
  --unregister - unregister app in system (also available via right-click menu)
  --edit - open linkrouter.json in global.defaultConfigEditor (also available via right-click menu)
  --help - open the online README.md from this repo in global.fallbackBrowserPath (also available via right-click menu)
  --version - show dialog window with version number
  any parameter not starting with -- is treated as a link and is matched against Rule-list or opened in global.fallbackBrowserPath
```

## ⚙️ Configuration
### GUI editor
For user convenience we developed a GUI config editor which can create/edit/search rules, validate regex syntax and indicate regex-matching on the fly AND it includes interactive mode - when no matching rule is found it opens a rule-creation dialog with the `regex` and `test URL` pre-filled. Search field searches in `program`, `arguments` and `regex` fields, and also checks search string against regexes, so you can easily find a matching rule by entering URL there. All actions are saved automatically.<br>
⌨️ **Hotkeys**:<br>
- `Ctrl+N` → create new rule<br>
- `Ctrl+C`/`Ctrl+V` → copy/paste rule<br>
- `Ctrl+D` → duplicate rule<br>
- `Ctrl+Z`/`Ctrl+Y` → undo/redo<br>
- `Ctrl+F`,`Ctrl+L`,`/` → focus search field<br>
- `Ctrl+S` in main window → save config as<br>
- `Ctrl+S`/`Ctrl+ENTER` in edit/settings dialog → save config, close dialog<br>
- `Ctrl+O` → open Test URL in browser. useful in interactive mode<br>
- `ARROWS` → navigate rules<br>
- `ENTER` → edit selected rule<br>
- `DELETE` → delete selected rule<br>
### Basic info
The app auto-creates `linkrouter.json` next to executable on its first launch and tries to detect your current default browser to use as the fallback one. If it fails, it tries to guess one from a list of known popular browsers locations.
When loading config, LinkRouter checks:
- `%LOCALAPPDATA%\LinkRouter\linkrouter.json`
- `linkrouter.json` in the same folder as the executable
When creating a new config, it tries to create it next to the executable first. If that fails, it falls back to `%LOCALAPPDATA%\LinkRouter\linkrouter.json`.

Every link passed to LinkRouter is tested against the rules in order. The first matching rule wins.

- `regex` – Golang-flavored regular expression
- `program` – full path to the target executable. environment variables are supported. If only a filename is provided, it is resolved via PATH.
- `arguments` – command-line arguments; `{URL}` is replaced with the original link, `$1`, `$2`… are replaced with capture-group contents.

Links that do not match any rule are passed to `global.fallbackBrowserPath` with `global.fallbackBrowserArgs` as arguments.

You can handle any protocol (mailto, ssh, steam, spotify, etc.). Just add the protocol to `global.supportedProtocols` and re-run `--register`.<br>
You can set `global.logPath` to enable logging. Path may be absolute or relative. Leave empty to disable (default). It is very helpful when composing new rules without GUI editor, since you can see captured groups, arguments and resulting commandline.<br>
In `global.defaultConfigEditor` parameter you can specify path to your preferred text-editor. It will be used to open `linkrouter.json` when double-clicking `linkrouter.exe` or when selecting `Edit LinkRouter config` in right-click menu of executable (may be hidden inside "show more options"). If empty - an attempt to find any known text-editor in PATH is made.<br>

Here's a sample config to get the idea. Notice, that all backslashes `\` have to be escaped like this `\\` in JSON. GUI config editor does it automatically under the hood.

```json
{
  "global": {
    "fallbackBrowserPath": "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
    "fallbackBrowserArgs": "{URL}",
    "defaultConfigEditor": "C:\\Program Files\\Microsoft VS Code\\Code.exe",
    "logPath": "linkrouter.log",
    "supportedProtocols": [
      "http",
      "https",
      "ssh",
      "mailto"
    ]
  },
  "rules": [
     {
      "regex": "https://store\\.steampowered\\.com.*",
      "program": "C:\\Program Files (x86)\\Steam\\steam.exe",
      "arguments": "steam://openurl/{URL}"
    },
    {
      "regex": "ssh://(.*@company1\\.com).*",
      "program": "C:\\Windows\\System32\\wsl.exe",
      "arguments": "ssh $1"
    },
    {
      "regex": "ssh://(.*)/",
      "program": "C:\\Windows\\System32\\OpenSSH\\ssh.exe",
      "arguments": "-i .ssh/id_rsa_personal user2@$1"
    },
    {
      "regex": "mailto:(.*@company1\\.com)",
      "program": "C:\\Program Files\\Microsoft Office\\root\\Office16\\OUTLOOK.EXE",
      "arguments": "/c ipm.note /m $1"
    },
    {
      "regex": "mailto:(.*)",
      "program": "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
      "arguments": "https://mail.google.com/mail/?view=cm&fs=1&to=$1"
    }
  ]
}
```
this config will make LinkRouter:
- turn links like `https://store.steampowered.com/....` into `steam://openurl/https://store.steampowered.com/....` and open them in Steam.
- open links like `ssh://.*.company1.com` in wsl via ssh
- open all other links like `ssh://.*` in windows openssh using key id_rsa_personal
- open links like `mailto:.*@company1.com` by opening "New email" window in outlook with prefilled recipient field.
- open all other links like `mailto:.*` in gmail in chrome.
- links that don't match any rule will be opened in chrome browser.
- write log to `linkrouter.log` next to `linkrouter.exe`
- opens config for editing in VS Code

Tip: you can specify `explorer.exe` in `program` and pass link to it, if you want Windows to handle that link. e.g. passing `steam://` link to explorer will open Steam, since Steam is registered in Windows as the default handler for that protocol.<br>It is usually a good idea to quote resulting url in `arguments` to prevent breakage of complex links (e.g. with spaces).

> [!Note]
> While LinkRouter works just fine without running as an administrator, if a program from config is being run as admin, LinkRouter can't launch such program unless also launched with admin privileges. In this case go to `linkrouter.exe` `Properties` - `Compatibility` and check `Run this program as an administrator`.

Check more example rules in [linkrouter.example.json](linkrouter.example.json) in root of this repo. Maybe the app you need is already there. Send us your rule ideas — we’ll add them to the examples!

> [!Note]
> Figuring out the correct command-line arguments/switches for third-party programs is **entirely the user’s responsibility**. LinkRouter only launches whatever you tell it to launch.
For testing regexes we recommend enabling logging via `global.logPath` or using `linkrouter-gui.exe` or [this wonderful website](https://regex101.com/?flavor=golang) (choose the Golang flavor).

## 🔒 Privacy & Security
- Zero network access
- No telemetry, no analytics, no crash reporting
- No data collection of any kind
- Fully open-source
- Single static portable binary or user-friendly installer

> [!WARNING]
> Because LinkRouter can execute arbitrary programs, only use rules you trust. Never download and run someone else’s linkrouter.json blindly — it could contain malicious commands. LinkRouter doesn't launch any programs except for those, specified in your config.

## 📦 Download
See the [Releases page](https://github.com/kolbasky/LinkRouter/releases/latest) for the latest linkrouter.exe.

## 🛠️ Build from source
For building you'll need to install Go and MinGW-w64 (needed for gcc compiler).

```
# clone
git clone https://github.com/kolbasky/LinkRouter.git
cd LinkRouter

# this steps are optional, to embed icon, manifest and metadata.
# note, that recursive launch protection will not work without proper metadata.
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
go generate .\cmd\linkrouter\

# build
go build -ldflags="-H windowsgui -s -w" -trimpath -o bin\ .\cmd\linkrouter\
```
