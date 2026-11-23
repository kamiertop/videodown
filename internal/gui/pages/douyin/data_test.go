package douyin

import (
	"testing"

	dyapi "github.com/kamiertop/videodown/douyin/api"
	"github.com/kamiertop/videodown/douyin/model"
)

func testItems(ids ...string) []model.AwemeItem {
	items := make([]model.AwemeItem, len(ids))
	for i, id := range ids {
		items[i] = model.AwemeItem{AwemeId: id}
	}
	return items
}

// TestFavoriteTabSwitchKeepsCatalogs pins the reported flow: switching
// between the 收藏夹 and 合集 tabs (and back) keeps each catalog, loading
// every kind at most once, as bilibili's separate sidebar resources do.
func TestFavoriteTabSwitchKeepsCatalogs(t *testing.T) {
	s := NewState()
	s.api = &dyapi.Douyin{}
	s.profile = resource[model.User]{Loaded: true, Data: model.User{Uid: "signed-in"}}
	jobs := 0
	s.runtime.Run = func(func() error, func(error)) { jobs++ }

	// 进入收藏夹页签：加载一次目录。
	s.enter("/douyin/favorite/folder")
	if jobs != 1 {
		t.Fatalf("folder catalog jobs=%d, want 1", jobs)
	}
	folders := s.catalog("folder", "")
	folders.folders = resource[[]folder]{Loaded: true, seq: 7, Data: []folder{
		{Key: "one", Title: "第一个"}, {Key: "two", Title: "第二个"},
	}}
	folders.cursor, folders.more, folders.version = 20, true, 9
	s.scroll(folders.scrollKey).Y = 180

	// 切到合集页签：只加载合集目录，收藏夹目录原样保留。
	s.enter("/douyin/favorite/collection")
	if jobs != 2 {
		t.Fatalf("collection catalog jobs=%d, want one more", jobs)
	}
	if !folders.folders.Loaded || folders.folders.seq != 7 || len(folders.folders.Data) != 2 ||
		folders.cursor != 20 || !folders.more || folders.version != 9 || s.scroll(folders.scrollKey).Y != 180 {
		t.Fatal("switching tabs reset the folder catalog or its pagination")
	}
	collections := s.catalog("collection", "")
	collections.folders = resource[[]folder]{Loaded: true, Data: []folder{{Key: "mix"}}}

	// 切回收藏夹页签：目录已在，零请求。
	s.enter("/douyin/favorite/folder")
	if jobs != 2 {
		t.Fatalf("switching back re-requested the folder catalog (jobs=%d)", jobs)
	}
	if len(folders.folders.Data) != 2 || s.scroll(folders.scrollKey).Y != 180 {
		t.Fatal("switching back lost the folder catalog")
	}
	// 切回合集页签同样零请求。
	s.enter("/douyin/favorite/collection")
	if jobs != 2 || !collections.folders.Loaded {
		t.Fatal("switching back to collections re-requested")
	}
}

// TestFolderRoutePreservesSidebar pins the folder routes with per-section
// lists: the catalog survives navigation, a different folder swaps the list,
// and returning to a visited folder or the directory re-requests nothing.
func TestFolderRoutePreservesSidebar(t *testing.T) {
	for _, kind := range []string{"collection", "folder"} {
		t.Run(kind, func(t *testing.T) {
			s := NewState()
			s.api = &dyapi.Douyin{}
			s.profile = resource[model.User]{Loaded: true, Data: model.User{Uid: "signed-in"}}
			root := "/douyin/favorite/" + kind
			jobs := 0
			s.runtime.Run = func(func() error, func(error)) { jobs++ }

			s.enter(root) // 首次进入目录：加载目录一次。
			if jobs != 1 || s.libraryTab != kind {
				t.Fatalf("catalog jobs=%d tab=%q, want 1 and %q", jobs, s.libraryTab, kind)
			}
			// 模拟目录加载完成（真实流程在 loadFolders 的回调里）。
			cat := s.catalog(kind, "")
			cat.folders = resource[[]folder]{Loaded: true, seq: 7, Data: []folder{
				{Key: "one", Title: "第一个"}, {Key: "two", Title: "第二个"},
			}}
			cat.cursor, cat.more, cat.version = 20, true, 9
			s.scroll(cat.scrollKey).Y = 180

			// 打开收藏夹 one：请求一次列表；目录与分页状态原样。
			s.enter(root + "/one")
			if jobs != 2 || s.folderFeed.src.Key != kind+":one" {
				t.Fatalf("folder one: jobs=%d src=%q", jobs, s.folderFeed.src.Key)
			}
			if cat.folders.seq != 7 || len(cat.folders.Data) != 2 || cat.cursor != 20 ||
				!cat.more || cat.version != 9 || s.scroll(cat.scrollKey).Y != 180 {
				t.Fatal("opening a folder reset the sidebar catalog or its pagination")
			}
			// 模拟列表加载完成，离开再回来：零请求，数据与滚动保留。
			s.folderFeed.videos.Loaded = true
			s.folderFeed.videos.Data = videoPage{Items: testItems("v1")}
			s.enter("/douyin/dynamic") // 动态是独立列表，自己请求一次。
			if jobs != 3 {
				t.Fatalf("dynamic feed jobs=%d, want 3", jobs)
			}
			s.enter(root + "/one")
			if jobs != 3 || !s.folderFeed.videos.Loaded || s.folderFeed.videos.Data.Items[0].AwemeId != "v1" {
				t.Fatal("returning to the folder re-requested or lost its list")
			}

			// 换另一个收藏夹 two：换列表、重新请求。
			s.enter(root + "/two")
			if jobs != 4 || s.folderFeed.src.Key != kind+":two" || s.folderFeed.videos.Loaded {
				t.Fatalf("folder two: jobs=%d src=%q", jobs, s.folderFeed.src.Key)
			}
			// 回目录视图：目录保留，列表清空，零请求。
			s.enter(root)
			if jobs != 4 || cat.folders.seq != 7 || len(cat.folders.Data) != 2 || s.scroll(cat.scrollKey).Y != 180 {
				t.Fatal("returning to the directory reloaded the sidebar")
			}
			if s.folderFeed.src.Key != "" || s.folderFeed.videos.Loaded {
				t.Fatal("directory view should clear the folder list")
			}
		})
	}
}

// TestSeriesTabPerAuthor verifies each author's series catalog is its own:
// switching authors reloads, switching back reuses.
func TestSeriesTabPerAuthor(t *testing.T) {
	s := NewState()
	s.api = &dyapi.Douyin{}
	s.profile = resource[model.User]{Loaded: true, Data: model.User{Uid: "signed-in"}}
	jobs := 0
	s.runtime.Run = func(func() error, func(error)) { jobs++ }

	s.authorTab = "series"
	s.author.Loaded = true
	s.author.Data = model.User{SecUid: "userA", NickName: "甲"}
	s.ensureFolders("series", "userA")
	if jobs != 1 {
		t.Fatalf("series catalog jobs=%d, want 1", jobs)
	}
	a := s.catalog("series", "userA")
	a.folders = resource[[]folder]{Loaded: true, Data: []folder{{Key: "s1"}}}

	// 同一作者再次进入：零请求。
	s.ensureFolders("series", "userA")
	if jobs != 1 {
		t.Fatal("same author's series catalog re-requested")
	}

	// 另一作者的系列是另一份目录。
	s.ensureFolders("series", "userB")
	if jobs != 2 || a.folders.seq != 0 || !a.folders.Loaded {
		t.Fatal("different author should not touch userA's catalog")
	}
	b := s.catalog("series", "userB")
	if len(b.folders.Data) != 0 {
		t.Fatal("userB catalog should be empty before its load completes")
	}
}
