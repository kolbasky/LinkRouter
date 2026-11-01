package main

import (
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

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	app *application.App
}

func NewApp(app *application.App) *App {
	return &App{app: app}
}

func (a *App) GetInteractiveMode() map[string]string {
	if *InteractiveMode && *InteractiveURL != "" {
		return map[string]string{
			"enabled": "true",
			"url":     *InteractiveURL,
		}
	} else if !*InteractiveMode && *InteractiveURL != "" {
		return map[string]string{
			"enabled": "false",
			"url":     *InteractiveURL,
		}
	} else {
		return map[string]string{"enabled": "false"}
	}
}

func (a *App) ShowCreateRule() bool {
	if *launchCreateRule && *InteractiveURL != "" {
		return true
	}
	return false
}

func (a *App) SpawnNewInstance(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	err = exec.Command(exe, args...).Start()
	return err
}

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

	saveDialog := a.app.Dialog.SaveFile()
	saveDialog.SetOptions(&application.SaveFileDialogOptions{
		Title:    "Save LinkRouter Config",
		Filename: "linkrouter.json",
		Filters: []application.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	filePath, err := saveDialog.PromptForSingleSelection()
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
	return config.GetConfigPath()
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

type FileFilter struct {
	DisplayName string `json:"displayName"`
	Pattern     string `json:"pattern"`
}

func (a *App) OpenFileDialog(title string, filters []FileFilter) (string, error) {
	if title == "" {
		title = "Select File"
	}

	dialog := a.app.Dialog.OpenFile().SetTitle(title)
	if len(filters) == 0 {
		dialog.AddFilter("All Files (*.*)", "*.*")
	} else {
		for _, f := range filters {
			dialog.AddFilter(f.DisplayName, f.Pattern)
		}
	}

	filePath, err := dialog.PromptForSingleSelection()
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
	return a.OpenFileDialog("Select LinkRouter Config File", []FileFilter{
		{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		{DisplayName: "All Files (*.*)", Pattern: "*.*"},
	})
}

// For programs
func (a *App) OpenProgramDialog() (string, error) {
	return a.OpenFileDialog("Select Program", []FileFilter{
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

func (a *App) OpenInFallbackBrowser(browserPath string, argsTemplate string, url string) {
	if argsTemplate == "" {
		argsTemplate = "\"{URL}\""
	}
	if strings.TrimSpace(browserPath) == "" {
		dialogs.ShowError("fallback browser is not set\n" +
			"go to settings and set it up")
		return
	}
	err := launcher.LaunchApp(browserPath, argsTemplate, url)
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
			dialogs.ShowError("Unable to compile regex:\n" + err.Error())
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
