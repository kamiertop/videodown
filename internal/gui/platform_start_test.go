package gui

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	bm "github.com/kamiertop/videodown/bilibili/model"
	dm "github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/gui/pages/bilibili"
	"github.com/kamiertop/videodown/internal/gui/pages/douyin"
	ypage "github.com/kamiertop/videodown/internal/gui/pages/youtube"
	yt "github.com/kamiertop/videodown/youtube"
)

type lazyBiliAPI struct {
	bilibili.API
	checks int
}

func (a *lazyBiliAPI) IsLoggedIn() bool { a.checks++; return true }
func (a *lazyBiliAPI) RefreshCookie() (bm.CookieRefreshData, error) {
	return bm.CookieRefreshData{}, nil
}
func (a *lazyBiliAPI) MyInfo() (bm.MyInfoProfile, error) { return bm.MyInfoProfile{}, nil }

type lazyDouyinAPI struct {
	douyin.API
	checks int
}

func (a *lazyDouyinAPI) GetCookie() (string, error) { return "fixture", nil }
func (a *lazyDouyinAPI) Profile() (dm.MyInfoResponse, error) {
	a.checks++
	return dm.MyInfoResponse{User: dm.User{Uid: "1"}}, nil
}

type lazyYouTubeAPI struct {
	ypage.API
	reads int
}

func (a *lazyYouTubeAPI) Cookie() string                 { return "" }
func (a *lazyYouTubeAPI) History() ([]yt.History, error) { a.reads++; return nil, nil }
func TestPlatformsInitializeOnlyOnFirstEntry(t *testing.T) {
	s := NewShell()
	bili := &lazyBiliAPI{}
	dy := &lazyDouyinAPI{}
	youtube := &lazyYouTubeAPI{}
	type job struct {
		work func() error
		done func(error)
	}
	var jobs []job
	run := func(work func() error, done func(error)) { jobs = append(jobs, job{work, done}) }
	runtime := guiruntime.Runtime{Run: run, Background: run}
	s.pages.Bilibili.Configure(bili, runtime)
	s.pages.Douyin.Configure(dy, nil, runtime)
	s.pages.YouTube.Configure(youtube, runtime, func(fn func()) { fn() })
	defer s.Wait()
	// Headless tests have no native window; activate at the same point before rendering.
	test := ui.NewTester(func(c *ui.Context) { s.activatePlatform(); s.view(c) }, 1200, 800)
	frame := func() {
		t.Helper()
		test.Frame()
		test.Advance(300 * time.Millisecond)
		for len(jobs) > 0 {
			j := jobs[0]
			jobs = jobs[1:]
			j.done(j.work())
		}
		test.Frame()
	}
	frame()
	if len(jobs) != 0 || bili.checks != 0 || dy.checks != 0 || youtube.reads != 0 {
		t.Fatal("home initialized platforms")
	}
	s.router.Push("/settings")
	frame()
	if bili.checks != 0 || dy.checks != 0 {
		t.Fatal("settings initialized platforms")
	}
	s.router.Push("/youtube/download")
	frame()
	if youtube.reads != 1 || bili.checks != 0 || dy.checks != 0 {
		t.Fatal("YouTube initialized another platform")
	}
	s.router.Push("/bilibili/download")
	frame()
	if bili.checks != 1 || dy.checks != 0 {
		t.Fatal("Bilibili did not initialize independently")
	}
	s.router.Push("/douyin/profile")
	frame()
	if dy.checks != 1 {
		t.Fatal("Douyin did not initialize")
	}
	for _, route := range []string{"/", "/bilibili/profile", "/youtube/history", "/douyin/profile", "/"} {
		s.router.Push(route)
		frame()
	}
	if bili.checks != 1 || dy.checks != 1 || youtube.reads != 1 {
		t.Fatalf("reentry initialized twice: bili=%d douyin=%d youtube=%d", bili.checks, dy.checks, youtube.reads)
	}
}
