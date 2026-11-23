package gui

import (
	"github.com/egoist/mygo/ui"
	newctheme "github.com/oneo-me/onevra/source/theme"

	"github.com/kamiertop/videodown/internal/gui/pages/settings"
)

// applyTheme installs the window's Onevra palette, following the configured
// mode and the desktop's high-contrast preference.
func (s *Shell) applyTheme(c *ui.Context) {
	var t *ui.Theme
	switch s.pages.Settings.Config.Theme {
	case "cupcake":
		t = cupcakeTheme()
	case "caramellatte":
		t = caramellatteTheme()
	default:
		dark := s.pages.Settings.Mode == settings.ModeDark
		if s.pages.Settings.Mode == settings.ModeSystem {
			dark = c.Theme().Dark
		}
		t = newctheme.Palette(dark)
	}
	if p := c.Preferences(); p.HighContrast {
		t.Border = t.Border.Mix(t.Text, 0.45)
		t.TextMuted = t.TextMuted.Mix(t.Text, 0.4)
		t.Focus.A = 255
		t.Scrollbar = t.Scrollbar.Alpha(1.6)
	}
	c.SetTheme(t)
}

// cupcakeTheme is the pastel cupcake appearance on Onevra's light palette.
func cupcakeTheme() *ui.Theme {
	t := newctheme.LightPalette()
	t.Background = ui.Hex("#FAF7F5")
	t.Surface = ui.Hex("#FFFFFF")
	t.SurfaceHover = ui.Hex("#F3E6E9")
	t.SurfacePressed = ui.Hex("#D9EFEF")
	t.Border = ui.Hex("#E7DCD8")
	t.Accent = ui.Hex("#65C3C8")
	t.AccentHover = ui.Hex("#52A9AF")
	t.AccentPressed = ui.Hex("#438C91")
	t.AccentText = ui.Hex("#123538")
	t.Selection = ui.Hex("#D9EFEF")
	t.Focus = t.Accent
	return t
}

// caramellatteTheme is the warm caramel appearance on Onevra's light palette.
func caramellatteTheme() *ui.Theme {
	t := newctheme.LightPalette()
	t.Background = ui.Hex("#F6EDE2")
	t.Surface = ui.Hex("#FFFAF5")
	t.Accent = ui.Hex("#A45A32")
	t.AccentHover = ui.Hex("#8E4C29")
	t.AccentPressed = ui.Hex("#753E21")
	t.AccentText = ui.Hex("#FFF8EE")
	t.Selection = ui.Hex("#F3DDC4")
	t.Focus = t.Accent
	return t
}
