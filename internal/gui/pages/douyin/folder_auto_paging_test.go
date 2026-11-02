package douyin

import (
	"errors"
	"testing"
)

func TestFolderListsAutoLoadAndStopOnErrorsOrObsoleteRoutes(t *testing.T) {
	s, _, _, r := setup(t)
	s.activePath = "/douyin/favorite/collection"
	fail := false
	s.folderNext = func(cursor int) ([]folder, int, bool, error) {
		if fail {
			return nil, 0, false, errors.New("合集加载失败")
		}
		if cursor == 0 {
			return []folder{{Key: "1", Title: "第一项"}}, 1, true, nil
		}
		return []folder{{Key: "2", Title: "第二项"}}, 2, false, nil
	}
	s.loadFolders(false)
	r.flush(t)
	scroll := s.scroll("folders")
	scroll.MaxY = 400
	scroll.Y = 0
	s.maybeMoreFolders()
	if len(r.timers) != 0 {
		t.Fatal("list loaded away from bottom")
	}
	scroll.Y = 350
	s.maybeMoreFolders()
	s.maybeMoreFolders()
	if len(r.timers) != 1 {
		t.Fatal("automatic loading should schedule once")
	}
	fail = true
	r.timers[0]()
	r.flush(t)
	if len(s.folders.Data) != 1 || s.folders.Error == "" {
		t.Fatal("failed page lost the current list")
	}
	n := len(r.timers)
	s.maybeMoreFolders()
	if len(r.timers) != n {
		t.Fatal("failed page retried automatically")
	}
	fail = false
	s.loadFolders(true)
	r.flush(t)
	if len(s.folders.Data) != 2 || s.folderMore {
		t.Fatal("next list page not merged")
	}
	s.loadFolders(false)
	r.flush(t)
	scroll.MaxY = 0
	s.maybeMoreFolders()
	tick := r.timers[len(r.timers)-1]
	s.enter("/douyin/download")
	tick()
	if len(r.jobs) != 0 {
		t.Fatal("left page continued loading")
	}
}
func TestDuplicateFolderPageEndsAutomaticLoading(t *testing.T) {
	s, _, _, r := setup(t)
	s.folderNext = func(cursor int) ([]folder, int, bool, error) { return []folder{{Key: "same"}}, cursor + 1, true, nil }
	s.loadFolders(false)
	r.flush(t)
	s.loadFolders(true)
	r.flush(t)
	if len(s.folders.Data) != 1 || s.folderMore {
		t.Fatal("duplicate pages should not loop")
	}
}
