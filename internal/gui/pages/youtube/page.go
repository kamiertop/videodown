// Package youtube owns the YouTube download form.
package youtube

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
)

func (s *State) Form(c *ui.Context, path string, scroll *ui.ScrollState) {
	// An edit's debounce may finish after another task started. Keep that edit
	// pending and commit it once network work is idle instead of losing it.
	if s.proxySavePending && !s.proxyBusy && !s.busy && !s.playlistLoading && s.detailsInFlight == 0 && s.readyCount() == 0 && s.PendingDownloadCount() == 0 {
		s.saveProxy()
	}
	if strings.HasSuffix(path, "/playlist") {
		s.playlistPage(c)
		return
	}
	if strings.HasSuffix(path, "/profile") {
		s.accountPage(c, scroll)
		return
	}
	if strings.HasSuffix(path, "/history") {
		s.historyPage(c, scroll)
		return
	}
	ui.Column(c).
		Label("YouTube 下载页面").
		Fill().
		MinWidth(0).
		MinHeight(0).
		Padding(style.PagePadding).
		Gap(style.SpaceMedium).
		Children(func() {
			k := c.Theme()
			ui.Row(c).Label("YouTube 视频输入栏").FillWidth().MinWidth(0).Padding(style.SpaceSmall).
				Gap(style.SpaceSmall).AlignItems(ui.Center).
				Radius(style.RadiusPanel).
				Border(style.BorderWidth, k.Border).
				Background(k.Surface).
				Children(func() {
					field, editor := roundedInput(c.Key("link"), &s.Link, "YouTube 视频链接", "粘贴 YouTube 视频或 Shorts 链接", s.busy, "清空视频链接")
					field.Grow(1).MinWidth(0)
					editor.OnSubmit(s.resolve)
					components.Button(c.Key("parse"), "解析视频", components.ButtonOptions{Variant: components.ButtonPrimary, Loading: s.busy, Disabled: s.busy || s.api == nil || strings.TrimSpace(s.Link) == ""}).Tooltip("也可在链接输入框按 Enter 解析").OnClick(s.resolve)
				})
			if s.message != "" {
				components.Muted(c, s.message)
			}
			s.downloadResults(c.Key("results"))
		})
}

func (s *State) downloadCard(c *ui.Context) {
	k := c.Theme()
	ui.Row(c).Label("YouTube 下载视频卡片").FillWidth().MinWidth(0).Padding(style.SpaceMedium).Gap(style.SpaceMedium).AlignItems(ui.Center).Radius(style.RadiusPanel).Border(style.BorderWidth, k.Border).Background(k.Surface).Children(func() {
		ui.Box(c).Label("YouTube 视频封面").Size(144, 81).Shrink(0).Clip().Radius(style.RadiusMedia).Background(k.Background).Center().Children(func() {
			if s.cover != nil {
				ui.Image(c, s.cover).Fill().Fit(ui.Cover)
			} else {
				if s.coverLoading {
					components.Spinner(c, components.SpinnerOptions{Size: 18, Label: "正在加载封面"})
				} else if s.coverError != "" {
					components.Button(c, "重试封面", components.ButtonOptions{Variant: components.ButtonGhost, Icon: icons.Must("refresh-cw")}).Tooltip(s.coverError).OnClick(s.loadCover)
				} else {
					ui.Icon(c, icons.Must("image")).Size(24, 24).TextColor(k.TextMuted)
				}
			}
		})
		ui.Column(c).Grow(1).MinWidth(0).MinHeight(81).Gap(4).Children(func() {
			ui.Column(c).Grow(1).FillWidth().MinWidth(0).MinHeight(0).Gap(4).Children(func() {
				ui.Text(c, s.video.Title).FillWidth().MaxLines(1).FontWeight(style.WeightSemibold).Tooltip(s.video.Title)
				ui.Text(c, fmt.Sprintf("%s · %s", s.video.Author, s.video.Duration.Round(1e9))).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
			})
			ui.Row(c.Key("actions")).FillWidth().MinWidth(0).Shrink(0).Wrap().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
				s.qualitySelect(c)
				ui.Box(c).Grow(1)
				controls := ui.Row(c.Key("buttons")).Shrink(0).Gap(style.SpaceSmall).AlignItems(ui.Center)
				controls.Children(func() {
					compactTaskButton(components.Button(c.Key("cover"), "封面", components.ButtonOptions{Label: "保存封面", Variant: components.ButtonGhost, Loading: s.coverSaving, Disabled: s.api == nil || s.runtime.Background == nil || s.coverSaving || s.video.Cover == ""}).OnClick(s.savePreviewCover))
					audio := components.Button(c.Key("audio-only"), "音频", components.ButtonOptions{Variant: components.ButtonGhost, Label: "解析卡片音频模式", Disabled: s.busy || len(s.video.AudioFormats) == 0}).Tooltip("仅下载音频，再次点击恢复视频").OnClick(func() { s.audioOnly = !s.audioOnly; s.selectDefaultFormat() })
					compactTaskButton(audio)
					if s.audioOnly {
						audio.Background(k.Selection).TextColor(k.Accent)
					}
					compactTaskButton(components.Button(c.Key("remove"), "移除", components.ButtonOptions{Variant: components.ButtonGhost}).OnClick(s.removePreview))
					compactTaskButton(components.Button(c.Key("download"), "下载", components.ButtonOptions{Variant: components.ButtonPrimary, Icon: icons.Must("download"), Disabled: s.busy || s.proxyBusy || s.api == nil || s.runtime.Background == nil || len(s.formats()) == 0}).OnClick(s.downloadPreview))
				})
			})
		})
	})
}

func transferStatus(p yt.Progress) string {
	text := p.Phase
	if p.Speed > 0 {
		text = byteSize(p.Speed) + fmt.Sprintf("/s · %.0f%% · ", p.Percent) + p.Phase
	} else {
		text += fmt.Sprintf(" · %.0f%%", p.Percent)
	}
	if p.Total > 0 {
		text += fmt.Sprintf(" · %s / %s", byteSize(float64(p.Downloaded)), byteSize(float64(p.Total)))
	}
	return text
}
func byteSize(bytes float64) string {
	for _, unit := range []string{"B", "KB", "MB", "GB"} {
		if bytes < 1024 || unit == "GB" {
			return fmt.Sprintf("%.1f %s", bytes, unit)
		}
		bytes /= 1024
	}
	return ""
}
