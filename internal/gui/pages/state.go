// Package pages assembles independent page states. Concrete views live in
// subpackages; platform-specific features belong to bilibili and douyin.
package pages

import (
	"github.com/kamiertop/videodown/internal/gui/pages/bilibili"
	"github.com/kamiertop/videodown/internal/gui/pages/douyin"
	"github.com/kamiertop/videodown/internal/gui/pages/settings"
)

type State struct {
	Settings *settings.State
	Bilibili *bilibili.State
	Douyin   *douyin.State
}

func NewState() *State {
	return &State{Settings: settings.NewState(), Bilibili: bilibili.NewState(), Douyin: douyin.NewState()}
}
