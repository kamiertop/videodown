package bilibili

import (
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// feed[T] 是带“加载更多”的板块（关注动态、播放历史）共有的翻页状态：
// 首屏与增量页两个资源、刷新代数（拦下迟到的分页响应）、自动加载的
// 滚动基准与空列表自动补页计数。板块特有的游标与勾选集放在外层。
type feed[T any] struct {
	items, more resource[T]
	hasMore     bool
	generation  uint64
	autoPending bool
	autoAtY     float32
	autoFill    int
}

// reset 在刷新前重置增量页与自动加载状态，并把代数加一，让迟到的旧
// 响应作废。
func (f *feed[T]) reset() {
	f.generation++
	f.more.seq++
	f.more.Loading, f.more.Error = false, ""
	var zero T
	f.items.Data, f.items.Loaded, f.items.Loading, f.items.Error = zero, false, false, ""
	f.autoPending, f.autoAtY, f.autoFill = false, 0, 0
}

// ready 判断是否贴近列表底部；空列表只自动补有限页，防止无限轮询。
func (f *feed[T]) ready(scroll *ui.ScrollState) bool {
	if scroll.MaxY <= 0 {
		return f.autoFill < style.DynamicAutoFillLimit
	}
	return scroll.Y > f.autoAtY+1 && scroll.Y+style.DynamicAutoLoadThreshold >= scroll.MaxY
}

// maybeAutoLoadFeed 滚动接近底部时补一页：canLoad 是板块特有的一致性
// 守卫（如动态的 offset 非空），unchanged 在定时器回调里确认游标未变；
// 两者都会在回调时重新求值。
func maybeAutoLoadFeed[T any](s *State, f *feed[T], scrollKey, section string, canLoad func() bool, unchanged func() bool, loadMore func()) {
	if canLoad == nil {
		canLoad = func() bool { return true }
	}
	scroll := s.scroll(scrollKey)
	if s.runtime.After == nil || f.autoPending || f.more.Loading || f.more.Error != "" || !f.hasMore || !canLoad() || !f.ready(scroll) {
		return
	}
	f.autoPending = true
	epoch, generation := s.epoch, f.generation
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || f.generation != generation {
			return
		}
		f.autoPending = false
		state := s.scroll(scrollKey)
		if s.section != section || !unchanged() || f.more.Error != "" || !canLoad() || !f.ready(state) {
			return
		}
		if state.MaxY <= 0 {
			f.autoFill++
		} else {
			f.autoFill = 0
		}
		loadMore()
	})
}

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

const autoPagePreloadDistance float32 = 600

func (p *autoPageState) nearEnd(scroll *ui.ScrollState) bool {
	if scroll.MaxY <= 0 {
		return p.fillPages < style.DynamicAutoFillLimit
	}
	return scroll.Y > p.atY+1 && scroll.Y+autoPagePreloadDistance >= scroll.MaxY
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
