package bilibili

import tea "charm.land/bubbletea/v2"

type Model struct {
	height int
	width  int
}

func (m Model) View() string {
	return "bilibili"
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	return m, nil
}
