package bilibili

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/style"
)

const historyPageSize = 50

func historyItemKey(item model.DownloadHistoryItem) string {
	return fmt.Sprintf("%d:%s", item.Cid, item.DownloadKind)
}

// historyState 是下载历史板块：首页与增量页两个资源、搜索词、变更互斥与
// 清空确认框。actionError 是无 toast 通道时的页面内回退提示。
type historyState struct {
	page, more    resource[biliapi.HistoryPage]
	query         string
	searchPending bool
	mutating      bool
	clearOpen     bool
	actionError   string
}

func (s *State) loadDownloadHistory(more bool) {
	if s.api == nil || s.history.mutating {
		return
	}
	api := s.api
	query := strings.TrimSpace(*s.search("history"))
	if more {
		if s.history.page.Loading || s.history.more.Loading || s.history.searchPending || !s.history.page.Data.HasMore {
			return
		}
		offset := len(s.history.page.Data.Items)
		load(s, &s.history.more, func() (biliapi.HistoryPage, error) {
			return api.DownloadHistoryPage(offset, historyPageSize, query)
		}, func(page biliapi.HistoryPage) {
			s.history.page.Data.Items = appendUnique(s.history.page.Data.Items, page.Items, historyItemKey)
			s.history.page.Data.Total, s.history.page.Data.HasMore = page.Total, page.HasMore && len(page.Items) > 0
		})
		return
	}
	s.history.query = query
	s.history.searchPending = false
	s.history.more.seq++
	s.history.more.Loading, s.history.more.Error = false, ""
	s.resetAutoPage("download-history", "download-history-items")
	load(s, &s.history.page, func() (biliapi.HistoryPage, error) {
		return api.DownloadHistoryPage(0, historyPageSize, query)
	}, nil)
}

// Reject old pages immediately; wait for a pause in typing before querying SQLite.
func (s *State) scheduleHistorySearch() {
	query := strings.TrimSpace(*s.search("history"))
	if s.history.query == query {
		return
	}
	s.history.query, s.history.searchPending = query, true
	s.history.page.seq++
	s.history.page.Loading = false
	s.history.more.seq++
	s.history.more.Loading = false
	if s.runtime.After == nil {
		s.loadDownloadHistory(false)
		return
	}
	epoch, seq := s.epoch, s.history.page.seq
	s.runtime.After(300*time.Millisecond, func() {
		if s.stopped || s.epoch != epoch || s.history.page.seq != seq {
			return
		}
		s.loadDownloadHistory(false)
	})
}

func (s *State) changeHistory(work func() error) {
	if s.history.mutating || s.runtime.Run == nil {
		return
	}
	s.history.mutating = true
	s.history.actionError = ""
	s.history.page.seq++
	s.history.page.Loading = false
	s.history.more.seq++
	s.history.more.Loading = false
	epoch := s.epoch
	s.runtime.Run(work, func(err error) {
		if s.stopped || s.epoch != epoch {
			return
		}
		s.history.mutating = false
		if err != nil {
			s.history.actionError = err.Error()
			if s.history.searchPending {
				s.loadDownloadHistory(false)
			}
			return
		}
		s.history.clearOpen = false
		s.loadDownloadHistory(false)
	})
}

// reportActionError 把打开文件等操作失败以 toast 弹出（右下角）；
// 没有 toast 通道的环境（测试、离屏渲染）退回页面内红字提示。
func (s *State) reportActionError(message string) {
	if s.runtime.Notify != nil {
		s.runtime.Notify(message, true)
		return
	}
	s.history.actionError = message
}

// openHistoryFile 打开下载文件或其所在目录；文件被移动或删除时弹出提示，
// 不再静默失败。
func (s *State) openHistoryFile(path string, directory bool) {
	action := s.runtime.OpenFile
	if directory {
		action = s.runtime.OpenLocation
	}
	if action == nil || s.runtime.Run == nil {
		return
	}
	if path == "" {
		s.reportActionError("打开失败：文件路径为空")
		return
	}
	epoch := s.epoch
	s.history.actionError = ""
	s.runtime.Run(func() error { return action(path) }, func(err error) {
		if s.stopped || s.epoch != epoch {
			return
		}
		if err != nil {
			s.reportActionError("打开失败：" + err.Error())
		}
	})
}

func (s *State) downloadHistory(c *ui.Context) {
	s.scheduleHistorySearch()
	k := c.Theme()
	panel(c.Key("toolbar"), "历史工具栏", func() {
		ui.Row(c).FillWidth().Height(style.LibraryHeaderHeight).Padding(style.SpaceSmall, style.SpaceMedium).
			Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("下载历史 · %d", s.history.page.Data.Total)).FontWeight(style.WeightSemibold).Shrink(0)
			textInput(c.Key("search"), s.search("history"), "搜索下载历史", "搜索标题、UP 主或 BV 号", func() { s.loadDownloadHistory(false) }, true).Grow(1).MinWidth(0)
			button(c.Key("refresh"), "刷新", components.ButtonOptions{Variant: components.ButtonOutline, Icon: icons.Must("refresh-cw"), Loading: s.history.page.Loading, Disabled: s.history.mutating || s.api == nil}).OnClick(func() { s.loadDownloadHistory(false) })
			button(c.Key("clear"), "清空", components.ButtonOptions{Variant: components.ButtonGhost, Icon: icons.Must("trash-2"), Disabled: s.history.mutating || s.history.page.Data.Total == 0 || s.api == nil}).OnClick(func() { s.history.clearOpen = true })
		})
	})
	s.errorRow(c, s.history.actionError, nil)
	if s.history.page.Loading || s.history.searchPending {
		ui.Text(c, "正在加载下载历史").FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
	}
	if s.history.page.Error != "" {
		s.errorRow(c, s.history.page.Error, func() { s.loadDownloadHistory(false) })
	}
	components.ScrollArea(c.Key("history-scroll"), s.scroll("download-history-items"), components.ScrollAreaOptions{}, func() {
		ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, style.SpaceMedium, 0).Gap(style.SpaceSmall).Children(func() {
			if len(s.history.page.Data.Items) == 0 && !s.history.page.Loading && !s.history.searchPending && s.history.page.Error == "" {
				message := "暂无下载历史"
				if s.history.query != "" {
					message = "没有匹配的下载历史"
				}
				empty(c, message, "history")
			}
			for _, item := range s.history.page.Data.Items {
				s.downloadHistoryCard(c.Key(historyItemKey(item)), item)
			}
			if s.history.more.Loading {
				caption(c, "正在加载更多")
			}
			s.errorRow(c, s.history.more.Error, func() { s.loadDownloadHistory(true) })
			if s.history.page.Data.HasMore && !s.history.more.Loading && s.history.more.Error == "" && s.autoPageFallback("download-history", "download-history-items") {
				button(c.Key("more"), "加载更多", components.ButtonOptions{Variant: components.ButtonOutline, Disabled: s.history.page.Loading || s.history.mutating || s.history.searchPending}).OnClick(func() { s.continueAutoPage("download-history", func() { s.loadDownloadHistory(true) }) })
			}
		})
	}).Label("下载历史列表").Grow(1).MinHeight(0).MinWidth(0)
	s.maybeAutoPage("download-history", "download-history-items", func() bool {
		return s.section == "history" && !s.history.page.Loading && !s.history.more.Loading && !s.history.mutating && !s.history.searchPending && s.history.page.Error == "" && s.history.more.Error == "" && s.history.page.Data.HasMore
	}, func() { s.loadDownloadHistory(true) })
	components.Dialog(c.Key("clear-dialog"), &s.history.clearOpen, components.DialogOptions{Title: "清空下载历史", Description: "清空所有下载记录？本地文件会保留。", KeepOnEscape: s.history.mutating, KeepOnBackdrop: true, HideClose: s.history.mutating}, func() {
		s.errorRow(c, s.history.actionError, nil)
		ui.Row(c).FillWidth().Gap(style.SpaceSmall).Justify(ui.End).Children(func() {
			button(c.Key("cancel"), "取消", components.ButtonOptions{Variant: components.ButtonOutline, Disabled: s.history.mutating}).OnClick(func() { s.history.clearOpen = false })
			button(c.Key("confirm"), "确认清空", components.ButtonOptions{Variant: components.ButtonDanger, Loading: s.history.mutating}).OnClick(func() {
				if s.api != nil {
					s.changeHistory(s.api.ClearDownloadHistory)
				}
			})
		})
	})
}

func (s *State) downloadHistoryCard(c *ui.Context, item model.DownloadHistoryItem) {
	k := c.Theme()
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
						downloadCardButton(c.Key("open"), "本地打开", components.ButtonOptions{Variant: components.ButtonPrimary, Icon: icons.Must("play"), Disabled: item.Path == "" || s.runtime.OpenFile == nil}).OnClick(func() { s.openHistoryFile(item.Path, false) })
						downloadCardButton(c.Key("location"), "打开目录", components.ButtonOptions{Variant: components.ButtonOutline, Icon: icons.Must("folder"), Disabled: item.Path == "" || s.runtime.OpenLocation == nil}).OnClick(func() { s.openHistoryFile(item.Path, true) })
						downloadCardButton(c.Key("delete"), "删除记录", components.ButtonOptions{Variant: components.ButtonGhost, Disabled: s.history.mutating || s.api == nil}).OnClick(func() {
							api := s.api
							s.changeHistory(func() error { return api.DeleteDownloadHistory(item.Cid, item.DownloadKind) })
						})
					})
				})
				var meta []string
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
