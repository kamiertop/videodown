package home

import (
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Home shows two centered platform links, no copy
// or cards. Buttons keep the icon-only entries accessible to the keyboard.
func Render(c *ui.Context, navigate func(string)) {
	width, _ := c.Size()
	size := float32(256)
	// Keep the platform links within small windows.
	if available := max(float32(32), (width-24-4*32-32)/2); available < size {
		size = available
	}
	ui.Column(c.Key("home")).Label("平台入口").Fill().Center().Children(func() {
		ui.Row(c).Gap(24).AlignItems(ui.Center).Children(func() {
			for _, item := range []struct{ id, label, path string }{
				{"bilibili", "进入BiliBili", "/bilibili/profile"},
				{"douyin", "进入抖音", "/douyin/profile"},
			} {
				ui.ButtonBase(c.Key(item.id)).Label(item.label).Padding(24, 32).Radius(style.RadiusPanel).
					Children(func() { ui.Image(c, assets.Platform(item.id)).Size(size, size) }).
					OnClick(func() { navigate(item.path) })
			}
		})
	})
}
