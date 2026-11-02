package gui

import (
	"log"

	"github.com/kamiertop/videodown/internal/gui/assets"

	"github.com/egoist/mygo"
)

// setupTray runs once after the app is ready. A menu makes restoring explicit,
// so opening a Linux tray context menu never also restores the window.
func (s *Shell) setupTray() {
	if s.tray != nil {
		return
	}
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Label: "显示窗口", Click: func(*mygo.MenuItem, *mygo.Window) { s.showWindow() }},
		{Label: "B 站下载", Click: func(*mygo.MenuItem, *mygo.Window) { s.openPage("/bilibili/download") }},
		{Label: "抖音下载", Click: func(*mygo.MenuItem, *mygo.Window) { s.openPage("/douyin/download") }},
		{Label: "设置", Click: func(*mygo.MenuItem, *mygo.Window) { s.openPage("/settings") }},
		mygo.Separator(),
		{Label: "退出", Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})
	tray, err := mygo.NewTray(mygo.TrayOptions{
		ID: "videodown-main", Icon: assets.AppIconPNG,
		ToolTip: "videodown - 视频下载工具", Menu: menu,
	})
	if err != nil {
		s.trayError = "系统托盘不可用，请选择退出程序或取消关闭。"
		log.Printf("create system tray: %v", err)
		return
	}
	s.tray = tray
	tray.OnRightClick(tray.PopUpMenu)
}

func (s *Shell) showWindow() {
	if s.win == nil || s.win.IsDestroyed() {
		return
	}
	s.win.Show()
	if s.win.IsMinimized() {
		s.win.Restore()
	}
	s.win.Focus()
}

func (s *Shell) openPage(path string) {
	if s.win == nil || s.win.IsDestroyed() {
		return
	}
	s.win.Update(func() { s.router.Push(path) })
	s.showWindow()
}
