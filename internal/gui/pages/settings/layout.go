package settings

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"github.com/kamiertop/videodown/internal/gui/preferences"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Layout uses a fixed category rail
// and an independently scrolling form, with a centered maximum content width.
func (s *State) Layout(c *ui.Context, p navigation.Page, path string, navigate func(string), content func()) {
	k := core.Tokens(c)
	width, _ := c.Size()
	compact := width < style.CompactBreakpoint
	root := ui.Row(c.Key("settings-layout")).Fill().MinHeight(0).AlignItems(ui.Stretch)
	if compact {
		root.Column()
	}
	root.Children(func() {
		nav := ui.Column(c).Label("设置分类").Shrink(0).Padding(12).Gap(4).BorderColor(k.Border)
		if compact {
			nav.FillWidth().Height(76).BorderWidth(0, 0, style.BorderWidth, 0)
		} else {
			nav.Width(style.SettingsNavWidth).BorderWidth(0, style.BorderWidth, 0, 0)
		}
		nav.Children(func() {
			build := func() {
				names := []string{"settings", "play", "download", "file-text", "video", "info"}
				for i, item := range p.Sections() {
					target := p.Path()
					if item.ID != "" {
						target += "/" + item.ID
					}
					icon, ok := icons.Get(names[i])
					if !ok {
						icon = icons.Must("settings")
					}
					active := p.CurrentSection(path) == item.ID
					button := components.MenuButton(c.Key(item.ID), item.Label, icon, active).Justify(ui.Start).Height(44).OnClick(func() { navigate(target) })
					if !compact {
						button.FillWidth()
					}
				}
			}
			if compact {
				ui.ScrollHorizontal(c).FillWidth().Children(func() { ui.Row(c).Gap(4).Children(build) })
			} else {
				build()
			}
		})
		content()
	})
}

// Form owns only the scrolling content of a settings subroute.
func (s *State) Form(c *ui.Context, p navigation.Page, path string, window Services, scroll *ui.ScrollState) {
	layout.ScrollArea(c, scroll, layout.ScrollAreaOptions{}, func() {
		ui.Column(c).FillWidth().Padding(24).AlignItems(ui.Center).Children(func() {
			maximum := float32(672)
			id := p.CurrentSection(path)
			if id == "filename" {
				maximum = 896
			} else if id == "about" {
				maximum = 512
			}
			ui.Column(c).Label("设置表单").FillWidth().MaxWidth(maximum).Gap(24).Children(func() {
				s.settingsPage(c, id, window)
			})
		})
	}).Grow(1).MinWidth(0).MinHeight(0)
}

func (s *State) settingsPage(c *ui.Context, id string, w Services) {
	switch id {
	case "":
		s.generalSettings(c, w)
	case "download":
		s.downloadSettings(c, w)
	case "bilibili":
		s.bilibiliSettings(c, w)
	case "filename":
		s.filenameSettings(c, w)
	case "ffmpeg":
		s.ffmpegSettings(c, w)
	case "about":
		s.aboutSettings(c)
	}
}
func settingCard(c *ui.Context, title, icon string, body func()) {
	k := core.Tokens(c)
	ui.Column(c).FillWidth().Padding(24).Gap(16).Radius(style.RadiusPanel).Background(k.Surface).Border(style.BorderWidth, k.Border.Alpha(.6)).Children(func() {
		if title != "" {
			ui.Row(c).FillWidth().Gap(10).AlignItems(ui.Center).Children(func() {
				svg, ok := icons.Get(icon)
				if !ok {
					svg = icons.Must("settings")
				}
				ui.Icon(c, svg).Size(20, 20).TextColor(k.AccentText)
				ui.Text(c, title).FontSize(style.Font(c, style.FontHeading)).FontWeight(style.WeightSemibold)
			})
		}
		body()
	})
}
func saveSetting(w Services, changes preferences.Changes) {
	if w.SaveSetting != nil {
		w.SaveSetting(changes)
	}
}
func toggleCard(c *ui.Context, title, icon, label, description string, value *bool, w Services, changes func(bool) preferences.Changes) {
	settingCard(c, title, icon, func() {
		insetToggle(c, label, description, value, false, func() { saveSetting(w, changes(*value)) })
	})
}
func insetToggle(c *ui.Context, label, description string, value *bool, disabled bool, changed func()) {
	k := core.Tokens(c)
	ui.Row(c).FillWidth().Padding(16).Gap(24).AlignItems(ui.Center).Radius(style.RadiusMenu).Background(k.Background).Border(style.BorderWidth, k.Border.Alpha(.6)).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(6).Children(func() {
			ui.Text(c, label).FillWidth().FontWeight(style.WeightSemibold)
			components.Muted(c, description)
		})
		components.SettingsSwitch(c, value, label+"开关", input.SwitchOptions{Disabled: disabled, Description: description}).OnClick(changed)
	})
}
