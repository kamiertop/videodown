package gui

import (
	"strings"

	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/gui/style"

	"github.com/egoist/mygo/ui"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/theme"
)

// ellipsisVertical is the vertical three-dots control, as Lucide's
// more-vertical, which MujicaUI's icon set lacks.
var ellipsisVertical = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/></svg>`))

// arrowLeft and arrowRight point through history, as a browser toolbar's,
// which MujicaUI's icon set lacks.
var (
	arrowLeft  = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m12 19-7-7 7-7"/><path d="M19 12H5"/></svg>`))
	arrowRight = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m12 5 7 7-7 7"/><path d="M5 12h14"/></svg>`))
)

// Equal side columns keep the brand at the window center independently of
// navigation and control widths.
func (s *Shell) titleBar(c *ui.Context) {
	k := core.Tokens(c)
	d := core.Density(c)
	u := d.Unit()
	ui.Row(c).FillWidth().MinWidth(0).MinHeight(style.TitlebarHeight).Shrink(0).
		Padding(0, u*3).Gap(u*2).
		Background(k.Surface).BorderWidth(0, 0, style.BorderWidth, 0).BorderColor(k.Border).
		DragWindow().Children(func() {
		ui.Row(c).Grow(1).Basis(0).MinWidth(0).AlignItems(ui.Center).Gap(u).Children(func() {
			s.historyButton(c, -1)
			s.historyButton(c, 1)
			s.topNavigation(c)
		})
		titleBrand(c)
		ui.Row(c).Grow(1).Basis(0).MinWidth(0).Justify(ui.End).AlignItems(ui.Center).Gap(8).Children(func() {
			s.githubButton(c)
			s.windowButtons(c)
		})
	})
}

func titleBrand(c *ui.Context) {
	ui.Row(c).Label("应用标识").Shrink(0).AlignItems(ui.Center).Gap(core.Density(c).Unit() * 1.5).Children(func() {
		ui.Image(c, assets.AppIcon).Size(25, 25)
		ui.Text(c, appName).
			Font(theme.SerifFont).
			FontSize(style.Font(c, style.FontHeading)).
			FontWeight(style.WeightSemibold).
			SingleLine()
	})
}

// topNavigation links the main application pages. All hit areas
// are buttons, so clicking them never starts a title-bar window drag.
func (s *Shell) topNavigation(c *ui.Context) {
	width, _ := c.Size()
	compact := width < 1000
	ui.Row(c.Key("top-nav")).Shrink(0).Gap(4).Margin(0, 8, 0, 0).Children(func() {
		for _, item := range []struct{ path, label, icon string }{
			{"/", "首页", "home"}, {"/bilibili", "B 站", "bilibili"}, {"/douyin", "抖音", "douyin"}, {"/youtube", "YouTube", "youtube"},
		} {
			path := s.router.Path()
			active := path == item.path || item.path != "/" && strings.HasPrefix(path, item.path+"/")
			k := core.Tokens(c)
			button := ui.ButtonBase(c.Key(item.path)).Height(32).Shrink(0).Padding(0, 8).Gap(5).
				Radius(style.RadiusChrome).FontWeight(style.WeightSemibold).Label("前往" + item.label).Tooltip(item.label)
			bg, fg := ui.Transparent, k.TextMuted
			if active {
				bg, fg = k.Selection, k.AccentText
			} else if button.Hovered() {
				bg, fg = k.SurfaceHover, k.Text
			}
			button.Background(bg).TextColor(fg).Children(func() {
				if item.icon == "home" {
					ui.Icon(c, homeIcon).Size(16, 16)
				} else {
					ui.Image(c, assets.Platform(item.icon)).Size(16, 16)
				}
				if !compact {
					ui.Text(c, item.label).SingleLine()
				}
			})
			button.OnClick(func() {
				if !active {
					s.router.Push(item.path)
				}
			})
		}
	})
}

func (s *Shell) githubButton(c *ui.Context) {
	services := c.Services()
	input.IconButton(c.Key("github"), assets.GitHub, "打开 GitHub 仓库", input.IconButtonOptions{Variant: input.Ghost}).
		Size(32, 32).Tooltip("GitHub").OnClick(func() { services.OpenURL(assets.RepositoryURL) })
}

var homeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-8H9v8H4a1 1 0 0 1-1-1Z"/></svg>`))

// historyButton goes d pages through the router's history, disabled at its
// ends.
func (s *Shell) historyButton(c *ui.Context, d int) {
	key, label, can, icon := "back", "后退", s.router.CanGoBack(), arrowLeft
	if d > 0 {
		key, label, can, icon = "forward", "前进", s.router.CanGoForward(), arrowRight
	}
	if windowButton(c, key, icon, label, false, !can).Clicked() {
		s.router.Go(d)
	}
}

// windowButtons draws the app's settings control and the window's.
func (s *Shell) windowButtons(c *ui.Context) {
	ui.Row(c).Gap(4).Shrink(0).Children(func() {
		if input.IconButton(c, ellipsisVertical, "打开设置", input.IconButtonOptions{Variant: input.Ghost}).
			Margin(0, 4, 0, 0).Clicked() {
			s.router.Push("/settings")
		}

		if windowButton(c, "win-min", icons.Must("minus"), "最小化", false, false).Clicked() {
			s.win.Minimize()
		}
		isMax := s.win != nil && s.win.IsMaximized()
		icon, label := icons.Must("square"), "最大化"
		if isMax {
			icon, label = icons.Must("copy"), "还原"
		}
		if windowButton(c, "win-max", icon, label, false, false).Clicked() {
			if isMax {
				s.win.Unmaximize()
			} else {
				s.win.Maximize()
			}
		}
		if windowButton(c, "win-close", icons.Must("x"), "关闭", true, false).Clicked() {
			// Closing during the view's build destroys the window on its
			// own render stack, which crashes GTK's teardown of the
			// transparent surface; Update defers past this frame.
			s.win.Update(func() { s.win.Close() })
		}
	})
}

// windowButton is one borderless chrome control, tinted red on hover when
// danger, as the close button's, and muted when disabled.
func windowButton(c *ui.Context, key string, icon *ui.SVG, label string, danger, disabled bool) ui.Element {
	k := core.Tokens(c)
	b := ui.ButtonBase(c.Key(key)).
		Size(theme.MinHitSize, theme.MinHitSize).
		Radius(style.RadiusChrome).
		Label(label).
		Disabled(disabled)
	bg, fg := ui.Color{}, k.TextMuted
	switch {
	case disabled:
	case danger && b.Pressed():
		bg, fg = k.DangerFill.Mix(ui.RGB(0, 0, 0), 0.25), k.OnDanger
	case danger && b.Hovered():
		bg, fg = k.DangerFill, k.OnDanger
	case b.Pressed():
		bg, fg = k.Selection, k.Text
	case b.Hovered():
		bg, fg = k.SurfaceHover, k.Text
	}
	b.Background(bg).TextColor(fg)
	b.Children(func() { ui.Icon(c, icon).Size(16, 16) })

	return b
}
