package bilibili

import (
	"fmt"

	"github.com/kamiertop/videodown/internal/gui/navigation"
)

func (s *State) maybeAutoAuthorVideos(mid uint64) {
	key := fmt.Sprintf("author-videos/%d", mid)
	active := func() bool {
		return s.activePath == navigation.BilibiliAuthorPath(mid) && s.author.info.Data.Mid == mid
	}
	switch {
	case s.author.tab == "uploads":
		s.maybeAutoPage(key, key, func() bool {
			return active() && s.author.tab == "uploads" && s.author.uploads.Loaded && !s.author.uploads.Loading && s.author.uploads.Error == "" &&
				len(s.author.uploads.Data.List.Vlist) < s.author.uploads.Data.Page.Count
		}, func() { s.loadUploads(s.author.uploadsPage+1, true) })
	case s.author.selectedSeason != 0:
		id := s.author.selectedSeason
		s.maybeAutoPage(key, key, func() bool {
			return active() && s.author.tab == "lists" && s.author.selectedSeason == id && s.author.seasonVideos.Loaded && !s.author.seasonVideos.Loading && s.author.seasonVideos.Error == "" &&
				len(s.author.seasonVideos.Data.Archives) < s.author.seasonVideos.Data.Page.Total
		}, func() { s.loadSeason(id, s.author.seasonPage+1, true) })
	case s.author.selectedSeries != 0:
		id := s.author.selectedSeries
		s.maybeAutoPage(key, key, func() bool {
			return active() && s.author.tab == "lists" && s.author.selectedSeries == id && s.author.seriesVideos.Loaded && !s.author.seriesVideos.Loading && s.author.seriesVideos.Error == "" &&
				len(s.author.seriesVideos.Data.Archives) < s.author.seriesVideos.Data.Page.Total
		}, func() { s.loadSeries(id, s.author.seriesPage+1, true) })
	}
}

func (s *State) maybeAutoAuthorLists(mid uint64) {
	if s.author.tab != "lists" {
		return
	}
	key := fmt.Sprintf("author-lists/%d", mid)
	scrollKey := fmt.Sprintf("author-videos/%d", mid)
	s.maybeAutoPage(key, scrollKey, func() bool {
		lists := s.author.lists.Data.ItemsLists
		return s.activePath == navigation.BilibiliAuthorPath(mid) && s.author.info.Data.Mid == mid && s.author.tab == "lists" &&
			s.author.lists.Loaded && !s.author.lists.Loading && s.author.lists.Error == "" &&
			len(lists.SeasonsList)+len(lists.SeriesList) < lists.Page.Total
	}, func() { s.loadAuthorLists(s.author.listPage+1, true) })
}

// maybeMoreAuthorFollows keeps appending the follow grid as the UP 主 page
// scrolls near its end, replacing the manual pager that jumped the layout.
func (s *State) maybeMoreAuthorFollows() {
	const key = "author-follows"
	s.maybeAutoPage(key, followsScrollKey, func() bool {
		return s.activePath == "/bilibili/up" &&
			s.follows.list.Loaded && !s.follows.list.Loading && s.follows.list.Error == "" &&
			len(s.follows.list.Data.List) < s.follows.list.Data.Total
	}, func() { s.loadFollows(true) })
}
