package app

import "github.com/wailsapp/wails/v3/pkg/application"

// SetupSystemTray creates the application menu and attaches it to the native
// system tray. The tray intentionally has no click-to-show-window handler;
// showing the window is an explicit menu action so Linux right-click does not
// also restore the window.
func SetupSystemTray(wailsApp *application.App, window *application.WebviewWindow, icon []byte, controller *Controller) {
	menu := wailsApp.NewMenu()
	menu.Add("显示窗口").OnClick(func(_ *application.Context) {
		window.Show().Focus()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(_ *application.Context) {
		controller.ForceQuit()
		wailsApp.Quit()
	})

	systemTray := wailsApp.
		SystemTray.
		New().
		SetIcon(icon).
		SetMenu(menu)

	systemTray.SetTooltip("videodown - 视频下载工具")
	systemTray.OnRightClick(systemTray.ShowMenu)
}
