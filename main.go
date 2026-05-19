package main

import (
	"embed"
	"os"

	bilibiliapi "github.com/kamiertop/videodown/bilibili/api"
	douyinapi "github.com/kamiertop/videodown/douyin/api"
	mylogger "github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
	"github.com/wailsapp/wails/v2/pkg/options/linux"

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

const appName = "videodown"

func main() {
	log := mylogger.New()
	if err := os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1"); err != nil {
		//	https://github.com/tauri-apps/tauri/issues/10702
		// 设置webgpu相关环境变量，解决Linux平台下WebView2渲染问题
		log.Fatalf("Failed to set environment variable: %v", err)
	}
	settings, err := utils.NewSettings(log)
	if err != nil {
		log.Fatalf("Failed to initialize settings: %v", err)
	}

	bilibili := bilibiliapi.New(log, settings)
	douyin := douyinapi.New(log, settings)
	// Create application with options
	err = wails.Run(&options.App{
		Title:             "",
		Width:             1300,
		Height:            1000,
		MinWidth:          1024,
		MinHeight:         800,
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
		Menu:             nil,
		Logger:           log.WithName("wails"),
		LogLevel:         logger.DEBUG,
		OnStartup:        settings.Startup,
		OnDomReady:       settings.DomReady,
		OnBeforeClose:    settings.BeforeClose,
		OnShutdown:       settings.OnShutdown,
		WindowStartState: options.Normal,
		Bind: []any{
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
				Title:   appName,
				Message: "",
				Icon:    macIcon,
			},
		},
		Linux: &linux.Options{
			Icon:                macIcon,
			WindowIsTranslucent: false,
			Messages:            nil,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
			ProgramName:         appName,
		},
	})

	if err != nil {
		panic(err)
	}
}
