package youtube

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) accountPage(c *ui.Context, scroll *ui.ScrollState) {
	layout.ScrollArea(c.Key("account-scroll"), scroll, layout.ScrollAreaOptions{}, func() {
		ui.Column(c).Label("YouTube 账号页面").FillWidth().Padding(style.PagePadding).AlignItems(ui.Center).Children(func() {
			k := core.Tokens(c)
			ui.Column(c).
				Label("YouTube 账号信息").
				FillWidth().
				MaxWidth(style.ProfileCardMaxWidth).
				Padding(style.SpaceXLarge).
				Gap(style.SpaceMedium).
				Radius(style.RadiusPanel).
				Border(style.BorderWidth, k.Border).
				Background(k.Surface).
				Children(func() {
					ui.Text(c, "YouTube 账号").FontWeight(style.WeightSemibold)
					status := "未设置 Cookie"
					if s.cookieSaved {
						status = "已设置 Cookie"
					}
					components.Muted(c, status)
					input.TextArea(c.Key("cookie-input"),
						&s.cookie,
						input.TextAreaOptions{
							Label:       "YouTube Cookie",
							Placeholder: "可选：粘贴浏览器请求头中的 Cookie；留空保存可清除",
							MinHeight:   400,
							Disabled:    s.busy || s.api == nil},
					).Element.FillWidth()
					components.Button(c.Key("save-cookie"), "保存 Cookie", input.ButtonOptions{Variant: input.Primary, Loading: s.busy && !s.downloading, Disabled: s.busy || s.api == nil}).OnClick(s.saveCookie)
					if s.authMessage != "" {
						components.Muted(c, s.authMessage)
					}
				})
		})
	}).Grow(1).MinWidth(0).MinHeight(0)
}
