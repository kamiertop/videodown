package youtube

import (
	"errors"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/internal/storage"
	yt "github.com/kamiertop/videodown/youtube"
)

func TestHistoryRowsAlignAndShowPaths(t *testing.T) {
	for _, width := range []int{760, 1200} {
		s := NewState()
		s.history = []yt.History{
			{Title: "短视频标题", Path: "/download/YouTube/短视频.mp4"},
			{Title: "这是一个很长的下载历史标题，用来检查标题和保存路径是否对齐，以及打开目录按钮是否会被挤出窗口", Path: "/download/作者/YouTube/较长的目录名称/这是一个很长的视频文件名.mkv"},
		}
		var opened, played string
		s.runtime.OpenFile = func(path string) error { played = path; return nil }
		s.runtime.OpenLocation = func(path string) error { opened = path; return nil }
		scroll := &ui.ScrollState{}
		test := ui.NewTester(func(c *ui.Context) {
			style.UseTheme(c)
			s.Layout(c, "/youtube/history", func(string) {}, func() { s.Form(c, "/youtube/history", scroll) })
		}, width, 600)
		test.Frame()
		var previousRow, previousButton, previousTitle ui.Rect
		for i, item := range s.history {
			row, ok := test.Find("下载记录：" + item.Path)
			if !ok {
				t.Fatal("missing history row")
			}
			title, ok := test.Find(item.Title)
			if !ok {
				t.Fatal("missing title")
			}
			path, ok := test.Find(item.Path)
			if !ok || path.Y <= title.Y {
				t.Fatal("path is not below title")
			}
			play, ok := test.Find("打开视频：" + item.Path)
			if !ok {
				t.Fatal("missing play action")
			}
			button, ok := test.Find("打开目录：" + item.Path)
			if !ok || button.X+button.W > float32(width) || title.X+title.W > play.X || play.X+play.W > button.X || button.W > 100 || button.H > 36 {
				t.Fatal("history columns overlap")
			}
			del, ok := test.Find("删除记录：" + item.Path)
			if !ok || button.X+button.W > del.X {
				t.Fatal("delete action missing or misplaced")
			}
			// 操作与标题同行并垂直居中（与 B 站/抖音卡片一致）。
			if math.Abs(float64(button.Y+button.H/2-title.Y-title.H/2)) > 8 {
				t.Fatal("button is not aligned with the title row")
			}
			if i > 0 && (row.X != previousRow.X || math.Abs(float64(row.H-previousRow.H)) > 1 || button.X != previousButton.X || title.X != previousTitle.X) {
				// 行高允许 1 DIP 的字体度量舍入差（中英文路径字形不同）。
				t.Fatal("history rows are not aligned")
			}
			previousRow, previousButton, previousTitle = row, button, title
		}
		if err := test.Click("打开视频：" + s.history[0].Path); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if played != s.history[0].Path {
			t.Fatal("opened wrong video")
		}
		if err := test.Click("打开目录：" + s.history[1].Path); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if opened != s.history[1].Path {
			t.Fatal("opened the wrong directory")
		}
		if width == 760 {
			f, err := os.Create(filepath.Join(os.TempDir(), "youtube-history-aligned.png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, test.Image())
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestHistorySearchAndClear(t *testing.T) {
	s := NewState()
	s.history = []yt.History{
		{ID: "AbC123", Title: "旅行视频", Author: "Alice", Path: "/download/travel.mp4"},
		{ID: "xyz456", Title: "音乐现场", Author: "Bob", Path: "/download/music.mkv"},
	}
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.historyPage(c, scroll)
	}, 760, 600)
	test.Frame()
	if err := test.Click("搜索 YouTube 下载历史"); err != nil {
		t.Fatal(err)
	}
	test.Type("旅行")
	test.Frame()
	if !test.HasText("旅行视频") || test.HasText("音乐现场") || !test.HasText("下载历史 · 1") {
		t.Fatal("typing did not filter history")
	}
	for _, query := range []string{" ALICE ", "abc123", "TRAVEL.MP4"} {
		s.historySearch = query
		test.Frame()
		if !test.HasText("旅行视频") || test.HasText("音乐现场") {
			t.Fatalf("search failed for %q", query)
		}
	}
	s.historySearch = "不存在"
	test.Frame()
	if !test.HasText("没有匹配的下载历史") || test.HasText("旅行视频") {
		t.Fatal("missing search empty state")
	}
	if err := test.Click("清除搜索"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if !test.HasText("旅行视频") || !test.HasText("音乐现场") || !test.HasText("下载历史 · 2") || len(s.history) != 2 {
		t.Fatal("clearing search did not restore history")
	}
}

// TestHistoryOpenFailureShowsMessage verifies that opening a moved or
// deleted file reports the reason as a toast instead of failing silently.
func TestHistoryOpenFailureShowsMessage(t *testing.T) {
	s := NewState()
	s.history = []yt.History{{ID: "gone1", Title: "被移动的视频", Path: "/download/moved/video.mp4"}}
	s.runtime.OpenFile = func(string) error { return errors.New("文件不存在") }
	s.runtime.OpenLocation = func(string) error { return errors.New("文件所在目录不存在") }
	var toasts []string
	s.runtime.Notify = func(message string, failed bool) {
		if failed {
			toasts = append(toasts, message)
		}
	}
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.Layout(c, "/youtube/history", func(string) {}, func() { s.Form(c, "/youtube/history", scroll) })
	}, 1200, 600)
	test.Frame()
	if err := test.Click("打开视频：/download/moved/video.mp4"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(toasts) != 1 || toasts[0] != "打开失败：文件不存在" {
		t.Fatalf("open failure toast = %v", toasts)
	}
	if err := test.Click("打开目录：/download/moved/video.mp4"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(toasts) != 2 || toasts[1] != "打开失败：文件所在目录不存在" {
		t.Fatalf("open-location failure toast = %v", toasts)
	}
	// 成功打开不再弹出错误。
	s.runtime.OpenFile = func(string) error { return nil }
	if err := test.Click("打开视频：/download/moved/video.mp4"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(toasts) != 2 {
		t.Fatal("successful open raised an error toast")
	}
}

// TestHistoryOpenFailureInlineFallback verifies the page-level red note when
// no toast channel exists (tests and offscreen rendering).
func TestHistoryOpenFailureInlineFallback(t *testing.T) {
	s := NewState()
	s.history = []yt.History{{ID: "gone2", Title: "被移动的视频", Path: "/download/moved/video.mp4"}}
	s.runtime.OpenFile = func(string) error { return errors.New("文件不存在") }
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.Layout(c, "/youtube/history", func(string) {}, func() { s.Form(c, "/youtube/history", scroll) })
	}, 1200, 600)
	test.Frame()
	if err := test.Click("打开视频：/download/moved/video.mp4"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if !test.HasText("打开失败：文件不存在") {
		t.Fatal("inline fallback message missing")
	}
}

// TestHistoryDeleteRecord verifies the delete action removes the record from
// the list, matching the bilibili and douyin history cards.
func TestHistoryDeleteRecord(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	api := yt.New(store)
	if err := store.SaveHistory("youtube:downloaded:vid1:18", `{"id":"vid1","title":"旅行视频","path":"/download/travel.mp4"}`); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveHistory("youtube:downloaded:vid2:22", `{"id":"vid2","title":"音乐现场","path":"/download/music.mkv"}`); err != nil {
		t.Fatal(err)
	}

	if err := store.SaveHistory("youtube:downloaded:vid1:cover", `{"id":"vid1","kind":"cover","title":"旅行封面","path":"/download/travel.jpg"}`); err != nil {
		t.Fatal(err)
	}

	s := NewState()
	s.api = api
	s.runtime.Run = func(work func() error, done func(error)) { done(work()) }
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.Layout(c, "/youtube/history", func(string) {}, func() { s.Form(c, "/youtube/history", scroll) })
	}, 1200, 600)
	// 模拟历史加载完成。
	s.mergeHistory([]yt.History{
		{ID: "vid1", Title: "旅行视频", Path: "/download/travel.mp4"},
		{ID: "vid2", Title: "音乐现场", Path: "/download/music.mkv"},
		{ID: "vid1", Kind: "cover", Title: "旅行封面", Path: "/download/travel.jpg"},
	})
	test.Frame()
	if !test.HasText("旅行视频") || !test.HasText("音乐现场") || !test.HasText("旅行封面") {
		t.Fatal("history items missing")
	}
	if err := test.Click("删除记录：/download/travel.jpg"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if test.HasText("旅行封面") {
		t.Fatal("deleted record still listed")
	}
	if !test.HasText("音乐现场") || !test.HasText("旅行视频") {
		t.Fatal("delete removed the wrong record")
	}
	// 存储层也已删除。
	rows, err := api.History()
	if err != nil || len(rows) != 2 || rows[0].Kind == "cover" || rows[1].Kind == "cover" {
		t.Fatalf("store rows after delete: %v (err=%v)", rows, err)
	}
}
