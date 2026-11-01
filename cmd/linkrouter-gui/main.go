//go:generate goversioninfo
package main

import (
	"embed"
	"flag"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

var InteractiveMode = flag.Bool("interactive", false, "Open in interactive mode")
var InteractiveURL = flag.String("url", "", "URL to prefill")
var launchCreateRule = flag.Bool("create-rule", false, "Launch GUI with Edit Rule dialog open")

func main() {
	flag.Parse()

	var title = "Linkrouter"
	var frameless = false
	var width = 1024
	var height = 768
	var minWidth = 512
	var minHeight = 384
	var resizeable = true
	var backgroundType = application.BackgroundTypeSolid
	var backgroundColour = application.NewRGB(27, 38, 54)

	// in interactive mode make window smaller and frameless
	if *InteractiveMode && *InteractiveURL != "" {
		frameless = true
		width = 500
		height = 100
		minWidth = 100
		minHeight = 100
		resizeable = false
		backgroundType = application.BackgroundTypeTransparent
		backgroundColour = application.NewRGBA(27, 38, 54, 0)
	}

	// Create application with options
	app := application.New(application.Options{
		Name: "linkrouter",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// Create the main window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           title,
		Width:           width,
		Height:          height,
		MinWidth:        minWidth,
		MinHeight:       minHeight,
		Frameless:       frameless,
		DisableResize:   !resizeable,
		BackgroundType:  backgroundType,
		BackgroundColour: backgroundColour,
		DefaultContextMenuDisabled: true,
		Windows: application.WindowsWindow{
			DisableIcon:                       true,
			Theme:                             application.Dark,
			DisableFramelessWindowDecorations: true,
			CustomTheme: application.ThemeSettings{
				DarkModeActive: &application.WindowTheme{
					TitleBarColour:  application.NewRGBPtr(27, 38, 54),
					TitleTextColour: application.NewRGBPtr(203, 213, 225),
					BorderColour:    application.NewRGBPtr(27, 38, 54),
				},
				DarkModeInactive: &application.WindowTheme{
					TitleBarColour:  application.NewRGBPtr(20, 28, 40),
					TitleTextColour: application.NewRGBPtr(148, 163, 184),
					BorderColour:    application.NewRGBPtr(20, 28, 40),
				},
			},
		},
	})

	// Register the App service
	app.RegisterService(application.NewService(NewApp(app)))

	if err := app.Run(); err != nil {
		println("Error:", err.Error())
	}
}
