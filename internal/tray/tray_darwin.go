//go:build darwin

package tray

import (
	"context"
	"sync/atomic"

	"github.com/kamiertop/videodown/logger"
)

// Tray 管理系统托盘图标和菜单。
// 在 macOS 上，Wails 已通过 NSApp 管理 Dock 和菜单栏，不需要 systray；
// 窗口的显示/隐藏由 Dock 图标点击自然支持。
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

// SetForceQuitCallback 设置强制退出回调（macOS 上由 Cmd+Q 或菜单栏 Quit 自然触发，
// 但保留回调以兼容 settings 的 ForceQuit 路径）。
func (t *Tray) SetForceQuitCallback(fn func()) {
	t.forceQuit = fn
}

// Start 在 macOS 上为空操作 —— Wails 负责 NSApp 和 Dock。
func (t *Tray) Start() {
	t.logger.Info("Tray Start on macOS: skipping systray (dock + menu bar handle this natively)")
}

// SetContext 注入 Wails 上下文。
func (t *Tray) SetContext(ctx context.Context) {
	t.ctx.Store(&ctx)
}

// Stop 退出托盘（macOS 上为空操作）。
func (t *Tray) Stop() {
	t.logger.Info("Tray Stop on macOS: no-op")
}
