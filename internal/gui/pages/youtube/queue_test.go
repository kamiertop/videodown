package youtube

import (
	"context"
	"errors"
	"image/png"
	"os"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	yt "github.com/kamiertop/videodown/youtube"
)

type queueAPI struct {
	fakeAPI
	fail    bool
	formats []int
}

func queueVideo(id string) yt.Video {
	return yt.Video{ID: id, Title: "视频 " + id, Author: "示例作者", Formats: []yt.Format{{ID: 137, Height: 1080, Label: "1080p · H.264 · MP4", Summary: "1080p · H.264"}, {ID: 136, Height: 720, Label: "720p · H.264 · MP4", Summary: "720p · H.264"}, {ID: 18, Height: 360, Label: "360p · H.264 · MP4", Summary: "360p · H.264"}}, AudioFormats: []yt.Format{{ID: 140, Label: "AAC · 128 kbps · M4A", Summary: "AAC · 128 kbps"}}}
}
func (a *queueAPI) Resolve(ctx context.Context, input string) (yt.Video, error) {
	if err := ctx.Err(); err != nil {
		return yt.Video{}, err
	}
	return queueVideo(input), nil
}
func (a *queueAPI) Download(ctx context.Context, v yt.Video, id int, progress func(yt.Progress)) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	a.formats = append(a.formats, id)
	if a.fail {
		a.fail = false
		return "", errors.New("网络中断")
	}
	progress(yt.Progress{Phase: "下载音频", Percent: 50})
	return "/download/" + v.ID + ".m4a", nil
}
func (a *queueAPI) Playlist(ctx context.Context, input, token string) (yt.Playlist, error) {
	if err := ctx.Err(); err != nil {
		return yt.Playlist{}, err
	}
	page := yt.Playlist{ID: "PLfixture", Title: "示例课程", Videos: []yt.Video{{ID: "jNQXAC9IVRw", Title: "第一课"}}}
	if token == "" {
		page.Continuation = "next"
	} else {
		page.Videos = append(page.Videos, yt.Video{ID: "KWGhxyu5osA", Title: "第二课"})
	}
	return page, nil
}

type queueJob struct {
	work func() error
	done func(error)
}

func (j queueJob) run() { j.done(j.work()) }
func newQueueTest() (*State, *queueAPI, *[]queueJob) {
	api := &queueAPI{}
	jobs := &[]queueJob{}
	s := NewState()
	s.Configure(api, guiruntime.Runtime{Background: func(work func() error, done func(error)) { *jobs = append(*jobs, queueJob{work, done}) }}, func(fn func()) { fn() })
	return s, api, jobs
}
func takeQueueJob(t *testing.T, jobs *[]queueJob) queueJob {
	t.Helper()
	if len(*jobs) == 0 {
		t.Fatal("missing scheduled job")
	}
	j := (*jobs)[0]
	*jobs = (*jobs)[1:]
	return j
}

func TestQueueSerialRetryCancelAndIndependentParsing(t *testing.T) {
	s, api, jobs := newQueueTest()
	s.enqueue(queueVideo("jNQXAC9IVRw"), 18, false)
	if s.enqueue(queueVideo("jNQXAC9IVRw"), 18, false) {
		t.Fatal("duplicate task admitted")
	}
	s.enqueue(yt.Video{ID: "KWGhxyu5osA", Title: "第二课"}, -1, true)
	s.startAll()
	if len(*jobs) != 1 || s.PendingDownloadCount() != 2 {
		t.Fatal("queue must start one job")
	}
	active := takeQueueJob(t, jobs)
	s.Link = "newVideo001"
	s.resolve()
	takeQueueJob(t, jobs).run()
	if s.video.ID != "newVideo001" || len(s.active) == 0 {
		t.Fatal("parsing blocked by active download")
	}
	s.cancelItem(s.queue[0])
	active.run()
	if s.queue[0].Status != "canceled" || s.active[s.queue[1]] == nil {
		t.Fatal("cancel did not advance queue")
	}
	api.fail = true
	takeQueueJob(t, jobs).run()
	if s.queue[1].Status != "error" || s.PendingDownloadCount() != 0 {
		t.Fatal("failure state missing")
	}
	s.retryItem(s.queue[1])
	takeQueueJob(t, jobs).run()
	if s.queue[1].Status != "done" || api.formats[len(api.formats)-1] != 140 || s.queue[1].AudioOnly != true {
		t.Fatal("retry lost audio choice")
	}
	s.queuePaused = true
	s.enqueue(queueVideo("nextVideo01"), 18, false)
	s.startAll()
	if len(*jobs) != 0 {
		t.Fatal("paused queue started")
	}
	s.cancelItem(s.queue[2])
	if s.PendingDownloadCount() != 0 {
		t.Fatal("queued cancel failed")
	}
	s.removeItem(s.queue[0])
	if len(s.queue) != 2 {
		t.Fatal("cannot remove ended task")
	}
	s.queuePaused = false
	s.enqueue(queueVideo("lateVideo01"), 18, false)
	s.startAll()
	j := takeQueueJob(t, jobs)
	s.Stop()
	j.run()
	if s.queue[len(s.queue)-1].Status == "done" {
		t.Fatal("closed state accepted worker result")
	}
}

func TestPlaylistSelectionPaginationAndStaleResults(t *testing.T) {
	s, _, jobs := newQueueTest()
	s.playlistLink = "https://www.youtube.com/playlist?list=PLfixture"
	s.loadPlaylist(false)
	old := takeQueueJob(t, jobs)
	s.loadPlaylist(false)
	old.run()
	if !s.playlistLoading || len(s.playlistRows) != 0 {
		t.Fatal("stale playlist completion applied")
	}
	takeQueueJob(t, jobs).run()
	s.playlistRows[0].Selected = false
	s.loadPlaylist(true)
	takeQueueJob(t, jobs).run()
	if len(s.playlistRows) != 2 || s.playlistRows[0].Selected {
		t.Fatal("pagination lost selection or duplicated video")
	}
	s.playlistRows[1].Selected = true
	s.enqueuePlaylist()
	if len(s.queue) != 1 || s.queue[0].Video.ID != "KWGhxyu5osA" || s.queue[0].AudioOnly || s.queue[0].Format != 0 {
		t.Fatal("playlist must stage videos without starting or choosing audio")
	}
	if len(*jobs) != 0 {
		t.Fatal("adding playlist started a job")
	}
	s.queue[0].AudioOnly = true
	s.startItem(s.queue[0])
	if len(*jobs) != 1 {
		t.Fatal("explicit start did not create job")
	}
	takeQueueJob(t, jobs).run()
	if s.queue[0].Status != "done" {
		t.Fatal("playlist task did not resolve and download")
	}
}

func TestQueuePlaylistAndAudioLayouts(t *testing.T) {
	for _, width := range []int{760, 1200} {
		s, _, jobs := newQueueTest()
		s.video = queueVideo("jNQXAC9IVRw")
		s.selectDefaultFormat()
		s.enqueue(s.video, 18, false)
		s.enqueue(yt.Video{ID: "audioVideo1", Title: "第二个任务：仅音频下载"}, -1, true)
		s.startAll()
		s.queue[0].Progress = yt.Progress{Phase: "下载视频", Percent: 35, Speed: 2 << 20}
		path := "/youtube/download"
		scrolls := map[string]*ui.ScrollState{path: {}, "/youtube/playlist": {}}
		test := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(next string) { path = next }, func() { s.Form(c, path, scrolls[path]) })
		}, width, 760)
		test.Frame()
		for _, label := range []string{"YouTube 队列工具栏", "YouTube 下载视频卡片", "YouTube 视频输入栏"} {
			b, ok := test.Find(label)
			if !ok || b.X+b.W > float32(width) {
				t.Fatalf("clipped %s %+v", label, b)
			}
		}
		inputBar, _ := test.Find("YouTube 视频输入栏")
		card, _ := test.Find("YouTube 下载视频卡片")
		audio, _ := test.Find("解析卡片音频模式")
		if audio.Y < inputBar.Y+inputBar.H || audio.Y < card.Y || audio.Y+audio.H > card.Y+card.H {
			t.Fatal("audio choice must belong to the video card")
		}
		if err := test.Click("解析卡片音频模式"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if s.format != 0 || !test.HasText("加入队列") || !test.HasText("最佳音质") {
			t.Fatal("audio mode did not update preview")
		}
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-queue.png")
		}
		if err := test.Click("YouTube播放列表"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		s.playlistLink = "https://www.youtube.com/playlist?list=PLfixture"
		s.loadPlaylist(false)
		// The active download is still scheduled; finish only the list request.
		j := (*jobs)[len(*jobs)-1]
		*jobs = (*jobs)[:len(*jobs)-1]
		j.run()
		s.loadPlaylist(true)
		j = (*jobs)[len(*jobs)-1]
		*jobs = (*jobs)[:len(*jobs)-1]
		j.run()
		test.Frame()
		if !test.HasText("第一课") {
			t.Fatal("playlist rows missing")
		}
		for _, label := range []string{"YouTube 播放列表输入栏", "YouTube 播放列表工具栏", "加入队列"} {
			b, ok := test.Find(label)
			if !ok || b.X+b.W > float32(width) {
				t.Fatalf("clipped %s %+v", label, b)
			}
		}
		if err := test.Click("全选"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if err := test.Click("取消全选"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if s.playlistRows[0].Selected {
			t.Fatal("deselect button failed")
		}
		if err := test.Click("全选"); err != nil {
			t.Fatal(err)
		}
		test.Frame()
		if !s.playlistRows[0].Selected || !s.playlistRows[1].Selected {
			t.Fatal("select all must allow removing queued source rows too")
		}
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-playlist.png")
		}
	}
}

func saveQueueScreenshot(t *testing.T, test *ui.Tester, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, test.Image()); err != nil {
		t.Fatal(err)
	}
}
