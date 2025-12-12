package registry

import (
	"fmt"
	"linkrouter/internal/config"
	"linkrouter/internal/dialogs"
	"linkrouter/internal/logger"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const appName = "LinkRouter"
const appDescription = "regex-based router for links"

func getExePath() string {
	exe, _ := os.Executable()
	return exe
}

func ParseProtocol(proto string) string {
	re := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*).*$`)
	proto = strings.TrimSpace(proto)
	match := re.FindStringSubmatch(proto)
	if len(match) < 2 {
		dialogs.ShowError("wrong proto name:\n" + proto)
		return ""
	}

	return strings.ToLower(match[1])
}

func getSupportedProtocols() []string {
	cfg, err := config.LoadConfig()
	if err != nil {
		return []string{"http", "https"}
	}
	var protos []string
	for _, p := range cfg.Global.SupportedProtocols {
		if cleaned := ParseProtocol(p); cleaned != "" {
			protos = append(protos, cleaned)
		}
	}
	if len(protos) == 0 {
		protos = []string{"http", "https"}
	}
	return protos
}

func RegisterApp() error {
	UnregisterApp()
	exePath := getExePath()
	cmd := fmt.Sprintf(`"%s" "%%1"`, exePath)

	// Computer\HKEY_CURRENT_USER\Software\Clients\StartMenuInternet
	protocols := getSupportedProtocols()
	logger.Log("LinkRouter was launched with --register key")
	appPath := `Software\Clients\StartMenuInternet\` + appName

	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", appPath))
	k, _, err := registry.CreateKey(registry.CURRENT_USER, appPath, registry.ALL_ACCESS)
	if err != nil {
		logger.Log("failed to create StartMenuInternet key: " + err.Error())
		dialogs.ShowError("failed to create StartMenuInternet key:\n" + err.Error())
		return fmt.Errorf("failed to create StartMenuInternet key:\n%w", err)
	}
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\DisplayName", appPath))
	k.SetStringValue("DisplayName", appName)
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\ApplicationName", appPath))
	k.SetStringValue("ApplicationName", appName)
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\ApplicationDescription", appPath))
	k.SetStringValue("ApplicationDescription", appDescription)
	k.Close()

	capPath := appPath + `\Capabilities`
	// Computer\HKEY_CURRENT_USER\Software\Clients\StartMenuInternet\LinkRouter\Capabilities
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", capPath))
	cap, _, _ := registry.CreateKey(registry.CURRENT_USER, capPath, registry.ALL_ACCESS)
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\ApplicationName", capPath))
	cap.SetStringValue("ApplicationName", appName)
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\ApplicationIcon", capPath))
	cap.SetStringValue("ApplicationIcon", exePath+",0")
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\ApplicationDescription", capPath))
	cap.SetStringValue("ApplicationDescription", appDescription)

	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\URLAssociations", capPath))
	urlAssoc, _, _ := registry.CreateKey(cap, "URLAssociations", registry.ALL_ACCESS)

	// Computer\HKEY_CURRENT_USER\Software\Classes
	// Here we make sure protocols are present in windows and announce our URLAssociations.
	for _, proto := range protocols {
		classPath := `Software\Classes\` + proto
		logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", classPath))
		k, _, err := registry.CreateKey(registry.CURRENT_USER, classPath, registry.ALL_ACCESS)
		if err != nil {
			continue
		}
		k.SetStringValue("", "URL: "+proto+" Protocol")
		k.SetStringValue("URL Protocol", "")
		k.Close()

		urlAssoc.SetStringValue(proto, appName+"HTML")

	}
	cap.Close()

	// Computer\HKEY_CURRENT_USER\Software\RegisteredApplications
	logger.Log("Creating: HKEY_CURRENT_USER\\Software\\RegisteredApplications")
	regApps, _, _ := registry.CreateKey(registry.CURRENT_USER, `Software\RegisteredApplications`, registry.ALL_ACCESS)
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\Software\\RegisteredApplications\\%s", appName))
	regApps.SetStringValue(appName, capPath)
	regApps.Close()

	htmlClass := appName + "HTML"
	// Computer\HKEY_CURRENT_USER\Software\Classes\LinkRouterHTML
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\Software\\Classes\\%s", htmlClass))
	html, _, _ := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\`+htmlClass, registry.ALL_ACCESS)
	html.SetStringValue("", appName+" Document")
	html.SetStringValue("FriendlyTypeName", appName)

	shellPath := `Software\Classes\` + htmlClass + `\shell\open\command`
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", shellPath))
	shell, _, _ := registry.CreateKey(registry.CURRENT_USER, shellPath, registry.ALL_ACCESS)
	shell.SetStringValue("", cmd)
	shell.Close()
	html.Close()

	program := "explorer.exe"
	args := "ms-settings:defaultapps?registeredAppUser=LinkRouter"
	fullCmdLine := strconv.Quote(program) + " " + strconv.Quote(args)
	cmd_settings := exec.Command(program)
	cmd_settings.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fullCmdLine,
	}
	cmd_settings.Start()

	return nil
}

func UnregisterApp() error {
	config.LoadConfig()
	// Computer\HKEY_CURRENT_USER\Software\Clients\StartMenuInternet\LinkRouter
	logger.Log("LinkRouter was launched with --unregister key")
	logger.Log(
		"Removing: HKEY_CURRENT_USER\\Software\\Clients\\StartMenuInternet\\" + appName + "\\Capabilities\\URLAssociations")
	registry.DeleteKey(
		registry.CURRENT_USER,
		`Software\Clients\StartMenuInternet\`+appName+`\Capabilities\URLAssociations`)

	logger.Log(
		"Removing: HKEY_CURRENT_USER\\Software\\Clients\\StartMenuInternet\\" + appName + "\\Capabilities")
	registry.DeleteKey(
		registry.CURRENT_USER,
		`Software\Clients\StartMenuInternet\`+appName+`\Capabilities`)

	logger.Log(
		"Removing: HKEY_CURRENT_USER\\Software\\Clients\\StartMenuInternet\\" + appName)
	registry.DeleteKey(registry.CURRENT_USER, `Software\Clients\StartMenuInternet\`+appName)

	// Computer\HKEY_CURRENT_USER\Software\Classes\LinkRouterHTML
	htmlClass := appName + "HTML"
	htmlPath := `Software\Classes\` + htmlClass

	logger.Log(
		"Removing: HKEY_CURRENT_USER\\" + htmlPath + "\\shell\\open\\command")
	registry.DeleteKey(registry.CURRENT_USER, htmlPath+`\shell\open\command`)
	logger.Log(
		"Removing: HKEY_CURRENT_USER\\" + htmlPath + "\\shell\\open")
	registry.DeleteKey(registry.CURRENT_USER, htmlPath+`\shell\open`)
	logger.Log(
		"Removing: HKEY_CURRENT_USER\\" + htmlPath + "\\shell")
	registry.DeleteKey(registry.CURRENT_USER, htmlPath+`\shell`)
	logger.Log(
		"Removing: HKEY_CURRENT_USER\\" + htmlPath)
	registry.DeleteKey(registry.CURRENT_USER, htmlPath)

	// Computer\HKEY_CURRENT_USER\Software\RegisteredApplications
	logger.Log(
		"Removing: HKEY_CURRENT_USER\\Software\\RegisteredApplications\\" + appName)
	regAppsKey, err := registry.OpenKey(registry.CURRENT_USER, `Software\RegisteredApplications`, registry.SET_VALUE)
	if err == nil {
		regAppsKey.DeleteValue(appName)
		regAppsKey.Close()
	}

	return nil
}
