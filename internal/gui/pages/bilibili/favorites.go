package bilibili

import (
	"fmt"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) favorites(c *ui.Context, navigate func(string)) {
	if s.requiresLogin(c, navigate) {
		return
	}
	_, height := c.Size()
	panel(c, "我的视频库", func() {
		ui.Row(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).BorderColor(core.Tokens(c).Border).BorderWidth(0, 0, style.BorderWidth, 0).Children(func() {
			for _, tab := range []struct{ id, label string }{{"favorite", "收藏夹"}, {"collection", "合集"}} {
				variant := input.Secondary
				if s.favoriteTab == tab.id {
					variant = input.Primary
				}
				button(c.Key(tab.id), tab.label, input.ButtonOptions{Variant: variant}).MinWidth(84).Height(style.InputHeight).OnClick(func() {
					if s.favoriteTab != tab.id {
						navigate("/bilibili/" + tab.id)
					}
				})
			}
		})
		title := "收藏夹"
		count := int(s.favoritesData.Data.Count)
		busy := s.favoritesData.Loading
		if s.favoriteTab == "collection" {
			title = "合集"
			count = s.collections.Data.Count
			busy = s.collections.Loading
		}
		folderWidth := style.LibraryListWidth
		if width, _ := c.Size(); width < style.AuthorControlsDownloadBreakpoint {
			folderWidth = style.LibraryListWidth - style.SpaceLarge
		}
		ui.Row(c.Key(s.favoriteTab)).FillWidth().Grow(1).MinHeight(0).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c.Key("folders")).Label("视频库分组").Width(folderWidth).Shrink(0).MinHeight(0).
				BorderColor(core.Tokens(c).Border).BorderWidth(0, style.BorderWidth, 0, 0).Children(func() {
				ui.Row(c).Label("视频库列表标题").FillWidth().Height(style.LibraryHeaderHeight).Padding(style.SpaceSmall, style.SpaceMedium).Gap(style.SpaceSmall).
					AlignItems(ui.Center).BorderColor(core.Tokens(c).Border).BorderWidth(0, 0, style.BorderWidth, 0).Children(func() {
					ui.Text(c, title).FontWeight(style.WeightSemibold).Shrink(0)
					ui.Box(c).MinWidth(28).Height(24).Padding(0, 6).Radius(12).Background(core.Tokens(c).Accent.Alpha(.12)).Center().Children(func() {
						ui.Text(c, fmt.Sprint(count)).SingleLine().FontSize(style.Font(c, style.FontMeta)).FontWeight(style.WeightSemibold).TextColor(core.Tokens(c).Accent)
					})
					ui.Spacer(c).Grow(1)
					button(c.Key("refresh"), "刷新", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("refresh-cw"), Label: "刷新" + title, Loading: busy, Disabled: s.api == nil}).Height(30).Padding(0, 4).OnClick(func() {
						if s.favoriteTab == "collection" {
							s.loadCollections(1, false)
						} else {
							s.loadFavorites()
						}
					})
				})
				layout.ScrollArea(c.Key("folders-scroll"), s.scroll("folders/"+s.favoriteTab), layout.ScrollAreaOptions{}, func() {
					ui.Column(c).FillWidth().Padding(style.SpaceSmall).Gap(4).Children(func() { s.libraryFolders(c) })
				}).Grow(1).MinHeight(0)
				s.maybeAutoLibraryFolders()
			})
			ui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Children(func() {
				s.libraryHeader(c, navigate)
				key := s.libraryScrollKey()
				layout.ScrollArea(c.Key(key), s.scroll(key), layout.ScrollAreaOptions{}, func() {
					ui.Column(c).FillWidth().Padding(style.SpaceMedium, style.ScrollbarGutter, style.SpaceMedium, style.SpaceMedium).
						Gap(style.SpaceMedium).Children(func() { s.libraryVideos(c, navigate) })
				}).Grow(1).MinHeight(0)
				s.maybeAutoLibraryVideos()
			})
		})
	}).Height(max(400, height-style.TitlebarHeight-2*style.PagePadding-2*style.BorderWidth))
}

func (s *State) libraryHeader(c *ui.Context, navigate func(string)) {
	k := core.Tokens(c)
	width, _ := c.Size()
	compact := width < style.AuthorControlsDownloadBreakpoint
	title, total := s.libraryTitleAndTotal()
	loaded := s.currentVideoCount()
	selected := s.selectedCount()
	keys := s.libraryVideoKeys()
	allSelected := len(keys) > 0 && selected == len(keys)
	actions := func() {
		label := "全选"
		if allSelected {
			label = "取消全选"
		}
		button(c.Key("select"), label, input.ButtonOptions{Variant: input.Ghost, Disabled: len(keys) == 0}).Height(32).Padding(0, 6).FontSize(style.Font(c, style.FontMeta)).OnClick(func() {
			for _, key := range keys {
				*s.selection(key) = !allSelected
			}
		})
		button(c.Key("selected"), fmt.Sprintf("下载已选 (%d)", selected), input.ButtonOptions{Variant: input.Outline, Disabled: selected == 0}).Height(32).Padding(0, 7).FontSize(style.Font(c, style.FontMeta)).OnClick(func() {
			s.stageSelectedLibrary()
			for _, key := range keys {
				*s.selection(key) = false
			}
			navigate("/bilibili/download")
		})
		{
			allLabel := "一键下载全部"
			if compact {
				allLabel = "下载全部"
			}
			button(c.Key("all"), allLabel, input.ButtonOptions{Variant: input.Primary, Label: "一键下载全部", Disabled: s.api == nil || len(keys) == 0}).Height(32).Padding(0, 7).FontSize(style.Font(c, style.FontMeta)).OnClick(func() { s.stageAllLibrary(navigate) })
		}
	}
	ui.Row(c).Label("视频库内容标题").FillWidth().Height(style.LibraryHeaderHeight).Padding(style.SpaceSmall, style.SpaceMedium).Gap(style.SpaceSmall).
		AlignItems(ui.Center).BorderColor(k.Border).BorderWidth(0, 0, style.BorderWidth, 0).Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, title).MinWidth(0).SingleLine().FontWeight(style.WeightSemibold)
			if total > 0 || loaded > 0 {
				countText := fmt.Sprintf("%d 个视频（已加载 %d）", max(total, loaded), loaded)
				if compact {
					countText = fmt.Sprintf("%d 个 · 已加载 %d", max(total, loaded), loaded)
				}
				ui.Text(c, countText).Shrink(0).SingleLine().FontSize(style.Font(c, style.FontMeta)).TextColor(k.Warning)
			}
		})
		actions()
	})
}

func (s *State) libraryTitleAndTotal() (string, int) {
	if s.favoriteTab == "collection" {
		for _, item := range s.collections.Data.List {
			if item.ID == s.collectionID {
				return item.Title, item.MediaCount
			}
		}
		return "合集", s.collectionTotal()
	}
	for _, item := range s.favoritesData.Data.List {
		if item.ID == s.favoriteID {
			return item.Title, int(item.MediaCount)
		}
	}
	return "收藏夹", s.favoriteDetail.Data.Info.MediaCount
}

func (s *State) libraryVideoKeys() []string {
	keys := make([]string, 0, s.currentVideoCount())
	if s.favoriteTab == "collection" {
		for _, video := range s.collectionDetail.Data.Medias {
			if strings.HasPrefix(video.Bvid, "BV") {
				keys = append(keys, video.Bvid)
			}
		}
	} else {
		for _, video := range s.favoriteDetail.Data.Medias {
			bvid := favoriteBvid(video.Bvid, video.BvID)
			if video.Type == 2 && video.Attr != 1 && video.Attr != 9 && strings.HasPrefix(bvid, "BV") {
				keys = append(keys, bvid)
			}
		}
	}
	return keys
}

func (s *State) libraryFolders(c *ui.Context) {
	if s.favoriteTab == "collection" {
		if s.collections.Loading {
			caption(c, "正在加载合集列表")
		}
		if len(s.collections.Data.List) == 0 {
			caption(c, "暂无合集")
		}
		for _, item := range s.collections.Data.List {
			libraryFolder(c.Key(item.ID), item.Title, "合集", item.MediaCount, s.collectionID == item.ID, func() { s.loadCollection(item.ID, 1, false) })
		}
		s.errorRow(c, s.collections.Error, func() {
			if s.collections.Loaded {
				s.loadCollections(s.collectionPage+1, true)
			} else {
				s.loadCollections(1, false)
			}
		})
		if s.collections.Data.HasMore {
			const key = "folders/collection"
			if s.autoPageFallback(key, key) && s.collections.Error == "" && !s.collections.Loading {
				button(c, "继续加载", input.ButtonOptions{Variant: input.Ghost}).OnClick(func() {
					s.continueAutoPage(key, func() { s.loadCollections(s.collectionPage+1, true) })
				})
			}
		}
	} else {
		s.errorRow(c, s.favoritesData.Error, s.loadFavorites)
		if len(s.favoritesData.Data.List) == 0 {
			caption(c, "暂无收藏夹")
		}
		for _, item := range s.favoritesData.Data.List {
			libraryFolder(c.Key(item.ID), item.Title, "收藏夹", int(item.MediaCount), s.favoriteID == item.ID, func() { s.loadFavorite(item.ID, 1, false) })
		}
	}
}

func libraryFolder(c *ui.Context, title, kind string, count int, selected bool, onClick func()) {
	k := core.Tokens(c)
	background, border := k.Surface, k.Border.Alpha(0)
	iconBackground, iconBorder, iconColor := k.Background, k.Border.Alpha(.7), k.Warning.Alpha(.6)
	badgeBackground, badgeColor := k.Background, k.TextMuted
	entry := ui.ButtonBase(c).Label("选择"+kind+title).FillWidth().MinWidth(0).
		Height(style.LibraryListItemHeight).Padding(0, style.SpaceSmall).Gap(style.SpaceSmall).
		Radius(style.RadiusMenu).Background(background).TextColor(k.Text).Clip()
	if selected {
		background, border = k.Selection, k.Accent.Alpha(.45)
		iconBackground, iconBorder, iconColor = k.Accent.Alpha(.14), k.Accent.Alpha(.45), k.Accent
		badgeBackground, badgeColor = k.Accent.Alpha(.14), k.Accent
	} else if entry.Hovered() || entry.FocusVisible() {
		background, border = k.SurfaceHover, k.Border.Alpha(.65)
	}
	entry.Background(background).Border(style.BorderWidth, border).Children(func() {
		iconName := "star"
		if kind == "合集" {
			iconName = "folder"
		}
		ui.Box(c).Size(style.LibraryListIconSize, style.LibraryListIconSize).Shrink(0).
			Radius(style.RadiusControl).Border(style.BorderWidth, iconBorder).Background(iconBackground).Center().Children(func() {
			ui.Icon(c, icons.Must(iconName)).Size(16, 16).TextColor(iconColor)
		})
		text := ui.Text(c, title).Grow(1).MinWidth(0).SingleLine().FontWeight(style.WeightSemibold)
		if selected {
			text.TextColor(k.Accent)
		}
		ui.Box(c).MinWidth(28).Height(23).Shrink(0).Padding(0, 6).Radius(12).Background(badgeBackground).Center().Children(func() {
			ui.Text(c, fmt.Sprint(count)).SingleLine().FontSize(style.Font(c, style.FontMeta)).FontWeight(style.WeightSemibold).TextColor(badgeColor)
		})
	}).OnClick(onClick)
}

func (s *State) libraryVideos(c *ui.Context, navigate func(string)) {
	s.stagingNotice(c)
	s.errorRow(c, s.stageError, nil)
	if s.favoriteTab == "collection" {
		if s.collectionDetail.Loading {
			caption(c, "正在加载合集视频")
		}
		if s.collectionID == 0 {
			empty(c, "请选择合集", "star")
			return
		}
		ui.Grid(c).Columns(s.columns(c, true)).FillWidth().Gap(style.SpaceMedium).Children(func() {
			for _, video := range s.collectionDetail.Data.Medias {
				s.mediaCardWithAuthor(c.Key(video.ID), video.Bvid, video.Title, video.Cover, video.Upper.Name, duration(video.Duration), video.CntInfo.Play, video.CntInfo.Danmaku, video.Pubtime, uint64(max(0, video.Upper.Mid)), navigate)
			}
		})
		s.errorRow(c, s.collectionDetail.Error, func() { s.loadCollection(s.collectionID, s.collectionDetailPage+1, s.collectionDetail.Loaded) })
		if len(s.collectionDetail.Data.Medias) == 0 && !s.collectionDetail.Loading {
			empty(c, "暂无视频", "play")
		}
		if s.collectionHasMore {
			key := s.libraryScrollKey()
			if s.autoPageFallback(key, key) && s.collectionDetail.Error == "" && !s.collectionDetail.Loading {
				button(c, "继续加载", input.ButtonOptions{Variant: input.Outline}).OnClick(func() {
					s.continueAutoPage(key, func() { s.loadCollection(s.collectionID, s.collectionDetailPage+1, true) })
				})
			}
		}
	} else {
		if s.favoriteDetail.Loading {
			caption(c, "正在加载收藏视频")
		}
		if s.favoriteID == 0 {
			empty(c, "请选择收藏夹", "star")
			return
		}
		ui.Grid(c).Columns(s.columns(c, true)).FillWidth().Gap(style.SpaceMedium).Children(func() {
			for _, video := range s.favoriteDetail.Data.Medias {
				bvid := favoriteBvid(video.Bvid, video.BvID)
				if video.Type != 2 || video.Attr == 1 || video.Attr == 9 {
					bvid = ""
				}
				s.mediaCardWithAuthor(c.Key(video.ID), bvid, video.Title, video.Cover, video.Upper.Name, duration(video.Duration), video.CntInfo.Play, video.CntInfo.Danmaku, video.PubTime, uint64(max(int64(0), video.Upper.Mid)), navigate)
			}
		})
		s.errorRow(c, s.favoriteDetail.Error, func() { s.loadFavorite(s.favoriteID, s.favoritePage+1, s.favoriteDetail.Loaded) })
		if len(s.favoriteDetail.Data.Medias) == 0 && !s.favoriteDetail.Loading {
			empty(c, "暂无视频", "play")
		}
		if s.favoriteDetail.Data.HasMore {
			key := s.libraryScrollKey()
			if s.autoPageFallback(key, key) && s.favoriteDetail.Error == "" && !s.favoriteDetail.Loading {
				button(c, "继续加载", input.ButtonOptions{Variant: input.Outline}).OnClick(func() {
					s.continueAutoPage(key, func() { s.loadFavorite(s.favoriteID, s.favoritePage+1, true) })
				})
			}
		}
	}
}
func favoriteBvid(bvid, legacy string) string {
	if bvid != "" {
		return bvid
	}
	return legacy
}
