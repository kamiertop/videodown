// Package youtube owns the YouTube download form.
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

func (s *State) Form(c *ui.Context, path string, scroll *ui.ScrollState) {
	if strings.HasSuffix(path, "/playlist") {
		s.playlistPage(c, scroll)
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
			k := core.Tokens(c)
			ui.Row(c).Label("YouTube 视频输入栏").FillWidth().MinWidth(0).Padding(style.SpaceSmall).
				Gap(style.SpaceSmall).AlignItems(ui.Center).
				Radius(style.RadiusPanel).
				Border(style.BorderWidth, k.Border).
				Background(k.Surface).
				Children(func() {
					field := input.TextInput(c.Key("link"),
						&s.Link,
						input.TextInputOptions{Label: "YouTube 视频链接", Placeholder: "粘贴 YouTube 视频或 Shorts 链接", Disabled: s.busy},
					)
					field.Element.Grow(1).MinWidth(0)
					field.Input.OnSubmit(s.resolve)
					components.Button(c.Key("parse"), "解析视频", input.ButtonOptions{Variant: input.Primary, Loading: s.busy, Disabled: s.busy || s.api == nil || strings.TrimSpace(s.Link) == ""}).OnClick(s.resolve)
				})
			if s.message != "" {
				components.Muted(c, s.message)
			}
			layout.ScrollArea(c.Key("results"), scroll, layout.ScrollAreaOptions{}, func() {
				ui.Column(c).
					FillWidth().
					MinWidth(0).
					Padding(0, style.ScrollbarGutter, 0, 0).
					Gap(style.SpaceMedium).
					Children(func() {
						if s.video.ID != "" {
							s.downloadCard(c)
						}
						s.queueView(c)
					})
			}).Grow(1).MinWidth(0).MinHeight(0)
		})
}

func (s *State) downloadCard(c *ui.Context) {
	k := core.Tokens(c)
	ui.Row(c).Label("YouTube 下载视频卡片").FillWidth().MinWidth(0).Padding(style.SpaceMedium).Gap(style.SpaceMedium).AlignItems(ui.Center).Radius(style.RadiusPanel).Border(style.BorderWidth, k.Border).Background(k.Surface).Children(func() {
		ui.Box(c).Label("YouTube 视频封面").Size(144, 81).Shrink(0).Clip().Radius(style.RadiusMedia).Background(k.Background).Center().Children(func() {
			if s.cover != nil {
				ui.Image(c, s.cover).Fill().Fit(ui.Cover)
			} else {
				if s.coverLoading {
					core.Spinner(c, core.SpinnerOptions{Size: 18, Label: "正在加载封面"})
				} else if s.coverError != "" {
					components.Button(c, "重试封面", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("refresh-cw")}).Tooltip(s.coverError).OnClick(s.loadCover)
				} else {
					ui.Icon(c, icons.Must("image")).Size(24, 24).TextColor(k.TextMuted)
				}
			}
		})
		ui.Column(c).Grow(1).MinWidth(0).Height(81).Gap(4).Children(func() {
			ui.Column(c).Grow(1).FillWidth().MinWidth(0).MinHeight(0).Gap(4).Children(func() {
				ui.Text(c, s.video.Title).FillWidth().MaxLines(1).FontWeight(style.WeightSemibold).Tooltip(s.video.Title)
				ui.Text(c, fmt.Sprintf("%s · %s", s.video.Author, s.video.Duration.Round(1e9))).FillWidth().MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
			})
			ui.Row(c).FillWidth().MinWidth(0).Shrink(0).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
				s.qualitySelect(c)
				ui.Box(c).Grow(1)
				audio := components.Button(c.Key("audio-only"), "音频", input.ButtonOptions{Variant: input.Ghost, Label: "解析卡片音频模式", Disabled: s.busy || len(s.video.AudioFormats) == 0}).MinWidth(0).Padding(4, 8).Tooltip("仅下载音频，再次点击恢复视频").OnClick(func() { s.audioOnly = !s.audioOnly; s.selectDefaultFormat() })
				if s.audioOnly {
					audio.Background(k.Selection).TextColor(k.AccentText)
				}
				components.Button(c.Key("download"), "加入队列", input.ButtonOptions{Variant: input.Primary, Disabled: s.busy || len(s.formats()) == 0}).Shrink(0).Tooltip("加入待下载区，确认设置后手动开始").OnClick(s.stagePreview)
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
