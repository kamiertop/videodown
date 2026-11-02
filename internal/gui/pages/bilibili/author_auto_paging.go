package bilibili

import (
	"fmt"

	"github.com/kamiertop/videodown/internal/gui/navigation"
)

func (s *State) maybeAutoAuthorVideos(mid uint64) {
	key := fmt.Sprintf("author-videos/%d", mid)
	active := func() bool {
		return s.activePath == navigation.BilibiliAuthorPath(mid) && s.authorInfo.Data.Mid == mid
	}
	switch {
	case s.authorTab == "uploads":
		s.maybeAutoPage(key, key, func() bool {
			return active() && s.authorTab == "uploads" && s.uploads.Loaded && !s.uploads.Loading && s.uploads.Error == "" &&
				len(s.uploads.Data.List.Vlist) < s.uploads.Data.Page.Count
		}, func() { s.loadUploads(s.authorPage+1, true) })
	case s.selectedSeason != 0:
		id := s.selectedSeason
		s.maybeAutoPage(key, key, func() bool {
			return active() && s.authorTab == "lists" && s.selectedSeason == id && s.seasonVideos.Loaded && !s.seasonVideos.Loading && s.seasonVideos.Error == "" &&
				len(s.seasonVideos.Data.Archives) < s.seasonVideos.Data.Page.Total
		}, func() { s.loadSeason(id, s.seasonPage+1, true) })
	case s.selectedSeries != 0:
		id := s.selectedSeries
		s.maybeAutoPage(key, key, func() bool {
			return active() && s.authorTab == "lists" && s.selectedSeries == id && s.seriesVideos.Loaded && !s.seriesVideos.Loading && s.seriesVideos.Error == "" &&
				len(s.seriesVideos.Data.Archives) < s.seriesVideos.Data.Page.Total
		}, func() { s.loadSeries(id, s.seriesPage+1, true) })
	}
}

func (s *State) maybeAutoAuthorLists(mid uint64) {
	if s.authorTab != "lists" {
		return
	}
	key := fmt.Sprintf("author-lists/%d", mid)
	scrollKey := fmt.Sprintf("author-videos/%d", mid)
	s.maybeAutoPage(key, scrollKey, func() bool {
		lists := s.authorLists.Data.ItemsLists
		return s.activePath == navigation.BilibiliAuthorPath(mid) && s.authorInfo.Data.Mid == mid && s.authorTab == "lists" &&
			s.authorLists.Loaded && !s.authorLists.Loading && s.authorLists.Error == "" &&
			len(lists.SeasonsList)+len(lists.SeriesList) < lists.Page.Total
	}, func() { s.loadAuthorLists(s.listPage+1, true) })
}
