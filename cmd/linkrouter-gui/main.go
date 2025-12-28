//go:generate goversioninfo
package main

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

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

// GetConfig returns the current config
//
//export GetConfig
func (a *App) GetConfig() (*Config, error) {
	exe, _ := os.Executable()
	confPath := filepath.Dir(exe)
	data, err := os.ReadFile(filepath.Join(confPath, "linkrouter.json"))
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

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "cmd/linkrouter-gui",
		Width:     1024,
		Height:    768,
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
