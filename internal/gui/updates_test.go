package gui

import (
	"context"
	"errors"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/updater"
)

type fakeUpdateChecker struct {
	calls  int
	result updater.Result
	err    error
}

func (f *fakeUpdateChecker) Check(ctx context.Context) (updater.Result, error) {
	f.calls++
	if err := ctx.Err(); err != nil {
		return updater.Result{}, err
	}
	return f.result, f.err
}
func TestStartupUpdateCheckIsDeferredOnceAndShowsRelease(t *testing.T) {
	checker := &fakeUpdateChecker{result: updater.Result{Available: true, CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0", ReleaseNotes: "修复下载与卡片布局"}}
	s := NewShellWithOptions(Options{UpdateChecker: checker})
	defer s.Wait()
	var work func() error
	var done func(error)
	scheduled := 0
	s.runUpdateJob = func(w func() error, d func(error)) { scheduled++; work, done = w, d }
	s.startUpdateCheck()
	s.startUpdateCheck()
	if scheduled != 1 || checker.calls != 0 || s.updateOpen {
		t.Fatal("startup check should be scheduled once without blocking")
	}
	done(work())
	if checker.calls != 1 || !s.updateOpen {
		t.Fatal("new release should open the update prompt")
	}
	view := ui.NewTester(s.view, 1200, 800)
	view.Frame()
	view.Advance(300 * time.Millisecond)
	for _, text := range []string{"发现新版本", "最新版本：v2.0.0", "当前版本：v1.0.0", "修复下载与卡片布局", "查看更新"} {
		if !view.HasText(text) {
			t.Fatalf("missing update content %q", text)
		}
	}
	f, err := os.Create("/tmp/videodown-update-dialog.png")
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, view.Image()); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	if err := view.Click("稍后再说"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	view.Advance(300 * time.Millisecond)
	if s.updateOpen {
		t.Fatal("dismiss did not close the prompt")
	}
	s.startUpdateCheck()
	if scheduled != 1 {
		t.Fatal("dismiss restarted update check")
	}
}
func TestStartupUpdatePreferencesFailuresAndClosure(t *testing.T) {
	for _, test := range []struct {
		name                                    string
		disableInitially, disableDuring, cancel bool
		err                                     error
		available                               bool
	}{
		{name: "disabled", disableInitially: true, available: true},
		{name: "disabled while checking", disableDuring: true, available: true},
		{name: "closed", cancel: true, available: true},
		{name: "network failure", err: errors.New("network unavailable"), available: true},
		{name: "already current"},
	} {
		t.Run(test.name, func(t *testing.T) {
			checker := &fakeUpdateChecker{result: updater.Result{Available: test.available, CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0"}, err: test.err}
			s := NewShellWithOptions(Options{UpdateChecker: checker})
			defer s.Wait()
			s.pages.Settings.Config.AutoUpdate = !test.disableInitially
			var work func() error
			var done func(error)
			s.runUpdateJob = func(w func() error, d func(error)) { work, done = w, d }
			s.startUpdateCheck()
			if test.disableInitially {
				if work != nil {
					t.Fatal("disabled check scheduled network work")
				}
				return
			}
			err := work()
			if test.disableDuring {
				s.pages.Settings.Config.AutoUpdate = false
			}
			if test.cancel {
				s.stopJobs()
			}
			done(err)
			if s.updateOpen {
				t.Fatal("obsolete, failed, or unnecessary update prompt shown")
			}
		})
	}
}
