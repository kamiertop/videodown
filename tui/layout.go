package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const appTitle = " 📥 VideoDown TUI"

// 渲染顶部信息
// 目前只显示appTitle，后面增加{哔哩哔哩/抖音}
func (r *Root) renderHeader() string {
	title := headerStyle.Width(r.width).Align(lipgloss.Center).Render(appTitle)
	line := borderStyle.Render(strings.Repeat("─", r.width))

	return lipgloss.JoinVertical(lipgloss.Center, title, line)
}

func (r *Root) renderFooter() string {
	var shortcuts string

	if r.state == viewHome {
		shortcuts = "[q] | [ctrl+c] 退出程序"
	}

	return footerStyle.Width(r.width).Render(shortcuts)
}

func (r *Root) renderBody(height int) string {

	// 利用 Place 让中间内容在计算出的 bodyHeight 区域内上下居中
	return lipgloss.Place(
		r.width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		"body",
	)
}
