package gui

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/assets"
)

func TestYouTubeEntryAndForm(t *testing.T) {
	if assets.Platform("youtube") == nil {
		t.Fatal("missing YouTube icon")
	}
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		s := NewShell()
		test := ui.NewTester(s.view, size[0], size[1])
		test.Frame()
		if err := test.Click("进入YouTube"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		test.Advance(300 * time.Millisecond)
		if s.router.Path() != "/youtube/download" {
			t.Fatal("home entry did not navigate")
		}
		for _, label := range []string{"YouTube 下载页面", "YouTube 视频链接", "解析视频"} {
			bounds, ok := test.Find(label)
			if !ok || bounds.W <= 0 || bounds.X < 0 || bounds.X+bounds.W > float32(size[0]) {
				t.Fatalf("clipped %s: %+v", label, bounds)
			}
		}
		if err := test.Click("YouTube 视频链接"); err != nil {
			t.Fatal(err)
		}
		test.Type("https://youtu.be/example")
		test.Frame()
		if err := test.Click("前往首页"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		test.Advance(300 * time.Millisecond)
		if err := test.Click("前往YouTube"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		test.Advance(300 * time.Millisecond)
		if s.pages.YouTube.Link != "https://youtu.be/example" {
			t.Fatal("navigation lost input")
		}
		if size[0] == 760 {
			file, err := os.Create("/tmp/youtube-page.png")
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, test.Image())
			file.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
