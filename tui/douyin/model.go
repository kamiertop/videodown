package douyin

import tea "charm.land/bubbletea/v2"

type Model struct {
	height int
	width  int
}

func (m Model) View() string {
	return "douyin"
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	return m, nil
}
