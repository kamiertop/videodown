package components

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// MenuButton shares stable geometry and selected decoration across page menus.
func MenuButton(c *ui.Context, label string, icon *ui.SVG, active bool) ui.Element {
	k := core.Tokens(c)
	variant := input.Ghost
	if active {
		variant = input.Secondary
	}
	button := Button(c, label, input.ButtonOptions{Variant: variant, Icon: icon}).
		Label(label).
		FontSize(style.Body(c)).
		Radius(style.RadiusMenu).
		Border(style.BorderWidth, ui.Transparent).Shrink(0)
	if active {
		button.Border(style.BorderWidth, k.AccentText).Background(k.Selection).TextColor(k.AccentText)
	}

	return button
}
