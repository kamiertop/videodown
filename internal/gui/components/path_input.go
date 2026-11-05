package components

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// PathInput shows a path the user can select and copy but not edit, next to
// a browse button. It frames mygo's editor itself instead of MujicaUI's
// TextInput, whose read-only mode keeps the caret and whose radius is
// hardcoded: the native editor's ReadOnly shows no caret, and the frame
// takes the app's control radius.
func PathInput(c *ui.Context, key string, value *string, label, placeholder string) ui.Element {
	k := core.Tokens(c)
	d := core.Density(c)
	frame := ui.Row(c.Key(key)).MinHeight(d.ControlHeight()).Padding(0, d.Unit()*3).
		Gap(d.Unit() * 2).Radius(style.RadiusControl).Cursor(ui.CursorText).
		Background(k.Background).Border(theme.BorderWidth, k.Border)
	var in ui.Element
	frame.Children(func() {
		in = ui.TextInputBase(c, value).Grow(1).MinWidth(d.Unit() * 10).
			Placeholder(placeholder).Label(label).ReadOnly(true).FocusRing(false)
	})
	if frame.Clicked() {
		in.Focus()
	}
	frame.DrawOver(func(p *ui.Painter, r ui.Rect) {
		if in.FocusVisible() {
			p.FocusRing(r, [4]float32{style.RadiusControl, style.RadiusControl, style.RadiusControl, style.RadiusControl})
		}
	})
	return frame
}
