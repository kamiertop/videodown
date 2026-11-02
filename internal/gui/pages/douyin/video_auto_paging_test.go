package douyin

import (
	"fmt"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/douyin/model"
)

func TestVideoAutoPagingRechecksCommittedLayout(t *testing.T) {
	s, _, _, r := setup(t)
	s.enter("/douyin/user/MS4wLjAB_author")
	r.flush(t)
	scroll := s.scroll("videos")
	// The first build still sees the empty region's scroll range.
	scroll.MaxY = 0
	s.maybeMore()
	s.maybeMore()
	if len(r.timers) != 1 {
		t.Fatal("expected one deferred layout check")
	}
	// Layout commits: the first page already fills more than one screen.
	scroll.MaxY = 1000
	r.timers[0]()
	if len(r.jobs) != 0 || s.autoPending || s.autoFill != 0 {
		t.Fatal("first page triggered an unnecessary second request")
	}
	s.maybeMore()
	if len(r.timers) != 1 {
		t.Fatal("loaded without scrolling")
	}
	scroll.Y = 450 // Preload before reaching the old 120-unit threshold.
	s.maybeMore()
	if len(r.timers) != 2 {
		t.Fatal("near-bottom scroll did not schedule next page")
	}
	r.timers[1]()
	r.flush(t)
	if len(s.videos.Data.Items) != 2 {
		t.Fatal("next page was not appended")
	}
}

func TestVideoAutoPagingStillFillsShortPage(t *testing.T) {
	s, _, _, r := setup(t)
	s.enter("/douyin/user/MS4wLjAB_author")
	r.flush(t)
	s.scroll("videos").MaxY = 0
	s.maybeMore()
	if len(r.timers) != 1 {
		t.Fatal("short page should schedule a refill")
	}
	r.timers[0]()
	r.flush(t)
	if len(s.videos.Data.Items) != 2 || s.autoFill != 1 {
		t.Fatal("short page did not load the next page")
	}
}

func TestVideoAutoPagingRechecksScrollBeforeRequest(t *testing.T) {
	s, _, _, r := setup(t)
	s.enter("/douyin/user/MS4wLjAB_author")
	r.flush(t)
	scroll := s.scroll("videos")
	scroll.MaxY, scroll.Y = 1000, 900
	s.maybeMore()
	scroll.Y = 100
	r.timers[0]()
	if len(r.jobs) != 0 || s.autoPending {
		t.Fatal("loaded after user scrolled away from bottom")
	}
}

func TestTwentyAuthorVideosWaitForScrolling(t *testing.T) {
	s, _, _, r := setup(t)
	path := "/douyin/user/MS4wLjAB_author"
	navigate := func(next string) { path = next }
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() {
			s.Layout(c, path, navigate, func() { s.Form(c, path, navigate, func() {}) })
		})
	}, 1200, 800)
	view.Frame()
	r.flush(t)
	calls := 0
	s.useSource(source{Key: "user:MS4wLjAB_author", Load: func(cursor int) (videoPage, error) {
		calls++
		items := make([]model.AwemeItem, 20)
		for i := range items {
			items[i] = fixture(fmt.Sprint(cursor + i + 1))
		}
		return videoPage{Items: items, Next: cursor + 20, More: true}, nil
	}})
	r.flush(t)
	view.Frame()
	if s.scroll("videos").MaxY <= 0 {
		t.Fatal("20 cards should exceed the viewport")
	}
	timers := r.timers
	r.timers = nil
	for _, tick := range timers {
		tick()
	}
	r.flush(t)
	view.Frame()
	if calls != 1 || len(s.videos.Data.Items) != 20 {
		t.Fatal("first visible page fetched twice without scrolling")
	}
}

func TestVideoPaginationKeepsExistingCardsInPlace(t *testing.T) {
	for _, path := range []string{"/douyin/user/MS4wLjAB_author", "/douyin/favorite"} {
		t.Run(path, func(t *testing.T) {
			s, _, _, r := setup(t)
			navigate := func(next string) { path = next }
			view := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				ui.Column(c).Fill().Children(func() {
					s.Layout(c, path, navigate, func() { s.Form(c, path, navigate, func() {}) })
				})
			}, 1200, 800)
			view.Frame()
			r.flush(t)
			s.useSource(source{Key: "stable-pagination", Load: func(cursor int) (videoPage, error) {
				items := make([]model.AwemeItem, 20)
				for i := range items {
					items[i] = fixture(fmt.Sprint(cursor + i + 1))
					items[i].Desc = fmt.Sprintf("稳定作品 %02d", cursor+i+1)
				}
				return videoPage{Items: items, Next: cursor + 20, More: cursor == 0}, nil
			}})
			r.flush(t)
			view.Frame()
			view.Frame()
			label := "选择抖音作品：稳定作品 01"
			before, ok := view.Find(label)
			if !ok {
				t.Fatal("missing first card")
			}
			s.loadVideos(true)
			view.Frame()
			during, ok := view.Find(label)
			if !ok || during.X != before.X || during.Y != before.Y || during.W != before.W || during.H != before.H {
				t.Fatalf("loading moved existing card: before=%+v during=%+v", before, during)
			}
			r.flush(t)
			view.Frame()
			view.Frame()
			after, ok := view.Find(label)
			if !ok || after.X != before.X || after.Y != before.Y || after.W != before.W || after.H != before.H {
				t.Fatalf("appending moved existing card: before=%+v after=%+v", before, after)
			}
			if len(s.videos.Data.Items) != 40 {
				t.Fatal("next page did not append")
			}
		})
	}
}
