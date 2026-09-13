package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kamiertop/videodown/tui/bilibili"
	"github.com/kamiertop/videodown/tui/douyin"
	"github.com/kamiertop/videodown/tui/home"
)

// Root videodown tui model 存储应用的完整状态
type Root struct {
	width    int
	height   int
	page     page
	bilibili bilibili.Model
	douyin   douyin.Model
	home     home.Model
}

type page int

const (
	pageHome page = iota
	pageBilibili
	pageDouyin
)

func NewRootModel() Root {
	return Root{
		page: pageHome,
	}
}

// View 根据当前Model渲染纯文本界面
func (r Root) View() tea.View {
	if r.width == 0 || r.height == 0 {
		return tea.NewView("初始化页面中")
	}
	header := r.renderHeader()

	footer := r.renderFooter()

	var content string
	switch r.page {
	case pageHome:
		content = r.home.View()
	case pageBilibili:
		content = r.bilibili.View()
	case pageDouyin:
		content = r.douyin.View()
	}

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, header, content, footer))
}

func (r Root) Init() tea.Cmd {
	return nil
}

// Update 接收事件，根据事件更新model
func (r Root) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return r, tea.Quit
		}
	case tea.WindowSizeMsg:
		r.height = msg.Height
		r.width = msg.Width
		r.home = r.home.SetSize(
			r.width, r.height-lipgloss.Height(r.renderHeader())-lipgloss.Height(r.renderFooter()),
		)
	}

	return r, nil
}
