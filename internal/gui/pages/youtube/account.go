package youtube

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
)

const cookieWikiURL = "https://github.com/yt-dlp/yt-dlp/wiki/Extractors#exporting-youtube-cookies"

func (s *State) accountPage(c *ui.Context, scroll *ui.ScrollState) {
	width, _ := c.Size()
	// Leave room for the platform navigation and both panels' padding.
	const twoColumnWidth = 960
	components.ScrollArea(c.Key("account-scroll"), scroll, components.ScrollAreaOptions{}, func() {
		ui.Column(c).Label("YouTube 账号页面").FillWidth().MinWidth(0).
			Padding(style.PagePadding).Gap(style.SpaceMedium).Children(func() {
			cookieExpiryNotice(c.Key("cookie-expiry"))
			ui.Row(c.Key("account-panels")).FillWidth().MinWidth(0).Wrap().
				AlignItems(ui.Start).Gap(style.SpaceMedium).Children(func() {
				editor := ui.Column(c.Key("login")).MinWidth(0)
				sidebarWidth := min(float32(400), width*0.34)
				if width < twoColumnWidth {
					editor.FillWidth().Shrink(0)
				} else {
					editor.Grow(1)
				}
				editor.Children(func() { s.cookieEditor(c) })

				sidebar := ui.Column(c.Key("sidebar")).MinWidth(0).Shrink(0).Gap(style.SpaceMedium)
				if width < twoColumnWidth {
					sidebar.FillWidth()
				} else {
					sidebar.Width(sidebarWidth)
				}
				sidebar.Children(func() {
					cookieGuide(c.Key("cookie-guide"))
					s.accountProxySettings(c.Key("proxy-settings"))
				})
			})
		})
	}).Grow(1).MinWidth(0).MinHeight(0)
}

func cookieExpiryNotice(c *ui.Context) {
	k := c.Theme()
	ui.Row(c).FillWidth().MinWidth(0).Padding(style.SpaceMedium).Gap(style.SpaceMedium).
		AlignItems(ui.Start).Radius(style.RadiusPanel).Background(k.Selection).
		Border(style.BorderWidth, k.Accent.Alpha(0.5)).Children(func() {
		ui.Icon(c, icons.Must("info")).Size(20, 20).Shrink(0).TextColor(k.Accent)
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Text(c, "Cookie 可能很快失效").FillWidth().FontWeight(style.WeightSemibold).TextColor(k.Accent)
			ui.Text(c, "YouTube 可能在浏览器继续使用时轮换登录 Cookie。刷新帐号失效时，可能需要重新粘贴最新的 Cookie").FillWidth()
		})
	})
}

func (s *State) cookieEditor(c *ui.Context) {
	panel(c, "Cookie 设置", func() {
		ui.Column(c).FillWidth().MinWidth(0).Padding(style.SpaceLarge).Gap(style.SpaceMedium).Children(func() {
			ui.Text(c, "登录 Cookie").FontWeight(style.WeightSemibold).FontSize(style.Font(c, style.FontTitle))
			s.accountStatus(c)
			components.TextArea(c.Key("cookie-input"), &s.cookie, components.TextAreaOptions{
				Label:       "YouTube Cookie",
				Placeholder: "粘贴单行 Cookie 请求头；留空保存可清除",
				MinHeight:   230,
				MaxHeight:   300,
				Disabled:    s.busy || s.api == nil,
			}).FillWidth()
			ui.Row(c).FillWidth().MinWidth(0).Wrap().Gap(style.SpaceSmall).Children(func() {
				components.Button(c.Key("save-cookie"), "保存并验证", components.ButtonOptions{Variant: components.ButtonPrimary, Loading: s.busy, Disabled: s.busy || s.proxyBusy || s.api == nil}).OnClick(s.saveCookie)
				if s.cookieSaved {
					components.Button(c.Key("refresh-account"), "刷新账号", components.ButtonOptions{Variant: components.ButtonOutline, Loading: s.accountBusy, Disabled: s.accountBusy || s.busy || s.proxyBusy || s.api == nil}).OnClick(s.checkAccount)
				}
			})
			if s.authMessage != "" {
				components.Muted(c, s.authMessage)
			}
		})
	})
}

func cookieGuide(c *ui.Context) {
	services := c.Services()
	panel(c, "Cookie 使用说明", func() {
		ui.Column(c).FillWidth().MinWidth(0).Padding(style.SpaceLarge).Gap(style.SpaceSmall).Children(func() {
			ui.Text(c, "普通公开视频和公开播放列表无需 Cookie").FillWidth().FontWeight(style.WeightSemibold)
			ui.Text(c, "即使保存了 Cookie，也会优先以访客身份解析。私人播放列表、会员专属或年龄限制内容可能需要登录。").FillWidth()
			components.Muted(c, "需要登录时，按指南在独立的无痕窗口登录 YouTube，从网络面板复制 YouTube 请求的单行 Cookie 值。关闭无痕窗口后，在这里保存并验证。")
			components.Button(c.Key("wiki"), "查看 yt-dlp Cookie 导出指南", components.ButtonOptions{Variant: components.ButtonOutline}).
				OnClick(func() { services.OpenURL(cookieWikiURL) })
		})
	})
}

func (s *State) accountProxySettings(c *ui.Context) {
	disabled := s.busy || s.playlistLoading || s.detailsInFlight > 0 || s.readyCount() > 0 || s.PendingDownloadCount() > 0 || s.api == nil
	panel(c, "YouTube 代理设置", func() {
		ui.Column(c).FillWidth().MinWidth(0).Padding(style.SpaceLarge).Gap(style.SpaceSmall).Children(func() {
			ui.Row(c).FillWidth().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "代理服务器").Grow(1).FontWeight(style.WeightSemibold)
				components.SettingsSwitch(c.Key("proxy-enabled"), &s.proxyEnabled, "启用 YouTube 代理", components.SwitchOptions{Disabled: disabled}).OnClick(s.saveProxy)
			})
			field, editor := roundedInput(c.Key("proxy-input"), &s.proxy, "YouTube 代理服务器", "http://127.0.0.1:7890", disabled, "")
			field.FillWidth()
			editor.OnChange(s.scheduleProxySave).OnSubmit(s.saveProxy)
			components.Muted(c, "支持 HTTP、HTTPS 和 SOCKS5，修改后自动保存。关闭后仅 YouTube 改为直连。")
			if s.proxyMessage != "" {
				components.Muted(c, s.proxyMessage)
			}
		})
	})
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
			if s.account.Avatar != "" || s.account.Name != "" || s.account.Email != "" {
				s.accountAvatarView(c, style.ProfileAvatarSize)
			}
			ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
				if s.account.Name != "" {
					ui.Text(c, s.account.Name).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontTitle)).FontWeight(style.WeightSemibold)
				}
				if s.account.Email != "" {
					components.Muted(c, s.account.Email)
				}
				ui.Text(c, fmt.Sprintf("Cookie 有效，已登录（%s 验证）", s.accountCheckedAt.Format("15:04:05"))).TextColor(k.Accent)
			})
		})
	case s.accountError != nil:
		ui.Text(c, "Cookie 验证失败："+s.accountError.Error()).FillWidth().MaxLines(3).TextColor(k.Danger)
		if errors.Is(s.accountError, yt.ErrAccountNetwork) {
			// 请求未到达 YouTube：与 Cookie 无效无关，引导检查网络出口。
			// 直连失败有两条出路：开启 TUN（程序直连会透明经过虚拟网卡），
			// 或启用与浏览器相同的应用层代理。
			hint := "程序当前直连 YouTube；可开启代理软件的 TUN（虚拟网卡）模式，或在此启用与浏览器相同的代理后重试"
			if s.proxyEnabled {
				hint = "当前已通过代理连接 YouTube，请确认代理可用后重试"
			}
			components.Muted(c, hint)
		} else if errors.Is(s.accountError, yt.ErrCookieNotLoggedIn) {
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
