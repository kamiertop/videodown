package douyin

import "github.com/kamiertop/videodown/internal/gui/style"

func (s *State) maybeMoreFolders() {
	if s.runtime.After == nil || s.folderAutoPending || !s.folders.Loaded || s.folders.Loading || s.folders.Error != "" || !s.folderMore {
		return
	}
	scroll := s.scroll("folders")
	if scroll.MaxY > 0 && scroll.Y+style.DynamicAutoLoadThreshold < scroll.MaxY {
		return
	}
	s.folderAutoPending = true
	epoch, version, path := s.epoch, s.folderVersion, s.activePath
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || s.folderVersion != version || s.activePath != path {
			return
		}
		s.folderAutoPending = false
		if scroll.MaxY > 0 && scroll.Y+style.DynamicAutoLoadThreshold < scroll.MaxY {
			return
		}
		s.loadFolders(true)
	})
}
