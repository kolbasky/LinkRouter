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

func main() {
	flag.Parse()
	// *InteractiveMode = true
	// *InteractiveURL = "https://music.yandex.ru/album/123456/track/7890"

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "LinkRouter Config Editor",
		Width:     1024,
		Height:    768,
		MinWidth:  800,
		MinHeight: 600,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			DisableWindowIcon:    true,
			WebviewUserDataPath:  "",
			// Theme:                windows.Dark,
			// CustomTheme: &windows.ThemeSettings{
			// 	DarkModeTitleBar:          windows.RGB(27, 38, 54),
			// 	DarkModeTitleBarInactive:  windows.RGB(27, 38, 54),
			// 	DarkModeTitleText:         windows.RGB(203, 213, 225),
			// 	DarkModeTitleTextInactive: windows.RGB(203, 213, 225),
			// 	DarkModeBorder:            windows.RGB(27, 38, 54),
			// },
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
