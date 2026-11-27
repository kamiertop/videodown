package bilibili

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) profile(c *ui.Context) {
	if s.auth.checking {
		empty(c, "正在读取账号信息", "loader-circle")
		return
	}
	ui.Column(c).FillWidth().Grow(1).AlignItems(ui.Center).Justify(ui.Center).Padding(16, 0).
		Children(func() {
			maximum := style.ProfileLoginMaxWidth
			if s.auth.loggedIn {
				maximum = style.ProfileCardMaxWidth
			}
			ui.Column(c).FillWidth().MaxWidth(maximum).
				Children(func() {
					if s.auth.loggedIn {
						s.loggedInProfile(c)
					} else {
						s.loginProfile(c)
					}
				})
		})
}

func (s *State) loggedInProfile(c *ui.Context) {
	panel(c, "账号资料", func() {
		ui.Column(c).FillWidth().Padding(style.SpaceXLarge).Gap(style.SpaceXLarge).Children(func() {
			ui.Row(c).FillWidth().Gap(style.SpaceLarge).AlignItems(ui.Center).Children(func() {
				s.image(c, s.auth.account.Face, style.ProfileAvatarSize, style.ProfileAvatarSize, true)
				ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
					caption(c, "当前登录账号")
					name := s.auth.account.Name
					if name == "" {
						name = "哔哩哔哩账号"
					}
					ui.Text(c, name).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontTitle)).FontWeight(style.WeightBold)
					caption(c, fmt.Sprintf("Lv.%d", s.auth.account.Level))
				})
			})
			vip := "普通用户"
			if s.auth.account.Vip.Status == 1 && s.auth.account.Vip.Type > 0 {
				vip = s.auth.account.Vip.Label.Text
				if vip == "" {
					vip = "大会员"
				}
			}
			ui.Grid(c).Columns(3).FillWidth().Gap(style.SpaceMedium).Children(func() {
				profileFact(c.Key("signature"), "签名", s.auth.account.Sign)
				profileFact(c.Key("membership"), "会员状态", vip)
				profileFact(c.Key("uid"), "UID", fmt.Sprint(s.auth.account.Mid))
			})
			s.errorRow(c, s.auth.error, s.checkAccount)
			ui.Row(c).FillWidth().Gap(style.SpaceSmall).Justify(ui.End).Children(func() {
				button(c.Key("refresh-account"), "刷新资料", components.ButtonOptions{Variant: components.ButtonOutline, Loading: s.auth.checking}).OnClick(s.checkAccount)
				button(c.Key("logout"), "退出登录", components.ButtonOptions{Variant: components.ButtonDanger, Loading: s.auth.loggingOut}).OnClick(s.logout)
			})
		})
	})
}

func profileFact(c *ui.Context, label, value string) {
	if value == "" {
		value = "暂无"
	}
	k := c.Theme()
	ui.Column(c).MinWidth(0).MinHeight(style.ProfileFactHeight).Padding(style.SpaceMedium).Gap(style.SpaceSmall).
		Radius(style.RadiusPanel).Background(k.Background).Children(func() {
		ui.Text(c, label).FillWidth().TextColor(k.TextMuted)
		ui.Text(c, value).FillWidth().MaxLines(3).FontWeight(style.WeightSemibold)
	})
}

func (s *State) loginProfile(c *ui.Context) {
	panel(c, "账号登录", func() {
		ui.Column(c).FillWidth().Padding(style.SpaceXLarge).Gap(20).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "扫码登录").FontWeight(style.WeightSemibold)
			ui.Column(c).Size(style.QRCodeSize, style.QRCodeSize).Center().Gap(style.SpaceMedium).Radius(style.RadiusPanel).Border(style.BorderWidth, c.Theme().Border).Background(ui.RGB(255, 255, 255)).Children(func() {
				if s.auth.qrImage != nil && !s.auth.qrExpired {
					ui.Image(c, s.auth.qrImage).Size(style.QRCodeSize, style.QRCodeSize)
				} else {
					ui.Icon(c, icons.Must("maximize")).Size(48, 48).TextColor(c.Theme().TextMuted)
					text := "二维码待获取"
					if s.auth.qrExpired {
						text = "二维码已过期"
					}
					caption(c, text)
				}
			})
			if s.auth.qrStatus != "" {
				caption(c, s.auth.qrStatus)
			}
			label := "获取登录二维码"
			if s.auth.qr.Loaded {
				label = "刷新二维码"
			}
			button(c.Key("qr"), label, components.ButtonOptions{Variant: components.ButtonOutline, Icon: icons.Must("refresh-cw"), Loading: s.auth.qr.Loading, Disabled: s.api == nil}).OnClick(s.newQRCode)
			s.errorRow(c, s.auth.qr.Error, s.newQRCode)
			s.errorRow(c, s.auth.error, s.checkAccount)
		})
	})
}

func (s *State) errorRow(c *ui.Context, message string, retry func()) {
	if message == "" {
		return
	}
	ui.Row(c).FillWidth().Wrap().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
		ui.Text(c, message).Grow(1).MinWidth(0).TextColor(c.Theme().Danger)
		if retry != nil {
			button(c.Key("retry"), "重试", components.ButtonOptions{Variant: components.ButtonGhost}).OnClick(retry)
		}
	})
}

func (s *State) requiresLogin(c *ui.Context, navigate func(string)) bool {
	if s.auth.loggedIn || s.api == nil {
		return false
	}
	panel(c, "登录提示", func() {
		ui.Column(c).FillWidth().Padding(style.SpaceXLarge).Gap(style.SpaceMedium).Center().Children(func() {
			caption(c, "请先登录哔哩哔哩账号")
			button(c, "前往账号", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() { navigate("/bilibili/profile") })
		})
	})
	return true
}
