package bilibili

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func testBatchSource(key string, count int) batchSource {
	return batchSource{Key: key, Title: "批量来源 " + key, Total: count, Load: func(page int) (batchPage, error) {
		return batchPage{Items: []bilidownload.Task{{Bvid: fmt.Sprintf("BV%s%d", key, page), Title: fmt.Sprintf("%s 的视频 %d", key, page)}}, Total: count, Received: 1, HasMore: page < count}, nil
	}}
}

func TestBatchesQueueSeparatelyAndDoNotConsumeManualVideos(t *testing.T) {
	s, _, backend, r := prepareAuthorDownload(t)
	s.stageManual(manualVideo{Key: "BVmanual", Bvid: "BVmanual", Title: "普通选择的视频"})
	path := ""
	navigate := func(next string) { path = next }
	s.enqueueBatch(testBatchSource("a", 2), navigate)
	s.enqueueBatch(testBatchSource("b", 1), navigate)
	s.enqueueBatch(testBatchSource("a", 2), navigate) // active duplicate returns to the existing batch
	if path != "/bilibili/download" || len(s.batches) != 2 || s.batches[1].Status != "queued" || s.PendingDownloadCount() != 2 {
		t.Fatal("batch navigation, isolation or queue order failed")
	}
	r.flush(t)
	got := []string{}
	for _, task := range backend.tasks {
		got = append(got, task.Bvid)
	}
	if !slices.Equal(got, []string{"BVa1", "BVa2", "BVb1"}) || !s.stagedKeys["BVmanual"] {
		t.Fatalf("wrong queue ordering or manual item consumed: %v", got)
	}
	if s.batches[0].Done != 2 || s.batches[1].Done != 1 || s.PendingDownloadCount() != 0 {
		t.Fatal("independent batch counts failed")
	}
}

func TestBatchWaitsForManualAndResumesAfterStop(t *testing.T) {
	s, _, backend, r := prepareAuthorDownload(t)
	s.queueDownloading = true
	s.enqueueBatch(testBatchSource("a", 2))
	b := s.batches[0]
	if len(r.jobs) != 0 || b.Status != "queued" {
		t.Fatal("batch overtook manual download")
	}
	s.queueDownloading = false
	s.pumpDownloads()
	r.finish(t, 0) // page one; starts one download
	s.stopBatch(b)
	r.flush(t)
	if b.Active || b.Done != 1 || len(backend.tasks) != 1 || b.Status != "stopped" {
		t.Fatal("stop did not wait for the current task")
	}
	s.resumeBatch(b)
	r.flush(t)
	if b.Active || b.Done != 2 || len(backend.tasks) != 2 {
		t.Fatal("resume repeated completed video or missed next page")
	}
}

func TestBatchRetryAndPageLoadRecovery(t *testing.T) {
	s, _, backend, r := prepareAuthorDownload(t)
	source := testBatchSource("a", 2)
	loader := source.Load
	failPage := true
	source.Load = func(page int) (batchPage, error) {
		if page == 2 && failPage {
			return batchPage{}, errors.New("加载第二页失败")
		}
		return loader(page)
	}
	backend.fail = "BVa1"
	s.enqueueBatch(source)
	r.flush(t)
	b := s.batches[0]
	if b.Active || b.Failed != 1 || b.Page != 2 || b.Error != "加载第二页失败" {
		t.Fatalf("failed source lost recovery state: %+v", b)
	}
	backend.fail, failPage = "", false
	s.resumeBatch(b)
	r.flush(t)
	if b.Done != 2 || b.Failed != 0 || len(backend.tasks) != 3 || b.Status != "done" {
		t.Fatal("retry did not recover failed item and next page")
	}
}

func TestBatchSourceSnapshotSurvivesNavigation(t *testing.T) {
	s, _, backend, r := prepareAuthorDownload(t)
	s.downloadAllAuthor()
	s.authorInfo.Data.Mid = 99
	s.authorTab = "lists"
	s.selectedSeries = 42
	r.flush(t)
	if len(backend.tasks) != 2 || backend.tasks[0].Bvid != "BVupload12-1" {
		t.Fatalf("batch source changed after leaving source page: %+v", backend.tasks)
	}
}

type incrementalBatchBackend struct{ authorDownloadBackend }

func (*incrementalBatchBackend) SkipDownloaded() bool { return true }
func (*incrementalBatchBackend) DownloadedVideoKeys(keys []bilidownload.DownloadKey) ([]int, error) {
	if keys[0].Bvid == "BVa1" {
		return []int{0}, nil
	}
	return nil, nil
}
func TestBatchSkipsSavedVideosBeforeResolving(t *testing.T) {
	s, api, _, r := prepareAuthorDownload(t)
	backend := &incrementalBatchBackend{}
	s.ConfigureDownloads(backend)
	s.enqueueBatch(testBatchSource("a", 2))
	r.flush(t)
	if s.batches[0].Skipped != 1 || s.batches[0].Done != 1 || !slices.Equal(api.resolved, []string{"BVa2"}) || len(backend.tasks) != 1 {
		t.Fatal("saved videos must be skipped before stream requests")
	}
}

func TestBatchCardsCollapseIndependentlyAndShowSleepAnimation(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, _, _, r := prepareAuthorDownload(t)
		s.enqueueBatch(testBatchSource("a", 2))
		s.enqueueBatch(testBatchSource("b", 1))
		r.finish(t, 0) // first page is visible; download worker remains pending
		s.stageManual(manualVideo{Key: "BVmanual", Bvid: "BVmanual", Title: "普通选择的视频"})
		s.UpdateDownloadProgress(bilidownload.Progress{Bvid: "BVa1", Cid: 42, Phase: "sleep", Percent: 100, SleepRemaining: 4, SleepTotal: 10})
		path := "/bilibili/download"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			style.Apply(c)
			s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() {}) })
		}, width, 750)
		view.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1})
		view.Frame()
		if !view.HasText("休眠中 · 4 秒") || !view.HasText("普通选择的视频") {
			t.Fatal("sleep state or manual video missing")
		}
		first, ok1 := view.Find("批量下载：1")
		second, ok2 := view.Find("批量下载：2")
		if !ok1 || !ok2 || first.Y+first.H > second.Y || first.X+first.W > float32(width) {
			t.Fatal("batch cards overlap or overflow")
		}
		if err := view.Click("展开或折叠批量：1"); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if s.batches[0].Expanded || !s.batches[1].Expanded || view.HasText("a 的视频 1") || !view.HasText("休眠中 · 4 秒") {
			t.Fatal("collapsing lost summary or affected other session")
		}
		if err := view.Click("展开或折叠批量：1"); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if !view.HasText("a 的视频 1") {
			t.Fatal("expanded batch lost its video")
		}
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-batches-%d.png", width))
	}
}
