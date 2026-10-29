package bilibili

import (
	"errors"
	"fmt"
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	"github.com/kamiertop/videodown/bilibili/model"
	"strconv"
	"testing"
	"time"
)

type fakeAPI struct {
	API
	logged           bool
	codes            []uint32
	favorite         func(int, int) (model.FavoriteData, error)
	collection       func(int, int) (model.CollectionItemData, error)
	collectionList   func(int, int) (model.CollectionData, error)
	upload           func(uint, int, int) (model.VideoListData, error)
	season           func(string, int, int, int) (model.SeasonsArchivesData, error)
	series           func(string, int, int, int) (model.SeriesArchivesData, error)
	authorLists      func(string, int, int) (model.SeasonsSeriesData, error)
	dynamic          func(string) (model.DynamicArchivePage, error)
	playHistory      func(int, int) (model.PlayHistoryData, error)
	failFollow       bool
	emptyAuthorLists bool
}

func (f *fakeAPI) IsLoggedIn() bool { return f.logged }
func (f *fakeAPI) RefreshCookie() (model.CookieRefreshData, error) {
	return model.CookieRefreshData{}, nil
}
func (f *fakeAPI) MyInfo() (model.MyInfoProfile, error) {
	return model.MyInfoProfile{Mid: 42, Name: "原生测试账号", Level: 6, Sign: "直接使用 Go model"}, nil
}
func (f *fakeAPI) ResolvePlayUrl(req biliapi.PlayUrlRequest, _ int) biliapi.PlayUrlResult {
	return biliapi.PlayUrlResult{
		Bvid: req.Bvid, Cid: 42,
		Detail:  &model.VideoDetailConciseData{View: model.VideoDetailView{Bvid: req.Bvid, Cid: 42, Title: req.Bvid}},
		PlayURL: &model.VideoURLData{Dash: model.Dash{Video: []model.VideoItem{{Id: 80, BaseUrl: "https://example.com/video", Bandwidth: 100}}}},
	}
}
func (f *fakeAPI) VideoDetailConciseBvid(bvid string) (model.VideoDetailConciseData, error) {
	return model.VideoDetailConciseData{View: model.VideoDetailView{Bvid: bvid, Cid: 42, Title: "手动解析视频", Owner: model.VideoOwner{Name: "测试 UP 主"}}}, nil
}
func (f *fakeAPI) VideoDetailConcise(aid int) (model.VideoDetailConciseData, error) {
	return f.VideoDetailConciseBvid("BVavfixture1")
}
func (f *fakeAPI) QRCode() (model.QRCodeData, error) {
	return model.QRCodeData{Url: "https://example.com/login/test-fixture", QRCodeKey: "fixture"}, nil
}
func (f *fakeAPI) PollQRCode(string) (model.PollQRCodeData, error) {
	code := f.codes[0]
	f.codes = f.codes[1:]
	if code == 0 {
		f.logged = true
	}
	return model.PollQRCodeData{Code: code}, nil
}
func (f *fakeAPI) LogOut() (model.LogOut, error) {
	f.logged = false
	return model.LogOut{}, errors.New("远端退出失败，本地凭证已清除")
}
func (f *fakeAPI) FollowList(page, size int) (model.FollowData, error) {
	if f.failFollow {
		return model.FollowData{}, errors.New("请求失败")
	}
	list := []model.FollowDataList{}
	for i := 0; i < 3; i++ {
		list = append(list, model.FollowDataList{Mid: int64((page-1)*20 + i + 1), Uname: fmt.Sprintf("测试 UP 主 %d", (page-1)*20+i+1), Sign: "签名"})
	}
	return model.FollowData{Total: 21, List: list}, nil
}
func (f *fakeAPI) Info(raw string) (model.UserInfoData, error) {
	id, _ := strconv.ParseUint(raw, 10, 64)
	return model.UserInfoData{Mid: id, Name: "测试 UP 主", Sign: "作者签名"}, nil
}
func (f *fakeAPI) VideoList(mid uint, size, page int) (model.VideoListData, error) {
	if f.upload != nil {
		return f.upload(mid, size, page)
	}
	return model.VideoListData{Page: model.VideoListPage{Count: 2}, List: model.VideoList{Vlist: []model.VListItem{{Bvid: fmt.Sprintf("BVupload%d-%d", mid, page), Title: fmt.Sprintf("投稿第 %d 页", page), Length: "03:00", Author: "测试 UP 主"}}}}, nil
}
func (f *fakeAPI) SeasonsSeriesList(mid string, size, page int) (model.SeasonsSeriesData, error) {
	if f.authorLists != nil {
		return f.authorLists(mid, size, page)
	}
	if f.emptyAuthorLists {
		return model.SeasonsSeriesData{}, nil
	}
	return model.SeasonsSeriesData{ItemsLists: model.SeasonsSeriesItemsLists{Page: model.SeasonsSeriesPage{Total: 2}, SeasonsList: []model.SeasonsItem{{Meta: model.SeasonsItemMeta{SeasonId: 3, Name: "作者合集", Total: 1}}}, SeriesList: []model.SeriesItem{{Meta: model.SeriesItemMeta{SeriesId: 4, Name: "作者系列", Total: 1}}}}}, nil
}
func (f *fakeAPI) SeasonsArchivesList(mid string, size, page, id int) (model.SeasonsArchivesData, error) {
	if f.season != nil {
		return f.season(mid, size, page, id)
	}
	return model.SeasonsArchivesData{Page: model.SeasonsArchivesPage{Total: 1}, Archives: []model.SeasonsArchivesItem{{Bvid: "BVseason", Title: "合集投稿"}}}, nil
}
func (f *fakeAPI) SeriesList(mid string, size, page, id int) (model.SeriesArchivesData, error) {
	if f.series != nil {
		return f.series(mid, size, page, id)
	}
	return model.SeriesArchivesData{Page: model.SeriesListPage{Total: 1}, Archives: []model.SeriesArchivesItem{{Bvid: "BVseries", Title: "系列投稿"}}}, nil
}
func (f *fakeAPI) FavoritesList() (model.FavoritesData, error) {
	return model.FavoritesData{Count: 2, List: []model.FavoriteItem{{ID: 1, Title: "音乐", MediaCount: 2}, {ID: 2, Title: "教程", MediaCount: 2}}}, nil
}
func (f *fakeAPI) Favorites(id, page, size int) (model.FavoriteData, error) {
	if f.favorite != nil {
		return f.favorite(id, page)
	}
	return model.FavoriteData{Info: model.FavoriteInfo{ID: int64(id), Title: fmt.Sprintf("收藏夹 %d", id), MediaCount: 2}, HasMore: page == 1, Medias: []model.FavoriteMedias{{ID: id*10 + page, Type: 2, Bvid: fmt.Sprintf("BVfavorite%d-%d", id, page), Title: fmt.Sprintf("收藏视频 %d-%d", id, page), Upper: model.FavoriteMediaUpper{Mid: 42, Name: "音乐 UP 主"}, Duration: 120}}}, nil
}
func (f *fakeAPI) Collection(page, size int) (model.CollectionData, error) {
	if f.collectionList != nil {
		return f.collectionList(page, size)
	}
	return model.CollectionData{Count: 2, HasMore: page == 1, List: []model.CollectionDataList{{ID: page, Title: fmt.Sprintf("订阅合集 %d", page), MediaCount: 2}}}, nil
}
func (f *fakeAPI) CollectionItem(id string, page, size int) (model.CollectionItemData, error) {
	n, _ := strconv.Atoi(id)
	if f.collection != nil {
		return f.collection(n, page)
	}
	return model.CollectionItemData{Info: model.CollectionItemInfo{ID: n, Title: "订阅合集", MediaCount: 2}, Medias: []model.CollectionItemMedias{{ID: int64(n*10 + page), Bvid: fmt.Sprintf("BVcollection%d-%d", n, page), Title: fmt.Sprintf("合集视频 %d", page), Upper: model.CollectionItemMediaUpper{Mid: 43, Name: "合集 UP 主"}}}}, nil
}
func (f *fakeAPI) DynamicVideosPage(offset string) (model.DynamicArchivePage, error) {
	if f.dynamic != nil {
		return f.dynamic(offset)
	}
	return model.DynamicArchivePage{Items: []model.DynamicArchiveItem{{Bvid: "BVdynamic1", Title: "动态视频", AuthorName: "动态 UP 主", Mid: 42, DurationText: "03:00"}}}, nil
}
func (f *fakeAPI) PlayHistory(cursor, viewAt int) (model.PlayHistoryData, error) {
	if f.playHistory != nil {
		return f.playHistory(cursor, viewAt)
	}
	return model.PlayHistoryData{}, nil
}

type queuedJob struct {
	work func() error
	done func(error)
}
type testTimer struct {
	fn       func()
	canceled bool
}
type fakeRuntime struct {
	jobs   []queuedJob
	timers []*testTimer
}

func (r *fakeRuntime) runtime() Runtime {
	return Runtime{
		Run: func(work func() error, done func(error)) { r.jobs = append(r.jobs, queuedJob{work, done}) },
		After: func(_ time.Duration, fn func()) func() {
			timer := &testTimer{fn: fn}
			r.timers = append(r.timers, timer)
			return func() { timer.canceled = true }
		},
	}
}
func (r *fakeRuntime) finish(t *testing.T, index int) {
	t.Helper()
	if index >= len(r.jobs) {
		t.Fatal("missing queued job")
	}
	job := r.jobs[index]
	r.jobs = append(r.jobs[:index], r.jobs[index+1:]...)
	job.done(job.work())
}
func (r *fakeRuntime) flush(t *testing.T) {
	t.Helper()
	for n := 0; len(r.jobs) > 0; n++ {
		if n > 100 {
			t.Fatal("unbounded work")
		}
		r.finish(t, 0)
	}
}
func (r *fakeRuntime) tick(t *testing.T) {
	t.Helper()
	for len(r.timers) > 0 {
		timer := r.timers[0]
		r.timers = r.timers[1:]
		if !timer.canceled {
			timer.fn()
			return
		}
	}
	t.Fatal("missing active timer")
}
func setup(t *testing.T, logged bool) (*State, *fakeAPI, *fakeRuntime) {
	t.Helper()
	s := NewState()
	api := &fakeAPI{logged: logged}
	runtime := &fakeRuntime{}
	s.Configure(api, runtime.runtime())
	s.Start()
	runtime.flush(t)
	t.Cleanup(s.Stop)
	return s, api, runtime
}

func TestQRCodeLifecycleAndLogout(t *testing.T) {
	s, api, r := setup(t, false)
	api.codes = []uint32{model.PollQRCodeStatusScannedUnconfirmed, model.PollQRCodeStatusSuccess}
	s.enter("profile")
	r.flush(t)
	if s.qrImage == nil || !s.qr.Loaded {
		t.Fatal("QR bitmap not generated")
	}
	r.tick(t)
	r.flush(t)
	if s.qrStatus != "已扫码，请在手机上确认" {
		t.Fatal(s.qrStatus)
	}
	r.tick(t)
	r.flush(t)
	if !s.loggedIn || s.account.Mid != 42 || s.qrImage != nil {
		t.Fatal("account not refreshed after QR confirmation")
	}
	s.enter("favorite")
	r.flush(t)
	s.stageFavorite(s.favoriteDetail.Data.Medias[0])
	s.loadFavorite(2, 1, false)
	s.logout()
	r.finish(t, 1)
	r.finish(t, 0)
	if s.loggedIn || s.account.Mid != 0 || len(s.favoritesData.Data.List) != 0 || len(s.favoriteDetail.Data.Medias) != 0 || len(s.stagedKeys) != 0 {
		t.Fatal("private state survived logout/late response")
	}
	if s.authError == "" {
		t.Fatal("logout error lost")
	}
}
func TestQRExpiryRefreshAndLeavingPage(t *testing.T) {
	s, api, r := setup(t, false)
	s.enter("profile")
	r.flush(t)
	api.codes = []uint32{model.PollQRCodeStatusExpired}
	r.tick(t)
	r.flush(t)
	if !s.qrExpired {
		t.Fatal("QR expiry ignored")
	}
	s.newQRCode()
	r.flush(t)
	if s.qrExpired {
		t.Fatal("new QR stayed expired")
	}
	r.tick(t) // poll in flight
	s.enter("up")
	api.codes = []uint32{model.PollQRCodeStatusScannedUnconfirmed}
	r.flush(t)
	if len(r.jobs) != 0 || s.qrStatus != "请使用 B 站 App 扫码" {
		t.Fatal("old poll mutated a hidden page")
	}
	s.enter("profile")
	s.Stop()
	r.flush(t)
	for _, timer := range r.timers {
		if !timer.canceled {
			t.Fatal("timer survived Stop")
		}
	}
}
func TestFollowAuthorPaginationAndLists(t *testing.T) {
	s, _, r := setup(t, true)
	s.enter("up")
	r.flush(t)
	s.loadFollows(2)
	r.flush(t)
	s.findAuthor("12", nil)
	r.flush(t)
	s.loadUploads(2, true)
	r.flush(t)
	if s.followPage != 2 || len(s.uploads.Data.List.Vlist) != 2 || s.authorInfo.Data.Mid != 12 {
		t.Fatal("author/pagination model mismatch")
	}
	s.loadUploads(3, true)
	s.findAuthor("13", nil)
	r.finish(t, 1)
	r.finish(t, 0)
	r.flush(t)
	if len(s.uploads.Data.List.Vlist) != 1 || s.uploads.Data.List.Vlist[0].Bvid != "BVupload13-1" {
		t.Fatal("old author response overwrote new author")
	}
	s.loadAuthorLists(1, false)
	r.flush(t)
	s.loadSeason(3, 1, false)
	r.flush(t)
	s.loadSeries(4, 1, false)
	r.flush(t)
	if len(s.authorLists.Data.ItemsLists.SeasonsList) != 1 || len(s.seasonVideos.Data.Archives) != 1 || len(s.seriesVideos.Data.Archives) != 1 {
		t.Fatal("author lists not loaded")
	}
}

func TestFavoriteStaleResponseAndRetry(t *testing.T) {
	s, api, r := setup(t, true)
	s.enter("favorite")
	r.finish(t, 0) // folders ready, first detail pending
	s.loadFavorite(2, 1, false)
	r.finish(t, 1)
	r.finish(t, 0)
	if s.favoriteDetail.Data.Info.ID != 2 {
		t.Fatal("stale folder overwrote current selection")
	}
	fail := true
	fallback := *api
	fallback.favorite = nil
	api.favorite = func(id, page int) (model.FavoriteData, error) {
		if page == 2 && fail {
			fail = false
			return model.FavoriteData{}, errors.New("page failure")
		}
		return fallback.Favorites(id, page, pageSize)
	}
	s.loadFavorite(2, 2, true)
	r.flush(t)
	if s.favoritePage != 1 || len(s.favoriteDetail.Data.Medias) != 1 || s.favoriteDetail.Error == "" {
		t.Fatal("failed pagination lost existing results")
	}
	s.loadFavorite(2, 2, true)
	r.flush(t)
	if s.favoritePage != 2 || len(s.favoriteDetail.Data.Medias) != 2 {
		t.Fatal("pagination retry did not append exactly once")
	}
	enableTestBatchDownloads(s, r)
	s.stageAllLibrary()
	r.flush(t)
	if len(s.batches) != 1 || s.batches[0].Done != 2 {
		t.Fatal("batch must fetch all pages")
	}
	s.stageAllLibrary()
	r.flush(t)
	if len(s.batches) != 2 || s.batches[1].Done != 2 {
		t.Fatal("a second invocation must create a separate session")
	}
}
func TestCollectionPagingSelectionAndStop(t *testing.T) {
	s, _, r := setup(t, true)
	s.favoriteTab = "collection"
	s.enter("favorite")
	r.flush(t)
	s.loadCollections(2, true)
	r.flush(t)
	s.loadCollection(1, 2, true)
	r.flush(t)
	if len(s.collections.Data.List) != 2 || len(s.collectionDetail.Data.Medias) != 2 {
		t.Fatal("collection pagination failed")
	}
	*s.search("collection") = "2"
	s.selectLibrary()
	s.stageSelectedLibrary()
	if len(s.stagedCollections) != 1 || s.stagedCollections[0].ID != 12 {
		t.Fatal("selection ignored local search")
	}
	s.loadCollection(2, 1, false)
	s.Stop()
	r.flush(t)
	if len(s.collectionDetail.Data.Medias) != 0 {
		t.Fatal("completion after Stop changed state")
	}
}
func TestFunctionalViews(t *testing.T) {
	for _, size := range [][2]int{{1200, 880}, {760, 600}} {
		s, _, r := setup(t, true)
		path := "/bilibili/profile"
		view := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, path, func(p string) { path = p }, func() { s.Form(c, path, func(p string) { path = p }, func() { path = "/bilibili/up" }) })
		}, size[0], size[1])
		for _, section := range []string{"profile", "up", "favorite", "collection", "download"} {
			path = "/bilibili/" + section
			view.Frame()
			r.flush(t)
			view.Frame()
			view.Frame()
			if section == "favorite" {
				s.selectLibrary()
				s.stageSelectedLibrary()
			}
			if section == "profile" && !view.HasText("原生测试账号") {
				t.Fatal("account model not rendered")
			}
			if section == "up" && !view.HasText("测试 UP 主 1") {
				t.Fatal("follow model not rendered")
			}
			if section == "favorite" && !view.HasText("收藏视频 1-1") {
				t.Fatal("favorite model not rendered")
			}
			if section == "collection" && !view.HasText("合集视频 1") {
				t.Fatal("collection model not rendered")
			}
			if section == "favorite" || section == "collection" {
				left, leftOK := view.Find("视频库列表标题")
				right, rightOK := view.Find("视频库内容标题")
				if !leftOK || !rightOK || abs(left.Y-right.Y) > 1 || abs(left.H-right.H) > 1 || left.H > 53 || right.H > 53 {
					t.Fatalf("library headers are misaligned at %v: left=%+v right=%+v", size, left, right)
				}
			}
			saveImage(t, view, fmt.Sprintf("/tmp/native-bili-functional-%s-%d.png", section, size[0]))
		}
	}
}

func TestBackRejectsPendingAuthorAndBatchCanCancel(t *testing.T) {
	s, _, r := setup(t, true)
	s.activePath = "/bilibili/up"
	navigated := false
	s.findAuthor("12", func(string) { navigated = true })
	s.Leave()
	r.flush(t)
	if navigated {
		t.Fatal("pending author search hijacked navigation after leaving the list")
	}
	s.enter("favorite")
	r.flush(t)
	enableTestBatchDownloads(s, r)
	s.stageAllLibrary()
	s.stopBatch(s.batches[0])
	r.flush(t)
	if len(s.stagedKeys) != 0 || s.batches[0].Active || s.batches[0].Status != "stopped" {
		t.Fatal("canceled batch still added records")
	}
}

func TestLibraryDownloadAllOpensDownloadPage(t *testing.T) {
	for _, tab := range []string{"favorite", "collection"} {
		t.Run(tab, func(t *testing.T) {
			s, _, runtime := setup(t, true)
			s.favoriteTab = tab
			s.enter("favorite")
			runtime.flush(t)
			path := ""
			enableTestBatchDownloads(s, runtime)
			s.stageAllLibrary(func(next string) { path = next })
			runtime.flush(t)
			if path != "/bilibili/download" || len(s.batches) != 1 || s.batches[0].Done != 2 {
				t.Fatalf("download all did not stage both videos and navigate: path=%s queued=%d", path, len(s.stagedKeys))
			}
		})
	}
}

func TestCollectionUsesFolderCountAndDeduplicatesPages(t *testing.T) {
	s, api, r := setup(t, true)
	api.collection = func(id, page int) (model.CollectionItemData, error) {
		return model.CollectionItemData{Medias: []model.CollectionItemMedias{{ID: int64(page), Bvid: fmt.Sprintf("BVitem%d", page)}}}, nil
	}
	s.favoriteTab = "collection"
	s.enter("favorite")
	r.flush(t)
	if !s.collectionHasMore {
		t.Fatal("missing detail total must fall back to the folder total")
	}
	s.loadCollection(1, 2, true)
	r.flush(t)
	if s.collectionHasMore || len(s.collectionDetail.Data.Medias) != 2 {
		t.Fatal("collection total not respected")
	}
	s.loadCollection(1, 2, true)
	r.flush(t)
	if len(s.collectionDetail.Data.Medias) != 2 {
		t.Fatal("repeated page duplicated media identities")
	}
}

func (f *fakeAPI) DownloadHistoryPage(int, int, string) (biliapi.HistoryPage, error) {
	return biliapi.HistoryPage{}, nil
}
func (f *fakeAPI) ClearDownloadHistory() error               { return nil }
func (f *fakeAPI) DeleteDownloadHistory(int64, string) error { return nil }

func enableTestBatchDownloads(s *State, r *fakeRuntime) {
	rt := r.runtime()
	rt.Background = rt.Run
	s.Configure(s.api, rt)
	s.ConfigureDownloads(&authorDownloadBackend{})
}
