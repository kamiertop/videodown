package douyin

import (
	"testing"

	dyapi "github.com/kamiertop/videodown/douyin/api"
	"github.com/kamiertop/videodown/douyin/model"
)

func TestSwitchFolderPreservesSidebar(t *testing.T) {
	for _, kind := range []string{"collection", "folder"} {
		t.Run(kind, func(t *testing.T) {
			s := NewState()
			s.api = &dyapi.Douyin{}
			s.profile = resource[model.User]{Loaded: true, Data: model.User{Uid: "signed-in"}}
			root := "/douyin/favorite/" + kind
			s.activePath = root
			s.libraryTab = kind
			s.folders = resource[[]folder]{Loaded: true, Loading: true, seq: 7, Data: []folder{
				{Key: "one", Title: "第一个"}, {Key: "two", Title: "第二个"},
			}}
			s.folderCursor, s.folderMore, s.folderVersion = 20, true, 9
			s.scroll("folders").Y = 180
			jobs := 0
			s.runtime.Run = func(func() error, func(error)) { jobs++ }
			for i, key := range []string{"one", "two"} {
				s.selected["old"] = true
				s.videos.Data.Items = []model.AwemeItem{{AwemeId: "old"}}
				s.scroll("videos").Y = 300
				s.enter(root + "/" + key)
				if s.source.Key != kind+":"+key || s.activePath != root+"/"+key {
					t.Fatal("wrong video source")
				}
				if jobs != i+1 {
					t.Fatalf("expected only one video request per switch, got %d", jobs)
				}
				if !s.folders.Loaded || !s.folders.Loading || s.folders.seq != 7 || len(s.folders.Data) != 2 ||
					s.folderCursor != 20 || !s.folderMore || s.folderVersion != 9 || s.scroll("folders").Y != 180 {
					t.Fatal("switch reset the sidebar or invalidated its in-flight pagination")
				}
				if len(s.selected) != 0 || len(s.videos.Data.Items) != 0 || s.scroll("videos").Y != 0 {
					t.Fatal("old video state was not cleared")
				}
			}
			s.enter(root)
			if jobs != 2 || s.source.Key != "" || len(s.folders.Data) != 2 || s.scroll("folders").Y != 180 {
				t.Fatal("returning to list reloaded the sidebar")
			}
		})
	}
}

func TestSwitchFolderDoesNotReuseDifferentList(t *testing.T) {
	s := NewState()
	s.activePath = "/douyin/favorite/collection/one"
	s.folders = resource[[]folder]{Loaded: true, Data: []folder{{Key: "one"}}}
	for _, path := range []string{"/douyin/favorite/folder/one", "/douyin/user/one", "/douyin/favorite/collection/unknown"} {
		if s.enterKnownFolder(path) {
			t.Fatalf("reused sidebar for %s", path)
		}
	}
}
