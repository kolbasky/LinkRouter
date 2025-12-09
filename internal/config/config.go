package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"linkrouter/internal/dialogs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Config represents the full configuration
type Config struct {
	Global GlobalConfig `json:"global"`
	Rules  []Rule       `json:"rules"`
}

// GlobalConfig holds global settings
type GlobalConfig struct {
	DefaultBrowserPath string   `json:"defaultBrowserPath"`
	DefaultBrowserArgs string   `json:"defaultBrowserArgs"`
	InteractiveMode    bool     `json:"interactiveMode"`
	LaunchAtStartup    bool     `json:"launchAtStartup"`
	DaemonMode         bool     `json:"daemonMode"`
	SupportedProtocols []string `json:"supportedProtocols"`
}

// Rule defines a URL routing rule
type Rule struct {
	Regex     string `json:"regex"`
	Program   string `json:"program"`
	Arguments string `json:"arguments"`
}

func getDefaultBrowserPath() string {
	fallbackBrowser := "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"
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

	return fallbackBrowser
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// still need a better way to identify ourselves
func IsLinkRouter(path string) bool {
	exePath, _ := os.Executable()
	if strings.EqualFold(filepath.Clean(path), filepath.Clean(exePath)) {
		return true
	}

	hash1, err1 := hashFile(path)
	hash2, err2 := hashFile(exePath)

	return err1 == nil && err2 == nil && hash1 == hash2
}

func DefaultConfig() *Config {
	browserPath := getDefaultBrowserPath()

	if browserPath != "" && IsLinkRouter(browserPath) {
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
			InteractiveMode:    false,
			LaunchAtStartup:    false,
			DaemonMode:         false,
			SupportedProtocols: []string{"http://", "https://"},
		},
		Rules: []Rule{},
	}
}

func LoadConfig() (*Config, error) {
	exePath, _ := os.Executable()
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")

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

	if IsLinkRouter(cfg.Global.DefaultBrowserPath) {
		dialogs.ShowError("Fallback browser is set to LinkRouter itself failing back to Edge.")
		cfg.Global.DefaultBrowserPath = "C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe"
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) MatchRule(url string) (*Rule, []string) {
	for _, rule := range c.Rules {
		re, err := regexp.Compile(rule.Regex)
		if err != nil {
			dialogs.ShowError("Invalid regex:\n" + err.Error())
			continue
		}
		if matches := re.FindStringSubmatch(url); len(matches) > 0 {
			return &rule, matches
		}
	}
	return nil, nil
}
