package gui

import (
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/pages/settings"
	"github.com/kamiertop/videodown/internal/gui/preferences"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/utils"
)

func (s *Shell) settingsServices() settings.Services {
	return settings.Services{
		TrayAvailable: s.tray != nil, TrayError: s.trayError, CloseToTray: s.closing.CloseToTray, CloseConfigured: s.closing.Configured, SavingClose: s.closing.Saving, CloseError: s.closing.Error,
		SaveClosePreference: func(value bool) { s.saveClosePreference(value, nil) },
		SettingsAvailable:   s.settingsStore != nil, SettingsBusy: s.settingsBusy, BrowsingSettings: s.browsingSettings,
		SaveSetting: s.saveSettings, BrowseDirectory: s.browseDirectory, BrowseFFmpeg: s.browseFFmpeg, CheckFFmpeg: s.checkFFmpeg,
		FFmpegChecking: s.ffmpegChecking, FFmpegChecked: s.ffmpegChecked, FFmpegPath: s.ffmpegPath, FFmpegVersion: s.ffmpegVersion, FFmpegError: s.ffmpegError,
	}
}
func (s *Shell) saveSettings(values preferences.Changes) {
	if s.settingsBusy {
		s.pendingSettings = s.pendingSettings.Merge(values)
		return
	}
	if s.settingsStore == nil {
		s.notifySettings("配置存储尚未初始化", true)
		return
	}
	values = (preferences.Changes{}).Merge(values)
	before := s.savedSettings
	s.settingsBusy = true
	s.runJob(func() error { return preferences.Save(s.settingsStore, values) }, func(err error) {
		s.settingsBusy = false
		if err != nil {
			_ = s.pages.Settings.Config.Apply(values.Restore(before, s.pendingSettings))
			s.pages.Settings.ApplyTheme()
			s.notifySettings("保存失败："+err.Error(), true)
		} else {
			_ = s.savedSettings.Apply(values)
			s.notifySettings("保存成功", false)
		}
		s.flushPendingSettings()
	})
}
func (s *Shell) browseDirectory() { s.browseSetting(true) }
func (s *Shell) browseFFmpeg()    { s.browseSetting(false) }
func (s *Shell) browseSetting(directory bool) {
	if s.win == nil || s.settingsStore == nil {
		return
	}
	if s.settingsBusy {
		s.pendingBrowse = &directory
		return
	}
	s.settingsBusy = true
	s.browsingSettings = true
	win := s.win
	title := "选择下载目录"
	changes := func(value string) preferences.Changes {
		return preferences.Changes{Directory: preferences.Value(value)}
	}
	if !directory {
		title = "选择 FFmpeg 可执行文件"
		changes = func(value string) preferences.Changes { return preferences.Changes{FFmpeg: preferences.Value(value)} }
	}
	var chosen, path, version string
	s.runJob(func() error {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: win, Title: title, Directory: directory, CreateDirectories: directory})
		if err != nil || len(paths) == 0 {
			return err
		}
		chosen = paths[0]
		if !directory {
			path, version, err = utils.CheckFFmpeg(chosen)
			if err != nil {
				return err
			}
		}
		return preferences.Save(s.settingsStore, changes(chosen))
	}, func(err error) {
		s.settingsBusy = false
		s.browsingSettings = false
		defer s.flushPendingSettings()
		if err != nil {
			s.notifySettings("保存失败："+err.Error(), true)
			return
		}
		if chosen == "" {
			return
		}
		values := changes(chosen)
		_ = s.pages.Settings.Config.Apply(values)
		_ = s.savedSettings.Apply(values)
		s.notifySettings("保存成功", false)
		if !directory {
			s.ffmpegPath, s.ffmpegVersion, s.ffmpegError, s.ffmpegChecked = path, version, "", true
		}
	})
}
func (s *Shell) checkFFmpeg() {
	if s.ffmpegChecking || s.win == nil {
		return
	}
	s.ffmpegChecking = true
	custom := s.pages.Settings.Config.FFmpeg
	var path, version string
	s.runJob(func() error { var err error; path, version, err = utils.CheckFFmpeg(custom); return err }, func(err error) {
		s.ffmpegChecking = false
		s.ffmpegChecked = true
		s.ffmpegPath, s.ffmpegVersion = path, version
		s.ffmpegError = ""
		if err != nil {
			s.ffmpegError = err.Error()
		}
	})
}

// Notifications live outside the form so saves never move its controls.
func (s *Shell) notifySettings(message string, failed bool) {
	duration := 3 * time.Second
	if failed {
		duration = 6 * time.Second
	}
	s.settingsToast = &settingsNotification{message: message, failed: failed, duration: duration}
}

type settingsNotification struct {
	message  string
	failed   bool
	duration time.Duration
	expires  time.Time
}

func (s *Shell) settingsNotification(c *ui.Context) {
	toast := s.settingsToast
	if toast == nil {
		return
	}
	if toast.expires.IsZero() {
		toast.expires = c.Now().Add(toast.duration)
	}
	remaining := toast.expires.Sub(c.Now())
	if remaining <= 0 {
		s.settingsToast = nil
		return
	}
	c.After(remaining)
	k := core.Tokens(c)
	color := k.Success
	if toast.failed {
		color = k.Danger
	}
	width, _ := c.Size()
	ui.Overlay(c, func() {
		ui.Column(c.Key("settings-notification")).Absolute().Right(24).Bottom(24).
			Width(min(360, width-48)).Padding(16).Radius(style.RadiusMenu).Background(k.Surface).
			Border(style.BorderWidth, color).PassThrough().Role(ui.RoleStatus).Label(toast.message).Children(func() {
			ui.Text(c, toast.message).FillWidth().TextColor(color)
		})
	})
}

func (s *Shell) flushPendingSettings() {
	if !s.pendingSettings.Empty() {
		pending := s.pendingSettings
		s.pendingSettings = preferences.Changes{}
		s.saveSettings(pending)
		return
	}
	if s.pendingBrowse != nil {
		directory := *s.pendingBrowse
		s.pendingBrowse = nil
		s.browseSetting(directory)
	}
}
