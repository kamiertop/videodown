package bilibili

import (
	"fmt"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"strconv"
	"strings"
)

const pageSize = 20

func (s *State) clearPrivateData() {
	s.history.seq++
	s.history.Loading = false
	s.historyMore.seq++
	s.historyMore.Loading = false
	s.historySearchPending, s.historyMutating, s.historyClearOpen = false, false, false
	s.follows = resource[model.FollowData]{}
	s.authorInfo = resource[model.UserInfoData]{}
	s.uploads = resource[model.VideoListData]{}
	s.authorLists = resource[model.SeasonsSeriesData]{}
	s.seasonVideos = resource[model.SeasonsArchivesData]{}
	s.seriesVideos = resource[model.SeriesArchivesData]{}
	s.favoritesData = resource[model.FavoritesData]{}
	s.favoriteDetail = resource[model.FavoriteData]{}
	s.collections = resource[model.CollectionData]{}
	s.collectionDetail = resource[model.CollectionItemData]{}
	s.dynamics = resource[model.DynamicArchivePage]{}
	s.dynamicMore = resource[model.DynamicArchivePage]{}
	s.dynamicOffset = ""
	s.dynamicHasMore = false
	s.dynamicGeneration++
	s.dynamicSelected = make(map[string]bool)
	s.dynamicAutoPending = false
	s.dynamicAutoAtY = 0
	s.dynamicAutoFillPages = 0
	s.playHistory = resource[model.PlayHistoryData]{}
	s.playHistoryMore = resource[model.PlayHistoryData]{}
	s.playHistoryCursor = model.PlayHistoryCursor{}
	s.playHistoryHasMore = false
	s.playHistoryGeneration++
	s.playHistorySelected = make(map[string]*bool)
	s.playHistoryAutoPending = false
	s.playHistoryAutoAtY = 0
	s.playHistoryAutoFillPages = 0
	s.autoPages = make(map[string]*autoPageState)
	s.favoriteID = 0
	s.collectionID = 0
	s.selectedSeason = 0
	s.selectedSeries = 0
	s.selected = make(map[string]*bool)
	s.images = make(map[string]*imageEntry)
	s.stagedFavorites = nil
	s.stagedCollections = nil
	s.stagedUploads = nil
	s.stagedSeasons = nil
	s.stagedSeries = nil
	s.stagedDynamics = nil
	s.stagedPlayHistory = nil
	s.stagedManual = nil
	s.manualGroup = nil
	s.manualParse = resource[model.VideoDetailConciseData]{}
	s.manualParseInput = ""
	s.notice = ""
	s.batches = nil
	s.pendingManualDownloads = nil
	s.activeManualDownloads = nil
	s.queueRequested = make(map[string]bool)
	s.stageError = ""
	s.stagedKeys = make(map[string]bool)
	s.stagedSourceNames = make(map[string]string)
	s.queueResolves = make(map[string]*queueResolve)
	s.queueProgress = make(map[string]bilidownload.Progress)
	s.queueExtraBusy = make(map[string]string)
	s.queueDownloading = false
	s.queueDownloadStop = false
	s.queueDownloadError = ""
	s.queueDownloadTotal, s.queueDownloadDone, s.queueDownloadFailed = 0, 0, 0
}
func (s *State) loadFollows(page int) {
	if s.follows.Loading || !s.loggedIn {
		return
	}
	load(s, &s.follows, func() (model.FollowData, error) { return s.api.FollowList(page, pageSize) }, func(model.FollowData) { s.followPage = page })
}
func (s *State) findAuthor(raw string, navigate func(string)) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !s.loggedIn {
		return
	}
	origin, routeSeq := s.activePath, s.routeSeq
	s.authorTarget = raw
	s.authorInfo.Loaded = false
	// Retain the follow page, but discard the previous author's content. Bump
	// each sequence before resetting so late pages cannot populate a new author.
	s.uploads.seq++
	s.uploads.Loading = false
	s.uploads.Loaded = false
	s.uploads.Data = model.VideoListData{}
	s.authorLists.seq++
	s.authorLists.Loading = false
	s.authorLists.Loaded = false
	s.authorLists.Data = model.SeasonsSeriesData{}
	s.seasonVideos.seq++
	s.seasonVideos.Loading = false
	s.seasonVideos.Loaded = false
	s.seriesVideos.seq++
	s.seriesVideos.Loading = false
	s.seriesVideos.Loaded = false
	s.authorTab = "uploads"
	s.selectedSeason = 0
	s.selectedSeries = 0
	s.authorPage = 0
	s.listPage = 0
	s.selected = make(map[string]*bool)
	load(s, &s.authorInfo, func() (model.UserInfoData, error) { return s.api.Info(raw) }, func(profile model.UserInfoData) {
		s.loadUploads(1, false)
		if navigate != nil && s.activePath == origin && s.routeSeq == routeSeq && profile.Mid > 0 {
			navigate(navigation.BilibiliAuthorPath(profile.Mid))
		}
	})
}
func (s *State) loadUploads(page int, more bool) {
	if s.uploads.Loading || s.authorInfo.Data.Mid == 0 {
		return
	}
	mid := s.authorInfo.Data.Mid
	if !more {
		key := fmt.Sprintf("author-videos/%d", mid)
		s.resetAutoPage(key, key)
	}
	previous := s.uploads.Data.List.Vlist
	load(s, &s.uploads, func() (model.VideoListData, error) { return s.api.VideoList(uint(mid), pageSize, page) }, func(data model.VideoListData) {
		if more {
			s.uploads.Data.List.Vlist = appendUnique(previous, data.List.Vlist, func(v model.VListItem) string { return v.Bvid })
		}
		s.authorPage = page
	})
}
func (s *State) loadAuthorLists(page int, more bool) {
	if s.authorLists.Loading || s.authorInfo.Data.Mid == 0 {
		return
	}
	if !more {
		s.resetAutoPage(fmt.Sprintf("author-lists/%d", s.authorInfo.Data.Mid), fmt.Sprintf("author-videos/%d", s.authorInfo.Data.Mid))
		s.authorLists.Data = model.SeasonsSeriesData{}
		s.authorLists.Loaded = false
		s.listPage = 0
		s.selectedSeason = 0
		s.selectedSeries = 0
		s.seasonVideos.seq++
		s.seasonVideos.Loading = false
		s.seasonVideos.Loaded = false
		s.seasonVideos.Data = model.SeasonsArchivesData{}
		s.seriesVideos.seq++
		s.seriesVideos.Loading = false
		s.seriesVideos.Loaded = false
		s.seriesVideos.Data = model.SeriesArchivesData{}
		s.selected = make(map[string]*bool)
	}
	mid := strconv.FormatUint(s.authorInfo.Data.Mid, 10)
	previous := s.authorLists.Data.ItemsLists
	load(s, &s.authorLists, func() (model.SeasonsSeriesData, error) { return s.api.SeasonsSeriesList(mid, pageSize, page) }, func(data model.SeasonsSeriesData) {
		if more {
			s.authorLists.Data.ItemsLists.SeasonsList = append(previous.SeasonsList, data.ItemsLists.SeasonsList...)
			s.authorLists.Data.ItemsLists.SeriesList = append(previous.SeriesList, data.ItemsLists.SeriesList...)
		}
		s.listPage = page
		if s.selectedSeason == 0 && s.selectedSeries == 0 && s.authorTab == "lists" {
			for _, item := range s.authorLists.Data.ItemsLists.SeasonsList {
				if item.Meta.SeasonId > 0 {
					s.loadSeason(item.Meta.SeasonId, 1, false)
					return
				}
			}
			for _, item := range s.authorLists.Data.ItemsLists.SeriesList {
				if item.Meta.SeriesId > 0 {
					s.loadSeries(item.Meta.SeriesId, 1, false)
					return
				}
			}
		}
	})
}
func (s *State) loadSeason(id, page int, more bool) {
	mid := strconv.FormatUint(s.authorInfo.Data.Mid, 10)
	if id == s.selectedSeason && s.seasonVideos.Loading {
		return
	}
	if !more {
		key := fmt.Sprintf("author-videos/%d", s.authorInfo.Data.Mid)
		s.resetAutoPage(key, key)
		s.seasonVideos.Data = model.SeasonsArchivesData{}
		s.seasonVideos.Loaded = false
		s.selectedSeason = id
		s.seasonPage = 0
		s.selectedSeries = 0
		s.selected = make(map[string]*bool)
	}
	previous := s.seasonVideos.Data.Archives
	load(s, &s.seasonVideos, func() (model.SeasonsArchivesData, error) { return s.api.SeasonsArchivesList(mid, pageSize, page, id) }, func(data model.SeasonsArchivesData) {
		if more {
			s.seasonVideos.Data.Archives = append(previous, data.Archives...)
		}
		s.seasonPage = page
	})
}
func (s *State) loadSeries(id, page int, more bool) {
	mid := strconv.FormatUint(s.authorInfo.Data.Mid, 10)
	if id == s.selectedSeries && s.seriesVideos.Loading {
		return
	}
	if !more {
		key := fmt.Sprintf("author-videos/%d", s.authorInfo.Data.Mid)
		s.resetAutoPage(key, key)
		s.seriesVideos.Data = model.SeriesArchivesData{}
		s.seriesVideos.Loaded = false
		s.selectedSeries = id
		s.seriesPage = 0
		s.selectedSeason = 0
		s.selected = make(map[string]*bool)
	}
	previous := s.seriesVideos.Data.Archives
	load(s, &s.seriesVideos, func() (model.SeriesArchivesData, error) { return s.api.SeriesList(mid, pageSize, page, id) }, func(data model.SeriesArchivesData) {
		if more {
			s.seriesVideos.Data.Archives = append(previous, data.Archives...)
		}
		s.seriesPage = page
	})
}
func (s *State) ensureLibrary() {
	if s.favoriteTab == "collection" {
		if !s.collections.Loaded && !s.collections.Loading && s.collections.Error == "" {
			s.loadCollections(1, false)
		}
	} else if !s.favoritesData.Loaded && !s.favoritesData.Loading && s.favoritesData.Error == "" {
		s.loadFavorites()
	}
}
func (s *State) loadFavorites() {
	if !s.loggedIn || s.favoritesData.Loading {
		return
	}
	load(s, &s.favoritesData, s.api.FavoritesList, func(data model.FavoritesData) {
		id := s.favoriteID
		if len(data.List) == 0 {
			s.favoriteID = 0
			s.favoriteDetail.seq++
			s.favoriteDetail.Loading = false
			s.favoriteDetail.Data = model.FavoriteData{}
			return
		}
		found := false
		for _, item := range data.List {
			if item.ID == id {
				found = true
				break
			}
		}
		if !found {
			id = data.List[0].ID
		}
		s.loadFavorite(id, 1, false)
	})
}
func (s *State) loadFavorite(id int64, page int, more bool) {
	if id == s.favoriteID && s.favoriteDetail.Loading && more {
		return
	}
	if !more {
		key := fmt.Sprintf("library/favorite/%d/0", id)
		s.resetAutoPage(key, key)
		s.favoriteID = id
		s.favoritePage = 0
		s.favoriteDetail.Data = model.FavoriteData{}
		s.favoriteDetail.Loaded = false
		s.selected = make(map[string]*bool)
	}
	previous := s.favoriteDetail.Data.Medias
	load(s, &s.favoriteDetail, func() (model.FavoriteData, error) { return s.api.Favorites(int(id), page, pageSize) }, func(data model.FavoriteData) {
		if more {
			s.favoriteDetail.Data.Medias = appendUnique(previous, data.Medias, func(v model.FavoriteMedias) int { return v.ID })
		}
		s.favoriteDetail.Data.HasMore = data.HasMore && len(data.Medias) > 0
		s.favoritePage = page
	})
}
func (s *State) loadCollections(page int, more bool) {
	if !s.loggedIn || s.collections.Loading {
		return
	}
	previous := s.collections.Data.List
	if !more {
		s.resetAutoPage("folders/collection", "folders/collection")
	}
	load(s, &s.collections, func() (model.CollectionData, error) { return s.api.Collection(page, pageSize) }, func(data model.CollectionData) {
		if more {
			s.collections.Data.List = appendUnique(previous, data.List, func(v model.CollectionDataList) int { return v.ID })
		}
		s.collectionPage = page
		if more {
			return
		}
		id := s.collectionID
		if len(data.List) == 0 {
			s.collectionID = 0
			s.collectionDetail.seq++
			s.collectionDetail.Loading = false
			s.collectionDetail.Data = model.CollectionItemData{}
			return
		}
		found := false
		for _, item := range data.List {
			if item.ID == id {
				found = true
				break
			}
		}
		if !found {
			id = data.List[0].ID
		}
		s.loadCollection(id, 1, false)
	})
}
func (s *State) loadCollection(id, page int, more bool) {
	if id == s.collectionID && s.collectionDetail.Loading && more {
		return
	}
	if !more {
		key := fmt.Sprintf("library/collection/0/%d", id)
		s.resetAutoPage(key, key)
		s.collectionID = id
		s.collectionDetailPage = 0
		s.collectionDetail.Data = model.CollectionItemData{}
		s.collectionDetail.Loaded = false
		s.selected = make(map[string]*bool)
	}
	previous := s.collectionDetail.Data.Medias
	load(s, &s.collectionDetail, func() (model.CollectionItemData, error) {
		return s.api.CollectionItem(strconv.Itoa(id), page, pageSize)
	}, func(data model.CollectionItemData) {
		if more {
			s.collectionDetail.Data.Medias = appendUnique(previous, data.Medias, func(v model.CollectionItemMedias) int64 { return v.ID })
		}
		total := s.collectionTotal()
		s.collectionHasMore = len(data.Medias) > 0 && (total == 0 && len(data.Medias) >= pageSize || len(s.collectionDetail.Data.Medias) < total)
		s.collectionDetailPage = page
	})
}

// ensureAuthor runs only for the active detail path, so outgoing transition
// views cannot start a request for an author the user already left.
func (s *State) ensureAuthor(mid uint64) {
	if s.authorInfo.Loaded && s.authorInfo.Data.Mid == mid {
		return
	}
	target := strconv.FormatUint(mid, 10)
	if s.authorTarget == target && (s.authorInfo.Loading || s.authorInfo.Error != "") {
		return
	}
	s.findAuthor(target, nil)
}
func (s *State) collectionTotal() int {
	if s.collectionDetail.Data.Info.MediaCount > 0 {
		return s.collectionDetail.Data.Info.MediaCount
	}
	for _, item := range s.collections.Data.List {
		if item.ID == s.collectionID {
			return item.MediaCount
		}
	}
	return 0
}

func appendUnique[T any, K comparable](previous, next []T, key func(T) K) []T {
	seen := make(map[K]bool, len(previous)+len(next))
	result := make([]T, 0, len(previous)+len(next))
	for _, items := range [][]T{previous, next} {
		for _, item := range items {
			value := key(item)
			if !seen[value] {
				seen[value] = true
				result = append(result, item)
			}
		}
	}
	return result
}
