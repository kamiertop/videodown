// Package bilibili owns the native Bilibili views and their persistent UI state.
package bilibili

import (
	"context"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// State contains data only: no frame-scoped elements or UI contexts.
type State struct {
	routeSeq uint64
	ctx      context.Context
	cancel   context.CancelFunc

	api             API
	downloadBackend DownloadBackend
	runtime         guiruntime.Runtime

	started, stopped, loggedIn, checking, loggingOut bool
	activePath, authorTarget                         string
	section, authError                               string
	epoch, pollSeq                                   uint64

	cancelPoll func()
	account    model.MyInfoProfile

	qr        resource[model.QRCodeData]
	qrImage   *ui.Bitmap
	qrExpired bool
	qrStatus  string
	images    imagecache.Cache

	follows    resource[model.FollowData]
	followPage int

	authorInfo   resource[model.UserInfoData]
	uploads      resource[model.VideoListData]
	authorLists  resource[model.SeasonsSeriesData]
	seasonVideos resource[model.SeasonsArchivesData]
	seriesVideos resource[model.SeriesArchivesData]
	authorTab    string

	authorPage, listPage, seasonPage, seriesPage, selectedSeason, selectedSeries int

	favoritesData  resource[model.FavoritesData]
	favoriteDetail resource[model.FavoriteData]
	favoriteID     int64
	favoritePage   int

	collections                                        resource[model.CollectionData]
	collectionDetail                                   resource[model.CollectionItemData]
	collectionHasMore                                  bool
	collectionID, collectionPage, collectionDetailPage int
	dynamics                                           resource[model.DynamicArchivePage]
	dynamicMore                                        resource[model.DynamicArchivePage]
	dynamicOffset                                      string
	dynamicHasMore                                     bool
	dynamicGeneration                                  uint64
	dynamicSelected                                    map[string]bool
	dynamicAutoPending                                 bool
	dynamicAutoAtY                                     float32
	dynamicAutoFillPages                               int

	playHistory              resource[model.PlayHistoryData]
	playHistoryMore          resource[model.PlayHistoryData]
	playHistoryCursor        model.PlayHistoryCursor
	playHistoryHasMore       bool
	playHistoryGeneration    uint64
	playHistorySelected      map[string]*bool
	playHistoryAutoPending   bool
	playHistoryAutoAtY       float32
	playHistoryAutoFillPages int

	selected                                                   map[string]*bool
	stagedUploads                                              []model.VListItem
	stagedFavorites                                            []model.FavoriteMedias
	stagedCollections                                          []model.CollectionItemMedias
	stagedSeasons                                              []model.SeasonsArchivesItem
	stagedSeries                                               []model.SeriesArchivesItem
	stagedDynamics                                             []model.DynamicArchiveItem
	stagedPlayHistory                                          []model.PlayHistoryItem
	stagedManual                                               []manualVideo
	notice                                                     string
	stageError                                                 string
	stagedKeys                                                 map[string]bool
	stagedSourceNames                                          map[string]string
	queueResolves                                              map[string]*queueResolve
	queueDownloading, queueDownloadStop                        bool
	queueDownloadTotal, queueDownloadDone, queueDownloadFailed int
	queueDownloadError                                         string
	queueProgress                                              map[string]bilidownload.Progress
	queueExtraBusy                                             map[string]string
	manualParse                                                resource[model.VideoDetailConciseData]
	manualGroup                                                *parsedVideoGroup
	manualParseInput                                           string

	batches                []*batchSession
	nextBatchID            uint64
	nextDownloadOrder      uint64
	pendingManualDownloads []manualDownloadRequest
	activeManualDownloads  *manualWork
	queueRequested         map[string]bool

	history                                                 resource[biliapi.HistoryPage]
	historyMore                                             resource[biliapi.HistoryPage]
	historyQuery                                            string
	historySearchPending, historyMutating, historyClearOpen bool
	historyActionError                                      string

	link        string
	author      string
	favoriteTab string
	searches    map[string]*string
	scrolls     map[string]*ui.ScrollState
	autoPages   map[string]*autoPageState
}

func NewState() *State {
	return &State{images: make(imagecache.Cache), stagedKeys: make(map[string]bool), stagedSourceNames: make(map[string]string), queueResolves: make(map[string]*queueResolve), queueProgress: make(map[string]bilidownload.Progress), queueExtraBusy: make(map[string]string), selected: make(map[string]*bool), dynamicSelected: make(map[string]bool), playHistorySelected: make(map[string]*bool), authorTab: "uploads", favoriteTab: "favorite", searches: make(map[string]*string), scrolls: make(map[string]*ui.ScrollState), autoPages: make(map[string]*autoPageState)}
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
	if (section == "favorite" || section == "collection") && s.favoriteTab != section {
		s.favoriteTab = section
		s.selected = make(map[string]*bool)
	}
	s.enter(section)
	k := core.Tokens(c)
	ui.Row(c.Key("bilibili-layout")).Fill().MinHeight(0).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c.Key("rail")).Label("哔哩哔哩导航").Width(style.PlatformNavWidth).Shrink(0).Padding(8).Gap(6).
			Background(k.Surface).BorderColor(k.Border).BorderWidth(0, style.BorderWidth, 0, 0).Children(func() {
			ui.Text(c, "B 站").FillWidth().TextAlign(ui.Center).FontSize(style.Font(c, style.FontCompact)).TextColor(k.TextMuted)
			layout.ScrollArea(c.Key("menu-scroll"), s.scroll("menu"), layout.ScrollAreaOptions{}, func() {
				ui.Column(c).FillWidth().Gap(6).Children(func() {
					names := []string{"download", "user", "play", "star", "history", "clock"}
					labels := []string{"下载", "UP 主", "动态", "收藏 / 合集", "播放历史", "下载历史"}
					for i, item := range p.Sections()[:6] {
						active := section == item.ID || item.ID == "favorite" && section == "collection"
						s.navButton(c.Key(item.ID), item.Label, labels[i], names[i], active, func() { navigate("/bilibili/" + item.ID) })
					}
				})
			}).Grow(1).MinHeight(0)
			if s.loggedIn && s.account.Face != "" {
				ui.Column(c.Key("account-avatar")).FillWidth().AlignItems(ui.Center).Gap(4).Children(func() { s.image(c, s.account.Face, 32, 32, true) })
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
	if detail && path == s.activePath && s.loggedIn {
		s.ensureAuthor(mid)
	}
	ui.Column(c.Key("content")).Label("哔哩哔哩内容").Grow(1).MinWidth(0).MinHeight(0).Children(func() {
		if section == "download" || section == "history" {
			ui.Column(c.Key(path)).Fill().MinWidth(0).MinHeight(0).Padding(style.PagePadding).Gap(style.SpaceMedium).FontSize(style.Body(c)).Children(func() {
				if section == "history" {
					s.downloadHistory(c)
				} else {
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
		layout.ScrollArea(c.Key(path), s.scroll(path), layout.ScrollAreaOptions{}, func() {
			_, height := c.Size()
			body := ui.Column(c.Key(section)).FillWidth().MinHeight(height - style.TitlebarHeight - 2*style.BorderWidth).Padding(style.PagePadding).Gap(style.SpaceMedium).FontSize(style.Body(c))
			body.Children(func() { s.Render(c, section, navigate) })
		}).Grow(1).MinWidth(0).MinHeight(0)
	})
}

func (s *State) navButton(c *ui.Context, label, caption, icon string, active bool, action func()) {
	k := core.Tokens(c)
	iconSVG := icons.Must(icon)
	b := components.Button(c, caption, input.ButtonOptions{Variant: input.Ghost, Icon: iconSVG, Label: label}).
		FillWidth().Height(style.PlatformNavItemHeight).Shrink(0).Column().Center().Gap(7).Padding(4).Radius(style.RadiusMenu).FontSize(style.Font(c, style.FontCompact)).
		Border(style.BorderWidth, k.Border.Alpha(.6)).Background(k.Background).TextColor(k.Text)
	if b.Hovered() {
		b.Background(k.Selection)
	}
	if active {
		b.Background(k.Selection).Border(style.BorderWidth, k.AccentText).TextColor(k.AccentText)
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
		s.dynamic(c, navigate)
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
