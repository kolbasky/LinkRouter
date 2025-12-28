//go:generate goversioninfo
package main

import (
	"embed"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS
var configPath string

// Config struct matching your linkrouter.json
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

func canWrite(path string) bool {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(path)
	return true
}

func getConfigPath() string {
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
	testFile := filepath.Join(exeDir, "linkrouter_write_test.9a928eb3-bfa9-4736-a262-00274e36d972")
	if canWrite(testFile) {
		return candidateExe
	}
	// Use localappdata then
	os.MkdirAll(filepath.Dir(candidateAppData), 0755)
	return candidateAppData
}

func (a *App) GetConfig() (*Config, error) {
	configPath = getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &Config{
			Global: GlobalConfig{}, // explicit if needed
			Rules:  []Rule{},
		}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Println("Error parsing config JSON:", err)
		return &Config{}, nil
	}

	return &cfg, nil
}

func (a *App) SaveConfig(cfg *Config) error {
	if configPath == "" {
		return errors.New("no file loaded")
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func (a *App) SaveConfigAs(cfg *Config) (string, error) {
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save LinkRouter Config",
		DefaultFilename: "linkrouter.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", nil // cancelled
	}
	if !strings.HasSuffix(strings.ToLower(filePath), ".json") {
		filePath += ".json"
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}
	configPath = filePath // update global
	return filePath, nil
}

func (a *App) GetCurrentConfigPath() string {
	configPath = getConfigPath()
	return configPath
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "cmd/linkrouter-gui",
		Width:     1024,
		Height:    768,
		MinWidth:  800,
		MinHeight: 600,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true, // or true if you want acrylic
			DisableWindowIcon:                 true,
			WebviewUserDataPath:               "",
			DisableFramelessWindowDecorations: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func (a *App) OpenFileDialog() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select LinkRouter Config File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", nil // User cancelled
	}
	return filePath, nil
}

// LoadConfigFromPath loads config from a user-selected path
func (a *App) LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
