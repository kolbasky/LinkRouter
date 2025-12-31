package main

import (
	"context"
	"encoding/json"
	"errors"
	"linkrouter/internal/config"
	"os"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

var configPath string

func (a *App) GetConfig() (*config.Config, error) {
	return config.LoadConfig()
}

func (a *App) SaveConfig(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	path := config.GetConfigPath()
	return cfg.Save(path)
}

func (a *App) SaveConfigAs(cfg *config.Config) (string, error) {
	if cfg == nil {
		return "", errors.New("config is nil")
	}

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

	if err := cfg.Save(filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func (a *App) GetCurrentConfigPath() string {
	configPath = config.GetConfigPath()
	return configPath
}

func (a *App) TestRegex(regexStr, url string) bool {
	if regexStr == "" {
		return false
	}
	re, err := regexp.Compile(regexStr)
	if err != nil {
		return false // invalid regex = no match
	}
	return re.MatchString(url)
}

func (a *App) OpenFileDialog(title string, filters []runtime.FileFilter) (string, error) {
	// Default to "Select File" if no title
	if title == "" {
		title = "Select File"
	}

	// Default filters if none provided
	if len(filters) == 0 {
		filters = []runtime.FileFilter{
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		}
	}

	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: filters,
	})
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", nil // user cancelled
	}
	return filePath, nil
}

// For config files
func (a *App) OpenConfigDialog() (string, error) {
	return a.OpenFileDialog("Select LinkRouter Config File", []runtime.FileFilter{
		{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		{DisplayName: "All Files (*.*)", Pattern: "*.*"},
	})
}

// For programs
func (a *App) OpenProgramDialog() (string, error) {
	return a.OpenFileDialog("Select Program", []runtime.FileFilter{
		{DisplayName: "Exe Files (*.exe)", Pattern: "*.exe"},
		{DisplayName: "All Files (*.*)", Pattern: "*.*"},
	})
}

// LoadConfigFromPath loads config from a user-selected path
func (a *App) LoadConfigFromPath(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
