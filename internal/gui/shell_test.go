package gui

import (
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestShellCornerAlpha renders the shell's view headlessly: outside the
// rounded frame, at the window's corners, the frame must stay transparent,
// so the desktop shows through them.
func TestShellCornerAlpha(t *testing.T) {
	s := NewShell()
	test := ui.NewTester(s.view, 1200, 880)
	test.SetScale(2)
	test.Frame()
	img := test.Image()
	if img == nil {
		t.Fatal("no image")
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for _, p := range [][2]int{{2, 2}, {w - 3, 2}, {2, h - 3}, {w - 3, h - 3}} {
		if a := img.RGBAAt(p[0], p[1]).A; a != 0 {
			t.Errorf("corner (%d,%d) alpha = %d, want 0 (transparent)", p[0], p[1], a)
		}
	}
	if a := img.RGBAAt(w/2, h/2).A; a == 0 {
		t.Error("center alpha = 0, want opaque content")
	}
	f, err := os.Create("/tmp/shell-frame.png")
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()
}

// TestRouterHistory checks the shell's navigation history: pushing walks
// forward, Back and Forward move through it, and a push cuts the pages
// ahead, as a browser's history does.
func TestRouterHistory(t *testing.T) {
	r := ui.NewRouter("/bilibili")
	if r.Path() != "/bilibili" || r.CanGoBack() || r.CanGoForward() {
		t.Fatalf("start: path=%s back=%v forward=%v", r.Path(), r.CanGoBack(), r.CanGoForward())
	}
	r.Push("/bilibili/download")
	r.Push("/settings")
	if !r.CanGoBack() || r.CanGoForward() {
		t.Fatal("after pushes: CanGoBack false, CanGoForward true expected")
	}
	r.Back()
	if r.Path() != "/bilibili/download" || !r.CanGoForward() {
		t.Fatalf("after Back: path=%s", r.Path())
	}
	r.Forward()
	if r.Path() != "/settings" {
		t.Fatalf("after Forward: path=%s", r.Path())
	}
	r.Back()
	r.Back()
	if r.Path() != "/bilibili" || r.CanGoBack() {
		t.Fatalf("after Back×2: path=%s", r.Path())
	}
	r.Push("/douyin")
	if r.CanGoForward() {
		t.Fatal("push cut the forward history: CanGoForward expected false")
	}
}

// The landing page must center all platform entry targets in the available
// content area, with no welcome cards or form controls.
func TestHomeLayout(t *testing.T) {
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		s := NewShell()
		test := ui.NewTester(s.view, size[0], size[1])
		test.Frame()
		content, ok := test.Find("平台入口")
		if !ok {
			t.Fatal("home content missing")
		}
		bili, okBili := test.Find("进入BiliBili")
		douyin, okDouyin := test.Find("进入抖音")
		youtube, okYouTube := test.Find("进入YouTube")
		if !okBili || !okDouyin || !okYouTube {
			t.Fatal("platform entries missing")
		}
		centerX := content.X + content.W/2
		centerY := content.Y + content.H/2
		if math.Abs(float64((bili.X+youtube.X+youtube.W)/2-centerX)) > 1 || math.Abs(float64(bili.Y+bili.H/2-centerY)) > 1 || bili.Y != douyin.Y || bili.Y != youtube.Y {
			t.Fatalf("entries are not centered at %v: content=%+v bili=%+v douyin=%+v", size, content, bili, douyin)
		}
		if bili.X < content.X || youtube.X+youtube.W > content.X+content.W || bili.Y < content.Y || bili.Y+bili.H > content.Y+content.H {
			t.Fatal("home entries overflow")
		}
		if test.HasText("欢迎使用") || test.HasText("快捷操作") {
			t.Fatal("home contains removed content")
		}
		if err := test.Click("进入抖音"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if s.router.Path() != "/douyin/profile" {
			t.Fatalf("Douyin entry: %s", s.router.Path())
		}
	}
}
