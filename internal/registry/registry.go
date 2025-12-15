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
	var criticalError error
	appPath := `Software\Clients\StartMenuInternet\` + appName
	linkrouterClass := appName

	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", appPath))
	k, _, err := registry.CreateKey(registry.CURRENT_USER, appPath, registry.ALL_ACCESS)
	if err != nil {
		criticalError = fmt.Errorf("failed to create registry key: %w", err)
		logger.Log(criticalError.Error())
		err = nil
	} else {
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\DisplayName", appPath))
		k.SetStringValue("DisplayName", appName)
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\ApplicationName", appPath))
		k.SetStringValue("ApplicationName", appName)
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\ApplicationDescription", appPath))
		k.SetStringValue("ApplicationDescription", appDescription)
		k.Close()
	}

	capPath := appPath + `\Capabilities`
	// Computer\HKEY_CURRENT_USER\Software\Clients\StartMenuInternet\LinkRouter\Capabilities
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", capPath))
	cap, _, err := registry.CreateKey(registry.CURRENT_USER, capPath, registry.ALL_ACCESS)
	if err != nil {
		criticalError = fmt.Errorf("failed to create registry key: %w", err)
		logger.Log(criticalError.Error())
		err = nil
	} else {
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\ApplicationName", capPath))
		cap.SetStringValue("ApplicationName", appName)
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\ApplicationIcon", capPath))
		cap.SetStringValue("ApplicationIcon", exePath+",0")
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\ApplicationDescription", capPath))
		cap.SetStringValue("ApplicationDescription", appDescription)
	}

	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s\\URLAssociations", capPath))
	urlAssoc, _, err := registry.CreateKey(cap, "URLAssociations", registry.ALL_ACCESS)
	if err != nil {
		criticalError = fmt.Errorf("failed to create registry key: %w", err)
		logger.Log(criticalError.Error())
		err = nil
	}

	// Computer\HKEY_CURRENT_USER\Software\Classes
	// Here we make sure protocols are present in windows and announce our URLAssociations.
	for _, proto := range protocols {
		classPath := `Software\Classes\` + proto
		logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", classPath))
		k, _, err := registry.CreateKey(registry.CURRENT_USER, classPath, registry.ALL_ACCESS)
		if err != nil {
			criticalError = fmt.Errorf("failed to create registry key for protocol: %w", err)
			logger.Log(criticalError.Error())
			err = nil
			continue
		} else {
			logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\(Default)", classPath))
			k.SetStringValue("", "URL: "+proto+" Protocol")
			logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\Url Protocol", classPath))
			k.SetStringValue("URL Protocol", "")
			k.Close()
			logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\URLAssociations\\%s\\(Default)", capPath, proto))
			urlAssoc.SetStringValue(proto, appName)
		}
	}
	cap.Close()

	// Computer\HKEY_CURRENT_USER\Software\RegisteredApplications
	logger.Log("Creating: HKEY_CURRENT_USER\\Software\\RegisteredApplications")
	regApps, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\RegisteredApplications`, registry.ALL_ACCESS)
	if err != nil {
		criticalError = fmt.Errorf("failed to create registry key: %w", err)
		logger.Log(criticalError.Error())
		err = nil
	} else {
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\Software\\RegisteredApplications\\%s", appName))
		regApps.SetStringValue(appName, capPath)
		regApps.Close()
	}

	// Computer\HKEY_CURRENT_USER\Software\Classes\LinkRouter
	logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\Software\\Classes\\%s", linkrouterClass))
	html, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\`+linkrouterClass, registry.ALL_ACCESS)
	if err != nil {
		criticalError = fmt.Errorf("failed to create registry key: %w", err)
		logger.Log(criticalError.Error())
		err = nil
	} else {
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\Software\\Classes\\%s\\(Default)", linkrouterClass))
		html.SetStringValue("", appName+" Document")
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\Software\\Classes\\%s\\FriendlyTypeName", linkrouterClass))
		html.SetStringValue("FriendlyTypeName", appName)

		shellPath := `Software\Classes\` + linkrouterClass + `\shell\open\command`
		logger.Log(fmt.Sprintf("Creating: HKEY_CURRENT_USER\\%s", shellPath))
		shell, _, err := registry.CreateKey(registry.CURRENT_USER, shellPath, registry.ALL_ACCESS)
		if err != nil {
			criticalError = fmt.Errorf("failed to create registry key: %w", err)
			logger.Log("failed to create registry key: " + err.Error())
			err = nil
		}
		logger.Log(fmt.Sprintf("Setting: HKEY_CURRENT_USER\\%s\\(Default)", shellPath))
		shell.SetStringValue("", cmd)
		shell.Close()
		html.Close()
	}

	program := `${SYSTEMROOT}\explorer.exe`
	args := "ms-settings:defaultapps?registeredAppUser=LinkRouter"
	fullCmdLine := strconv.Quote(os.ExpandEnv(program)) + " " + strconv.Quote(args)
	cmd_settings := exec.Command(os.ExpandEnv(program))
	cmd_settings.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fullCmdLine,
	}
	err = cmd_settings.Start()
	if err != nil {
		logger.Log("Error: can't open windows settings.")
		msg := "Registered successfully. Now set LinkRouter as defaul app for desired link types in Windows Settings (Win+I and start typing \"default\")"
		dialogs.ShowMessageBox("LinkRouter", msg, 0x00000040)
		return nil
	}

	if criticalError != nil {
		dialogs.ShowError("Errors during registration.\nSet global.logPath in linkrouter.json and rerun --register for troubleshooting.")
		logger.Log("Registration failed.")
		return criticalError
	}

	logger.Log("Registration completed successfully")
	return nil
}

func UnregisterApp() error {
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

	// Computer\HKEY_CURRENT_USER\Software\Classes\LinkRouter
	linkrouterClass := appName
	htmlPath := `Software\Classes\` + linkrouterClass

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
