// Package closing implements the configuration and decisions behind window
// close and explicit app quit, independently of MyGo window operations.
package closing

import "errors"

type Preferences interface {
	CloseToTray() (enabled, configured bool, err error)
	SetCloseToTray(bool) error
}
type Action int

const (
	Exit Action = iota
	Hide
	Choose
	ConfirmTasks
	Wait
)

type Controller struct {
	preferences Preferences
	CloseToTray bool
	Configured  bool
	Saving      bool
	Error       string
	Prompt      Action
	Open        bool
	Remember    bool
	Active      int
}

func New(preferences Preferences) *Controller {
	c := &Controller{preferences: preferences, CloseToTray: true}
	if preferences != nil {
		enabled, configured, err := preferences.CloseToTray()
		if err != nil {
			c.Error = "读取关闭配置失败：" + err.Error()
		} else {
			c.CloseToTray, c.Configured = enabled, configured
		}
	}
	return c
}
func (c *Controller) OnClose(tray bool, active int) Action {
	if c.Saving {
		return Wait
	}
	if !c.Configured {
		return Choose
	}
	if c.CloseToTray && tray {
		return Hide
	}
	if active > 0 {
		return ConfirmTasks
	}
	return Exit
}
func (c *Controller) OnQuit(active int, confirmed bool) Action {
	if c.Saving {
		return Wait
	}
	if active > 0 && !confirmed {
		return ConfirmTasks
	}
	return Exit
}
func (c *Controller) Show(action Action, active int) {
	if !c.Open || c.Prompt != action {
		c.Remember = false
	}
	c.Prompt, c.Active, c.Open = action, active, true
}

// WritePreference may run in a worker; mutable UI state is committed separately.
func (c *Controller) WritePreference(value bool) error {
	if c.preferences == nil {
		return errors.New("配置存储尚未初始化")
	}
	return c.preferences.SetCloseToTray(value)
}
func (c *Controller) Saved(value bool, err error) {
	c.Saving = false
	if err != nil {
		c.Error = "保存关闭配置失败：" + err.Error()
		return
	}
	c.CloseToTray, c.Configured, c.Error = value, true, ""
}
