package bilibili

import (
	"fmt"
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"slices"
	"strings"
	"testing"
	"time"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
)

type signedQueueAPI struct {
	*authorDownloadAPI
	calls     map[string]int
	removeLow bool
}

func (a *signedQueueAPI) ResolvePlayUrl(req biliapi.PlayUrlRequest, count int) biliapi.PlayUrlResult {
	result := a.authorDownloadAPI.ResolvePlayUrl(req, count)
	a.calls[req.Bvid]++
	revision := a.calls[req.Bvid]
	for i := range result.PlayURL.Dash.Video {
		stream := &result.PlayURL.Dash.Video[i]
		stream.BaseUrl = fmt.Sprintf("https://example.com/video/%d?revision=%d", stream.Id, revision)
		stream.BaseUrl1 = ""
	}
	for i := range result.PlayURL.Dash.Audio {
		stream := &result.PlayURL.Dash.Audio[i]
		stream.BaseUrl = fmt.Sprintf("https://example.com/audio/%d?revision=%d", stream.Id, revision)
	}
	if a.removeLow && revision > 1 {
		result.PlayURL.Dash.Video = result.PlayURL.Dash.Video[1:]
	}
	return result
}

func stageSignedManual(t *testing.T) (*State, *signedQueueAPI, *authorDownloadBackend, *fakeRuntime) {
	t.Helper()
	s, base, backend, r := prepareAuthorDownload(t)
	api := &signedQueueAPI{authorDownloadAPI: base, calls: map[string]int{}}
	s.Configure(api, s.runtime)
	s.stageManual(manualVideo{Key: "BVsingle", Bvid: "BVsingle", Title: "单独下载的视频"})
	s.ensureQueueResolves()
	r.flush(t)
	s.queueResolves["BVsingle"].videoID = 64
	s.queueResolves["BVsingle"].audioID = 30216
	return s, api, backend, r
}

func TestManualRequestWaitsInFIFOAndReusesValidChosenStreams(t *testing.T) {
	s, api, backend, r := stageSignedManual(t)
	s.enqueueBatch(testBatchSource("a", 1))
	s.enqueueBatch(testBatchSource("b", 1))
	s.startSingleQueuedDownload("BVsingle")
	s.startSingleQueuedDownload("BVsingle") // repeated clicks must not queue duplicate work
	if !s.queueRequested["BVsingle"] || len(s.pendingManualDownloads) != 1 || s.queueDownloading || api.calls["BVsingle"] != 1 {
		t.Fatal("single video was not queued behind the existing batches")
	}
	s.enqueueBatch(testBatchSource("c", 1))
	r.flush(t)
	got := []string{}
	for _, task := range backend.tasks {
		got = append(got, task.Bvid)
	}
	if !slices.Equal(got, []string{"BVa1", "BVb1", "BVsingle", "BVc1"}) {
		t.Fatalf("FIFO ordering: %v", got)
	}
	task := backend.tasks[2]
	if task.VideoURL != "https://example.com/video/64?revision=1" || task.AudioURL != "https://example.com/audio/30216?revision=1" || api.calls["BVsingle"] != 1 {
		t.Fatalf("execution unnecessarily resolved again or changed the user's selection: %+v", task)
	}
	if len(s.queueRequested) != 0 || len(s.pendingManualDownloads) != 0 || s.stagedKeys["BVsingle"] {
		t.Fatal("completed manual request was not cleaned up")
	}
}

func TestRemovingWaitingVideoCancelsItsDownload(t *testing.T) {
	s, api, backend, r := stageSignedManual(t)
	s.enqueueBatch(testBatchSource("a", 1))
	s.startSingleQueuedDownload("BVsingle")
	s.removeStaged("BVsingle")
	r.flush(t)
	if len(backend.tasks) != 1 || backend.tasks[0].Bvid != "BVa1" || api.calls["BVsingle"] != 1 || len(s.pendingManualDownloads) != 0 {
		t.Fatal("removed video still resolved or downloaded")
	}
}

func TestUnavailableSelectedQualityRequiresExplicitReselection(t *testing.T) {
	s, api, backend, r := stageSignedManual(t)
	api.removeLow = true
	s.queueResolves["BVsingle"].resolvedAt = time.Now().Add(-2 * streamCacheTTL)
	s.startSingleQueuedDownload("BVsingle")
	r.flush(t)
	if len(backend.tasks) != 0 || !s.stagedKeys["BVsingle"] || s.queueRequested["BVsingle"] || !strings.Contains(s.queueDownloadError, "所选画质已不可用") || s.queueResolves["BVsingle"].data.Error == "" {
		t.Fatal("missing selected quality silently fell back or could not be retried")
	}
}

func TestOnlySuccessfulBatchesAutoDismissAfterCompletion(t *testing.T) {
	s, _, backend, r := prepareAuthorDownload(t)
	rt := s.runtime
	var timers []struct {
		delay time.Duration
		run   func()
	}
	rt.After = func(delay time.Duration, run func()) func() {
		timers = append(timers, struct {
			delay time.Duration
			run   func()
		}{delay, run})
		return func() {}
	}
	s.Configure(s.api, rt)
	s.enqueueBatch(testBatchSource("success", 1))
	r.flush(t)
	if len(s.batches) != 1 || len(timers) != 1 || timers[0].delay != 3*time.Second {
		t.Fatal("completed batch did not get a brief completion display")
	}
	timers[0].run()
	if len(s.batches) != 0 {
		t.Fatal("successful batch remained visible")
	}
	backend.fail = "BVfailed1"
	s.enqueueBatch(testBatchSource("failed", 1))
	r.flush(t)
	failed := s.batches[0]
	if failed.Failed != 1 || len(timers) != 1 {
		t.Fatal("failed batch was scheduled for dismissal")
	}
	backend.fail = ""
	s.resumeBatch(failed)
	r.flush(t)
	if len(timers) != 2 || failed.Failed != 0 {
		t.Fatal("successful retry did not schedule dismissal")
	}
	timers[1].run()
	if len(s.batches) != 0 {
		t.Fatal("retried successful batch did not close")
	}
	s.enqueueBatch(testBatchSource("stopped", 1))
	s.stopBatch(s.batches[0])
	r.flush(t)
	if len(s.batches) != 1 || s.batches[0].Status != "stopped" || len(timers) != 2 {
		t.Fatal("stopped batch must remain resumable")
	}
}

func TestSingleDownloadButtonCanQueueDuringBatch(t *testing.T) {
	s, api, backend, r := stageSignedManual(t)
	s.enqueueBatch(testBatchSource("a", 1))
	path := "/bilibili/download"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(next string) { path = next }, func() { s.Form(c, path, func(next string) { path = next }, func() {}) })
	}, 1200, 750)
	view.Frame()
	if err := view.Click("下载视频"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if !s.queueRequested["BVsingle"] || !view.HasText("已排队") || api.calls["BVsingle"] != 1 || len(backend.tasks) != 0 {
		t.Fatal("single download button is disabled or resolves too early during a batch")
	}
	r.flush(t)
	view.Frame()
	if len(backend.tasks) != 2 || backend.tasks[0].Bvid != "BVa1" || backend.tasks[1].Bvid != "BVsingle" {
		t.Fatal("queued button action did not download after the batch")
	}
}

func TestExpiredQueuedStreamRefreshesAtExecution(t *testing.T) {
	s, api, backend, r := stageSignedManual(t)
	s.queueResolves["BVsingle"].resolvedAt = time.Now().Add(-2 * streamCacheTTL)
	s.enqueueBatch(testBatchSource("a", 1))
	s.startSingleQueuedDownload("BVsingle")
	if api.calls["BVsingle"] != 1 {
		t.Fatal("queued URL refreshed too early")
	}
	r.flush(t)
	task := backend.tasks[1]
	if api.calls["BVsingle"] != 2 || task.VideoURL != "https://example.com/video/64?revision=2" || task.AudioURL != "https://example.com/audio/30216?revision=2" {
		t.Fatalf("expired stream not refreshed with chosen IDs: %+v", task)
	}
}
