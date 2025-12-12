# ![LinkRouter icon](resources/icon/48.png?raw=true) LinkRouter 

A lightweight portable Windows app that routes links to specific applications based on **regex rules**.

Windows lets you choose a program to handle specific protocols, but there is no way to choose an app based on link contents. This app aims to fill that gap and suits power users who want **total control** over how links open on their machine.

- ✅ Open links like `https://store.steampowered.com/…`, `https://music.yandex.ru/…` etc. directly in their native apps  
- ✅ Route different domains to different browsers, browser profiles, or open them in incognito/private mode  
- ✅ Add custom regex rules for **any app** and **any protocol**  
- ✅ Use capture groups to reformat the URL any way you want  
- ✅ All unhandled links fall back to your default browser (unchanged behavior)  
- ✅ **No installer**, **no telemetry**, **no network access whatsoever**  
- ✅ Tiny, fast, single .exe
- ✅ Zero memory footprint - fire and exit.

## 🚀 Quick Start

1. **Download** [`linkrouter.exe`](https://github.com/kolbasky/link-router/releases/latest) 
2. **Open PowerShell or Command Prompt** in folder where `linkrouter.exe` is placed
3. Run:
   ```powershell
   .\linkrouter.exe --register
   ```
   this will create registry keys, necessary for setting LinkRouter as a browser. Use `--unregister` to remove the registry entries later.
4. Go to `Windows Settings` → `Apps` → `Default apps` and **select LinkRouter as the default handler for HTTP, HTTPS, or any other protocols** you want it to handle. You only need to do this once; LinkRouter will then intercept links for those protocols.
5. **Edit the config** next to executable and add your rules (see example below).
6. **Optionally**, move config to `%LOCALAPPDATA%\LinkRouter\linkrouter.json`


## ⚙️ Configuration
The app auto-creates `linkrouter.json` next to executable on its first launch and tries to detect your current default browser to use as the fallback one. If it fails, it defaults to Edge.
User may store config in one of these places (searched in this order):
  - %LOCALAPPDATA%\LinkRouter\linkrouter.json
  - .\linkrouter.json

Every link passed to LinkRouter is tested against the rules in order. The first matching rule wins.

- `regex` – Golang-flavored regular expression
- `program` – full path to the target executable
- `arguments` – command-line arguments; `{URL}` is replaced with the original link, `$1`, `$2`… are replaced with capture-group contents

You can handle any protocol (mailto, ssh, steam, spotify, etc.). Just add the protocol to `global.supportedProtocols` and re-run `--register`.

Here's a sample config to get the idea. Notice, that all backslashes `\` have to be escaped like this `\\` in JSON.

```json
{
  "global": {
    "defaultBrowserPath": "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
    "defaultBrowserArgs": "{URL}",
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
      "regex": "ssh://(.*@company1\\.com)/",
      "program": "C:\\Windows\\System32\\OpenSSH\\ssh.exe",
      "arguments": "-i .ssh/id_rsa_work user1@$1"
    },
    {
      "regex": "ssh://(.*)/",
      "program": "C:\\Windows\\System32\\OpenSSH\\ssh.exe",
      "arguments": "-i .ssh/id_rsa_personal user2@$1"
    },
    {
      "regex": "mailto:.*@company1\\.com",
      "program": "C:\\Program Files\\Microsoft Office\\root\\Office16\\OUTLOOK.EXE",
      "arguments": "/c ipm.note /profile \"work\" /m $1"
    },
    {
      "regex": "mailto:(.*)",
      "program": "C:\\Program Files\\Microsoft Office\\root\\Office16\\OUTLOOK.EXE",
      "arguments": "/c ipm.note /profile \"personal\" /m $1"
    }
  ]
}
```
this config will make LinkRouter:
- turn links like `https://store.steampowered.com/....` into `steam://openurl/https://store.steampowered.com/....` and open them in Steam.
- opens links like `ssh://.*.company1.com` in openssh using key id_rsa_work
- opens all other links like `ssh://.*` in openssh using key id_rsa_personal
- opens links like `mailto:.*@company1.com` by opening "New email" window in outlook with `work` profile and prefilled recipient field.
- opens all other links like `mailto:.*` by opening "New email" window in outlook with `personal` profile and prefilled recipient field.
- links that don't match any rule will be opened in chrome browser.
> [!Note]
> While LinkRouter works just fine without running as administrator, if a program from config is being run as admin, LinkRouter can't launch such program unless also launched with admin privileges. Go to linkrouter.exe properties - Compatibility and check "Run this programm as an administrator"

Check more example rules in [linkrouter.json.example](https://github.com/kolbasky/LinkRouter/blob/main/linkrouter.json.example) in root of this repo. Maybe the app you need is already there.

> [!Note]
> Figuring out the correct command-line arguments/switches for third-party programs is **entirely the user’s responsibility**. LinkRouter only launches whatever you tell it to launch.
For testing regexes we recommend [this wonderful website](https://regex101.com/?flavor=golang) (choose the Golang flavor).

## 🔒 Privacy & Security
- Zero network access
- No telemetry, no analytics, no crash reporting
- No data collection of any kind
- Fully open-source
- Single static binary, portable, no installer

> [!WARNING]
> Because LinkRouter can execute arbitrary programs, only use rules you trust. Never download and run someone else’s linkrouter.json blindly — it could contain malicious commands. LinkRouter doesn't launch any programs except for those, specified in your config.

## 📦 Download
See the [Releases page](https://github.com/kolbasky/link-router/releases/latest) for the latest linkrouter.exe.

## 🛠️ Build from source
```
git clone https://github.com/kolbasky/LinkRouter.git
go build -ldflags="-H windowsgui -s -w" -trimpath -o bin\ .\cmd\linkrouter\
```
