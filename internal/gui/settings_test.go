package gui

import (
	"image/png"
	"math"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/preferences"
	"github.com/kamiertop/videodown/internal/storage"
)

func TestSettingsLayoutAndPersistence(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InitDefaults("/downloads"); err != nil {
		t.Fatal(err)
	}
	s := NewShellWithOptions(Options{Settings: store, Preferences: store})
	s.runJob = func(work func() error, done func(error)) { done(work()) }
	s.router.Push("/settings")
	test := ui.NewTester(s.view, 1200, 880)
	test.Frame()
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		test.SetSize(size[0], size[1])
		for _, id := range []string{"", "bilibili", "download", "filename", "ffmpeg", "about"} {
			path := "/settings"
			if id != "" {
				path += "/" + id
			}
			s.router.Push(path)
			test.Frame()
			test.Advance(300 * time.Millisecond)
			nav, ok := test.Find("设置分类")
			if !ok || size[0] >= 768 && nav.W != 176 || size[0] < 768 && nav.W < float32(size[0]-2) {
				t.Fatalf("settings category rail at %v: %+v", size, nav)
			}
			form, ok := test.Find("设置表单")
			if !ok || size[0] >= 768 && form.X < nav.X+nav.W || size[0] < 768 && form.Y < nav.Y+nav.H || form.X+form.W > float32(size[0]) {
				t.Fatalf("settings form overflows: %+v", form)
			}
			if id == "bilibili" {
				var first ui.Rect
				for index, label := range []string{"不休眠", "1 秒", "2 秒", "3 秒", "5 秒", "10 秒", "15 秒", "20 秒", "30 秒", "60 秒"} {
					bounds, ok := test.Find(label)
					if !ok {
						t.Fatalf("missing sleep preset %q", label)
					}
					if index == 0 {
						first = bounds
					}
					if math.Abs(float64(bounds.W-first.W)) > .5 || bounds.H != first.H || bounds.X < form.X || bounds.X+bounds.W > form.X+form.W+.5 {
						t.Fatalf("sleep preset dimensions differ at %v: %s %+v, first %+v", size, label, bounds, first)
					}
				}
			}
			if id == "download" {
				slider, ok := test.Find("下载后休眠滑块")
				if !ok || slider.W < 140 {
					t.Fatalf("sleep slider collapsed at %v: %+v", size, slider)
				}
			}
			if id == "filename" {
				for _, label := range []string{"{title}", "{author}", "{publish_date}", "{date}"} {
					bounds, ok := test.Find(label)
					if !ok || bounds.X+bounds.W > float32(size[0]) {
						t.Fatalf("filename field overflows: %s %+v", label, bounds)
					}
				}
			}
			if size[0] == 1200 {
				name := id
				if name == "" {
					name = "general"
				}
				file, err := os.Create("/tmp/settings-" + name + ".png")
				if err != nil {
					t.Fatal(err)
				}
				if err := png.Encode(file, test.Image()); err != nil {
					t.Fatal(err)
				}
				file.Close()
			}
		}
	}
	test.SetSize(1200, 1000)
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
	if value := store.Config().App.Theme; value != "dark" {
		t.Fatalf("theme was not saved: %v", value)
	}
	if err := test.Click("启动时检查新版本开关"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if value := store.Config().App.AutoUpdate; value != false {
		t.Fatalf("auto-update was not saved: %v", value)
	}
	if err := test.Click("下载"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if !test.HasText("下载后休眠") {
		t.Fatal("category click did not switch settings")
	}
	if err := test.Click("120 秒"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if value := store.Config().Download.SleepSeconds; value != 120 {
		t.Fatalf("sleep was not saved: %v", value)
	}
	s.router.Push("/settings/bilibili")
	test.Frame()
	test.Advance(300 * time.Millisecond)
	if err := test.Click("15 秒"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if value := store.Config().Bilibili.ParseSleepSeconds; value != 15 {
		t.Fatalf("parse sleep preset was not saved: %v", value)
	}
	next := NewShellWithOptions(Options{Settings: store, Preferences: store})
	if next.pages.Settings.Config.Theme != "dark" || next.pages.Settings.Config.Sleep != 120 || next.pages.Settings.Config.AutoUpdate || next.pages.Settings.Config.ParseSleep != 15 {
		t.Fatal("settings did not survive a new shell")
	}
}
func TestQueuedSettingsKeepLatestValue(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := NewShellWithOptions(Options{Settings: store, Preferences: store})
	var jobs []func()
	s.runJob = func(work func() error, done func(error)) { jobs = append(jobs, func() { done(work()) }) }
	s.saveSettings(preferences.Changes{Sleep: preferences.Value(int64(10))})
	s.saveSettings(preferences.Changes{Sleep: preferences.Value(int64(30))})
	s.saveSettings(preferences.Changes{Sleep: preferences.Value(int64(120))})
	if len(jobs) != 1 {
		t.Fatal("concurrent config writers were started")
	}
	jobs[0]()
	if len(jobs) != 2 {
		t.Fatal("latest slider edit was not queued")
	}
	jobs[1]()
	if value := store.Config().Download.SleepSeconds; value != 120 || s.settingsBusy {
		t.Fatalf("latest config not committed: %v", value)
	}
}

func TestSleepNumericValueAboveSliderRange(t *testing.T) {
	s := NewShell()
	s.pages.Settings.Config.Sleep = 1234
	s.router.Push("/settings/download")
	test := ui.NewTester(s.view, 1200, 880)
	test.Frame()
	test.Frame()
	if s.pages.Settings.Config.Sleep != 1234 {
		t.Fatal("slider clamped the numeric field's larger supported range")
	}
}

// The settings rail must stay fixed throughout child-route transitions, not
// just settle back into place once the whole-page animation finishes.
func TestSettingsCategoryRailStaysFixed(t *testing.T) {
	s := NewShell()
	s.router.Push("/settings")
	test := ui.NewTester(s.view, 1200, 880)
	test.Advance(300 * time.Millisecond)
	rail, ok := test.Find("设置分类")
	if !ok {
		t.Fatal("missing category rail")
	}
	for _, path := range []string{"/settings/bilibili", "/settings", "/settings/download"} {
		s.router.Push(path)
		test.Frame()
		for frame := 0; frame < 20; frame++ {
			test.Advance(16 * time.Millisecond)
			got, ok := test.Find("设置分类")
			if !ok || got != rail {
				t.Fatalf("category rail moved during transition: want %+v, got %+v", rail, got)
			}
		}
	}
}

func TestSettingsSaveDoesNotMoveFormAndToastExpires(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := NewShellWithOptions(Options{Settings: store})
	var finish func()
	s.runJob = func(work func() error, done func(error)) { finish = func() { done(work()) } }
	s.router.Push("/settings")
	test := ui.NewTester(s.view, 1200, 1000)
	test.Advance(300 * time.Millisecond)
	before, ok := test.Find("选择目录")
	if !ok {
		t.Fatal("missing directory button")
	}
	checkBounds := func() {
		t.Helper()
		got, ok := test.Find("选择目录")
		if !ok || got != before {
			t.Fatalf("saving moved directory button: want %+v, got %+v", before, got)
		}
	}
	for save := 0; save < 2; save++ {
		if err := test.Click("启动时检查新版本开关"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if !s.settingsBusy {
			t.Fatal("switch did not start saving")
		}
		checkBounds()
		finish()
		test.Frame()
		test.Advance(200 * time.Millisecond)
		checkBounds()
		if !test.HasText("保存成功") {
			t.Fatal("missing save feedback")
		}
		if save == 0 {
			file, err := os.Create("/tmp/settings-save-feedback.png")
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, test.Image())
			file.Close()
			if err != nil {
				t.Fatal(err)
			}
			test.Advance(2 * time.Second)
		}
	}
	// A second successful save replaces the existing toast and starts a fresh lifetime.
	test.Advance(2 * time.Second)
	if !test.HasText("保存成功") {
		t.Fatal("replacement toast expired too soon")
	}
	test.Advance(2 * time.Second)
	if test.HasText("保存成功") {
		t.Fatal("success toast did not expire")
	}
	checkBounds()
	s.notifySettings("保存失败：测试错误", true)
	test.Frame()
	test.Advance(7 * time.Second)
	if test.HasText("保存失败：测试错误") {
		t.Fatal("error toast did not expire")
	}
}

func TestSettingsSectionsFadeWithoutSliding(t *testing.T) {
	s := NewShell()
	s.router.Push("/settings")
	test := ui.NewTester(s.view, 1200, 1000)
	test.Advance(300 * time.Millisecond)
	before, ok := test.Find("设置表单")
	if !ok {
		t.Fatal("missing settings form")
	}
	// General has no section path segment; it should animate exactly like its peers.
	for _, path := range []string{"/settings/bilibili", "/settings/download", "/settings", "/settings/ffmpeg", "/settings/about", "/settings/filename", "/settings"} {
		s.router.Push(path)
		test.Frame()
		// Each form is centered within the right pane; its width may differ by section.
		for frame := 0; frame < 12; frame++ {
			test.Advance(16 * time.Millisecond)
			form, ok := test.Find("设置表单")
			if !ok || form.X+form.W/2 != before.X+before.W/2 {
				t.Fatalf("settings form slid during %s: %+v", path, form)
			}
		}
	}
}

func TestFFmpegInstallGuideLinkAndCopy(t *testing.T) {
	s := NewShell()
	s.router.Push("/settings/ffmpeg")
	test := ui.NewTester(s.view, 1200, 1500)
	test.Advance(300 * time.Millisecond)
	if err := test.Click("ffmpeg.org ↗"); err != nil {
		t.Fatal(err)
	}
	if urls := test.OpenedURLs(); len(urls) != 1 || urls[0] != "https://ffmpeg.org/download.html" {
		t.Fatalf("FFmpeg inline link opened unexpected URL: %v", urls)
	}
	for _, command := range []string{"brew install ffmpeg", "sudo apt install ffmpeg", "sudo dnf install ffmpeg"} {
		if err := test.Click("复制代码"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if got := test.Clipboard(); got != command {
			t.Fatalf("copied %q, want %q", got, command)
		}
	}
	test.Advance(3 * time.Second)
	if !test.HasText("brew install ffmpeg") {
		t.Fatal("copy feedback removed the code block")
	}
	file, err := os.Create("/tmp/settings-ffmpeg-guide.png")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, test.Image()); err != nil {
		t.Fatal(err)
	}
}
