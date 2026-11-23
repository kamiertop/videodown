package douyin

import (
	"testing"

	dyapi "github.com/kamiertop/videodown/douyin/api"
	"github.com/kamiertop/videodown/douyin/model"
)

// TestSectionBackNavigationReusesData pins the reported flow: visiting
// 用户 → 动态 → 收藏 and then navigating back step by step must not
// re-request anything, matching bilibili's per-section state.
func TestSectionBackNavigationReusesData(t *testing.T) {
	s := NewState()
	s.api = &dyapi.Douyin{}
	s.profile = resource[model.User]{Loaded: true, Data: model.User{Uid: "signed-in"}}
	jobs := 0
	s.runtime.Run = func(func() error, func(error)) { jobs++ }

	// 进入用户列表：一次请求，然后模拟完成加载。
	s.enter("/douyin/user")
	if jobs != 1 {
		t.Fatalf("follow list requested %d times, want 1", jobs)
	}
	s.follows.Loaded = true

	// 进入动态：一次请求，模拟列表加载完成并滚动。
	s.enter("/douyin/dynamic")
	if jobs != 2 {
		t.Fatalf("dynamic feed requested %d times, want 1", jobs)
	}
	s.dynamicFeed.videos.Loaded = true
	s.dynamicFeed.videos.Data = videoPage{Items: testItems("dyn-1", "dyn-2"), Next: 10, More: true}
	s.scroll("videos/dynamic").Y = 240

	// 进入收藏：独立列表再次请求（数据不同），模拟完成。
	s.enter("/douyin/favorite")
	if jobs != 3 {
		t.Fatalf("favorite list requested %d times, want 1", jobs)
	}
	s.favoriteFeed.videos.Loaded = true
	s.favoriteFeed.videos.Data = videoPage{Items: testItems("fav-1")}

	// 后退到动态：独立列表原样恢复，零请求。
	s.enter("/douyin/dynamic")
	if jobs != 3 {
		t.Fatalf("back to dynamic re-requested (jobs=%d)", jobs)
	}
	list := s.activeList("/douyin/dynamic")
	if list != &s.dynamicFeed || !list.videos.Loaded || len(list.videos.Data.Items) != 2 ||
		list.videos.Data.Items[0].AwemeId != "dyn-1" || !list.videos.Data.More {
		t.Fatal("dynamic list was not restored in place")
	}
	if got := s.scroll("videos/dynamic").Y; got != 240 {
		t.Fatalf("scroll position lost on back navigation: %.1f", got)
	}

	// 后退到用户：关注列表仍在，零请求。
	s.enter("/douyin/user")
	if jobs != 3 {
		t.Fatalf("back to user re-requested (jobs=%d)", jobs)
	}
	if !s.follows.Loaded {
		t.Fatal("follow list state was dropped")
	}
}

// TestAuthorBackNavigationRestoresCachedPage verifies a visited user's page
// (profile and video list are the author section's own state) survives
// navigating away and back without a request.
func TestAuthorBackNavigationRestoresCachedPage(t *testing.T) {
	s := NewState()
	s.api = &dyapi.Douyin{}
	s.profile = resource[model.User]{Loaded: true, Data: model.User{Uid: "signed-in"}}
	jobs := 0
	s.runtime.Run = func(func() error, func(error)) { jobs++ }

	s.enter("/douyin/user/MS4wLjAB")
	if jobs != 1 {
		t.Fatalf("author page requested %d times, want 1", jobs)
	}
	// 模拟资料与列表加载完成（真实流程在各自回调里，回调会设置来源）。
	s.author.Loaded = true
	s.author.Data = model.User{SecUid: "MS4wLjAB", NickName: "测试用户"}
	s.authorFeed.src = s.authorSource("MS4wLjAB", "测试用户")
	s.authorFeed.videos.Loaded = true
	s.authorFeed.videos.Data = videoPage{Items: testItems("post-1")}

	s.enter("/douyin/dynamic")
	if jobs != 2 {
		t.Fatalf("dynamic feed requested %d times, want 1", jobs)
	}
	s.dynamicFeed.videos.Loaded = true
	s.dynamicFeed.videos.Data = videoPage{Items: testItems("dyn-1")}

	// 返回用户主页：资料与作品列表都是用户板块自己的状态，零请求。
	s.enter("/douyin/user/MS4wLjAB")
	if jobs != 2 {
		t.Fatalf("back to author page re-requested (jobs=%d)", jobs)
	}
	if !s.author.Loaded || s.author.Data.SecUid != "MS4wLjAB" {
		t.Fatal("author profile was not restored")
	}
	if s.authorFeed.src.Key != "user:MS4wLjAB" || !s.authorFeed.videos.Loaded ||
		s.authorFeed.videos.Data.Items[0].AwemeId != "post-1" {
		t.Fatal("author video list was not restored")
	}
}

// TestHistoryBackNavigationSkipsReload verifies returning to the history
// page keeps its loaded list.
func TestHistoryBackNavigationSkipsReload(t *testing.T) {
	s := NewState()
	s.backend = nil // loadHistory guards on the backend; simulate a loaded list.
	s.enter("/douyin/history")
	s.history.Loaded = true
	s.history.Data.Total = 3
	s.enter("/douyin/dynamic")
	s.enter("/douyin/history")
	if !s.history.Loaded || s.history.Data.Total != 3 {
		t.Fatal("history state was dropped on back navigation")
	}
}
