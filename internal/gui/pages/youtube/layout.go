package youtube

import (
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) Layout(c *ui.Context, path string, navigate func(string), content func()) {
	section := strings.TrimPrefix(path, "/youtube/")
	if path == "/youtube" {
		section = "download"
	}
	k := core.Tokens(c)
	ui.Row(c.Key("youtube-layout")).Fill().MinHeight(0).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c.Key("rail")).
			Label("YouTube 导航").Width(style.PlatformNavWidth).Shrink(0).FillHeight().
			Padding(8).Gap(6).Background(k.Surface).BorderColor(k.Border).BorderWidth(0, 1, 0, 0).
			Children(func() {
				ui.Text(c, "YouTube").FillWidth().TextAlign(ui.Center).FontSize(style.Font(c, style.FontCompact)).TextColor(k.TextMuted)
				for _, item := range []struct{ id, label, icon string }{{"download", "下载", "download"}, {"history", "下载历史", "clock"}} {
					s.navButton(c, item.id, item.label, item.icon, section == item.id, navigate)
				}
				ui.Column(c).Grow(1)
				s.navButton(c, "profile", "账号", "user", section == "profile", navigate)
			})
		ui.Column(c.Key("body")).Label("YouTube 功能区域").Grow(1).MinWidth(0).MinHeight(0).Children(content)
	})
}
func (s *State) navButton(c *ui.Context, id, label, icon string, active bool, navigate func(string)) {
	k := core.Tokens(c)
	b := components.Button(c.Key(id), label, input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must(icon), Label: "YouTube" + label}).FillWidth().Height(style.PlatformNavItemHeight).Column().Center().Gap(7).Padding(4).FontSize(style.Font(c, style.FontCompact)).Border(1, k.Border).Background(k.Background)
	if active {
		b.Background(k.Selection).TextColor(k.AccentText).Border(1, k.AccentText)
	}
	b.OnClick(func() { navigate("/youtube/" + id) })
}
