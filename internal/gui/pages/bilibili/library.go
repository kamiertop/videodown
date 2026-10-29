package bilibili

import (
	"fmt"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/ELY-Open-Source/MujicaUI/overlay"
	"github.com/egoist/mygo/ui"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

const historyPageSize = 50

func historyItemKey(item model.DownloadHistoryItem) string {
	return fmt.Sprintf("%d:%s", item.Cid, item.DownloadKind)
}

func (s *State) loadDownloadHistory(more bool) {
	if s.api == nil || s.historyMutating {
		return
	}
	api := s.api
	query := strings.TrimSpace(*s.search("history"))
	if more {
		if s.history.Loading || s.historyMore.Loading || s.historySearchPending || !s.history.Data.HasMore {
			return
		}
		offset := len(s.history.Data.Items)
		load(s, &s.historyMore, func() (biliapi.HistoryPage, error) {
			return api.DownloadHistoryPage(offset, historyPageSize, query)
		}, func(page biliapi.HistoryPage) {
			s.history.Data.Items = appendUnique(s.history.Data.Items, page.Items, historyItemKey)
			s.history.Data.Total, s.history.Data.HasMore = page.Total, page.HasMore && len(page.Items) > 0
		})
		return
	}
	s.historyQuery = query
	s.historySearchPending = false
	s.historyMore.seq++
	s.historyMore.Loading, s.historyMore.Error = false, ""
	s.resetAutoPage("download-history", "download-history-items")
	load(s, &s.history, func() (biliapi.HistoryPage, error) {
		return api.DownloadHistoryPage(0, historyPageSize, query)
	}, nil)
}

// Reject old pages immediately; wait for a pause in typing before querying SQLite.
func (s *State) scheduleHistorySearch() {
	query := strings.TrimSpace(*s.search("history"))
	if s.historyQuery == query {
		return
	}
	s.historyQuery, s.historySearchPending = query, true
	s.history.seq++
	s.history.Loading = false
	s.historyMore.seq++
	s.historyMore.Loading = false
	if s.runtime.After == nil {
		s.loadDownloadHistory(false)
		return
	}
	epoch, seq := s.epoch, s.history.seq
	s.runtime.After(300*time.Millisecond, func() {
		if s.stopped || s.epoch != epoch || s.history.seq != seq {
			return
		}
		s.loadDownloadHistory(false)
	})
}

func (s *State) changeHistory(work func() error) {
	if s.historyMutating || s.runtime.Run == nil {
		return
	}
	s.historyMutating = true
	s.historyActionError = ""
	s.history.seq++
	s.history.Loading = false
	s.historyMore.seq++
	s.historyMore.Loading = false
	epoch := s.epoch
	s.runtime.Run(work, func(err error) {
		if s.stopped || s.epoch != epoch {
			return
		}
		s.historyMutating = false
		if err != nil {
			s.historyActionError = err.Error()
			if s.historySearchPending {
				s.loadDownloadHistory(false)
			}
			return
		}
		s.historyClearOpen = false
		s.loadDownloadHistory(false)
	})
}

func (s *State) openHistoryFile(path string, directory bool) {
	action := s.runtime.OpenFile
	if directory {
		action = s.runtime.OpenLocation
	}
	if action == nil || s.runtime.Run == nil {
		return
	}
	epoch := s.epoch
	s.historyActionError = ""
	s.runtime.Run(func() error { return action(path) }, func(err error) {
		if s.stopped || s.epoch != epoch {
			return
		}
		if err != nil {
			s.historyActionError = err.Error()
		}
	})
}

func (s *State) downloadHistory(c *ui.Context) {
	s.scheduleHistorySearch()
	k := core.Tokens(c)
	panel(c.Key("toolbar"), "历史工具栏", func() {
		ui.Row(c).FillWidth().Height(style.LibraryHeaderHeight).Padding(style.SpaceSmall, style.SpaceMedium).
			Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("下载历史 · %d", s.history.Data.Total)).FontWeight(style.WeightSemibold).Shrink(0)
			textInput(c.Key("search"), s.search("history"), "搜索下载历史", "搜索标题、UP 主或 BV 号", func() { s.loadDownloadHistory(false) }).Grow(1).MinWidth(0)
			button(c.Key("refresh"), "刷新", input.ButtonOptions{Variant: input.Outline, Icon: icons.Must("refresh-cw"), Loading: s.history.Loading, Disabled: s.historyMutating || s.api == nil}).OnClick(func() { s.loadDownloadHistory(false) })
			button(c.Key("clear"), "清空", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("trash-2"), Disabled: s.historyMutating || s.history.Data.Total == 0 || s.api == nil}).OnClick(func() { s.historyClearOpen = true })
		})
	})
	s.errorRow(c, s.historyActionError, nil)
	if s.history.Loading || s.historySearchPending {
		ui.Text(c, "正在加载下载历史").FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
	}
	if s.history.Error != "" {
		s.errorRow(c, s.history.Error, func() { s.loadDownloadHistory(false) })
	}
	layout.ScrollArea(c.Key("history-scroll"), s.scroll("download-history-items"), layout.ScrollAreaOptions{}, func() {
		ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, style.SpaceMedium, 0).Gap(style.SpaceSmall).Children(func() {
			if len(s.history.Data.Items) == 0 && !s.history.Loading && !s.historySearchPending && s.history.Error == "" {
				message := "暂无下载历史"
				if s.historyQuery != "" {
					message = "没有匹配的下载历史"
				}
				empty(c, message, "history")
			}
			for _, item := range s.history.Data.Items {
				s.downloadHistoryCard(c.Key(historyItemKey(item)), item)
			}
			if s.historyMore.Loading {
				caption(c, "正在加载更多")
			}
			s.errorRow(c, s.historyMore.Error, func() { s.loadDownloadHistory(true) })
			if s.history.Data.HasMore && !s.historyMore.Loading && s.historyMore.Error == "" && s.autoPageFallback("download-history", "download-history-items") {
				button(c.Key("more"), "加载更多", input.ButtonOptions{Variant: input.Outline, Disabled: s.history.Loading || s.historyMutating || s.historySearchPending}).OnClick(func() { s.continueAutoPage("download-history", func() { s.loadDownloadHistory(true) }) })
			}
		})
	}).Label("下载历史列表").Grow(1).MinHeight(0).MinWidth(0)
	s.maybeAutoPage("download-history", "download-history-items", func() bool {
		return s.section == "history" && !s.history.Loading && !s.historyMore.Loading && !s.historyMutating && !s.historySearchPending && s.history.Error == "" && s.historyMore.Error == "" && s.history.Data.HasMore
	}, func() { s.loadDownloadHistory(true) })
	overlay.Dialog(c.Key("clear-dialog"), &s.historyClearOpen, overlay.DialogOptions{Title: "清空下载历史", Description: "清空所有下载记录？本地文件会保留。", KeepOnEscape: s.historyMutating, KeepOnBackdrop: true, HideClose: s.historyMutating}, func() {
		s.errorRow(c, s.historyActionError, nil)
		ui.Row(c).FillWidth().Gap(style.SpaceSmall).Justify(ui.End).Children(func() {
			button(c.Key("cancel"), "取消", input.ButtonOptions{Variant: input.Outline, Disabled: s.historyMutating}).OnClick(func() { s.historyClearOpen = false })
			button(c.Key("confirm"), "确认清空", input.ButtonOptions{Variant: input.Danger, Loading: s.historyMutating}).OnClick(func() {
				if s.api != nil {
					s.changeHistory(s.api.ClearDownloadHistory)
				}
			})
		})
	})
}

func (s *State) downloadHistoryCard(c *ui.Context, item model.DownloadHistoryItem) {
	k := core.Tokens(c)
	panel(c, "下载记录："+item.Title, func() {
		ui.Row(c).FillWidth().Padding(style.DownloadCardPadding).Gap(style.SpaceMedium).AlignItems(ui.Start).Children(func() {
			ui.Box(c.Key("cover")).Size(style.DownloadCoverWidth, style.DownloadCoverHeight).Clip().Radius(style.RadiusMedia).Children(func() {
				s.imageWithRadius(c, item.Cover, style.DownloadCoverWidth, style.DownloadCoverHeight, false, style.RadiusMedia)
				badge := duration(item.Duration)
				if item.DownloadKind == "audio" {
					badge = "音频"
				} else if item.DownloadKind == "cover" || item.Cid < 0 {
					badge = "封面"
				}
				ui.Text(c, badge).Absolute().End(4).Bottom(4).Padding(1, 4).Radius(style.RadiusChrome).FontSize(style.Font(c, style.FontMeta)).Background(ui.RGB(0, 0, 0).Alpha(style.MediaDurationOpacity)).TextColor(ui.RGB(255, 255, 255))
			})
			ui.Column(c.Key("details")).Grow(1).MinWidth(0).Gap(style.DownloadCardGap).Children(func() {
				ui.Row(c.Key("heading")).FillWidth().Gap(style.SpaceSmall).AlignItems(ui.Start).Children(func() {
					ui.Text(c, item.Title).Grow(1).MinWidth(0).MaxLines(2).FontWeight(style.WeightSemibold)

					ui.Row(c.Key("actions")).Gap(style.SpaceSmall).Children(func() {
						downloadCardButton(c.Key("open"), "本地打开", input.ButtonOptions{Variant: input.Primary, Icon: icons.Must("play"), Disabled: item.Path == "" || s.runtime.OpenFile == nil}).OnClick(func() { s.openHistoryFile(item.Path, false) })
						downloadCardButton(c.Key("location"), "打开目录", input.ButtonOptions{Variant: input.Outline, Icon: icons.Must("folder"), Disabled: item.Path == "" || s.runtime.OpenLocation == nil}).OnClick(func() { s.openHistoryFile(item.Path, true) })
						downloadCardButton(c.Key("delete"), "删除记录", input.ButtonOptions{Variant: input.Ghost, Disabled: s.historyMutating || s.api == nil}).OnClick(func() {
							api := s.api
							s.changeHistory(func() error { return api.DeleteDownloadHistory(item.Cid, item.DownloadKind) })
						})
					})
				})
				meta := []string{}
				if item.UpperName != "" {
					meta = append(meta, item.UpperName)
				}
				meta = append(meta, "下载于 "+historyDownloadedAt(item.Downloaded))
				ui.Text(c, strings.Join(meta, " · ")).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
				ui.Text(c, item.Path).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
			})
		})
	})
}

func historyDownloadedAt(raw string) string {
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		if raw == "" {
			return "未知时间"
		}
		return raw
	}
	return parsed.Local().Format(time.DateTime)
}
