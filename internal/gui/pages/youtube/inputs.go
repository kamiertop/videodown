package youtube

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Use the same six-unit control radius as the other native platform pages.
func roundedInput(c *ui.Context, value *string, label, placeholder string, disabled bool) (ui.Element, ui.Element) {
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
	})
	if editor.Focused() {
		frame.Border(style.BorderWidth, k.Focus)
	}
	if !disabled && frame.Clicked() {
		editor.Focus()
	}
	return frame, editor
}

// The clear button sits beside the parse action and returns focus to the editor.
func clearLinkButton(c *ui.Context, field, editor ui.Element, value *string, disabled bool) {
	focus := ui.Local(field, "editor-focus", func() ui.Handle { return ui.Handle{} })
	editor.Bind(focus)
	components.Button(c, "清空", components.ButtonOptions{Variant: components.ButtonOutline, Disabled: disabled || strings.TrimSpace(*value) == ""}).
		OnClick(func() {
			*value = ""
			focus.Focus(c)
		})
}
