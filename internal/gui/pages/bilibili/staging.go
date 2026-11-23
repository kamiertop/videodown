package bilibili

import (
	"strings"

	"github.com/kamiertop/videodown/bilibili/model"
)

// stageState 是下载页的待下载暂存队列：按来源保留原始 API 记录（供画质
// 选择与下载使用），keys 维持入队顺序，sourceNames 记录来源名称。
type stageState struct {
	notice, error string
	keys          map[string]bool
	sourceNames   map[string]string
	uploads       []model.VListItem
	favorites     []model.FavoriteMedias
	collections   []model.CollectionItemMedias
	seasons       []model.SeasonsArchivesItem
	series        []model.SeriesArchivesItem
	dynamics      []model.DynamicArchiveItem
	playHistory   []model.PlayHistoryItem
	manual        []manualVideo
}

func (s *State) stageKey(key string) bool {
	if !strings.HasPrefix(key, "BV") {
		return false
	}
	if s.stage.keys[key] {
		s.stage.notice = "该视频已在下载列表中"
		return false
	}
	s.stage.keys[key] = true
	s.stage.notice = ""
	return true
}
func (s *State) stageUpload(v model.VListItem) {
	if s.stageKey(v.Bvid) {
		s.stage.uploads = append(s.stage.uploads, v)
	}
}
func (s *State) stageFavorite(v model.FavoriteMedias) {
	if v.Type != 2 || v.Attr == 1 || v.Attr == 9 {
		return
	}
	if s.stageKey(favoriteBvid(v.Bvid, v.BvID)) {
		s.stage.favorites = append(s.stage.favorites, v)
	}
}
func (s *State) stageCollection(v model.CollectionItemMedias) {
	if s.stageKey(v.Bvid) {
		s.stage.collections = append(s.stage.collections, v)
	}
}
func (s *State) stageSeason(v model.SeasonsArchivesItem) {
	if s.stageKey(v.Bvid) {
		s.stage.seasons = append(s.stage.seasons, v)
	}
}
func (s *State) stageSeries(v model.SeriesArchivesItem) {
	if s.stageKey(v.Bvid) {
		s.stage.series = append(s.stage.series, v)
	}
}
func (s *State) stageDynamic(v model.DynamicArchiveItem) {
	if s.stageKey(v.Bvid) {
		s.stage.dynamics = append(s.stage.dynamics, v)
		s.stage.sourceNames[v.Bvid] = "关注动态"
	}
}
func (s *State) stagePlayHistory(v model.PlayHistoryItem) {
	if s.stageKey(v.History.Bvid) {
		s.stage.playHistory = append(s.stage.playHistory, v)
		s.stage.sourceNames[v.History.Bvid] = "B 站播放历史"
	}
}

// stageSelectedAuthorVideos keeps the original API records for the download
// page, where stream selection and the actual download are handled.
func (s *State) stageSelectedAuthorVideos() int {
	count := 0
	selected := func(bvid string) bool { value := s.selected[bvid]; return value != nil && *value }
	switch {
	case s.author.tab == "uploads":
		for _, video := range s.author.uploads.Data.List.Vlist {
			if selected(video.Bvid) {
				count++
				s.stageUpload(video)
			}
		}
	case s.author.selectedSeason != 0:
		for _, video := range s.author.seasonVideos.Data.Archives {
			if selected(video.Bvid) {
				count++
				s.stageSeason(video)
				s.stage.sourceNames[video.Bvid] = s.authorSourceName()
			}
		}
	case s.author.selectedSeries != 0:
		for _, video := range s.author.seriesVideos.Data.Archives {
			if selected(video.Bvid) {
				count++
				s.stageSeries(video)
				s.stage.sourceNames[video.Bvid] = s.authorSourceName()
			}
		}
	}

	return count
}

func (s *State) authorSourceName() string {
	if s.author.selectedSeason != 0 {
		return s.author.seasonVideos.Data.Meta.Title
	}
	if s.author.selectedSeries != 0 {
		for _, item := range s.author.lists.Data.ItemsLists.SeriesList {
			if item.Meta.SeriesId == s.author.selectedSeries {
				return item.Meta.Name
			}
		}
	}

	return ""
}
func (s *State) currentVideoCount() int {
	if s.library.tab == "collection" {
		return len(s.library.collectionDetail.Data.Medias)
	}
	return len(s.library.folderDetail.Data.Medias)
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
	query := *s.search(s.library.tab)
	if s.library.tab == "collection" {
		for _, v := range s.library.collectionDetail.Data.Medias {
			if matches(v.Title, query) && strings.HasPrefix(v.Bvid, "BV") {
				*s.selection(v.Bvid) = true
			}
		}
	} else {
		for _, v := range s.library.folderDetail.Data.Medias {
			bvid := favoriteBvid(v.Bvid, v.BvID)
			if matches(v.Title, query) && v.Type == 2 && v.Attr != 1 && v.Attr != 9 && strings.HasPrefix(bvid, "BV") {
				*s.selection(bvid) = true
			}
		}
	}
}

func (s *State) stageSelectedLibrary() {
	if s.library.tab == "collection" {
		for _, v := range s.library.collectionDetail.Data.Medias {
			if *s.selection(v.Bvid) {
				s.stageCollection(v)
			}
		}
	} else {
		for _, v := range s.library.folderDetail.Data.Medias {
			if *s.selection(favoriteBvid(v.Bvid, v.BvID)) {
				s.stageFavorite(v)
			}
		}
	}
}
