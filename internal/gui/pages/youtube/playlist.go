package youtube

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
)

type playlistRow struct {
	Video    yt.Video
	Selected bool
}

func (s *State) loadPlaylist(more bool) {
	if s.stopped || s.api == nil || s.runtime.Background == nil {
		return
	}
	input, token := strings.TrimSpace(s.playlistLink), ""
	if more {
		if s.playlistLoading || s.playlist.Continuation == "" {
			return
		}
		input, token = s.playlistSource, s.playlist.Continuation
	}
	if s.playlistCancel != nil {
		s.playlistCancel()
	}
	s.playlistGeneration++
	generation := s.playlistGeneration
	s.playlistLoading, s.playlistError = true, ""
	if !more {
		s.playlistRows = nil
		s.playlistRemoved = nil
		s.playlist = yt.Playlist{}
		s.playlistSource = input
	}
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	s.playlistCancel = cancel
	api := s.api
	var result yt.Playlist
	s.runtime.Background(func() error { defer cancel(); var err error; result, err = api.Playlist(ctx, input, token); return err }, func(err error) {
		if s.stopped || s.playlistGeneration != generation {
			return
		}
		s.playlistLoading = false
		if err != nil {
			s.playlistError = err.Error()
			return
		}
		s.playlist = result
		seen := map[string]bool{}
		for _, row := range s.playlistRows {
			seen[row.Video.ID] = true
		}
		for _, video := range result.Videos {
			if seen[video.ID] || s.playlistRemoved[video.ID] {
				continue
			}
			seen[video.ID] = true
			row := &playlistRow{Video: video}
			row.Selected = false
			s.playlistRows = append(s.playlistRows, row)
		}
	})
}

// Clearing the parsed result invalidates pending pagination, preserving download tasks.
func (s *State) clearPlaylist() {
	if s.playlistCancel != nil {
		s.playlistCancel()
		s.playlistCancel = nil
	}
	s.playlistGeneration++
	s.playlistRows = nil
	s.playlistRemoved = nil
	s.playlist = yt.Playlist{}
	s.playlistSource, s.playlistError = "", ""
	s.playlistLoading = false
}

func (s *State) removePlaylistRows(remove func(*playlistRow) bool) {
	if s.playlistRemoved == nil {
		s.playlistRemoved = map[string]bool{}
	}
	s.playlistRows = slices.DeleteFunc(s.playlistRows, func(row *playlistRow) bool {
		if !remove(row) {
			return false
		}
		s.playlistRemoved[row.Video.ID] = true
		return true
	})
	s.playlistError = ""
	if len(s.playlistRows) == 0 {
		s.clearPlaylist()
	}
}

func (s *State) enqueuePlaylist() {
	added := 0
	for _, row := range s.playlistRows {
		if !row.Selected {
			continue
		}
		row.Selected = false
		if !s.playlistAvailable(row) {
			continue
		}
		if s.enqueueGrouped(row.Video, 0, false, "playlist:"+s.playlist.ID, s.playlist.Title, false) {
			added++
			row.Selected = false
		}
	}
	s.playlistError = fmt.Sprintf("已加入 %d 个视频，请到下载页选择画质后开始下载", added)
}

func (s *State) playlistPage(c *ui.Context, scroll *ui.ScrollState) {
	ui.Column(c).Label("YouTube 播放列表页面").Fill().MinWidth(0).MinHeight(0).Padding(style.PagePadding).Gap(style.SpaceMedium).Children(func() {
		panel(c.Key("playlist-input"), "YouTube 播放列表输入栏", func() {
			ui.Row(c).FillWidth().MinWidth(0).Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
				field := input.TextInput(c.Key("link"), &s.playlistLink, input.TextInputOptions{Label: "YouTube 播放列表链接", Placeholder: "粘贴包含 list= 的 YouTube 播放列表链接"})
				field.Element.Grow(1).MinWidth(0)
				field.Input.OnSubmit(func() { s.loadPlaylist(false) })
				components.Button(c.Key("parse"), "解析列表", input.ButtonOptions{Variant: input.Primary, Loading: s.playlistLoading, Disabled: s.api == nil || strings.TrimSpace(s.playlistLink) == "" || s.playlistLoading}).OnClick(func() { s.loadPlaylist(false) })
			})
		})

		selected, addable := 0, 0
		for _, row := range s.playlistRows {

			if row.Selected {
				selected++
				if s.playlistAvailable(row) {
					addable++
				}
			}
		}
		if len(s.playlistRows) > 0 {
			panel(c.Key("playlist-toolbar"), "YouTube 播放列表工具栏", func() {
				ui.Row(c).FillWidth().MinWidth(0).Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
					title := s.playlist.Title
					if title == "" {
						title = "播放列表"
					}
					ui.Text(c, title).Label("YouTube 播放列表标题").Grow(1).MinWidth(0).MaxLines(1).FontWeight(style.WeightSemibold).Tooltip(title)
					allSelected := selected == len(s.playlistRows)
					selectLabel := "全选"
					if allSelected {
						selectLabel = "取消全选"
					}
					components.Button(c.Key("all"), selectLabel, input.ButtonOptions{Variant: input.Outline, Disabled: len(s.playlistRows) == 0}).OnClick(func() {
						for _, row := range s.playlistRows {
							row.Selected = !allSelected
						}
					})
					components.Button(c.Key("remove-selected"), "移除已选", input.ButtonOptions{Variant: input.Ghost, Disabled: selected == 0}).OnClick(func() { s.removePlaylistRows(func(row *playlistRow) bool { return row.Selected }) })
					components.Button(c.Key("clear"), "清空列表", input.ButtonOptions{Variant: input.Ghost}).OnClick(s.clearPlaylist)
					ui.Text(c, fmt.Sprintf("%d / %d", selected, len(s.playlistRows))).Shrink(0).MaxLines(1).TextColor(core.Tokens(c).TextMuted)
					components.Button(c.Key("add"), "加入队列", input.ButtonOptions{Variant: input.Primary, Disabled: addable == 0 || s.api == nil}).OnClick(s.enqueuePlaylist)
				})
			})
		}
		if s.playlistError != "" {
			components.Muted(c, s.playlistError)
		}
		layout.ScrollArea(c.Key("playlist-scroll"), scroll, layout.ScrollAreaOptions{}, func() {
			ui.Column(c).FillWidth().MinWidth(0).Padding(0, style.ScrollbarGutter, 0, 0).Gap(style.SpaceSmall).Children(func() {
				if len(s.playlistRows) == 0 && !s.playlistLoading && s.playlist.Title != "" {
					components.Muted(c, "暂无可下载的视频，私密或已删除的视频会跳过")
				}
				for _, row := range s.playlistRows {
					s.playlistCard(c.Key(row.Video.ID), row)
				}
				if len(s.playlistRows) > 0 && s.playlist.Continuation != "" {
					components.Button(c.Key("more"), "加载更多", input.ButtonOptions{Variant: input.Outline, Loading: s.playlistLoading, Disabled: s.playlistLoading}).OnClick(func() { s.loadPlaylist(true) })
				}
			})
		}).Grow(1).MinWidth(0).MinHeight(0)
	})
}

func (s *State) playlistCard(c *ui.Context, row *playlistRow) {
	k := core.Tokens(c)
	card := ui.ButtonBase(c).Row().Label("选择视频："+row.Video.ID).FillWidth().MinWidth(0).Padding(style.SpaceMedium).Gap(style.SpaceMedium).AlignItems(ui.Center).Radius(style.RadiusPanel).Background(k.Surface).Border(style.BorderWidth, k.Border).Cursor(ui.CursorPointer)
	if row.Selected {
		card.Background(k.Selection)
		card.DrawOver(func(p *ui.Painter, r ui.Rect) { p.Stroke(r, k.Accent, style.RadiusPanel, style.BorderEmphasis) })
	} else if card.Hovered() || card.FocusVisible() {
		card.Background(k.SurfaceHover)
	}
	card.Children(func() {
		s.thumbnailFrame(c.Key("cover"), row.Video, 144, 81, "列表封面："+row.Video.ID)
		ui.Column(c).Grow(1).MinWidth(0).Height(81).Gap(style.SpaceSmall).Children(func() {
			ui.Text(c, row.Video.Title).FillWidth().MaxLines(2).FontWeight(style.WeightSemibold).Tooltip(row.Video.Title)
			ui.Box(c).Grow(1)
			ui.Text(c, fmt.Sprintf("%s · %s", row.Video.Author, row.Video.Duration.Round(time.Second))).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
		})
		ui.Column(c).Shrink(0).Gap(style.SpaceSmall).AlignItems(ui.End).Children(func() {
			if label := s.playlistTaskLabel(row); label != "" {
				ui.Text(c, label).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
			}
			components.Button(c.Key("remove"), "移除", input.ButtonOptions{Variant: input.Ghost, Label: "移除列表视频：" + row.Video.ID}).MinWidth(0).Padding(4, 8).OnClick(func() {
				s.removePlaylistRows(func(item *playlistRow) bool { return item == row })
			}).Tooltip("从当前解析列表移除")
		})
	}).OnClick(func() {
		row.Selected = !row.Selected
	})
}
