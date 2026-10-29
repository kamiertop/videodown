package youtube

import (
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Keep the closed control short while the popup preserves every format detail.
func (s *State) qualitySelect(c *ui.Context) {
	width, _ := c.Size()
	compactWidth := float32(260)
	if width < 1000 {
		compactWidth = 180
	}
	selectBox := ui.SelectBase(c.Key("quality"), &s.format)
	label, detail := "选择画质", ""
	for _, format := range s.video.Formats {
		if format.ID == s.format {
			label = format.Summary
			if label == "" {
				parts := strings.Split(format.Label, " · ")
				label = strings.Join(parts[:min(2, len(parts))], " · ")
			}
			detail = format.Label
			break
		}
	}
	selectBox.Trigger.Label("YouTube 下载画质").Width(compactWidth).Disabled(s.busy).Tooltip(detail)
	input.SelectField(c, selectBox.Trigger, selectBox.Open(), func() { ui.Text(c, label).SingleLine().Grow(1).Shrink(1) })
	selectBox.Trigger.Radius(style.RadiusControl)
	if s.busy {
		return
	}
	selectBox.Popup(func(panel ui.Element) {
		input.SelectPanel(c, panel)
		panel.Label("YouTube 画质选项").Width(compactWidth).Radius(style.RadiusMenu)
		ui.Scroll(c).FillWidth().MaxHeight(320).Children(func() {
			for _, format := range s.video.Formats {
				row := selectBox.Item(format.ID)
				k := core.Tokens(c)
				chosen := format.ID == s.format
				row.Label(format.Label).FillWidth().MinWidth(0).Padding(8).Gap(8).AlignItems(ui.Center).Radius(style.RadiusControl)
				if chosen {
					row.Background(k.Selection).TextColor(k.AccentText)
				} else if row.Highlighted() {
					row.Background(k.SurfaceHover)
				}
				heading, details := qualityLines(format.Label)
				row.Children(func() {
					input.SelectCheck(c, chosen)
					ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
						ui.Text(c, heading).FillWidth().MaxLines(2).FontWeight(style.WeightSemibold)
						if details != "" {
							ui.Text(c, details).FillWidth().MaxLines(3).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
						}
					})
				})
			}
		})
	})
}

func qualityLines(label string) (string, string) {
	parts := strings.Split(label, " · ")
	split := min(3, len(parts))
	return strings.Join(parts[:split], " · "), strings.Join(parts[split:], " · ")
}
