package youtube

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
	newc "github.com/oneo-me/onevra/source/components"
)

// Keep the closed control short while the popup preserves every format detail.
func (s *State) qualitySelect(c *ui.Context) {
	width, _ := c.Size()
	compactWidth := float32(260)
	if width < 1000 {
		compactWidth = 180
	}
	qualitySelect(c, &s.format, formatOptions(s.video, s.audioOnly), compactWidth, "YouTube 下载画质", "YouTube 画质选项", s.busy || len(s.formats()) == 0)
}

func qualitySelect(c *ui.Context, value *int, formats []yt.Format, compactWidth float32, controlLabel, popupLabel string, disabled bool) ui.Element {
	sel := newc.Select(c, value)
	label, detail := "选择画质", ""
	if len(formats) == 0 {
		label = "没有可用格式"
	}
	for _, format := range formats {
		if format.ID == *value {
			label = format.Summary
			if label == "" {
				parts := strings.Split(format.Label, " · ")
				label = strings.Join(parts[:min(2, len(parts))], " · ")
			}
			detail = format.Label
			break
		}
	}
	sel.Trigger.Label(controlLabel).Width(compactWidth).MinWidth(0).MinHeight(style.DownloadControlHeight).Disabled(disabled).Tooltip(detail).Radius(style.RadiusControl)
	sel.Trigger.Children(func() { ui.Text(c, label).SingleLine().Grow(1).Shrink(1) })
	if disabled {
		return sel.Trigger
	}
	sel.Popup(func(panel ui.Element) {
		panel.Label(popupLabel).Width(compactWidth).Radius(style.RadiusMenu)
		ui.Scroll(c).FillWidth().MaxHeight(320).Children(func() {
			for _, format := range formats {
				row := sel.Item(format.ID)
				row.Label(format.Label).FillWidth().MinWidth(0).Padding(8).Gap(8).AlignItems(ui.Center).Radius(style.RadiusControl)
				heading, details := qualityLines(format.Label)
				row.Children(func() {
					ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
						ui.Text(c, heading).FillWidth().MaxLines(2).FontWeight(style.WeightSemibold)
						if details != "" {
							ui.Text(c, details).FillWidth().MaxLines(3).FontSize(style.Font(c, style.FontMeta)).TextColor(c.Theme().TextMuted)
						}
					})
				})
			}
		})
	})
	return sel.Trigger
}

func qualityLines(label string) (string, string) {
	parts := strings.Split(label, " · ")
	split := min(3, len(parts))
	return strings.Join(parts[:split], " · "), strings.Join(parts[split:], " · ")
}
