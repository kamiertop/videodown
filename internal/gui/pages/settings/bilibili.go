package settings

import (
	"fmt"

	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/preferences"
)

func (s *State) bilibiliSettings(c *ui.Context, w Services) {
	settingCard(c, fmt.Sprintf("同时解析数 [ %d ]", int(s.Config.ParseConcurrency)), "play", func() {
		components.Muted(c, "批量下载时并发请求播放地址的数量。越小越不容易触发风控，越大解析越快。")
		var marks []input.SliderMark
		for i := 1; i <= 5; i++ {
			marks = append(marks, input.SliderMark{Value: float64(i), Label: fmt.Sprint(i)})
		}
		result := input.Slider(c.Key("parse-count"),
			&s.Config.ParseConcurrency,
			input.SliderOptions{Label: "同时解析数", Min: 1, Max: 5, Step: 1, Marks: marks})
		if result.Changed() {
			saveSetting(w, preferences.Changes{ParseConcurrency: preferences.Value(int(s.Config.ParseConcurrency))})
		}
	})
	settingCard(c, "请求间隔随机休眠", "moon", func() {
		components.Muted(c, fmt.Sprintf("每个播放地址解析任务请求前随机等待 [0-%d] 秒，打散批量解析节奏，降低风控概率", int(s.Config.ParseSleep)))
		result := input.Slider(c.Key("parse-sleep"), &s.Config.ParseSleep, input.SliderOptions{Label: "解析请求间隔", Min: 0, Max: 120, Step: 1})
		if result.Changed() {
			saveSetting(w, preferences.Changes{ParseSleep: preferences.Value(int(s.Config.ParseSleep))})
		}
		presets(c, &s.Config.ParseSleep, []int{0, 1, 2, 3, 5, 10, 15, 20, 30, 60}, w, "ParseSleep", func(v float64) preferences.Changes { return preferences.Changes{ParseSleep: preferences.Value(int(v))} })
	})
}
