package douyin

import "github.com/kamiertop/videodown/internal/gui/style"

// maybeMoreFolders 在目录滚动接近底部时自动追加下一页；
// 每类目录的加载状态各自独立。
func (s *State) maybeMoreFolders(cat *folderCatalog) {
	if s.runtime.After == nil || cat.autoPending || !cat.folders.Loaded || cat.folders.Loading || cat.folders.Error != "" || !cat.more {
		return
	}
	scroll := s.scroll(cat.scrollKey)
	if scroll.MaxY > 0 && scroll.Y+style.DynamicAutoLoadThreshold < scroll.MaxY {
		return
	}
	cat.autoPending = true
	epoch, version, path := s.epoch, cat.version, s.activePath
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || cat.version != version || s.activePath != path {
			return
		}
		cat.autoPending = false
		if scroll.MaxY > 0 && scroll.Y+style.DynamicAutoLoadThreshold < scroll.MaxY {
			return
		}
		s.loadFolders(cat, true)
	})
}
