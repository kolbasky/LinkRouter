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
	data, err := os.ReadFile(filepath.Join(confPath, "linkrouter.json")) // adjust path as needed
	if err != nil {
		log.Println("Error reading config:", err)
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Println("Error parsing config JSON:", err)
		return nil, err
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
