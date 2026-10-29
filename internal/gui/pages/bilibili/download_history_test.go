package bilibili

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

type historyTestAPI struct {
	*fakeAPI
	items                []model.DownloadHistoryItem
	failRead, failDelete bool
	reads                []string
	deleted              []string
}

func (a *historyTestAPI) DownloadHistoryPage(offset, limit int, query string) (biliapi.HistoryPage, error) {
	a.reads = append(a.reads, query)
	if a.failRead {
		return biliapi.HistoryPage{}, errors.New("历史读取失败")
	}
	items := []model.DownloadHistoryItem{}
	for _, item := range a.items {
		if strings.Contains(strings.ToLower(item.Title+item.UpperName), strings.ToLower(query)) {
			items = append(items, item)
		}
	}
	end := min(offset+limit, len(items))
	return biliapi.HistoryPage{Items: slices.Clone(items[min(offset, len(items)):end]), Total: len(items), HasMore: end < len(items)}, nil
}
func (a *historyTestAPI) DeleteDownloadHistory(cid int64, kind string) error {
	if a.failDelete {
		return errors.New("无法删除记录")
	}
	a.deleted = append(a.deleted, fmt.Sprintf("%d:%s", cid, kind))
	a.items = slices.DeleteFunc(a.items, func(item model.DownloadHistoryItem) bool { return item.Cid == cid && item.DownloadKind == kind })
	return nil
}
func (a *historyTestAPI) ClearDownloadHistory() error { a.items = nil; return nil }

func historyView(s *State, width int) *ui.Tester {
	path := "/bilibili/history"
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() {}) })
	}, width, 600)
}

func TestDownloadHistoryLocalActionsSearchAndClear(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, base, runtime := setup(t, false)
		api := &historyTestAPI{fakeAPI: base, items: []model.DownloadHistoryItem{
			{Cid: 1, Title: "已保存的视频", UpperName: "测试 UP", Path: "/tmp/video.mp4", Downloaded: "2026-10-28T01:00:00Z"},
			{Cid: 1, Title: "独立音频", DownloadKind: "audio", Path: "/tmp/audio.m4a"},
			{Cid: -1, Title: "独立封面", DownloadKind: "cover", Path: "/tmp/cover.jpg"},
		}}
		rt := runtime.runtime()
		var opened, location string
		rt.OpenFile = func(path string) error { opened = path; return errors.New("文件不存在") }
		rt.OpenLocation = func(path string) error { location = path; return nil }
		s.Configure(api, rt)
		view := historyView(s, width)
		view.Frame()
		runtime.flush(t)
		view.Frame()
		if !view.HasText("下载历史 · 3") || view.HasText("扫码登录") {
			t.Fatal("local history requires login or failed to load")
		}
		click := func(label string) {
			t.Helper()
			if err := view.Click(label); err != nil {
				t.Fatal(err)
			}
			runtime.flush(t)
			view.Frame()
		}
		click("本地打开")
		if opened != "/tmp/video.mp4" || !view.HasText("文件不存在") {
			t.Fatal("open local file/error missing")
		}
		click("打开目录")
		if location != opened {
			t.Fatal("directory action received wrong file")
		}
		if err := view.Click("搜索下载历史"); err != nil {
			t.Fatal(err)
		}
		view.Type("独立音频")
		view.Frame()
		runtime.tick(t)
		runtime.flush(t)
		view.Frame()
		if !view.HasText("下载历史 · 1") || view.HasText("已保存的视频") {
			t.Fatal("history search did not filter")
		}
		api.failDelete = true
		click("删除记录")
		if !view.HasText("无法删除记录") || len(api.items) != 3 {
			t.Fatal("delete failure lost record")
		}
		api.failDelete = false
		click("删除记录")
		if fmt.Sprint(api.deleted) != "[1:audio]" || !view.HasText("没有匹配的下载历史") {
			t.Fatal("audio record deletion is not separate")
		}
		*s.search("history") = ""
		view.Frame()
		runtime.tick(t)
		runtime.flush(t)
		view.Frame()
		toolbar, _ := view.Find("历史工具栏")
		list, _ := view.Find("下载历史列表")
		open, _ := view.Find("本地打开")
		title, _ := view.Find("已保存的视频")
		card, _ := view.Find("下载记录：已保存的视频")
		if open.Y > title.Y+title.H || title.Y > open.Y+open.H || card.H > 115 {
			t.Fatalf("history actions should share the title row: button=%+v title=%+v card=%+v", open, title, card)
		}
		if toolbar.H > 60 || list.Y < toolbar.Y+toolbar.H || open.X+open.W > float32(width) {
			t.Fatal("history layout is clipped")
		}
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-download-history-%d.png", width))
		click("清空")
		if !view.HasText("本地文件会保留") {
			t.Fatal("clear confirmation missing")
		}
		click("取消")
		if len(api.items) != 2 {
			t.Fatal("cancel cleared history")
		}
		click("清空")
		click("确认清空")
		if len(api.items) != 0 || !view.HasText("暂无下载历史") {
			t.Fatal("clear did not refresh")
		}
	}
}

func TestDownloadHistoryAutomaticPagingAndStaleSearch(t *testing.T) {
	s, base, runtime := setup(t, false)
	api := &historyTestAPI{fakeAPI: base}
	for i := 0; i < 75; i++ {
		api.items = append(api.items, model.DownloadHistoryItem{Cid: int64(i + 1), Title: fmt.Sprintf("历史视频 %02d", i)})
	}
	s.Configure(api, runtime.runtime())
	view := historyView(s, 1200)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	if len(s.history.Data.Items) != 50 {
		t.Fatal("initial history page missing")
	}
	area, _ := view.Find("下载历史列表")
	toolbar, _ := view.Find("历史工具栏")
	scroll := s.scroll("download-history-items")
	view.Scroll(area.X+area.W/2, area.Y+area.H/2, 0, scroll.MaxY)
	view.Frame()
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	after, _ := view.Find("历史工具栏")
	if len(s.history.Data.Items) != 75 || s.history.Data.HasMore || toolbar.Y != after.Y {
		t.Fatal("history automatic pagination or fixed toolbar failed")
	}
	// Supersede a pending load before it publishes its old result.
	s.loadDownloadHistory(false)
	*s.search("history") = "不存在"
	s.scheduleHistorySearch()
	runtime.flush(t)
	runtime.tick(t)
	runtime.flush(t)
	if len(s.history.Data.Items) != 0 || s.history.Data.Total != 0 {
		t.Fatal("stale search results returned")
	}
	api.failRead = true
	s.loadDownloadHistory(false)
	runtime.flush(t)
	view.Frame()
	if !view.HasText("历史读取失败") {
		t.Fatal("read failure not visible")
	}
	api.failRead = false
	s.loadDownloadHistory(false)
	runtime.flush(t)
	view.Frame()
	if s.history.Error != "" {
		t.Fatal("history retry failed")
	}
}
