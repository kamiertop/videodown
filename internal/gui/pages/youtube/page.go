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
					components.Button(c.Key("parse"), "解析视频", input.ButtonOptions{Variant: input.Primary, Loading: s.busy && !s.downloading, Disabled: s.busy || s.api == nil || strings.TrimSpace(s.Link) == ""}).OnClick(s.resolve)
				})
			if s.message != "" {
				components.Muted(c, s.message)
			}
			layout.ScrollArea(c.Key("results"), scroll, layout.ScrollAreaOptions{}, func() {
				ui.Column(c).
					FillWidth().
					MinWidth(0).
					Gap(style.SpaceMedium).
					Children(func() {
						if s.video.ID != "" {
							s.downloadCard(c)
						}
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
		ui.Column(c).
			Grow(1).
			MinWidth(0).
			Height(81).
			Gap(style.SpaceSmall).
			Children(func() {
				ui.Text(c, s.video.Title).FillWidth().MaxLines(2).FontWeight(style.WeightSemibold).Tooltip(s.video.Title)
				ui.Box(c).Grow(1)
				if s.downloading {
					text := transferStatus(s.progress)
					ui.Text(c, text).FillWidth().MinWidth(0).MaxLines(1).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted).Tooltip(text)
				} else {
					ui.Text(c, fmt.Sprintf("%s · %s", s.video.Author, s.video.Duration.Round(1e9))).
						FillWidth().
						MinWidth(0).MaxLines(1).
						FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
				}

			})
		ui.Row(c).Shrink(0).Gap(style.SpaceSmall).AlignItems(ui.Center).
			Children(func() {
				s.qualitySelect(c)
				if s.downloading {
					components.Button(c.Key("cancel"), "取消下载", input.ButtonOptions{Variant: input.Outline}).Shrink(0).OnClick(func() {
						if s.cancel != nil {
							s.cancel()
						}
					})
				} else {
					components.Button(c.Key("download"), "下载视频", input.ButtonOptions{Variant: input.Primary, Disabled: s.busy}).
						Shrink(0).
						OnClick(s.download)
				}
				if s.output != "" && s.runtime.OpenLocation != nil {
					components.Button(c.Key("open"), "", input.ButtonOptions{Variant: input.Outline, Icon: icons.Must("folder"), Label: "打开所在目录"}).
						Size(32, 32).MinWidth(32).Padding(0).
						Shrink(0).Tooltip("打开所在目录").
						OnClick(func() { _ = s.runtime.OpenLocation(s.output) })
				}
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
