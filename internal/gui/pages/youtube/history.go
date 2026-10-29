package youtube

import (
	"fmt"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
)

func (s *State) historyPage(c *ui.Context, scroll *ui.ScrollState) {
	ui.Column(c).Label("YouTube 下载历史页面").Fill().MinWidth(0).MinHeight(0).Padding(style.PagePadding).Gap(style.SpaceMedium).Children(func() {
		k := core.Tokens(c)
		items := s.filteredHistory()
		ui.Row(c).Label("YouTube 历史工具栏").FillWidth().MinWidth(0).Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Radius(style.RadiusPanel).Border(style.BorderWidth, k.Border).Background(k.Surface).Children(func() {
			ui.Text(c, fmt.Sprintf("下载历史 · %d", len(items))).FontWeight(style.WeightSemibold).Shrink(0)
			field := input.TextInput(c.Key("history-search"), &s.historySearch, input.TextInputOptions{Label: "搜索 YouTube 下载历史", Placeholder: "搜索标题、作者、视频 ID 或路径"})
			field.Element.Grow(1).MinWidth(0)
			components.Button(c.Key("clear-search"), "清除搜索", input.ButtonOptions{Variant: input.Ghost, Disabled: s.historySearch == ""}).Shrink(0).OnClick(func() { s.historySearch = "" })
		})
		layout.ScrollArea(c.Key("history-scroll"), scroll, layout.ScrollAreaOptions{}, func() {
			ui.Column(c).FillWidth().MinWidth(0).Children(func() {
				if len(items) == 0 {
					message := "暂无下载记录"
					if strings.TrimSpace(s.historySearch) != "" {
						message = "没有匹配的下载历史"
					}
					components.Muted(c, message)
					return
				}
				k := core.Tokens(c)
				ui.Column(c).FillWidth().MinWidth(0).Radius(style.RadiusPanel).Border(style.BorderWidth, k.Border).Background(k.Surface).Clip().Children(func() {
					for i, item := range items {
						row := ui.Row(c.Key(item.Path)).Label("下载记录：" + item.Path).FillWidth().MinWidth(0).MinHeight(76).Padding(style.SpaceMedium).Gap(style.SpaceMedium).AlignItems(ui.Center)
						if i < len(items)-1 {
							row.BorderWidth(0, 0, style.BorderWidth, 0).BorderColor(k.Border)
						}
						row.Children(func() {
							ui.Column(c).Grow(1).MinWidth(0).Gap(style.SpaceSmall).Children(func() {
								ui.Text(c, item.Title).FillWidth().MaxLines(1).FontWeight(style.WeightSemibold).Tooltip(item.Title)
								path := item.Path
								if path == "" {
									path = "未记录保存路径"
								}
								ui.Text(c, path).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted).Tooltip(path)
							})
							ui.Row(c).Shrink(0).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
								components.Button(c.Key("open"), "打开", input.ButtonOptions{Variant: input.Outline, Icon: icons.Must("play"), Label: "打开视频：" + item.Path, Disabled: s.runtime.OpenFile == nil || item.Path == ""}).MinWidth(0).MinHeight(32).Padding(4, 8).Gap(6).Shrink(0).Tooltip("使用默认播放器打开").OnClick(func() { _ = s.runtime.OpenFile(item.Path) })
								components.Button(c.Key("directory"), "打开目录", input.ButtonOptions{Variant: input.Outline, Icon: icons.Must("folder"), Label: "打开目录：" + item.Path, Disabled: s.runtime.OpenLocation == nil || item.Path == ""}).MinWidth(0).MinHeight(32).Padding(4, 8).Gap(6).Shrink(0).OnClick(func() { _ = s.runtime.OpenLocation(item.Path) })
							})
						})
					}
				})
			})
		}).Grow(1).MinWidth(0).MinHeight(0)
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
