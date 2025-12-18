package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var logFile *os.File
var enabled bool

func Init(logPath string) error {
	if strings.TrimSpace(logPath) == "" {
		enabled = false
		return nil
	}

	// in GO %VARS% are not expanded. so convert then to unix-style
	re := regexp.MustCompile(`%([_a-zA-Z][_a-zA-Z0-9\-]*)%`)
	converted := re.ReplaceAllString(logPath, `$${$1}`)
	logPath = os.ExpandEnv(converted)

	if !filepath.IsAbs(logPath) {
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		logPath = filepath.Join(exeDir, logPath)
	}

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logFile = f
	enabled = true
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
		enabled = false
	}
}

func Log(message string) {
	if !enabled || logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, message)
	logFile.WriteString(line)
}

func FormatCaptureGroups(groups []string) string {
	if !enabled {
		return ""
	}

	var parts []string
	for i, g := range groups {
		parts = append(parts, fmt.Sprintf("$%d=%q", i, g))
	}
	return strings.Join(parts, ", ")
}
