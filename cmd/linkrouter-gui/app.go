package main

import (
	"context"
	"encoding/json"
	"errors"
	"linkrouter/internal/config"
	"linkrouter/internal/dialogs"
	"linkrouter/internal/launcher"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetInteractiveMode() map[string]string {
	if *InteractiveMode && *InteractiveURL != "" {
		return map[string]string{
			"enabled": "true",
			"url":     *InteractiveURL,
		}
	}
	return map[string]string{"enabled": "false"}
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
		return "", nil
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

func (a *App) IsValidRegex(regexStr string) string {
	if regexStr == "" {
		return ""
	}
	_, err := regexp.Compile(regexStr)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) TestRegex(regexStr, url string) bool {
	if regexStr == "" {
		return false
	}
	re, err := regexp.Compile(regexStr)
	if err != nil {
		return false
	}
	return re.MatchString(url)
}

func (a *App) OpenFileDialog(title string, filters []runtime.FileFilter) (string, error) {
	if title == "" {
		title = "Select File"
	}

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
		return "", nil
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

func getExePath() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}

func (a *App) RegisterLinkRouter(silent bool) error {
	dir := getExePath()
	cmdPath := filepath.Join(dir, "linkrouter.exe")
	if _, err := os.Stat(cmdPath); err != nil {
		dialogs.ShowError("linkrouter.exe not found\nplace it near linkrouter-gui.exe")
		return nil
	}

	err := exec.Command(cmdPath, "--register", "--quiet").Start()
	if err != nil {
		dialogs.ShowError(err.Error())
		return err
	}
	if silent {
		err = exec.Command(cmdPath, "--default-apps").Start()
	}
	if err != nil {
		dialogs.ShowError(err.Error())
		return err
	}
	return nil
}

func (a *App) UnregisterLinkRouter() error {
	dir := getExePath()
	cmdPath := filepath.Join(dir, "linkrouter.exe")
	if _, err := os.Stat(cmdPath); err != nil {
		dialogs.ShowError("linkrouter.exe not found\nplace it near linkrouter-gui.exe")
		return nil
	}
	fullCmdLine := cmdPath + " --unregister"

	cmd := exec.Command(cmdPath)
	cmd.Path = cmdPath
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fullCmdLine,
	}
	return cmd.Start()
}

func (a *App) OpenInFallbackBrowser(browserPath string, url string) {
	if strings.TrimSpace(browserPath) == "" {
		dialogs.ShowError("fallback browser is not set\n" +
			"go to settings and set it up")
		return
	}
	err := launcher.LaunchApp(browserPath, "{URL}", url)
	if err != nil {
		dialogs.ShowError("unable to launch fallback browser: \n" + err.Error())
	}
}

func (a *App) OpenConfigInExplorer(configPath string) error {
	if configPath == "" {
		return nil
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("explorer", "/select,", absPath)
	return cmd.Start()
}

type Rule struct {
	Regex     string `json:"regex"`
	Program   string `json:"program"`
	Arguments string `json:"arguments"`
}

func (a *App) TestRule(rule Rule, url string) error {
	go func() {
		re, err := regexp.Compile(rule.Regex)
		if err != nil {
			dialogs.ShowError("Unable to complie regex:\n" + err.Error())
			return
		}
		matches := re.FindStringSubmatch(url)

		if len(matches) == 0 {
			dialogs.ShowError("Test URL doesn't match regex")
		}

		expandedArgs := launcher.ExpandPlaceholders(rule.Arguments, matches)
		err = launcher.LaunchApp(rule.Program, expandedArgs, url)
		if err != nil {
			dialogs.ShowError("Unable to launch program:\n" + err.Error())
		}
	}()
	return nil
}
