package bilibili

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kamiertop/videodown/internal/gui/components"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/icons"
	newc "github.com/oneo-me/onevra/source/components"

	"github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) download(c *ui.Context) {
	panel(c.Key("input-panel"), "视频下载输入栏", func() {
		ui.Row(c).FillWidth().Padding(style.SpaceMedium).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			textInput(c.Key("link"), &s.link, "BiliBili视频链接", "输入视频链接，支持 BV 号、AV 号、视频 URL", s.parseDownloadInput).Grow(1)
			parseLabel := "解析"
			if s.manual.parse.Loading {
				parseLabel = "解析中"
			}
			button(c.Key("parse"), parseLabel, components.ButtonOptions{Variant: components.ButtonPrimary, Disabled: strings.TrimSpace(s.link) == "" || s.manual.parse.Loading || s.api == nil}).OnClick(s.parseDownloadInput)
			button(c.Key("clear"), "清空", components.ButtonOptions{Variant: components.ButtonOutline, Disabled: strings.TrimSpace(s.link) == ""}).OnClick(func() {
				s.link = ""
				s.manual.input = ""
				s.manual.parse.seq++
				s.manual.parse.Loading = false
				s.manual.parse.Loaded = false
				s.manual.parse.Error = ""
				s.manual.group = nil
			})
		})
	})
	s.errorRow(c, s.manual.parse.Error, nil)
	if s.stage.notice != "" {
		ui.Text(c, s.stage.notice).FontSize(style.Font(c, style.FontMeta)).TextColor(c.Theme().TextMuted).MaxLines(1)
	}
	rows, ready := s.downloadRows()
	if len(s.stage.keys) > 0 {
		panel(c.Key("queue-toolbar"), "下载操作栏", func() {
			ui.Column(c).FillWidth().Padding(style.SpaceSmall, style.SpaceMedium).Gap(style.SpaceSmall).Children(func() {
				ui.Row(c).FillWidth().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
					ui.Text(c, fmt.Sprintf("待下载 · %d", len(s.stage.keys))).FontWeight(style.WeightSemibold)
					ui.Text(c, fmt.Sprintf("已就绪 %d · 已排队 %d", ready, len(s.queue.requested))).Grow(1).FontSize(style.Font(c, style.FontMeta)).TextColor(c.Theme().TextMuted)
					if s.queue.downloading {
						button(c.Key("stop-queue"), "停止后续任务", components.ButtonOptions{Variant: components.ButtonOutline, Disabled: s.queue.stop}).OnClick(s.stopQueuedDownloads)
					} else {
						button(c.Key("download"), "开始下载", components.ButtonOptions{Variant: components.ButtonPrimary, Icon: icons.Must("download"), Disabled: len(s.stage.keys) == 0 || ready != len(s.stage.keys) || s.downloadBackend == nil || s.runtime.Background == nil || len(s.queue.requested) == len(s.stage.keys)}).OnClick(s.startQueuedDownloads)
					}
					button(c.Key("clear-queue"), "清空列表", components.ButtonOptions{Variant: components.ButtonGhost, Disabled: len(s.stage.keys) == 0 || s.queue.downloading || len(s.queue.extraBusy) != 0}).OnClick(func() { s.clearStaged() })
				})
				if s.queue.total > 0 {
					caption(c, fmt.Sprintf("已完成 %d/%d · 失败 %d", s.queue.done, s.queue.total, s.queue.failed))
				}
				s.errorRow(c, s.queue.error, nil)
			})
		})
	}
	if group := s.manual.group; group != s.manual.shown {
		s.manual.shown = group
		if group != nil {
			s.downloadList.ScrollTo(0, ui.Start)
		}
	}
	s.downloadList.Key = func(i int) any { return rows[i].key }
	s.downloadList.Label = func(i int) string { return rows[i].title }
	list := ui.List(c.Key("download-items"), &s.downloadList, len(rows)).
		Label("下载任务区域").Grow(1).MinWidth(0).MinHeight(0).
		Gap(style.SpaceMedium).Padding(0, style.ScrollbarGutter, style.SpaceMedium, 0)
	if len(rows) == 0 {
		list.Children(func() { empty(c, "下载列表为空", "download") })
	}
	list.Rows(func(row ui.ListRow) {
		item := rows[row.Index]
		switch key := item.key.(type) {
		case *parsedVideoGroup:
			s.parsedGroupPanel(row.Context, key)
		case uint64:
			s.batchDownloadCard(row.Context, item.batch)
		case string:
			s.stagedMediaCard(row.Context, key, item.title, item.cover, item.author, item.length)
		}
	})
}

type downloadRow struct {
	// Parsed groups, batches and staged videos have distinct key types.
	key                          any
	title, cover, author, length string
	batch                        *batchSession
}

func (s *State) downloadRows() ([]downloadRow, int) {
	rows := make([]downloadRow, 0, 1+len(s.batches.sessions)+len(s.stage.keys))
	if group := s.manual.group; group != nil {
		rows = append(rows, downloadRow{key: group, title: group.Title})
	}
	for _, batch := range s.batches.sessions {
		rows = append(rows, downloadRow{key: batch.ID, title: batch.Source.Title, batch: batch})
	}
	ready := 0
	add := func(key, title, cover, author, length string) {
		rows = append(rows, downloadRow{key: key, title: title, cover: cover, author: author, length: length})
		if entry := s.queue.resolves[key]; entry.resolved() && queueVideoURL(entry.data.Data.PlayURL, entry.videoID) != "" {
			ready++
		}
	}
	for _, v := range s.stage.uploads {
		add(v.Bvid, v.Title, v.Pic, v.Author, v.Length)
	}
	for _, v := range s.stage.favorites {
		add(favoriteBvid(v.Bvid, v.BvID), v.Title, v.Cover, v.Upper.Name, duration(v.Duration))
	}
	for _, v := range s.stage.collections {
		add(v.Bvid, v.Title, v.Cover, v.Upper.Name, duration(v.Duration))
	}
	for _, v := range s.stage.seasons {
		add(v.Bvid, v.Title, v.Pic, "", duration(v.Duration))
	}
	for _, v := range s.stage.series {
		add(v.Bvid, v.Title, v.Pic, "", duration(v.Duration))
	}
	for _, v := range s.stage.dynamics {
		add(v.Bvid, v.Title, v.Cover, v.AuthorName, v.DurationText)
	}
	for _, v := range s.stage.playHistory {
		add(v.History.Bvid, playHistoryTitle(v), v.Cover, v.AuthorName, playHistoryDuration(v))
	}
	for _, v := range s.stage.manual {
		add(v.Key, v.Title, v.Cover, v.Author, duration(v.Duration))
	}
	return rows, ready
}

func (s *State) stagedMediaCard(c *ui.Context, bvid, title, cover, author, length string) {
	k := c.Theme()
	width, _ := c.Size()
	wide := width >= style.DownloadCompactBreakpoint
	panel(c, bvid, func() {
		row := ui.Row(c).FillWidth().Padding(style.DownloadCardPadding).Gap(style.SpaceMedium).AlignItems(ui.Start)
		if !wide {
			row.Column()
		}
		row.Children(func() {
			ui.Column(c.Key("thumbnail")).Width(style.DownloadCoverWidth).Shrink(0).Children(func() {
				s.imageWithRadius(c, cover, style.DownloadCoverWidth, style.DownloadCoverHeight, false, style.RadiusMedia)
			})
			details := ui.Column(c.Key("details")).MinWidth(0).Gap(style.DownloadCardGap)
			if wide {
				details.Grow(1)
			} else {
				details.FillWidth()
			}
			details.Children(func() {
				progress, downloading := s.progressForKey(bvid)
				header := ui.Row(c.Key("header")).FillWidth().Gap(style.SpaceMedium).AlignItems(ui.Start)
				if width < style.DownloadInlineBreakpoint {
					header.Column()
				}
				header.Children(func() {
					info := ui.Column(c.Key("info")).MinWidth(0).Gap(style.DownloadCardGap)
					if width >= style.DownloadInlineBreakpoint {
						info.Grow(1)
					} else {
						info.FillWidth()
					}
					info.Children(func() {
						ui.Text(c, title).FillWidth().FontWeight(style.WeightSemibold).MaxLines(2).Tooltip(title)
						var metadata []string
						if author != "" {
							metadata = append(metadata, author)
						}
						if length != "" && length != "0:00" {
							metadata = append(metadata, length)
						}
						if len(metadata) > 0 {
							ui.Text(c, strings.Join(metadata, " · ")).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted).MaxLines(1)
						}
					})
					// Reserve a compact status area above the actions on wide cards.
					// Narrow cards put progress on its own line to keep titles readable.
					if downloading {
						status := ui.Column(c.Key("progress")).Label("视频下载进度").MinWidth(0)
						if width >= style.DownloadInlineBreakpoint {
							status.Width(200).Shrink(0)
						} else {
							status.FillWidth()
						}
						status.Children(func() { downloadProgress(c, progress) })
					}
				})
				task, ready := s.queuedTask(bvid)
				busy := s.queue.extraBusy[bvid] != ""
				extrasBusy := busy || s.queue.requested[bvid]
				ui.Row(c.Key("actions")).Label("视频下载选项").FillWidth().Wrap().Gap(style.SpaceSmall).AlignItems(ui.Center).Padding(4, 0, 0, 0).Children(func() {
					s.queueStreamControls(c.Key("streams"), bvid)
					ui.Box(c).Grow(1)
					actions := ui.Row(c.Key("buttons")).Gap(4).AlignItems(ui.Center)
					if width < style.DownloadInlineBreakpoint {
						actions.FillWidth().Justify(ui.End)
					}
					actions.Children(func() {
						if s.downloadBackend != nil {
							downloadCardButton(c.Key("cover"), "封面", components.ButtonOptions{Label: "保存封面", Variant: components.ButtonGhost, Disabled: !ready || task.Cover == "" || s.queue.downloading || s.batchesRunning() || extrasBusy}).OnClick(func() { s.startQueueExtra(bvid, "cover") })
							downloadCardButton(c.Key("audio-only"), "音频", components.ButtonOptions{Label: "下载音频", Variant: components.ButtonGhost, Disabled: !ready || task.AudioURL == "" || s.queue.downloading || s.batchesRunning() || extrasBusy}).OnClick(func() { s.startQueueExtra(bvid, "audio") })
						}
						downloadCardButton(c.Key("remove"), "移除", components.ButtonOptions{Variant: components.ButtonGhost, Disabled: s.queue.downloading || busy}).OnClick(func() { s.removeStaged(bvid) })
						downloadCardButton(c.Key("single-download"), downloadRequestLabel(s.queue.requested[bvid]), components.ButtonOptions{Label: "下载视频", Variant: components.ButtonPrimary, Icon: icons.Must("download"), Disabled: !ready || s.queue.requested[bvid] || s.runtime.Background == nil || s.downloadBackend == nil}).OnClick(func() { s.startSingleQueuedDownload(bvid) })
					})
				})
			})
		})
	})
}

// Keep download card controls compact while retaining the shared button style.
func downloadCardButton(c *ui.Context, label string, opts components.ButtonOptions) ui.Element {
	return button(c, label, opts).MinHeight(style.DownloadControlHeight).Height(style.DownloadControlHeight).Padding(0, style.SpaceSmall)
}

func (s *State) parsedGroupPanel(c *ui.Context, group *parsedVideoGroup) {
	count := 0
	for _, item := range group.Items {
		if group.Selected[item.Key] {
			count++
		}
	}
	title := fmt.Sprintf("%s · %s · %d 个视频 · 已选 %d", group.Kind, group.Title, len(group.Items), count)
	newc.Accordion(c, func(g *newc.AccordionGroup) {
		g.Item(title, &group.Open, func() {
			ui.Column(c.Key("items")).FillWidth().Gap(4).Children(func() {
				for _, item := range group.Items {
					selected := group.Selected[item.Key]
					k := c.Theme()
					row := ui.ButtonBase(c.Key(item.Key)).Label("选择待添加："+item.Title).FillWidth().MinWidth(0).
						MinHeight(style.InputHeight).Padding(6, style.SpaceSmall).Gap(style.SpaceSmall).Radius(style.RadiusControl).FocusRing(true)
					if selected {
						row.Background(k.Selection).TextColor(k.Accent)
					} else {
						row.TextColor(k.Text)
					}
					if row.Hovered() {
						row.Background(k.SurfaceHover)
					}
					row.Children(func() {
						icon := "plus"
						if selected {
							icon = "check"
						}
						ui.Icon(c, icons.Must(icon)).Size(16, 16).Shrink(0)
						ui.Text(c, item.Title).Grow(1).MinWidth(0).MaxLines(1)
						if item.Duration > 0 {
							ui.Text(c, duration(item.Duration)).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted).Shrink(0)
						}
					}).OnClick(func() { group.Selected[item.Key] = !group.Selected[item.Key] })
				}
			})
			ui.Row(c.Key("group-actions")).FillWidth().Wrap().Gap(style.SpaceSmall).Children(func() {
				button(c.Key("add-current"), "添加当前视频", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() {
					s.stageManual(group.Current)
					s.manual.group = nil
				})
				button(c.Key("add-selected"), fmt.Sprintf("添加已选 (%d)", count), components.ButtonOptions{Variant: components.ButtonPrimary, Disabled: count == 0}).OnClick(func() {
					for _, item := range group.Items {
						if group.Selected[item.Key] {
							s.stageManual(item)
						}
					}
					s.manual.group = nil
				})
				button(c.Key("cancel-group"), "取消", components.ButtonOptions{Variant: components.ButtonGhost}).OnClick(func() { s.manual.group = nil })
			})
		})
	}).Label("解析到的" + group.Kind).FillWidth().MinWidth(0).Radius(style.RadiusPanel).Clip()
}

func (s *State) queueStreamControls(c *ui.Context, bvid string) {
	entry := s.queue.resolves[bvid]
	if entry == nil || entry.data.Loading {
		caption(c, "正在解析画质与音频")
		return
	}
	if entry.data.Error != "" {
		s.errorRow(c, entry.data.Error, nil)
		button(c.Key("retry-resolve"), "重新解析", components.ButtonOptions{Variant: components.ButtonOutline}).OnClick(func() { s.retryQueueItem(bvid) })
		return
	}
	if !entry.data.Loaded || entry.data.Data.PlayURL == nil {
		caption(c, "等待解析")
		return
	}
	play := entry.data.Data.PlayURL

	ui.Row(c.Key("video-field")).Gap(4).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "画质").FontSize(style.Font(c, style.FontMeta)).TextColor(c.Theme().TextMuted).Shrink(0)
		video, _ := components.Select(c.Key("quality"), &entry.videoID, queueVideoOptions(play), components.SelectOptions{Label: "视频画质", Disabled: s.queue.requested[bvid]})
		video.Width(style.DownloadStreamWidth).MinWidth(0).MinHeight(style.DownloadControlHeight).Height(style.DownloadControlHeight).Padding(0, style.SpaceSmall).FontSize(style.Font(c, style.FontButton)).Radius(style.RadiusControl)
	})
	audioOptions := queueAudioOptions(play)
	if len(audioOptions) > 0 {
		ui.Row(c.Key("audio-field")).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "音质").FontSize(style.Font(c, style.FontMeta)).TextColor(c.Theme().TextMuted).Shrink(0)
			audio, _ := components.Select(c.Key("audio"), &entry.audioID, audioOptions, components.SelectOptions{Label: "音频音质", Disabled: s.queue.requested[bvid]})
			audio.Width(style.DownloadAudioWidth).MinWidth(0).MinHeight(style.DownloadControlHeight).Height(style.DownloadControlHeight).Padding(0, style.SpaceSmall).FontSize(style.Font(c, style.FontButton)).Radius(style.RadiusControl)
		})
	} else {
		caption(c, "无独立音轨")
	}
}

func (s *State) removeStaged(key string) {
	s.cancelPendingManualKey(key)
	if task, ok := s.queuedTask(key); ok {
		delete(s.queue.progress, queueProgressKey(task.Bvid, task.Cid))
	}
	delete(s.stage.keys, key)
	delete(s.stage.sourceNames, key)
	delete(s.queue.resolves, key)
	s.stage.uploads = slices.DeleteFunc(s.stage.uploads, func(v model.VListItem) bool { return v.Bvid == key })
	s.stage.favorites = slices.DeleteFunc(s.stage.favorites, func(v model.FavoriteMedias) bool { return favoriteBvid(v.Bvid, v.BvID) == key })
	s.stage.collections = slices.DeleteFunc(s.stage.collections, func(v model.CollectionItemMedias) bool { return v.Bvid == key })
	s.stage.seasons = slices.DeleteFunc(s.stage.seasons, func(v model.SeasonsArchivesItem) bool { return v.Bvid == key })
	s.stage.series = slices.DeleteFunc(s.stage.series, func(v model.SeriesArchivesItem) bool { return v.Bvid == key })
	s.stage.dynamics = slices.DeleteFunc(s.stage.dynamics, func(v model.DynamicArchiveItem) bool { return v.Bvid == key })
	s.stage.playHistory = slices.DeleteFunc(s.stage.playHistory, func(v model.PlayHistoryItem) bool { return v.History.Bvid == key })
	s.stage.manual = slices.DeleteFunc(s.stage.manual, func(v manualVideo) bool { return v.Key == key })
	delete(s.queue.progress, key)
	delete(s.queue.extraBusy, key)
	s.stage.notice = ""
}
func (s *State) clearStaged() {
	s.queue.pending = nil
	s.queue.requested = make(map[string]bool)
	s.stage.keys = make(map[string]bool)
	s.stage.sourceNames = make(map[string]string)
	s.queue.resolves = make(map[string]*queueResolve)
	s.stage.uploads = nil
	s.stage.favorites = nil
	s.stage.collections = nil
	s.stage.seasons = nil
	s.stage.series = nil
	s.stage.dynamics = nil
	s.stage.playHistory = nil
	s.stage.manual = nil
	s.queue.progress = make(map[string]download.Progress)
	s.queue.extraBusy = make(map[string]string)
	s.stage.notice = ""
}

func downloadRequestLabel(queued bool) string {
	if queued {
		return "已排队"
	}
	return "下载"
}
