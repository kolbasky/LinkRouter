package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	DefaultBrowserPath string `json:"defaultBrowserPath"`
	DefaultBrowserArgs string `json:"defaultBrowserArgs"`
	InteractiveMode    bool   `json:"interactiveMode"`
	LaunchAtStartup    bool   `json:"launchAtStartup"`
	DaemonMode         bool   `json:"daemonMode"`
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

// isSameBinary checks if two paths point to the same executable content
func isSameBinary(path1, path2 string) bool {
	if path1 == "" || path2 == "" {
		return false
	}
	hash1, err1 := hashFile(path1)
	hash2, err2 := hashFile(path2)
	return err1 == nil && err2 == nil && hash1 == hash2
}

// DefaultConfig returns a sensible default config
func DefaultConfig() *Config {
	browserPath := getDefaultBrowserPath()

	// get our executable path to avoid recursive launches
	exePath, _ := os.Executable()
	exePath = filepath.Clean(exePath)
	if browserPath != "" && isSameBinary(exePath, browserPath) {
		browserPath = ""
	}

	if browserPath == "" {
		browserPath = `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
		// Also check 64-bit path if needed
		if _, err := os.Stat(browserPath); os.IsNotExist(err) {
			browserPath = `C:\Program Files\Microsoft\Edge\Application\msedge.exe`
		}
		if _, err := os.Stat(browserPath); os.IsNotExist(err) {
			browserPath = "" // leave empty if Edge not found
		}
	}

	return &Config{
		Global: GlobalConfig{
			DefaultBrowserPath: browserPath,
			DefaultBrowserArgs: "{URL}",
			InteractiveMode:    true,
			LaunchAtStartup:    false,
			DaemonMode:         false,
		},
		Rules: []Rule{},
	}
}

// LoadConfig loads config from file next to the executable
func LoadConfig() (*Config, error) {
	exePath, _ := os.Executable()
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Create default config
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

	return &cfg, nil
}

// Save writes config to file
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
			fmt.Fprintf(os.Stderr, "⚠️ Invalid regex: %v\n", err)
			continue
		}
		if matches := re.FindStringSubmatch(url); len(matches) > 0 {
			return &rule, matches
		}
	}
	return nil, nil
}
