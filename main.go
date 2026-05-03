package main

import (
	"context"
	"embed"

	bilibiliapi "github.com/kamiertop/videodown/bilibili/api"
	douyinapi "github.com/kamiertop/videodown/douyin/api"
	mylogger "github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var macIcon []byte

func main() {
	log := mylogger.New()
	settings, err := utils.NewSettings(log)
	if err != nil {
		log.Fatalf("Failed to initialize settings: %v", err)
	}
	app := NewApp(settings)
	bilibili := bilibiliapi.New(log, settings, func() context.Context {
		return app.ctx
	})
	douyin := douyinapi.New(log, settings, func() context.Context {
		return app.ctx
	})
	// Create application with options
	err = wails.Run(&options.App{
		Title:             "videodown",
		Width:             1300,
		Height:            800,
		MinWidth:          1024,
		MinHeight:         768,
		MaxWidth:          2560,
		MaxHeight:         1440,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false, // 是否无边框，无边框时需要额外处理窗口拖动等功能
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:          nil,
		Logger:        log.WithName("wails"),
		LogLevel:      logger.DEBUG,
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnBeforeClose: app.beforeClose,
		OnShutdown: func(ctx context.Context) {
			if err := settings.CloseDB(); err != nil {
				log.Errorf("Failed to close settings DB: %v", err)
				return
			}
			log.Info("Close BadgerDB and shutdown application")
		},
		WindowStartState: options.Normal,
		Bind: []any{
			app,
			bilibili,
			douyin,
			settings,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
			ZoomFactor:          1.0,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "videodown",
				Message: "",
				Icon:    macIcon,
			},
		},
	})

	if err != nil {
		panic(err)
	}
}
