package bilibili

import (
	"fmt"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

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
	if !s.loggedIn || s.api == nil {
		return
	}
	if more {
		if s.playHistoryMore.Loading || s.playHistory.Loading || !s.playHistoryHasMore {
			return
		}
		s.playHistoryAutoAtY = s.scroll("play-history-videos").Y
		cursor, generation := s.playHistoryCursor, s.playHistoryGeneration
		load(s, &s.playHistoryMore, func() (model.PlayHistoryData, error) {
			return s.api.PlayHistory(cursor.Max, cursor.ViewAt)
		}, func(page model.PlayHistoryData) {
			if generation != s.playHistoryGeneration {
				return
			}
			s.playHistory.Data.List = appendUnique(s.playHistory.Data.List, page.List, playHistoryKey)
			s.playHistoryCursor = page.Cursor
			s.playHistoryHasMore = playHistoryCursorChanged(cursor, page.Cursor) && len(page.List) > 0
		})
		return
	}
	s.playHistoryGeneration++
	s.playHistoryMore.seq++
	s.playHistoryMore.Loading = false
	s.playHistoryMore.Error = ""
	s.playHistory.Data = model.PlayHistoryData{}
	s.playHistory.Loaded = false
	s.playHistory.Loading = false
	s.playHistory.Error = ""
	s.playHistoryCursor = model.PlayHistoryCursor{}
	s.playHistoryHasMore = false
	s.playHistorySelected = make(map[string]*bool)
	s.playHistoryAutoPending = false
	s.playHistoryAutoAtY = 0
	s.playHistoryAutoFillPages = 0
	s.scroll("play-history-videos").Y = 0
	s.scroll("play-history-videos").MaxY = 0
	load(s, &s.playHistory, func() (model.PlayHistoryData, error) {
		return s.api.PlayHistory(0, 0)
	}, func(page model.PlayHistoryData) {
		s.playHistoryCursor = page.Cursor
		s.playHistoryHasMore = len(page.List) > 0 && playHistoryCursorChanged(model.PlayHistoryCursor{}, page.Cursor)
	})
}

func playHistoryCursorChanged(a, b model.PlayHistoryCursor) bool {
	return a.Max != b.Max || a.ViewAt != b.ViewAt
}

func (s *State) playHistorySelectedCount() int {
	count := 0
	for _, item := range s.playHistory.Data.List {
		if selected := s.playHistorySelected[item.History.Bvid]; selected != nil && *selected {
			count++
		}
	}
	return count
}

func (s *State) playHistorySelection(bvid string) *bool {
	if s.playHistorySelected[bvid] == nil {
		s.playHistorySelected[bvid] = new(bool)
	}
	return s.playHistorySelected[bvid]
}

func (s *State) stageSelectedPlayHistory(navigate func(string)) {
	if s.playHistorySelectedCount() == 0 {
		return
	}
	for _, item := range s.playHistory.Data.List {
		if selected := s.playHistorySelected[item.History.Bvid]; selected != nil && *selected {
			s.stagePlayHistory(item)
		}
	}
	s.playHistorySelected = make(map[string]*bool)
	navigate("/bilibili/download")
}

func (s *State) maybeAutoLoadPlayHistory() {
	state := s.scroll("play-history-videos")
	if s.runtime.After == nil || s.playHistoryAutoPending || s.playHistoryMore.Loading || s.playHistoryMore.Error != "" ||
		!s.playHistoryHasMore || !s.playHistoryAutoReady(state) {
		return
	}
	s.playHistoryAutoPending = true
	epoch, generation, cursor := s.epoch, s.playHistoryGeneration, s.playHistoryCursor
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || s.playHistoryGeneration != generation {
			return
		}
		s.playHistoryAutoPending = false
		state := s.scroll("play-history-videos")
		if s.section != "play-history" || s.playHistoryCursor != cursor || s.playHistoryMore.Error != "" || !s.playHistoryAutoReady(state) {
			return
		}
		if state.MaxY <= 0 {
			s.playHistoryAutoFillPages++
		} else {
			s.playHistoryAutoFillPages = 0
		}
		s.loadPlayHistory(true)
	})
}

func (s *State) playHistoryAutoReady(state *ui.ScrollState) bool {
	if state.MaxY <= 0 {
		return s.playHistoryAutoFillPages < style.DynamicAutoFillLimit
	}
	return state.Y > s.playHistoryAutoAtY+1 && state.Y+style.DynamicAutoLoadThreshold >= state.MaxY
}

func (s *State) playHistoryPage(c *ui.Context, navigate func(string)) {
	if s.requiresLogin(c, navigate) {
		return
	}
	k := core.Tokens(c)
	items := s.playHistory.Data.List
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
			button(c.Key("refresh"), "刷新", input.ButtonOptions{Variant: input.Outline, Icon: icons.Must("refresh-cw"), Loading: s.playHistory.Loading}).
				Height(32).OnClick(func() { s.loadPlayHistory(false) })
			allSelected := selectable > 0 && selected == selectable
			selectLabel := "全选"
			if allSelected {
				selectLabel = "取消全选"
			}
			button(c.Key("select-all"), selectLabel, input.ButtonOptions{Variant: input.Ghost, Disabled: selectable == 0}).
				Height(32).OnClick(func() {
				for _, item := range items {
					if strings.HasPrefix(item.History.Bvid, "BV") {
						*s.playHistorySelection(item.History.Bvid) = !allSelected
					}
				}
			})
			button(c.Key("selected-download"), fmt.Sprintf("下载已选 (%d)", selected), input.ButtonOptions{Variant: input.Primary, Disabled: selected == 0}).
				Height(32).OnClick(func() { s.stageSelectedPlayHistory(navigate) })
		})
		if s.playHistory.Loading && !s.playHistory.Loaded {
			empty(c, "正在加载播放历史", "loader-circle")
			return
		}
		if s.playHistory.Error != "" {
			ui.Column(c).FillWidth().Padding(style.SpaceLarge).Children(func() {
				s.errorRow(c, s.playHistory.Error, func() { s.loadPlayHistory(false) })
			})
			return
		}
		if len(items) == 0 {
			empty(c, "暂无播放历史", "history")
			return
		}
		layout.ScrollArea(c.Key("play-history-scroll"), s.scroll("play-history-videos"), layout.ScrollAreaOptions{}, func() {
			ui.Column(c).FillWidth().Padding(style.SpaceMedium, style.ScrollbarGutter, style.SpaceMedium, style.SpaceMedium).
				Gap(style.SpaceMedium).Children(func() {
				ui.Grid(c).Columns(s.columns(c, false)).FillWidth().Gap(style.SpaceMedium).Children(func() {
					for _, item := range items {
						s.mediaCardWithSelection(c.Key(playHistoryKey(item)), item.History.Bvid, playHistoryTitle(item), item.Cover,
							item.AuthorName, playHistoryDuration(item), 0, 0, item.ViewAt, uint64(max(int64(0), item.AuthorMid)), navigate,
							s.playHistorySelection(item.History.Bvid))
					}
				})
				if s.playHistoryMore.Loading {
					caption(c, "正在加载更多")
				}
				if s.playHistoryMore.Error != "" {
					s.errorRow(c, s.playHistoryMore.Error, func() { s.loadPlayHistory(true) })
				}
				shortListFallback := s.playHistoryHasMore && (s.runtime.After == nil ||
					s.scroll("play-history-videos").MaxY <= 0 && s.playHistoryAutoFillPages >= style.DynamicAutoFillLimit)
				if shortListFallback && s.playHistoryMore.Error == "" && !s.playHistoryMore.Loading {
					button(c.Key("more"), "继续加载", input.ButtonOptions{Variant: input.Outline}).OnClick(func() {
						s.playHistoryAutoFillPages = 0
						s.loadPlayHistory(true)
					})
				}
			})
		}).Grow(1).MinHeight(0)
		s.maybeAutoLoadPlayHistory()
	}).Height(max(400, height-style.TitlebarHeight-2*style.PagePadding-2*style.BorderWidth))
}
