package bilibili

import (
	"testing"

	"github.com/egoist/mygo/ui"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"

	"github.com/kamiertop/videodown/internal/gui/style"
)

// TestSleepIndicatorStaysBelowProgressBar pins the download card's status
// order: the progress bar keeps the first row whatever the phase, and the
// sleep spinner joins below it instead of pushing the bar down.
func TestSleepIndicatorStaysBelowProgressBar(t *testing.T) {
	progress := bilidownload.Progress{Phase: "sleep", SleepTotal: 10, SleepRemaining: 3}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		ui.Column(c).FillWidth().Children(func() {
			ui.Text(c, "卡片标题")
			downloadProgress(c, progress)
		})
	}, 400, 120)
	test.Frame()
	title, ok := test.Find("卡片标题")
	if !ok {
		t.Fatal("missing title anchor")
	}
	bar, ok := test.Find("休眠进度")
	if !ok {
		t.Fatal("missing sleep progress bar")
	}
	spinner, ok := test.Find("休眠中 · 3 秒")
	if !ok {
		t.Fatal("missing sleep spinner")
	}
	// 进度条必须仍是状态区的第一行，紧跟标题，不被休眠行顶下去。
	if diff := bar.Y - (title.Y + title.H); diff < -0.5 || diff > 0.5 {
		t.Fatalf("progress bar moved %.2f DIPs below the title; it must stay the first row", diff)
	}
	// 休眠指示在进度条下面。
	if spinner.Y < bar.Y+bar.H-0.5 {
		t.Fatalf("sleep spinner at Y=%.2f overlaps or precedes the bar (Y=%.2f, H=%.2f)", spinner.Y, bar.Y, bar.H)
	}
}
