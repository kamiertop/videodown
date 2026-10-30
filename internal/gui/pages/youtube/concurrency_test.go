package youtube

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/storage"
	yt "github.com/kamiertop/videodown/youtube"
)

type concurrentYouTubeAPI struct {
	queueAPI
	service      *yt.Service
	started      chan string
	release      chan struct{}
	active, peak atomic.Int64
}

func (a *concurrentYouTubeAPI) DownloadConcurrency() int { return a.service.DownloadConcurrency() }
func (a *concurrentYouTubeAPI) Download(ctx context.Context, v yt.Video, _ int, notify func(yt.Progress)) (string, error) {
	n := a.active.Add(1)
	defer a.active.Add(-1)
	for old := a.peak.Load(); n > old; old = a.peak.Load() {
		if a.peak.CompareAndSwap(old, n) {
			break
		}
	}
	a.started <- v.ID
	notify(yt.Progress{Phase: "下载视频", Speed: 1 << 20, Percent: 30})
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-a.release:
		return "/download/" + v.ID + ".mp4", nil
	}
}

func receiveQueueEvent[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case event := <-ch:
		return event
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for download worker")
		var zero T
		return zero
	}
}

func TestQueueUsesConfiguredConcurrencyAcrossGroups(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.UpdateConfig(func(c *storage.Config) error { c.Download.Concurrency = 3; return nil }); err != nil {
		t.Fatal(err)
	}
	api := &concurrentYouTubeAPI{service: yt.New(store), started: make(chan string, 10), release: make(chan struct{}, 10)}
	events := make(chan func(), 100)
	s := NewState()
	s.Configure(api, guiruntime.Runtime{Background: func(work func() error, done func(error)) {
		go func() { err := work(); events <- func() { done(err) } }()
	}}, func(fn func()) { events <- fn })
	until := func(predicate func() bool) {
		t.Helper()
		for !predicate() {
			receiveQueueEvent(t, events)()
		}
	}
	defer func() { s.Stop(); until(func() bool { return len(s.active) == 0 }) }()
	for i := range 5 {
		key := "course-a"
		if i >= 3 {
			key = "course-b"
		}
		s.enqueueGrouped(queueVideo(fmt.Sprintf("video%06d", i)), 18, false, key, key, true)
	}
	if len(s.active) != 0 {
		t.Fatal("staging started downloads")
	}
	s.startAll()
	for range 3 {
		receiveQueueEvent(t, api.started)
	}
	if len(s.active) != 3 || s.waitingCount() != 2 || s.PendingDownloadCount() != 5 || api.peak.Load() != 3 {
		t.Fatal("configured concurrency was not honored")
	}
	until(func() bool {
		return s.queue[0].Progress.Speed > 0 && s.queue[1].Progress.Speed > 0 && s.queue[2].Progress.Speed > 0
	})
	for _, group := range s.queueGroups {
		group.Expanded = false
	}
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) { core.Use(c, core.Settings{}); s.Form(c, "/youtube/download", scroll) }, 1200, 760)
	test.Frame()
	rate := transferStatus(s.queue[0].Progress)
	if test.HasText(rate) || !test.HasText("待下载 0 · 下载中 3 · 等待 2") {
		t.Fatal("collapsed group repeats speed or queued count is reported as running")
	}
	for _, group := range s.queueGroups {
		group.Expanded = true
	}
	test.Frame()
	if _, ok := test.Find("任务进度：1"); !ok || s.queue[0].Progress.Speed == 0 {
		t.Fatal("individual video lost its transfer rate")
	}
	saveQueueScreenshot(t, test, "/tmp/youtube-concurrent-downloads.png")
	s.cancelItem(s.queue[0])
	until(func() bool { return s.queue[0].Status == "canceled" })
	if id := receiveQueueEvent(t, api.started); id != s.queue[3].Video.ID {
		t.Fatal("completion did not refill one slot")
	}
	s.queuePaused = true
	api.release <- struct{}{}
	until(func() bool { return len(s.active) == 2 })
	select {
	case <-api.started:
		t.Fatal("paused queue launched a task")
	default:
	}
	s.queuePaused = false
	s.pumpQueue()
	if id := receiveQueueEvent(t, api.started); id != s.queue[4].Video.ID {
		t.Fatal("resume did not fill remaining slot")
	}
	s.removeGroup(s.queueGroups[1])
	until(func() bool { return len(s.active) == 1 })
	if api.peak.Load() > 3 || s.PendingDownloadCount() != 1 {
		t.Fatal("group removal canceled another group or exceeded limit")
	}
}

type settingsQueueAPI struct {
	queueAPI
	service *yt.Service
}

func (a *settingsQueueAPI) DownloadConcurrency() int { return a.service.DownloadConcurrency() }

func TestQueueRereadsLimitWithoutCancelingActiveDownloads(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	setLimit := func(limit int) {
		t.Helper()
		if err := store.UpdateConfig(func(c *storage.Config) error { c.Download.Concurrency = limit; return nil }); err != nil {
			t.Fatal(err)
		}
	}
	setLimit(3)
	api := &settingsQueueAPI{service: yt.New(store)}
	var jobs []queueJob
	s := NewState()
	s.Configure(api, guiruntime.Runtime{Background: func(work func() error, done func(error)) { jobs = append(jobs, queueJob{work, done}) }}, func(fn func()) { fn() })
	defer s.Stop()
	for i := range 6 {
		s.enqueue(queueVideo(fmt.Sprintf("video%06d", i)), 18, false)
	}
	s.startAll()
	if len(s.active) != 3 || len(jobs) != 3 {
		t.Fatal("initial limit ignored")
	}
	setLimit(1)
	for _, remaining := range []int{2, 1} {
		takeQueueJob(t, &jobs).run()
		if len(s.active) != remaining || len(jobs) != remaining {
			t.Fatal("lower limit launched more work or canceled an active download")
		}
	}
	takeQueueJob(t, &jobs).run()
	if len(s.active) != 1 || len(jobs) != 1 {
		t.Fatal("lower limit did not retain one running slot")
	}
	setLimit(3)
	s.pumpQueue()
	if len(s.active) != 3 || len(jobs) != 3 {
		t.Fatal("raised limit did not refill slots")
	}
}

func TestHistoryMergeKeepsOutOfOrderCompletions(t *testing.T) {
	s := NewState()
	a := yt.History{Path: "/a.mp4", Downloaded: "2026-10-30T01:00:00Z"}
	b := yt.History{Path: "/b.mp4", Downloaded: "2026-10-30T01:01:00Z"}
	s.mergeHistory([]yt.History{b, a})
	s.mergeHistory([]yt.History{a})
	if len(s.history) != 2 || s.history[0].Path != b.Path {
		t.Fatal("older completion hid a finished file")
	}
}
