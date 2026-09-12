package tui

import "charm.land/lipgloss/v2"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5A56E0")). // 紫色背景
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#222222")).
			Padding(0, 1)
)
