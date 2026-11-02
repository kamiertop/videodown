package settings

import (
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/internal/updater"
)

func (s *State) aboutSettings(c *ui.Context) {
	// One page root keeps the settings form's Gap from adding space between
	// these sections. Card spacing and the footer margin are controlled here.
	ui.Column(c).FillWidth().Children(func() {
		ui.Column(c).
			FillWidth().
			Gap(20).
			Children(func() {
				ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(4).
					Children(
						func() {
							ui.Image(c, assets.AppIcon).Size(80, 80)
							ui.Text(c, "VideoDown").FontSize(style.Font(c, style.FontHero)).FontWeight(style.WeightBold)
							ui.Text(c, updater.CurrentVersion())
						})
				settingCard(c, "功能特性", "play", func() {
					for _, line := range []string{"哔哩哔哩 — 视频、番剧、收藏夹、UP 主作品", "抖音 — 视频、合集、喜欢列表、用户作品", "多线程并发下载，下载后自动随机休眠", "多主题切换、历史管理、自动分目录整理"} {
						components.Muted(c, "▸ "+line)
					}
				})
				settingCard(c, "开源说明", "info", func() {
					components.Muted(c, "本软件开源免费，源代码托管于 GitHub，欢迎 Star")
					components.Muted(c, "禁止用于商业用途，禁止转售、倒卖、二次打包收费")
				})
				settingCard(c, "免责声明", "info", func() {
					components.Muted(c, "本软件仅提供视频下载技术实现，不收集任何用户信息。用户下载的视频版权归属原平台及创作者所有。请遵守相关平台的使用条款，在合法授权范围内使用。因使用本软件产生的任何法律纠纷与开发者无关。")
				})
			})
		ui.Column(c).
			FillWidth().
			AlignItems(ui.Center).
			Margin(4, 0, 0, 0).
			Children(func() {
				services := c.Services()
				components.Button(c.Key("source"), "GitHub", input.ButtonOptions{Variant: input.Ghost, Icon: assets.GitHub}).
					OnClick(func() { services.OpenURL(assets.RepositoryURL) })
			})
	})
}
