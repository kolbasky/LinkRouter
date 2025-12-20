package config

import (
	"encoding/json"
	"fmt"
	"linkrouter/internal/dialogs"
	"linkrouter/internal/logger"
	"linkrouter/internal/utils"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/sys/windows/registry"
)

var SupportedProtocols []string

// Config represents the full configuration
type Config struct {
	Global GlobalConfig `json:"global"`
	Rules  []Rule       `json:"rules"`
}

// GlobalConfig holds global settings
type GlobalConfig struct {
	FallbackBrowserPath string   `json:"fallbackBrowserPath"`
	FallbackBrowserArgs string   `json:"fallbackBrowserArgs"`
	DefaultConfigEditor string   `json:"defaultConfigEditor"`
	LogPath             string   `json:"logPath"`
	SupportedProtocols  []string `json:"supportedProtocols"`
}

// Rule defines a URL routing rule
type Rule struct {
	Regex     string `json:"regex"`
	Program   string `json:"program"`
	Arguments string `json:"arguments"`
}

func getDefaultBrowserPath() string {
	// try to get default browser from registry
	// Step 1: Get ProgId from UserChoice for .html
	logger.Log("Trying to get default system browser.")
	userChoiceKey, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\Shell\Associations\UrlAssociations\https\UserChoice`,
		registry.READ)
	if err == nil {
		defer userChoiceKey.Close()
	}
	err = nil

	progId, _, _ := userChoiceKey.GetStringValue("Progid")
	logger.Log("Got browser progid from registry: " + progId)
	// Step 2: Get command from HKCR\<ProgId>\shell\open\command
	cmdKey, err := registry.OpenKey(registry.CLASSES_ROOT, progId+`\shell\open\command`, registry.READ)
	if err == nil {
		defer userChoiceKey.Close()
	}

	cmdLine, _, _ := cmdKey.GetStringValue("")
	logger.Log("Got cmdline from registry: " + cmdLine)
	// Step 3: Extract quoted executable
	if len(cmdLine) > 0 {
		re := regexp.MustCompile(`^"([^"]+)"`)
		matches := re.FindStringSubmatch(cmdLine)
		if len(matches) > 1 && !utils.IsLinkRouter(matches[1]) {
			logger.Log("Found " + matches[1])
			return matches[1]
		}
		if utils.IsLinkRouter(matches[1]) {
			logger.Log("LinkRouter is already set as default browser. Trying to guess fallback browser.")
			dialogs.ShowError("LinkRouter is already set as default browser. Trying to guess fallback browser.")
		}
	}

	// if not found in registry - search known file locations
	fallback_candidates := []string{
		// Chrome
		`${ProgramFiles}\Google\Chrome\Application\chrome.exe`,
		`${ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe`,
		`${LOCALAPPDATA}\Google\Chrome\Application\chrome.exe`,
		// Chrome canary
		`${ProgramFiles}\Google\Chrome SxS\Application\chrome.exe`,
		`${ProgramFiles(x86)}\Google\Chrome SxS\Application\chrome.exe`,
		`${LOCALAPPDATA}\Google\Chrome SxS\Application\chrome.exe`,
		// Brave
		`${ProgramFiles}\BraveSoftware\Brave-Browser\Application\brave.exe`,
		`${ProgramFiles(x86)}\BraveSoftware\Brave-Browser\Application\brave.exe`,
		`${LOCALAPPDATA}\BraveSoftware\Brave-Browser\Application\brave.exe`,
		// Firefox
		`${ProgramFiles}\Mozilla Firefox\firefox.exe`,
		`${ProgramFiles(x86)}\Mozilla Firefox\firefox.exe`,
		`${LOCALAPPDATA}\Mozilla Firefox\firefox.exe`,
		// Yandex Browser
		`${LOCALAPPDATA}\Yandex\YandexBrowser\Application\browser.exe`,
		// Vivaldi
		`${LOCALAPPDATA}\Vivaldi\Application\vivaldi.exe`,
		`${ProgramFiles}\Vivaldi\Application\vivaldi.exe`,
		`${ProgramFiles(x86)}\Vivaldi\Application\vivaldi.exe`,
		// Opera
		`${LOCALAPPDATA}\Programs\Opera\launcher.exe`,
		`${ProgramFiles}\Opera\launcher.exe`,
		// Edge
		`${ProgramFiles}\Microsoft\Edge\Application\msedge.exe`,
		`${ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe`,
		`${LOCALAPPDATA}\Microsoft\Edge\Application\msedge.exe`,
		// iexplorer
		`${ProgramFiles}\Internet Explorer\iexplore.exe`,
		`${ProgramFiles(x86)}\Internet Explorer\iexplore.exe`,
	}
	for _, path := range fallback_candidates {
		if _, err := os.Stat(os.ExpandEnv(path)); err == nil {
			defaultBrowser := os.ExpandEnv(path)
			logger.Log("Found " + defaultBrowser)
			return defaultBrowser
		}
	}

	return ""
}

func canWrite(path string) bool {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(path)
	return true
}

func GetConfigPath() string {
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	candidateExe := filepath.Join(exeDir, "linkrouter.json")

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return candidateExe
	}
	candidateAppData := filepath.Join(localAppData, "LinkRouter", "linkrouter.json")

	// Prefer AppData if exists
	if _, err := os.Stat(candidateAppData); err == nil {
		return candidateAppData
	}
	if _, err := os.Stat(candidateExe); err == nil {
		return candidateExe
	}

	// No config exists exist. Try exedir first
	// But may fail when in Program Files and without admin privs
	testFile := filepath.Join(exeDir, "linkrouter_write_test.9a928eb3-bfa9-4736-a262-00274e36d973")
	if canWrite(testFile) {
		return candidateExe
	}
	// Use localappdata then
	dialogs.ShowError("Unable to create config next to exe file. Config will be created in " + candidateAppData)
	os.MkdirAll(filepath.Dir(candidateAppData), 0755)
	return candidateAppData
}

func DefaultConfig() *Config {
	browserPath := getDefaultBrowserPath()

	if browserPath != "" && utils.IsLinkRouter(browserPath) {
		browserPath = ""
	}

	return &Config{
		Global: GlobalConfig{
			FallbackBrowserPath: browserPath,
			FallbackBrowserArgs: "{URL}",
			DefaultConfigEditor: "",
			LogPath:             "",
			SupportedProtocols:  []string{"http", "https"},
		},
		Rules: []Rule{
			{
				Regex:     `https://(.*)`,
				Program:   browserPath,
				Arguments: "{URL}",
			},
		},
	}
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		cfg := DefaultConfig()
		cfg.Save(configPath)
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	err_init := logger.Init(cfg.Global.LogPath)
	if err_init != nil {
		logger.Log(fmt.Sprintf("Error: can't open global.logPath %q, %s", cfg.Global.LogPath, err_init))
		dialogs.ShowError(fmt.Sprintf("can't open global.logPath %q, %s", cfg.Global.LogPath, err_init))
	}

	if utils.IsLinkRouter(cfg.Global.FallbackBrowserPath) {
		dialogs.ShowError("Fallback browser is set to LinkRouter itself.\nTrying to guess fallback browser.")
		logger.Log("Error: Fallback browser is set to LinkRouter itself. Trying to guess fallback browser.")
		cfg.Global.FallbackBrowserPath = getDefaultBrowserPath()
	}

	SupportedProtocols = cfg.Global.SupportedProtocols

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) MatchRule(url string) (*Rule, []string, int) {
	for i, rule := range c.Rules {
		re, err := regexp.Compile(rule.Regex)
		if err != nil {
			logger.Log("Invalid regex: " + err.Error())
			logger.Log(fmt.Sprintf("Failed rule: regex=%q", rule.Regex))
			dialogs.ShowError("invalid regex:\n" + err.Error())
			continue
		}
		if matches := re.FindStringSubmatch(url); len(matches) > 0 {
			return &rule, matches, i
		}
	}
	logger.Log("Matched rule: None")
	return nil, nil, -1
}
