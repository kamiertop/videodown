package douyin

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	dyDownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
)

type testAPI struct {
	API
	cookie      string
	detailCalls int
	fail        bool
}

func (a *testAPI) GetCookie() (string, error) { return a.cookie, nil }
func (a *testAPI) SetCookie(v string) error   { a.cookie = v; return nil }
func (a *testAPI) Profile() (model.MyInfoResponse, error) {
	return model.MyInfoResponse{User: model.User{Uid: "1", SecUid: "MS4wLjAB_me", NickName: "原生抖音账号", Signature: "Go model 直接展示"}}, nil
}
func (a *testAPI) FollowList(cursor int) (model.FollowResponse, error) {
	return model.FollowResponse{Total: 2, Followings: []model.FollowItem{{SecUid: "MS4wLjAB_author", Nickname: "测试作者", Signature: "作者简介"}}}, nil
}
func (a *testAPI) ParseVideo(v string) (string, error)     { return "123", nil }
func (a *testAPI) ParseSecUserID(v string) (string, error) { return "MS4wLjAB_author", nil }
func (a *testAPI) VideoDetail(id string) (model.AwemeItem, error) {
	a.detailCalls++
	return fixture(id), nil
}
func (a *testAPI) User(uid string) (model.UserResponse, error) {
	return model.UserResponse{User: model.User{SecUid: uid, NickName: "测试作者", Signature: "作者简介"}}, nil
}
func (a *testAPI) UserVideoList(uid string, size, cursor int) (model.UserVideoListResponse, error) {
	return model.UserVideoListResponse{AwemeList: []model.AwemeItem{fixture(fmt.Sprint(cursor + 1))}, MaxCursor: int64(cursor + 1), HasMore: boolInt(cursor < 1)}, nil
}
func (a *testAPI) UserSeries(uid string, cursor, size int) (model.UserSeriesResponse, error) {
	return model.UserSeriesResponse{SeriesInfos: []model.SeriesInfoItem{{SeriesID: "42", SeriesName: "作者合集"}}}, nil
}
func (a *testAPI) FavoriteVideo(size, cursor uint) (model.FavoriteVideoResponse, error) {
	if a.fail {
		return model.FavoriteVideoResponse{}, errors.New("分页失败")
	}
	return model.FavoriteVideoResponse{AwemeList: []model.AwemeItem{fixture(fmt.Sprint(cursor + 1))}, Cursor: cursor + 1, HasMore: boolInt(cursor < 1)}, nil
}
func (a *testAPI) CollectList(cursor int) (model.CollectListResponse, error) {
	return model.CollectListResponse{CollectsList: []model.CollectsList{{CollectsIDStr: "7", CollectsName: "收藏夹示例", TotalNumber: 2}}}, nil
}
func (a *testAPI) FavoritesVideoList(id string, cursor, size int) (model.FavoriteVideoResponse, error) {
	return a.FavoriteVideo(uint(size), uint(cursor))
}
func (a *testAPI) Collection(size, cursor int) (model.CollectionResponse, error) {
	return model.CollectionResponse{MixInfos: []model.CollectionItem{{MixId: "42", MixName: "收藏的合集", Author: model.Author{SecUid: "MS4wLjAB_author"}}}}, nil
}
func (a *testAPI) CollectionList(uid, id string, cursor, size int) (model.CollectionListResponse, error) {
	return model.CollectionListResponse{AwemeList: []model.AwemeItem{fixture("collection-video")}}, nil
}
func (a *testAPI) Dynamic(cursor int) (model.DynamicResponse, error) {
	return model.DynamicResponse{}, nil
}
func (a *testAPI) History(cursor, t1, t2, t3 int) (model.HistoryResponse, error) {
	return model.HistoryResponse{AwemeList: []model.HistoryItem{{AwemeId: "history", Desc: "播放历史作品", Author: model.Author{Nickname: "测试作者"}, Video: fixture("1").Video}}}, nil
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func fixture(id string) model.AwemeItem {
	return model.AwemeItem{AwemeId: id, Desc: "原生抖音作品 " + id, Duration: 93000, Author: model.Author{SecUid: "MS4wLjAB_author", Nickname: "测试作者"}, Video: model.Video{PlayAddr: model.PlayInfo{UrlList: []string{"https://fixture.test/" + id + ".mp4"}}}}
}

type job struct {
	work func() error
	done func(error)
}
type testRuntime struct {
	jobs   []job
	timers []func()
}

func (r *testRuntime) runtime() guiruntime.Runtime {
	run := func(w func() error, d func(error)) { r.jobs = append(r.jobs, job{w, d}) }
	return guiruntime.Runtime{Run: run, Background: run, After: func(_ time.Duration, f func()) func() { r.timers = append(r.timers, f); return func() {} }}
}
func (r *testRuntime) finish(t *testing.T, index int) {
	t.Helper()
	if index >= len(r.jobs) {
		t.Fatal("missing job")
	}
	j := r.jobs[index]
	r.jobs = append(r.jobs[:index], r.jobs[index+1:]...)
	j.done(j.work())
}
func (r *testRuntime) flush(t *testing.T) {
	t.Helper()
	for n := 0; len(r.jobs) > 0; n++ {
		if n > 100 {
			t.Fatal("unbounded work")
		}
		r.finish(t, 0)
	}
}

type testBackend struct {
	DownloadBackend
	tasks   []dyDownload.Task
	limit   int
	fail    string
	removed string
	cleared bool
}

func (d *testBackend) DownloadConcurrency() int { return d.limit }
func (d *testBackend) DownloadVideos(tasks []dyDownload.Task) (dyDownload.BatchResult, error) {
	d.tasks = append(d.tasks, tasks...)
	if tasks[0].AwemeID == d.fail {
		return dyDownload.BatchResult{Failed: 1}, errors.New("下载失败")
	}
	return dyDownload.BatchResult{Success: len(tasks)}, nil
}
func (d *testBackend) DownloadHistoryPage(offset, size int, q string) (dyDownload.HistoryPage, error) {
	return dyDownload.HistoryPage{Total: 1, Items: []dyDownload.HistoryItem{{AwemeID: "local", Title: "本地历史", Path: "/tmp/local.mp4"}}}, nil
}
func (d *testBackend) DeleteDownloadHistory(id string) error { d.removed = id; return nil }
func (d *testBackend) ClearDownloadHistory() error           { d.cleared = true; return nil }
func setup(t *testing.T) (*State, *testAPI, *testBackend, *testRuntime) {
	t.Helper()
	s := NewState()
	a := &testAPI{cookie: "fixture=test"}
	b := &testBackend{limit: 3}
	r := &testRuntime{}
	s.Configure(a, b, r.runtime())
	s.Start()
	r.flush(t)
	t.Cleanup(s.Stop)
	return s, a, b, r
}

func TestAccountSelectionPaginationAndRoutes(t *testing.T) {
	s, a, _, r := setup(t)
	if !s.loggedIn() {
		t.Fatal("profile unavailable")
	}
	s.enter("/douyin/favorite")
	r.flush(t)
	s.selected["1"] = true
	s.loadVideos(true)
	r.flush(t)
	if len(s.videos.Data.Items) != 2 || !s.selected["1"] || s.videos.Data.More {
		t.Fatal("pagination lost selection")
	}
	var path string
	s.stageSelected(func(p string) { path = p })
	if path != "/douyin/download" || len(s.staged) != 1 || s.PendingDownloadCount() != 0 {
		t.Fatal("selection should navigate without downloading")
	}
	s.enter("/douyin/favorite/folder/7")
	r.flush(t)
	if s.source.Key != "folder:7" || s.source.Title != "收藏夹示例" {
		t.Fatalf("folder route not restored: %+v", s.source)
	}
	s.enter("/douyin/favorite/collection/42")
	r.flush(t)
	if s.source.Key != "collection:42" {
		t.Fatal("collection route not restored")
	}
	s.enter("/douyin/user/MS4wLjAB_author")
	s.enter("/douyin/play-history")
	r.flush(t)
	if s.source.Key != "play-history" || s.videos.Data.Items[0].AwemeId != "history" {
		t.Fatal("stale author response replaced history")
	}
	s.cookie = ""
	s.saveCookie(true)
	r.flush(t)
	if s.loggedIn() || a.cookie != "" || len(s.staged) != 0 {
		t.Fatal("logout retained private data")
	}
}
func TestAutoLoadingErrorsAndObsoleteSources(t *testing.T) {
	s, a, _, r := setup(t)
	s.enter("/douyin/favorite")
	r.flush(t)
	s.scroll("videos").MaxY = 500
	s.scroll("videos").Y = 450
	s.maybeMore()
	if len(r.timers) != 1 {
		t.Fatal("near-bottom did not schedule")
	}
	a.fail = true
	r.timers[0]()
	r.flush(t)
	if s.more.Error == "" || len(s.videos.Data.Items) != 1 {
		t.Fatal("paging failure should retain cards")
	}
	n := len(r.timers)
	s.maybeMore()
	if len(r.timers) != n {
		t.Fatal("automatic retry loop")
	}
	a.fail = false
	s.loadVideos(true)
	r.flush(t)
	if len(s.videos.Data.Items) != 2 {
		t.Fatal("retry failed")
	}
	s.useSource(s.baseSource("favorite"))
	s.enter("/douyin/play-history")
	r.flush(t)
	if s.videos.Data.Items[0].AwemeId != "history" {
		t.Fatal("obsolete source result accepted")
	}
}
func TestDownloadGroupConcurrencyStopRetryAndManualQueue(t *testing.T) {
	s, a, b, r := setup(t)
	s.source = source{Key: "author", Title: "作者全部作品", Load: func(cursor int) (videoPage, error) {
		items := []model.AwemeItem{}
		for i := 1; i <= 5; i++ {
			items = append(items, fixture(fmt.Sprint(i)))
		}
		return videoPage{Items: items}, nil
	}}
	s.requestAll(func(string) {})
	s.requestAll(func(string) {})
	if len(s.groups) != 1 {
		t.Fatal("duplicate batch")
	}
	r.finish(t, 0)
	g := s.groups[0]
	if g.InFlight != 3 || len(r.jobs) != 3 {
		t.Fatalf("configured concurrency ignored: %d / %d", g.InFlight, len(r.jobs))
	}
	s.stage(fixture("manual"))
	s.request(s.staged)
	if len(r.jobs) != 3 || s.groups[1].Status != "queued" {
		t.Fatal("manual task should wait")
	}
	s.stopGroup(g)
	r.flush(t)
	if g.Done != 3 || g.Status != "stopped" || len(b.tasks) != 4 {
		t.Fatal("stop should drain active tasks then run next group")
	}
	b.fail = "4"
	s.resume(g)
	r.flush(t)
	if g.Failed != 1 || g.Done != 4 {
		t.Fatalf("failed task state: %+v", g)
	}
	b.fail = ""
	s.resume(g)
	r.flush(t)
	if g.Done != 5 || g.Failed != 0 || a.detailCalls != 0 {
		t.Fatal("retry repeated completed work or reparsed fresh URLs")
	}
	for _, tick := range r.timers {
		tick()
	}
	if len(s.groups) != 0 {
		t.Fatal("successful group not dismissed")
	}
}
func TestTaskUsesModelGalleryAndExpiredDetail(t *testing.T) {
	v := fixture("gallery")
	v.Images = []model.ImageItem{{UrlList: []string{"https://fixture.test/1.jpg"}, Video: model.ImageVideo{PlayAddr: model.PlayInfo{UrlList: []string{"https://fixture.test/live.mp4"}}}}}
	task, err := taskFor(v, -1, "合集")
	if err != nil || len(task.Assets) != 2 || task.Assets[1].Kind != "video" || task.SourceName != "合集" {
		t.Fatalf("live photo assets: %+v %v", task, err)
	}
	s, a, b, r := setup(t)
	s.stage(fixture("expired"))
	s.staged[0].ResolvedAt = time.Now().Add(-time.Hour)
	s.request(s.staged)
	r.flush(t)
	if a.detailCalls != 1 || len(b.tasks) != 1 {
		t.Fatal("expired URL should resolve once when task starts")
	}
}
func TestHistoryLocalOperationsWithoutLogin(t *testing.T) {
	s, _, b, r := setup(t)
	s.profile = resource[model.User]{}
	s.enter("/douyin/history")
	r.flush(t)
	if s.history.Data.Total != 1 {
		t.Fatal("local history required login")
	}
	s.changeHistory(func() error { return b.DeleteDownloadHistory("local") })
	r.flush(t)
	if b.removed != "local" {
		t.Fatal("delete not forwarded")
	}
	s.changeHistory(b.ClearDownloadHistory)
	r.flush(t)
	if !b.cleared {
		t.Fatal("clear not forwarded")
	}
}
func TestNativePagesInteractionAndRendering(t *testing.T) {
	s, _, _, r := setup(t)
	path := "/douyin/favorite"
	navigate := func(p string) { path = p }
	view := func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() {
			s.Layout(c, path, navigate, func() { s.Form(c, path, navigate, func() { path = "/douyin/user" }) })
		})
	}
	test := ui.NewTester(view, 1200, 800)
	for _, size := range [][2]int{{1200, 800}, {760, 600}} {
		test.SetSize(size[0], size[1])
		for _, p := range []string{"/douyin/profile", "/douyin/user", "/douyin/user/MS4wLjAB_author", "/douyin/favorite/folder/7", "/douyin/favorite/collection/42", "/douyin/dynamic", "/douyin/play-history", "/douyin/download", "/douyin/history"} {
			path = p
			test.Frame()
			r.flush(t)
			test.Frame()
			test.Advance(300 * time.Millisecond)
			test.Frame()
		}
	}
	path = "/douyin/favorite"
	test.SetSize(1200, 800)
	test.Frame()
	r.flush(t)
	test.Frame()
	if err := test.Click("选择抖音作品：原生抖音作品 1"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if !s.selected["1"] {
		t.Fatal("click did not select card")
	}
	if err := test.Click("下载已选 (1)"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if path != "/douyin/download" || len(s.staged) != 1 {
		t.Fatal("selected video did not reach download page")
	}
	r.flush(t)
	test.Frame()
	test.Advance(time.Second)
	file, err := os.Create("/tmp/douyin-native-download.png")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, test.Image()); err != nil {
		t.Fatal(err)
	}
	path = "/douyin/favorite/folder/7"
	test.Frame()
	r.flush(t)
	test.Frame()
	test.Advance(time.Second)
	file2, err := os.Create("/tmp/douyin-native-library.png")
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()
	png.Encode(file2, test.Image())
}

func TestAuthorLinkDoesNotSelectVideo(t *testing.T) {
	s, _, _, r := setup(t)
	path := "/douyin/favorite"
	navigate := func(p string) { path = p }
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() {
			s.Layout(c, path, navigate, func() { s.Form(c, path, navigate, func() { path = "/douyin/user" }) })
		})
	}, 1200, 800)
	test.Frame()
	r.flush(t)
	test.Frame()
	if err := test.Click("进入抖音作者主页：测试作者"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	r.flush(t)
	test.Frame()
	if path != "/douyin/user/MS4wLjAB_author" || s.selected["1"] {
		t.Fatal("author link selected the containing card")
	}
}

func TestAccountChangeAndCloseRejectOldResults(t *testing.T) {
	s, _, _, r := setup(t)
	s.enter("/douyin/favorite")
	s.cookie = "replacement=fixture"
	s.saveCookie(false)
	r.finish(t, 1) // cookie write finishes before the obsolete list response
	r.flush(t)
	if len(s.videos.Data.Items) != 1 || !s.loggedIn() {
		t.Fatal("account refresh failed")
	}
	s.link = "fixture"
	s.parse()
	s.Stop()
	r.flush(t)
	if len(s.staged) > 0 {
		t.Fatal("closed state accepted a parse result")
	}
}

func TestLoggedOutPromptKeepsMessageAndActionTogether(t *testing.T) {
	s := NewState()
	s.profile.Error = "Cookie未设置，请先设置Cookie"
	path := "/douyin/user"
	navigate := func(next string) { path = next }
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() { s.Layout(c, path, navigate, func() { s.Form(c, path, navigate, func() {}) }) })
	}, 1200, 800)
	for _, size := range [][2]int{{1200, 800}, {760, 600}} {
		test.SetSize(size[0], size[1])
		for _, route := range []string{"/douyin/user", "/douyin/dynamic", "/douyin/favorite", "/douyin/play-history"} {
			path = route
			test.Frame()
			test.Advance(300 * time.Millisecond)
			card, ok := test.Find("登录提示")
			if !ok {
				t.Fatal("missing login prompt")
			}
			action, ok := test.Find("前往账号")
			if !ok {
				t.Fatal("missing account action")
			}
			if card.H > 160 || action.W > 160 || action.Y < card.Y || action.Y+action.H > card.Y+card.H {
				t.Fatalf("login action separated from compact prompt at %v: card=%+v action=%+v", size, card, action)
			}
			pane, ok := test.Find("抖音内容")
			if !ok {
				t.Fatal("missing content pane")
			}
			dx := card.X + card.W/2 - (pane.X + pane.W/2)
			dy := card.Y + card.H/2 - (pane.Y + pane.H/2)
			if dx < -1 || dx > 1 || dy < -1 || dy > 1 || card.W > 360 {
				t.Fatalf("login prompt not centered at %v: card=%+v pane=%+v", size, card, pane)
			}
			if test.HasText(s.profile.Error) {
				t.Fatal("duplicate account error on protected page")
			}
		}
	}
	test.SetSize(1200, 800)
	test.Frame()
	f, err := os.Create("/tmp/douyin-login-prompt.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, test.Image()); err != nil {
		t.Fatal(err)
	}
	if err := test.Click("前往账号"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if path != "/douyin/profile" || !test.HasText(s.profile.Error) {
		t.Fatal("account action should open profile with detailed error")
	}
}

func TestSidebarShowsAccountAvatarOnlyWhenLoggedIn(t *testing.T) {
	s, _, _, _ := setup(t)
	s.profile.Data.AvatarThumb.UrlList = []string{"https://fixture.test/avatar.jpg"}
	view := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() {
			s.Layout(c, "/douyin/download", func(string) {}, func() { s.Form(c, "/douyin/download", func(string) {}, func() {}) })
		})
	}, 1200, 800)
	for _, size := range [][2]int{{1200, 800}, {760, 600}} {
		view.SetSize(size[0], size[1])
		view.Frame()
		avatar, ok := view.Find("抖音账号头像")
		if !ok || avatar.H != 32 {
			t.Fatalf("missing avatar at %v: %+v", size, avatar)
		}
		account, ok := view.Find("抖音账号")
		if !ok || avatar.Y+avatar.H > account.Y {
			t.Fatal("avatar must sit above account entry")
		}
	}
	s.profile.Loaded = false
	view.Frame()
	if _, ok := view.Find("抖音账号头像"); ok {
		t.Fatal("signed-out account retained avatar")
	}
}
