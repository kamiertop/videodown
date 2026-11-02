package bilibili

import (
	"fmt"
	"strconv"

	bilidownload "github.com/kamiertop/videodown/bilibili/download"
)

func (s *State) downloadAllAuthor(navigate ...func(string)) {
	if !s.loggedIn || s.api == nil || s.authorInfo.Data.Mid == 0 {
		return
	}
	api, mid, name := s.api, s.authorInfo.Data.Mid, s.authorInfo.Data.Name
	kind, seasonID, seriesID := s.authorTab, s.selectedSeason, s.selectedSeries
	sourceName := s.authorSourceName()
	key, title := fmt.Sprintf("up:%d", mid), name+" · 全部投稿"
	if kind != "uploads" {
		if seasonID != 0 {
			key = fmt.Sprintf("season:%d:%d", mid, seasonID)
			title = name + " · " + sourceName
		} else if seriesID != 0 {
			key = fmt.Sprintf("series:%d:%d", mid, seriesID)
			title = name + " · " + sourceName
		} else {
			return
		}
	}
	total := s.uploads.Data.Page.Count
	if kind != "uploads" {
		if seasonID != 0 {
			total = max(s.seasonVideos.Data.Page.Total, s.seasonVideos.Data.Meta.Total)
		} else {
			total = s.seriesVideos.Data.Page.Total
		}
	}
	source := batchSource{Key: key, Title: title, Total: total, Load: func(page int) (batchPage, error) {
		result := batchPage{}
		switch {
		case kind == "uploads":
			data, err := api.VideoList(uint(mid), pageSize, page)
			if err != nil {
				return result, err
			}
			result.Total, result.Received = data.Page.Count, len(data.List.Vlist)
			for _, v := range data.List.Vlist {
				author := v.Author
				if author == "" {
					author = name
				}
				result.Items = append(result.Items, bilidownload.Task{Bvid: v.Bvid, Title: v.Title, Cover: v.Pic, UpperName: author, Play: v.Play, Danmaku: v.VideoReview, Pubtime: v.Created})
			}
		case seasonID != 0:
			data, err := api.SeasonsArchivesList(strconv.FormatUint(mid, 10), pageSize, page, seasonID)
			if err != nil {
				return result, err
			}
			result.Total, result.Received = max(data.Page.Total, data.Meta.Total), len(data.Archives)
			for _, v := range data.Archives {
				result.Items = append(result.Items, bilidownload.Task{Bvid: v.Bvid, Title: v.Title, Cover: v.Pic, UpperName: name, SourceName: sourceName, Duration: v.Duration, Play: v.Stat.View, Danmaku: v.Stat.Danmaku, Pubtime: v.Pubdate})
			}
		case seriesID != 0:
			data, err := api.SeriesList(strconv.FormatUint(mid, 10), pageSize, page, seriesID)
			if err != nil {
				return result, err
			}
			result.Total, result.Received = data.Page.Total, len(data.Archives)
			for _, v := range data.Archives {
				result.Items = append(result.Items, bilidownload.Task{Bvid: v.Bvid, Title: v.Title, Cover: v.Pic, UpperName: name, SourceName: sourceName, Duration: v.Duration, Play: v.Stat.View, Danmaku: v.Stat.Danmaku, Pubtime: v.PubDate})
			}
		}
		result.HasMore = result.Total > 0 || result.Received == pageSize
		return result, nil
	}}
	s.enqueueBatch(source, navigate...)
}

func (s *State) stageAllLibrary(navigate ...func(string)) {
	if !s.loggedIn || s.api == nil {
		return
	}
	api, tab, favID, colID := s.api, s.favoriteTab, s.favoriteID, s.collectionID
	name, total := s.libraryTitleAndTotal()
	if tab == "collection" && colID == 0 || tab != "collection" && favID == 0 {
		return
	}
	id := favID
	if tab == "collection" {
		id = int64(colID)
	}
	source := batchSource{Key: fmt.Sprintf("%s:%d", tab, id), Title: name, Total: total, Load: func(page int) (batchPage, error) {
		result := batchPage{}
		if tab == "collection" {
			data, err := api.CollectionItem(strconv.Itoa(colID), page, pageSize)
			if err != nil {
				return result, err
			}
			result.Total, result.Received = max(data.Info.MediaCount, total), len(data.Medias)
			for _, v := range data.Medias {
				result.Items = append(result.Items, bilidownload.Task{Bvid: v.Bvid, Title: v.Title, Cover: v.Cover, UpperName: v.Upper.Name, SourceName: name, Duration: v.Duration, Play: v.CntInfo.Play, Danmaku: v.CntInfo.Danmaku, Pubtime: v.Pubtime})
			}
			result.HasMore = result.Total > 0 || result.Received == pageSize
		} else {
			data, err := api.Favorites(int(favID), page, pageSize)
			if err != nil {
				return result, err
			}
			result.Total, result.Received, result.HasMore = data.Info.MediaCount, len(data.Medias), data.HasMore
			for _, v := range data.Medias {
				if v.Type != 2 || v.Attr == 1 || v.Attr == 9 {
					continue
				}
				result.Items = append(result.Items, bilidownload.Task{Bvid: favoriteBvid(v.Bvid, v.BvID), Title: v.Title, Cover: v.Cover, UpperName: v.Upper.Name, SourceName: name, Duration: v.Duration, Play: v.CntInfo.Play, Danmaku: v.CntInfo.Danmaku, Pubtime: v.PubTime})
			}
		}
		return result, nil
	}}
	s.enqueueBatch(source, navigate...)
}
