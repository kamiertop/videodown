package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/closing"
	"github.com/kamiertop/videodown/internal/storage"
	"testing"
	"time"
)

func TestFirstCloseDialog(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := NewShellWithOptions(Options{Preferences: store})
	s.tray = &mygo.Tray{}
	s.runJob = func(work func() error, done func(error)) { done(work()) }
	s.closing.Show(closing.Choose, 2)
	test := ui.NewTester(s.view, 760, 600)
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if !test.HasText("仍有 2 个任务正在下载") {
		t.Fatal("active download warning missing")
	}
	if err := test.Click("取消"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if s.closing.Open || s.closing.Configured {
		t.Fatal("cancel must not save a preference")
	}
	s.closing.Show(closing.Choose, 0)
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if err := test.Click("最小化到托盘"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if _, configured, _ := store.CloseToTray(); configured {
		t.Fatal("one-time choice was persisted")
	}
	s.closing.Show(closing.Choose, 0)
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if err := test.Click("记住选择，下次不再询问"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if err := test.Click("最小化到托盘"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if enabled, configured, err := store.CloseToTray(); err != nil || !enabled || !configured {
		t.Fatal("remembered choice not saved")
	}
	next := NewShellWithOptions(Options{Preferences: store})
	if !next.closing.Configured || !next.closing.CloseToTray {
		t.Fatal("saved choice not loaded on next startup")
	}
}

func TestCloseBehaviorSwitchShowsSaveFeedback(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := NewShellWithOptions(Options{Preferences: store, Settings: store})
	s.tray = &mygo.Tray{}
	var finish func()
	s.runJob = func(work func() error, done func(error)) { finish = func() { done(work()) } }
	s.router.Push("/settings")
	test := ui.NewTester(s.view, 1200, 1000)
	test.Advance(300 * time.Millisecond)
	for _, expected := range []bool{false, true} {
		if err := test.Click("关闭时缩小到任务栏托盘开关"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if !s.closing.Saving || finish == nil {
			t.Fatal("close behavior switch did not start saving")
		}
		finish()
		test.Frame()
		if s.closing.CloseToTray != expected {
			t.Fatalf("switch saved %v, want %v", s.closing.CloseToTray, expected)
		}
		if !test.HasText("保存成功") {
			t.Fatal("close behavior switch did not show save feedback")
		}
		test.Advance(4 * time.Second)
		if test.HasText("保存成功") {
			t.Fatal("close behavior save feedback did not expire")
		}
	}
}

func TestCloseBehaviorSwitchReportsSaveFailure(t *testing.T) {
	s := NewShell()
	s.tray = &mygo.Tray{}
	// No preferences store: the attempted write must fail and keep the old value.
	s.runJob = func(work func() error, done func(error)) { done(work()) }
	s.router.Push("/settings")
	test := ui.NewTester(s.view, 1200, 1000)
	test.Advance(300 * time.Millisecond)
	if err := test.Click("关闭时缩小到任务栏托盘开关"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if !s.closing.CloseToTray || s.closing.Configured || s.closing.Error == "" {
		t.Fatal("failed save did not preserve close behavior and report the error")
	}
	if s.settingsToast == nil || !s.settingsToast.failed {
		t.Fatal("failed save did not show an error notification")
	}
	test.Advance(7 * time.Second)
	if s.settingsToast != nil {
		t.Fatal("error notification did not expire")
	}
}
