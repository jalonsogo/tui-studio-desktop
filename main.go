package main

import (
	"context"
	"embed"
	"io/fs"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed native-bridge/bridge.js
var bridgeJS string

// assets is populated by Wails during `wails build` — see wails.json "frontend:dir".
// NOTE: Run `wails build` (not plain `go build`) so the frontend is compiled first
//       and web/dist/ exists for the embed directive below.
//
//go:embed all:web/dist
var assets embed.FS

func main() {
	app := NewApp()

	// Serve from the embedded FS sub-tree rooted at web/dist
	webDist, err := fs.Sub(assets, "web/dist")
	if err != nil {
		_, _ = os.Stderr.WriteString("assets sub error: " + err.Error() + "\n")
		os.Exit(1)
	}

	err = wails.Run(&options.App{
		Title:             "TUIStudio",
		Width:             1400,
		Height:            900,
		MinWidth:          900,
		MinHeight:         600,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 17, G: 17, B: 17, A: 255},
		AssetServer: &assetserver.Options{
			Assets: webDist,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			// Inject the native bridge before any user interaction is possible.
			// Overrides showSaveFilePicker / showOpenFilePicker / showDirectoryPicker
			// to route through Go IPC methods exposed as window.go.main.App.*
			runtime.WindowExecJS(ctx, bridgeJS)
		},
		OnShutdown: func(_ context.Context) {},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "TUIStudio",
				Message: "A visual TUI designer.\n\nhttps://github.com/jalonsogo/tui-studio-desktop",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			DisablePinchZoom:     true,
			IsZoomControlEnabled: false,
			Theme:                windows.Dark,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:   0x111111,
				DarkModeTitleText:  0xffffff,
				DarkModeBorder:     0x222222,
				LightModeTitleBar:  0xffffff,
				LightModeTitleText: 0x000000,
				LightModeBorder:    0xcccccc,
			},
		},
	})
	if err != nil {
		_, _ = os.Stderr.WriteString("Error: " + err.Error() + "\n")
		os.Exit(1)
	}
}
