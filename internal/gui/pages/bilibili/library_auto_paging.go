package bilibili

import "fmt"

func (s *State) libraryScrollKey() string {
	return fmt.Sprintf("library/%s/%d/%d", s.library.tab, s.library.folderID, s.library.collectionID)
}

func (s *State) maybeAutoLibraryFolders() {
	if s.library.tab != "collection" {
		return
	}
	const key = "folders/collection"
	s.maybeAutoPage(key, key, func() bool {
		return s.activePath == "/bilibili/collection" && s.library.tab == "collection" &&
			s.library.collections.Loaded && !s.library.collections.Loading && s.library.collections.Error == "" && s.library.collections.Data.HasMore
	}, func() { s.loadCollections(s.library.collectionPage+1, true) })
}

func (s *State) maybeAutoLibraryVideos() {
	key := s.libraryScrollKey()
	if s.library.tab == "collection" {
		id := s.library.collectionID
		s.maybeAutoPage(key, key, func() bool {
			return id != 0 && s.activePath == "/bilibili/collection" && s.library.tab == "collection" && s.library.collectionID == id &&
				s.library.collectionDetail.Loaded && !s.library.collectionDetail.Loading && s.library.collectionDetail.Error == "" && s.library.collectionHasMore
		}, func() { s.loadCollection(id, s.library.collectionDetailPage+1, true) })
		return
	}
	id := s.library.folderID
	s.maybeAutoPage(key, key, func() bool {
		return id != 0 && s.activePath == "/bilibili/favorite" && s.library.tab == "favorite" && s.library.folderID == id &&
			s.library.folderDetail.Loaded && !s.library.folderDetail.Loading && s.library.folderDetail.Error == "" && s.library.folderDetail.Data.HasMore
	}, func() { s.loadFavorite(id, s.library.folderPage+1, true) })
}
