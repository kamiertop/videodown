package components

import (
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Button is the shared app button. Change RadiusControl in style/tokens.go
// to adjust every button using this component.
func Button(c *ui.Context, label string, opts input.ButtonOptions) ui.Element {
	return input.
		Button(c, label, opts).
		Radius(style.RadiusControl).
		FontSize(style.Font(c, style.FontButton)).
		FontWeight(style.WeightSemibold)
}
