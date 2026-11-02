package bilibili

import (
	"fmt"
	"github.com/kamiertop/videodown/internal/gui/style"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
)

type manualDownloadAPI struct {
	*fakeAPI
	detail   model.VideoDetailConciseData
	requests []biliapi.PlayUrlRequest
}

type manualExtraBackend struct {
	authorDownloadBackend
	covers, audios []bilidownload.Task
}

func (d *manualExtraBackend) DownloadCover(_ string, task bilidownload.Task) (string, error) {
	d.covers = append(d.covers, task)
	return "/tmp/cover.png", nil
}
func (d *manualExtraBackend) DownloadAudio(task bilidownload.Task) (string, error) {
	d.audios = append(d.audios, task)
	return "/tmp/audio.m4a", nil
}

func (a *manualDownloadAPI) VideoDetailConciseBvid(string) (model.VideoDetailConciseData, error) {
	return a.detail, nil
}
func (a *manualDownloadAPI) VideoDetailConcise(int) (model.VideoDetailConciseData, error) {
	return a.detail, nil
}
func (a *manualDownloadAPI) ResolvePlayUrl(req biliapi.PlayUrlRequest, _ int) biliapi.PlayUrlResult {
	a.requests = append(a.requests, req)
	cid := req.Cid
	if cid == 0 {
		cid = a.detail.View.Cid
	}
	return biliapi.PlayUrlResult{Bvid: req.Bvid, Cid: cid, Detail: &a.detail, PlayURL: &model.VideoURLData{Dash: model.Dash{
		Video: []model.VideoItem{{Id: 80, BaseUrl: "https://example.com/video", Bandwidth: 200}},
		Audio: []model.AudioItem{{Id: 30280, BaseUrl: "https://example.com/audio", Bandwidth: 128000}},
	}}}
}

func TestParseVideoReference(t *testing.T) {
	for _, tc := range []struct {
		raw, bvid  string
		avid, part int
		valid      bool
	}{
		{"BV1xx411c7mD", "BV1xx411c7mD", 0, 1, true},
		{"https://www.bilibili.com/video/BV1xx411c7mD?p=2", "BV1xx411c7mD", 0, 2, true},
		{"av12345", "", 12345, 1, true},
		{"https://www.bilibili.com/video/av12345?page=3", "", 12345, 3, true},
		{"something else", "", 0, 0, false},
	} {
		bvid, avid, part, err := parseVideoReference(tc.raw)
		if (err == nil) != tc.valid || tc.valid && (bvid != tc.bvid || avid != tc.avid || part != tc.part) {
			t.Fatalf("%s: %q %d %d %v", tc.raw, bvid, avid, part, err)
		}
	}
}

func TestManualMultiPartSelectionReachesDownloader(t *testing.T) {
	s, base, runtime := setup(t, true)
	api := &manualDownloadAPI{fakeAPI: base, detail: model.VideoDetailConciseData{View: model.VideoDetailView{
		Bvid: "BV1xx411c7mD", Cid: 11, Title: "多P测试", Pic: "//example.com/cover", Owner: model.VideoOwner{Name: "测试 UP 主"},
		Pages: []model.VideoPage{{Cid: 11, Page: 1, Part: "第一集", Duration: 60}, {Cid: 22, Page: 2, Part: "第二集", Duration: 90}},
	}}}
	background := runtime.runtime()
	background.Background = background.Run
	s.Configure(api, background)
	backend := &authorDownloadBackend{}
	s.ConfigureDownloads(backend)
	path := "/bilibili/download"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() {}) })
	}, 900, 750)
	s.link = "https://www.bilibili.com/video/BV1xx411c7mD?p=2"
	view.Frame()
	if err := view.Click("解析"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if s.manualGroup == nil || s.manualGroup.Current.Cid != 22 || len(s.stagedKeys) != 0 {
		t.Fatalf("group: %+v staged=%d", s.manualGroup, len(s.stagedKeys))
	}
	p1, _ := view.Find("选择待添加：多P测试 · P1 第一集")
	p2, _ := view.Find("选择待添加：多P测试 · P2 第二集")
	if p1.Y+p1.H > p2.Y {
		t.Fatal("分P选项相互遮挡")
	}
	if err := view.Click("选择待添加：多P测试 · P1 第一集"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	view.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1})
	groupTitle := "分P · 多P测试 · 2 个视频 · 已选 1"
	if err := view.Click(groupTitle); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if len(s.manualGroup.OpenSections) != 0 || view.HasText("第一集") {
		t.Fatal("合集中内容未折叠")
	}
	if err := view.Click(groupTitle); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	p1, _ = view.Find("选择待添加：多P测试 · P1 第一集")
	p2, _ = view.Find("选择待添加：多P测试 · P2 第二集")
	if p1.Y+p1.H > p2.Y || s.manualGroup.Selected["BV1xx411c7mD:11"] {
		t.Fatal("展开后列表或选择状态不正确")
	}
	saveImage(t, view, "/tmp/bilibili-parsed-group.png")
	if err := view.Click("添加已选 (1)"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if len(s.stagedManual) != 1 || s.stagedManual[0].Cid != 22 || len(api.requests) != 1 || api.requests[0].Cid != 22 {
		t.Fatalf("manual=%+v requests=%+v", s.stagedManual, api.requests)
	}
	task, ok := s.queuedTask("BV1xx411c7mD:22")
	if !ok || task.Title != "多P测试 · P2 第二集" || task.Cid != 22 || task.VideoURL == "" || task.AudioURL == "" {
		t.Fatalf("task=%+v ok=%v", task, ok)
	}
	s.UpdateDownloadProgress(bilidownload.Progress{Bvid: task.Bvid, Cid: task.Cid, Phase: "video", Percent: 35})
	view.Frame()
	if !view.HasText("下载视频") || !view.HasText("35%") {
		t.Fatal("download progress missing")
	}
	if err := view.Click("开始下载"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if len(backend.tasks) != 1 || backend.tasks[0].Cid != 22 || len(s.stagedKeys) != 0 || s.queueDownloadDone != 1 {
		t.Fatalf("download=%+v staged=%d done=%d", backend.tasks, len(s.stagedKeys), s.queueDownloadDone)
	}
}

func TestManualSingleVideoStagesImmediately(t *testing.T) {
	s, base, runtime := setup(t, true)
	api := &manualDownloadAPI{fakeAPI: base, detail: model.VideoDetailConciseData{View: model.VideoDetailView{Bvid: "BV1xx411c7mD", Cid: 42, Title: "单视频"}}}
	s.Configure(api, runtime.runtime())
	s.link = "av12345"
	s.parseDownloadInput()
	runtime.flush(t)
	if s.manualGroup != nil || len(s.stagedManual) != 1 || s.stagedManual[0].Bvid != "BV1xx411c7mD" {
		t.Fatalf("single video not staged: %+v", s.stagedManual)
	}
}

func TestManualVideoOffersSeparateVideoAudioAndCoverActions(t *testing.T) {
	s, base, runtime := setup(t, true)
	api := &manualDownloadAPI{fakeAPI: base, detail: model.VideoDetailConciseData{View: model.VideoDetailView{
		Bvid: "BV1xx411c7mD", Cid: 42, Title: "单视频", Pic: "//example.com/cover",
	}}}
	background := runtime.runtime()
	background.Background = background.Run
	s.Configure(api, background)
	backend := &manualExtraBackend{}
	s.ConfigureDownloads(backend)
	path := "/bilibili/download"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() {}) })
	}, 1200, 880)
	s.link = "BV1xx411c7mD"
	view.Frame()
	if err := view.Click("解析"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	saveImage(t, view, "/tmp/bilibili-manual-download-actions.png")
	coverBounds, _ := view.Find("保存封面")
	audioBounds, _ := view.Find("下载音频")
	if coverBounds.X+coverBounds.W > audioBounds.X {
		t.Fatal("下载操作按钮相互遮挡")
	}
	if err := view.Click("保存封面"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if err := view.Click("下载音频"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if len(backend.covers) != 1 || len(backend.audios) != 1 || backend.audios[0].AudioURL == "" {
		t.Fatalf("extra downloads: covers=%d audios=%+v", len(backend.covers), backend.audios)
	}
	if err := view.Click("下载视频"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if len(backend.tasks) != 1 || len(s.stagedKeys) != 0 {
		t.Fatalf("single download: %+v staged=%d", backend.tasks, len(s.stagedKeys))
	}
}

func TestClearingInputDiscardsPendingParse(t *testing.T) {
	s, base, runtime := setup(t, true)
	s.Configure(&manualDownloadAPI{fakeAPI: base, detail: model.VideoDetailConciseData{View: model.VideoDetailView{Bvid: "BV1xx411c7mD", Title: "稍后返回"}}}, runtime.runtime())
	s.link = "BV1xx411c7mD"
	s.parseDownloadInput()
	s.link = ""
	s.manualParseInput = ""
	s.manualParse.seq++
	runtime.flush(t)
	if len(s.stagedKeys) != 0 || s.manualGroup != nil {
		t.Fatal("cleared input staged stale result")
	}
}

func TestDownloadToolbarStaysFixedAndActionsFit(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, base, runtime := setup(t, true)
		background := runtime.runtime()
		background.Background = background.Run
		api := &manualDownloadAPI{fakeAPI: base, detail: model.VideoDetailConciseData{View: model.VideoDetailView{Bvid: "BV1xx411c7mD", Cid: 42, Title: "测试视频"}}}
		s.Configure(api, background)
		s.ConfigureDownloads(&manualExtraBackend{})
		for i := 0; i < 8; i++ {
			s.stageManual(manualVideo{Key: fmt.Sprintf("BVtest%d", i), Bvid: fmt.Sprintf("BVtest%d", i), Title: fmt.Sprintf("测试视频 %d", i)})
		}
		path := "/bilibili/download"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			style.Apply(c)
			s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() {}) })
		}, width, 600)
		view.Frame()
		runtime.flush(t)
		view.Frame()
		toolbar, _ := view.Find("下载操作栏")
		input, _ := view.Find("视频下载输入栏")
		region, _ := view.Find("下载任务区域")
		video, _ := view.Find("下载视频")
		quality, _ := view.Find("视频画质")
		card, _ := view.Find("BVtest0")
		if width == 1200 && (abs(video.Y+video.H/2-quality.Y-quality.H/2) > 1 || card.H > 115) {
			t.Fatalf("download card is not compact: card=%+v quality=%+v download=%+v", card, quality, video)
		}
		remove, _ := view.Find("移除")
		if abs(remove.Y+remove.H/2-video.Y-video.H/2) > 1 {
			t.Fatal("下载按钮组被拆开换行")
		}
		if toolbar.H > 60 || video.X < region.X || remove.X+remove.W > region.X+region.W {
			t.Fatalf("download controls overflow at %d: toolbar=%+v video=%+v remove=%+v", width, toolbar, video, remove)
		}
		scroll := s.scroll("download-items")
		if scroll.MaxY <= 0 {
			t.Fatal("download list does not scroll")
		}
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-download-list-%d.png", width))
		view.Scroll(region.X+region.W/2, region.Y+region.H/2, 0, 250)
		view.Frame()
		inputAfter, _ := view.Find("视频下载输入栏")
		toolbarAfter, _ := view.Find("下载操作栏")
		if scroll.Y <= 0 || inputAfter.Y != input.Y || toolbarAfter.Y != toolbar.Y {
			t.Fatal("download toolbar moved while scrolling")
		}
	}
}
