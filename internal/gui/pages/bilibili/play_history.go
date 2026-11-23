package bilibili

import (
	"fmt"
	"github.com/kamiertop/videodown/internal/gui/components"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// playState 是播放历史板块：内嵌共用的翻页 feed，另带接口的 max/view_at
// 游标与按 BV 号的勾选集。
type playState struct {
	feed[model.PlayHistoryData]
	cursor model.PlayHistoryCursor
	picked map[string]*bool
}

func playHistoryTitle(item model.PlayHistoryItem) string {
	if title := strings.TrimSpace(item.Title); title != "" {
		return title
	}
	if title := strings.TrimSpace(item.LongTitle); title != "" {
		return title
	}
	return "未命名视频"
}

func playHistoryDuration(item model.PlayHistoryItem) string {
	if item.Duration <= 0 {
		return ""
	}
	return duration(item.Duration)
}

func playHistoryKey(item model.PlayHistoryItem) string {
	if item.History.Bvid != "" {
		return item.History.Bvid
	}
	return fmt.Sprintf("%s:%d:%d", item.History.Business, item.History.Oid, item.History.Epid)
}

func (s *State) loadPlayHistory(more bool) {
	if !s.auth.loggedIn || s.api == nil {
		return
	}
	if more {
		if s.play.more.Loading || s.play.items.Loading || !s.play.hasMore {
			return
		}
		s.play.autoAtY = s.scroll("play-history-videos").Y
		cursor, generation := s.play.cursor, s.play.generation
		load(s, &s.play.more, func() (model.PlayHistoryData, error) {
			return s.api.PlayHistory(cursor.Max, cursor.ViewAt)
		}, func(page model.PlayHistoryData) {
			if generation != s.play.generation {
				return
			}
			s.play.items.Data.List = appendUnique(s.play.items.Data.List, page.List, playHistoryKey)
			s.play.cursor = page.Cursor
			s.play.hasMore = playHistoryCursorChanged(cursor, page.Cursor) && len(page.List) > 0
		})
		return
	}
	s.play.feed.reset()
	s.play.cursor = model.PlayHistoryCursor{}
	s.play.picked = make(map[string]*bool)
	scroll := s.scroll("play-history-videos")
	scroll.Y, scroll.MaxY = 0, 0
	load(s, &s.play.items, func() (model.PlayHistoryData, error) {
		return s.api.PlayHistory(0, 0)
	}, func(page model.PlayHistoryData) {
		s.play.cursor = page.Cursor
		s.play.hasMore = len(page.List) > 0 && playHistoryCursorChanged(model.PlayHistoryCursor{}, page.Cursor)
	})
}

func playHistoryCursorChanged(a, b model.PlayHistoryCursor) bool {
	return a.Max != b.Max || a.ViewAt != b.ViewAt
}

func (s *State) playHistorySelectedCount() int {
	count := 0
	for _, item := range s.play.items.Data.List {
		if selected := s.play.picked[item.History.Bvid]; selected != nil && *selected {
			count++
		}
	}
	return count
}

func (s *State) playHistorySelection(bvid string) *bool {
	if s.play.picked[bvid] == nil {
		s.play.picked[bvid] = new(bool)
	}
	return s.play.picked[bvid]
}

func (s *State) stageSelectedPlayHistory(navigate func(string)) {
	if s.playHistorySelectedCount() == 0 {
		return
	}
	for _, item := range s.play.items.Data.List {
		if selected := s.play.picked[item.History.Bvid]; selected != nil && *selected {
			s.stagePlayHistory(item)
		}
	}
	s.play.picked = make(map[string]*bool)
	navigate("/bilibili/download")
}

func (s *State) maybeAutoLoadPlayHistory() {
	cursor := s.play.cursor
	maybeAutoLoadFeed(s, &s.play.feed, "play-history-videos", "play-history",
		nil,
		func() bool { return s.play.cursor == cursor },
		func() { s.loadPlayHistory(true) })
}

func (s *State) playHistoryPage(c *ui.Context, navigate func(string)) {
	if s.requiresLogin(c, navigate) {
		return
	}
	k := c.Theme()
	items := s.play.items.Data.List
	selected := s.playHistorySelectedCount()
	selectable := 0
	for _, item := range items {
		if strings.HasPrefix(item.History.Bvid, "BV") {
			selectable++
		}
	}
	_, height := c.Size()
	panel(c, "播放历史记录", func() {
		ui.Row(c).Label("播放历史工具栏").FillWidth().Height(style.LibraryHeaderHeight).Padding(style.SpaceSmall, style.SpaceMedium).
			Gap(style.SpaceSmall).AlignItems(ui.Center).BorderColor(k.Border).BorderWidth(0, 0, style.BorderWidth, 0).Children(func() {
			ui.Text(c, "播放历史").Shrink(0).SingleLine().FontWeight(style.WeightSemibold)
			ui.Text(c, fmt.Sprintf("已加载 %d 个", len(items))).Grow(1).MinWidth(0).SingleLine().
				FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
			button(c.Key("refresh"), "刷新", components.ButtonOptions{Variant: components.ButtonOutline, Icon: icons.Must("refresh-cw"), Loading: s.play.items.Loading}).
				Height(32).OnClick(func() { s.loadPlayHistory(false) })
			allSelected := selectable > 0 && selected == selectable
			selectLabel := "全选"
			if allSelected {
				selectLabel = "取消全选"
			}
			button(c.Key("select-all"), selectLabel, components.ButtonOptions{Variant: components.ButtonGhost, Disabled: selectable == 0}).
				Height(32).OnClick(func() {
				for _, item := range items {
					if strings.HasPrefix(item.History.Bvid, "BV") {
						*s.playHistorySelection(item.History.Bvid) = !allSelected
					}
				}
			})
			button(c.Key("selected-download"), fmt.Sprintf("下载已选 (%d)", selected), components.ButtonOptions{Variant: components.ButtonPrimary, Disabled: selected == 0}).
				Height(32).OnClick(func() { s.stageSelectedPlayHistory(navigate) })
		})
		if s.play.items.Loading && !s.play.items.Loaded {
			empty(c, "正在加载播放历史", "loader-circle")
			return
		}
		if s.play.items.Error != "" {
			ui.Column(c).FillWidth().Padding(style.SpaceLarge).Children(func() {
				s.errorRow(c, s.play.items.Error, func() { s.loadPlayHistory(false) })
			})
			return
		}
		if len(items) == 0 {
			empty(c, "暂无播放历史", "history")
			return
		}
		components.ScrollArea(c.Key("play-history-scroll"), s.scroll("play-history-videos"), components.ScrollAreaOptions{}, func() {
			ui.Column(c).FillWidth().Padding(style.SpaceMedium, style.ScrollbarGutter, style.SpaceMedium, style.SpaceMedium).
				Gap(style.SpaceMedium).Children(func() {
				ui.Grid(c).Columns(s.columns(c, false)).FillWidth().Gap(style.SpaceMedium).Children(func() {
					for _, item := range items {
						s.mediaCardWithSelection(c.Key(playHistoryKey(item)), item.History.Bvid, playHistoryTitle(item), item.Cover,
							item.AuthorName, playHistoryDuration(item), 0, 0, item.ViewAt, uint64(max(int64(0), item.AuthorMid)), navigate,
							s.playHistorySelection(item.History.Bvid))
					}
				})
				if s.play.more.Loading {
					caption(c, "正在加载更多")
				}
				if s.play.more.Error != "" {
					s.errorRow(c, s.play.more.Error, func() { s.loadPlayHistory(true) })
				}
				shortListFallback := s.play.hasMore && (s.runtime.After == nil ||
					s.scroll("play-history-videos").MaxY <= 0 && s.play.autoFill >= style.DynamicAutoFillLimit)
				if shortListFallback && s.play.more.Error == "" && !s.play.more.Loading {
					button(c.Key("more"), "继续加载", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() {
						s.play.autoFill = 0
						s.loadPlayHistory(true)
					})
				}
			})
		}).Grow(1).MinHeight(0)
		s.maybeAutoLoadPlayHistory()
	}).Height(max(400, height-style.TitlebarHeight-2*style.PagePadding-2*style.BorderWidth))
}
