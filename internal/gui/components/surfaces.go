// Package components provides shared themed UI elements.
package components

import (
	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"

	"github.com/kamiertop/videodown/internal/gui/style"
)

func Muted(c *ui.Context, text string) {
	ui.Text(c, text).
		FillWidth().
		FontSize(style.Body(c)).
		LineHeight(style.LineHeight).
		TextColor(c.Theme().TextMuted)
}

// PathInput shows a path the user can select and copy but not edit, next to
// a browse button. Onevra's read-only TextBox shows no caret and takes the
// app's control radius.
func PathInput(c *ui.Context, key string, value *string, label, placeholder string) ui.Element {
	frame := ui.Row(c.Key(key)).FillWidth().MinWidth(0)
	frame.Children(func() {
		newc.TextBox(c, value, newc.TextBoxOptions{ReadOnly: true}).
			Label(label).
			Placeholder(placeholder).
			FillWidth().
			MinWidth(0).
			Radius(style.RadiusControl)
	})
	return frame
}
