package youtube

import (
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Use the same six-unit control radius as the other native platform pages.
func roundedInput(c *ui.Context, value *string, label, placeholder string, disabled bool, clearLabel string) (ui.Element, ui.Element) {
	k := c.Theme()
	frame := ui.Row(c).
		MinWidth(0).
		MinHeight(style.InputHeight).
		Padding(0, 12).
		Radius(style.RadiusControl).
		Border(style.BorderWidth, k.Border).
		Background(k.Surface).
		Disabled(disabled)
	var editor ui.Element
	frame.Children(func() {
		editor = components.TextInputEditor(c.Key("editor"), frame, value, disabled, ui.TextInputBase).
			Label(label).
			Placeholder(placeholder).
			Grow(1).
			MinWidth(0).
			FocusRing(false)
		if clearLabel != "" {
			focus := ui.Local(frame, "editor-focus", func() ui.Handle { return ui.Handle{} })
			editor.Bind(focus)
			if *value != "" && !disabled {
				components.IconButton(c.Key("clear"), icons.Must("x"), clearLabel, components.IconButtonOptions{}).
					Size(24, 24).MinWidth(24).MinHeight(24).Padding(0).Shrink(0).
					OnClick(func() {
						*value = ""
						focus.Focus(c)
					})
			}
		}
	})
	if editor.Focused() {
		frame.Border(style.BorderWidth, k.Focus)
	}
	if !disabled && frame.Clicked() {
		editor.Focus()
	}
	return frame, editor
}
