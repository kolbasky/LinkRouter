//go:generate goversioninfo
package main

import (
	"flag"

	"linkrouter/internal/launcher"
	"linkrouter/internal/registry"
)

func main() {
	register := flag.Bool("register", false, "Register ourself in registry")
	unregister := flag.Bool("unregister", false, "Unregister ourself in registry")
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	args := flag.Args()

	if *help {
		launcher.Help()
		return
	}

	if *register {
		registry.RegisterApp()
		return
	}

	if *unregister {
		registry.UnregisterApp()
		return
	}

	if len(args) == 1 && launcher.IsCorrectURL(args[0]) {
		launcher.HandleURL(args[0])
		return
	}

	launcher.HandleNoArgs()
}
