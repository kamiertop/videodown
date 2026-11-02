package settings

import (
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/preferences"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) generalSettings(c *ui.Context, w Services) {
	settingCard(c, "下载路径", "folder", func() {
		ui.Row(c).FillWidth().Gap(8).Children(func() {
			input.TextInput(c.Key("directory"), &s.Config.Directory, input.TextInputOptions{Label: "下载目录", Placeholder: "选择下载路径", ReadOnly: true}).Element.Grow(1).MinWidth(0)
			components.Button(c.Key("browse"), "选择目录", input.ButtonOptions{Disabled: !w.SettingsAvailable || w.BrowsingSettings}).OnClick(func() {
				if w.BrowseDirectory != nil {
					w.BrowseDirectory()
				}
			})
		})
	})
	settingCard(c, "界面主题", "palette", func() {
		ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "主题模式").Grow(1)
			choice := input.Select(
				c.Key("theme"),
				&s.Config.Theme,
				[]input.SelectOption[string]{
					{Value: "dark", Label: "dark - 深色模式"},
					{Value: "light", Label: "light - 浅色模式"},
					{Value: "cupcake", Label: "cupcake - 纸杯蛋糕"},
					{Value: "caramellatte", Label: "caramellatte 焦糖"},
					{Value: "system", Label: "跟随系统"},
				},
				input.SelectOptions{Label: "界面主题"},
			)
			choice.Element.Radius(style.RadiusMenu)
			if choice.Changed() {
				s.ApplyTheme()
				saveSetting(w, preferences.Changes{Theme: preferences.Value(s.Config.Theme)})
			}
		})
		components.Muted(c, "当前主题："+s.Config.Theme)
	})
	toggleCard(c, "自动更新", "refresh-cw", "启动时检查新版本", "在后台检查 GitHub 是否有可用更新", &s.Config.AutoUpdate, w, func(v bool) preferences.Changes { return preferences.Changes{AutoUpdate: preferences.Value(v)} })
	settingCard(c, "窗口关闭行为", "settings", func() {
		value := w.CloseToTray
		insetToggle(c, "关闭时缩小到任务栏托盘", "关闭时隐藏窗口到系统托盘，下载不受影响", &value, w.SavingClose || !w.TrayAvailable || w.SaveClosePreference == nil, func() { w.SaveClosePreference(value) })
		if !w.CloseConfigured {
			components.Muted(c, "当前未设置，首次关闭窗口时仍会询问。")
		}
		if w.CloseError != "" {
			components.Muted(c, w.CloseError)
		}
		if w.TrayError != "" {
			components.Muted(c, w.TrayError)
		}
	})
}
