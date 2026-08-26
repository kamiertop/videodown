// Package app contains Wails runtime behaviour that does not belong to
// persistent application settings.
package app

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/kamiertop/videodown/utils"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Controller owns application-window interactions exposed to the frontend.
type Controller struct {
	settings  *utils.Settings
	app       *application.App
	window    *application.WebviewWindow
	forceQuit atomic.Bool
}

func New(settings *utils.Settings) *Controller {
	return &Controller{settings: settings}
}

// Configure supplies the Wails objects after their construction. It is a
// package function so it is not exported as a frontend service binding.
func Configure(controller *Controller, wailsApp *application.App, window *application.WebviewWindow) {
	controller.app = wailsApp
	controller.window = window
}

// EventEmitter returns a runtime adapter for backend services.
func EventEmitter(controller *Controller) utils.EventEmitter {
	return eventEmitter{controller: controller}
}

type eventEmitter struct{ controller *Controller }

func (e eventEmitter) EmitEvent(name string, data any) bool {
	return emitEvent(e.controller, name, data)
}

// BeforeClose reports whether a main-window close event should be cancelled.
func BeforeClose(controller *Controller) bool {
	if controller.forceQuit.Swap(false) {
		return false
	}

	closeToTray, err := controller.settings.GetCloseToTray()
	if err != nil {
		emitEvent(controller, "before-close-prompt", nil)
		return true
	}
	if closeToTray {
		controller.HideWindow()
		return true
	}
	return false
}

func emitEvent(controller *Controller, name string, data any) bool {
	window := controller.window
	if window == nil {
		return false
	}
	return window.EmitEvent(name, data)
}

// ForceQuit allows the next close event to exit the application.
func (c *Controller) ForceQuit() {
	c.forceQuit.Store(true)
}

// HideWindow hides the main window.
func (c *Controller) HideWindow() {
	window := c.window
	if window != nil {
		window.Hide()
	}
}

// SetStorage opens a directory chooser and persists the selected path.
func (c *Controller) SetStorage() (string, error) {
	a, window, err := c.wailsObjects()
	if err != nil {
		return "", err
	}
	dir, err := a.Dialog.OpenFile().CanChooseFiles(false).CanChooseDirectories(true).SetTitle("选择下载目录").AttachToWindow(window).PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("打开目录选择对话框失败: %w", err)
	}
	if dir == "" {
		return "", nil
	}
	if err := c.settings.SetStoragePath(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// SelectFFmpegPath opens an executable chooser and persists a validated path.
func (c *Controller) SelectFFmpegPath() (string, error) {
	wailsApp, window, err := c.wailsObjects()
	if err != nil {
		return "", err
	}
	path, err := wailsApp.Dialog.OpenFile().SetTitle("选择 FFmpeg 可执行文件").AddFilter("可执行文件", "ffmpeg;ffmpeg.exe;*").AttachToWindow(window).PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("打开文件选择对话框失败: %w", err)
	}
	if path == "" {
		return "", nil
	}
	if err := c.settings.SetFFmpegPath(path); err != nil {
		return "", err
	}
	return path, nil
}

func (c *Controller) wailsObjects() (*application.App, *application.WebviewWindow, error) {
	if c.app == nil || c.window == nil {
		return nil, nil, errors.New("应用尚未初始化")
	}
	return c.app, c.window, nil
}
