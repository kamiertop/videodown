package bilibili

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func TestPlayHistorySelectionAuthorAndDownload(t *testing.T) {
	s, api, runtime := setup(t, true)
	api.playHistory = func(int, int) (model.PlayHistoryData, error) {
		return model.PlayHistoryData{List: []model.PlayHistoryItem{
			{Title: "看过的视频一", Cover: "", AuthorName: "历史作者", AuthorMid: 42, ViewAt: int(time.Date(2026, 10, 28, 12, 0, 0, 0, time.Local).Unix()), Duration: 120, History: model.PlayHistory{Bvid: "BVhistory1"}},
			{LongTitle: "长标题视频二", AuthorName: "另一个作者", History: model.PlayHistory{Bvid: "BVhistory2"}},
			{Title: "无 BV 记录", History: model.PlayHistory{Oid: 9, Business: "article"}},
		}}, nil
	}
	path := "/bilibili/play-history"
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
	if !view.HasText("已加载 3 个") || !view.HasText("看过的视频一") || !view.HasText("长标题视频二") {
		t.Fatal("play history model did not render")
	}
	saveImage(t, view, "/tmp/bilibili-play-history-760.png")
	if err := view.Click("进入 UP 主主页：历史作者"); err != nil {
		t.Fatal(err)
	}
	if path != "/bilibili/up/42" || s.playHistorySelectedCount() != 0 {
		t.Fatal("author click should navigate without selecting the video")
	}
	path = "/bilibili/play-history"
	view.Frame()
	if err := view.Click("选择视频：看过的视频一"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if s.playHistorySelectedCount() != 1 || !view.HasText("下载已选 (1)") {
		t.Fatal("history card did not select")
	}
	if err := view.Click("全选"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if s.playHistorySelectedCount() != 2 || !view.HasText("取消全选") {
		t.Fatal("select all should include only BV videos")
	}
	if err := view.Click("下载已选 (2)"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	runtime.flush(t)
	view.Frame()
	if path != "/bilibili/download" || !s.stagedKeys["BVhistory1"] || !s.stagedKeys["BVhistory2"] || len(s.stagedKeys) != 2 {
		t.Fatalf("history selection did not reach download page: path=%s keys=%v", path, s.stagedKeys)
	}
	if !view.HasText("看过的视频一") || !view.HasText("长标题视频二") || s.stagedSourceNames["BVhistory1"] != "B 站播放历史" {
		t.Fatal("download page did not retain the original history records")
	}
}

func TestPlayHistoryAutoPaginationAndRefresh(t *testing.T) {
	s, api, runtime := setup(t, true)
	api.playHistory = func(cursor, viewAt int) (model.PlayHistoryData, error) {
		if cursor == 5 {
			return model.PlayHistoryData{List: []model.PlayHistoryItem{{Title: "下一页的视频", History: model.PlayHistory{Bvid: "BVlast"}}}, Cursor: model.PlayHistoryCursor{Max: 5, ViewAt: viewAt}}, nil
		}
		items := make([]model.PlayHistoryItem, 24)
		for i := range items {
			items[i] = model.PlayHistoryItem{Title: fmt.Sprintf("历史视频 %02d", i), History: model.PlayHistory{Bvid: fmt.Sprintf("BVhistory%02d", i)}}
		}
		return model.PlayHistoryData{List: items, Cursor: model.PlayHistoryCursor{Max: 5, ViewAt: 100}}, nil
	}
	path := "/bilibili/play-history"
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
	scroll := s.scroll("play-history-videos")
	if scroll.MaxY <= 0 || len(runtime.timers) != 1 || view.HasText("继续加载") {
		t.Fatalf("scrollable history should wait for scrolling: scroll=%+v timers=%d", scroll, len(runtime.timers))
	}
	runtime.tick(t)
	if len(runtime.jobs) != 0 {
		t.Fatal("second history page loaded before scrolling")
	}
	if err := view.Click("选择视频：历史视频 00"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	list, ok := view.Find("播放历史记录")
	if !ok {
		t.Fatal("missing history panel")
	}
	view.Scroll(list.X+list.W/2, list.Y+list.H/2, 0, scroll.MaxY)
	view.Frame()
	if len(runtime.timers) != 1 {
		t.Fatal("scrolling near the end should schedule one page")
	}
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if !view.HasText("下一页的视频") || !view.HasText("已加载 25 个") || s.playHistorySelectedCount() != 1 {
		t.Fatal("next history page did not append without losing selection")
	}
	if err := view.Click("刷新"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if view.HasText("下一页的视频") || !view.HasText("已加载 24 个") || s.playHistorySelectedCount() != 0 || scroll.Y != 0 {
		t.Fatal("refresh should replace the history and reset selection and scroll")
	}
}

func TestPlayHistoryPaginationErrorCanRetry(t *testing.T) {
	s, api, runtime := setup(t, true)
	fail := true
	api.playHistory = func(cursor, _ int) (model.PlayHistoryData, error) {
		if cursor == 1 {
			if fail {
				fail = false
				return model.PlayHistoryData{}, errors.New("历史分页失败")
			}
			return model.PlayHistoryData{List: []model.PlayHistoryItem{{Title: "重试后的视频", History: model.PlayHistory{Bvid: "BVretry"}}}}, nil
		}
		return model.PlayHistoryData{List: []model.PlayHistoryItem{{Title: "首页视频", History: model.PlayHistory{Bvid: "BVfirst"}}}, Cursor: model.PlayHistoryCursor{Max: 1}}, nil
	}
	path := "/bilibili/play-history"
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
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if !view.HasText("历史分页失败") || !view.HasText("重试") || !view.HasText("首页视频") {
		t.Fatal("pagination error should keep the first page and offer retry")
	}
	if err := view.Click("重试"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if !view.HasText("重试后的视频") || view.HasText("历史分页失败") {
		t.Fatal("retry did not append the recovered page")
	}
}

func TestPlayHistoryRefreshIgnoresOldPage(t *testing.T) {
	s, api, runtime := setup(t, true)
	api.playHistory = func(cursor, _ int) (model.PlayHistoryData, error) {
		if cursor == 1 {
			return model.PlayHistoryData{List: []model.PlayHistoryItem{{Title: "过期历史", History: model.PlayHistory{Bvid: "BVold"}}}}, nil
		}
		return model.PlayHistoryData{List: []model.PlayHistoryItem{{Title: "当前历史", History: model.PlayHistory{Bvid: "BVcurrent"}}}, Cursor: model.PlayHistoryCursor{Max: 1}}, nil
	}
	s.loadPlayHistory(false)
	runtime.flush(t)
	s.loadPlayHistory(true)
	s.loadPlayHistory(false)
	runtime.finish(t, 1)
	runtime.finish(t, 0)
	if len(s.playHistory.Data.List) != 1 || s.playHistory.Data.List[0].History.Bvid != "BVcurrent" {
		t.Fatalf("old page reappeared after refresh: %+v", s.playHistory.Data.List)
	}
}
