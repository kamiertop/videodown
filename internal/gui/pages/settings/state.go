// Package settings owns settings form state and views. Window operations
// and persistent storage are provided through callbacks by the shell.
package settings

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/kamiertop/videodown/internal/gui/preferences"
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
	SaveSetting                            func(map[string]string)
	BrowseDirectory                        func()
	BrowseFFmpeg                           func()
	CheckFFmpeg                            func()
	FFmpegChecking, FFmpegChecked          bool
	FFmpegPath, FFmpegVersion, FFmpegError string
}

func (s *State) ApplyTheme() {
	s.Mode = int(core.Light)
	if s.Config.Theme == "dark" {
		s.Mode = int(core.Dark)
	} else if s.Config.Theme == "system" {
		s.Mode = int(core.System)
	}
}
