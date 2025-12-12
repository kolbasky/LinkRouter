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
	fallbackBrowser := ""
	// if not found in registry - search known file locations
	fallback_candidates := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
		`C:\Program Files (x86)\BraveSoftware\Brave-Browser\Application\brave.exe`,
		`C:\Program Files\Mozilla Firefox\firefox.exe`,
		`C:\Program Files (x86)\Mozilla Firefox\firefox.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Internet Explorer\iexplore.exe`,
		`C:\Program Files (x86)\Internet Explorer\iexplore.exe`,
	}
	for _, path := range fallback_candidates {
		if _, err := os.Stat(path); err == nil {
			fallbackBrowser = path
			break
		}
	}

	// try to get default browser from registry
	// Step 1: Get ProgId from UserChoice for .html
	userChoiceKey, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\Shell\Associations\UrlAssociations\https\UserChoice`,
		registry.READ)
	if err != nil {
		return fallbackBrowser
	}
	defer userChoiceKey.Close()

	progId, _, err := userChoiceKey.GetStringValue("Progid")
	if err != nil || progId == "" {
		return fallbackBrowser
	}

	// Step 2: Get command from HKCR\<ProgId>\shell\open\command
	cmdKey, err := registry.OpenKey(registry.CLASSES_ROOT, progId+`\shell\open\command`, registry.READ)
	if err != nil {
		return fallbackBrowser
	}
	defer cmdKey.Close()

	cmdLine, _, err := cmdKey.GetStringValue("")
	if err != nil {
		return fallbackBrowser
	}

	if strings.Contains(strings.ToLower(cmdLine), "cmd.exe") {
		return fallbackBrowser
	}
	// Step 3: Extract quoted executable
	re := regexp.MustCompile(`^"([^"]+)"`)
	matches := re.FindStringSubmatch(cmdLine)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: first token
	parts := strings.Fields(cmdLine)
	if len(parts) > 0 {
		return parts[0]
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

	if browserPath == "" {
		browserPath = `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
		if _, err := os.Stat(browserPath); os.IsNotExist(err) {
			browserPath = `C:\Program Files\Microsoft\Edge\Application\msedge.exe`
		}
		if _, err := os.Stat(browserPath); os.IsNotExist(err) {
			browserPath = ""
		}
	}

	return &Config{
		Global: GlobalConfig{
			DefaultBrowserPath: browserPath,
			DefaultBrowserArgs: "{URL}",
			LogPath:            "",
			SupportedProtocols: []string{"http://", "https://"},
		},
		Rules: []Rule{},
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

	logger.Init(cfg.Global.LogPath)

	if utils.IsLinkRouter(cfg.Global.DefaultBrowserPath) {
		dialogs.ShowError("fallback browser is set to LinkRouter itself.\nusing Edge as fallback")
		cfg.Global.DefaultBrowserPath = "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"
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
