package bilibili

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
)

type authorDownloadAPI struct {
	*fakeAPI
	resolved []string
}

func (a *authorDownloadAPI) ResolvePlayUrl(req biliapi.PlayUrlRequest, _ int) biliapi.PlayUrlResult {
	a.resolved = append(a.resolved, req.Bvid)
	return biliapi.PlayUrlResult{
		Bvid: req.Bvid, Cid: 42,
		Detail: &model.VideoDetailConciseData{View: model.VideoDetailView{Bvid: req.Bvid, Cid: 42, Title: req.Bvid}},
		PlayURL: &model.VideoURLData{Dash: model.Dash{
			Video: []model.VideoItem{{Id: 64, BaseUrl: "https://example.com/low", Bandwidth: 100}, {Id: 80, BaseUrl1: "https://example.com/high", Bandwidth: 200}},
			Audio: []model.AudioItem{{Id: 30216, BaseUrl: "https://example.com/audio-low", Bandwidth: 64000}, {Id: 30280, BaseUrl: "https://example.com/audio", Bandwidth: 128000}},
		}, SupportFormats: []model.SupportFormat{{Quality: 64, DisplayDesc: "720P"}, {Quality: 80, DisplayDesc: "1080P"}}},
	}
}

type authorDownloadBackend struct {
	tasks []bilidownload.Task
	fail  string
}

func (d *authorDownloadBackend) DownloadVideosByDash(tasks []bilidownload.Task) (bilidownload.BatchResult, error) {
	d.tasks = append(d.tasks, tasks...)
	if len(tasks) != 1 {
		return bilidownload.BatchResult{}, errors.New("expected one task per call")
	}
	if tasks[0].Bvid == d.fail {
		return bilidownload.BatchResult{Failed: 1, Results: []bilidownload.Result{{Bvid: tasks[0].Bvid, Error: "模拟下载失败"}}}, nil
	}
	return bilidownload.BatchResult{Success: 1, Results: []bilidownload.Result{{Bvid: tasks[0].Bvid}}}, nil
}

func prepareAuthorDownload(t *testing.T) (*State, *authorDownloadAPI, *authorDownloadBackend, *fakeRuntime) {
	t.Helper()
	s, base, runtime := setup(t, true)
	api := &authorDownloadAPI{fakeAPI: base}
	background := runtime.runtime()
	background.Background = background.Run
	s.Configure(api, background)
	backend := &authorDownloadBackend{}
	s.ConfigureDownloads(backend)
	s.findAuthor("12", nil)
	runtime.flush(t)
	return s, api, backend, runtime
}

func TestAuthorOneClickDownloadsEveryPage(t *testing.T) {
	s, api, backend, runtime := prepareAuthorDownload(t)
	path := "/bilibili/up/12"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() {
			s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
		})
	}, 760, 700)
	view.Frame()
	saveImage(t, view, "/tmp/bilibili-author-download-760.png")
	controls, ok := view.Find("UP 主内容操作")
	if !ok {
		t.Fatal("missing author operation card")
	}
	profile, ok := view.Find("UP 主信息")
	if !ok {
		t.Fatal("missing author profile card")
	}
	download, ok := view.Find("一键下载全部")
	if !ok || download.Y < controls.Y || download.Y+download.H > controls.Y+controls.H || download.Y < profile.Y+profile.H {
		t.Fatalf("download button must be inside operation card, below profile: button=%+v controls=%+v profile=%+v", download, controls, profile)
	}
	if err := view.Click("一键下载全部"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	runtime.flush(t)
	view.Frame()
	if s.batches[0].Active || s.batches[0].Done != 2 || s.batches[0].Failed != 0 || len(backend.tasks) != 2 {
		t.Fatalf("download state: active=%v done=%d failed=%d tasks=%d", s.batches[0].Active, s.batches[0].Done, s.batches[0].Failed, len(backend.tasks))
	}
	for i, task := range backend.tasks {
		want := fmt.Sprintf("BVupload12-%d", i+1)
		if task.Bvid != want || task.Cid != 42 || task.VideoURL != "https://example.com/high" || task.AudioURL != "https://example.com/audio" || task.UpperName != "测试 UP 主" || api.resolved[i] != want {
			t.Fatalf("task %d: %+v resolved=%v", i, task, api.resolved)
		}
	}
	if path != "/bilibili/download" || !view.HasText("已完成 2") {
		t.Fatal("batch did not navigate to download page or show completion")
	}
}

func TestAuthorSelectedCardsOpenDownloadPage(t *testing.T) {
	s, api, backend, runtime := prepareAuthorDownload(t)
	path := "/bilibili/up/12"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() {
			s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
		})
	}, 760, 700)
	view.Frame()
	if view.HasText("添加全部") || !view.HasText("下载已选 (0)") {
		t.Fatal("author actions should offer selected or all downloads only")
	}
	if err := view.Click("选择视频：投稿第 1 页"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if err := view.Click("下载已选 (1)"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if path != "/bilibili/download" || len(s.stagedKeys) != 1 || len(backend.tasks) != 0 {
		t.Fatalf("selection must enter the download queue without starting: path=%s staged=%d tasks=%d", path, len(s.stagedKeys), len(backend.tasks))
	}
	runtime.flush(t)
	view.Frame()
	if !view.HasText("视频画质") || !view.HasText("音频音质") || len(api.resolved) != 1 || api.resolved[0] != "BVupload12-1" || len(backend.tasks) != 0 {
		t.Fatalf("download page must resolve stream choices before starting: resolved=%v tasks=%d", api.resolved, len(backend.tasks))
	}
	saveImage(t, view, "/tmp/bilibili-queued-selected-760.png")
	if err := view.Click("1080P"); err != nil {
		t.Fatal(err)
	}
	if err := view.Click("720P"); err != nil {
		t.Fatal(err)
	}
	if err := view.Click("128 kbps"); err != nil {
		t.Fatal(err)
	}
	if err := view.Click("64 kbps"); err != nil {
		t.Fatal(err)
	}
	if err := view.Click("开始下载"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if len(backend.tasks) != 1 || backend.tasks[0].Bvid != "BVupload12-1" || backend.tasks[0].VideoURL != "https://example.com/low" || backend.tasks[0].AudioURL != "https://example.com/audio-low" || s.queueDownloadDone != 1 || len(s.stagedKeys) != 0 {
		t.Fatalf("queued selected video did not download: tasks=%+v done=%d staged=%d", backend.tasks, s.queueDownloadDone, len(s.stagedKeys))
	}
}

func TestAuthorSelectAllOnlyMarksLoadedVideos(t *testing.T) {
	s, _, backend, runtime := prepareAuthorDownload(t)
	s.uploads.Data.List.Vlist = append(s.uploads.Data.List.Vlist, model.VListItem{Bvid: "BVextra", Title: "已加载的第二个视频"})
	s.uploads.Data.Page.Count = 3
	path := "/bilibili/up/12"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() {
			s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
		})
	}, 760, 700)
	view.Frame()
	if err := view.Click("全选"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if !*s.selection("BVupload12-1") || !*s.selection("BVextra") || !view.HasText("下载已选 (2)") || !view.HasText("取消全选") || len(backend.tasks) != 0 || len(s.stagedKeys) != 0 {
		t.Fatal("select all must mark only loaded cards without starting a download")
	}
	if err := view.Click("取消全选"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if *s.selection("BVupload12-1") || *s.selection("BVextra") || !view.HasText("下载已选 (0)") {
		t.Fatal("cancel select all did not clear the loaded selection")
	}
	if err := view.Click("全选"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if err := view.Click("下载已选 (2)"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if path != "/bilibili/download" || len(s.stagedKeys) != 2 || len(backend.tasks) != 0 {
		t.Fatalf("selected cards should be queued, not downloaded: path=%s staged=%d tasks=%d", path, len(s.stagedKeys), len(backend.tasks))
	}
	runtime.flush(t)
	view.Frame()
	if err := view.Click("开始下载"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	if len(backend.tasks) != 2 || s.queueDownloadTotal != 2 || s.queueDownloadDone != 2 {
		t.Fatalf("selected download should contain only loaded videos: tasks=%d total=%d done=%d", len(backend.tasks), s.queueDownloadTotal, s.queueDownloadDone)
	}
}

func TestAuthorSelectedDownloadUsesCurrentCollection(t *testing.T) {
	for _, kind := range []string{"season", "series"} {
		t.Run(kind, func(t *testing.T) {
			s, _, backend, runtime := prepareAuthorDownload(t)
			s.authorTab = "lists"
			want := "BVseason"
			if kind == "season" {
				s.loadSeason(3, 1, false)
			} else {
				s.loadSeries(4, 1, false)
				want = "BVseries"
			}
			runtime.flush(t)
			*s.selection(want) = true
			if s.stageSelectedAuthorVideos() != 1 || len(backend.tasks) != 0 || !s.stagedKeys[want] {
				t.Fatalf("selected %s should only be staged: tasks=%+v staged=%v", kind, backend.tasks, s.stagedKeys)
			}
		})
	}
}

func TestQueuedDownloadWaitsForAllStreamChoices(t *testing.T) {
	s, _, backend, runtime := prepareAuthorDownload(t)
	s.stageUpload(model.VListItem{Bvid: "BVqueueA", Title: "队列 A"})
	s.stageUpload(model.VListItem{Bvid: "BVqueueB", Title: "队列 B"})
	path := "/bilibili/download"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() {
			s.Form(c, path, func(p string) { path = p }, func() {})
		})
	}, 760, 700)
	view.Frame()
	if len(runtime.jobs) != 2 {
		t.Fatalf("expected two stream resolves, got %d", len(runtime.jobs))
	}
	runtime.finish(t, 0)
	view.Frame()
	_ = view.Click("开始下载")
	view.Frame()
	if len(backend.tasks) != 0 || s.queueDownloading {
		t.Fatal("queue started before every video's streams were ready")
	}
	runtime.finish(t, 0)
	view.Frame()
	if err := view.Click("开始下载"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	if len(backend.tasks) != 2 || len(s.stagedKeys) != 0 {
		t.Fatalf("queue did not download both ready videos: tasks=%d staged=%d", len(backend.tasks), len(s.stagedKeys))
	}
}

func TestQueuedDownloadKeepsFailedVideoForRetry(t *testing.T) {
	s, _, backend, runtime := prepareAuthorDownload(t)
	backend.fail = "BVretry"
	s.stageUpload(model.VListItem{Bvid: "BVretry", Title: "重试视频"})
	path := "/bilibili/download"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() {
			s.Form(c, path, func(p string) { path = p }, func() {})
		})
	}, 760, 700)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	if err := view.Click("开始下载"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if len(s.stagedKeys) != 1 || s.queueDownloadFailed != 1 || s.queueDownloadError == "" {
		t.Fatal("failed video must remain available for retry")
	}
	backend.fail = ""
	if err := view.Click("开始下载"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	if len(s.stagedKeys) != 0 || s.queueDownloadFailed != 0 || len(backend.tasks) != 2 {
		t.Fatalf("retry did not clear successful video: staged=%d failed=%d tasks=%d", len(s.stagedKeys), s.queueDownloadFailed, len(backend.tasks))
	}
}

func TestAuthorOneClickCanStopAfterCurrentTask(t *testing.T) {
	s, _, backend, runtime := prepareAuthorDownload(t)
	path := "/bilibili/up/12"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() {
			s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
		})
	}, 760, 700)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	s.downloadAllAuthor(func(p string) { path = p })
	runtime.finish(t, 0) // fetch all pages, then queue the first download
	if !s.batches[0].Active || len(runtime.jobs) != 1 {
		t.Fatal("first download was not queued")
	}
	view.Frame()
	if err := view.Click("停止批量：1"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	runtime.flush(t)
	if s.batches[0].Active || len(backend.tasks) != 1 || s.batches[0].Done != 1 {
		t.Fatalf("stop did not prevent the next task: tasks=%d done=%d", len(backend.tasks), s.batches[0].Done)
	}
}

func TestAuthorOneClickReportsIndividualFailure(t *testing.T) {
	s, _, backend, runtime := prepareAuthorDownload(t)
	backend.fail = "BVupload12-1"
	s.downloadAllAuthor()
	runtime.flush(t)
	if s.batches[0].Done != 1 || s.batches[0].Failed != 1 || len(backend.tasks) != 2 || s.batches[0].Error == "" {
		t.Fatalf("single failure stopped the batch: done=%d failed=%d tasks=%d error=%q", s.batches[0].Done, s.batches[0].Failed, len(backend.tasks), s.batches[0].Error)
	}
}

func TestAuthorOneClickCanCancelPageCollection(t *testing.T) {
	s, _, backend, runtime := prepareAuthorDownload(t)
	s.downloadAllAuthor()
	s.stopBatch(s.batches[0])
	runtime.flush(t)
	if s.batches[0].Active || len(backend.tasks) != 0 || s.batches[0].Status != "stopped" {
		t.Fatal("canceled source still downloaded")
	}
}

func TestAuthorOneClickUsesSelectedSeasonOrSeries(t *testing.T) {
	for _, kind := range []string{"season", "series"} {
		t.Run(kind, func(t *testing.T) {
			s, _, backend, runtime := prepareAuthorDownload(t)
			s.authorTab = "lists"
			want := "BVseason"
			if kind == "season" {
				s.selectedSeason = 3
			} else {
				s.selectedSeries = 4
				want = "BVseries"
			}
			s.downloadAllAuthor()
			runtime.flush(t)
			if len(backend.tasks) != 1 || backend.tasks[0].Bvid != want || s.batches[0].Done != 1 {
				t.Fatalf("wrong %s source: tasks=%+v done=%d", kind, backend.tasks, s.batches[0].Done)
			}
		})
	}
}
