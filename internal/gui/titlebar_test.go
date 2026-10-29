package gui

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestTitleBarNavigation(t *testing.T) {
	for _, width := range []int{760, 1200} {
		s := NewShell()
		test := ui.NewTester(s.view, width, 600)
		test.Frame()
		for _, label := range []string{"前往首页", "前往B 站", "前往抖音", "打开 GitHub 仓库", "打开设置", "最小化", "最大化", "关闭"} {
			r, ok := test.Find(label)
			if !ok || r.W <= 0 || r.X < 0 || r.X+r.W > float32(width) {
				t.Errorf("width=%d %s: bounds=%+v exists=%v", width, label, r, ok)
			}
		}
		brand, ok := test.Find("应用标识")
		if !ok || math.Abs(float64(brand.X+brand.W/2)-float64(width)/2) > 1 {
			t.Errorf("brand is not centered at width %d: %+v", width, brand)
		}
		nav, _ := test.Find("前往抖音")
		github, _ := test.Find("打开 GitHub 仓库")
		if nav.X+nav.W > brand.X || github.X < brand.X+brand.W {
			t.Errorf("navigation overlaps brand at width %d", width)
		}
		if err := test.Click("前往B 站"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if s.router.Path() != "/bilibili" {
			t.Fatalf("Bilibili navigation: %s", s.router.Path())
		}
		if err := test.Click("前往抖音"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if s.router.Path() != "/douyin" {
			t.Fatalf("Douyin navigation: %s", s.router.Path())
		}
		if err := test.Click("前往首页"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if s.router.Path() != "/" {
			t.Fatalf("Home navigation: %s", s.router.Path())
		}
	}
}
