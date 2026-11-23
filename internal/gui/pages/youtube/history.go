package youtube

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/icons"

	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
)

// Concurrent jobs can publish history snapshots out of completion order.
// Merge by saved path so an older callback cannot hide another finished file.
func (s *State) mergeHistory(items []yt.History) {
	byPath := make(map[string]int, len(s.history))
	for i, item := range s.history {
		byPath[item.Path] = i
	}
	for _, item := range items {
		if i, ok := byPath[item.Path]; ok {
			s.history[i] = item
		} else {
			byPath[item.Path] = len(s.history)
			s.history = append(s.history, item)
		}
	}
	sort.SliceStable(s.history, func(i, j int) bool { return s.history[i].Downloaded > s.history[j].Downloaded })
}

func (s *State) historyPage(c *ui.Context, scroll *ui.ScrollState) {
	ui.Column(c).Label("YouTube 下载历史页面").Fill().MinWidth(0).MinHeight(0).Padding(style.PagePadding).Gap(style.SpaceMedium).Children(func() {
		k := c.Theme()
		items := s.filteredHistory()
		ui.Row(c).
			Label("YouTube 历史工具栏").
			FillWidth().
			MinWidth(0).
			Padding(style.SpaceSmall).
			Gap(style.SpaceSmall).
			AlignItems(ui.Center).
			Radius(style.RadiusPanel).
			Border(style.BorderWidth, k.Border).
			Background(k.Surface).
			Children(func() {
				ui.Text(c, fmt.Sprintf("下载历史 · %d", len(items))).FontWeight(style.WeightSemibold).Shrink(0)
				field := components.TextInput(c.Key("history-search"), &s.historySearch, components.TextInputOptions{Label: "搜索 YouTube 下载历史", Placeholder: "搜索标题、作者、视频 ID 或路径"})
				field.Grow(1).MinWidth(0)
				components.Button(c.Key("clear-search"), "清除搜索", components.ButtonOptions{Variant: components.ButtonGhost, Disabled: s.historySearch == ""}).Shrink(0).OnClick(func() { s.historySearch = "" })
			})
		if s.actionError != "" {
			ui.Text(c, s.actionError).FillWidth().FontSize(style.Font(c, style.FontMeta)).TextColor(k.Danger)
		}
		components.ScrollArea(c.Key("history-scroll"), scroll, components.ScrollAreaOptions{}, func() {
			ui.Column(c).FillWidth().MinWidth(0).Children(func() {
				if len(items) == 0 {
					message := "暂无下载记录"
					if strings.TrimSpace(s.historySearch) != "" {
						message = "没有匹配的下载历史"
					}
					components.Muted(c, message)
					return
				}
				k := c.Theme()
				ui.Column(c).
					FillWidth().
					MinWidth(0).
					Radius(style.RadiusPanel).
					Border(style.BorderWidth, k.Border).
					Background(k.Surface).
					Clip().
					Children(func() {
						for i, item := range items {
							row := ui.Row(c.Key(item.Path)).
								Label("下载记录：" + item.Path).
								FillWidth().
								MinWidth(0).
								MinHeight(76).
								Padding(style.SpaceMedium).
								Gap(style.SpaceMedium).
								AlignItems(ui.Center)
							if i < len(items)-1 {
								row.BorderWidth(0, 0, style.BorderWidth, 0).BorderColor(k.Border)
							}
							row.Children(func() {
								ui.Column(c).Grow(1).MinWidth(0).Gap(style.SpaceSmall).Children(func() {
									// 标题与操作同行，与哔哩哔哩、抖音的历史卡片一致。
									ui.Row(c).FillWidth().MinWidth(0).Gap(style.SpaceSmall).AlignItems(ui.Center).
										Children(func() {
											ui.Text(c, item.Title).
												Grow(1).
												MinWidth(0).
												MaxLines(1).
												FontWeight(style.WeightSemibold).
												Tooltip(item.Title)
											if item.Kind == "audio" {
												ui.Text(c, "音频").
													Shrink(0).
													FontSize(style.Font(c, style.FontMeta)).
													TextColor(k.TextMuted)
											}
											ui.Row(c.Key("actions")).Shrink(0).Gap(style.SpaceSmall).Children(func() {
												historyCardButton(c.Key("open"), "本地打开", components.ButtonOptions{
													Variant:  components.ButtonPrimary,
													Icon:     icons.Must("play"),
													Label:    "打开视频：" + item.Path,
													Disabled: s.runtime.OpenFile == nil || item.Path == "",
												}).Tooltip("使用默认播放器打开").OnClick(func() { s.openFile(item.Path, false) })
												historyCardButton(c.Key("directory"), "打开目录", components.ButtonOptions{
													Variant:  components.ButtonOutline,
													Icon:     icons.Must("folder"),
													Label:    "打开目录：" + item.Path,
													Disabled: s.runtime.OpenLocation == nil || item.Path == "",
												}).OnClick(func() { s.openFile(item.Path, true) })
												historyCardButton(c.Key("delete"), "删除记录", components.ButtonOptions{
													Variant:  components.ButtonGhost,
													Label:    "删除记录：" + item.Path,
													Disabled: s.api == nil || s.historyBusy,
												}).OnClick(func() { s.deleteHistory(item) })
											})
										})
									path := item.Path
									if path == "" {
										path = "未记录保存路径"
									}
									ui.Text(c, path).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted).Tooltip(path)
								})
							})
						}
					})
			})
		}).
			Grow(1).
			MinWidth(0).
			MinHeight(0)
	})
}

func (s *State) filteredHistory() []yt.History {
	query := strings.ToLower(strings.TrimSpace(s.historySearch))
	if query == "" {
		return s.history
	}
	items := make([]yt.History, 0, len(s.history))
	for _, item := range s.history {
		if strings.Contains(strings.ToLower(item.Title), query) || strings.Contains(strings.ToLower(item.Author), query) || strings.Contains(strings.ToLower(item.ID), query) || strings.Contains(strings.ToLower(item.Path), query) {
			items = append(items, item)
		}
	}

	return items
}

// reportActionError 把打开、删除等操作失败以 toast 弹出（右下角）；
// 没有 toast 通道的环境（测试、离屏渲染）退回页面内红字提示。
func (s *State) reportActionError(message string) {
	if s.runtime.Notify != nil {
		s.runtime.Notify(message, true)
		return
	}
	s.actionError = message
}

// openFile 打开下载文件或其所在目录；文件被移动或删除时弹出提示，
// 不再静默失败。
func (s *State) openFile(path string, directory bool) {
	action := s.runtime.OpenFile
	if directory {
		action = s.runtime.OpenLocation
	}
	if action == nil {
		return
	}
	if path == "" {
		s.reportActionError("打开失败：文件路径为空")
		return
	}
	s.actionError = ""
	if s.runtime.Run == nil {
		// 测试与离屏环境没有调度器，同步执行即可。
		if err := action(path); err != nil {
			s.reportActionError("打开失败：" + err.Error())
		}
		return
	}
	generation := s.generation
	s.runtime.Run(func() error { return action(path) }, func(err error) {
		if !s.stopped && generation == s.generation && err != nil {
			s.reportActionError("打开失败：" + err.Error())
		}
	})
}

// historyCardButton 是历史卡片的紧凑操作按钮，尺寸与 B 站/抖音的下载卡片一致。
func historyCardButton(c *ui.Context, label string, opts components.ButtonOptions) ui.Element {
	return components.Button(c, label, opts).Height(style.DownloadControlHeight).MinHeight(style.DownloadControlHeight).Padding(0, style.SpaceSmall)
}

// deleteHistory 删除一条下载记录；已保存到本地的文件不受影响。
func (s *State) deleteHistory(item yt.History) {
	if s.api == nil || s.runtime.Run == nil || s.historyBusy {
		return
	}
	s.historyBusy = true
	s.actionError = ""
	generation := s.generation
	api := s.api
	s.runtime.Run(func() error { return api.DeleteHistory(item.ID) }, func(err error) {
		if s.stopped || generation != s.generation {
			return
		}
		s.historyBusy = false
		if err != nil {
			s.reportActionError("删除记录失败：" + err.Error())
			return
		}
		s.history = slices.DeleteFunc(s.history, func(v yt.History) bool { return v.ID == item.ID })
	})
}
