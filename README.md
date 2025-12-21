# ![LinkRouter icon](resources/icon/32.png?raw=true) LinkRouter

A lightweight portable Windows app that routes links to specific applications based on **regex rules**.

Windows lets you choose a program to handle specific protocols, but there is no way to choose an app based on link contents. This app aims to fill that gap and suits power users who want **total control** over how links open on their machine.

- ✅ Open links like `https://store.steampowered.com/…`, `https://music.yandex.ru/…` etc. directly in their native apps (Steam, Discord, Spotify, YaMusic etc)
- ✅ Route different domains to different browsers, browser profiles, or open them in incognito/private mode
- ✅ Open mailto and ssh links from different companies in different apps or with different args
- ✅ Add custom regex rules for **any app** and **any protocol**
- ✅ Use capture groups to reformat the URL any way you want
- ✅ All unhandled links fall back to your default browser (unchanged behavior)
- ✅ **No installer**, **no telemetry**, **no network access**
- ✅ Tiny, fast, single .exe
- ✅ Zero memory footprint - fire and exit.

> [!WARNING]
> There are some false-positives on VirusTotal for this program: [3 out of 72 AVs mark this file suspicious](https://www.virustotal.com/gui/file/22e1ce428e06fce8556077a719db7830c7b32627d7face91b5c59e28d6c9a18e). We can do nothing about it at the moment without sacrificing functionality. If the project lives we will try code signing and contacting AV vendors.

## 🚀 Quick Start

1. **Download** [`linkrouter.exe`](https://github.com/kolbasky/LinkRouter/releases/latest)
2. Place `linkrouter.exe` where you want and **run it** by double-clicking.
3. Select "Yes" in dialog to register the app in the system.
4. `Windows Settings` → `Apps` → `Default apps` dialog should pop-up automatically. If not press `Win+I` and start typing "default".
5. **select LinkRouter as the default handler for HTTP, HTTPS, or any other protocols** you want it to handle.
6. **Edit the config** by double-clicking `linkrouter.exe` again.

> [!NOTE]
> When registered, double-clicking `linkrouter.exe` opens the config for editing.
> Additional right-click menu entries are available on `linkrouter.exe` for you convenience after registration (may be hidden inside "show more options"):
> - Register LinkRouter
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
The app auto-creates `linkrouter.json` next to executable on its first launch and tries to detect your current default browser to use as the fallback one. If it fails, it tries to guess one from a list of known popular browsers locations.
When loading config, LinkRouter checks:
- `%LOCALAPPDATA%\LinkRouter\linkrouter.json`
- `linkrouter.json` in the same folder as the executable
When creating a new config, it tries to create it next to the executable first. If that fails (e.g., in Program Files without admin rights), it falls back to `%LOCALAPPDATA%\LinkRouter\linkrouter.json`.

Every link passed to LinkRouter is tested against the rules in order. The first matching rule wins.

- `regex` – Golang-flavored regular expression
- `program` – full path to the target executable
- `arguments` – command-line arguments; `{URL}` is replaced with the original link, `$1`, `$2`… are replaced with capture-group contents

Links that do not match any rule are passed to `global.fallbackBrowserPath` with `global.fallbackBrowserArgs` as arguments.

You can handle any protocol (mailto, ssh, steam, spotify, etc.). Just add the protocol to `global.supportedProtocols` and re-run `--register`.<br>
You can set `global.logPath` to enable logging. Path may be absolute or relative. Leave empty to disable (default). It is very helpful when composing new rules, since you can see captured groups, arguments and resulting commandline.<br>
In `global.defaultConfigEditor` parameter you can specify path to you preferred text-editor. It will be used to open `linkrouter.json` when double-clicking `linkrouter.exe` or when selecting `Edit LinkRouter config` in right-click menu of executable (may be hidden inside "show more options"). If empty - an attempt to find any known text-editor in PATH is made.<br>

Here's a sample config to get the idea. Notice, that all backslashes `\` have to be escaped like this `\\` in JSON.

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
- opens links like `ssh://.*.company1.com` in wsl in ssh
- opens all other links like `ssh://.*` in windows openssh using key id_rsa_personal
- opens links like `mailto:.*@company1.com` by opening "New email" window in outlook with prefilled recipient field.
- opens all other links like `mailto:.*` in gmail in chrome.
- links that don't match any rule will be opened in chrome browser.
- write log to `linkrouter.log` next to `linkrouter.exe`
- open config for editing in VSCode

Tip: you can specify `explorer.exe` in program and pass link to it, if you want Windows to handle that link. i.e. passing steam:// link to explorer will open Steam, since Steam is registered in Windows as the default handler for that protocol.

> [!Note]
> While LinkRouter works just fine without running as an administrator, if a program from config is being run as admin, LinkRouter can't launch such program unless also launched with admin privileges. In this case go to `linkrouter.exe` `Properties` - `Compatibility` and check `Run this program as an administrator`.

Check more example rules in [linkrouter.example.json](linkrouter.example.json) in root of this repo. Maybe the app you need is already there.

> [!Note]
> Figuring out the correct command-line arguments/switches for third-party programs is **entirely the user’s responsibility**. LinkRouter only launches whatever you tell it to launch.
For testing regexes we recommend enabling logging via `global.logPath` or using [this wonderful website](https://regex101.com/?flavor=golang) (choose the Golang flavor).

## 🔒 Privacy & Security
- Zero network access
- No telemetry, no analytics, no crash reporting
- No data collection of any kind
- Fully open-source
- Single static binary, portable, no installer

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
