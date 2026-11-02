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
	"github.com/kamiertop/videodown/douyin/model"
)

type followsAPI struct {
	*testAPI
	follow func(int) (model.FollowResponse, error)
	search func(string) ([]model.SearchFollowResponse, error)
}

func (a *followsAPI) FollowList(cursor int) (model.FollowResponse, error)         { return a.follow(cursor) }
func (a *followsAPI) SearchFollow(q string) ([]model.SearchFollowResponse, error) { return a.search(q) }
func followPage(cursor int) model.FollowResponse {
	items := []model.FollowItem{}
	for i := 0; i < 20; i++ {
		items = append(items, model.FollowItem{SecUid: fmt.Sprintf("MS4wLjAB_%d", cursor+i), Nickname: fmt.Sprintf("关注作者 %d", cursor+i), Signature: "关注用户签名"})
	}
	return model.FollowResponse{Total: 40, Offset: cursor + 20, HasMore: cursor == 0, Followings: items}
}
func TestFollowsLoadAutomaticallyAndKeepCardsOnError(t *testing.T) {
	s, a, _, r := setup(t)
	calls := 0
	fail := false
	s.api = &followsAPI{testAPI: a, follow: func(cursor int) (model.FollowResponse, error) {
		calls++
		if fail {
			return model.FollowResponse{}, errors.New("加载失败")
		}
		return followPage(cursor), nil
	}}
	s.enter("/douyin/user")
	r.flush(t)
	scroll := s.scroll("authors")
	scroll.MaxY = 600
	scroll.Y = 0
	s.maybeMoreFollows()
	if len(r.timers) != 0 {
		t.Fatal("loaded before approaching bottom")
	}
	scroll.Y = 550
	s.maybeMoreFollows()
	s.maybeMoreFollows()
	if len(r.timers) != 1 {
		t.Fatal("missing or duplicated auto paging timer")
	}
	fail = true
	r.timers[0]()
	r.flush(t)
	if len(s.follows.Data.Followings) != 20 || s.follows.Error == "" {
		t.Fatal("failed page discarded existing users")
	}
	n := len(r.timers)
	s.maybeMoreFollows()
	if len(r.timers) != n {
		t.Fatal("failed page retried in a loop")
	}
	fail = false
	s.loadFollows(true)
	r.flush(t)
	if len(s.follows.Data.Followings) != 40 || s.follows.Data.HasMore || calls != 3 {
		t.Fatal("retry did not append the next page")
	}
	s.loadFollows(false)
	r.flush(t)
	scroll.MaxY = 0
	s.maybeMoreFollows()
	tick := r.timers[len(r.timers)-1]
	s.enter("/douyin/download")
	tick()
	if len(r.jobs) != 0 {
		t.Fatal("obsolete route continued paging")
	}
}
func TestFollowSearchDebouncesAndDiscardsStaleResults(t *testing.T) {
	s, a, _, r := setup(t)
	calls := []string{}
	s.api = &followsAPI{testAPI: a, follow: func(cursor int) (model.FollowResponse, error) { return followPage(cursor), nil }, search: func(q string) ([]model.SearchFollowResponse, error) {
		calls = append(calls, q)
		return []model.SearchFollowResponse{{SecUid: "MS4wLjAB_search", NickName: q, FollowStatus: "follow"}}, nil
	}}
	s.enter("/douyin/user")
	r.flush(t)
	s.followSearchInput = "旧关键词"
	s.syncFollowSearch()
	s.followSearchInput = "新关键词"
	s.syncFollowSearch()
	r.timers[0]()
	if len(r.jobs) != 0 {
		t.Fatal("superseded debounce started a search")
	}
	r.timers[1]()
	if len(r.jobs) != 1 {
		t.Fatal("current search missing")
	}
	s.followSearchInput = "最终关键词"
	s.syncFollowSearch()
	r.finish(t, 0)
	if s.followResults.Loaded {
		t.Fatal("obsolete search response accepted")
	}
	s.maybeMoreFollows()
	if s.followAutoPending {
		t.Fatal("search should suppress list pagination")
	}
	r.timers[2]()
	r.flush(t)
	if len(calls) != 2 || len(s.followResults.Data) != 1 || s.followResults.Data[0].NickName != "最终关键词" {
		t.Fatal("search result missing")
	}
	s.followSearchInput = ""
	s.syncFollowSearch()
	if s.followResults.Loaded || s.followSearchPending || len(s.follows.Data.Followings) != 20 || s.follows.Data.Offset != 20 {
		t.Fatal("clearing search did not restore cached follows")
	}
}
func TestFollowsSearchLayoutAndNavigation(t *testing.T) {
	s, a, _, r := setup(t)
	s.api = &followsAPI{testAPI: a, follow: func(cursor int) (model.FollowResponse, error) { return followPage(cursor), nil }, search: func(q string) ([]model.SearchFollowResponse, error) {
		return []model.SearchFollowResponse{{SecUid: "MS4wLjAB_search", NickName: "搜索到的关注作者", FollowStatus: "follow"}}, nil
	}}
	path := "/douyin/user"
	navigate := func(p string) { path = p }
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() { s.Layout(c, path, navigate, func() { s.Form(c, path, navigate, func() {}) }) })
	}, 1200, 800)
	test.Frame()
	r.flush(t)
	test.Frame()
	for _, size := range [][2]int{{760, 600}, {1200, 800}} {
		test.SetSize(size[0], size[1])
		test.Frame()
		test.Advance(time.Second)
		for _, label := range []string{"搜索已关注用户", "抖音用户主页", "解析"} {
			b, ok := test.Find(label)
			if !ok || b.X+b.W > float32(size[0]) || b.Y < 0 {
				t.Fatalf("toolbar overflow at %v: %s %+v", size, label, b)
			}
		}
		if test.HasText("加载更多用户") {
			t.Fatal("manual paging button still visible")
		}
	}
	f, err := os.Create("/tmp/douyin-follows.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, test.Image()); err != nil {
		t.Fatal(err)
	}
	if err := test.Click("搜索已关注用户"); err != nil {
		t.Fatal(err)
	}
	test.Type("测试")
	test.Frame()
	if !s.followSearchPending {
		t.Fatal("typing did not schedule search")
	}
	r.timers[len(r.timers)-1]()
	r.flush(t)
	test.Frame()
	if err := test.Click("查看抖音用户：搜索到的关注作者"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if path != "/douyin/user/MS4wLjAB_search" {
		t.Fatal("search result did not open author")
	}
}
