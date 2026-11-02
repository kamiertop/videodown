package gui

import (
	"fmt"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/navigation"
)

func TestShellNavigationAndInput(t *testing.T) {
	s := NewShell()
	test := ui.NewTester(s.view, 1200, 880)
	test.Frame()
	test.Advance(300 * time.Millisecond)
	click := func(label string) {
		t.Helper()
		if err := test.Click(label); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		test.Advance(300 * time.Millisecond)
	}
	click("进入BiliBili")
	if s.router.Path() != "/bilibili/profile" || !test.HasText("扫码登录") {
		t.Fatal("platform entry must open account page")
	}
	click("视频下载")
	click("BiliBili视频链接")
	test.Type("https://www.bilibili.com/video/example")
	test.Frame()
	test.Advance(300 * time.Millisecond)
	click("收藏夹")
	if s.router.Path() != "/bilibili/favorite" {
		t.Fatalf("section route: %s", s.router.Path())
	}
	click("抖音")
	if s.router.Path() != "/douyin" {
		t.Fatalf("top navigation route: %s", s.router.Path())
	}
	click("抖音视频链接")
	test.Type("https://www.douyin.com/video/example")
	test.Frame()
	if got := *s.pages.Douyin.Link(); got != "https://www.douyin.com/video/example" {
		t.Fatalf("douyin input not retained: %q", got)
	}
	click("前往B 站")
	if got := *s.pages.Bilibili.Link(); got != "https://www.bilibili.com/video/example" {
		t.Fatalf("link lost after switching platform: %q", got)
	}
	click("后退")
	if s.router.Path() != "/douyin" {
		t.Fatalf("back: %s", s.router.Path())
	}
	click("前进")
	if s.router.Path() != "/bilibili" {
		t.Fatalf("forward: %s", s.router.Path())
	}
	click("打开设置")
	if !test.HasText("窗口关闭行为") {
		t.Fatal("settings page missing")
	}
	click("关于")
	if s.router.Path() != "/settings/about" {
		t.Fatalf("settings route: %s", s.router.Path())
	}
}

func TestShellAllRoutesAndSizes(t *testing.T) {
	s := NewShell()
	test := ui.NewTester(s.view, 1200, 880)
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		test.SetSize(size[0], size[1])
		for _, p := range navigation.All() {
			paths := []string{p.Path()}
			for _, section := range p.Sections() {
				if section.ID != "" {
					paths = append(paths, p.Path()+"/"+section.ID)
				}
			}
			for _, path := range paths {
				s.router.Push(path)
				test.Frame()
				test.Advance(300 * time.Millisecond)
				test.Frame()
				if path == p.Path() && p.ID == "douyin" {
					for _, label := range []string{"抖音下载", "抖音用户", "抖音动态", "抖音收藏 / 合集", "抖音播放历史", "抖音下载历史", "抖音账号"} {
						bounds, ok := test.Find(label)
						if !ok || bounds.W < 60 || bounds.W > 88 {
							t.Fatalf("douyin menu %s at %v: %+v", label, size, bounds)
						}
					}
					if size[0] == 1200 {
						file, err := os.Create("/tmp/" + p.ID + "-navigation.png")
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
				if p.ID == "bilibili" && path == "/bilibili/favorite" {
					f, err := os.Create(fmt.Sprintf("/tmp/bilibili-native-%d.png", size[0]))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(f, test.Image())
					f.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
				if test.HasText("页面不存在：" + path) {
					t.Fatalf("missing route: %s", path)
				}
				if _, ok := navigation.ByPath(path); !ok {
					t.Fatalf("unresolved route: %s", path)
				}
			}
		}
	}
	s.router.Push("/")
	test.SetSize(1200, 880)
	test.Frame()
	test.Advance(300 * time.Millisecond)
	f, err := os.Create("/tmp/shell-home.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, test.Image()); err != nil {
		t.Fatal(err)
	}
	s.router.Push("/settings")
	test.SetSize(760, 600)
	test.Frame()
	test.Advance(300 * time.Millisecond)
	narrow, err := os.Create("/tmp/shell-settings.png")
	if err != nil {
		t.Fatal(err)
	}
	defer narrow.Close()
	if err := png.Encode(narrow, test.Image()); err != nil {
		t.Fatal(err)
	}
}

func TestSettingsFormState(t *testing.T) {
	s := NewShell()
	s.router.Push("/settings/filename")
	test := ui.NewTester(s.view, 1200, 880)
	test.Frame()
	if err := test.Click("{author}"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if !test.HasText("示例视频 - 作者.mp4") {
		t.Fatal("filename preview did not update")
	}
	s.router.Push("/settings/download")
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if err := test.Click("一键下载全部时自动跳过已下载内容开关"); err != nil {
		t.Fatal(err)
	}
	test.Frame()

	s.router.Push("/settings/filename")
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if !test.HasText("示例视频 - 作者.mp4") {
		t.Fatal("settings draft lost after switching page")
	}
	s.router.Push("/settings")
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if err := test.Click("light - 浅色模式"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if err := test.Click("dark - 深色模式"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if s.pages.Settings.Mode != 2 {
		t.Fatalf("theme did not change: %d", s.pages.Settings.Mode)
	}
}
