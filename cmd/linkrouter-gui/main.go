//go:generate goversioninfo
package main

import (
	"embed"
	"flag"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS
var InteractiveMode = flag.Bool("interactive", false, "Open in interactive mode")
var InteractiveURL = flag.String("url", "", "URL to prefill")
var launchCreateRule = flag.Bool("create-rule", false, "Launch GUI with Edit Rule dialog open")

func main() {
	flag.Parse()
	// *InteractiveMode = true
	// *InteractiveURL = "https://music.yandex.ru/album/123456/track/7890"

	// Create an instance of the app structure
	app := NewApp()

	var title = "Linkrouter"
	var frameless = false
	var width = 1024
	var height = 768
	var minWidth = 512
	var minHeight = 384
	var resizeable = true
	var alpha = 128

	// in interactive mode make window smaller and frameless
	if *InteractiveMode && *InteractiveURL != "" {
		frameless = true
		width = 600
		height = 650
		minWidth = 100
		minHeight = 100
		resizeable = true
		alpha = 0
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:         title,
		Width:         width,
		Height:        height,
		MinWidth:      minWidth,
		MinHeight:     minHeight,
		Frameless:     frameless,
		DisableResize: !resizeable,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: uint8(alpha)},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableWindowIcon:                 true,
			WebviewUserDataPath:               "",
			Theme:                             windows.Dark,
			DisableFramelessWindowDecorations: true,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:          windows.RGB(27, 38, 54),
				DarkModeTitleText:         windows.RGB(203, 213, 225),
				DarkModeBorder:            windows.RGB(27, 38, 54),
				DarkModeTitleBarInactive:  windows.RGB(20, 28, 40),
				DarkModeTitleTextInactive: windows.RGB(148, 163, 184),
				DarkModeBorderInactive:    windows.RGB(20, 28, 40),
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
