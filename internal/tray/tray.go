package tray

import (
	"context"
	"sync/atomic"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/kamiertop/videodown/logger"
)

// Tray 管理系统托盘图标和菜单。
type Tray struct {
	logger    *logger.Logger
	icon      []byte
	ctx       atomic.Pointer[context.Context]
	forceQuit func()
}

// New 创建托盘控制器。
func New(log *logger.Logger, icon []byte) *Tray {
	t := &Tray{
		logger: log.WithName("Tray"),
		icon:   icon,
	}
	return t
}

// SetForceQuitCallback 设置退出前回调。
func (t *Tray) SetForceQuitCallback(fn func()) {
	t.forceQuit = fn
}

// Start 注册 systray 回调，必须在 Wails 启动前从主 goroutine 调用。
func (t *Tray) Start() {
	systray.Register(t.onReady, t.onExit)
}

// SetContext 由 Startup 调用，注入 Wails 上下文。
func (t *Tray) SetContext(ctx context.Context) {
	t.ctx.Store(&ctx)
}

func (t *Tray) getCtx() context.Context {
	c := t.ctx.Load()
	if c == nil {
		return nil
	}
	return *c
}

// Stop 退出系统托盘。
func (t *Tray) Stop() {
	systray.Quit()
}

func (t *Tray) onReady() {
	systray.SetIcon(t.icon)
	systray.SetTitle("videodown")
	systray.SetTooltip("videodown - 视频下载工具")

	mShow := systray.AddMenuItem("显示窗口", "显示/恢复主窗口")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "完全退出程序")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				t.showWindow()
			case <-mQuit.ClickedCh:
				t.quitApp()
			}
		}
	}()
}

func (t *Tray) onExit() {
	t.logger.Info("System tray exited")
}

func (t *Tray) showWindow() {
	ctx := t.getCtx()
	if ctx == nil {
		return
	}
	runtime.WindowShow(ctx)
}

func (t *Tray) quitApp() {
	if t.forceQuit != nil {
		t.forceQuit()
	}
	ctx := t.getCtx()
	if ctx == nil {
		return
	}
	runtime.Quit(ctx)
}
