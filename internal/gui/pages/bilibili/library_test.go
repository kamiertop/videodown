package bilibili

import (
	"errors"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// TestHistoryOpenFailureShowsToast verifies that opening a moved or deleted
// file reports the reason as a toast instead of failing silently, matching
// the YouTube history page.
func TestHistoryOpenFailureShowsToast(t *testing.T) {
	s := NewState()
	s.history.page.Data.Items = []model.DownloadHistoryItem{{Cid: 1, Title: "被移动的视频", Path: "/download/moved/video.mp4"}}
	s.runtime.OpenFile = func(string) error { return errors.New("文件不存在") }
	s.runtime.OpenLocation = func(string) error { return errors.New("文件所在目录不存在") }
	s.runtime.Run = func(work func() error, done func(error)) { done(work()) }
	var toasts []string
	s.runtime.Notify = func(message string, failed bool) {
		if failed {
			toasts = append(toasts, message)
		}
	}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.downloadHistory(c)
	}, 1200, 720)
	test.Frame()
	if err := test.Click("本地打开"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(toasts) != 1 || toasts[0] != "打开失败：文件不存在" {
		t.Fatalf("open failure toast = %v", toasts)
	}
	if err := test.Click("打开目录"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(toasts) != 2 || toasts[1] != "打开失败：文件所在目录不存在" {
		t.Fatalf("open-location failure toast = %v", toasts)
	}
	// 成功打开不再弹出错误。
	s.runtime.OpenFile = func(string) error { return nil }
	if err := test.Click("本地打开"); err != nil {
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
	s.history.page.Data.Items = []model.DownloadHistoryItem{{Cid: 2, Title: "被移动的视频", Path: "/download/moved/video.mp4"}}
	s.runtime.OpenFile = func(string) error { return errors.New("文件不存在") }
	s.runtime.Run = func(work func() error, done func(error)) { done(work()) }
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.downloadHistory(c)
	}, 1200, 720)
	test.Frame()
	if err := test.Click("本地打开"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if !test.HasText("打开失败：文件不存在") {
		t.Fatal("inline fallback message missing")
	}
}
