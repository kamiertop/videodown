package bilibili

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func autoPagingView(s *State, path *string) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		style.Apply(c)
		s.Layout(c, *path, func(next string) { *path = next }, func() {
			s.Form(c, *path, func(next string) { *path = next }, func() { *path = "/bilibili/up" })
		})
	}, 760, 600)
}

func TestAuthorUploadsLoadWhenScrolledNearEnd(t *testing.T) {
	s, api, runtime := setup(t, true)
	s.authorInfo = resource[model.UserInfoData]{Data: model.UserInfoData{Mid: 1, Name: "自动翻页作者"}, Loaded: true}
	s.follows.Loaded = true
	items := make([]model.VListItem, 24)
	for i := range items {
		items[i] = model.VListItem{Bvid: fmt.Sprintf("BVupload%02d", i), Title: fmt.Sprintf("投稿 %02d", i)}
	}
	s.uploads = resource[model.VideoListData]{Data: model.VideoListData{Page: model.VideoListPage{Count: 25}, List: model.VideoList{Vlist: items}}, Loaded: true}
	s.authorPage = 1
	api.upload = func(_ uint, _, page int) (model.VideoListData, error) {
		if page != 2 {
			t.Fatalf("unexpected upload page %d", page)
		}
		return model.VideoListData{Page: model.VideoListPage{Count: 25}, List: model.VideoList{Vlist: []model.VListItem{{Bvid: "BVupload24", Title: "最后一条投稿"}}}}, nil
	}
	path := "/bilibili/up/1"
	view := autoPagingView(s, &path)
	view.Frame()
	scroll := s.scroll("author-videos/1")
	if scroll.MaxY <= 0 || len(runtime.timers) != 1 || view.HasText("加载更多投稿") {
		t.Fatalf("upload grid should wait for scrolling: scroll=%+v timers=%d", scroll, len(runtime.timers))
	}
	runtime.tick(t)
	if len(runtime.jobs) != 0 {
		t.Fatal("scrollable uploads fetched before scrolling")
	}
	if err := view.Click("选择视频：投稿 00"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	region, ok := view.Find("UP 主视频区域")
	if !ok {
		t.Fatal("missing author scroll region")
	}
	view.Scroll(region.X+region.W/2, region.Y+region.H/2, 0, scroll.MaxY)
	view.Frame()
	if len(runtime.timers) != 1 {
		t.Fatal("scrolling to the end did not schedule the next upload page")
	}
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if len(s.uploads.Data.List.Vlist) != 25 || !view.HasText("最后一条投稿") || !*s.selection("BVupload00") {
		t.Fatal("next upload page did not append or selection was lost")
	}
}

func TestAuthorSeasonAndSeriesAutoFill(t *testing.T) {
	for _, kind := range []string{"season", "series"} {
		t.Run(kind, func(t *testing.T) {
			s, api, runtime := setup(t, true)
			s.authorInfo = resource[model.UserInfoData]{Data: model.UserInfoData{Mid: 1, Name: "合集作者"}, Loaded: true}
			s.authorTab = "lists"
			s.authorLists = resource[model.SeasonsSeriesData]{Data: model.SeasonsSeriesData{ItemsLists: model.SeasonsSeriesItemsLists{
				Page:        model.SeasonsSeriesPage{Total: 2},
				SeasonsList: []model.SeasonsItem{{Meta: model.SeasonsItemMeta{SeasonId: 3, Name: "测试合集", Total: 2}}},
				SeriesList:  []model.SeriesItem{{Meta: model.SeriesItemMeta{SeriesId: 4, Name: "测试系列", Total: 2}}},
			}}, Loaded: true}
			if kind == "season" {
				s.selectedSeason = 3
				s.seasonPage = 1
				s.seasonVideos = resource[model.SeasonsArchivesData]{Data: model.SeasonsArchivesData{Page: model.SeasonsArchivesPage{Total: 2}, Archives: []model.SeasonsArchivesItem{{Bvid: "BVseason1", Title: "合集首页"}}}, Loaded: true}
				api.season = func(_ string, _, page, id int) (model.SeasonsArchivesData, error) {
					if page != 2 || id != 3 {
						t.Fatalf("wrong season request: page=%d id=%d", page, id)
					}
					return model.SeasonsArchivesData{Page: model.SeasonsArchivesPage{Total: 2}, Archives: []model.SeasonsArchivesItem{{Bvid: "BVseason2", Title: "合集下一页"}}}, nil
				}
			} else {
				s.selectedSeries = 4
				s.seriesPage = 1
				s.seriesVideos = resource[model.SeriesArchivesData]{Data: model.SeriesArchivesData{Page: model.SeriesListPage{Total: 2}, Archives: []model.SeriesArchivesItem{{Bvid: "BVseries1", Title: "系列首页"}}}, Loaded: true}
				api.series = func(_ string, _, page, id int) (model.SeriesArchivesData, error) {
					if page != 2 || id != 4 {
						t.Fatalf("wrong series request: page=%d id=%d", page, id)
					}
					return model.SeriesArchivesData{Page: model.SeriesListPage{Total: 2}, Archives: []model.SeriesArchivesItem{{Bvid: "BVseries2", Title: "系列下一页"}}}, nil
				}
			}
			path := "/bilibili/up/1"
			view := autoPagingView(s, &path)
			view.Frame()
			if len(runtime.timers) != 1 || view.HasText("加载更多视频") {
				t.Fatalf("short %s page should auto-fill without a normal button", kind)
			}
			runtime.tick(t)
			runtime.flush(t)
			view.Frame()
			if kind == "season" && (len(s.seasonVideos.Data.Archives) != 2 || !view.HasText("合集下一页")) {
				t.Fatal("season video page did not auto-fill")
			}
			if kind == "series" && (len(s.seriesVideos.Data.Archives) != 2 || !view.HasText("系列下一页")) {
				t.Fatal("series video page did not auto-fill")
			}
		})
	}
}

func TestAuthorCollectionListAutoFill(t *testing.T) {
	s, api, runtime := setup(t, true)
	s.follows.Loaded = true
	s.authorInfo = resource[model.UserInfoData]{Data: model.UserInfoData{Mid: 1, Name: "合集作者"}, Loaded: true}
	s.authorTab = "lists"
	s.authorLists = resource[model.SeasonsSeriesData]{Data: model.SeasonsSeriesData{ItemsLists: model.SeasonsSeriesItemsLists{
		Page: model.SeasonsSeriesPage{Total: 2}, SeasonsList: []model.SeasonsItem{{Meta: model.SeasonsItemMeta{SeasonId: 3, Name: "第一页合集", Total: 1}}},
	}}, Loaded: true}
	s.listPage = 1
	s.selectedSeason = 3
	s.seasonVideos = resource[model.SeasonsArchivesData]{Data: model.SeasonsArchivesData{Page: model.SeasonsArchivesPage{Total: 1}, Archives: []model.SeasonsArchivesItem{{Bvid: "BVonly", Title: "合集视频"}}}, Loaded: true}
	api.authorLists = func(_ string, _, page int) (model.SeasonsSeriesData, error) {
		if page != 2 {
			t.Fatalf("wrong author list page %d", page)
		}
		return model.SeasonsSeriesData{ItemsLists: model.SeasonsSeriesItemsLists{
			Page: model.SeasonsSeriesPage{Total: 2}, SeriesList: []model.SeriesItem{{Meta: model.SeriesItemMeta{SeriesId: 4, Name: "下一页系列", Total: 1}}},
		}}, nil
	}
	path := "/bilibili/up/1"
	view := autoPagingView(s, &path)
	view.Frame()
	if len(runtime.timers) != 1 || view.HasText("加载更多") {
		t.Fatalf("short author list should auto-fill: timers=%d", len(runtime.timers))
	}
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if !view.HasText("下一页系列") || len(s.authorLists.Data.ItemsLists.SeriesList) != 1 {
		t.Fatal("author collection list did not append its next page")
	}
}

func TestFavoriteAndCollectionVideosAutoFill(t *testing.T) {
	for _, kind := range []string{"favorite", "collection"} {
		t.Run(kind, func(t *testing.T) {
			s, api, runtime := setup(t, true)
			if kind == "favorite" {
				s.favoritesData = resource[model.FavoritesData]{Data: model.FavoritesData{Count: 1, List: []model.FavoriteItem{{ID: 1, Title: "测试收藏夹", MediaCount: 2}}}, Loaded: true}
				s.favoriteID = 1
				s.favoritePage = 1
				s.favoriteDetail = resource[model.FavoriteData]{Data: model.FavoriteData{Info: model.FavoriteInfo{ID: 1, MediaCount: 2}, HasMore: true, Medias: []model.FavoriteMedias{{ID: 11, Type: 2, Bvid: "BVfav1", Title: "收藏首页"}}}, Loaded: true}
				api.favorite = func(_, page int) (model.FavoriteData, error) {
					if page != 2 {
						t.Fatalf("wrong favorite page %d", page)
					}
					return model.FavoriteData{Info: model.FavoriteInfo{ID: 1, MediaCount: 2}, Medias: []model.FavoriteMedias{{ID: 12, Type: 2, Bvid: "BVfav2", Title: "收藏下一页"}}}, nil
				}
			} else {
				s.collections = resource[model.CollectionData]{Data: model.CollectionData{Count: 1, List: []model.CollectionDataList{{ID: 1, Title: "测试合集", MediaCount: 2}}}, Loaded: true}
				s.collectionID = 1
				s.collectionDetailPage = 1
				s.collectionDetail = resource[model.CollectionItemData]{Data: model.CollectionItemData{Info: model.CollectionItemInfo{ID: 1, MediaCount: 2}, Medias: []model.CollectionItemMedias{{ID: 11, Bvid: "BVcol1", Title: "合集首页"}}}, Loaded: true}
				s.collectionHasMore = true
				api.collection = func(_, page int) (model.CollectionItemData, error) {
					if page != 2 {
						t.Fatalf("wrong collection page %d", page)
					}
					return model.CollectionItemData{Info: model.CollectionItemInfo{ID: 1, MediaCount: 2}, Medias: []model.CollectionItemMedias{{ID: 12, Bvid: "BVcol2", Title: "合集下一页"}}}, nil
				}
			}
			path := "/bilibili/" + kind
			view := autoPagingView(s, &path)
			view.Frame()
			if len(runtime.timers) != 1 || view.HasText("加载更多视频") {
				t.Fatalf("short %s page should auto-fill without a normal button: timers=%d", kind, len(runtime.timers))
			}
			runtime.tick(t)
			runtime.flush(t)
			view.Frame()
			if kind == "favorite" && (len(s.favoriteDetail.Data.Medias) != 2 || !view.HasText("收藏下一页")) {
				t.Fatal("favorite video page did not auto-fill")
			}
			if kind == "collection" && (len(s.collectionDetail.Data.Medias) != 2 || !view.HasText("合集下一页")) {
				t.Fatal("collection video page did not auto-fill")
			}
			saveImage(t, view, "/tmp/bilibili-"+kind+"-auto-page.png")
		})
	}
}

func TestCollectionFoldersAutoFillAndSwitchCancelsOldVideoPage(t *testing.T) {
	s, api, runtime := setup(t, true)
	s.collections = resource[model.CollectionData]{Data: model.CollectionData{Count: 2, HasMore: true, List: []model.CollectionDataList{{ID: 1, Title: "第一个合集", MediaCount: 2}}}, Loaded: true}
	s.collectionPage = 1
	s.collectionID = 1
	s.collectionDetail = resource[model.CollectionItemData]{Data: model.CollectionItemData{Info: model.CollectionItemInfo{ID: 1, MediaCount: 2}, Medias: []model.CollectionItemMedias{{ID: 11, Bvid: "BVold1", Title: "旧合集视频"}}}, Loaded: true}
	s.collectionDetailPage = 1
	s.collectionHasMore = true
	api.collectionList = func(page, _ int) (model.CollectionData, error) {
		if page != 2 {
			t.Fatalf("wrong folder page %d", page)
		}
		return model.CollectionData{Count: 2, List: []model.CollectionDataList{{ID: 2, Title: "第二个合集", MediaCount: 1}}}, nil
	}
	api.collection = func(id, page int) (model.CollectionItemData, error) {
		if id != 2 || page != 1 {
			t.Fatalf("old collection page loaded after switch: id=%d page=%d", id, page)
		}
		return model.CollectionItemData{Info: model.CollectionItemInfo{ID: 2, MediaCount: 1}, Medias: []model.CollectionItemMedias{{ID: 21, Bvid: "BVnew", Title: "新合集视频"}}}, nil
	}
	path := "/bilibili/collection"
	view := autoPagingView(s, &path)
	view.Frame()
	if len(runtime.timers) != 2 {
		t.Fatalf("both short folder and video regions should schedule pagination: timers=%d", len(runtime.timers))
	}
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if len(s.collections.Data.List) != 2 || !view.HasText("第二个合集") {
		t.Fatal("collection folder list did not auto-fill")
	}
	if err := view.Click("选择合集第二个合集"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	runtime.tick(t) // The old collection's pending timer must be inert.
	if len(runtime.jobs) != 0 || s.collectionID != 2 || !view.HasText("新合集视频") {
		t.Fatal("old collection timer loaded after switching folders")
	}
}

func TestFavoriteScrollPaginationErrorAndRetry(t *testing.T) {
	s, api, runtime := setup(t, true)
	s.favoritesData = resource[model.FavoritesData]{Data: model.FavoritesData{Count: 1, List: []model.FavoriteItem{{ID: 1, Title: "长收藏夹", MediaCount: 25}}}, Loaded: true}
	s.favoriteID = 1
	s.favoritePage = 1
	medias := make([]model.FavoriteMedias, 24)
	for i := range medias {
		medias[i] = model.FavoriteMedias{ID: i + 1, Type: 2, Bvid: fmt.Sprintf("BVfav%02d", i), Title: fmt.Sprintf("收藏视频 %02d", i)}
	}
	s.favoriteDetail = resource[model.FavoriteData]{Data: model.FavoriteData{Info: model.FavoriteInfo{ID: 1, MediaCount: 25}, HasMore: true, Medias: medias}, Loaded: true}
	fail := true
	api.favorite = func(_, page int) (model.FavoriteData, error) {
		if page != 2 {
			t.Fatalf("wrong favorite page %d", page)
		}
		if fail {
			fail = false
			return model.FavoriteData{}, errors.New("收藏分页失败")
		}
		return model.FavoriteData{Info: model.FavoriteInfo{ID: 1, MediaCount: 25}, Medias: []model.FavoriteMedias{{ID: 25, Type: 2, Bvid: "BVfav24", Title: "重试后的收藏"}}}, nil
	}
	path := "/bilibili/favorite"
	view := autoPagingView(s, &path)
	view.Frame()
	scroll := s.scroll("library/favorite/1/0")
	if scroll.MaxY <= 0 || len(runtime.timers) != 1 {
		t.Fatalf("long favorite grid should wait for scroll: scroll=%+v timers=%d", scroll, len(runtime.timers))
	}
	runtime.tick(t)
	if len(runtime.jobs) != 0 {
		t.Fatal("favorite page fetched before scrolling")
	}
	if err := view.Click("选择视频：收藏视频 00"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	panel, ok := view.Find("我的视频库")
	if !ok {
		t.Fatal("missing library panel")
	}
	view.Scroll(panel.X+panel.W-100, panel.Y+panel.H/2, 0, scroll.MaxY)
	view.Frame()
	if len(runtime.timers) != 1 {
		t.Fatal("scrolling near favorite bottom did not schedule the next page")
	}
	runtime.tick(t)
	runtime.flush(t)
	view.Frame()
	if !view.HasText("收藏分页失败") || !view.HasText("重试") || !*s.selection("BVfav00") {
		t.Fatal("failed page should keep existing cards and offer retry")
	}
	view.Scroll(panel.X+panel.W-100, panel.Y+panel.H/2, 0, scroll.MaxY)
	view.Frame()
	if err := view.Click("重试"); err != nil {
		t.Fatal(err)
	}
	runtime.flush(t)
	view.Frame()
	if len(s.favoriteDetail.Data.Medias) != 25 || !view.HasText("重试后的收藏") || view.HasText("收藏分页失败") {
		t.Fatalf("favorite retry did not append: count=%d page=%d error=%q hasNew=%v hasOldError=%v", len(s.favoriteDetail.Data.Medias), s.favoritePage, s.favoriteDetail.Error, view.HasText("重试后的收藏"), view.HasText("收藏分页失败"))
	}
}
