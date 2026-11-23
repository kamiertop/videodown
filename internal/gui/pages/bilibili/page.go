// Package bilibili owns the native Bilibili views and their persistent UI state.
package bilibili

import (
	"context"

	"github.com/egoist/mygo/ui"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// State contains data only: no frame-scoped elements or UI contexts.
// 各板块状态按域拆入子结构（定义在对应板块的文件里），不再全部平铺；
// 顶层只保留基础设施与跨板块共享的输入、选择状态。
type State struct {
	// 基础设施与生命周期。
	routeSeq, epoch uint64
	ctx             context.Context
	cancel          context.CancelFunc

	api             *biliapi.BiliBili
	downloadBackend *bilidownload.Service
	runtime         guiruntime.Runtime

	started, stopped    bool
	activePath, section string

	auth   authState        // 登录、二维码、账号资料（account.go）
	images imagecache.Cache // 封面与头像缓存

	follows followsState // 关注列表（authors.go）
	author  authorState  // UP 主主页：投稿、合集、系列（authors.go）
	library libraryState // 收藏夹与合集（favorites.go）
	dynamic dynamicState // 关注动态（dynamic.go）
	play    playState    // 播放历史（play_history.go）
	history historyState // 下载历史（library.go）
	stage   stageState   // 待下载暂存队列（staging.go）
	queue   queueState   // 暂存队列的解析与下载调度（download_queue.go）
	manual  manualState  // 链接解析（manual_download.go）
	batches batchState   // 批量下载会话（batch_download.go）

	// 顶层输入与跨板块共享的卡片选择状态。
	link, authorInput string
	selected          map[string]*bool
	searches          map[string]*string
	scrolls           map[string]*ui.ScrollState
	autoPages         map[string]*autoPageState
	downloadList      ui.ListState
}

func NewState() *State {
	return &State{
		images:    make(imagecache.Cache),
		selected:  make(map[string]*bool),
		author:    authorState{tab: "uploads"},
		library:   libraryState{tab: "favorite"},
		dynamic:   dynamicState{picked: make(map[string]bool)},
		play:      playState{picked: make(map[string]*bool)},
		stage:     stageState{keys: make(map[string]bool), sourceNames: make(map[string]string)},
		queue:     queueState{resolves: make(map[string]*queueResolve), progress: make(map[string]bilidownload.Progress), extraBusy: make(map[string]string), requested: make(map[string]bool)},
		searches:  make(map[string]*string),
		scrolls:   make(map[string]*ui.ScrollState),
		autoPages: make(map[string]*autoPageState),
	}
}
func (s *State) Link() *string { return &s.link }

// Layout has a narrow left rail, a separate account
// entry at its foot, and a right pane that owns all content scrolling.
func (s *State) Layout(c *ui.Context, path string, navigate func(string), content func()) {
	p, _ := navigation.ByPath("/bilibili")
	section := p.CurrentSection(path)
	if s.activePath != path {
		s.routeSeq++
	}
	s.activePath = path
	if (section == "favorite" || section == "collection") && s.library.tab != section {
		s.library.tab = section
		s.selected = make(map[string]*bool)
	}
	s.enter(section)
	k := c.Theme()
	ui.Row(c.Key("bilibili-layout")).Fill().MinHeight(0).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c.Key("rail")).Label("哔哩哔哩导航").Width(style.PlatformNavWidth).Shrink(0).Padding(8).Gap(6).
			Background(k.Surface).BorderColor(k.Border).BorderWidth(0, style.BorderWidth, 0, 0).Children(func() {
			ui.Text(c, "B 站").FillWidth().TextAlign(ui.Center).FontSize(style.Font(c, style.FontCompact)).TextColor(k.TextMuted)
			components.ScrollArea(c.Key("menu-scroll"), s.scroll("menu"), components.ScrollAreaOptions{}, func() {
				ui.Column(c).FillWidth().Gap(6).Children(func() {
					names := []string{"download", "user", "play", "star", "history", "clock"}
					labels := []string{"下载", "UP 主", "动态", "收藏 / 合集", "播放历史", "下载历史"}
					for i, item := range p.Sections()[:6] {
						active := section == item.ID || item.ID == "favorite" && section == "collection"
						s.navButton(c.Key(item.ID), item.Label, labels[i], names[i], active, func() { navigate("/bilibili/" + item.ID) })
					}
				})
			}).Grow(1).MinHeight(0)
			if s.auth.loggedIn && s.auth.account.Face != "" {
				ui.Column(c.Key("account-avatar")).FillWidth().AlignItems(ui.Center).Gap(4).Children(func() { s.image(c, s.auth.account.Face, 32, 32, true) })
			}
			s.navButton(c.Key("profile"), "账号", "账号", "user", section == "profile", func() { navigate("/bilibili/profile") })
		})
		content()
	})
}

// Form scrolls independently of the rail and of other subroutes.
func (s *State) Form(c *ui.Context, path string, navigate func(string), goBack func()) {
	p, _ := navigation.ByPath("/bilibili")
	section := p.CurrentSection(path)
	mid, detail := navigation.BilibiliAuthorID(path)
	if detail && path == s.activePath && s.auth.loggedIn {
		s.ensureAuthor(mid)
	}
	ui.Column(c.Key("content")).Label("哔哩哔哩内容").Grow(1).MinWidth(0).MinHeight(0).Children(func() {
		// download、history 与 up 自管内容滚动：up 的工具栏固定在顶部，
		// 仅关注列表在面板内滚动；up 的深层路由（UP 主主页）走下面的 detail 分支。
		if section == "download" || section == "history" || section == "up" && !detail {
			ui.Column(c.Key(path)).Fill().MinWidth(0).MinHeight(0).Padding(style.PagePadding).Gap(style.SpaceMedium).FontSize(style.Body(c)).Children(func() {
				switch section {
				case "history":
					s.downloadHistory(c)
				case "up":
					s.authors(c, navigate)
				default:
					s.download(c)
				}
			})
			return
		}
		if detail {
			ui.Column(c.Key(path)).Fill().MinWidth(0).MinHeight(0).Padding(style.PagePadding).
				Gap(style.SpaceMedium).FontSize(style.Body(c)).Children(func() {
				s.authorRoute(c, mid, navigate, goBack)
			})
			return
		}
		components.ScrollArea(c.Key(path), s.scroll(path), components.ScrollAreaOptions{}, func() {
			_, height := c.Size()
			body := ui.Column(c.Key(section)).FillWidth().MinHeight(height - style.TitlebarHeight - 2*style.BorderWidth).Padding(style.PagePadding).Gap(style.SpaceMedium).FontSize(style.Body(c))
			body.Children(func() { s.Render(c, section, navigate) })
		}).Grow(1).MinWidth(0).MinHeight(0)
	})
}

func (s *State) navButton(c *ui.Context, label, caption, icon string, active bool, action func()) {
	k := c.Theme()
	iconSVG := icons.Must(icon)
	b := components.Button(c, caption, components.ButtonOptions{Variant: components.ButtonGhost, Icon: iconSVG, Label: label}).
		FillWidth().Height(style.PlatformNavItemHeight).Shrink(0).Column().Center().Gap(7).Padding(4).Radius(style.RadiusMenu).FontSize(style.Font(c, style.FontCompact)).
		Border(style.BorderWidth, k.Border.Alpha(.6)).Background(k.Background).TextColor(k.Text)
	if b.Hovered() {
		b.Background(k.Selection)
	}
	if active {
		b.Background(k.Selection).Border(style.BorderWidth, k.Accent).TextColor(k.Accent)
	}
	b.OnClick(action)
}

func (s *State) Render(c *ui.Context, section string, navigate func(string)) {
	switch section {
	case "download":
		s.download(c)
	case "up":
		s.authors(c, navigate)
	case "dynamic":
		s.dynamicPage(c, navigate)
	case "favorite", "collection":
		s.favorites(c, navigate)
	case "play-history":
		s.playHistoryPage(c, navigate)
	case "history":
		s.downloadHistory(c)
	case "profile":
		s.profile(c)
	}
}

func (s *State) scroll(key string) *ui.ScrollState {
	if s.scrolls[key] == nil {
		s.scrolls[key] = &ui.ScrollState{}
	}
	return s.scrolls[key]
}
func (s *State) search(key string) *string {
	if s.searches[key] == nil {
		s.searches[key] = new(string)
	}
	return s.searches[key]
}
