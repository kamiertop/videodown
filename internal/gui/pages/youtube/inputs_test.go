package youtube

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/internal/storage"
	yt "github.com/kamiertop/videodown/youtube"
)

func TestLinkInputClearAndEnter(t *testing.T) {
	for _, tc := range []struct {
		page, field, clear, link string
		value                    func(*State) *string
	}{
		{"/youtube/download", "YouTube 视频链接", "清空视频链接", "https://youtu.be/jNQXAC9IVRw", func(s *State) *string { return &s.Link }},
		{"/youtube/playlist", "YouTube 播放列表链接", "清空播放列表链接", "https://www.youtube.com/playlist?list=PLfixture", func(s *State) *string { return &s.playlistLink }},
	} {
		t.Run(tc.page, func(t *testing.T) {
			store, err := storage.OpenMemory()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			s := NewState()
			runtime := syncRuntime()
			scheduled := 0
			runtime.Background = func(func() error, func(error)) { scheduled++ }
			s.Configure(yt.New(store), runtime)
			t.Cleanup(s.Stop)
			value := tc.value(s)
			*value = "old link"
			test := ui.NewTester(func(c *ui.Context) {
				style.UseTheme(c)
				s.Form(c, tc.page, nil)
			}, 800, 650)
			test.Frame()
			if err := test.Click(tc.field); err != nil {
				t.Fatal(err)
			}
			if err := test.Click(tc.clear); err != nil {
				t.Fatal(err)
			}
			test.Frame()
			if *value != "" {
				t.Fatalf("clear left value %q", *value)
			}
			test.Key(0, ui.KeyEnter)
			if scheduled != 0 {
				t.Fatal("Enter scheduled work with an empty input")
			}
			test.Type(tc.link)
			test.Frame()
			if *value != tc.link {
				t.Fatalf("typing after clear wrote %q, want %q", *value, tc.link)
			}
			test.Key(0, ui.KeyEnter)
			test.Frame()
			test.Key(0, ui.KeyEnter)
			test.Frame()
			if scheduled != 1 {
				t.Fatalf("scheduled %d parse requests, want one", scheduled)
			}
		})
	}
}

func TestAccountCookieGuideOpensWiki(t *testing.T) {
	s := NewState()
	scroll := &ui.ScrollState{}
	test := ui.NewTester(func(c *ui.Context) {
		style.UseTheme(c)
		s.Form(c, "/youtube/profile", scroll)
	}, 1200, 760)
	test.Frame()
	notice, noticeVisible := test.Find("Cookie 可能很快失效")
	guide, guideVisible := test.Find("普通公开视频和公开播放列表无需 Cookie")
	editor, editorVisible := test.Find("YouTube Cookie")
	if !noticeVisible || !guideVisible || !editorVisible || notice.Y >= editor.Y || guide.X <= editor.X {
		t.Fatalf("expected a top notice and side-by-side panels, got notice=%+v guide=%+v editor=%+v", notice, guide, editor)
	}
	if err := test.Click("查看 yt-dlp Cookie 导出指南"); err != nil {
		t.Fatal(err)
	}
	urls := test.OpenedURLs()
	if len(urls) != 1 || urls[0] != cookieWikiURL {
		t.Fatalf("opened %v, want the yt-dlp Cookie guide", urls)
	}

	test.SetSize(680, 980)
	test.Frame()
	guide, guideVisible = test.Find("普通公开视频和公开播放列表无需 Cookie")
	editor, editorVisible = test.Find("YouTube Cookie")
	if !guideVisible || !editorVisible || guide.Y <= editor.Y || guide.X != editor.X {
		t.Fatalf("expected stacked panels, got guide=%+v editor=%+v", guide, editor)
	}
}
