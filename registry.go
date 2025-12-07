// registry.go
package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const appName = "LinkRouter"
const appDescription = "Smart link router with custom rules"

func getExePath() string {
	exe, _ := os.Executable()
	return exe // Already uses \ on Windows
}

func RegisterAsBrowser() error {
	exePath := getExePath()
	cmd := fmt.Sprintf(`"%s" "%%1"`, exePath)

	fmt.Println("📝 Registering browser with:")
	fmt.Printf("  EXE: %s\n", exePath)
	fmt.Printf("  CMD: %s\n", cmd)

	appPath := `Software\Clients\StartMenuInternet\` + appName
	fmt.Printf("  → HKCU\\%s\n", appPath)

	k, _, err := registry.CreateKey(registry.CURRENT_USER, appPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create StartMenuInternet key: %w", err)
	}
	k.SetStringValue("DisplayName", appName)
	k.SetStringValue("ApplicationName", appName)
	k.SetStringValue("ApplicationDescription", appDescription)
	k.Close()

	capPath := appPath + `\Capabilities`
	fmt.Printf("  → HKCU\\%s\n", capPath)
	cap, _, _ := registry.CreateKey(registry.CURRENT_USER, capPath, registry.ALL_ACCESS)
	cap.SetStringValue("ApplicationName", appName)
	cap.SetStringValue("ApplicationIcon", exePath+",0")
	cap.SetStringValue("ApplicationDescription", appDescription)

	urlAssoc, _, _ := registry.CreateKey(cap, "URLAssociations", registry.ALL_ACCESS)
	urlAssoc.SetStringValue("http", appName+"HTML")
	urlAssoc.SetStringValue("https", appName+"HTML")
	urlAssoc.Close()
	cap.Close()

	fmt.Printf("  → HKCU\\Software\\RegisteredApplications (%s)\n", appName)
	regApps, _, _ := registry.CreateKey(registry.CURRENT_USER, `Software\RegisteredApplications`, registry.ALL_ACCESS)
	regApps.SetStringValue(appName, capPath)
	regApps.Close()

	htmlClass := appName + "HTML"
	fmt.Printf("  → HKCU\\Software\\Classes\\%s\n", htmlClass)
	html, _, _ := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\`+htmlClass, registry.ALL_ACCESS)
	html.SetStringValue("", appName+" Document")

	shellPath := `Software\Classes\` + htmlClass + `\shell\open\command`
	fmt.Printf("    → CMD: %s\n", shellPath)
	shell, _, _ := registry.CreateKey(registry.CURRENT_USER, shellPath, registry.ALL_ACCESS)
	shell.SetStringValue("", cmd)
	shell.Close()
	html.Close()

	fmt.Println("✅ Registration complete.")
	fmt.Println("👉 Go to Settings → Default apps → Web browser → Choose LinkRouter")
	return nil
}

func UnregisterAsBrowser() error {
	// 1. Remove StartMenuInternet entry (and all children)
	registry.DeleteKey(registry.CURRENT_USER, `Software\Clients\StartMenuInternet\`+appName)

	// 2. Recursively remove LinkRouterHTML class
	htmlClass := appName + "HTML"
	htmlPath := `Software\Classes\` + htmlClass

	// Delete 'shell\open\command'
	registry.DeleteKey(registry.CURRENT_USER, htmlPath+`\shell\open\command`)
	// Delete 'shell\open'
	registry.DeleteKey(registry.CURRENT_USER, htmlPath+`\shell\open`)
	// Delete 'shell'
	registry.DeleteKey(registry.CURRENT_USER, htmlPath+`\shell`)
	// Now delete the root class
	registry.DeleteKey(registry.CURRENT_USER, htmlPath)

	// 3. Remove from RegisteredApplications
	regAppsKey, err := registry.OpenKey(registry.CURRENT_USER, `Software\RegisteredApplications`, registry.SET_VALUE)
	if err == nil {
		regAppsKey.DeleteValue(appName)
		regAppsKey.Close()
	}

	fmt.Println("✅ Safely unregistered as browser.")
	return nil
}
