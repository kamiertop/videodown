// Package components provides shared themed UI elements.
package components

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func Muted(c *ui.Context, text string) {
	ui.Text(c, text).FillWidth().FontSize(style.Body(c)).LineHeight(style.LineHeight).TextColor(core.Tokens(c).TextMuted)
}
