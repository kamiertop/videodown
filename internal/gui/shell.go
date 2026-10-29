// Package gui is videodown's native UI: a MyGo window drawn with MujicaUI
// components for the Go-native application.
package gui

import (
	"context"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	dydownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/internal/gui/pages/douyin"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/internal/updater"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/gui/closing"
	"github.com/kamiertop/videodown/internal/gui/pages"
	"github.com/kamiertop/videodown/internal/gui/pages/bilibili"
	"github.com/kamiertop/videodown/internal/gui/preferences"
)

// appName names the app across its window and title bar.
const appName = assets.AppName

// windowRadius is the frame's corner radius in DIPs, as GNOME's own
// windows; the frame squares off when the window maximizes.
const windowRadius = style.RadiusWindow

// Shell is the state the window shows.
type Shell struct {
	biliJobs                  bilibiliJobs
	router                    *ui.Router
	win                       *mygo.Window
	scrolls                   map[string]*ui.ScrollState
	tray                      *mygo.Tray
	trayError                 string
	quitting                  bool
	pages                     *pages.State
	closing                   *closing.Controller
	activeDownloads           func() int
	confirmedQuit             bool
	runJob                    func(func() error, func(error))
	settingsStore             preferences.Store
	pendingSettings           map[string]string
	savedSettings             preferences.Data
	settingsBusy              bool
	settingsToast             *settingsNotification
	browsingSettings          bool
	pendingBrowse             *bool
	ffmpegChecking            bool
	ffmpegChecked             bool
	ffmpegPath                string
	ffmpegVersion             string
	ffmpegError               string
	updateChecker             UpdateChecker
	updateStarted, updateOpen bool
	updateResult              updater.Result
	updateNotes               ui.ScrollState
	runUpdateJob              func(func() error, func(error))
}

// Options supplies shared persistent preferences and the download queue probe.
// A nil probe reports zero active tasks.
type Options struct {
	Bilibili         bilibili.API
	BilibiliDownload bilibili.DownloadBackend
	Douyin           douyin.API
	DouyinDownload   douyin.DownloadBackend
	Preferences      closing.Preferences
	Settings         preferences.Store
	ActiveDownloads  func() int
	UpdateChecker    UpdateChecker
}

func NewShell() *Shell { return NewShellWithOptions(Options{}) }
func NewShellWithOptions(opts Options) *Shell {
	s := &Shell{router: ui.NewRouter("/"), scrolls: make(map[string]*ui.ScrollState), pages: pages.NewState(), closing: closing.New(opts.Preferences), activeDownloads: opts.ActiveDownloads}
	data, err := preferences.Load(opts.Settings)
	s.pages.Settings.Config = data
	s.savedSettings = data
	s.settingsStore = opts.Settings
	s.pages.Settings.ApplyTheme()
	if err != nil {
		s.notifySettings("读取配置失败："+err.Error(), true)
	}
	s.runJob = func(work func() error, done func(error)) {
		win := s.win
		go func() { err := work(); win.Update(func() { done(err) }) }()
	}
	s.biliJobs.ctx, s.biliJobs.cancel = context.WithCancel(context.Background())
	s.biliJobs.images = make(chan struct{}, 4)
	s.updateChecker = opts.UpdateChecker
	s.runUpdateJob = s.bilibiliRuntime().Background
	s.pages.Bilibili.Configure(opts.Bilibili, s.bilibiliRuntime())
	s.pages.Bilibili.ConfigureDownloads(opts.BilibiliDownload)
	if source, ok := opts.BilibiliDownload.(interface {
		SetProgressListener(func(bilidownload.Progress))
	}); ok {
		source.SetProgressListener(func(progress bilidownload.Progress) {
			win := s.win
			if win == nil || win.IsDestroyed() || s.biliJobs.ctx.Err() != nil {
				return
			}
			win.Update(func() { s.pages.Bilibili.UpdateDownloadProgress(progress) })
		})
	}
	s.pages.Douyin.Configure(opts.Douyin, opts.DouyinDownload, s.bilibiliRuntime())
	if source, ok := opts.DouyinDownload.(interface {
		SetProgressListener(func(dydownload.Progress))
	}); ok {
		source.SetProgressListener(func(p dydownload.Progress) {
			win := s.win
			if win != nil && !win.IsDestroyed() && s.biliJobs.ctx.Err() == nil {
				win.Update(func() { s.pages.Douyin.UpdateDownloadProgress(p) })
			}
		})
	}

	return s
}

// OpenWindow creates the application's window.
func (s *Shell) OpenWindow() {
	if s.win != nil && !s.win.IsDestroyed() {
		s.showWindow()
		return
	}
	s.win = mygo.NewWindow(mygo.WindowOptions{
		Title:             appName,
		Width:             1200,
		Height:            880,
		MinWidth:          760,
		MinHeight:         600,
		TitleBarStyle:     mygo.TitleBarHidden,
		HideWindowButtons: true,
		Transparent:       true,
		Content:           ui.View(s.view),
	})
	s.pages.Bilibili.Start()
	s.pages.Douyin.Start()
	s.startUpdateCheck()
	s.win.OnClosed(s.stopBilibili)
	// The taskbar and the window list show the app's icon.
	_ = s.win.SetIcon(assets.AppIconPNG)
	s.setupTray()
	s.win.OnClose(s.beforeClose)
	mygo.App.OnBeforeQuit(s.beforeQuit)
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) {
		s.stopBilibili()
		if s.tray != nil {
			s.tray.Destroy()
			s.tray = nil
		}
	})
	mygo.App.OnActivate(func(bool) { s.showWindow() })
}

// view builds the interface from the state, on the main thread, whenever the
// window needs a frame.
func (s *Shell) view(c *ui.Context) {
	core.Use(c, s.themeSettings())
	style.Apply(c)
	registerShortcut(c)
	// The transparent window shows the desktop around the rounded frame,
	// which squares off when the window maximizes.
	radius := float32(windowRadius)
	if s.win != nil && s.win.IsMaximized() {
		radius = 0
	}
	k := core.Tokens(c)
	c.Root().Background(k.Background).Radius(radius).Clip().Border(style.BorderWidth, k.Border.Alpha(0.5))
	ui.Column(c).Fill().
		Children(func() {
			s.titleBar(c)
			s.content(c)
		})
	s.updateDialog(c)
	s.closeDialogs(c)
	s.settingsNotification(c)
}

// scroll returns the content scroll state of one page, keeping its offset.
func (s *Shell) scroll(id string) *ui.ScrollState {
	if _, ok := s.scrolls[id]; !ok {
		s.scrolls[id] = &ui.ScrollState{}
	}
	return s.scrolls[id]
}

// registerShortcut registers global keyboard shortcuts for the window.
func registerShortcut(c *ui.Context) {
	c.OnShortcut(ui.Cmd, ui.KeyP, func() {
		c.Router().Push("/settings")
	})
}
