package bilibili

import (
	"fmt"
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
)

type concurrentTestAPI struct {
	*authorDownloadAPI
	mu sync.Mutex
}

func (a *concurrentTestAPI) ResolvePlayUrl(req biliapi.PlayUrlRequest, n int) biliapi.PlayUrlResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.authorDownloadAPI.ResolvePlayUrl(req, n)
}

type blockingConcurrentBackend struct {
	started      chan string
	release      chan struct{}
	active, peak atomic.Int64
}

func (*blockingConcurrentBackend) DownloadConcurrency() int { return 3 }
func (d *blockingConcurrentBackend) DownloadVideosByDash(tasks []bilidownload.Task) (bilidownload.BatchResult, error) {
	n := d.active.Add(1)
	defer d.active.Add(-1)
	for old := d.peak.Load(); n > old; old = d.peak.Load() {
		if d.peak.CompareAndSwap(old, n) {
			break
		}
	}
	d.started <- tasks[0].Bvid
	<-d.release
	return bilidownload.BatchResult{Success: 1, Results: []bilidownload.Result{{Bvid: tasks[0].Bvid}}}, nil
}
func parallelRuntime(t *testing.T, s *State) (*blockingConcurrentBackend, chan func()) {
	t.Helper()
	backend := &blockingConcurrentBackend{started: make(chan string, 10), release: make(chan struct{}, 10)}
	done := make(chan func(), 10)
	rt := s.runtime
	rt.Background = func(work func() error, finish func(error)) {
		go func() { err := work(); done <- func() { finish(err) } }()
	}
	s.Configure(s.api, rt)
	s.ConfigureDownloads(backend)
	t.Cleanup(func() { close(backend.release) })
	return backend, done
}
func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for concurrent worker")
		var zero T
		return zero
	}
}

func TestBatchRunsThreeWorkersAndRefillsOneSlot(t *testing.T) {
	s, base, _, r := prepareAuthorDownload(t)
	s.Configure(&concurrentTestAPI{authorDownloadAPI: base}, s.runtime)
	backend, done := parallelRuntime(t, s)
	s.enqueueBatch(batchSource{Key: "parallel", Title: "并发测试", Load: func(int) (batchPage, error) {
		page := batchPage{Total: 6, Received: 6}
		for i := 0; i < 6; i++ {
			page.Items = append(page.Items, bilidownload.Task{Bvid: fmt.Sprintf("BVparallel%d", i), Title: fmt.Sprint(i)})
		}
		return page, nil
	}})
	r.finish(t, 0)
	first := []string{receive(t, backend.started), receive(t, backend.started), receive(t, backend.started)}
	b := s.batches[0]
	if b.InFlight != 3 || backend.peak.Load() != 3 {
		t.Fatal("configured concurrency of 3 did not launch three workers")
	}
	s.UpdateDownloadProgress(bilidownload.Progress{Bvid: first[0], Phase: "sleep", SleepRemaining: 5, SleepTotal: 10})
	sleeping := 0
	for _, item := range b.Items {
		if item.Progress.Phase == "sleep" {
			sleeping++
		}
	}
	if sleeping != 1 || b.InFlight != 3 {
		t.Fatal("one worker's sleep overwrote other workers' progress")
	}
	path := "/bilibili/download"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() {}) })
	}, 1200, 750)
	view.Frame()
	if !view.HasText("运行中 3 · 休眠 1") {
		t.Fatal("download page does not show concurrent workers separately")
	}
	saveImage(t, view, "/tmp/bilibili-concurrent-downloads.png")
	select {
	case <-backend.started:
		t.Fatal("concurrency limit exceeded")
	default:
	}
	backend.release <- struct{}{}
	receive(t, done)()
	receive(t, backend.started)
	if b.InFlight != 3 || backend.peak.Load() != 3 {
		t.Fatal("free slot did not refill independently")
	}
	s.stopBatch(b)
	for range 3 {
		backend.release <- struct{}{}
	}
	for range 3 {
		receive(t, done)()
	}
	if b.Active || b.Status != "stopped" || b.Done != 4 || backend.peak.Load() > 3 {
		t.Fatal("stop should wait for in-flight workers without starting remaining videos")
	}
}

func TestSeparateManualClicksCanUseThreeSlots(t *testing.T) {
	s, _, _, r := stageSignedManual(t)
	for _, key := range []string{"BVsecond", "BVthird"} {
		s.stageManual(manualVideo{Key: key, Bvid: key, Title: key})
	}
	s.ensureQueueResolves()
	r.flush(t)
	backend, done := parallelRuntime(t, s)
	for _, key := range []string{"BVsingle", "BVsecond", "BVthird"} {
		s.startSingleQueuedDownload(key)
	}
	for range 3 {
		receive(t, backend.started)
	}
	if s.activeManualDownloads.inFlight != 3 || backend.peak.Load() != 3 {
		t.Fatal("single clicks were serialized instead of sharing available slots")
	}
	for range 3 {
		backend.release <- struct{}{}
	}
	for range 3 {
		receive(t, done)()
	}
	if s.queueDownloading || s.queueDownloadDone != 3 || len(s.queueRequested) != 0 || len(s.stagedKeys) != 0 {
		t.Fatal("concurrent manual completions did not clean up the queue")
	}
}

func TestStreamCacheUsesEarliestDeadlineAndBoundedLifetime(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	short := now.Add(2 * time.Minute)
	until := streamValidity(now, "https://example.com/video?deadline="+fmt.Sprint(short.Unix()))
	if !until.Equal(short.Add(-30 * time.Second)) {
		t.Fatal("signed URL expiry was ignored")
	}
	long := now.Add(time.Hour)
	if !streamValidity(now, "https://example.com/video?deadline="+fmt.Sprint(long.Unix())).Equal(long.Add(-30 * time.Second)) {
		t.Fatal("known expiry should not cause a redundant fallback refresh")
	}
	if !streamValidity(now, "https://example.com/video").Equal(now.Add(streamCacheTTL)) {
		t.Fatal("URL without expiry must have bounded cache lifetime")
	}
	if !streamValidity(time.Time{}, "https://example.com/video").IsZero() {
		t.Fatal("missing resolution time must not be considered fresh")
	}
}
