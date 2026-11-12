package youtube

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	yt "github.com/kamiertop/videodown/youtube"
)

func TestDownloadListVirtualizesExpandedGroup(t *testing.T) {
	s := NewState()
	s.ctx = context.Background()
	var scheduled int
	s.runtime.Background = func(func() error, func(error)) { scheduled++ }
	group := &queueGroup{Key: "playlist", Title: "测试分组"}
	s.queueGroups = []*queueGroup{group}
	for i := range 60 {
		s.queue = append(s.queue, &queueItem{
			ID: uint64(i + 1), Group: group, Status: "queued",
			Video: yt.Video{ID: fmt.Sprint(i), Title: fmt.Sprintf("队列视频 %d", i)},
		})
	}
	target := s.queue[40]
	target.Status = "ready"
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Form(c, "/youtube/download", nil)
	}, 1200, 720)
	test.Frame()
	if test.HasText("YouTube 队列任务：") || scheduled != 0 {
		t.Fatal("collapsed group rendered or prepared its tasks")
	}
	if err := test.Click("展开或收起分组：playlist"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	test.Advance(2 * time.Millisecond)
	if !group.Expanded || !test.HasText("YouTube 队列任务：1") || scheduled != 0 {
		t.Fatalf("expand: open=%v, first task visible=%v, scheduled=%d", group.Expanded, test.HasText("YouTube 队列任务：1"), scheduled)
	}

	s.downloadList.ScrollTo(42, ui.Start)
	test.Frame()
	test.Advance(2 * time.Millisecond)
	if _, ok := test.Find(fmt.Sprintf("YouTube 队列任务：%d", target.ID)); !ok || !target.DetailsLoading || scheduled != 1 {
		t.Fatalf("visible task should schedule one format lookup, scheduled %d", scheduled)
	}
	s.removeItem(s.queue[0])
	test.Frame()
	first, _ := s.downloadList.Visible()
	if first != 41 || s.downloadList.Key(first) != target {
		t.Fatalf("first visible row = %d, want the same task at 41", first)
	}

	s.video = yt.Video{ID: "preview", Title: "新解析视频"}
	s.revealPreview = true
	test.Frame()
	first, _ = s.downloadList.Visible()
	if first != 0 || !test.HasText("新解析视频") || scheduled != 1 {
		t.Fatal("new preview should reveal the first row without scheduling another task lookup")
	}
	s.downloadList.ScrollTo(42, ui.Start)
	test.Frame()
	test.Advance(2 * time.Millisecond)
	if scheduled != 1 {
		t.Fatalf("revisiting a task with a lookup in progress scheduled %d lookups", scheduled)
	}
}

func TestPlaylistListKeepsSelectionAndPagination(t *testing.T) {
	s := NewState()
	s.api = yt.New(nil)
	s.ctx = context.Background()
	var scheduled int
	s.runtime.Background = func(func() error, func(error)) { scheduled++ }
	s.playlist = yt.Playlist{Title: "测试播放列表", Continuation: "next"}
	for i := range 60 {
		s.playlistRows = append(s.playlistRows, &playlistRow{
			Video: yt.Video{ID: fmt.Sprint(i), Title: fmt.Sprintf("列表视频 %d", i)},
		})
	}
	firstRow := s.playlistRows[0]
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Form(c, "/youtube/playlist", nil)
	}, 1200, 720)
	test.Frame()
	if err := test.Click("选择视频：0"); err != nil {
		t.Fatal(err)
	}
	if !firstRow.Selected {
		t.Fatal("click did not select the video")
	}

	s.playlistList.ScrollToEnd()
	test.Frame()
	if _, ok := test.Find("列表视频 59"); !ok {
		t.Fatal("last video missing at end of list")
	}
	if err := test.Click("加载更多"); err != nil {
		t.Fatal(err)
	}
	if scheduled != 1 || !s.playlistLoading {
		t.Fatalf("pagination should schedule one request, scheduled %d", scheduled)
	}

	s.playlistList.ScrollTo(0, ui.Start)
	test.Frame()
	if !firstRow.Selected || !test.HasText("1 / 60") {
		t.Fatal("selection was lost after scrolling out of view")
	}
	if err := test.Click("移除列表视频：0"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(s.playlistRows) != 59 || s.playlistList.Key(0) != "1" {
		t.Fatal("removing a playlist row did not update the list")
	}
}
