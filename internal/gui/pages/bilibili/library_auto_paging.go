package bilibili

import "fmt"

func (s *State) libraryScrollKey() string {
	return fmt.Sprintf("library/%s/%d/%d", s.favoriteTab, s.favoriteID, s.collectionID)
}

func (s *State) maybeAutoLibraryFolders() {
	if s.favoriteTab != "collection" {
		return
	}
	const key = "folders/collection"
	s.maybeAutoPage(key, key, func() bool {
		return s.activePath == "/bilibili/collection" && s.favoriteTab == "collection" &&
			s.collections.Loaded && !s.collections.Loading && s.collections.Error == "" && s.collections.Data.HasMore
	}, func() { s.loadCollections(s.collectionPage+1, true) })
}

func (s *State) maybeAutoLibraryVideos() {
	key := s.libraryScrollKey()
	if s.favoriteTab == "collection" {
		id := s.collectionID
		s.maybeAutoPage(key, key, func() bool {
			return id != 0 && s.activePath == "/bilibili/collection" && s.favoriteTab == "collection" && s.collectionID == id &&
				s.collectionDetail.Loaded && !s.collectionDetail.Loading && s.collectionDetail.Error == "" && s.collectionHasMore
		}, func() { s.loadCollection(id, s.collectionDetailPage+1, true) })
		return
	}
	id := s.favoriteID
	s.maybeAutoPage(key, key, func() bool {
		return id != 0 && s.activePath == "/bilibili/favorite" && s.favoriteTab == "favorite" && s.favoriteID == id &&
			s.favoriteDetail.Loaded && !s.favoriteDetail.Loading && s.favoriteDetail.Error == "" && s.favoriteDetail.Data.HasMore
	}, func() { s.loadFavorite(id, s.favoritePage+1, true) })
}
