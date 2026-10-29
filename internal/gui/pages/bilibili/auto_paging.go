package bilibili

import (
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// One scroll region can display several sources (for example an UP 主's
// uploads and series). Reset invalidates its pending callback on a switch.
type autoPageState struct {
	pending   bool
	atY       float32
	fillPages int
	version   uint64
}

func (s *State) autoPage(key string) *autoPageState {
	if s.autoPages == nil {
		s.autoPages = make(map[string]*autoPageState)
	}
	if s.autoPages[key] == nil {
		s.autoPages[key] = new(autoPageState)
	}
	return s.autoPages[key]
}

func (s *State) resetAutoPage(key, scrollKey string) {
	page := s.autoPage(key)
	page.version++
	page.pending = false
	page.atY = 0
	page.fillPages = 0
	scroll := s.scroll(scrollKey)
	scroll.Y = 0
	scroll.MaxY = 0
}

func (p *autoPageState) nearEnd(scroll *ui.ScrollState) bool {
	if scroll.MaxY <= 0 {
		return p.fillPages < style.DynamicAutoFillLimit
	}
	return scroll.Y > p.atY+1 && scroll.Y+style.DynamicAutoLoadThreshold >= scroll.MaxY
}

// ready is evaluated again when the timer fires, after the new layout has
// committed its scroll range and after any source or route change.
func (s *State) maybeAutoPage(key, scrollKey string, ready func() bool, loadNext func()) {
	page := s.autoPage(key)
	scroll := s.scroll(scrollKey)
	if s.runtime.After == nil || page.pending || !ready() || !page.nearEnd(scroll) {
		return
	}
	page.pending = true
	epoch, version := s.epoch, page.version
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || page.version != version {
			return
		}
		page.pending = false
		if !ready() || !page.nearEnd(scroll) {
			return
		}
		if scroll.MaxY <= 0 {
			page.fillPages++
		} else {
			page.fillPages = 0
		}
		page.atY = scroll.Y
		loadNext()
	})
}

func (s *State) autoPageFallback(key, scrollKey string) bool {
	return s.runtime.After == nil || s.scroll(scrollKey).MaxY <= 0 && s.autoPage(key).fillPages >= style.DynamicAutoFillLimit
}

func (s *State) continueAutoPage(key string, loadNext func()) {
	s.autoPage(key).fillPages = 0
	loadNext()
}
