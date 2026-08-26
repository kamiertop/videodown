package main

import (
	"embed"
	"log/slog"

	bilibiliapi "github.com/kamiertop/videodown/bilibili/api"
	douyinapi "github.com/kamiertop/videodown/douyin/api"
	"github.com/kamiertop/videodown/internal/app"
	"github.com/kamiertop/videodown/internal/storage"
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
	controller := app.New(settings)
	bilibili := bilibiliapi.New(log, settings, store, app.EventEmitter(controller))
	douyin := douyinapi.New(log, settings, store, app.EventEmitter(controller))

	wailsApp := application.New(application.Options{
		Name:     appName,
		Logger:   log.Slog(),
		LogLevel: slog.LevelError,
		Services: []application.Service{
			application.NewService(bilibili),
			application.NewService(douyin),
			application.NewService(settings),
			application.NewService(controller),
		},
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
	app.SetupSystemTray(wailsApp, window, icon, controller)

	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if app.BeforeClose(controller) {
			event.Cancel()
		}
	})

	panic(wailsApp.Run())
}
