package bilibili

import (
	"fmt"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// A refresh owns a new generation so a late pagination response cannot add
// old videos back to the refreshed feed.
func (s *State) loadDynamics(more bool) {
	if !s.loggedIn || s.api == nil {
		return
	}
	if more {
		if s.dynamicMore.Loading || s.dynamics.Loading || !s.dynamicHasMore || s.dynamicOffset == "" {
			return
		}
		s.dynamicAutoAtY = s.scroll("dynamic-videos").Y
		offset, generation := s.dynamicOffset, s.dynamicGeneration
		load(s, &s.dynamicMore, func() (model.DynamicArchivePage, error) {
			return s.api.DynamicVideosPage(offset)
		}, func(page model.DynamicArchivePage) {
			if generation != s.dynamicGeneration {
				return
			}
			s.dynamics.Data.Items = appendUnique(s.dynamics.Data.Items, page.Items, func(item model.DynamicArchiveItem) string { return item.Bvid })
			s.dynamicOffset = page.Offset
			s.dynamicHasMore = page.HasMore && page.Offset != "" && page.Offset != offset
		})
		return
	}
	s.dynamicGeneration++
	s.dynamicMore.seq++
	s.dynamicMore.Loading = false
	s.dynamicMore.Error = ""
	s.dynamics.Data = model.DynamicArchivePage{}
	s.dynamics.Loaded = false
	s.dynamicOffset = ""
	s.dynamicHasMore = false
	s.dynamicSelected = make(map[string]bool)
	s.dynamicAutoPending = false
	s.dynamicAutoAtY = 0
	s.dynamicAutoFillPages = 0
	s.scroll("dynamic-videos").Y = 0
	s.scroll("dynamic-videos").MaxY = 0
	load(s, &s.dynamics, func() (model.DynamicArchivePage, error) {
		return s.api.DynamicVideosPage("")
	}, func(page model.DynamicArchivePage) {
		s.dynamicOffset = page.Offset
		s.dynamicHasMore = page.HasMore && page.Offset != ""
	})
}

func dynamicTime(item model.DynamicArchiveItem) string {
	if item.PubTs > 0 {
		return time.Unix(item.PubTs, 0).Format("2006-01-02 15:04")
	}
	return strings.TrimSpace(item.PubTime)
}

func (s *State) dynamicSelectedCount() int {
	count := 0
	for _, item := range s.dynamics.Data.Items {
		if s.dynamicSelected[item.Bvid] {
			count++
		}
	}
	return count
}

func (s *State) stageSelectedDynamics(navigate func(string)) {
	if s.dynamicSelectedCount() == 0 {
		return
	}
	for _, item := range s.dynamics.Data.Items {
		if s.dynamicSelected[item.Bvid] {
			s.stageDynamic(item)
		}
	}
	s.dynamicSelected = make(map[string]bool)
	navigate("/bilibili/download")
}

// Schedule one page after a scroll approaches the end. A short first page
// fills automatically too; the limit keeps empty feeds from polling forever.
func (s *State) maybeAutoLoadDynamics() {
	state := s.scroll("dynamic-videos")
	if s.runtime.After == nil || s.dynamicAutoPending || s.dynamicMore.Loading || s.dynamicMore.Error != "" ||
		!s.dynamicHasMore || s.dynamicOffset == "" || !s.dynamicShouldAutoLoad(state) {
		return
	}
	s.dynamicAutoPending = true
	epoch, generation, offset := s.epoch, s.dynamicGeneration, s.dynamicOffset
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || s.dynamicGeneration != generation {
			return
		}
		s.dynamicAutoPending = false
		state := s.scroll("dynamic-videos")
		if s.section != "dynamic" || s.dynamicOffset != offset || s.dynamicMore.Error != "" ||
			!s.dynamicShouldAutoLoad(state) {
			return
		}
		if state.MaxY <= 0 {
			s.dynamicAutoFillPages++
		} else {
			s.dynamicAutoFillPages = 0
		}
		s.loadDynamics(true)
	})
}

func (s *State) dynamicShouldAutoLoad(state *ui.ScrollState) bool {
	if state.MaxY <= 0 {
		return s.dynamicAutoFillPages < style.DynamicAutoFillLimit
	}
	return state.Y > s.dynamicAutoAtY+1 && state.Y+style.DynamicAutoLoadThreshold >= state.MaxY
}

func (s *State) dynamic(c *ui.Context, navigate func(string)) {
	if s.requiresLogin(c, navigate) {
		return
	}
	k := core.Tokens(c)
	items := s.dynamics.Data.Items
	selected := s.dynamicSelectedCount()
	_, height := c.Size()
	panel(c, "动态工具栏", func() {
		ui.Row(c).FillWidth().Height(style.LibraryHeaderHeight).Padding(style.SpaceSmall, style.SpaceMedium).
			Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "关注动态").Shrink(0).SingleLine().FontWeight(style.WeightSemibold)
			ui.Text(c, fmt.Sprintf("已加载 %d 个 · 已选 %d 个", len(items), selected)).Grow(1).MinWidth(0).SingleLine().
				FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
			button(c.Key("refresh"), "刷新", input.ButtonOptions{Variant: input.Outline, Icon: icons.Must("refresh-cw"), Loading: s.dynamics.Loading}).
				Height(32).OnClick(func() { s.loadDynamics(false) })
			if selected > 0 {
				button(c.Key("clear-selection"), "清空选择", input.ButtonOptions{Variant: input.Ghost}).
					Height(32).OnClick(func() { s.dynamicSelected = make(map[string]bool) })
			}
			button(c.Key("selected-download"), fmt.Sprintf("下载已选 (%d)", selected), input.ButtonOptions{Variant: input.Primary, Disabled: selected == 0}).
				Height(32).OnClick(func() { s.stageSelectedDynamics(navigate) })
		})
	})
	panel(c, "动态视频列表", func() {
		if s.dynamics.Loading && !s.dynamics.Loaded {
			empty(c, "正在加载关注动态", "loader-circle")
			return
		}
		if s.dynamics.Error != "" {
			ui.Column(c).FillWidth().Padding(style.SpaceLarge).Children(func() {
				s.errorRow(c, s.dynamics.Error, func() { s.loadDynamics(false) })
			})
			return
		}
		if len(items) == 0 {
			s.maybeAutoLoadDynamics()
			if s.dynamicMore.Error != "" {
				empty(c, "加载关注视频动态失败", "circle-alert")
			} else if s.dynamicHasMore || s.dynamicMore.Loading {
				empty(c, "正在查找关注视频动态", "loader-circle")
			} else {
				empty(c, "暂无关注视频动态", "play")
			}
			if s.dynamicMore.Error != "" {
				ui.Column(c).FillWidth().Padding(style.SpaceMedium).Children(func() {
					s.errorRow(c, s.dynamicMore.Error, func() { s.loadDynamics(true) })
				})
			} else if s.dynamicHasMore && (s.runtime.After == nil || s.dynamicAutoFillPages >= style.DynamicAutoFillLimit) {
				ui.Row(c).FillWidth().Justify(ui.Center).Padding(style.SpaceMedium).Children(func() {
					button(c.Key("more"), "继续加载", input.ButtonOptions{Variant: input.Outline}).
						OnClick(func() { s.dynamicAutoFillPages = 0; s.loadDynamics(true) })
				})
			}
			return
		}
		ui.Column(c).FillWidth().Grow(1).MinHeight(0).Children(func() {
			s.dynamicColumns(c)
			layout.ScrollArea(c.Key("dynamic-scroll"), s.scroll("dynamic-videos"), layout.ScrollAreaOptions{}, func() {
				ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, 0, 0).Children(func() {
					for _, item := range items {
						s.dynamicRow(c.Key(item.Bvid), item, navigate)
					}
				})
			}).Grow(1).MinHeight(0)
			s.maybeAutoLoadDynamics()
			shortListFallback := s.dynamicHasMore && (s.runtime.After == nil ||
				s.scroll("dynamic-videos").MaxY <= 0 && s.dynamicAutoFillPages >= style.DynamicAutoFillLimit)
			if shortListFallback || s.dynamicMore.Loading || s.dynamicMore.Error != "" {
				ui.Column(c.Key("dynamic-more")).Label("动态加载状态").FillWidth().Padding(style.SpaceSmall, style.SpaceMedium).
					Gap(style.SpaceSmall).BorderColor(k.Border).BorderWidth(style.BorderWidth, 0, 0, 0).
					AlignItems(ui.Center).Children(func() {
					if s.dynamicMore.Error != "" {
						s.errorRow(c, s.dynamicMore.Error, func() { s.loadDynamics(true) })
					}
					if s.dynamicMore.Loading {
						caption(c, "正在加载更多")
					} else if shortListFallback && s.dynamicMore.Error == "" {
						button(c.Key("more"), "继续加载", input.ButtonOptions{Variant: input.Outline}).
							Height(32).OnClick(func() {
							s.dynamicAutoFillPages = 0
							s.loadDynamics(true)
						})
					}
				})
			}
		})
	}).Height(max(360, height-style.TitlebarHeight-2*style.PagePadding-style.LibraryHeaderHeight-style.SpaceMedium-2*style.BorderWidth))
}

func (s *State) dynamicColumns(c *ui.Context) {
	k := core.Tokens(c)
	ui.Row(c).FillWidth().Height(style.DynamicColumnsHeight).
		Padding(0, style.ScrollbarGutter+style.SpaceMedium, 0, style.SpaceMedium).Gap(style.DynamicColumnGap).
		AlignItems(ui.Center).Background(k.Background).BorderColor(k.Border).BorderWidth(0, 0, style.BorderWidth, 0).
		FontSize(style.Font(c, style.FontMeta)).FontWeight(style.WeightSemibold).TextColor(k.TextMuted).
		Children(func() {
			ui.Text(c, "视频").Grow(1)
			ui.Text(c, "作者").Width(style.DynamicAuthorWidth).Shrink(0)
			ui.Text(c, "时间").Width(style.DynamicTimeWidth).Shrink(0)
			ui.Text(c, "操作").Width(style.DynamicActionWidth).Shrink(0).TextAlign(ui.End)
		})
}

func (s *State) dynamicRow(c *ui.Context, item model.DynamicArchiveItem, navigate func(string)) {
	k := core.Tokens(c)
	title := item.Title
	if title == "" {
		title = "未命名视频"
	}
	valid := strings.HasPrefix(item.Bvid, "BV")
	selected := valid && s.dynamicSelected[item.Bvid]
	label := "选择动态视频：" + title
	if selected {
		label = "取消选择动态视频：" + title
	}
	row := ui.ButtonBase(c).Label(label).FillWidth().MinHeight(style.DynamicRowHeight).
		Padding(style.SpaceSmall, style.SpaceMedium).Gap(style.DynamicColumnGap).AlignItems(ui.Center).
		BorderColor(k.Border.Alpha(.5)).BorderWidth(0, 0, style.BorderWidth, 0).Cursor(ui.CursorPointer).
		Disabled(!valid && item.Mid <= 0)
	authorHovered, downloadHovered := false, false
	row.Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).Gap(style.SpaceMedium).AlignItems(ui.Center).Children(func() {
			length := item.DurationText
			if length == "" {
				length = "--:--"
			}
			ui.Box(c).Size(style.DynamicThumbWidth, style.DynamicThumbHeight).Shrink(0).Children(func() {
				s.mediaThumbnail(c, item.Cover, length, 0, 0)
			})
			ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
				ui.Text(c, title).FillWidth().MaxLines(2).FontWeight(style.WeightSemibold)
				ui.Text(c, item.Bvid).FillWidth().SingleLine().FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
			})
		})
		author := item.AuthorName
		if author == "" {
			author = "未知 UP"
		}
		if item.Mid > 0 {
			ui.Row(c).Width(style.DynamicAuthorWidth).Shrink(0).MinWidth(0).Clip().Children(func() {
				link := ui.ButtonBase(c.Key("author")).Label("进入 UP 主主页："+author).
					MinWidth(0).MaxWidthPercent(100).Clip().Padding(0, 0, style.MediaAuthorUnderlineGap, 0).
					BorderWidth(0, 0, style.MediaAuthorUnderlineWidth, 0).BorderColor(ui.Transparent).
					Background(ui.Transparent).TextColor(k.TextMuted).Cursor(ui.CursorPointer)
				if link.Hovered() || link.FocusVisible() {
					authorHovered = true
					link.TextColor(k.AccentText).BorderColor(k.Accent)
				}
				link.Children(func() { ui.Text(c, author).SingleLine().MaxWidthPercent(100) }).
					OnClick(func() { navigate(navigation.BilibiliAuthorPath(uint64(item.Mid))) })
			})
		} else {
			ui.Text(c, author).Width(style.DynamicAuthorWidth).Shrink(0).SingleLine().TextColor(k.TextMuted)
		}
		ui.Text(c, dynamicTime(item)).Width(style.DynamicTimeWidth).Shrink(0).SingleLine().
			FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
		ui.Row(c).Width(style.DynamicActionWidth).Shrink(0).Justify(ui.End).Children(func() {
			one := button(c.Key("download"), "下载", input.ButtonOptions{Variant: input.Primary, Label: "下载视频：" + title, Disabled: !valid}).Height(30)
			downloadHovered = one.Hovered() || one.FocusVisible()
			one.OnClick(func() {
				s.stageDynamic(item)
				navigate("/bilibili/download")
			})
		})
	})
	if selected {
		row.Background(k.Selection).BorderColor(k.Accent).BorderWidth(0, 0, style.BorderEmphasis, 0)
	} else if valid && !authorHovered && !downloadHovered && (row.Hovered() || row.FocusVisible()) {
		row.Background(k.SurfaceHover)
	}
	row.OnClick(func() {
		if valid {
			s.dynamicSelected[item.Bvid] = !s.dynamicSelected[item.Bvid]
		}
	})
}
