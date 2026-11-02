package youtube

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	yt "github.com/kamiertop/videodown/youtube"
)

func TestPlaylistGroupCollapseRemoveAndSourceStatus(t *testing.T) {
	s, _, jobs := newQueueTest()
	s.playlist = yt.Playlist{ID: "course", Title: "Go 入门课程"}
	s.playlistRows = []*playlistRow{{Video: queueVideo("jNQXAC9IVRw"), Selected: true}, {Video: queueVideo("KWGhxyu5osA"), Selected: true}}
	s.enqueuePlaylist()
	if len(s.queueGroups) != 1 || s.queueGroups[0].Expanded || s.playlistRows[0].Selected || s.playlistAvailable(s.playlistRows[0]) {
		t.Fatal("playlist group must start folded and mark source rows")
	}
	s.enqueuePlaylist()
	if len(s.queue) != 2 {
		t.Fatal("repeated add duplicated a group")
	}
	if len(*jobs) != 0 {
		t.Fatal("adding a playlist must not start downloading")
	}
	s.startGroup(s.queueGroups[0])
	active := takeQueueJob(t, jobs)
	group := s.queueGroups[0]
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) { core.Use(c, core.Settings{}); s.Form(c, "/youtube/download", scroll) }, 760, 600)
	test.Frame()
	if _, ok := test.Find("YouTube 队列任务：1"); ok {
		t.Fatal("folded group rendered its task rows")
	}
	if !test.HasText("Go 入门课程") {
		t.Fatal("playlist theme missing")
	}
	if err := test.Click("展开或收起分组：" + group.Key); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if _, ok := test.Find("YouTube 队列任务：1"); !ok {
		t.Fatal("expand did not show tasks")
	}
	if err := test.Click("展开或收起分组：" + group.Key); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if _, ok := test.Find("YouTube 队列任务：1"); ok {
		t.Fatal("collapse did not hide tasks")
	}
	if err := test.Click("移除分组：" + group.Key); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(s.queue) != 0 || len(s.queueGroups) != 0 || !s.playlistAvailable(s.playlistRows[0]) || s.PendingDownloadCount() != 1 {
		t.Fatal("remove did not cancel group or lost active cleanup accounting")
	}
	s.enqueue(queueVideo("nextVideo01"), 18, false)
	s.startAll()
	if len(*jobs) != 0 {
		t.Fatal("next transfer started before canceled worker drained")
	}
	active.run()
	if len(s.active) != 1 || s.active[s.queue[len(s.queue)-1]] == nil {
		t.Fatal("removed group blocked next group")
	}
	takeQueueJob(t, jobs).run()
	if s.PendingDownloadCount() != 0 {
		t.Fatal("finished queue still pending")
	}
}

func TestVisibleCoversSharedBetweenPlaylistAndQueue(t *testing.T) {
	for _, width := range []int{760, 1200} {
		s, _, _ := newQueueTest()
		var imageJobs []queueJob
		s.runtime.Image = func(work func() error, done func(error)) { imageJobs = append(imageJobs, queueJob{work, done}) }
		bitmapImage := image.NewRGBA(image.Rect(0, 0, 160, 90))
		draw.Draw(bitmapImage, bitmapImage.Bounds(), image.NewUniform(color.RGBA{R: 98, G: 134, B: 158, A: 255}), image.Point{}, draw.Src)
		bitmap := ui.NewBitmap(bitmapImage)
		s.thumbnailLoader = func(_ context.Context, urls []string) (*ui.Bitmap, error) {
			if len(urls) < 2 || urls[1] == "" {
				t.Fatal("missing thumbnail fallback")
			}
			return bitmap, nil
		}
		s.playlist = yt.Playlist{ID: "course", Title: "Go 入门课程"}
		for i := range 30 {
			s.playlistRows = append(s.playlistRows, &playlistRow{Video: yt.Video{ID: fmt.Sprintf("video%06d", i), Title: fmt.Sprintf("第 %d 课：学习 Go", i+1), Author: "课程作者", Duration: 10 * time.Minute}, Selected: true})
		}
		path := "/youtube/playlist"
		scrolls := map[string]*ui.ScrollState{path: {}, "/youtube/download": {}}
		test := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(next string) { path = next }, func() { s.Form(c, path, scrolls[path]) })
		}, width, 760)
		test.Frame()
		test.Advance(10 * time.Millisecond)
		test.Frame()
		if len(imageJobs) == 0 || len(imageJobs) >= 30 {
			t.Fatalf("visible covers not loaded lazily: %d", len(imageJobs))
		}
		for _, job := range imageJobs {
			job.run()
		}
		imageJobs = nil
		test.Frame()
		first := s.playlistRows[0]
		if s.thumbnails[first.Video.ID].Bitmap != bitmap {
			t.Fatal("cover was not published")
		}
		s.enqueuePlaylist()
		test.Frame()
		if !test.HasText("已加入队列") || first.Selected || s.playlistAvailable(first) {
			t.Fatal("source page failed to reflect queued status")
		}
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-playlist-status.png")
		}
		path = "/youtube/download"
		test.Frame()
		test.Advance(10 * time.Millisecond)
		test.Frame()
		if len(imageJobs) != 0 {
			t.Fatal("folded group loaded child covers or refetched shared cover")
		}
		group := s.queueGroups[0]
		b, ok := test.Find("YouTube 队列分组：" + group.Key)
		if !ok || b.W > float32(width) || b.H > 130 {
			t.Fatalf("folded group too large: %+v", b)
		}
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-group-collapsed.png")
		}
		if err := test.Click("展开或收起分组：" + group.Key); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		test.Advance(10 * time.Millisecond)
		test.Frame()
		for _, job := range imageJobs {
			job.run()
		}
		imageJobs = nil
		test.Frame()
		cover, ok := test.Find("任务封面：1")
		if !ok || cover.W <= 0 || cover.X+cover.W > float32(width) {
			t.Fatal("expanded task cover missing")
		}
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-group-expanded.png")
		}
	}
}
