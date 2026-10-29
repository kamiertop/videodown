package bilibili

import (
	"fmt"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func TestDynamicFeedLayoutAndActions(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, api, runtime := setup(t, true)
		published := time.Date(2026, time.October, 26, 13, 42, 0, 0, time.Local).Unix()
		api.dynamic = func(offset string) (model.DynamicArchivePage, error) {
			if offset == "cursor" {
				return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{{Bvid: "BVdynamic2", Title: "第二条动态", AuthorName: "很长很长的第二位 UP 主名称", Mid: 43, PubTs: published}}}, nil
			}
			return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{{Bvid: "BVdynamic1", Title: "第一条动态", AuthorName: "动态 UP 主", Mid: 42, DurationText: "03:00", PubTime: "1小时前", PubTs: published}}, Offset: "cursor", HasMore: true}, nil
		}
		path := "/bilibili/dynamic"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			style.Apply(c)
			s.Layout(c, path, func(next string) { path = next }, func() {
				s.Form(c, path, func(next string) { path = next }, func() { path = "/bilibili/dynamic" })
			})
		}, width, 740)
		view.Frame()
		runtime.flush(t)
		view.Frame()
		selectRow := func(label string) {
			t.Helper()
			row, found := view.Find(label)
			if !found {
				t.Fatalf("missing dynamic row %s", label)
			}
			view.ClickAt(row.X+style.SpaceMedium+20, row.Y+row.H/2)
			view.Frame()
		}
		header, ok := view.Find("动态工具栏")
		if !ok || header.H > style.LibraryHeaderHeight+2 {
			t.Fatalf("dynamic header should be one compact row at %d: %+v", width, header)
		}
		for _, label := range []string{"已加载 1 个 · 已选 0 个", "刷新", "下载已选 (0)"} {
			item, found := view.Find(label)
			if !found || item.Y < header.Y || item.Y+item.H > header.Y+header.H+1 {
				t.Fatalf("%s escaped the single header row at %d: %+v / %+v", label, width, item, header)
			}
		}
		if view.HasText("下载全部") || !view.HasText("下载已选 (0)") {
			t.Fatal("dynamic toolbar still offers infinite-feed download or is missing selection actions")
		}
		if view.HasText("加载更多") || view.HasText("继续加载") || view.HasText("清空选择") || len(runtime.timers) != 1 {
			t.Fatal("short first page should schedule automatic pagination without a load-more button")
		}
		if !view.HasText("第一条动态") || !view.HasText("动态 UP 主") {
			t.Fatal("first dynamic page did not render")
		}
		if !view.HasText("2026-10-26 13:42") || view.HasText("1小时前") {
			t.Fatal("dynamic publish time should show a full date and time")
		}
		shortLink, ok := view.Find("进入 UP 主主页：动态 UP 主")
		if !ok || shortLink.W >= style.DynamicAuthorWidth-4 {
			t.Fatalf("short author underline should follow the visible name at %d: %+v", width, shortLink)
		}
		timeCell, ok := view.Find("2026-10-26 13:42")
		if !ok || timeCell.X < shortLink.X+style.DynamicAuthorWidth+style.DynamicColumnGap-1 {
			t.Fatalf("author and full timestamp overlap at %d: author=%+v time=%+v", width, shortLink, timeCell)
		}
		selectRow("选择动态视频：第一条动态")
		if !s.dynamicSelected["BVdynamic1"] || !view.HasText("下载已选 (1)") {
			t.Fatalf("clicking a video row should select it: selected=%v texts=%v", s.dynamicSelected, view.Texts())
		}
		for _, label := range []string{"清空选择", "下载已选 (1)"} {
			item, found := view.Find(label)
			if !found || item.Y < header.Y || item.Y+item.H > header.Y+header.H+1 {
				t.Fatalf("%s is not beside refresh at %d: %+v header=%+v", label, width, item, header)
			}
		}
		runtime.tick(t)
		runtime.flush(t)
		view.Frame()
		if !view.HasText("第二条动态") || !view.HasText("已加载 2 个 · 已选 1 个") {
			t.Fatal("dynamic pagination did not append videos")
		}
		longLink, ok := view.Find("进入 UP 主主页：很长很长的第二位 UP 主名称")
		if !ok || longLink.W > style.DynamicAuthorWidth+1 {
			t.Fatalf("long author name escapes its column at %d: %+v", width, longLink)
		}
		selectRow("选择动态视频：第二条动态")
		if !s.dynamicSelected["BVdynamic2"] || !view.HasText("下载已选 (2)") {
			t.Fatal("multiple dynamic rows should remain selected")
		}
		view.Move(longLink.X+longLink.W/2, longLink.Y+longLink.H/2)
		view.Frame()
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-dynamic-long-author-%d.png", width))
		if err := view.Click("刷新"); err != nil {
			t.Fatal(err)
		}
		runtime.flush(t)
		view.Frame()
		if view.HasText("第二条动态") || !view.HasText("已加载 1 个 · 已选 0 个") {
			t.Fatal("refresh did not replace the old feed")
		}
		if s.dynamicSelectedCount() != 0 || !view.HasText("下载已选 (0)") {
			t.Fatal("refresh should clear selections from the old feed")
		}
		if err := view.Click("进入 UP 主主页：动态 UP 主"); err != nil {
			t.Fatal(err)
		}
		if path != "/bilibili/up/42" {
			t.Fatalf("author link navigated to %s", path)
		}
		path = "/bilibili/dynamic"
		view.Frame()
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-dynamic-feed-%d.png", width))
		if err := view.Click("下载视频：第一条动态"); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if path != "/bilibili/download" || !s.stagedKeys["BVdynamic1"] || !view.HasText("第一条动态") || s.dynamicSelectedCount() != 0 {
			t.Fatalf("dynamic video did not reach the download page: path=%s staged=%v texts=%v", path, s.stagedKeys, view.Texts())
		}
	}
}

func TestDynamicRefreshIgnoresOldPagination(t *testing.T) {
	s, api, runtime := setup(t, true)
	api.dynamic = func(offset string) (model.DynamicArchivePage, error) {
		if offset == "cursor" {
			return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{{Bvid: "BVold", Title: "过期分页"}}}, nil
		}
		return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{{Bvid: "BVnew", Title: "当前动态"}}, Offset: "cursor", HasMore: true}, nil
	}
	s.loadDynamics(false)
	runtime.flush(t)
	s.loadDynamics(true)
	s.loadDynamics(false)
	// Complete the refresh first; the older page must not reappear afterward.
	runtime.finish(t, 1)
	runtime.finish(t, 0)
	if len(s.dynamics.Data.Items) != 1 || s.dynamics.Data.Items[0].Bvid != "BVnew" {
		t.Fatalf("stale pagination replaced refreshed data: %+v", s.dynamics.Data.Items)
	}
	path := "/bilibili/dynamic"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		s.Layout(c, path, func(next string) { path = next }, func() {
			s.Form(c, path, func(next string) { path = next }, func() {})
		})
	}, 760, 600)
	view.Frame()
	row, ok := view.Find("选择动态视频：当前动态")
	if !ok {
		t.Fatal("refreshed row missing")
	}
	view.ClickAt(row.X+style.SpaceMedium+20, row.Y+row.H/2)
	view.Frame()
	if !view.HasText("下载已选 (1)") {
		t.Fatal("refreshed video was not selected")
	}
	if err := view.Click("下载已选 (1)"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if path != "/bilibili/download" || !s.stagedKeys["BVnew"] || s.stagedKeys["BVold"] {
		t.Fatalf("selected download staged wrong feed: path=%s keys=%v", path, s.stagedKeys)
	}
}

func TestDynamicMultiSelectStagesOnlyChosenVideos(t *testing.T) {
	s, api, runtime := setup(t, true)
	api.dynamic = func(string) (model.DynamicArchivePage, error) {
		return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{
			{Bvid: "BVone", Title: "视频一", AuthorName: "作者一", Mid: 42},
			{Bvid: "BVtwo", Title: "视频二", AuthorName: "作者二", Mid: 43},
			{Bvid: "BVthree", Title: "视频三", AuthorName: "作者三", Mid: 44},
		}}, nil
	}
	path := "/bilibili/dynamic"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		s.Layout(c, path, func(next string) { path = next }, func() {
			s.Form(c, path, func(next string) { path = next }, func() {})
		})
	}, 760, 600)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	selectRow := func(label string) {
		t.Helper()
		row, found := view.Find(label)
		if !found {
			t.Fatalf("missing dynamic row %s", label)
		}
		view.ClickAt(row.X+style.SpaceMedium+20, row.Y+row.H/2)
		view.Frame()
	}
	for _, label := range []string{"选择动态视频：视频一", "选择动态视频：视频三"} {
		selectRow(label)
	}
	if !view.HasText("下载已选 (2)") {
		t.Fatal("two dynamic videos were not selected")
	}
	selectRow("取消选择动态视频：视频一")
	if s.dynamicSelectedCount() != 1 || !view.HasText("下载已选 (1)") {
		t.Fatal("clicking a selected row should deselect only that video")
	}
	if err := view.Click("清空选择"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if s.dynamicSelectedCount() != 0 || !view.HasText("下载已选 (0)") {
		t.Fatal("clear selection did not reset the footer")
	}
	for _, label := range []string{"选择动态视频：视频一", "选择动态视频：视频三"} {
		selectRow(label)
	}
	if err := view.Click("进入 UP 主主页：作者二"); err != nil {
		t.Fatal(err)
	}
	if s.dynamicSelectedCount() != 2 || path != "/bilibili/up/43" {
		t.Fatal("author link changed dynamic selection")
	}
	path = "/bilibili/dynamic"
	view.Frame()
	if err := view.Click("下载已选 (2)"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if path != "/bilibili/download" || !s.stagedKeys["BVone"] || !s.stagedKeys["BVthree"] || s.stagedKeys["BVtwo"] || s.dynamicSelectedCount() != 0 {
		t.Fatalf("selected download staged the wrong videos: path=%s keys=%v selected=%d", path, s.stagedKeys, s.dynamicSelectedCount())
	}
}

func TestDynamicAutomaticallyLoadsNearScrollEnd(t *testing.T) {
	s, api, runtime := setup(t, true)
	api.dynamic = func(offset string) (model.DynamicArchivePage, error) {
		if offset == "next" {
			return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{{Bvid: "BVfinal", Title: "自动加载的视频"}}}, nil
		}
		items := make([]model.DynamicArchiveItem, 20)
		for i := range items {
			items[i] = model.DynamicArchiveItem{Bvid: fmt.Sprintf("BVauto%02d", i), Title: fmt.Sprintf("动态视频 %02d", i)}
		}
		return model.DynamicArchivePage{Items: items, Offset: "next", HasMore: true}, nil
	}
	path := "/bilibili/dynamic"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		s.Layout(c, path, func(next string) { path = next }, func() {
			s.Form(c, path, func(next string) { path = next }, func() {})
		})
	}, 760, 600)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	scroll := s.scroll("dynamic-videos")
	if scroll.MaxY <= 0 || len(runtime.timers) != 1 || view.HasText("加载更多") {
		t.Fatalf("scrollable feed should hide the button and verify its viewport: scroll=%+v timers=%d", scroll, len(runtime.timers))
	}
	runtime.tick(t) // The first-page layout check sees that scrolling is possible.
	if len(runtime.jobs) != 0 {
		t.Fatal("scrollable feed loaded a second page before scrolling")
	}
	list, ok := view.Find("动态视频列表")
	if !ok {
		t.Fatal("missing dynamic list")
	}
	view.Scroll(list.X+list.W/2, list.Y+list.H/2, 0, scroll.MaxY)
	view.Frame()
	if scroll.Y <= 0 || len(runtime.timers) != 1 {
		t.Fatalf("near-end scrolling did not schedule one next page: Y=%v timers=%d", scroll.Y, len(runtime.timers))
	}
	view.Frame()
	if len(runtime.timers) != 1 {
		t.Fatal("rebuild scheduled duplicate automatic loads")
	}
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if !view.HasText("自动加载的视频") || !view.HasText("已加载 21 个 · 已选 0 个") || s.dynamicHasMore {
		t.Fatal("automatic pagination failed to append its next page")
	}
}

func TestDynamicAutomaticallySkipsEmptyVideoPage(t *testing.T) {
	s, api, runtime := setup(t, true)
	api.dynamic = func(offset string) (model.DynamicArchivePage, error) {
		if offset == "next" {
			return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{{Bvid: "BVfound", Title: "翻页找到的视频"}}}, nil
		}
		return model.DynamicArchivePage{Offset: "next", HasMore: true}, nil
	}
	path := "/bilibili/dynamic"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		s.Layout(c, path, func(next string) { path = next }, func() {
			s.Form(c, path, func(next string) { path = next }, func() {})
		})
	}, 760, 600)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	if !view.HasText("正在查找关注视频动态") || len(runtime.timers) != 1 {
		t.Fatal("empty video page should schedule the next page")
	}
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if !view.HasText("翻页找到的视频") || view.HasText("继续加载") {
		t.Fatal("automatic pagination did not show the first available video")
	}
}

func TestDynamicStopsAutomaticPagingAfterRepeatedEmptyPages(t *testing.T) {
	s, api, runtime := setup(t, true)
	calls := 0
	api.dynamic = func(string) (model.DynamicArchivePage, error) {
		calls++
		return model.DynamicArchivePage{Offset: fmt.Sprintf("cursor-%d", calls), HasMore: true}, nil
	}
	path := "/bilibili/dynamic"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		s.Layout(c, path, func(next string) { path = next }, func() {
			s.Form(c, path, func(next string) { path = next }, func() {})
		})
	}, 760, 600)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	for range style.DynamicAutoFillLimit {
		if len(runtime.timers) != 1 {
			t.Fatal("expected one scheduled page while the feed cannot scroll")
		}
		runtime.tick(t)
		runtime.flush(t)
		view.Frame()
	}
	if len(runtime.timers) != 0 || !view.HasText("继续加载") || calls != 1+style.DynamicAutoFillLimit {
		t.Fatalf("repeated empty pages should pause automatic requests: calls=%d timers=%d", calls, len(runtime.timers))
	}
	if err := view.Click("继续加载"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	if calls != 2+style.DynamicAutoFillLimit {
		t.Fatal("manual retry did not load the next empty page")
	}
}
