// main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func main() {
	register := flag.Bool("register", false, "Register as browser app")
	unregister := flag.Bool("unregister", false, "Unregister as browser app")
	daemonMode := flag.Bool("daemon", false, "Run in background (tray mode)")
	testMode := flag.Bool("test", false, "Test mode: execute rule immediately")
	flag.Parse()

	args := flag.Args()

	// Handle CLI flags first
	if *register {
		RegisterAsBrowser()
		return
	}
	if *unregister {
		UnregisterAsBrowser()
		return
	}
	if *daemonMode {
		fmt.Println("🔽 Daemon mode (future: tray icon)")
		return
	}
	if *testMode {
		if len(args) != 1 {
			showError("Usage: --test <url>")
			os.Exit(1)
		}
		handleURL(args[0])
		return
	}

	// Handle URL: exactly one non-flag arg
	if len(args) == 1 && isHTTPURL(args[0]) {
		handleURL(args[0])
		return
	}

	// No valid args → user double-clicked the EXE
	handleDoubleClick()
}

func handleDoubleClick() {
	cfg, err := LoadConfig()
	if err != nil {
		showError("Failed to load config:\n" + err.Error())
		os.Exit(1)
	}

	if cfg.Global.DefaultBrowserPath != "" {
		argsTemplate := cfg.Global.DefaultBrowserArgs
		if !strings.Contains(argsTemplate, "{URL}") {
			argsTemplate += " {URL}"
		}
		launchApp(cfg.Global.DefaultBrowserPath, argsTemplate, "")
	} else {
		showInstructions()
	}
	os.Exit(0)
}

func isHTTPURL(s string) bool {
	return len(s) > 8 && (s[:7] == "http://" || s[:8] == "https://")
}

func handleURL(url string) {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Config error: %v\n", err)
		return
	}

	if rule, matches := cfg.MatchRule(url); rule != nil {
		expandedArgs := expandPlaceholders(rule.Arguments, matches)
		launchApp(rule.Program, expandedArgs, url)
		return
	}

	if cfg.Global.DefaultBrowserPath != "" {
		argsTemplate := cfg.Global.DefaultBrowserArgs
		if argsTemplate == "" {
			argsTemplate = "{URL}"
		}
		launchApp(cfg.Global.DefaultBrowserPath, argsTemplate, url)
	} else {
		showError("No rule matched and no default browser configured.")
	}
}

// in GO %VARS% are not expanded. so convert then to unix-style
func expandPath(path string) string {
	re := regexp.MustCompile(`%([_a-zA-Z][_a-zA-Z0-9\-]*)%`)
	converted := re.ReplaceAllString(path, `$${$1}`)
	ready_path := os.ExpandEnv(converted)
	return ready_path
}

func launchApp(programPath, argsTemplate, url string) error {
	if programPath == "" {
		return fmt.Errorf("program path is empty")
	}
	program := expandPath(programPath)

	// Always quote the program path (handles spaces)
	quotedProgram := strconv.Quote(program)

	var argsLine string
	if argsTemplate == "" {
		argsLine = ""
	} else {
		// Substitute {URL} — and do NOT add extra quoting here
		argsLine = strings.ReplaceAll(argsTemplate, "{URL}", url)
	}

	// Construct FULL command line: "program" args...
	var fullCmdLine string
	if argsLine == "" {
		fullCmdLine = quotedProgram
	} else {
		fullCmdLine = quotedProgram + " " + argsLine
	}

	fmt.Fprintf(os.Stderr, "🚀 LAUNCH CMD: %s\n", fullCmdLine)

	cmd := exec.Command(program) // still needed for Go internals
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

// --- Native Windows Dialogs (Safe & Standard) ---

func showError(msg string) {
	showMessageBox("LinkRouter Error", msg, 0x00000010) // MB_ICONERROR
}

func showInstructions() {
	msg := "To use LinkRouter:\n" +
		"1. Open Command Prompt or PowerShell\n" +
		"2. Run: LinkRouter.exe --register\n" +
		"3. Go to Settings → Apps → Default apps → Web browser\n" +
		"4. Choose 'LinkRouter'\n" +
		"5. Edit config.json (next to this EXE) to add rules\n\n"
	showMessageBox("LinkRouter — Setup Required", msg, 0x00000040) // MB_ICONINFORMATION
}

func showMessageBox(title, text string, icon uint) {
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)

	msgBox.Call(
		0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(icon|0x00001000), // + MB_TOPMOST
	)
}
