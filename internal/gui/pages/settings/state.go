// Package settings owns settings form state and views. Window operations
// and persistent storage are provided through callbacks by the shell.
package settings

import (
	"github.com/kamiertop/videodown/internal/gui/preferences"
)

// 窗口外观模式；与桌面设置对应。
const (
	ModeSystem = iota
	ModeLight
	ModeDark
)

type State struct {
	Mode      int
	Config    preferences.Data
	separator string
}

func NewState() *State { return &State{Config: preferences.Defaults(), separator: " - "} }

// Services supplies live capabilities without coupling pages to native windows.
type Services struct {
	TrayAvailable                          bool
	TrayError                              string
	CloseToTray                            bool
	CloseConfigured                        bool
	SavingClose                            bool
	CloseError                             string
	SaveClosePreference                    func(bool)
	SettingsAvailable                      bool
	SettingsBusy                           bool
	BrowsingSettings                       bool
	SaveSetting                            func(preferences.Changes)
	BrowseDirectory                        func()
	BrowseFFmpeg                           func()
	CheckFFmpeg                            func()
	FFmpegChecking, FFmpegChecked          bool
	FFmpegPath, FFmpegVersion, FFmpegError string
}

func (s *State) ApplyTheme() {
	s.Mode = ModeLight
	if s.Config.Theme == "dark" {
		s.Mode = ModeDark
	} else if s.Config.Theme == "system" {
		s.Mode = ModeSystem
	}
}
