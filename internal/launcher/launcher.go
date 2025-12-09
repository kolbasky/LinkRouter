package launcher

import (
	"fmt"
	"linkrouter/internal/config"
	"linkrouter/internal/dialogs"
	"linkrouter/internal/registry"
	"linkrouter/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

func HandleNoArgs() {
	cfg, err := config.LoadConfig()
	if err != nil {
		dialogs.ShowError("Failed to load config:\n" + err.Error())
		os.Exit(1)
	}

	if cfg.Global.DefaultBrowserPath != "" {
		argsTemplate := cfg.Global.DefaultBrowserArgs
		if !strings.Contains(argsTemplate, "{URL}") {
			argsTemplate += " {URL}"
		}
		launchApp(cfg.Global.DefaultBrowserPath, argsTemplate, "")
	} else {
		dialogs.ShowError("DefaultBrowserPath in config.json is empty!")
	}
	os.Exit(0)
}

func IsCorrectURL(s string) bool {
	return len(strings.TrimSpace(s)) > 1
}

func HandleURL(url string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		dialogs.ShowError("Config error:\n" + err.Error())
		return
	}

	if rule, matches := cfg.MatchRule(url); rule != nil {
		expandedArgs := expandPlaceholders(rule.Arguments, matches)
		err := launchApp(rule.Program, expandedArgs, url)
		if err == nil {
			return
		} else {
			dialogs.ShowError(
				fmt.Sprintf(
					"Failed to launch app %s:\n%s",
					rule.Program,
					err,
				),
			)
		}
	}

	if cfg.Global.DefaultBrowserPath != "" {
		argsTemplate := cfg.Global.DefaultBrowserArgs
		if argsTemplate == "" {
			argsTemplate = "{URL}"
		}
		err := launchApp(cfg.Global.DefaultBrowserPath, argsTemplate, url)
		if err == nil {
			return
		} else {
			dialogs.ShowError(
				fmt.Sprintf(
					"Failed to launch fallback browser:\n%s\nProgram: %s",
					err,
					cfg.Global.DefaultBrowserPath,
				),
			)
		}
	} else {
		dialogs.ShowError("No rule matched and no default browser configured.")
	}
}

// in GO %VARS% are not expanded. so convert then to unix-style
func expandPath(path string) string {
	re := regexp.MustCompile(`%([_a-zA-Z][_a-zA-Z0-9\-]*)%`)
	converted := re.ReplaceAllString(path, `$${$1}`)
	ready_path := os.ExpandEnv(converted)
	return ready_path
}

func containsSupportedProtocol(argsLine string) bool {
	for _, proto := range config.SupportedProtocols {
		cleanProto := registry.ParseProtocol(proto) // your existing func
		if cleanProto == "" {
			continue
		}
		pattern := `(^|[ \t])` + regexp.QuoteMeta(cleanProto) + `:`
		if matched, _ := regexp.MatchString(pattern, argsLine); matched {
			return true
		}
	}
	return false
}

func isExplorer(path string) bool {
	return strings.EqualFold(filepath.Base(path), "explorer.exe")
}

func launchApp(programPath, argsTemplate, url string) error {
	if programPath == "" {
		return fmt.Errorf("program path is empty")
	}
	program := expandPath(programPath)

	if utils.IsLinkRouter(program) {
		return fmt.Errorf("recursion prevented. " +
			"program specified in rule is linkrouter itself. " +
			"skipping rule")
	}

	quotedProgram := strconv.Quote(program)

	var argsLine string
	if argsTemplate == "" {
		argsLine = ""
	} else {
		argsLine = strings.ReplaceAll(argsTemplate, "{URL}", url)
	}
	if isExplorer(program) && containsSupportedProtocol(argsLine) {
		return fmt.Errorf("recursion prevented. link is passed to explorer.exe and LinkRouter is default for this type of links")
	}
	var fullCmdLine string
	if argsLine == "" {
		fullCmdLine = quotedProgram
	} else {
		fullCmdLine = quotedProgram + " " + argsLine
	}

	cmd := exec.Command(program)
	cmd.Path = program
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fullCmdLine,
	}
	return cmd.Start()
}

func expandPlaceholders(template string, matches []string) string {
	result := template
	for i, match := range matches {
		placeholder := "$" + strconv.Itoa(i)
		result = strings.ReplaceAll(result, placeholder, match)
	}
	return result
}
