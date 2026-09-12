package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/kamiertop/videodown/tui"
)

func main() {
	program := tea.NewProgram(tui.NewRootModel())

	if _, err := program.Run(); err != nil {
		panic(err)
	}
}
