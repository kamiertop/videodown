package youtube

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) accountPage(c *ui.Context, scroll *ui.ScrollState) {
	components.ScrollArea(c.Key("account-scroll"), scroll, components.ScrollAreaOptions{}, func() {
		ui.Column(c).
			Label("YouTube 账号页面").
			FillWidth().
			Padding(style.PagePadding).
			AlignItems(ui.Center).
			Children(func() {
				k := c.Theme()
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
						s.accountStatus(c)
						cookieField := components.TextArea(c.Key("cookie-input"),
							&s.cookie,
							components.TextAreaOptions{
								Label:       "YouTube Cookie",
								Placeholder: "可选：粘贴浏览器请求头中的 Cookie；留空保存可清除",
								MinHeight:   400,
								Disabled:    s.busy || s.api == nil},
						).FillWidth()
						_ = cookieField
						ui.Row(c).FillWidth().Gap(style.SpaceSmall).Children(func() {
							components.Button(c.Key("save-cookie"), "保存并验证", components.ButtonOptions{Variant: components.ButtonPrimary, Loading: s.busy, Disabled: s.busy || s.proxyBusy || s.api == nil}).OnClick(s.saveCookie)
							if s.cookieSaved {
								components.Button(c.Key("refresh-account"), "刷新账号", components.ButtonOptions{Variant: components.ButtonOutline, Loading: s.accountBusy, Disabled: s.accountBusy || s.busy || s.proxyBusy || s.api == nil}).OnClick(s.checkAccount)
							}
						})
						if s.authMessage != "" {
							components.Muted(c, s.authMessage)
						}
						proxyDisabled := s.busy || s.playlistLoading || s.detailsInFlight > 0 || s.readyCount() > 0 || s.PendingDownloadCount() > 0 || s.api == nil
						ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
							ui.Text(c, "代理服务器").Grow(1).FontWeight(style.WeightSemibold)
							components.SettingsSwitch(c.Key("proxy-enabled"), &s.proxyEnabled, "启用 YouTube 代理", components.SwitchOptions{Disabled: proxyDisabled}).OnClick(s.saveProxy)
						})
						proxyField := roundedInput(c.Key("proxy-input"), &s.proxy, "YouTube 代理服务器", "http://127.0.0.1:7890", proxyDisabled)
						proxyField.FillWidth().OnChange(s.scheduleProxySave).OnSubmit(s.saveProxy)
						components.Muted(c, "修改后自动保存。支持 HTTP、HTTPS、SOCKS5；关闭后直接连接，仅影响 YouTube。")

						if s.proxyMessage != "" {
							components.Muted(c, s.proxyMessage)
						}

					})
			})
	}).Grow(1).MinWidth(0).MinHeight(0)
}

// accountStatus 汇总 Cookie 验证状态：登录账号卡（头像、昵称、邮箱）、
// 验证失败原因或待验证提示，替换 Cookie 后会显示新账号信息。
func (s *State) accountStatus(c *ui.Context) {
	k := c.Theme()
	switch {
	case s.accountBusy:
		components.Spinner(c, components.SpinnerOptions{Size: 14, Label: "正在验证 Cookie", ShowLabel: true})
	case s.account.LoggedIn:
		ui.Row(c).FillWidth().Gap(style.SpaceLarge).AlignItems(ui.Center).Children(func() {
			s.accountAvatarView(c, style.ProfileAvatarSize)
			ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
				components.Muted(c, "当前登录账号")
				name := s.account.Name
				if name == "" {
					name = "YouTube 账号"
				}
				ui.Text(c, name).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontTitle)).FontWeight(style.WeightSemibold)
				if s.account.Email != "" {
					components.Muted(c, s.account.Email)
				}
				ui.Text(c, fmt.Sprintf("Cookie 有效，已登录（%s 验证）", s.accountCheckedAt.Format("15:04:05"))).TextColor(k.Accent)
			})
		})
	case s.accountError != "":
		ui.Text(c, "Cookie 验证失败："+s.accountError).FillWidth().MaxLines(3).TextColor(k.Danger)
		if s.accountNetwork {
			// 请求未到达 YouTube：与 Cookie 无效无关，引导检查网络出口。
			// 直连失败有两条出路：开启 TUN（程序直连会透明经过虚拟网卡），
			// 或启用与浏览器相同的应用层代理。
			hint := "程序当前直连 YouTube；可开启代理软件的 TUN（虚拟网卡）模式，或在此启用与浏览器相同的代理后重试"
			if s.proxyEnabled {
				hint = "当前已通过代理连接 YouTube，请确认代理可用后重试"
			}
			components.Muted(c, hint)
		} else {
			components.Muted(c, "请从已登录的浏览器重新复制 Cookie 后保存")
		}
	case s.cookieSaved:
		components.Muted(c, "已设置 Cookie（尚未验证）")
	default:
		components.Muted(c, "未设置 Cookie")
	}
}

// accountAvatarView 渲染圆形账号头像：加载中显示转圈，失败可重试。
func (s *State) accountAvatarView(c *ui.Context, size float32) {
	k := c.Theme()
	ui.Box(c).Label("YouTube 账号头像").Size(size, size).Shrink(0).Clip().Radius(size / 2).Background(k.Background).Center().Children(func() {
		switch {
		case s.accountAvatar != nil:
			ui.Image(c, s.accountAvatar).Fill().Fit(ui.Cover)
		case s.accountAvatarLoading:
			components.Spinner(c, components.SpinnerOptions{Size: 16, Label: "正在加载头像"})
		case s.accountAvatarError != "":
			components.Button(c.Key("retry-account-avatar"), "", components.ButtonOptions{Variant: components.ButtonGhost, Icon: icons.Must("refresh-cw"), Label: "重试账号头像"}).Size(28, 28).MinWidth(28).Padding(0).Tooltip(s.accountAvatarError).OnClick(func() {
				s.accountAvatarError = ""
				s.loadAccountAvatar()
			})
		default:
			ui.Icon(c, icons.Must("user")).Size(24, 24).TextColor(k.TextMuted)
		}
	})
}

// loadAccountAvatar 用代理感知的图片客户端加载账号头像（与封面同一管线）。
func (s *State) loadAccountAvatar() {
	target := s.account.Avatar
	if target == "" || s.runtime.Image == nil || s.accountAvatarLoading || s.stopped {
		return
	}
	if strings.HasPrefix(target, "//") {
		target = "https:" + target
	}
	s.accountAvatarLoading = true
	s.accountAvatarError = ""
	s.accountAvatarGeneration++
	generation, ctx, client := s.accountAvatarGeneration, s.ctx, s.imageClient
	var bitmap *ui.Bitmap
	s.runtime.Image(func() error {
		var err error
		loadCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
		defer cancel()
		if client != nil {
			bitmap, err = imagecache.LoadThumbnailWithClient(loadCtx, client, target, "https://www.youtube.com/")
		} else {
			bitmap, err = imagecache.LoadThumbnail(loadCtx, target, "https://www.youtube.com/")
		}
		return err
	}, func(err error) {
		if s.stopped || s.accountAvatarGeneration != generation {
			return
		}
		s.accountAvatarLoading = false
		if err != nil {
			s.accountAvatarError = "头像加载失败，点击重试"
			return
		}
		s.accountAvatar = bitmap
	})
}
