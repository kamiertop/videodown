package youtube

import (
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	yt "github.com/kamiertop/videodown/youtube"
)

func TestParsedCardSupportsAutomaticBestFormat(t *testing.T) {
	for _, audio := range []bool{false, true} {
		s, api, jobs := newQueueTest()
		s.Link = "jNQXAC9IVRw"
		s.resolve()
		takeQueueJob(t, jobs).run()
		scroll := &ui.ScrollState{}
		test := ui.NewTester(func(c *ui.Context) { core.Use(c, core.Settings{}); s.Form(c, "/youtube/download", scroll) }, 760, 600)
		click := func(label string) {
			t.Helper()
			if err := test.Click(label); err != nil {
				t.Fatal(err)
			}
			test.Frame()
			test.Advance(20 * time.Millisecond)
			test.Frame()
		}
		test.Frame()
		if s.format != 0 || !test.HasText("最高画质") {
			t.Fatal("parsed card must default to automatic best quality")
		}
		click("YouTube 下载画质")
		click("720p · H.264 · MP4")
		if s.format != 136 {
			t.Fatal("explicit quality no longer selectable")
		}
		click("YouTube 下载画质")
		click("最高画质")
		if s.format != 0 {
			t.Fatal("highest quality did not restore automatic selection")
		}
		if audio {
			click("解析卡片音频模式")
			if !test.HasText("最佳音质") {
				t.Fatal("missing automatic best audio")
			}
		}
		click("加入队列")
		if s.queue[0].Format != 0 || len(api.formats) != 0 {
			t.Fatal("staging froze automatic selection or started media")
		}
		s.startItem(s.queue[0])
		takeQueueJob(t, jobs).run()
		want := 137
		if audio {
			want = 140
		}
		if len(api.formats) != 1 || api.formats[0] != want {
			t.Fatal("automatic format not selected on download")
		}
	}
}

func TestParsedCardAudioChoiceIsIndependent(t *testing.T) {
	s, _, jobs := newQueueTest()
	s.Link = "jNQXAC9IVRw"
	s.resolve()
	takeQueueJob(t, jobs).run()
	s.audioOnly = true
	s.selectDefaultFormat()
	s.stagePreview()
	if len(s.queue) != 1 || !s.queue[0].AudioOnly || len(*jobs) != 0 {
		t.Fatal("card choice was not staged independently")
	}
	s.Link = "KWGhxyu5osA"
	s.resolve()
	takeQueueJob(t, jobs).run()
	if s.audioOnly || s.format != 0 || !s.queue[0].AudioOnly {
		t.Fatal("audio mode leaked between parsed cards")
	}
}

func TestPlaylistClickSelectionStageConfigureAndStart(t *testing.T) {
	for _, width := range []int{760, 1200} {
		s, api, jobs := newQueueTest()
		s.playlistLink = "https://www.youtube.com/playlist?list=PLfixture"
		s.loadPlaylist(false)
		takeQueueJob(t, jobs).run()
		s.loadPlaylist(true)
		takeQueueJob(t, jobs).run()
		path := "/youtube/playlist"
		scrolls := map[string]*ui.ScrollState{path: {}, "/youtube/download": {}}
		test := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(next string) { path = next }, func() { s.Form(c, path, scrolls[path]) })
		}, width, 760)
		frame := func() { test.Frame(); test.Advance(20 * time.Millisecond); test.Frame() }
		click := func(label string) {
			t.Helper()
			if err := test.Click(label); err != nil {
				t.Fatal(err)
			}
			frame()
		}
		frame()
		if test.HasText("仅下载音频") || s.playlistRows[0].Selected {
			t.Fatal("playlist should only pick videos, starting unselected")
		}
		title, _ := test.Find("YouTube 播放列表标题")
		toolbar, _ := test.Find("YouTube 播放列表工具栏")
		cover, _ := test.Find("列表封面：jNQXAC9IVRw")
		if title.Y < toolbar.Y || title.Y+title.H > toolbar.Y+toolbar.H || cover.W < 144 {
			t.Fatal("playlist header/cover layout incorrect")
		}
		click("选择视频：jNQXAC9IVRw")
		if !s.playlistRows[0].Selected {
			t.Fatal("card click did not select")
		}
		click("选择视频：jNQXAC9IVRw")
		if s.playlistRows[0].Selected {
			t.Fatal("card click did not deselect")
		}
		click("全选")
		click("移除列表视频：KWGhxyu5osA")
		if len(s.playlistRows) != 1 || !s.playlistRows[0].Selected {
			t.Fatal("remove affected another selection")
		}
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-playlist-selection.png")
		}
		click("加入队列")
		if len(api.formats) != 0 || len(*jobs) != 0 || len(s.queue) != 1 || s.queue[0].Status != "ready" || s.PendingDownloadCount() != 0 {
			t.Fatal("staging started downloading")
		}
		click("YouTube下载")
		click("展开或收起分组：playlist:PLfixture")
		for len(*jobs) > 0 {
			takeQueueJob(t, jobs).run()
		}
		frame()
		if len(api.formats) != 0 || len(s.queue[0].Video.Formats) == 0 {
			t.Fatal("format preparation must not transfer media")
		}
		click("任务画质：1")
		click("720p · H.264 · MP4")
		if s.queue[0].Format != 136 {
			t.Fatal("quality choice was not retained")
		}
		click("下载方式：1")
		if !s.queue[0].AudioOnly || s.queue[0].Format != 0 {
			t.Fatal("audio mode did not reset video format")
		}
		click("任务画质：1")
		click("AAC · 128 kbps · M4A")
		if s.queue[0].Format != 140 {
			t.Fatal("audio selection missing")
		}
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-staged-options.png")
		}
		click("下载任务：1")
		if len(api.formats) != 0 || len(*jobs) != 1 {
			t.Fatal("manual start did not schedule exactly one transfer")
		}
		takeQueueJob(t, jobs).run()
		frame()
		if s.queue[0].Status != "done" || len(api.formats) != 1 || api.formats[0] != 140 {
			t.Fatal("download ignored audio choice")
		}
	}
}

func TestReadyTaskRemovalAndResolutionLimit(t *testing.T) {
	s, api, jobs := newQueueTest()
	s.enqueue(yt.Video{ID: "jNQXAC9IVRw"}, 0, false)
	s.enqueue(yt.Video{ID: "KWGhxyu5osA"}, 0, false)
	first, second := s.queue[0], s.queue[1]
	s.prepareItem(first)
	s.removeItem(first)
	takeQueueJob(t, jobs).run()
	if len(s.queue) != 1 || first.DetailsLoading || len(first.Video.Formats) != 0 || len(api.formats) != 0 {
		t.Fatal("removed item accepted late format data or transferred media")
	}
	second.Format = -720
	s.startGroup(second.Group)
	takeQueueJob(t, jobs).run()
	if len(api.formats) != 1 || api.formats[0] != 136 {
		t.Fatal("resolution limit ignored")
	}
}
