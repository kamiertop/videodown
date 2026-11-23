package bilibili

import (
	"fmt"
	"strconv"
	"strings"

	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/gui/navigation"
)

const pageSize = 20

// followsScrollKey 是“UP 主”页关注列表面板的独立滚动区域，
// 与页面级滚动区分开，让顶部工具栏固定不动。
const followsScrollKey = "up-follows"

// clearPrivateData 退出登录后整块丢弃各板块状态：子结构直接整体重建，
// 页签（library.tab、author.tab）保留，导航同步时会校正。
func (s *State) clearPrivateData() {
	s.history = historyState{}
	s.follows = followsState{}
	s.author = authorState{tab: s.author.tab}
	s.library = libraryState{tab: s.library.tab}
	s.dynamic = dynamicState{picked: make(map[string]bool)}
	s.play = playState{picked: make(map[string]*bool)}
	s.stage = stageState{keys: make(map[string]bool), sourceNames: make(map[string]string)}
	s.queue = queueState{resolves: make(map[string]*queueResolve), progress: make(map[string]bilidownload.Progress), extraBusy: make(map[string]string), requested: make(map[string]bool)}
	s.manual = manualState{}
	s.batches = batchState{}
	s.autoPages = make(map[string]*autoPageState)
	s.selected = make(map[string]*bool)
	s.images = make(imagecache.Cache)
}
func (s *State) loadFollows(more bool) {
	if s.follows.list.Loading || !s.auth.loggedIn {
		return
	}
	const key = "author-follows"
	if !more {
		s.follows.page = 0
		s.resetAutoPage(key, followsScrollKey)
	}
	if more && len(s.follows.list.Data.List) >= s.follows.list.Data.Total {
		return
	}
	page := 1
	if more {
		page = s.follows.page + 1
	}
	old := s.follows.list.Data
	load(s, &s.follows.list, func() (model.FollowData, error) {
		d, err := s.api.FollowList(page, pageSize)
		if err == nil && more {
			d.List = mergeFollowPages(old.List, d.List)
		}
		return d, err
	}, func(model.FollowData) { s.follows.page = page })
}

// mergeFollowPages appends a fetched page to the loaded items. The list is
// ordered by follow time, which can shift between requests, so duplicates
// are dropped by mid.
func mergeFollowPages(oldItems, page []model.FollowDataList) []model.FollowDataList {
	seen := make(map[int64]bool, len(oldItems)+len(page))
	for _, v := range oldItems {
		seen[v.Mid] = true
	}
	items := append([]model.FollowDataList(nil), oldItems...)
	for _, v := range page {
		if !seen[v.Mid] {
			items = append(items, v)
			seen[v.Mid] = true
		}
	}
	return items
}
func (s *State) findAuthor(raw string, navigate func(string)) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !s.auth.loggedIn {
		return
	}
	origin, routeSeq := s.activePath, s.routeSeq
	s.author.target = raw
	s.author.info.Loaded = false
	// Retain the follow page, but discard the previous author's content. Bump
	// each sequence before resetting so late pages cannot populate a new author.
	s.author.uploads.seq++
	s.author.uploads.Loading = false
	s.author.uploads.Loaded = false
	s.author.uploads.Data = model.VideoListData{}
	s.author.lists.seq++
	s.author.lists.Loading = false
	s.author.lists.Loaded = false
	s.author.lists.Data = model.SeasonsSeriesData{}
	s.author.seasonVideos.seq++
	s.author.seasonVideos.Loading = false
	s.author.seasonVideos.Loaded = false
	s.author.seriesVideos.seq++
	s.author.seriesVideos.Loading = false
	s.author.seriesVideos.Loaded = false
	s.author.tab = "uploads"
	s.author.selectedSeason = 0
	s.author.selectedSeries = 0
	s.author.uploadsPage = 0
	s.author.listPage = 0
	s.selected = make(map[string]*bool)
	load(s, &s.author.info, func() (model.UserInfoData, error) { return s.api.Info(raw) }, func(profile model.UserInfoData) {
		s.loadUploads(1, false)
		if navigate != nil && s.activePath == origin && s.routeSeq == routeSeq && profile.Mid > 0 {
			navigate(navigation.BilibiliAuthorPath(profile.Mid))
		}
	})
}
func (s *State) loadUploads(page int, more bool) {
	if s.author.uploads.Loading || s.author.info.Data.Mid == 0 {
		return
	}
	mid := s.author.info.Data.Mid
	if !more {
		key := fmt.Sprintf("author-videos/%d", mid)
		s.resetAutoPage(key, key)
	}
	previous := s.author.uploads.Data.List.Vlist
	load(s, &s.author.uploads, func() (model.VideoListData, error) { return s.api.VideoList(uint(mid), pageSize, page) }, func(data model.VideoListData) {
		if more {
			s.author.uploads.Data.List.Vlist = appendUnique(previous, data.List.Vlist, func(v model.VListItem) string { return v.Bvid })
		}
		s.author.uploadsPage = page
	})
}
func (s *State) loadAuthorLists(page int, more bool) {
	if s.author.lists.Loading || s.author.info.Data.Mid == 0 {
		return
	}
	if !more {
		s.resetAutoPage(fmt.Sprintf("author-lists/%d", s.author.info.Data.Mid), fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid))
		s.author.lists.Data = model.SeasonsSeriesData{}
		s.author.lists.Loaded = false
		s.author.listPage = 0
		s.author.selectedSeason = 0
		s.author.selectedSeries = 0
		s.author.seasonVideos.seq++
		s.author.seasonVideos.Loading = false
		s.author.seasonVideos.Loaded = false
		s.author.seasonVideos.Data = model.SeasonsArchivesData{}
		s.author.seriesVideos.seq++
		s.author.seriesVideos.Loading = false
		s.author.seriesVideos.Loaded = false
		s.author.seriesVideos.Data = model.SeriesArchivesData{}
		s.selected = make(map[string]*bool)
	}
	mid := strconv.FormatUint(s.author.info.Data.Mid, 10)
	previous := s.author.lists.Data.ItemsLists
	load(s, &s.author.lists, func() (model.SeasonsSeriesData, error) { return s.api.SeasonsSeriesList(mid, pageSize, page) }, func(data model.SeasonsSeriesData) {
		if more {
			s.author.lists.Data.ItemsLists.SeasonsList = append(previous.SeasonsList, data.ItemsLists.SeasonsList...)
			s.author.lists.Data.ItemsLists.SeriesList = append(previous.SeriesList, data.ItemsLists.SeriesList...)
		}
		s.author.listPage = page
		if s.author.selectedSeason == 0 && s.author.selectedSeries == 0 && s.author.tab == "lists" {
			for _, item := range s.author.lists.Data.ItemsLists.SeasonsList {
				if item.Meta.SeasonId > 0 {
					s.loadSeason(item.Meta.SeasonId, 1, false)
					return
				}
			}
			for _, item := range s.author.lists.Data.ItemsLists.SeriesList {
				if item.Meta.SeriesId > 0 {
					s.loadSeries(item.Meta.SeriesId, 1, false)
					return
				}
			}
		}
	})
}
func (s *State) loadSeason(id, page int, more bool) {
	mid := strconv.FormatUint(s.author.info.Data.Mid, 10)
	if id == s.author.selectedSeason && s.author.seasonVideos.Loading {
		return
	}
	if !more {
		key := fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid)
		s.resetAutoPage(key, key)
		s.author.seasonVideos.Data = model.SeasonsArchivesData{}
		s.author.seasonVideos.Loaded = false
		s.author.selectedSeason = id
		s.author.seasonPage = 0
		s.author.selectedSeries = 0
		s.selected = make(map[string]*bool)
	}
	previous := s.author.seasonVideos.Data.Archives
	load(s, &s.author.seasonVideos, func() (model.SeasonsArchivesData, error) { return s.api.SeasonsArchivesList(mid, pageSize, page, id) }, func(data model.SeasonsArchivesData) {
		if more {
			s.author.seasonVideos.Data.Archives = append(previous, data.Archives...)
		}
		s.author.seasonPage = page
	})
}

func (s *State) loadSeries(id, page int, more bool) {
	mid := strconv.FormatUint(s.author.info.Data.Mid, 10)
	if id == s.author.selectedSeries && s.author.seriesVideos.Loading {
		return
	}
	if !more {
		key := fmt.Sprintf("author-videos/%d", s.author.info.Data.Mid)
		s.resetAutoPage(key, key)
		s.author.seriesVideos.Data = model.SeriesArchivesData{}
		s.author.seriesVideos.Loaded = false
		s.author.selectedSeries = id
		s.author.seriesPage = 0
		s.author.selectedSeason = 0
		s.selected = make(map[string]*bool)
	}
	previous := s.author.seriesVideos.Data.Archives
	load(s, &s.author.seriesVideos, func() (model.SeriesArchivesData, error) { return s.api.SeriesList(mid, pageSize, page, id) }, func(data model.SeriesArchivesData) {
		if more {
			s.author.seriesVideos.Data.Archives = append(previous, data.Archives...)
		}
		s.author.seriesPage = page
	})
}
func (s *State) ensureLibrary() {
	if s.library.tab == "collection" {
		if !s.library.collections.Loaded && !s.library.collections.Loading && s.library.collections.Error == "" {
			s.loadCollections(1, false)
		}
	} else if !s.library.folders.Loaded && !s.library.folders.Loading && s.library.folders.Error == "" {
		s.loadFavorites()
	}
}
func (s *State) loadFavorites() {
	if !s.auth.loggedIn || s.library.folders.Loading {
		return
	}
	load(s, &s.library.folders, s.api.FavoritesList, func(data model.FavoritesData) {
		id := s.library.folderID
		if len(data.List) == 0 {
			s.library.folderID = 0
			s.library.folderDetail.seq++
			s.library.folderDetail.Loading = false
			s.library.folderDetail.Data = model.FavoriteData{}
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
	if id == s.library.folderID && s.library.folderDetail.Loading && more {
		return
	}
	if !more {
		key := fmt.Sprintf("library/favorite/%d/0", id)
		s.resetAutoPage(key, key)
		s.library.folderID = id
		s.library.folderPage = 0
		s.library.folderDetail.Data = model.FavoriteData{}
		s.library.folderDetail.Loaded = false
		s.selected = make(map[string]*bool)
	}
	previous := s.library.folderDetail.Data.Medias
	load(s, &s.library.folderDetail, func() (model.FavoriteData, error) { return s.api.Favorites(int(id), page, pageSize) }, func(data model.FavoriteData) {
		if more {
			s.library.folderDetail.Data.Medias = appendUnique(previous, data.Medias, func(v model.FavoriteMedias) int { return v.ID })
		}
		s.library.folderDetail.Data.HasMore = data.HasMore && len(data.Medias) > 0
		s.library.folderPage = page
	})
}
func (s *State) loadCollections(page int, more bool) {
	if !s.auth.loggedIn || s.library.collections.Loading {
		return
	}
	previous := s.library.collections.Data.List
	if !more {
		s.resetAutoPage("folders/collection", "folders/collection")
	}
	load(s, &s.library.collections, func() (model.CollectionData, error) { return s.api.Collection(page, pageSize) }, func(data model.CollectionData) {
		if more {
			s.library.collections.Data.List = appendUnique(previous, data.List, func(v model.CollectionDataList) int { return v.ID })
		}
		s.library.collectionPage = page
		if more {
			return
		}
		id := s.library.collectionID
		if len(data.List) == 0 {
			s.library.collectionID = 0
			s.library.collectionDetail.seq++
			s.library.collectionDetail.Loading = false
			s.library.collectionDetail.Data = model.CollectionItemData{}
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
	if id == s.library.collectionID && s.library.collectionDetail.Loading && more {
		return
	}
	if !more {
		key := fmt.Sprintf("library/collection/0/%d", id)
		s.resetAutoPage(key, key)
		s.library.collectionID = id
		s.library.collectionDetailPage = 0
		s.library.collectionDetail.Data = model.CollectionItemData{}
		s.library.collectionDetail.Loaded = false
		s.selected = make(map[string]*bool)
	}
	previous := s.library.collectionDetail.Data.Medias
	load(s, &s.library.collectionDetail, func() (model.CollectionItemData, error) {
		return s.api.CollectionItem(strconv.Itoa(id), page, pageSize)
	}, func(data model.CollectionItemData) {
		if more {
			s.library.collectionDetail.Data.Medias = appendUnique(previous, data.Medias, func(v model.CollectionItemMedias) int64 { return v.ID })
		}
		total := s.collectionTotal()
		s.library.collectionHasMore = len(data.Medias) > 0 && (total == 0 && len(data.Medias) >= pageSize || len(s.library.collectionDetail.Data.Medias) < total)
		s.library.collectionDetailPage = page
	})
}

// ensureAuthor runs only for the active detail path, so outgoing transition
// views cannot start a request for an author the user already left.
func (s *State) ensureAuthor(mid uint64) {
	if s.author.info.Loaded && s.author.info.Data.Mid == mid {
		return
	}
	target := strconv.FormatUint(mid, 10)
	if s.author.target == target && (s.author.info.Loading || s.author.info.Error != "") {
		return
	}
	s.findAuthor(target, nil)
}
func (s *State) collectionTotal() int {
	if s.library.collectionDetail.Data.Info.MediaCount > 0 {
		return s.library.collectionDetail.Data.Info.MediaCount
	}
	for _, item := range s.library.collections.Data.List {
		if item.ID == s.library.collectionID {
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
