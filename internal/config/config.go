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
	"strings"

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
	DefaultBrowserPath string   `json:"defaultBrowserPath"`
	DefaultBrowserArgs string   `json:"defaultBrowserArgs"`
	LogPath            string   `json:"logPath"`
	SupportedProtocols []string `json:"supportedProtocols"`
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
	userChoiceKey, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\Shell\Associations\UrlAssociations\https\UserChoice`,
		registry.READ)
	if err == nil {
		defer userChoiceKey.Close()
	}
	err = nil

	progId, _, _ := userChoiceKey.GetStringValue("Progid")

	// Step 2: Get command from HKCR\<ProgId>\shell\open\command
	cmdKey, err := registry.OpenKey(registry.CLASSES_ROOT, progId+`\shell\open\command`, registry.READ)
	if err == nil {
		defer userChoiceKey.Close()
	}

	cmdLine, _, _ := cmdKey.GetStringValue("")

	// Step 3: Extract quoted executable
	re := regexp.MustCompile(`^"([^"]+)"`)
	matches := re.FindStringSubmatch(cmdLine)
	if len(matches) > 1 && !utils.IsLinkRouter(matches[1]) {
		return matches[1]
	}

	// Fallback: first token
	parts := strings.Fields(cmdLine)
	first_token := strings.ReplaceAll(parts[0], "\"", "")
	if len(parts) > 0 && !utils.IsLinkRouter(first_token) {
		return parts[0]
	}

	if utils.IsLinkRouter(first_token) || utils.IsLinkRouter(matches[1]) {
		dialogs.ShowError("LinkRouter is already set as default browser. Trying to guess fallback browser.")
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
			return os.ExpandEnv(path)
		}
	}

	return ""
}

func getConfigPath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		userConfig := filepath.Join(localAppData, "LinkRouter", "linkrouter.json")
		if _, err := os.Stat(userConfig); err == nil {
			return userConfig
		}
		userConfig = filepath.Join(localAppData, "linkrouter.json")
		if _, err := os.Stat(userConfig); err == nil {
			return userConfig
		}
	}

	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "linkrouter.json")
}

func DefaultConfig() *Config {
	browserPath := getDefaultBrowserPath()

	if browserPath != "" && utils.IsLinkRouter(browserPath) {
		browserPath = ""
	}

	return &Config{
		Global: GlobalConfig{
			DefaultBrowserPath: browserPath,
			DefaultBrowserArgs: "{URL}",
			LogPath:            "",
			SupportedProtocols: []string{"http", "https"},
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
	configPath := getConfigPath()

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

	if utils.IsLinkRouter(cfg.Global.DefaultBrowserPath) {
		dialogs.ShowError("fallback browser is set to LinkRouter itself.\nusing fallback")
		cfg.Global.DefaultBrowserPath = getDefaultBrowserPath()
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
