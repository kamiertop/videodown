package gui

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/ui"
)

// Apply the configured theme using native tokens.
func (s *Shell) themeSettings() core.Settings {
	settings := core.Settings{Mode: core.Mode(s.pages.Settings.Mode), Density: theme.Default, Locale: core.ZhCN}
	if s.pages.Settings.Config.Theme == "cupcake" {
		k := theme.Light()
		k.Background = ui.Hex("#FAF7F5")
		k.Surface = ui.Hex("#FFFFFF")
		k.SurfaceHover = ui.Hex("#F3E6E9")
		k.Border = ui.Hex("#E7DCD8")
		k.ControlBorder = ui.Hex("#B6A8A3")
		k.Accent = ui.Hex("#65C3C8")
		k.AccentHover = ui.Hex("#52A9AF")
		k.AccentPressed = ui.Hex("#438C91")
		k.AccentText = ui.Hex("#286E73")
		k.Selection = ui.Hex("#D9EFEF")
		k.OnAccent = ui.Hex("#123538")
		k.Focus = k.Accent
		settings.Light = &k
	} else if s.pages.Settings.Config.Theme == "caramellatte" {
		k := theme.Light()
		k.Background = ui.Hex("#F6EDE2")
		k.Surface = ui.Hex("#FFFAF5")
		k.Accent = ui.Hex("#A45A32")
		k.AccentHover = ui.Hex("#8E4C29")
		k.AccentPressed = ui.Hex("#753E21")
		k.AccentText = ui.Hex("#854523")
		k.Selection = ui.Hex("#F3DDC4")
		k.Focus = k.Accent
		settings.Light = &k
	}
	return settings
}
