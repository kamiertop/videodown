package douyin

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	dyDownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func TestDownloadStatusSharesTitleAndQualityRow(t *testing.T) {
	s, _, _, _ := setup(t)
	item := &queuedVideo{Data: fixture("compact-download"), Track: -1, Status: "active"}
	item.Data.Desc = "夏倾月：云澈你怎么还没睡呀，云澈：完了解释不清了。"
	item.Data.Author.Nickname = "动画动漫"
	batch := true
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		ui.Column(c).Fill().Padding(style.PagePadding).Children(func() { s.downloadCard(c, item, batch) })
	}, 1200, 400)
	for _, phase := range []string{"sleep", "video"} {
		item.Progress = dyDownload.Progress{Phase: phase, Percent: 48, SleepRemaining: 52}
		test.Frame()
		card, ok := test.Find("抖音下载作品：" + title(item.Data))
		if !ok || card.H > 90 {
			t.Fatalf("download card still too tall: %+v", card)
		}
		heading, ok := test.Find(title(item.Data))
		if !ok {
			t.Fatal("missing title")
		}
		quality, ok := test.Find("抖音下载画质")
		if !ok {
			t.Fatal("missing quality control")
		}
		progress, ok := test.Find("抖音下载进度")
		if !ok {
			t.Fatal("missing inline progress")
		}
		if math.Abs(float64(heading.Y+heading.H/2-quality.Y-quality.H/2)) > 1 || math.Abs(float64(progress.Y+progress.H/2-quality.Y-quality.H/2)) > 1 {
			t.Fatalf("status not on title row: title=%+v quality=%+v progress=%+v", heading, quality, progress)
		}
		f, err := os.Create(fmt.Sprintf("/tmp/douyin-compact-download-%s.png", phase))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, test.Image()); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
	}
	batch = false
	test.SetSize(760, 400)
	test.Frame()
	for _, label := range []string{"抖音下载画质", "抖音下载进度", "下载中", "封面", "移除"} {
		bounds, ok := test.Find(label)
		if !ok || bounds.X+bounds.W > 760 {
			t.Fatalf("narrow download controls clipped: %s %+v", label, bounds)
		}
	}
}
