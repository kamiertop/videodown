package youtube

import (
	"fmt"
	"slices"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
)

// panel and toolbar share the same surface, radius and spacing as history cards.
func panel(c *ui.Context, label string, children func()) {
	k := core.Tokens(c)
	ui.Column(c).Label(label).FillWidth().MinWidth(0).Radius(style.RadiusPanel).Border(style.BorderWidth, k.Border).Background(k.Surface).Children(children)
}

func (s *State) queueView(c *ui.Context) {
	if len(s.queue) == 0 {
		return
	}
	panel(c.Key("queue-toolbar"), "YouTube 队列工具栏", func() {
		ui.Row(c).FillWidth().MinWidth(0).Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("待下载 %d · 下载中 %d · 等待 %d", s.readyCount(), len(s.active), s.waitingCount())).Grow(1).MinWidth(0).FontWeight(style.WeightSemibold)
			components.Button(c.Key("start-all"), "下载全部", input.ButtonOptions{Variant: input.Primary, Disabled: s.readyCount() == 0}).OnClick(s.startAll)
			label := "暂停队列"
			if s.queuePaused {
				label = "继续队列"
			}
			components.Button(c.Key("pause"), label, input.ButtonOptions{Variant: input.Outline, Disabled: s.PendingDownloadCount() == 0}).OnClick(func() { s.queuePaused = !s.queuePaused; s.pumpQueue() }).Tooltip("暂停后当前任务继续，后续任务等待")
			components.Button(c.Key("clear"), "清理已结束", input.ButtonOptions{Variant: input.Ghost}).OnClick(func() {
				s.queue = slices.DeleteFunc(s.queue, func(item *queueItem) bool {
					return item.Status != "ready" && item.Status != "active" && item.Status != "queued"
				})
				s.pruneQueueGroups()
			})
		})
	})
	for _, group := range s.queueGroups {
		items := s.groupItems(group)
		if len(items) == 0 {
			continue
		}
		panel(c.Key(group.Key), "YouTube 队列分组："+group.Key, func() {
			ui.Row(c).Label("队列分组标题：" + group.Key).FillWidth().MinWidth(0).Padding(style.SpaceMedium).Gap(style.SpaceMedium).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Gap(style.SpaceSmall).Children(func() {
					ui.Text(c, group.Title).FillWidth().MaxLines(2).FontWeight(style.WeightSemibold).Tooltip(group.Title)
					summary := groupSummary(items)
					ui.Text(c, summary).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(core.Tokens(c).TextMuted).Tooltip(summary)
				})
				label, icon := "展开", "chevron-down"
				if group.Expanded {
					label, icon = "收起", "chevron-up"
				}
				ready := false
				for _, item := range items {
					if item.Status == "ready" {
						ready = true
						break
					}
				}
				components.Button(c.Key("start-group"), "下载全部", input.ButtonOptions{Variant: input.Primary, Disabled: !ready, Label: "下载分组：" + group.Key}).MinWidth(0).Padding(4, 8).Shrink(0).OnClick(func() { s.startGroup(group) })
				components.Button(c.Key("toggle"), label, input.ButtonOptions{Variant: input.Outline, Icon: icons.Must(icon), Label: "展开或收起分组：" + group.Key}).MinWidth(0).Padding(4, 8).Shrink(0).OnClick(func() { group.Expanded = !group.Expanded })
				components.Button(c.Key("remove"), "移除", input.ButtonOptions{Variant: input.Ghost, Label: "移除分组：" + group.Key}).MinWidth(0).Padding(4, 8).Shrink(0).Tooltip("取消本组未完成任务，保留已下载文件").OnClick(func() { s.removeGroup(group) })
			})
			if group.Expanded {
				for _, item := range items {
					s.queueItemRow(c.Key(item.ID), item)
				}
			}
		})
	}
}

func (s *State) queueItemRow(c *ui.Context, item *queueItem) {
	k := core.Tokens(c)
	row := ui.Row(c).Label(fmt.Sprintf("YouTube 队列任务：%d", item.ID)).FillWidth().MinWidth(0).Padding(style.DownloadCardPadding).Gap(style.SpaceMedium).AlignItems(ui.Start).BorderWidth(1, 0, 0, 0).BorderColor(k.Border)
	if row.Bounds().W == 0 {
		c.After(time.Millisecond)
	}
	if row.InView() {
		s.prepareItem(item)
	}
	row.Children(func() {
		s.thumbnailFrame(c.Key("cover"), item.Video, style.DownloadCoverWidth, style.DownloadCoverHeight, fmt.Sprintf("任务封面：%d", item.ID))
		ui.Column(c).Label(fmt.Sprintf("任务信息：%d", item.ID)).Grow(1).MinWidth(0).Height(style.DownloadCoverHeight).Gap(4).Children(func() {
			ui.Column(c).Grow(1).MinHeight(0).FillWidth().MinWidth(0).Gap(4).Children(func() {
				ui.Text(c, item.Video.Title).Label(fmt.Sprintf("任务标题：%d", item.ID)).FillWidth().MinWidth(0).MaxLines(1).Shrink(0).FontWeight(style.WeightSemibold).Tooltip(item.Video.Title)
				metadata := fmt.Sprintf("%s · %s", item.Video.Author, item.Video.Duration.Round(time.Second))
				detail := metadata
				if item.DetailsLoading {
					metadata += " · 读取画质中"
				}
				if editableTask(item) && item.DetailsError != "" {
					metadata += " · 画质解析失败"
					detail = item.DetailsError
				}
				if item.Status == "error" {
					metadata += " · 下载失败"
					detail = item.Error
				}
				ui.Text(c, metadata).FillWidth().MaxLines(1).Shrink(0).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted).Tooltip(detail)
			})
			ui.Row(c).Label(fmt.Sprintf("任务底栏：%d", item.ID)).FillWidth().MinWidth(0).Shrink(0).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {

				if editableTask(item) {
					ui.Row(c).Label(fmt.Sprintf("任务格式栏：%d", item.ID)).MinWidth(0).Shrink(0).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
						label := "画质"
						if item.AudioOnly {
							label = "音质"
						}
						ui.Text(c, label).Shrink(0).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
						qualitySelect(c.Key("stream"), &item.Format, queueFormatOptions(item), 170, fmt.Sprintf("任务画质：%d", item.ID), fmt.Sprintf("任务画质选项：%d", item.ID), false).Height(style.DownloadControlHeight).MinHeight(style.DownloadControlHeight).Padding(0, style.SpaceSmall)
						if item.DetailsError != "" {
							components.Button(c.Key("resolve"), "", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("refresh-cw"), Label: fmt.Sprintf("重新解析任务：%d", item.ID)}).Size(30, 30).MinWidth(30).Padding(0).Tooltip("重新解析画质").OnClick(func() { item.DetailsError = ""; s.prepareItem(item) })
						}
					})
				} else {
					text := queueStatus(item)
					ui.Text(c, text).Label(fmt.Sprintf("任务进度：%d", item.ID)).Grow(1).MinWidth(0).MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted).Tooltip(text)
				}
				if editableTask(item) {
					ui.Box(c).Grow(1)
				}
				s.queueItemActions(c.Key("actions"), item)
			})
		})
	})
}

func (s *State) queueItemActions(c *ui.Context, item *queueItem) {
	k := core.Tokens(c)
	ui.Row(c).Label(fmt.Sprintf("任务操作：%d", item.ID)).Shrink(0).AlignSelf(ui.End).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
		if editableTask(item) {
			mode := components.Button(c.Key("mode"), "音频", input.ButtonOptions{Variant: input.Ghost, Label: fmt.Sprintf("下载方式：%d", item.ID)}).Tooltip("仅下载音频，再次点击恢复视频").OnClick(func() { item.AudioOnly = !item.AudioOnly; item.Format = 0 })
			compactTaskButton(mode)
			if item.AudioOnly {
				mode.Background(k.Selection).TextColor(k.AccentText)
			}
		}
		if item.Path != "" {
			if s.runtime.OpenFile != nil {
				compactTaskButton(components.Button(c.Key("open"), "打开", input.ButtonOptions{Variant: input.Ghost, Label: fmt.Sprintf("打开任务：%d", item.ID)}).OnClick(func() { _ = s.runtime.OpenFile(item.Path) }))
			}
			if s.runtime.OpenLocation != nil {
				compactTaskButton(components.Button(c.Key("directory"), "", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("folder"), Label: fmt.Sprintf("打开任务目录：%d", item.ID)}).Tooltip("打开目录").OnClick(func() { _ = s.runtime.OpenLocation(item.Path) }))
			}
		}
		compactTaskButton(components.Button(c.Key("remove"), "移除", input.ButtonOptions{Variant: input.Ghost, Label: fmt.Sprintf("移除任务：%d", item.ID)}).OnClick(func() { s.removeItem(item) }))
		if editableTask(item) {
			label := "下载"
			if item.Status == "error" || item.Status == "canceled" {
				label = "重试"
			}
			compactTaskButton(components.Button(c.Key("start"), label, input.ButtonOptions{Variant: input.Primary, Icon: icons.Must("download"), Label: fmt.Sprintf("下载任务：%d", item.ID)}).OnClick(func() { s.startItem(item) }))
		} else if item.Status == "active" || item.Status == "queued" {
			compactTaskButton(components.Button(c.Key("cancel"), "取消", input.ButtonOptions{Variant: input.Outline, Label: fmt.Sprintf("取消任务：%d", item.ID)}).OnClick(func() { s.cancelItem(item) }))
		}
	})
}

func compactTaskButton(button ui.Element) {
	button.MinWidth(0).Height(style.DownloadControlHeight).MinHeight(style.DownloadControlHeight).Padding(0, style.SpaceSmall).Gap(4).Shrink(0)
}

func queueFormatOptions(item *queueItem) []yt.Format {
	return formatOptions(item.Video, item.AudioOnly)
}

func formatOptions(video yt.Video, audioOnly bool) []yt.Format {
	label := "最高画质"
	formats := video.Formats
	if audioOnly {
		label = "最佳音质"
		formats = video.AudioFormats
	}
	options := []yt.Format{{ID: 0, Label: label, Summary: label}}
	if len(formats) > 0 {
		return append(options, formats...)
	}
	if !audioOnly {
		for _, height := range []int{2160, 1440, 1080, 720, 480, 360} {
			text := fmt.Sprintf("最高 %dp", height)
			options = append(options, yt.Format{ID: -height, Label: text, Summary: text})
		}
	}
	return options
}
