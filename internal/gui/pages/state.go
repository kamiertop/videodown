// Package pages assembles independent page states. Concrete views live in
// subpackages; platform-specific features belong to bilibili and douyin.
package pages

import (
	"github.com/kamiertop/videodown/internal/gui/pages/bilibili"
	"github.com/kamiertop/videodown/internal/gui/pages/douyin"
	"github.com/kamiertop/videodown/internal/gui/pages/settings"
	"github.com/kamiertop/videodown/internal/gui/pages/youtube"
)

type State struct {
	YouTube  *youtube.State
	Settings *settings.State
	Bilibili *bilibili.State
	Douyin   *douyin.State
}

func NewState() *State {
	return &State{YouTube: youtube.NewState(), Settings: settings.NewState(), Bilibili: bilibili.NewState(), Douyin: douyin.NewState()}
}
