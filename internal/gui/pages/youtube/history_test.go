package youtube

import (
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
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
			core.Use(c, core.Settings{})
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
			if math.Abs(float64(button.Y+button.H/2-row.Y-row.H/2)) > 1 {
				t.Fatal("button is not vertically centered")
			}
			if i > 0 && (row.X != previousRow.X || row.H != previousRow.H || button.X != previousButton.X || title.X != previousTitle.X) {
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
			f, err := os.Create("/tmp/youtube-history-aligned.png")
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
		core.Use(c, core.Settings{})
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
