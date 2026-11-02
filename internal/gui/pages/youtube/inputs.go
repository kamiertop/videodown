package youtube

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Use the same six-unit control radius as the other native platform pages.
func roundedInput(c *ui.Context, value *string, label, placeholder string, disabled bool) input.InputResult {
	k := core.Tokens(c)
	frame := ui.Row(c).
		MinWidth(0).
		MinHeight(style.InputHeight).
		Padding(0, 12).
		Radius(style.RadiusControl).
		Border(style.BorderWidth, k.ControlBorder).
		Background(k.Surface).
		Disabled(disabled)
	var editor ui.Element
	frame.Children(func() {
		editor = input.TextInputEditor(c.Key("editor"), frame, value, disabled, ui.TextInputBase).
			Label(label).
			Placeholder(placeholder).
			Grow(1).
			MinWidth(0).
			FocusRing(false)
	})
	if editor.Focused() {
		frame.Border(style.BorderWidth, k.Focus)
	}
	if !disabled && frame.Clicked() {
		editor.Focus()
	}
	return input.NewInputResult(frame, editor, editor.Changed() && !disabled, editor.Submitted() && !disabled)
}
