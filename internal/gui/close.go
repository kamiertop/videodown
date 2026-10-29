package gui

import (
	"fmt"
	"github.com/kamiertop/videodown/internal/gui/style"

	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/overlay"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/closing"
	"github.com/kamiertop/videodown/internal/gui/components"
)

func (s *Shell) activeCount() int {
	count := s.pages.Bilibili.PendingDownloadCount() + s.pages.Douyin.PendingDownloadCount()
	if s.activeDownloads != nil {
		count = max(count, s.activeDownloads())
	}
	return count
}

func (s *Shell) beforeClose(event *mygo.CloseEvent) {
	if s.quitting {
		return
	}
	if s.settingsBusy {
		event.PreventDefault()
		return
	}
	action := s.closing.OnClose(s.tray != nil, s.activeCount())
	if action == closing.Exit {
		return
	}
	event.PreventDefault()
	if action == closing.Hide {
		s.win.Hide()
		return
	}
	if action != closing.Wait {
		s.promptClose(action)
	}
}

func (s *Shell) beforeQuit(event *mygo.QuitEvent) {
	if s.settingsBusy {
		event.PreventDefault()
		return
	}
	action := s.closing.OnQuit(s.activeCount(), s.confirmedQuit)
	if action != closing.Exit {
		event.PreventDefault()
		if action != closing.Wait {
			s.promptClose(action)
		}
		return
	}
	s.quitting = true
}

func (s *Shell) promptClose(action closing.Action) {
	s.closing.Show(action, s.activeCount())
	s.showWindow()
	if s.win != nil {
		s.win.Invalidate()
	}
}
func (s *Shell) saveClosePreference(value bool, after func()) {
	if s.closing.Saving {
		return
	}
	s.closing.Saving = true
	s.closing.Error = ""
	s.runJob(func() error { return s.closing.WritePreference(value) }, func(err error) {
		s.closing.Saved(value, err)
		if err != nil {
			s.notifySettings(s.closing.Error, true)
		} else if after == nil {
			s.notifySettings("保存成功", false)
		}
		if err == nil && after != nil {
			after()
		}
	})
}
func (s *Shell) chooseClose(hide bool) {
	finish := func() {
		s.closing.Open = false
		if hide {
			if s.win != nil {
				s.win.Hide()
			}
		} else {
			mygo.App.Quit()
		}
	}
	if s.closing.Remember {
		s.saveClosePreference(hide, finish)
	} else {
		finish()
	}
}
func (s *Shell) closeDialogs(c *ui.Context) {
	state := s.closing
	title, description := "关闭窗口", "你希望现在如何关闭窗口？"
	if state.Prompt == closing.ConfirmTasks {
		title, description = "还有任务未完成", "退出会中断正在进行的下载任务。"
	}
	overlay.Dialog(c, &state.Open,
		overlay.DialogOptions{Title: title, Description: description, KeepOnBackdrop: true, KeepOnEscape: state.Saving, HideClose: state.Saving},
		func() {
			if state.Active > 0 {
				components.Muted(c, fmt.Sprintf("仍有 %d 个任务正在下载，退出会中断这些任务。", state.Active))
			}
			if state.Prompt == closing.Choose {
				input.Checkbox(c.Key("remember-close"), &state.Remember, "记住选择，下次不再询问", input.CheckboxOptions{Disabled: state.Saving})
			}
			if s.trayError != "" {
				components.Muted(c, s.trayError)
			}
			if state.Error != "" {
				components.Muted(c, state.Error)
			}
			ui.Column(c).FillWidth().
				Radius(style.RadiusMenu).
				Gap(8).
				Children(func() {
					label := "最小化到托盘"
					if state.Prompt == closing.ConfirmTasks {
						label = "最小化到托盘，继续下载"
					}
					components.Button(c.Key("hide"), label, input.ButtonOptions{Variant: input.Outline, Disabled: s.tray == nil || state.Saving}).
						FillWidth().
						Radius(style.RadiusMenu).
						OnClick(func() {
							if state.Prompt == closing.Choose {
								s.chooseClose(true)
							} else {
								state.Open = false
								s.win.Hide()
							}
						})
					label = "退出程序"
					if state.Prompt == closing.ConfirmTasks {
						label = "仍要退出"
					}
					components.Button(c.Key("exit"), label, input.ButtonOptions{Variant: input.Ghost, Loading: state.Saving}).
						FillWidth().
						Radius(style.RadiusMenu).
						OnClick(func() {
							if state.Prompt == closing.Choose {
								s.chooseClose(false)
							} else {
								state.Open = false
								s.confirmedQuit = true
								mygo.App.Quit()
							}
						})
					components.Button(c.Key("cancel"), "取消", input.ButtonOptions{Variant: input.Ghost, Disabled: state.Saving}).
						FillWidth().
						Radius(style.RadiusMenu).
						OnClick(func() { state.Open = false; state.Remember = false })
				})
		})
}
