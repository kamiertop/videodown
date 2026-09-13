package tui

import (
	_ "charm.land/bubbles/v2"
	"charm.land/lipgloss/v2"
)

const appTitle = " 📥 VideoDown TUI"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5A56E0")). // 紫色背景
			Border(lipgloss.RoundedBorder(), true, true, false, true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#222222")).
			Border(lipgloss.RoundedBorder(), false, true, true, true)
)

// 渲染顶部信息
// 目前只显示appTitle，后面增加{哔哩哔哩/抖音}
func (r Root) renderHeader() string {
	return headerStyle.
		Width(r.width).
		Align(lipgloss.Center).
		Render(appTitle)
}

func (r Root) renderFooter() string {
	var shortcuts string

	if r.page == pageHome {
		shortcuts = "[q] | [ctrl+c] 退出程序"
	}

	return footerStyle.Width(r.width).Render(shortcuts)
}
