package bilibili

import (
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/icons"

	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func panel(c *ui.Context, label string, body func()) ui.Element {
	k := c.Theme()
	return ui.Column(c).Label(label).AlignSelf(ui.Stretch).MinWidth(0).Radius(style.RadiusPanel).Clip().
		Border(style.BorderWidth, k.Border.Alpha(.7)).Background(k.Surface).Children(body)
}

func button(c *ui.Context, label string, opts components.ButtonOptions) ui.Element {
	return components.Button(c, label, opts)
}

func unavailable(c *ui.Context, label, icon string) {
	button(c, label, components.ButtonOptions{Variant: components.ButtonOutline, Icon: icons.Must(icon), Disabled: true})
}

// Use the native editor in a rounded frame with the app's control radius.
// Keep the custom frame local to Bilibili views.
func textInput(c *ui.Context, value *string, label, placeholder string, submit func(), showClear bool) ui.Element {
	k := c.Theme()
	frame := ui.Row(c).FillWidth().MinWidth(0).MinHeight(style.InputHeight).Padding(0, 12).Gap(6).
		Radius(style.RadiusControl).Border(style.BorderWidth, k.Border).Background(k.Surface)
	var editor ui.Element
	frame.Children(func() {
		editor = components.TextInputEditor(c.Key("editor"), frame, value, false, ui.TextInputBase).
			Label(label).Placeholder(placeholder).Grow(1).MinWidth(0).FocusRing(false)
		if submit != nil {
			editor.OnSubmit(submit)
		}
		if showClear && *value != "" {
			button(c.Key("clear-input"), "", components.ButtonOptions{Variant: components.ButtonGhost, Icon: icons.Must("x"), Label: "清除" + label}).
				Size(24, 24).MinHeight(24).MinWidth(24).Padding(0).OnClick(func() { *value = "" })
		}
	})
	if editor.Focused() {
		frame.Border(style.BorderWidth, k.Focus)
	}

	return frame
}

func empty(c *ui.Context, text, icon string) {
	k := c.Theme()
	ui.Column(c.Key("empty")).FillWidth().Grow(1).MinHeight(260).Padding(style.SpaceXLarge).Center().Gap(style.SpaceMedium).Children(func() {
		ui.Icon(c, icons.Must(icon)).Size(32, 32).TextColor(k.TextMuted)
		ui.Text(c, text).FontSize(style.Body(c)).TextColor(k.TextMuted)
	})
}

func caption(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(style.Body(c)).TextColor(c.Theme().TextMuted)
}
