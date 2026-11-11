package gui

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Exercise the shell router as well as the page: state-only tests cannot
// detect the fade/remount applied around Douyin.Form by the shell.
func TestDouyinFolderRoutesDoNotFadeWholeForm(t *testing.T) {
	for _, kind := range []string{"collection", "folder"} {
		t.Run(kind, func(t *testing.T) {
			s := NewShell()
			root := "/douyin/favorite/" + kind
			s.router.Push(root)
			test := ui.NewTester(s.view, 1200, 880)
			test.Advance(300 * time.Millisecond)
			bounds, ok := test.Find("登录提示")
			if !ok {
				t.Fatal("missing Douyin content")
			}
			before := test.Image()
			for _, path := range []string{root + "/one", root + "/two", root} {
				s.router.Push(path)
				test.Frame()
				for frame := range 20 {
					test.Advance(16 * time.Millisecond)
					got, ok := test.Find("登录提示")
					if !ok || got != bounds {
						t.Fatal("content moved during folder switch")
					}
					after := test.Image()
					// The unchanged form must remain fully visible throughout navigation.
					for y := int(bounds.Y) + 2; y < int(bounds.Y+bounds.H)-2; y++ {
						for x := int(bounds.X) + 2; x < int(bounds.X+bounds.W)-2; x++ {
							if before.RGBAAt(x, y) != after.RGBAAt(x, y) {
								t.Fatalf("form faded during %s at frame %d", path, frame)
							}
						}
					}
				}
			}
		})
	}
}
