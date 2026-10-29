package bilibili

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func TestLayoutAndNavigation(t *testing.T) {
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		s := NewState()
		path := "/bilibili/profile"
		test := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" }) })
		}, size[0], size[1])
		test.Frame()
		for _, item := range []struct{ label, section, text string }{
			{"视频下载", "download", "清空"}, {"UP 主", "up", "我的关注 · 0"},
			{"关注动态", "dynamic", "暂无关注视频动态"}, {"收藏夹", "favorite", "请选择收藏夹"},
			{"播放历史", "play-history", "暂无播放历史"}, {"下载历史", "history", "下载历史 · 0"},
			{"账号", "profile", "扫码登录"},
		} {
			if err := test.Click(item.label); err != nil {
				t.Fatal(err)
			}
			test.Frame()
			test.Frame() // additional rebuild catches stale elements and input identity.
			if path != "/bilibili/"+item.section || !test.HasText(item.text) {
				t.Fatalf("%s: path=%s, missing %s", item.label, path, item.text)
			}
			nav, ok := test.Find("哔哩哔哩导航")
			if !ok || nav.W != 88 {
				t.Fatalf("navigation width: %+v", nav)
			}
			pane, ok := test.Find("哔哩哔哩内容")
			if !ok || pane.X < nav.X+nav.W || pane.X+pane.W > float32(size[0])+.5 || pane.H != nav.H {
				t.Fatalf("pane overflows at %v: nav=%+v pane=%+v", size, nav, pane)
			}
			for _, label := range []string{"视频下载", "UP 主", "关注动态", "收藏夹", "播放历史", "下载历史", "账号"} {
				b, ok := test.Find(label)
				if !ok || b.X < nav.X || b.X+b.W > nav.X+nav.W+.5 || b.H < 60 {
					t.Fatalf("menu clipped at %v: %s %+v", size, label, b)
				}
			}
			for _, description := range []string{"添加视频", "待下载视频", "浏览我的关注，或通过 UID、空间链接查找 UP 主。", "管理收藏的视频与订阅的合集。", "先完成扫码登录，再进入下载流程"} {
				if test.HasText(description) {
					t.Fatalf("unexpected explanatory copy: %s", description)
				}
			}
			if item.section == "download" || item.section == "up" {
				label := "BiliBili视频链接"
				if item.section == "up" {
					label = "UP 主 UID 或空间链接"
				}
				field, ok := test.Find(label)
				if !ok || field.W < 180 || field.X < pane.X || field.X+field.W > pane.X+pane.W {
					t.Fatalf("input clipped at %v: %s %+v", size, label, field)
				}
			}
			suffix := ""
			if size[0] == 760 {
				suffix = "-760"
			}
			saveImage(t, test, "/tmp/bilibili-"+item.section+suffix+".png")
		}
	}
}

func TestInputAndTabsSurviveNavigation(t *testing.T) {
	s := NewState()
	path := "/bilibili/download"
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" }) })
	}, 760, 600)
	test.Frame()
	click := func(label string) {
		t.Helper()
		if err := test.Click(label); err != nil {
			t.Fatal(err)
		}
		test.Frame()
	}
	click("BiliBili视频链接")
	test.Type("BV1xx411c7mD")
	test.Frame()
	click("收藏夹")
	click("合集")
	if path != "/bilibili/collection" || s.favoriteTab != "collection" {
		t.Fatal("collection tab did not navigate to its route")
	}
	click("账号")
	click("收藏夹")
	if path != "/bilibili/favorite" || s.favoriteTab != "favorite" {
		t.Fatal("favorite route did not restore its tab")
	}
	click("视频下载")
	if s.link != "BV1xx411c7mD" {
		t.Fatal("download input lost")
	}
	click("清空")
	if s.link != "" {
		t.Fatal("clear did not reset input")
	}
}

func TestLibraryActionsUseVideoCardsAndDownloadPage(t *testing.T) {
	for _, tab := range []string{"favorite", "collection"} {
		t.Run(tab, func(t *testing.T) {
			s, _, runtime := setup(t, true)
			s.favoriteTab = tab
			path := "/bilibili/" + tab
			view := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				s.Layout(c, path, func(p string) { path = p }, func() {
					s.Form(c, path, func(p string) { path = p }, func() {})
				})
			}, 1200, 880)
			view.Frame()
			runtime.flush(t)
			view.Frame()
			if _, found := view.Find("搜索" + map[string]string{"favorite": "收藏夹", "collection": "合集"}[tab]); found {
				t.Fatal("library search field should be removed")
			}
			video := "收藏视频 1-1"
			if tab == "collection" {
				video = "合集视频 1"
			}
			if err := view.Click("选择视频：" + video); err != nil {
				t.Fatal(err)
			}
			view.Frame()
			if !view.HasText("下载已选 (1)") {
				t.Fatal("selected card did not update toolbar")
			}
			if err := view.Click("下载已选 (1)"); err != nil {
				t.Fatal(err)
			}
			view.Frame()
			if path != "/bilibili/download" || len(s.stagedKeys) != 1 {
				t.Fatalf("selection did not open download page: path=%s queued=%d", path, len(s.stagedKeys))
			}
		})
	}
}

func TestLibraryAuthorLinkDoesNotSelectCard(t *testing.T) {
	for _, tab := range []string{"favorite", "collection"} {
		t.Run(tab, func(t *testing.T) {
			s, _, runtime := setup(t, true)
			path := "/bilibili/" + tab
			view := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				s.Layout(c, path, func(next string) { path = next }, func() {
					s.Form(c, path, func(next string) { path = next }, func() {})
				})
			}, 1200, 880)
			view.Frame()
			runtime.flush(t)
			view.Frame()
			name, want := "音乐 UP 主", "/bilibili/up/42"
			if tab == "collection" {
				name, want = "合集 UP 主", "/bilibili/up/43"
			}
			if err := view.Click("进入 UP 主主页：" + name); err != nil {
				t.Fatal(err)
			}
			if path != want || s.selectedCount() != 0 {
				t.Fatalf("author link selected the card or missed route: path=%s selected=%d", path, s.selectedCount())
			}
		})
	}
}

func TestLibraryAuthorHoverLeavesCardNeutral(t *testing.T) {
	s, _, runtime := setup(t, true)
	path := "/bilibili/favorite"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(next string) { path = next }, func() {
			s.Form(c, path, func(next string) { path = next }, func() {})
		})
	}, 1200, 880)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	card, cardOK := view.Find("选择视频：收藏视频 1-1")
	link, linkOK := view.Find("进入 UP 主主页：音乐 UP 主")
	if !cardOK || !linkOK {
		t.Fatal("missing video card or author link")
	}
	borderX, borderY := int(card.X+1), int(card.Y+card.H/2)
	view.Move(link.X+link.W/2, link.Y+link.H/2)
	view.Frame()
	linkHoverBorder := view.Image().RGBAAt(borderX, borderY)
	saveImage(t, view, "/tmp/bilibili-library-author-hover.png")
	view.Move(card.X+card.W/2, card.Y+card.H/3)
	view.Frame()
	cardHoverBorder := view.Image().RGBAAt(borderX, borderY)
	if linkHoverBorder == cardHoverBorder {
		t.Fatal("author hover still triggers the card highlight")
	}
}

func saveImage(t *testing.T, test *ui.Tester, path string) {
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

func TestLoggedOutCardCenteredInContent(t *testing.T) {
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		for _, header := range []float32{38} {
			s := NewState()
			view := ui.NewTester(func(c *ui.Context) {
				core.Use(c, core.Settings{})
				ui.Column(c).Fill().Children(func() {
					if header > 0 {
						ui.Box(c).Height(header).Shrink(0)
					}
					s.Layout(c, "/bilibili/profile", func(string) {}, func() { s.Form(c, "/bilibili/profile", func(string) {}, func() {}) })
				})
			}, size[0], size[1])
			view.Frame()
			view.Frame()
			pane, ok := view.Find("哔哩哔哩内容")
			if !ok {
				t.Fatal("missing content")
			}
			card, ok := view.Find("账号登录")
			if !ok {
				t.Fatal("missing login card")
			}
			if abs(card.X+card.W/2-pane.X-pane.W/2) > 1 || abs(card.Y+card.H/2-pane.Y-pane.H/2) > 1 {
				t.Fatalf("login card not centered at %v header=%v: card=%+v pane=%+v", size, header, card, pane)
			}
			saveImage(t, view, fmt.Sprintf("/tmp/bilibili-qr-centered-%d-%d.png", size[0], int(header)))
		}
	}
}
func TestLoggedInCardCenteredInContent(t *testing.T) {
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		s, _, runtime := setup(t, true)
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			ui.Column(c).Fill().Children(func() {
				ui.Box(c).Height(style.TitlebarHeight).Shrink(0)
				s.Layout(c, "/bilibili/profile", func(string) {}, func() { s.Form(c, "/bilibili/profile", func(string) {}, func() {}) })
			})
		}, size[0], size[1])
		view.Frame()
		runtime.flush(t)
		view.Frame()
		pane, ok := view.Find("哔哩哔哩内容")
		if !ok {
			t.Fatal("missing content")
		}
		card, ok := view.Find("账号资料")
		if !ok || abs(card.X+card.W/2-pane.X-pane.W/2) > 1 || abs(card.Y+card.H/2-pane.Y-pane.H/2) > 1 {
			t.Fatalf("account card not centered at %v: card=%+v pane=%+v", size, card, pane)
		}
		for _, label := range []string{"当前登录账号", "原生测试账号", "签名", "会员状态", "UID", "退出登录"} {
			if !view.HasText(label) {
				t.Fatalf("missing account detail %q", label)
			}
		}
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-account-centered-%d.png", size[0]))
	}
}
func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func TestAuthorCardHoverClickAndKeyboard(t *testing.T) {
	s, _, runtime := setup(t, true)
	path := "/bilibili/up"
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" }) })
	}, 1200, 880)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	label := "进入测试 UP 主 1的主页"
	card, ok := view.Find(label)
	if !ok {
		t.Fatal("whole author card is not an accessible control")
	}
	if view.HasText("查看投稿") {
		t.Fatal("old action button remains")
	}
	if view.HasText("进入主页") {
		t.Fatal("hover hint should be hidden initially")
	}
	view.Move(card.X+card.W/2, card.Y+10)
	view.Frame()
	if !view.HasText("进入主页") {
		t.Fatal("missing hover hint")
	}
	hovered, _ := view.Find(label)
	if card != hovered {
		t.Fatal("hover moved/resized the card")
	}
	saveImage(t, view, "/tmp/bilibili-author-hover.png")
	view.ClickAt(card.X+10, card.Y+10)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	if !s.authorInfo.Loaded || s.authorInfo.Data.Mid != 1 {
		t.Fatal("clicking the card background did not enter the author page")
	}
	path = "/bilibili/up"
	view.Frame()
	// Reach the same card through the normal tab order, then activate it.
	for i := 0; i < 30 && !view.Focused(label); i++ {
		view.Key(0, ui.KeyTab)
		view.Frame()
	}
	if !view.Focused(label) {
		t.Fatal("author card missing from tab order")
	}
	view.Key(0, ui.KeyEnter)
	view.Frame()
	runtime.flush(t)
	view.Frame()
	if !s.authorInfo.Loaded || s.authorInfo.Data.Mid != 1 {
		t.Fatal("keyboard activation did not enter the author page")
	}
}

func TestAuthorCardHintStaysInsideCompactGrid(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, _, runtime := setup(t, true)
		path := "/bilibili/up"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(p string) { path = p }, func() {
				s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
			})
		}, width, 880)
		view.Frame()
		runtime.flush(t)
		view.Frame()
		card, ok := view.Find("进入测试 UP 主 1的主页")
		if !ok || card.W > 215 {
			t.Fatalf("author card is not compact at %d: %+v", width, card)
		}
		view.Move(card.X+card.W/2, card.Y+10)
		view.Frame()
		hint, ok := view.Find("进入主页")
		if !ok || hint.X < card.X || hint.X+hint.W > card.X+card.W || hint.Y < card.Y || hint.Y+hint.H > card.Y+card.H {
			t.Fatalf("author hover hint escapes card at %d: card=%+v hint=%+v", width, card, hint)
		}
	}
}

func TestAuthorDetailListsAndSelection(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, _, runtime := setup(t, true)
		path := "/bilibili/up/1"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(p string) { path = p }, func() {
				s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
			})
		}, width, 880)
		frame := func() { view.Frame(); runtime.flush(t); view.Frame() }
		click := func(label string) {
			t.Helper()
			if err := view.Click(label); err != nil {
				t.Fatal(err)
			}
			frame()
		}
		frame()
		if !view.HasText("投稿第 1 页") {
			t.Fatal("author uploads missing")
		}
		controls, ok := view.Find("UP 主内容操作")
		if !ok {
			t.Fatal("missing shared author controls")
		}
		tab, ok := view.Find("投稿")
		if !ok {
			t.Fatal("missing upload tab")
		}
		if view.HasText("全部投稿") {
			t.Fatal("redundant upload heading takes space from video cards")
		}
		if view.HasText("添加全部") {
			t.Fatal("redundant add-all control is still visible")
		}
		for _, label := range []string{"合集 / 系列", "已加载 1/2", "全选", "下载已选 (0)", "一键下载全部"} {
			item, found := view.Find(label)
			if !found || item.X < controls.X || item.X+item.W > controls.X+controls.W || item.Y < controls.Y || item.Y+item.H > controls.Y+controls.H {
				t.Fatalf("%s escaped shared author controls at %d: item=%+v controls=%+v", label, width, item, controls)
			}
			if label != "一键下载全部" && abs(item.Y+item.H/2-tab.Y-tab.H/2) > 3 {
				t.Fatalf("%s must share the tab row at %d: item=%+v tab=%+v", label, width, item, tab)
			}
		}
		click("合集 / 系列")
		if !view.HasText("合集投稿") {
			t.Fatal("first author collection was not selected")
		}
		videos, ok := view.Find("UP 主视频区域")
		if !ok {
			t.Fatal("missing author content scroll area")
		}
		listPanel, ok := view.Find("合集与系列")
		if !ok || abs(listPanel.X-videos.X) > 1 || listPanel.H < videos.H-1 {
			t.Fatalf("list pane should align left and fill the viewport at %d: pane=%+v scroll=%+v", width, listPanel, videos)
		}
		season, ok := view.Find("合集 · 作者合集")
		if !ok || season.H < style.AuthorListItemHeight-1 || season.X < listPanel.X || season.X+season.W > listPanel.X+listPanel.W {
			t.Fatalf("collection item is not a full-width card at %d: item=%+v pane=%+v", width, season, listPanel)
		}
		click("系列 · 作者系列")
		if !view.HasText("系列投稿") {
			t.Fatal("author series was not opened")
		}
		click("选择视频：系列投稿")
		if !view.HasText("下载已选 (1)") || len(s.stagedKeys) != 0 {
			t.Fatal("selected series video should be ready for direct download")
		}
		click("投稿")
		click("合集 / 系列")
		if !view.HasText("系列投稿") {
			t.Fatal("author list selection was lost after switching tabs")
		}
		view.Advance(300 * time.Millisecond)
		view.Frame()
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-author-lists-%d.png", width))
		click("返回")
		if path != "/bilibili/up" {
			t.Fatal("detail return button ignored")
		}
	}
}

func TestAuthorDetailHeaderStaysFixedWhenVideosScroll(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, _, runtime := setup(t, true)
		path := "/bilibili/up/1"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			style.Apply(c)
			s.Layout(c, path, func(p string) { path = p }, func() {
				s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
			})
		}, width, 600)
		view.Frame()
		runtime.flush(t)
		for i := 0; i < 40; i++ {
			s.uploads.Data.List.Vlist = append(s.uploads.Data.List.Vlist, model.VListItem{
				Bvid: fmt.Sprintf("BVscroll%02d", i), Title: fmt.Sprintf("滚动测试视频 %02d", i), Length: "03:00",
			})
		}
		s.uploads.Data.Page.Count = len(s.uploads.Data.List.Vlist)
		view.Frame()
		profile, ok := view.Find("UP 主信息")
		if !ok || profile.H > 64 {
			t.Fatalf("UP profile card is too tall at %d: %+v", width, profile)
		}
		controls, ok := view.Find("UP 主内容操作")
		if !ok {
			t.Fatal("missing fixed operation row")
		}
		videos, ok := view.Find("UP 主视频区域")
		if !ok || videos.Y < controls.Y+controls.H-1 {
			t.Fatalf("video scroll region is not below fixed rows at %d: %+v %+v", width, videos, controls)
		}
		leftmost, rightmost := float32(width), float32(0)
		for i := 0; i < 40; i++ {
			if card, found := view.Find(fmt.Sprintf("选择视频：滚动测试视频 %02d", i)); found {
				leftmost = min(leftmost, card.X)
				rightmost = max(rightmost, card.X+card.W)
			}
		}
		if leftmost < videos.X+style.ScrollbarGutter-1 || rightmost <= videos.X+videos.W/2 || rightmost > videos.X+videos.W-style.ScrollbarGutter+1 {
			t.Fatalf("video grid padding is uneven at %d: leftmost=%v rightmost=%v region=%+v", width, leftmost, rightmost, videos)
		}
		scroll := s.scroll("author-videos/1")
		if scroll.MaxY <= 0 {
			t.Fatal("long video grid is not independently scrollable")
		}
		view.Scroll(videos.X+videos.W/2, videos.Y+videos.H/2, 0, 200)
		view.Frame()
		profileAfter, _ := view.Find("UP 主信息")
		controlsAfter, _ := view.Find("UP 主内容操作")
		if scroll.Y <= 0 || abs(profileAfter.Y-profile.Y) > 1 || abs(controlsAfter.Y-controls.Y) > 1 {
			t.Fatalf("fixed rows moved while scrolling at %d: scroll=%v profile=%+v/%+v controls=%+v/%+v", width, scroll.Y, profile, profileAfter, controls, controlsAfter)
		}
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-author-fixed-scroll-%d.png", width))
	}
}

func TestVideoCardSelectionAndCompactLayout(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, _, runtime := setup(t, true)
		path := "/bilibili/up/1"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(p string) { path = p }, func() {
				s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
			})
		}, width, 880)
		view.Frame()
		runtime.flush(t)
		view.Frame()
		label := "选择视频：投稿第 1 页"
		card, ok := view.Find(label)
		if !ok || card.W < 100 || card.H < 100 {
			t.Fatalf("missing video card at %d: %+v", width, card)
		}
		_, checkbox := view.Find("选择视频")
		_, add := view.Find("添加到下载")
		_, open := view.Find("打开")
		if checkbox || add || open {
			t.Fatal("old checkbox or card actions are still visible")
		}
		view.Move(card.X+card.W/2, card.Y+card.H/2)
		view.Frame()
		hover, ok := view.Find(label)
		if !ok || abs(card.W-hover.W) > 1 || abs(card.H-hover.H) > 1 {
			t.Fatalf("hover changed card layout at %d: before=%+v after=%+v", width, card, hover)
		}
		if view.Cursor() != ui.CursorPointer {
			t.Fatal("card hover does not show pointer cursor")
		}
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-video-hover-%d.png", width))
		if err := view.Click(label); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if !*s.selection("BVupload1-1") || !view.HasText("下载已选 (1)") {
			t.Fatal("click did not select card and update batch action")
		}
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-video-selected-%d.png", width))
		if err := view.Click(label); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if *s.selection("BVupload1-1") || !view.HasText("下载已选 (0)") {
			t.Fatalf("second click did not clear selection: selected=%v texts=%v", *s.selection("BVupload1-1"), view.Texts())
		}
		if err := view.Click(label); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if !view.HasText("下载已选 (1)") || len(s.stagedKeys) != 0 {
			t.Fatal("selected video should be ready for direct download")
		}
	}
}

func TestDownloadQueueStillRemovesIndividualCard(t *testing.T) {
	s := NewState()
	s.stageUpload(model.VListItem{Bvid: "BVqueued", Title: "待下载测试视频", Length: "02:00"})
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.Layout(c, "/bilibili/download", func(string) {}, func() {
			s.Form(c, "/bilibili/download", func(string) {}, func() {})
		})
	}, 760, 600)
	view.Frame()
	if !view.HasText("待下载测试视频") {
		t.Fatal("queued video missing")
	}
	if err := view.Click("移除"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if len(s.stagedKeys) != 0 || view.HasText("待下载测试视频") {
		t.Fatal("individual queued video was not removed")
	}
}

func TestVideoCoverTouchesCardFrame(t *testing.T) {
	s := NewState()
	cover := image.NewRGBA(image.Rect(0, 0, 160, 90))
	want := color.RGBA{R: 220, G: 40, B: 40, A: 255}
	draw.Draw(cover, cover.Bounds(), &image.Uniform{C: want}, image.Point{}, draw.Src)
	s.images["test://cover"] = &imagecache.Entry{Bitmap: ui.NewBitmap(cover)}
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Padding(20).Width(240).Children(func() {
			s.mediaCard(c, "BVtest", "封面贴边测试", "test://cover", "UP 主", "1:00", 1142, 3, 0)
		})
	}, 300, 280)
	view.Frame()
	card, ok := view.Find("选择视频：封面贴边测试")
	if !ok {
		t.Fatal("video card missing")
	}
	// The painted two-pixel outline overlays the cover; no surface-colored
	// inset should appear between the outline and the image.
	got := view.Image().RGBAAt(int(card.X)+3, int(card.Y)+12)
	if got.R < 180 || got.G > 80 || got.B > 80 {
		t.Fatalf("cover does not meet the card outline: got %+v", got)
	}
	overlayCenter := view.Image().RGBAAt(int(card.X+card.W/2), int(card.Y+card.W/style.MediaAspectRatio)-8)
	if overlayCenter.R < 180 || overlayCenter.G > 80 || overlayCenter.B > 80 {
		t.Fatalf("statistics overlay still darkens the cover: got %+v", overlayCenter)
	}
	for _, x := range []int{int(card.X), int(card.X + card.W - 1)} {
		corner := view.Image().RGBAAt(x, int(card.Y))
		if corner.R > 180 && corner.G < 80 && corner.B < 80 {
			t.Fatalf("cover leaks into rounded top corner at x=%d: %+v", x, corner)
		}
	}
	if view.HasText("播放") || view.HasText("弹幕") || !view.HasText("1142") || !view.HasText("3") {
		t.Fatal("video statistics should use icons with compact counts")
	}
	saveImage(t, view, "/tmp/bilibili-video-icon-corners.png")
}

func TestPortraitCoverFillsVideoCard(t *testing.T) {
	s := NewState()
	cover := image.NewRGBA(image.Rect(0, 0, 60, 160))
	draw.Draw(cover, cover.Bounds(), &image.Uniform{C: color.RGBA{R: 220, G: 40, B: 40, A: 255}}, image.Point{}, draw.Src)
	s.images["test://portrait"] = &imagecache.Entry{Bitmap: ui.NewBitmap(cover)}
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Padding(20).Width(240).Children(func() {
			s.mediaCard(c, "BVportrait", "竖版封面", "test://portrait", "UP 主", "1:00", 0, 0, 0)
		})
	}, 300, 280)
	view.Frame()
	card, ok := view.Find("选择视频：竖版封面")
	if !ok {
		t.Fatal("portrait card missing")
	}
	for _, x := range []int{int(card.X) + 3, int(card.X+card.W) - 4} {
		pixel := view.Image().RGBAAt(x, int(card.Y)+12)
		if pixel.R < 180 || pixel.G > 80 || pixel.B > 80 {
			t.Fatalf("portrait cover left a horizontal gap at x=%d: %+v", x, pixel)
		}
	}
}

func TestAuthorWithoutListsUsesFullWidthEmptyState(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s, api, runtime := setup(t, true)
		api.emptyAuthorLists = true
		path := "/bilibili/up/1"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(p string) { path = p }, func() {
				s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" })
			})
		}, width, 880)
		view.Frame()
		runtime.flush(t)
		view.Frame()
		if err := view.Click("合集 / 系列"); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if !view.HasText("正在加载合集与系列") || view.HasText("列表") {
			t.Fatal("empty author lists showed a sidebar during loading")
		}
		runtime.flush(t)
		view.Frame()
		if !view.HasText("暂无合集或系列") || view.HasText("列表") || view.HasText("选择一个合集或系列") || view.HasText("全选") {
			t.Fatal("empty author lists retained split layout or disabled actions")
		}
		pane, ok := view.Find("哔哩哔哩内容")
		if !ok {
			t.Fatal("missing content pane")
		}
		emptyPanel, ok := view.Find("合集与系列空状态")
		if !ok || emptyPanel.W < pane.W*.8 {
			t.Fatalf("empty state did not span the pane at %d: pane=%+v empty=%+v", width, pane, emptyPanel)
		}
		view.Advance(300 * time.Millisecond)
		view.Frame()
		saveImage(t, view, fmt.Sprintf("/tmp/bilibili-author-empty-lists-%d.png", width))
		if err := view.Click("查看投稿"); err != nil {
			t.Fatal(err)
		}
		view.Frame()
		if !view.HasText("投稿第 1 页") {
			t.Fatal("empty state could not return to uploads")
		}
	}
}
