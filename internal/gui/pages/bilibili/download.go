package bilibili

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/data"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"

	"github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) download(c *ui.Context) {
	panel(c.Key("input-panel"), "视频下载输入栏", func() {
		ui.Row(c).FillWidth().Padding(style.SpaceMedium).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			textInput(c.Key("link"), &s.link, "BiliBili视频链接", "输入视频链接，支持 BV 号、AV 号、视频 URL", s.parseDownloadInput).Grow(1)
			parseLabel := "解析"
			if s.manualParse.Loading {
				parseLabel = "解析中"
			}
			button(c.Key("parse"), parseLabel, input.ButtonOptions{Variant: input.Primary, Disabled: strings.TrimSpace(s.link) == "" || s.manualParse.Loading || s.api == nil}).OnClick(s.parseDownloadInput)
			button(c.Key("clear"), "清空", input.ButtonOptions{Variant: input.Outline, Disabled: strings.TrimSpace(s.link) == ""}).OnClick(func() {
				s.link = ""
				s.manualParseInput = ""
				s.manualParse.seq++
				s.manualParse.Loading = false
				s.manualParse.Loaded = false
				s.manualParse.Error = ""
				s.manualGroup = nil
			})
		})
	})
	s.errorRow(c, s.manualParse.Error, nil)
	if s.notice != "" {
		ui.Text(c, s.notice).FontSize(style.Font(c, style.FontMeta)).TextColor(core.Tokens(c).TextMuted).MaxLines(1)
	}
	ready := 0
	for _, key := range s.stagedOrder() {
		if _, ok := s.queuedTask(key); ok {
			ready++
		}
	}
	if len(s.stagedKeys) > 0 || len(s.batches) == 0 {
		panel(c.Key("queue-toolbar"), "下载操作栏", func() {
			ui.Column(c).FillWidth().Padding(style.SpaceSmall, style.SpaceMedium).Gap(style.SpaceSmall).Children(func() {
				ui.Row(c).FillWidth().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
					ui.Text(c, fmt.Sprintf("待下载 · %d", len(s.stagedKeys))).FontWeight(style.WeightSemibold)
					ui.Text(c, fmt.Sprintf("已就绪 %d · 已排队 %d", ready, len(s.queueRequested))).Grow(1).FontSize(style.Font(c, style.FontMeta)).TextColor(core.Tokens(c).TextMuted)
					if s.queueDownloading {
						button(c.Key("stop-queue"), "停止后续任务", input.ButtonOptions{Variant: input.Outline, Disabled: s.queueDownloadStop}).OnClick(s.stopQueuedDownloads)
					} else {
						button(c.Key("download"), "开始下载", input.ButtonOptions{Variant: input.Primary, Icon: icons.Must("download"), Disabled: len(s.stagedKeys) == 0 || ready != len(s.stagedKeys) || s.downloadBackend == nil || s.runtime.Background == nil || len(s.queueRequested) == len(s.stagedKeys)}).OnClick(s.startQueuedDownloads)
					}
					button(c.Key("clear-queue"), "清空列表", input.ButtonOptions{Variant: input.Ghost, Disabled: len(s.stagedKeys) == 0 || s.queueDownloading || len(s.queueExtraBusy) != 0}).OnClick(func() { s.clearStaged() })
				})
				if s.queueDownloadTotal > 0 {
					caption(c, fmt.Sprintf("已完成 %d/%d · 失败 %d", s.queueDownloadDone, s.queueDownloadTotal, s.queueDownloadFailed))
				}
				s.errorRow(c, s.queueDownloadError, nil)
			})
		})
	}
	layout.ScrollArea(c.Key("download-scroll"), s.scroll("download-items"), layout.ScrollAreaOptions{}, func() {
		ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, style.SpaceMedium, 0).Gap(style.SpaceMedium).Children(func() {
			if s.manualGroup != nil {
				s.parsedGroupPanel(c.Key("parsed-group"))
			}
			if len(s.stagedKeys) == 0 && s.manualGroup == nil && len(s.batches) == 0 {
				empty(c, "下载列表为空", "download")
			}
			for _, batch := range s.batches {
				s.batchDownloadCard(c.Key(batch.ID), batch)
			}
			ui.Column(c.Key("queue-items")).FillWidth().Gap(style.SpaceMedium).Children(func() {
				for _, v := range s.stagedUploads {
					s.stagedMediaCard(c.Key(v.Bvid), v.Bvid, v.Title, v.Pic, v.Author, v.Length)
				}
				for _, v := range s.stagedFavorites {
					key := favoriteBvid(v.Bvid, v.BvID)
					s.stagedMediaCard(c.Key(key), key, v.Title, v.Cover, v.Upper.Name, duration(v.Duration))
				}
				for _, v := range s.stagedCollections {
					s.stagedMediaCard(c.Key(v.Bvid), v.Bvid, v.Title, v.Cover, v.Upper.Name, duration(v.Duration))
				}
				for _, v := range s.stagedSeasons {
					s.stagedMediaCard(c.Key(v.Bvid), v.Bvid, v.Title, v.Pic, "", duration(v.Duration))
				}
				for _, v := range s.stagedSeries {
					s.stagedMediaCard(c.Key(v.Bvid), v.Bvid, v.Title, v.Pic, "", duration(v.Duration))
				}
				for _, v := range s.stagedDynamics {
					s.stagedMediaCard(c.Key(v.Bvid), v.Bvid, v.Title, v.Cover, v.AuthorName, v.DurationText)
				}
				for _, v := range s.stagedPlayHistory {
					s.stagedMediaCard(c.Key(v.History.Bvid), v.History.Bvid, playHistoryTitle(v), v.Cover, v.AuthorName, playHistoryDuration(v))
				}
				for _, v := range s.stagedManual {
					s.stagedMediaCard(c.Key(v.Key), v.Key, v.Title, v.Cover, v.Author, duration(v.Duration))
				}
			})
		})
	}).Label("下载任务区域").Grow(1).MinWidth(0).MinHeight(0)
}

func (s *State) stagedMediaCard(c *ui.Context, bvid, title, cover, author, length string) {
	k := core.Tokens(c)
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
				ui.Text(c, title).FillWidth().FontWeight(style.WeightSemibold).MaxLines(2)
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
				task, ready := s.queuedTask(bvid)
				busy := s.queueExtraBusy[bvid] != ""
				extrasBusy := busy || s.queueRequested[bvid]
				ui.Row(c.Key("actions")).Label("视频下载选项").FillWidth().Wrap().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
					s.queueStreamControls(c.Key("streams"), bvid)
					ui.Box(c).Grow(1)
					actions := ui.Row(c.Key("buttons")).Gap(4).AlignItems(ui.Center)
					if width < style.DownloadInlineBreakpoint {
						actions.FillWidth().Justify(ui.End)
					}
					actions.Children(func() {
						if _, ok := s.downloadBackend.(DownloadExtras); ok {
							downloadCardButton(c.Key("cover"), "封面", input.ButtonOptions{Label: "保存封面", Variant: input.Ghost, Disabled: !ready || task.Cover == "" || s.queueDownloading || s.batchesRunning() || extrasBusy}).OnClick(func() { s.startQueueExtra(bvid, "cover") })
							downloadCardButton(c.Key("audio-only"), "音频", input.ButtonOptions{Label: "下载音频", Variant: input.Ghost, Disabled: !ready || task.AudioURL == "" || s.queueDownloading || s.batchesRunning() || extrasBusy}).OnClick(func() { s.startQueueExtra(bvid, "audio") })
						}
						downloadCardButton(c.Key("remove"), "移除", input.ButtonOptions{Variant: input.Ghost, Disabled: s.queueDownloading || busy}).OnClick(func() { s.removeStaged(bvid) })
						downloadCardButton(c.Key("single-download"), downloadRequestLabel(s.queueRequested[bvid]), input.ButtonOptions{Label: "下载视频", Variant: input.Primary, Icon: icons.Must("download"), Disabled: !ready || s.queueRequested[bvid] || s.runtime.Background == nil || s.downloadBackend == nil}).OnClick(func() { s.startSingleQueuedDownload(bvid) })
					})
				})
				s.queueProgressBar(c.Key("progress"), bvid)
			})
		})
	})
}

// Keep download card controls compact while retaining the shared button style.
func downloadCardButton(c *ui.Context, label string, opts input.ButtonOptions) ui.Element {
	return button(c, label, opts).MinHeight(style.DownloadControlHeight).Height(style.DownloadControlHeight).Padding(0, style.SpaceSmall)
}

func (s *State) parsedGroupPanel(c *ui.Context) {
	group := s.manualGroup
	count := 0
	for _, item := range group.Items {
		if group.Selected[item.Key] {
			count++
		}
	}
	title := fmt.Sprintf("%s · %s · %d 个视频 · 已选 %d", group.Kind, group.Title, len(group.Items), count)
	data.Accordion(c, &group.OpenSections, []data.AccordionItem{{ID: "videos", Title: title, Content: func() {
		ui.Column(c.Key("items")).FillWidth().Gap(4).Children(func() {
			for _, item := range group.Items {
				selected := group.Selected[item.Key]
				k := core.Tokens(c)
				row := ui.ButtonBase(c.Key(item.Key)).Label("选择待添加："+item.Title).FillWidth().MinWidth(0).
					MinHeight(style.InputHeight).Padding(6, style.SpaceSmall).Gap(style.SpaceSmall).Radius(style.RadiusControl).FocusRing(true)
				if selected {
					row.Background(k.Selection).TextColor(k.AccentText)
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
			button(c.Key("add-current"), "添加当前视频", input.ButtonOptions{Variant: input.Outline}).OnClick(func() {
				s.stageManual(group.Current)
				s.manualGroup = nil
			})
			button(c.Key("add-selected"), fmt.Sprintf("添加已选 (%d)", count), input.ButtonOptions{Variant: input.Primary, Disabled: count == 0}).OnClick(func() {
				for _, item := range group.Items {
					if group.Selected[item.Key] {
						s.stageManual(item)
					}
				}
				s.manualGroup = nil
			})
			button(c.Key("cancel-group"), "取消", input.ButtonOptions{Variant: input.Ghost}).OnClick(func() { s.manualGroup = nil })
		})
	}}}, data.AccordionOptions{}).Element.Label("解析到的" + group.Kind).FillWidth().MinWidth(0).Radius(style.RadiusPanel).Clip()
}

func (s *State) queueProgressBar(c *ui.Context, key string) {
	p, ok := s.progressForKey(key)
	if !ok {
		return
	}
	downloadProgress(c, p)
}

func (s *State) queueStreamControls(c *ui.Context, bvid string) {
	entry := s.queueResolves[bvid]
	if entry == nil || entry.data.Loading {
		caption(c, "正在解析画质与音频")
		return
	}
	if entry.data.Error != "" {
		s.errorRow(c, entry.data.Error, nil)
		button(c.Key("retry-resolve"), "重新解析", input.ButtonOptions{Variant: input.Outline}).OnClick(func() { s.retryQueueItem(bvid) })
		return
	}
	if !entry.data.Loaded || entry.data.Data.PlayURL == nil {
		caption(c, "等待解析")
		return
	}
	play := entry.data.Data.PlayURL

	ui.Row(c.Key("video-field")).Gap(4).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "画质").FontSize(style.Font(c, style.FontMeta)).TextColor(core.Tokens(c).TextMuted).Shrink(0)
		video := input.Select(c.Key("quality"), &entry.videoID, queueVideoOptions(play), input.SelectOptions{Label: "视频画质", Disabled: s.queueRequested[bvid]})
		video.Element.Width(style.DownloadStreamWidth).MinWidth(0).MinHeight(style.DownloadControlHeight).Height(style.DownloadControlHeight).Padding(0, style.SpaceSmall).FontSize(style.Font(c, style.FontButton)).Radius(style.RadiusControl)
	})
	audioOptions := queueAudioOptions(play)
	if len(audioOptions) > 0 {
		ui.Row(c.Key("audio-field")).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "音质").FontSize(style.Font(c, style.FontMeta)).TextColor(core.Tokens(c).TextMuted).Shrink(0)
			audio := input.Select(c.Key("audio"), &entry.audioID, audioOptions, input.SelectOptions{Label: "音频音质", Disabled: s.queueRequested[bvid]})
			audio.Element.Width(style.DownloadAudioWidth).MinWidth(0).MinHeight(style.DownloadControlHeight).Height(style.DownloadControlHeight).Padding(0, style.SpaceSmall).FontSize(style.Font(c, style.FontButton)).Radius(style.RadiusControl)
		})
	} else {
		caption(c, "无独立音轨")
	}
}

func (s *State) removeStaged(key string) {
	s.cancelPendingManualKey(key)
	if task, ok := s.queuedTask(key); ok {
		delete(s.queueProgress, queueProgressKey(task.Bvid, task.Cid))
	}
	delete(s.stagedKeys, key)
	delete(s.stagedSourceNames, key)
	delete(s.queueResolves, key)
	s.stagedUploads = slices.DeleteFunc(s.stagedUploads, func(v model.VListItem) bool { return v.Bvid == key })
	s.stagedFavorites = slices.DeleteFunc(s.stagedFavorites, func(v model.FavoriteMedias) bool { return favoriteBvid(v.Bvid, v.BvID) == key })
	s.stagedCollections = slices.DeleteFunc(s.stagedCollections, func(v model.CollectionItemMedias) bool { return v.Bvid == key })
	s.stagedSeasons = slices.DeleteFunc(s.stagedSeasons, func(v model.SeasonsArchivesItem) bool { return v.Bvid == key })
	s.stagedSeries = slices.DeleteFunc(s.stagedSeries, func(v model.SeriesArchivesItem) bool { return v.Bvid == key })
	s.stagedDynamics = slices.DeleteFunc(s.stagedDynamics, func(v model.DynamicArchiveItem) bool { return v.Bvid == key })
	s.stagedPlayHistory = slices.DeleteFunc(s.stagedPlayHistory, func(v model.PlayHistoryItem) bool { return v.History.Bvid == key })
	s.stagedManual = slices.DeleteFunc(s.stagedManual, func(v manualVideo) bool { return v.Key == key })
	delete(s.queueProgress, key)
	delete(s.queueExtraBusy, key)
}
func (s *State) clearStaged() {
	s.pendingManualDownloads = nil
	s.queueRequested = make(map[string]bool)
	s.stagedKeys = make(map[string]bool)
	s.stagedSourceNames = make(map[string]string)
	s.queueResolves = make(map[string]*queueResolve)
	s.stagedUploads = nil
	s.stagedFavorites = nil
	s.stagedCollections = nil
	s.stagedSeasons = nil
	s.stagedSeries = nil
	s.stagedDynamics = nil
	s.stagedPlayHistory = nil
	s.stagedManual = nil
	s.queueProgress = make(map[string]download.Progress)
	s.queueExtraBusy = make(map[string]string)
	s.notice = ""
}

func downloadRequestLabel(queued bool) string {
	if queued {
		return "已排队"
	}
	return "下载"
}
