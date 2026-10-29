package bilibili

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func panel(c *ui.Context, label string, body func()) ui.Element {
	k := core.Tokens(c)
	return ui.Column(c).Label(label).AlignSelf(ui.Stretch).MinWidth(0).Radius(style.RadiusPanel).Clip().
		Border(style.BorderWidth, k.Border.Alpha(.7)).Background(k.Surface).Children(body)
}

func button(c *ui.Context, label string, opts input.ButtonOptions) ui.Element {
	return components.Button(c, label, opts)
}

func unavailable(c *ui.Context, label, icon string) {
	button(c, label, input.ButtonOptions{Variant: input.Outline, Icon: icons.Must(icon), Disabled: true})
}

// Use MujicaUI's editor in a rounded frame; TextInput's built-in frame has a
// fixed small radius. Keep the custom frame local to Bilibili views.
func textInput(c *ui.Context, value *string, label, placeholder string, submit ...func()) ui.Element {
	k := core.Tokens(c)
	frame := ui.Row(c).FillWidth().MinWidth(0).MinHeight(style.InputHeight).Padding(0, 12).Gap(6).
		Radius(style.RadiusControl).Border(style.BorderWidth, k.ControlBorder).Background(k.Surface)
	var editor ui.Element
	frame.Children(func() {
		editor = input.TextInputEditor(c.Key("editor"), frame, value, false, ui.TextInputBase).
			Label(label).Placeholder(placeholder).Grow(1).MinWidth(0).FocusRing(false)
		if len(submit) > 0 {
			editor.OnSubmit(submit[0])
		}
		if *value != "" {
			button(c.Key("clear-input"), "", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("x"), Label: "清除" + label}).
				Size(24, 24).MinHeight(24).MinWidth(24).Padding(0).OnClick(func() { *value = "" })
		}
	})
	if editor.Focused() {
		frame.Border(style.BorderWidth, k.Focus)
	}
	return frame
}

func empty(c *ui.Context, text, icon string) {
	k := core.Tokens(c)
	ui.Column(c.Key("empty")).FillWidth().Grow(1).MinHeight(260).Padding(style.SpaceXLarge).Center().Gap(style.SpaceMedium).Children(func() {
		ui.Icon(c, icons.Must(icon)).Size(32, 32).TextColor(k.TextMuted)
		ui.Text(c, text).FontSize(style.Body(c)).TextColor(k.TextMuted)
	})
}

func caption(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(style.Body(c)).TextColor(core.Tokens(c).TextMuted)
}
