package home

import "charm.land/lipgloss/v2"

type Model struct {
	width  int
	height int
}

var homeStyle = lipgloss.NewStyle().Background(lipgloss.BrightCyan).Border(lipgloss.RoundedBorder(), false, true, false, true)

func (m Model) View() string {
	return homeStyle.Width(m.width).Height(m.height).Render("Home")
}

func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height

	return m
}
