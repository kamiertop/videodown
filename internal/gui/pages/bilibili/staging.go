package bilibili

import (
	"strings"

	"github.com/kamiertop/videodown/bilibili/model"
)

func (s *State) stageKey(key string) bool {
	if !strings.HasPrefix(key, "BV") {
		return false
	}
	if s.stagedKeys[key] {
		s.notice = "该视频已在下载列表中"
		return false
	}
	s.stagedKeys[key] = true
	s.notice = ""
	return true
}
func (s *State) stageUpload(v model.VListItem) {
	if s.stageKey(v.Bvid) {
		s.stagedUploads = append(s.stagedUploads, v)
	}
}
func (s *State) stageFavorite(v model.FavoriteMedias) {
	if v.Type != 2 || v.Attr == 1 || v.Attr == 9 {
		return
	}
	if s.stageKey(favoriteBvid(v.Bvid, v.BvID)) {
		s.stagedFavorites = append(s.stagedFavorites, v)
	}
}
func (s *State) stageCollection(v model.CollectionItemMedias) {
	if s.stageKey(v.Bvid) {
		s.stagedCollections = append(s.stagedCollections, v)
	}
}
func (s *State) stageSeason(v model.SeasonsArchivesItem) {
	if s.stageKey(v.Bvid) {
		s.stagedSeasons = append(s.stagedSeasons, v)
	}
}
func (s *State) stageSeries(v model.SeriesArchivesItem) {
	if s.stageKey(v.Bvid) {
		s.stagedSeries = append(s.stagedSeries, v)
	}
}
func (s *State) stageDynamic(v model.DynamicArchiveItem) {
	if s.stageKey(v.Bvid) {
		s.stagedDynamics = append(s.stagedDynamics, v)
		s.stagedSourceNames[v.Bvid] = "关注动态"
	}
}
func (s *State) stagePlayHistory(v model.PlayHistoryItem) {
	if s.stageKey(v.History.Bvid) {
		s.stagedPlayHistory = append(s.stagedPlayHistory, v)
		s.stagedSourceNames[v.History.Bvid] = "B 站播放历史"
	}
}

// stageSelectedAuthorVideos keeps the original API records for the download
// page, where stream selection and the actual download are handled.
func (s *State) stageSelectedAuthorVideos() int {
	count := 0
	selected := func(bvid string) bool { value := s.selected[bvid]; return value != nil && *value }
	switch {
	case s.authorTab == "uploads":
		for _, video := range s.uploads.Data.List.Vlist {
			if selected(video.Bvid) {
				count++
				s.stageUpload(video)
			}
		}
	case s.selectedSeason != 0:
		for _, video := range s.seasonVideos.Data.Archives {
			if selected(video.Bvid) {
				count++
				s.stageSeason(video)
				s.stagedSourceNames[video.Bvid] = s.authorSourceName()
			}
		}
	case s.selectedSeries != 0:
		for _, video := range s.seriesVideos.Data.Archives {
			if selected(video.Bvid) {
				count++
				s.stageSeries(video)
				s.stagedSourceNames[video.Bvid] = s.authorSourceName()
			}
		}
	}
	return count
}

func (s *State) authorSourceName() string {
	if s.selectedSeason != 0 {
		return s.seasonVideos.Data.Meta.Title
	}
	if s.selectedSeries != 0 {
		for _, item := range s.authorLists.Data.ItemsLists.SeriesList {
			if item.Meta.SeriesId == s.selectedSeries {
				return item.Meta.Name
			}
		}
	}
	return ""
}
func (s *State) currentVideoCount() int {
	if s.favoriteTab == "collection" {
		return len(s.collectionDetail.Data.Medias)
	}
	return len(s.favoriteDetail.Data.Medias)
}
func (s *State) selectedCount() int {
	count := 0
	for _, key := range s.libraryVideoKeys() {
		if value := s.selected[key]; value != nil && *value {
			count++
		}
	}
	return count
}
func (s *State) selectLibrary() {
	query := *s.search(s.favoriteTab)
	if s.favoriteTab == "collection" {
		for _, v := range s.collectionDetail.Data.Medias {
			if matches(v.Title, query) && strings.HasPrefix(v.Bvid, "BV") {
				*s.selection(v.Bvid) = true
			}
		}
	} else {
		for _, v := range s.favoriteDetail.Data.Medias {
			bvid := favoriteBvid(v.Bvid, v.BvID)
			if matches(v.Title, query) && v.Type == 2 && v.Attr != 1 && v.Attr != 9 && strings.HasPrefix(bvid, "BV") {
				*s.selection(bvid) = true
			}
		}
	}
}
func (s *State) stageSelectedLibrary() {
	if s.favoriteTab == "collection" {
		for _, v := range s.collectionDetail.Data.Medias {
			if *s.selection(v.Bvid) {
				s.stageCollection(v)
			}
		}
	} else {
		for _, v := range s.favoriteDetail.Data.Medias {
			if *s.selection(favoriteBvid(v.Bvid, v.BvID)) {
				s.stageFavorite(v)
			}
		}
	}
}
