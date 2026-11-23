package bilibili

import (
	"fmt"
	"github.com/kamiertop/videodown/internal/gui/components"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/icons"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// followsState 是“UP 主”页关注列表的分页状态：list 随滚动持续追加，
// page 记录已连续加载的页数，退出登录后从第一页重新开始。
type followsState struct {
	list resource[model.FollowData]
	page int
}

// authorState 是单个 UP 主主页的完整状态：资料、投稿、合集/系列列表与
// 当前选中的合集/系列，以及各自的页码。target 是输入框里的原始解析目标。
type authorState struct {
	target         string
	info           resource[model.UserInfoData]
	uploads        resource[model.VideoListData]
	lists          resource[model.SeasonsSeriesData]
	seasonVideos   resource[model.SeasonsArchivesData]
	seriesVideos   resource[model.SeriesArchivesData]
	tab            string
	uploadsPage    int
	listPage       int
	seasonPage     int
	seriesPage     int
	selectedSeason int
	selectedSeries int
}

func (s *State) authors(c *ui.Context, navigate func(string)) {
	if s.requiresLogin(c, navigate) {
		return
	}
	// 顶部工具栏固定不随列表滚动；内边距收紧一档，给列表多留高度。
	panel(c, "UP 主工具栏", func() {
		ui.Row(c).
			FillWidth().
			Padding(style.SpaceSmall).
			Gap(style.SpaceSmall).
			AlignItems(ui.Center).
			Children(func() {
				ui.Text(c, fmt.Sprintf("我的关注 · %d", s.follows.list.Data.Total)).
					FontWeight(style.WeightSemibold).
					Shrink(0)
				button(c.Key("refresh"),
					"刷新",
					components.ButtonOptions{Variant: components.ButtonOutline, Loading: s.follows.list.Loading, Disabled: s.api == nil},
				).OnClick(func() { s.loadFollows(false) })
				ui.Spacer(c).Grow(1)
				textInput(c.Key("author"),
					&s.authorInput,
					"UP 主 UID 或空间链接",
					"输入 UP 主 mid / 空间链接",
					func() { s.findAuthor(s.authorInput, navigate) },
				).Width(450)

				button(c.Key("parse-author"), "解析",
					components.ButtonOptions{Loading: s.author.info.Loading, Disabled: s.api == nil},
				).OnClick(func() { s.findAuthor(s.authorInput, navigate) })
			})
	})
	s.errorRow(c, s.author.info.Error, func() { s.findAuthor(s.authorInput, navigate) })
	s.errorRow(c, s.follows.list.Error, func() { s.loadFollows(s.follows.list.Loaded) })
	panel(c, "关注列表", func() {
		components.ScrollArea(c.Key(followsScrollKey), s.scroll(followsScrollKey), components.ScrollAreaOptions{Label: "关注列表区域"}, func() {
			ui.Column(c).FillWidth().Padding(style.SpaceSmall, style.SpaceMedium, style.SpaceMedium).Gap(style.SpaceMedium).Children(func() {
				if len(s.follows.list.Data.List) == 0 {
					if s.follows.list.Loading {
						components.Spinner(c, components.SpinnerOptions{Size: 16, Label: "正在加载关注列表", ShowLabel: true})
					} else {
						empty(c, "暂无关注列表", "user")
					}
					return
				}
				ui.Grid(c).
					Columns(authorColumns(c)).
					FillWidth().
					Gap(style.SpaceMedium).
					Children(func() {
						for _, author := range s.follows.list.Data.List {
							s.authorCard(c.Key(author.Mid), author, navigate)
						}
					})
				// 下拉到底部自动追加下一页，不再用整页替换的翻页按钮。
				if s.follows.list.Loading {
					components.Spinner(c, components.SpinnerOptions{Size: 16, Label: "正在加载关注", ShowLabel: true})
				} else if len(s.follows.list.Data.List) < s.follows.list.Data.Total {
					caption(c, fmt.Sprintf("已加载 %d / %d 位 UP 主，滚动到底部自动加载", len(s.follows.list.Data.List), s.follows.list.Data.Total))
					const key = "author-follows"
					if s.autoPageFallback(key, followsScrollKey) && s.follows.list.Error == "" {
						button(c, "继续加载", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() {
							s.continueAutoPage(key, func() { s.loadFollows(true) })
						})
					}
				} else {
					caption(c, fmt.Sprintf("已加载全部 %d 位 UP 主", s.follows.list.Data.Total))
				}
			})
		}).Grow(1).MinWidth(0).MinHeight(0)
	}).Grow(1).MinHeight(0)
	s.maybeMoreAuthorFollows()
}

func authorColumns(c *ui.Context) int {
	width, _ := c.Size()
	available := width - style.PlatformNavWidth - 2*style.PagePadding - 2*style.SpaceMedium
	columns := int((available + style.SpaceMedium) / (style.AuthorCardMinWidth + style.SpaceMedium))

	return max(1, min(style.AuthorGridMaxColumns, columns))
}

func (s *State) authorDetail(c *ui.Context, navigate func(string), goBack func()) {
	profile := s.author.info.Data
	panel(c, "UP 主信息", func() {
		ui.Row(c).FillWidth().
			Padding(style.AuthorDetailPaddingY, style.AuthorDetailPaddingX).
			Gap(style.AuthorDetailGap).AlignItems(ui.Center).
			Children(func() {
				s.image(c, profile.Face, style.AuthorDetailAvatarSize, style.AuthorDetailAvatarSize, true)
				ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
					ui.Row(c).FillWidth().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
						ui.Text(c, profile.Name).FontWeight(style.WeightBold).MaxLines(1)
						caption(c, fmt.Sprintf("Lv.%d", profile.Level))
						if profile.IsFollowed {
							caption(c, "已关注")
						}
					})
					if strings.TrimSpace(profile.Sign) != "" {
						ui.Text(c, profile.Sign).FillWidth().MaxLines(1).TextColor(c.Theme().TextMuted)
					}
				})
				button(c.Key("back-to-authors"), "返回", components.ButtonOptions{Variant: components.ButtonOutline, Icon: icons.Must("chevron-left")}).OnClick(goBack)
			})
	})
	s.authorContentControls(c, navigate)
	components.ScrollArea(c.Key("author-videos-scroll"), s.scroll(fmt.Sprintf("author-videos/%d", profile.Mid)),
		components.ScrollAreaOptions{Label: "UP 主视频区域"}, func() {
			leftPadding := style.ScrollbarGutter
			if s.author.tab == "lists" {
				leftPadding = 0
			}
			ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, 0, leftPadding).
				Gap(style.SpaceMedium).Children(func() { s.authorVideoContent(c) })
		}).Grow(1).MinWidth(0).MinHeight(0)
	s.maybeAutoAuthorVideos(profile.Mid)
	s.maybeAutoAuthorLists(profile.Mid)
}

func (s *State) authorVideoContent(c *ui.Context) {
	s.stagingNotice(c)
	s.errorRow(c, s.stage.error, nil)
	switch {
	case s.author.tab == "uploads":
		if len(s.author.uploads.Data.List.Vlist) == 0 {
			if s.author.uploads.Loading {
				caption(c, "正在加载投稿")
			} else {
				empty(c, "暂无投稿", "play")
			}
			return
		}
		ui.Grid(c).
			Columns(s.columns(c, false)).
			FillWidth().
			Gap(style.SpaceMedium).
			Children(func() {
				for _, video := range s.author.uploads.Data.List.Vlist {
					s.mediaCard(c.Key(video.Bvid), video.Bvid, video.Title, video.Pic, video.Author, video.Length, video.Play, video.VideoReview, video.Created)
				}
			})
		if s.author.uploads.Loading {
			caption(c, "正在加载投稿")
		}
		s.errorRow(c, s.author.uploads.Error, func() { s.loadUploads(max(1, s.author.uploadsPage+1), s.author.uploads.Loaded) })
		if len(s.author.uploads.Data.List.Vlist) < s.author.uploads.Data.Page.Count {
			key := fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid)
			if s.autoPageFallback(key, key) && s.author.uploads.Error == "" && !s.author.uploads.Loading {
				button(c, "继续加载", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() {
					s.continueAutoPage(key, func() { s.loadUploads(s.author.uploadsPage+1, true) })
				})
			}
		}
	default:
		s.authorListsView(c)
	}
}

func (s *State) authorListsView(c *ui.Context) {
	lists := s.author.lists.Data.ItemsLists
	validLists := 0
	for _, item := range lists.SeasonsList {
		if item.Meta.SeasonId > 0 {
			validLists++
		}
	}
	for _, item := range lists.SeriesList {
		if item.Meta.SeriesId > 0 {
			validLists++
		}
	}
	if validLists == 0 {
		s.authorListsEmpty(c)
		return
	}
	width, height := c.Size()
	viewportOffset := style.AuthorListViewportOffsetWide
	if width < style.AuthorControlsDownloadBreakpoint {
		viewportOffset = style.AuthorListViewportOffsetCompact
	}
	ui.Row(c).
		FillWidth().
		MinWidth(0).
		MinHeight(max(320, height-viewportOffset)).
		Gap(style.SpaceMedium).
		AlignItems(ui.Stretch).
		Children(func() {
			panel(c.Key("author-lists"), "合集与系列", func() {
				ui.Column(c).Width(style.AuthorListWidth).Padding(style.SpaceSmall).Gap(6).Children(func() {
					ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
						ui.Text(c, "列表").Grow(1).FontWeight(style.WeightSemibold)
						button(c.Key("refresh-lists"), "",
							components.ButtonOptions{Variant: components.ButtonGhost, Icon: icons.Must("refresh-cw"), Label: "刷新合集与系列"},
						).Size(30, 30).
							Padding(0).
							OnClick(func() { s.loadAuthorLists(1, false) })
					})
					s.errorRow(c, s.author.lists.Error, func() {
						if s.author.lists.Loaded {
							s.loadAuthorLists(s.author.listPage+1, true)
						} else {
							s.loadAuthorLists(1, false)
						}
					})
					for _, item := range lists.SeasonsList {
						if item.Meta.SeasonId <= 0 {
							continue
						}
						name := item.Meta.Title
						if name == "" {
							name = item.Meta.Name
						}
						authorListItem(c.Key(fmt.Sprintf("season-%d", item.Meta.SeasonId)), name, "合集", item.Meta.Total,
							item.Meta.SeasonId == s.author.selectedSeason, func() { s.loadSeason(item.Meta.SeasonId, 1, false) })
					}
					for _, item := range lists.SeriesList {
						if item.Meta.SeriesId <= 0 {
							continue
						}
						authorListItem(c.Key(fmt.Sprintf("series-%d", item.Meta.SeriesId)), item.Meta.Name, "系列", item.Meta.Total,
							item.Meta.SeriesId == s.author.selectedSeries, func() { s.loadSeries(item.Meta.SeriesId, 1, false) })
					}
					if s.author.lists.Loading {
						caption(c, "正在加载列表")
					}
					if len(lists.SeasonsList)+len(lists.SeriesList) < lists.Page.Total {
						key := fmt.Sprintf("author-lists/%d", s.author.info.Data.Mid)
						scrollKey := fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid)
						if s.autoPageFallback(key, scrollKey) && s.author.lists.Error == "" && !s.author.lists.Loading {
							button(c.Key("more-lists"), "继续加载", components.ButtonOptions{Variant: components.ButtonOutline}).
								FillWidth().
								OnClick(func() {
									s.continueAutoPage(key, func() { s.loadAuthorLists(s.author.listPage+1, true) })
								})
						}
					}
				})
			}).Shrink(0)
			ui.Column(c).Grow(1).MinWidth(0).Gap(style.SpaceMedium).Children(func() {
				switch {
				case s.author.selectedSeason != 0:
					s.authorSeason(c)
				case s.author.selectedSeries != 0:
					s.authorSeries(c)
				default:
					empty(c, "选择一个合集或系列", "folder")
				}
			})
		})
}

func authorListItem(c *ui.Context, name, kind string, count int, selected bool, onClick func()) {
	k := c.Theme()
	background, border := k.Surface, k.Border.Alpha(.35)
	iconBackground, iconColor := k.Background, k.TextMuted
	badgeBackground, badgeColor := k.Background, k.TextMuted
	entry := ui.ButtonBase(c).Label(kind+" · "+name).FillWidth().MinWidth(0).
		Height(style.AuthorListItemHeight).Padding(6, style.SpaceSmall).Gap(style.SpaceSmall).
		Radius(style.RadiusMenu).Border(style.BorderWidth, border).Background(background).TextColor(k.Text).Clip()
	if selected {
		background, border = k.Selection, k.Accent.Alpha(.65)
		iconBackground, iconColor = k.Accent.Alpha(.15), k.Accent
		badgeBackground, badgeColor = k.Accent.Alpha(.12), k.Accent
	} else if entry.Hovered() || entry.FocusVisible() {
		background, border = k.SurfaceHover, k.Accent.Alpha(.5)
		iconBackground, iconColor = k.Accent.Alpha(.1), k.Accent
	}
	entry.Background(background).Border(style.BorderWidth, border).Children(func() {
		iconName := "folder"
		if kind == "系列" {
			iconName = "list"
		}
		ui.Box(c).
			Size(style.AuthorListIconSize, style.AuthorListIconSize).
			Shrink(0).
			Radius(style.RadiusControl).
			Border(style.BorderWidth, border).
			Background(iconBackground).
			Center().
			Children(func() {
				ui.Icon(c, icons.Must(iconName)).Size(17, 17).TextColor(iconColor)
			})

		ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
			title := ui.Text(c, name).FillWidth().SingleLine().FontWeight(style.WeightSemibold)
			if selected {
				title.TextColor(k.Accent)
			}
			ui.Text(c, kind).FillWidth().SingleLine().TextColor(k.TextMuted)
		})
		ui.Row(c).MinWidth(28).Height(24).Shrink(0).Padding(0, 6).
			Radius(12).Background(badgeBackground).Center().Children(func() {
			ui.Text(c, fmt.Sprint(count)).SingleLine().FontWeight(style.WeightSemibold).TextColor(badgeColor)
		})
	}).OnClick(onClick)
}

func (s *State) authorListsEmpty(c *ui.Context) {
	panel(c, "合集与系列空状态", func() {
		ui.Column(c).FillWidth().Grow(1).MinHeight(320).Padding(style.SpaceXLarge).Center().Gap(style.SpaceMedium).Children(func() {
			switch {
			case s.author.lists.Loading:
				components.Spinner(c, components.SpinnerOptions{Size: 32})
				caption(c, "正在加载合集与系列")
			case s.author.lists.Error != "":
				ui.Icon(c, icons.Must("folder")).Size(36, 36).TextColor(c.Theme().TextMuted)
				ui.Text(c, s.author.lists.Error).MaxLines(2).TextAlign(ui.Center).TextColor(c.Theme().Danger)
				button(c.Key("retry-lists"), "重试", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() { s.loadAuthorLists(1, false) })
			default:
				ui.Icon(c, icons.Must("folder")).Size(36, 36).TextColor(c.Theme().TextMuted)
				ui.Text(c, "暂无合集或系列").FontWeight(style.WeightSemibold)
				button(c.Key("show-uploads"), "查看投稿", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() { s.author.tab = "uploads" })
			}
		})
	}).Grow(1)
}

func (s *State) authorVideoKeys() []string {
	var keys []string
	switch {
	case s.author.tab == "uploads":
		for _, v := range s.author.uploads.Data.List.Vlist {
			if strings.HasPrefix(v.Bvid, "BV") {
				keys = append(keys, v.Bvid)
			}
		}
	case s.author.selectedSeason != 0:
		for _, v := range s.author.seasonVideos.Data.Archives {
			if strings.HasPrefix(v.Bvid, "BV") {
				keys = append(keys, v.Bvid)
			}
		}
	case s.author.selectedSeries != 0:
		for _, v := range s.author.seriesVideos.Data.Archives {
			if strings.HasPrefix(v.Bvid, "BV") {
				keys = append(keys, v.Bvid)
			}
		}
	}
	return keys
}

func (s *State) authorContentControls(c *ui.Context, navigate func(string)) {
	loaded, total := len(s.author.uploads.Data.List.Vlist), s.author.uploads.Data.Page.Count
	if s.author.tab == "lists" {
		loaded, total = 0, 0
		if s.author.selectedSeason != 0 {
			loaded, total = len(s.author.seasonVideos.Data.Archives), s.author.seasonVideos.Data.Page.Total
		} else if s.author.selectedSeries != 0 {
			loaded, total = len(s.author.seriesVideos.Data.Archives), s.author.seriesVideos.Data.Page.Total
		}
	}
	keys := s.authorVideoKeys()
	selected := 0
	for _, key := range keys {
		if *s.selection(key) {
			selected++
		}
	}
	all := len(keys) > 0 && selected == len(keys)
	tabs := func() {
		for _, tab := range []struct{ id, label string }{{"uploads", "投稿"}, {"lists", "合集 / 系列"}} {
			variant := components.ButtonSecondary
			if s.author.tab == tab.id {
				variant = components.ButtonPrimary
			}
			button(c.Key(tab.id), tab.label, components.ButtonOptions{Variant: variant}).MinWidth(104).Height(style.InputHeight).FontWeight(style.WeightSemibold).Shrink(0).OnClick(func() {
				if s.author.tab != tab.id {
					videoKey := fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid)
					listKey := fmt.Sprintf("author-lists/%d", s.author.info.Data.Mid)
					s.resetAutoPage(videoKey, videoKey)
					s.resetAutoPage(listKey, videoKey)
					s.author.tab = tab.id
					s.selected = make(map[string]*bool)
				}
				if tab.id == "lists" && !s.author.lists.Loaded {
					s.loadAuthorLists(1, false)
				}
			})
		}
		if loaded > 0 || total > 0 {
			caption(c, fmt.Sprintf("已加载 %d/%d", loaded, max(loaded, total)))
		}
	}
	actions := func() {
		if len(keys) > 0 {
			label := "全选"
			if all {
				label = "取消全选"
			}
			button(c.Key("select-all"), label, components.ButtonOptions{Variant: components.ButtonGhost}).OnClick(func() {
				for _, key := range keys {
					*s.selection(key) = !all
				}
			})
			button(c.Key("download-selected"), fmt.Sprintf("下载已选 (%d)", selected), components.ButtonOptions{Variant: components.ButtonOutline, Disabled: selected == 0}).OnClick(func() {
				if s.stageSelectedAuthorVideos() == 0 {
					return
				}
				for _, key := range keys {
					*s.selection(key) = false
				}
				navigate("/bilibili/download")
			})
		}
	}
	downloadControls := func() {
		button(c.Key("download-all"), "一键下载全部", components.ButtonOptions{Variant: components.ButtonPrimary, Disabled: len(keys) == 0 || s.api == nil || s.downloadBackend == nil}).OnClick(func() { s.downloadAllAuthor(navigate) })
	}

	panel(c, "UP 主内容操作", func() {
		width, _ := c.Size()
		ui.Column(c).FillWidth().MinWidth(0).Padding(style.SpaceMedium).Gap(style.SpaceSmall).Children(func() {
			ui.Row(c).FillWidth().MinWidth(0).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
				tabs()
				ui.Spacer(c).Grow(1)
				actions()
				if width >= style.AuthorControlsDownloadBreakpoint {
					downloadControls()
				}
			})
			if width < style.AuthorControlsDownloadBreakpoint {
				ui.Row(c).FillWidth().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
					ui.Spacer(c).Grow(1)
					downloadControls()
				})
			}
		})
	})
}
func (s *State) authorSeason(c *ui.Context) {
	ui.Grid(c).Columns(s.columns(c, true)).FillWidth().Gap(style.SpaceMedium).Children(func() {
		for _, video := range s.author.seasonVideos.Data.Archives {
			s.mediaCard(c.Key(video.Bvid), video.Bvid, video.Title, video.Pic, s.author.info.Data.Name, duration(video.Duration), video.Stat.View, video.Stat.Danmaku, video.Pubdate)
		}
	})
	if s.author.seasonVideos.Loading {
		caption(c, "正在加载合集视频")
	}
	s.errorRow(c, s.author.seasonVideos.Error, func() {
		s.loadSeason(s.author.selectedSeason, max(1, s.author.seasonPage+1), s.author.seasonVideos.Loaded)
	})
	if len(s.author.seasonVideos.Data.Archives) == 0 && !s.author.seasonVideos.Loading {
		empty(c, "暂无合集视频", "play")
	}
	if len(s.author.seasonVideos.Data.Archives) < s.author.seasonVideos.Data.Page.Total {
		key := fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid)
		if s.autoPageFallback(key, key) && s.author.seasonVideos.Error == "" && !s.author.seasonVideos.Loading {
			button(c, "继续加载", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() {
				s.continueAutoPage(key, func() { s.loadSeason(s.author.selectedSeason, s.author.seasonPage+1, true) })
			})
		}
	}
}
func (s *State) authorSeries(c *ui.Context) {
	ui.Grid(c).Columns(s.columns(c, true)).FillWidth().Gap(style.SpaceMedium).Children(func() {
		for _, video := range s.author.seriesVideos.Data.Archives {
			s.mediaCard(c.Key(video.Bvid), video.Bvid, video.Title, video.Pic, s.author.info.Data.Name, duration(video.Duration), video.Stat.View, video.Stat.Danmaku, video.PubDate)
		}
	})
	if s.author.seriesVideos.Loading {
		caption(c, "正在加载系列视频")
	}
	s.errorRow(c, s.author.seriesVideos.Error, func() {
		s.loadSeries(s.author.selectedSeries, max(1, s.author.seriesPage+1), s.author.seriesVideos.Loaded)
	})
	if len(s.author.seriesVideos.Data.Archives) == 0 && !s.author.seriesVideos.Loading {
		empty(c, "暂无系列视频", "play")
	}
	if len(s.author.seriesVideos.Data.Archives) < s.author.seriesVideos.Data.Page.Total {
		key := fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid)
		if s.autoPageFallback(key, key) && s.author.seriesVideos.Error == "" && !s.author.seriesVideos.Loading {
			button(c, "继续加载", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() {
				s.continueAutoPage(key, func() { s.loadSeries(s.author.selectedSeries, s.author.seriesPage+1, true) })
			})
		}
	}
}

// The card itself is the control, keeping pointer and keyboard activation on
// the same target. Reserve hint space so hover never moves the grid.
func (s *State) authorCard(c *ui.Context, author model.FollowDataList, navigate func(string)) {
	k := c.Theme()
	card := ui.ButtonBase(c).Label("进入"+author.Uname+"的主页").
		Column().AlignSelf(ui.Stretch).MinWidth(0).Padding(style.AuthorCardPadding).Gap(style.AuthorCardGap).AlignItems(ui.Center).Clip().
		Radius(style.RadiusPanel).Border(style.BorderEmphasis, k.Border.Alpha(.7)).Background(k.Surface).TextColor(k.Text)
	highlighted := card.Hovered() || card.FocusVisible()
	if highlighted {
		card.Border(style.BorderEmphasis, k.Accent).Background(k.Selection).Shadow(0, 3, 12, 0, k.Accent.Alpha(.25))
	}
	card.Children(func() {
		s.image(c, author.Face, style.AuthorAvatarSize, style.AuthorAvatarSize, true)
		ui.Text(c, author.Uname).FillWidth().TextAlign(ui.Center).FontWeight(style.WeightSemibold).MaxLines(1)
		ui.Text(c, author.Sign).FillWidth().TextAlign(ui.Center).MaxLines(2).TextColor(k.TextMuted)
		ui.Row(c).FillWidth().MinWidth(0).Height(style.AuthorHintHeight).Clip().Justify(ui.Center).AlignItems(ui.Center).Gap(4).Children(func() {
			if highlighted {
				ui.Text(c, "进入主页").SingleLine().MaxLines(1).FontWeight(style.WeightSemibold).TextColor(k.Accent)
				ui.Icon(c, icons.Must("chevron-right")).Size(16, 16).Shrink(0).TextColor(k.Accent)
			}
		})
	}).OnClick(func() { navigate(navigation.BilibiliAuthorPath(uint64(author.Mid))) })
}

func (s *State) authorRoute(c *ui.Context, mid uint64, navigate func(string), goBack func()) {
	if s.requiresLogin(c, navigate) {
		return
	}
	if !s.author.info.Loaded || s.author.info.Data.Mid != mid {
		if s.author.info.Error != "" {
			s.errorRow(c, s.author.info.Error, func() { s.findAuthor(strconv.FormatUint(mid, 10), nil) })
		} else {
			empty(c, "正在加载 UP 主视频", "loader-circle")
		}
		return
	}
	s.authorDetail(c, navigate, goBack)
}
