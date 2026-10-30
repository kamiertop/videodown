package youtube

import (
	"context"
	"errors"
	"image"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	yt "github.com/kamiertop/videodown/youtube"
)

func TestCoverFallbackAndQueuedVideoKeepPendingImage(t *testing.T) {
	bitmap, err := loadCoverWith(context.Background(), []string{"missing", "image"}, func(ctx context.Context, url, referer string) (*ui.Bitmap, error) {
		if url == "missing" {
			return nil, errors.New("404")
		}
		return ui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 16, 9))), nil
	})
	if err != nil || bitmap == nil {
		t.Fatalf("cover fallback: %v", err)
	}
	s := NewState()
	var jobs []queueJob
	s.thumbnailLoader = func(context.Context, []string) (*ui.Bitmap, error) { return bitmap, nil }
	s.Configure(&fakeAPI{}, guiruntime.Runtime{Background: func(work func() error, done func(error)) { done(work()) }, Image: func(work func() error, done func(error)) { jobs = append(jobs, queueJob{work, done}) }}, func(fn func()) { fn() })
	s.resolve()
	id := s.video.ID
	s.stagePreview()
	if len(jobs) != 1 {
		t.Fatal("preview must request one shared image")
	}
	jobs[0].run()
	if s.thumbnails[id].Bitmap != bitmap || s.video.ID != "" {
		t.Fatal("queue lost the pending image or retained duplicate preview")
	}
}

func TestThumbnailRetryAndClosedState(t *testing.T) {
	s, _, _ := newQueueTest()
	var jobs []queueJob
	s.runtime.Image = func(work func() error, done func(error)) { jobs = append(jobs, queueJob{work, done}) }
	bitmap := ui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 16, 9)))
	failed := true
	s.thumbnailLoader = func(context.Context, []string) (*ui.Bitmap, error) {
		if failed {
			failed = false
			return nil, errors.New("network")
		}
		return bitmap, nil
	}
	video := yt.Video{ID: "jNQXAC9IVRw"}
	test := ui.NewTester(func(c *ui.Context) { core.Use(c, core.Settings{}); s.thumbnailFrame(c, video, 96, 54, "封面") }, 200, 100)
	test.Frame()
	test.Advance(10 * time.Millisecond)
	test.Frame()
	if len(jobs) != 1 {
		t.Fatal("missing initial thumbnail request")
	}
	jobs[0].run()
	jobs = nil
	test.Frame()
	if err := test.Click("重试封面：" + video.ID); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if len(jobs) != 1 {
		t.Fatal("retry was not scheduled")
	}
	jobs[0].run()
	jobs = nil
	test.Frame()
	if s.thumbnails[video.ID].Bitmap != bitmap {
		t.Fatal("retry image not published")
	}
	other := yt.Video{ID: "KWGhxyu5osA"}
	entry := s.thumbnailEntry(other)
	s.requestThumbnail(other, entry)
	s.Stop()
	jobs[0].run()
	if entry.Bitmap != nil {
		t.Fatal("closed state accepted late thumbnail")
	}
}
