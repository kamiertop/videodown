package settings

import (
	"fmt"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/preferences"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) downloadSettings(c *ui.Context, w Services) {
	settingCard(c, "并发数量", "zap", func() {
		components.Muted(c, fmt.Sprintf("同时下载视频的数量 [ %d ]", int(s.Config.Concurrency)))
		var marks []input.SliderMark
		for i := 1; i <= 10; i++ {
			marks = append(marks, input.SliderMark{Value: float64(i), Label: fmt.Sprint(i)})
		}
		slider := input.Slider(c.Key("concurrency"), &s.Config.Concurrency, input.SliderOptions{Label: "并发下载数量", Min: 1, Max: 10, Step: 1, Marks: marks})
		if slider.Changed() {
			saveSetting(w, preferences.Changes{Concurrency: preferences.Value(int(s.Config.Concurrency))})
		}
	})
	settingCard(c, "下载后休眠", "moon", func() {
		components.Muted(c, fmt.Sprintf("每个视频下载完成后随机休眠 [0-%d] 秒", int(s.Config.Sleep)))
		ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				sliderValue := min(s.Config.Sleep, 600)
				result := input.Slider(c.Key("sleep"), &sliderValue, input.SliderOptions{Label: "下载后休眠滑块", Min: 0, Max: 600, Step: 5})
				if result.Changed() {
					s.Config.Sleep = sliderValue
					saveSetting(w, preferences.Changes{Sleep: preferences.Value(int64(s.Config.Sleep))})
				}
			})
			ui.Column(c).Width(132).Shrink(0).Children(func() {
				field := input.NumberInput(c.Key("seconds"), &s.Config.Sleep, input.NumberInputOptions{Label: "下载后休眠秒数", Min: 0, Max: 3600, Step: 1})
				if field.Changed() {
					saveSetting(w, preferences.Changes{Sleep: preferences.Value(int64(s.Config.Sleep))})
				}
			})
			ui.Text(c, "秒").Shrink(0)
		})
		presets(c, &s.Config.Sleep, []int{0, 10, 30, 60, 120, 300}, w, "Sleep", func(v float64) preferences.Changes { return preferences.Changes{Sleep: preferences.Value(int64(v))} })
	})
	toggleCard(c, "增量下载（跳过已下载）", "download", "一键下载全部时自动跳过已下载内容", "B 站 / 抖音跳过历史中已存在且文件仍在的视频；文件删除后会重新下载。关闭后恢复全量下载。", &s.Config.SkipDownloaded, w, func(v bool) preferences.Changes { return preferences.Changes{SkipDownloaded: preferences.Value(v)} })
}
func presets(c *ui.Context, value *float64, values []int, w Services, key string, changes func(float64) preferences.Changes) {
	columns := len(values)
	if columns > 6 {
		columns = 5
	}
	k := core.Tokens(c)
	// Compose the library's individual Toggles so each item's radius can be
	// styled; ToggleGroup currently exposes only its outer element.
	ui.Grid(c.Key(key+"-presets")).Columns(columns).FillWidth().AlignSelf(ui.Stretch).
		Radius(style.RadiusMenu).Border(style.BorderWidth, k.ControlBorder).Background(k.Surface).Padding(4).Gap(4).
		Role(ui.RoleGroup).Label("休眠快捷选项").FocusGroup(ui.Horizontal).Children(func() {
		for _, n := range values {
			label := fmt.Sprintf("%d 秒", n)
			if n == 0 {
				label = "不休眠"
			}
			on := *value == float64(n)
			input.Toggle(c.Key(n), &on, label, input.ToggleOptions{Label: label}).
				FillWidth().
				MinWidth(0).
				Radius(style.RadiusChoice).
				BorderWidth(0).
				OnClick(func() {
					if *value != float64(n) {
						*value = float64(n)
						saveSetting(w, changes(*value))
					}
				})
		}
	})
}
