package main

import (
	"embed"
	"log/slog"

	bilibiliapi "github.com/kamiertop/videodown/bilibili/api"
	"github.com/kamiertop/videodown/bilibili/download"
	douyinapi "github.com/kamiertop/videodown/douyin/api"
	douyinDown "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/internal/app"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/internal/updater"
	mylogger "github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

const appName = "videodown"

func main() {
	log := mylogger.New()

	store, err := storage.OpenDefault()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	settings, err := utils.NewSettings(log, store)
	if err != nil {
		log.Fatalf("Failed to initialize settings: %v", err)
	}

	controller := app.New(store)

	wailsApp := application.New(application.Options{
		Name:     appName,
		Logger:   log.Slog(slog.LevelError),
		LogLevel: slog.LevelError,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})
	window := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     appName,
		Width:     1200,
		Height:    880,
		MinWidth:  760,
		MinHeight: 600,
		MaxWidth:  2560,
		MaxHeight: 1440,
	})
	app.Configure(controller, wailsApp, window)

	bilibili := bilibiliapi.New(log, store, wailsApp.Event)
	douyin := douyinapi.New(log, store, wailsApp.Event)
	updateService := updater.New(log, store, wailsApp.Event)
	biliDownService := download.NewService(log, store, wailsApp.Event, bilibili.CookieFunc())
	douyinDownService := douyinDown.New(log, store, wailsApp.Event, douyin.PublicHeaders())

	wailsApp.RegisterService(application.NewService(bilibili))
	wailsApp.RegisterService(application.NewService(douyin))
	wailsApp.RegisterService(application.NewService(settings))
	wailsApp.RegisterService(application.NewService(controller))
	wailsApp.RegisterService(application.NewService(updateService))
	wailsApp.RegisterService(application.NewService(biliDownService))
	wailsApp.RegisterService(application.NewService(douyinDownService))

	app.SetupSystemTray(wailsApp, window, icon, controller)

	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if app.BeforeClose(controller) {
			event.Cancel()
		}
	})

	if err := wailsApp.Run(); err != nil {
		log.Fatalf("Application exited with error: %v", err)
	}
}
