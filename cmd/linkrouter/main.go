//go:generate goversioninfo
package main

import (
	"flag"

	"linkrouter/internal/dialogs"
	"linkrouter/internal/globals"
	"linkrouter/internal/launcher"
	"linkrouter/internal/logger"
	"linkrouter/internal/registry"
)

func main() {
	register := flag.Bool("register", false, "Register ourself in registry")
	quiet := flag.Bool("quiet", false, "Do not show popups when registering")
	unregister := flag.Bool("unregister", false, "Unregister ourself in registry")
	help := flag.Bool("help", false, "Show help message")
	version := flag.Bool("version", false, "Show version")
	edit := flag.Bool("edit", false, "Edit config")
	showDefaultApps := flag.Bool("default-apps", false, "Show default apps dialog")
	flag.Parse()

	args := flag.Args()

	globals.QuietMode = *quiet

	if *help {
		launcher.Help()
		return
	}

	if *version {
		dialogs.ShowMessageBox("LinkRouter", "version 2.5.4", 0x00000040)
		return
	}

	if *edit {
		launcher.EditConfig()
		return
	}

	if *register {
		registry.RegisterApp()
		defer logger.Close()
		return
	}

	if *showDefaultApps {
		registry.ShowWinDefaultApps()
		defer logger.Close()
		return
	}

	if *unregister {
		registry.UnregisterApp()
		defer logger.Close()
		return
	}

	if len(args) == 1 && launcher.IsCorrectURL(args[0]) {
		launcher.HandleURL(args[0])
		defer logger.Close()
		return
	}

	launcher.HandleNoArgs()
	defer logger.Close()
}
