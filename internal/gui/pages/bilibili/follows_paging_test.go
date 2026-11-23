package bilibili

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// TestMergeFollowPagesAppendAndDedup covers the follow grid's append paging:
// new items join the end and repeated mids (the list shifts as follows
// change) never produce a second card.
func TestMergeFollowPagesAppendAndDedup(t *testing.T) {
	old := []model.FollowDataList{{Mid: 1}, {Mid: 2}}
	page := []model.FollowDataList{{Mid: 2}, {Mid: 3}, {Mid: 3}, {Mid: 4}}
	merged := mergeFollowPages(old, page)
	if len(merged) != 4 {
		t.Fatalf("merged %d items, want 4: %+v", len(merged), merged)
	}
	for i, want := range []int64{1, 2, 3, 4} {
		if merged[i].Mid != want {
			t.Fatalf("merged[%d] = %d, want %d", i, merged[i].Mid, want)
		}
	}
}

func followsTester(s *State) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.activePath = "/bilibili/up"
		s.Render(c, "up", func(string) {})
	}, 1200, 880)
}

// TestFollowsGridAutoLoadFooter verifies the pager buttons are gone and the
// grid keeps its items while more pages remain, with the headless fallback
// continuing the load.
func TestFollowsGridAutoLoadFooter(t *testing.T) {
	s := NewState()
	s.auth.loggedIn = true
	s.follows.list.Loaded = true
	s.follows.list.Data = model.FollowData{Total: 50}
	for i := range 20 {
		s.follows.list.Data.List = append(s.follows.list.Data.List, model.FollowDataList{Mid: int64(i + 1), Uname: fmt.Sprintf("UP%02d", i)})
	}
	tester := followsTester(s)
	tester.Frame()
	if _, ok := tester.Find("下一页"); ok {
		t.Fatal("manual pager survived the migration")
	}
	if _, ok := tester.Find("UP00"); !ok {
		t.Fatal("follow card missing")
	}
	if !tester.HasText("已加载 20 / 50") {
		t.Fatal("loaded-count footer missing")
	}
	// The tester has no timer, so the fallback button continues the pages.
	if err := tester.Click("继续加载"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if s.autoPage("author-follows").fillPages != 0 {
		t.Fatal("continue button did not reset the fill counter")
	}

	// All items present: the footer settles and the fallback disappears.
	s.follows.list.Data.List = s.follows.list.Data.List[:0]
	for i := range 50 {
		s.follows.list.Data.List = append(s.follows.list.Data.List, model.FollowDataList{Mid: int64(i + 1), Uname: fmt.Sprintf("UP%02d", i)})
	}
	tester.Frame()
	if !tester.HasText("已加载全部 50 位 UP 主") {
		t.Fatal("all-loaded footer missing")
	}
	if _, ok := tester.Find("继续加载"); ok {
		t.Fatal("fallback button shown after the last page")
	}
}

// TestUpToolbarStaysFixedWhileListScrolls pins the UP 主 page's layout: the
// toolbar lives outside the follow list's own scroll area, so scrolling the
// cards never moves it.
func TestUpToolbarStaysFixedWhileListScrolls(t *testing.T) {
	s := NewState()
	s.auth.loggedIn = true
	s.follows.list.Loaded = true
	s.follows.list.Data = model.FollowData{Total: 60}
	for i := range 60 {
		s.follows.list.Data.List = append(s.follows.list.Data.List, model.FollowDataList{Mid: int64(i + 1), Uname: fmt.Sprintf("UP%02d", i)})
	}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.activePath = "/bilibili/up"
		s.Render(c, "up", func(string) {})
	}, 1200, 880)
	test.Frame()
	toolbar, ok := test.Find("刷新")
	if !ok {
		t.Fatal("missing toolbar anchor")
	}
	scroll := s.scroll(followsScrollKey)
	if scroll.MaxY <= 0 {
		t.Fatalf("follow list does not scroll (MaxY=%.1f)", scroll.MaxY)
	}
	scroll.Y = 150
	test.Frame()
	after, ok := test.Find("刷新")
	if !ok {
		t.Fatal("toolbar disappeared while scrolling")
	}
	if after != toolbar {
		t.Fatalf("toolbar moved while the list scrolled: %+v -> %+v", toolbar, after)
	}
	if _, ok := test.Find("UP45"); !ok {
		t.Fatal("scrolled-in card missing")
	}
}
