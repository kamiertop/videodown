package bilibili

import (
	"fmt"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	"github.com/kamiertop/videodown/bilibili/model"
)

func downloadTester(s *State) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.download(c)
	}, 1200, 880)
}

func stageTestUploads(s *State, count int) {
	for i := range count {
		s.stageUpload(model.VListItem{Bvid: fmt.Sprintf("BV%d", i), Title: fmt.Sprintf("视频 %03d", i)})
	}
}

func TestDownloadListEmptyAndRemove(t *testing.T) {
	s := NewState()
	tester := downloadTester(s)
	tester.Frame()
	if _, ok := tester.Find("下载列表为空"); !ok {
		t.Fatal("empty queue state missing")
	}

	stageTestUploads(s, 1)
	tester.Frame()
	if _, ok := tester.Find("视频 000"); !ok {
		t.Fatal("staged card missing")
	}
	if err := tester.Click("移除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if s.stagedKeys["BV0"] {
		t.Fatal("remove button did not remove the task")
	}
	if _, ok := tester.Find("下载列表为空"); !ok {
		t.Fatal("empty queue state did not return")
	}
}

func TestDownloadListVirtualizesAndKeepsAnchor(t *testing.T) {
	s := NewState()
	stageTestUploads(s, 100)
	tester := downloadTester(s)
	tester.Frame()
	if _, ok := tester.Find("视频 050"); ok {
		t.Fatal("distant card was built before scrolling")
	}

	s.downloadList.ScrollTo(50, ui.Start)
	tester.Frame()
	first, _ := s.downloadList.Visible()
	if first != 50 {
		t.Fatalf("first visible row = %d, want 50", first)
	}
	if _, ok := tester.Find("视频 050"); !ok {
		t.Fatal("card missing after scrolling to it")
	}
	s.removeStaged("BV0")
	tester.Frame()
	first, _ = s.downloadList.Visible()
	if first != 49 || s.downloadList.Key(first) != "BV50" {
		t.Fatalf("first visible row = %d, want the original card at 49", first)
	}
}

func TestDownloadListRevealsNewParsedGroup(t *testing.T) {
	s := NewState()
	stageTestUploads(s, 20)
	tester := downloadTester(s)
	tester.Frame()
	s.downloadList.ScrollTo(10, ui.Start)
	tester.Frame()

	s.manualGroup = &parsedVideoGroup{
		Kind: "合集", Title: "解析测试",
		Items:        []manualVideo{{Key: "BVpart", Title: "分 P"}},
		OpenSections: []string{"videos"},
	}
	tester.Frame()
	first, _ := s.downloadList.Visible()
	if first != 0 {
		t.Fatalf("first visible row = %d, want parsed group at 0", first)
	}
	if _, ok := tester.Find("添加当前视频"); !ok {
		t.Fatal("parsed group actions missing")
	}

	s.downloadList.ScrollTo(10, ui.Start)
	tester.Frame()
	first, _ = s.downloadList.Visible()
	if first != 10 {
		t.Fatalf("first visible row = %d, want 10 after scrolling away from parsed group", first)
	}
}

func TestDownloadListShowsBatchWithStagedTasks(t *testing.T) {
	s := NewState()
	stageTestUploads(s, 1)
	s.batches = []*batchSession{{
		ID: 1, Source: batchSource{Title: "批量任务"}, Status: "done",
	}}
	tester := downloadTester(s)
	tester.Frame()
	for _, label := range []string{"批量任务", "视频 000"} {
		if _, ok := tester.Find(label); !ok {
			t.Fatalf("missing row %q", label)
		}
	}
	if s.downloadList.Key(0) != uint64(1) || s.downloadList.Key(1) != "BV0" {
		t.Fatal("batch should precede staged task")
	}
	if err := tester.Click("关闭批量：1"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(s.batches) != 0 {
		t.Fatal("batch close button did not dismiss the batch")
	}
	if _, ok := tester.Find("视频 000"); !ok {
		t.Fatal("staged task missing after dismissing the batch")
	}
}

func TestDownloadListReadiness(t *testing.T) {
	s := NewState()
	stageTestUploads(s, 2)
	s.queueResolves["BV0"] = &queueResolve{data: resource[biliapi.PlayUrlResult]{
		Loaded: true,
		Data: biliapi.PlayUrlResult{
			Detail:  &model.VideoDetailConciseData{},
			PlayURL: &model.VideoURLData{Durl: []model.DurlItem{{Url: "https://example.com/video.mp4"}}},
		},
	}}
	s.queueResolves["BV1"] = &queueResolve{data: resource[biliapi.PlayUrlResult]{
		Loaded: true,
		Data: biliapi.PlayUrlResult{
			Detail:  &model.VideoDetailConciseData{},
			PlayURL: &model.VideoURLData{},
		},
	}}
	_, ready := s.downloadRows()
	if ready != 1 {
		t.Fatalf("ready downloads = %d, want 1", ready)
	}
}
