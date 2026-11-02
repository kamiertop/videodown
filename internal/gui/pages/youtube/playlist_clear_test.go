package youtube

import (
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
)

func TestPlaylistBulkRemoveClearAndLatePagination(t *testing.T) {
	for _, width := range []int{760, 1200} {
		s, _, jobs := newQueueTest()
		s.playlistLink = "https://www.youtube.com/playlist?list=PLfixture"
		s.loadPlaylist(false)
		takeQueueJob(t, jobs).run()
		s.loadPlaylist(true)
		takeQueueJob(t, jobs).run()
		for _, row := range s.playlistRows {
			row.Selected = true
		}
		s.enqueuePlaylist()
		s.playlist.Continuation = "next"
		scroll := &ui.ScrollState{}
		test := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, "/youtube/playlist", func(string) {}, func() { s.Form(c, "/youtube/playlist", scroll) })
		}, width, 600)
		frame := func() { test.Frame(); test.Advance(20 * time.Millisecond); test.Frame() }
		click := func(label string) {
			t.Helper()
			if err := test.Click(label); err != nil {
				t.Fatal(err)
			}
			frame()
		}
		assertEmpty := func() {
			t.Helper()
			if len(s.playlistRows) != 0 || s.playlist.Continuation != "" || s.playlistError != "" || s.playlistLoading {
				t.Fatal("parsed state not cleared")
			}
			for _, label := range []string{"YouTube 播放列表工具栏", "加载更多", "加入队列", "0 / 0"} {
				if _, ok := test.Find(label); ok {
					t.Fatalf("empty page retained %s", label)
				}
			}
			if s.playlistLink == "" || len(s.queue) != 2 || s.queue[0].Status != "ready" {
				t.Fatal("clearing source changed link or download tasks")
			}
		}
		frame()
		for _, label := range []string{"移除已选", "清空列表", "加入队列"} {
			b, ok := test.Find(label)
			if !ok || b.X+b.W > float32(width) {
				t.Fatalf("toolbar clipped: %s %+v", label, b)
			}
		}
		click("全选")
		if !s.playlistRows[0].Selected || !s.playlistRows[1].Selected {
			t.Fatal("queued source rows cannot be selected for bulk removal")
		}
		click("移除已选")
		assertEmpty()
		if width == 760 {
			saveQueueScreenshot(t, test, "/tmp/youtube-playlist-cleared.png")
		}
		s.loadPlaylist(false)
		takeQueueJob(t, jobs).run()
		s.loadPlaylist(true)
		late := takeQueueJob(t, jobs)
		frame()
		click("清空列表")
		late.run()
		frame()
		assertEmpty()
		s.loadPlaylist(false)
		takeQueueJob(t, jobs).run()
		frame()
		click("移除列表视频：jNQXAC9IVRw")
		assertEmpty()
	}
}

func TestRemovedSourceVideoDoesNotReappearOnMore(t *testing.T) {
	s, _, jobs := newQueueTest()
	s.playlistLink = "https://www.youtube.com/playlist?list=PLfixture"
	s.loadPlaylist(false)
	takeQueueJob(t, jobs).run()
	removed := s.playlistRows[0]
	s.playlistRows = append(s.playlistRows, &playlistRow{Video: queueVideo("otherVideo1")})
	s.loadPlaylist(true)
	late := takeQueueJob(t, jobs)
	s.removePlaylistRows(func(row *playlistRow) bool { return row == removed })
	late.run()
	if len(s.playlistRows) != 2 {
		t.Fatal("pagination failed to retain remaining entries")
	}
	for _, row := range s.playlistRows {
		if row.Video.ID == removed.Video.ID {
			t.Fatal("removed source video reappeared")
		}
	}
}
