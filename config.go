package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// DefaultConfig returns a sensible default config
func DefaultConfig() *Config {
	return &Config{
		Global: GlobalConfig{
			DefaultBrowserPath: "",
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
