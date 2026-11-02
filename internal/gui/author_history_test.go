package gui

import (
	"fmt"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/gui/pages/bilibili"
)

type historyAuthorAPI struct {
	bilibili.API
	infoCalls int
}

func (a *historyAuthorAPI) IsLoggedIn() bool { return true }
func (a *historyAuthorAPI) RefreshCookie() (model.CookieRefreshData, error) {
	return model.CookieRefreshData{}, nil
}
func (a *historyAuthorAPI) MyInfo() (model.MyInfoProfile, error) {
	return model.MyInfoProfile{Mid: 42}, nil
}
func (a *historyAuthorAPI) FollowList(page, size int) (model.FollowData, error) {
	return model.FollowData{Total: 21, List: []model.FollowDataList{{Mid: int64(page), Uname: fmt.Sprintf("作者 %d", page)}}}, nil
}
func (a *historyAuthorAPI) Info(raw string) (model.UserInfoData, error) {
	a.infoCalls++
	return model.UserInfoData{Mid: 2, Name: "作者 2"}, nil
}
func (a *historyAuthorAPI) VideoList(uint, int, int) (model.VideoListData, error) {
	return model.VideoListData{Page: model.VideoListPage{Count: 1}, List: model.VideoList{Vlist: []model.VListItem{{Bvid: "BVfixture", Title: "作者的视频卡片", Length: "01:00"}}}}, nil
}
func (a *historyAuthorAPI) FavoritesList() (model.FavoritesData, error) {
	return model.FavoritesData{Count: 1, List: []model.FavoriteItem{{ID: 1, Title: "测试收藏夹", MediaCount: 1}}}, nil
}
func (a *historyAuthorAPI) Favorites(int, int, int) (model.FavoriteData, error) {
	return model.FavoriteData{Info: model.FavoriteInfo{ID: 1, Title: "测试收藏夹", MediaCount: 1}, Medias: []model.FavoriteMedias{{ID: 1, Bvid: "BVfavorite", Type: 2, Title: "收藏视频", Upper: model.FavoriteMediaUpper{Mid: 2, Name: "测试作者"}}}}, nil
}
func (a *historyAuthorAPI) Collection(int, int) (model.CollectionData, error) {
	return model.CollectionData{Count: 1, List: []model.CollectionDataList{{ID: 3, Title: "测试合集", MediaCount: 1}}}, nil
}
func (a *historyAuthorAPI) CollectionItem(string, int, int) (model.CollectionItemData, error) {
	return model.CollectionItemData{Info: model.CollectionItemInfo{ID: 3, Title: "测试合集", MediaCount: 1}, Medias: []model.CollectionItemMedias{{ID: 7, Bvid: "BVcollection", Title: "合集视频", Upper: model.CollectionItemMediaUpper{Mid: 2, Name: "测试作者"}}}}, nil
}

func TestLibraryTabAndAuthorUseTitlebarHistory(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s := NewShell()
		api := &historyAuthorAPI{}
		type job struct {
			work func() error
			done func(error)
		}
		var jobs []job
		s.pages.Bilibili.Configure(api, guiruntime.Runtime{Run: func(work func() error, done func(error)) { jobs = append(jobs, job{work, done}) }})
		s.pages.Bilibili.Start()
		flush := func() {
			for len(jobs) > 0 {
				next := jobs[0]
				jobs = jobs[1:]
				next.done(next.work())
			}
		}
		flush()
		s.router.Push("/bilibili/favorite")
		view := ui.NewTester(s.view, width, 880)
		frame := func() { view.Frame(); flush(); view.Frame(); view.Advance(300 * time.Millisecond); view.Frame() }
		click := func(label string) {
			t.Helper()
			if err := view.Click(label); err != nil {
				t.Fatal(err)
			}
			frame()
		}
		frame()
		click("合集")
		if s.router.Path() != "/bilibili/collection" || !view.HasText("合集视频") {
			t.Fatalf("collection tab did not push route: %s", s.router.Path())
		}
		click("后退")
		if s.router.Path() != "/bilibili/favorite" || !view.HasText("收藏视频") {
			t.Fatalf("back did not restore favorites: %s", s.router.Path())
		}
		click("前进")
		if s.router.Path() != "/bilibili/collection" {
			t.Fatalf("forward did not restore collections: %s", s.router.Path())
		}
		click("进入 UP 主主页：测试作者")
		if s.router.Path() != "/bilibili/up/2" {
			t.Fatalf("author link did not push detail route: %s", s.router.Path())
		}
		click("后退")
		if s.router.Path() != "/bilibili/collection" || !view.HasText("合集视频") {
			t.Fatalf("author back did not restore collection: %s", s.router.Path())
		}
		click("选择视频：合集视频")
		if !view.HasText("下载已选 (1)") {
			t.Fatal("card click no longer selects the video")
		}
		s.pages.Bilibili.Stop()
	}
}

func TestAuthorDetailUsesTitlebarHistory(t *testing.T) {
	for _, width := range []int{1200, 760} {
		s := NewShell()
		api := &historyAuthorAPI{}
		type job struct {
			work func() error
			done func(error)
		}
		var jobs []job
		s.pages.Bilibili.Configure(api, guiruntime.Runtime{Run: func(work func() error, done func(error)) { jobs = append(jobs, job{work, done}) }})
		s.pages.Bilibili.Start()
		defer s.pages.Bilibili.Stop()
		flush := func() {
			for len(jobs) > 0 {
				next := jobs[0]
				jobs = jobs[1:]
				next.done(next.work())
			}
		}
		flush()
		s.router.Push("/bilibili/up")
		view := ui.NewTester(s.view, width, 880)
		frame := func() { view.Frame(); flush(); view.Frame(); view.Advance(300 * time.Millisecond); view.Frame() }
		frame()
		click := func(label string) {
			t.Helper()
			if err := view.Click(label); err != nil {
				t.Fatal(err)
			}
			frame()
		}
		click("下一页")
		if !view.HasText("第 2 / 2 页") {
			t.Fatal("follow page not selected")
		}
		click("进入作者 2的主页")
		if s.router.Path() != "/bilibili/up/2" || !view.HasText("作者的视频卡片") {
			t.Fatalf("detail not routed: %s", s.router.Path())
		}
		file, err := os.Create(fmt.Sprintf("/tmp/bilibili-up-detail-%d.png", width))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(file, view.Image())
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		if view.HasText("我的关注 · 21") || view.HasText("返回关注") {
			t.Fatal("follow toolbar leaked into the detail page")
		}
		calls := api.infoCalls
		click("后退")
		if s.router.Path() != "/bilibili/up" || !view.HasText("第 2 / 2 页") || !view.HasText("作者 2") {
			t.Fatal("titlebar back lost the follow page")
		}
		click("前进")
		if s.router.Path() != "/bilibili/up/2" || !view.HasText("作者的视频卡片") || api.infoCalls != calls {
			t.Fatal("forward did not restore the loaded detail")
		}
		if err := view.Click("后退"); err != nil {
			t.Fatal(err)
		}
		frame()
		if err := view.Click("UP 主 UID 或空间链接"); err != nil {
			t.Fatal(err)
		}
		view.Type("2")
		view.Key(0, ui.KeyEnter)
		frame()
		if s.router.Path() != "/bilibili/up/2" {
			t.Fatal("query submission must also create a detail history entry")
		}
		click("返回")
		if s.router.Path() != "/bilibili/up" || !view.HasText("第 2 / 2 页") {
			t.Fatal("detail return button did not restore the follow page")
		}
	}
}
