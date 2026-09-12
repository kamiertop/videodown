package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Root videodown tui model 存储应用的完整状态
type Root struct {
	width  int
	height int
	state  viewState
}

type viewState int

const (
	viewHome viewState = iota
	viewDouyin
	viewBilibili
)

func NewRootModel() *Root {
	return &Root{
		width:  30,
		height: 60,
		state:  viewHome,
	}
}

// View 根据当前Model渲染纯文本界面
func (r *Root) View() tea.View {
	if r.width == 0 || r.height == 0 {
		return tea.NewView("初始化页面中")
	}
	header := r.renderHeader()
	headerHeight := lipgloss.Height(header)

	footer := r.renderFooter()
	footerHeight := lipgloss.Height(footer)

	bodyHeight := r.height - headerHeight - footerHeight

	body := r.renderBody(bodyHeight)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
}

func (r *Root) Init() tea.Cmd {
	return nil
}

// Update 接收时间，根据事件更新model，并可选择返回副作用命令
func (r *Root) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return r, tea.Quit
		}
	case tea.WindowSizeMsg:
		r.height = msg.Height
		r.width = msg.Width

	}

	return r, nil
}
