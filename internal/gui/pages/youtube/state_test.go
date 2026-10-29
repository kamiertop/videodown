package youtube

import (
	"context"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	yt "github.com/kamiertop/videodown/youtube"
)

type fakeAPI struct {
	cookie   string
	canceled bool
}

func (a *fakeAPI) Cookie() string           { return a.cookie }
func (a *fakeAPI) SetCookie(v string) error { a.cookie = v; return nil }
func (a *fakeAPI) Resolve(ctx context.Context, v string) (yt.Video, error) {
	return yt.Video{ID: "jNQXAC9IVRw", Title: "YouTube 测试视频", Author: "测试作者", Formats: []yt.Format{{ID: 18, Label: "360p · H.264 · MP4 · 30帧 · 1.20 Mbps · 含音频 · #18", Audio: true}, {ID: 137, Label: "1080p · AV1 · MP4 · 60帧 · 2.50 Mbps · 需合并音频 · #399"}}}, nil
}
func (a *fakeAPI) History() ([]yt.History, error) { return nil, nil }
func (a *fakeAPI) Download(ctx context.Context, v yt.Video, id int, notify func(yt.Progress)) (string, error) {
	if ctx.Err() != nil {
		a.canceled = true
		return "", ctx.Err()
	}
	notify(yt.Progress{Phase: "已完成", Percent: 100})
	return "/tmp/video.mp4", nil
}
func TestFormResolveDownloadCancelAndCookie(t *testing.T) {
	type job struct {
		work func() error
		done func(error)
	}
	var jobs []job
	api := &fakeAPI{}
	s := NewState()
	s.Configure(api, guiruntime.Runtime{Background: func(work func() error, done func(error)) { jobs = append(jobs, job{work, done}) }}, func(fn func()) { fn() })
	flush := func() {
		t.Helper()
		for len(jobs) > 0 {
			j := jobs[0]
			jobs = jobs[1:]
			j.done(j.work())
		}
	}
	s.Link = "https://youtu.be/jNQXAC9IVRw"
	s.resolve()
	flush()
	if s.video.ID == "" || s.busy {
		t.Fatal("resolve failed")
	}
	s.download()
	if s.PendingDownloadCount() != 1 {
		t.Fatal("download not counted")
	}
	s.cancel()
	flush()
	if !api.canceled || s.PendingDownloadCount() != 0 {
		t.Fatal("cancellation failed")
	}
	s.download()
	flush()
	if s.output != "/tmp/video.mp4" || s.progress.Percent != 100 {
		t.Fatal("download completion failed")
	}
	for _, width := range []int{760, 1200} {
		scroll := &ui.ScrollState{}
		test := ui.NewTester(func(c *ui.Context) { core.Use(c, core.Settings{}); s.Form(c, "/youtube/download", scroll) }, width, 600)
		test.Frame()
		card, _ := test.Find("YouTube 下载视频卡片")
		cover, _ := test.Find("YouTube 视频封面")
		selectBox, _ := test.Find("YouTube 下载画质")
		download, _ := test.Find("下载视频")
		if card.H > 120 || selectBox.Y >= cover.Y+cover.H || download.Y >= cover.Y+cover.H {
			t.Fatalf("card too tall or controls below cover: card=%+v cover=%+v select=%+v download=%+v", card, cover, selectBox, download)
		}
		for _, label := range []string{"YouTube 下载画质", "下载视频", "YouTube 测试视频"} {
			if bounds, ok := test.Find(label); !ok || bounds.X+bounds.W > float32(width) {
				t.Fatalf("clipped %s: %+v", label, bounds)
			}
		}
		if width == 760 {
			f, err := os.Create("/tmp/youtube-download-ready.png")
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
	s.cookie = "SID=fixture"
	s.saveCookie()
	flush()
	if api.cookie != "SID=fixture" || s.video.ID != "" {
		t.Fatal("cookie save did not clear stale video")
	}
	s.resolve()
	s.Stop()
	flush()
	if s.video.ID != "" {
		t.Fatal("closed page accepted result")
	}
}

func TestSidebarAccountCookieAndHistory(t *testing.T) {
	for _, width := range []int{760, 1200} {
		s := NewState()
		api := &fakeAPI{}
		s.Configure(api, guiruntime.Runtime{Background: func(work func() error, done func(error)) { done(work()) }}, func(fn func()) { fn() })
		s.Link = "https://youtu.be/jNQXAC9IVRw"
		s.resolve()
		path := "/youtube/download"
		navigate := func(next string) { path = next }
		scrolls := map[string]*ui.ScrollState{}
		test := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			if scrolls[path] == nil {
				scrolls[path] = &ui.ScrollState{}
			}
			s.Layout(c, path, navigate, func() { s.Form(c, path, scrolls[path]) })
		}, width, 600)
		test.Frame()
		rail, _ := test.Find("YouTube 导航")
		body, _ := test.Find("YouTube 功能区域")
		if rail.W <= 0 || rail.X+rail.W > body.X || body.X+body.W > float32(width) {
			t.Fatalf("sidebar overlaps content: %+v %+v", rail, body)
		}
		if _, ok := test.Find("YouTube Cookie"); ok {
			t.Fatal("cookie editor leaked into download page")
		}
		if err := test.Click("YouTube账号"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		field, ok := test.Find("YouTube Cookie")
		if !ok || field.X < body.X || field.X+field.W > float32(width) {
			t.Fatal("cookie editor outside account content")
		}
		if err := test.Click("YouTube Cookie"); err != nil {
			t.Fatal(err)
		}
		test.Type("SID=fixture")
		test.Frame()
		if err := test.Click("保存 Cookie"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if api.cookie != "SID=fixture" || !test.HasText("已设置 Cookie") {
			t.Fatal("account cookie was not saved")
		}
		if width == 760 {
			f, err := os.Create("/tmp/youtube-account-sidebar.png")
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, test.Image())
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := test.Click("YouTube下载历史"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if !test.HasText("暂无下载记录") {
			t.Fatal("missing history page")
		}
		if err := test.Click("YouTube下载"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if s.Link != "https://youtu.be/jNQXAC9IVRw" {
			t.Fatal("navigation lost download input")
		}
		if width == 760 {
			f, err := os.Create("/tmp/youtube-download-sidebar.png")
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

func TestCompactCardWithSidebarAndFormatPopup(t *testing.T) {
	s := NewState()
	api := &fakeAPI{}
	s.Configure(api, guiruntime.Runtime{Background: func(work func() error, done func(error)) { done(work()) }}, func(fn func()) { fn() })
	s.resolve()
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, "/youtube/download", func(string) {}, func() { s.Form(c, "/youtube/download", scroll) })
	}, 760, 600)
	test.Frame()
	for _, label := range []string{"YouTube 下载画质", "下载视频"} {
		b, ok := test.Find(label)
		if !ok || b.X+b.W > 760 {
			t.Fatalf("clipped %s: %+v", label, b)
		}
	}
	save := func(path string) {
		t.Helper()
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, test.Image())
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	save("/tmp/youtube-compact-sidebar.png")
	if err := test.Click("YouTube 下载画质"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	test.Advance(300000000)
	for _, format := range s.video.Formats {
		heading, details := qualityLines(format.Label)
		if !test.HasText(heading) || (details != "" && !test.HasText(details)) {
			t.Fatalf("format detail missing: %s", format.Label)
		}
	}
	popup, ok := test.Find("YouTube 画质选项")
	trigger, triggerOK := test.Find("YouTube 下载画质")
	if !ok || !triggerOK || popup.W != trigger.W || popup.X+popup.W > 760 {
		t.Fatalf("format popup must match trigger width: popup=%+v trigger=%+v", popup, trigger)
	}
	save("/tmp/youtube-format-popup.png")
}

func TestTransferStatusShowsSpeedBeforeLongDetails(t *testing.T) {
	text := transferStatus(yt.Progress{Phase: "下载视频", Downloaded: 30 << 20, Total: 100 << 20, Speed: 2 << 20, Percent: 27})
	if !strings.HasPrefix(text, "2.0 MB/s · 27%") || !strings.Contains(text, "30.0 MB / 100.0 MB") {
		t.Fatalf("missing transfer details: %s", text)
	}
}
