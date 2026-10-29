package douyin

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kamiertop/videodown/douyin/model"
)

const pageSize = 20

func (s *State) useSource(src source) {
	s.sourceVersion++
	s.source = src
	s.videos.seq++
	s.more.seq++
	s.videos = resource[videoPage]{seq: s.videos.seq}
	s.more = resource[videoPage]{seq: s.more.seq}
	s.selected = map[string]bool{}
	s.autoPending = false
	s.autoAt = 0
	s.autoFill = 0
	s.scroll("videos").Y = 0
	s.loadVideos(false)
}
func (s *State) loadVideos(more bool) {
	if s.source.Load == nil || s.videos.Loading || s.more.Loading {
		return
	}
	loader := s.source.Load
	if !more {
		load(s, &s.videos, func() (videoPage, error) { return loader(0) }, nil)
		return
	}
	if !s.videos.Data.More {
		return
	}
	cursor, version := s.videos.Data.Next, s.sourceVersion
	load(s, &s.more, func() (videoPage, error) { return loader(cursor) }, func(page videoPage) {
		if version != s.sourceVersion {
			return
		}
		seen := map[string]bool{}
		for _, v := range s.videos.Data.Items {
			seen[v.AwemeId] = true
		}
		for _, v := range page.Items {
			if v.AwemeId != "" && !seen[v.AwemeId] {
				seen[v.AwemeId] = true
				s.videos.Data.Items = append(s.videos.Data.Items, v)
			}
		}
		s.videos.Data.Next, s.videos.Data.More = page.Next, page.More && page.Next != cursor && len(page.Items) > 0
	})
}
func (s *State) baseSource(kind string) source {
	api := s.api
	return source{Key: kind,
		Title: map[string]string{"favorite": "收藏视频", "dynamic": "关注动态", "play-history": "播放历史"}[kind],
		Load: func(cursor int) (videoPage, error) {
			out := videoPage{}
			switch kind {
			case "favorite":
				d, err := api.FavoriteVideo(pageSize, uint(cursor))
				out = videoPage{d.AwemeList, int(d.Cursor), d.HasMore != 0}
				if err == nil {
					err = status(d.StatusCode)
				}
				return out, err
			case "dynamic":
				d, err := api.Dynamic(cursor)
				if err == nil {
					err = status(d.StatusCode)
				}
				for _, v := range d.Data {
					out.Items = append(out.Items, v.Aweme)
				}
				out.Next, out.More = int(d.Cursor), d.HasMore != 0
				return out, err
			case "play-history":
				d, err := api.History(cursor, 0, 0, 0)
				if err == nil {
					err = status(d.StatusCode)
				}
				for _, v := range d.AwemeList {
					out.Items = append(out.Items, model.AwemeItem{AwemeId: v.AwemeId, Desc: v.Desc, Caption: v.Caption, ItemTitle: v.ItemTitle, CreateTime: v.CreateTime, Author: v.Author, Images: v.Images, Video: v.Video, Music: v.Music, Statistics: v.Statistics, MediaType: v.MediaType, IsLivePhoto: v.IsLivePhoto, IsSliedes: v.IsSliedes, AwemeType: v.AwemeType, Duration: int(v.Video.Duration)})
				}
				out.Next, out.More = int(d.MaxCursor), d.HasMore != 0
				return out, err
			}
			return out, errors.New("未知视频来源")
		}}
}
func (s *State) authorSource(uid, name string) source {
	api := s.api
	return source{Key: "user:" + uid, Title: name + " · 全部作品", Load: func(cursor int) (videoPage, error) {
		d, err := api.UserVideoList(uid, pageSize, cursor)
		if err == nil {
			err = status(d.StatusCode)
		}
		return videoPage{d.AwemeList, int(d.MaxCursor), d.HasMore != 0}, err
	}}
}
func (s *State) folderSource(kind string, f folder) source {
	api := s.api
	return source{Key: kind + ":" + f.Key, Title: f.Title, Load: func(cursor int) (videoPage, error) {
		if kind == "folder" {
			d, err := api.FavoritesVideoList(f.Key, cursor, pageSize)
			if err == nil {
				err = status(d.StatusCode)
			}
			return videoPage{d.AwemeList, int(d.Cursor), d.HasMore != 0}, err
		}
		d, err := api.CollectionList(f.Author, f.Key, cursor, pageSize)
		if err == nil {
			err = status(d.StatusCode)
		}
		return videoPage{d.AwemeList, d.MaxCursor, d.HasMore != 0}, err
	}}
}
func (s *State) loadFollows(more bool) {
	if s.api == nil || s.follows.Loading {
		return
	}
	offset := 0
	if !more {
		s.followGeneration++
		s.followAutoPending = false
		s.scroll("authors").Y = 0
	}
	if more && (!s.follows.Data.HasMore || strings.TrimSpace(s.followSearchInput) != "") {
		return
	}
	if more {
		offset = s.follows.Data.Offset
	}
	old := s.follows.Data
	api := s.api
	load(s, &s.follows, func() (model.FollowResponse, error) {
		d, err := api.FollowList(offset)
		if err == nil {
			err = status(d.StatusCode)
		}
		if more {
			seen := map[string]bool{}
			for _, v := range old.Followings {
				seen[v.SecUid] = true
			}
			items := append([]model.FollowItem(nil), old.Followings...)
			for _, v := range d.Followings {
				if !seen[v.SecUid] {
					items = append(items, v)
					seen[v.SecUid] = true
				}
			}
			d.Followings = items
			d.HasMore = d.HasMore && d.Offset != offset && len(items) > len(old.Followings)
		}
		return d, err
	}, nil)
}
func (s *State) findAuthor(navigate func(string)) {
	raw := strings.TrimSpace(s.authorInput)
	if raw == "" || s.api == nil || s.runtime.Run == nil {
		return
	}
	api, epoch, origin := s.api, s.epoch, s.activePath
	s.authorSeq++
	seq := s.authorSeq
	s.actionError = ""
	var uid string
	s.runtime.Run(func() (err error) {
		s.apiMu.Lock()
		defer s.apiMu.Unlock()
		uid, err = api.ParseSecUserID(raw)
		return err
	}, func(err error) {
		if s.stopped || s.epoch != epoch || s.authorSeq != seq || s.activePath != origin {
			return
		}
		if err != nil {
			s.actionError = err.Error()
			return
		}
		navigate("/douyin/user/" + url.PathEscape(uid))
	})
}
func (s *State) loadAuthor(uid string) {
	s.useSource(source{})
	s.author.Data = model.User{}
	s.folders.seq++
	s.folders.Loaded = false
	s.authorTab = "posts"
	api := s.api
	load(s, &s.author, func() (model.User, error) {
		d, err := api.User(uid)
		if err == nil {
			err = status(d.StatusCode)
		}
		return d.User, err
	}, func(user model.User) {
		if user.SecUid == "" {
			user.SecUid = uid
			s.author.Data = user
		}
		s.useSource(s.authorSource(uid, user.NickName))
	})
}
func (s *State) prepareFolders(kind, uid string) {
	s.folderVersion++
	s.folderAutoPending = false
	api := s.api
	s.folders.seq++
	s.folders = resource[[]folder]{seq: s.folders.seq}
	s.folderCursor = 0
	s.folderMore = false
	s.folderNext = func(cursor int) ([]folder, int, bool, error) {
		items := []folder{}
		switch kind {
		case "folder":
			d, err := api.CollectList(cursor)
			if err == nil {
				err = status(d.StatusCode)
			}
			for _, v := range d.CollectsList {
				id := v.CollectsIDStr
				if id == "" {
					id = strconv.FormatUint(uint64(v.CollectsID), 10)
				}
				items = append(items, folder{Key: id, Title: v.CollectsName, Count: v.TotalNumber})
			}
			return items, d.Cursor, d.HasMore, err
		case "series":
			d, err := api.UserSeries(uid, cursor, pageSize)
			if err == nil {
				err = status(d.StatusCode)
			}
			for _, v := range d.SeriesInfos {
				items = append(items, folder{Key: v.SeriesID, Title: v.SeriesName, Author: uid, Count: max(v.Stats.TotalEpisode, v.Stats.UpdatedToEpisode)})
			}
			return items, d.Cursor, d.HasMore != 0, err
		default:
			d, err := api.Collection(pageSize, cursor)
			if err == nil {
				err = status(d.StatusCode)
			}
			for _, v := range d.MixInfos {
				items = append(items, folder{Key: v.MixId, Title: v.MixName, Author: v.Author.SecUid, Count: v.Statis.UpdatedToEpisode})
			}
			return items, d.Cursor, d.HasMore != 0, err
		}
	}
	s.loadFolders(false)
}
func (s *State) loadFolders(more bool) {
	if s.folderNext == nil || s.folders.Loading {
		return
	}
	cursor := 0
	if !more {
		s.folderVersion++
		s.folderAutoPending = false
		s.scroll("folders").Y = 0
	}
	if more && !s.folderMore {
		return
	}
	if more {
		cursor = s.folderCursor
	}
	old := append([]folder(nil), s.folders.Data...)
	oldCount := len(old)
	loader, version := s.folderNext, s.folderVersion
	var next int
	var hasMore bool
	load(s, &s.folders, func() ([]folder, error) {
		items, n, m, err := loader(cursor)
		next, hasMore = n, m
		if len(items) == 0 {
			hasMore = false
		}
		if more {
			seen := map[string]bool{}
			for _, v := range old {
				seen[v.Key] = true
			}
			for _, v := range items {
				if !seen[v.Key] {
					old = append(old, v)
					seen[v.Key] = true
				}
			}
			if len(old) == oldCount {
				hasMore = false
			}
			items = old
		}
		return items, err
	}, func(items []folder) {
		if version != s.folderVersion {
			return
		}
		s.folderCursor, s.folderMore = next, hasMore && next != cursor
		if s.pendingFolder != "" {
			for _, f := range items {
				if f.Key == s.pendingFolder {
					s.pendingFolder = ""
					kind := s.libraryTab
					if s.authorTab == "series" && section(s.activePath) == "user" {
						kind = "collection"
					}
					s.useSource(s.folderSource(kind, f))
					break
				}
			}
			if s.pendingFolder != "" && s.folderMore {
				s.loadFolders(true)
			}
		}
	})
}
func (s *State) maybeMore() {
	scroll := s.scroll("videos")
	if s.runtime.After == nil || s.autoPending || !s.videos.Data.More || s.videos.Loading || s.more.Loading || s.more.Error != "" {
		return
	}
	if scroll.MaxY > 0 && (scroll.Y <= s.autoAt+1 || scroll.Y+120 < scroll.MaxY) {
		return
	}
	if scroll.MaxY <= 0 && s.autoFill >= 6 {
		return
	}
	s.autoPending = true
	epoch, version, path := s.epoch, s.sourceVersion, s.activePath
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || s.sourceVersion != version || s.activePath != path {
			return
		}
		s.autoPending = false
		if scroll.MaxY <= 0 {
			s.autoFill++
		}
		s.autoAt = scroll.Y
		s.loadVideos(true)
	})
}
func (s *State) enter(path string) {
	if path == s.activePath {
		return
	}
	s.activePath = path
	s.followGeneration++
	s.followAutoPending = false
	if s.cancelFollowSearch != nil {
		s.cancelFollowSearch()
		s.cancelFollowSearch = nil
	}
	s.followSearchPending = false
	s.followResults.seq++
	s.followResults.Loading = false
	s.followQuery = ""
	s.authorSeq++
	s.author.seq++
	s.author.Loading = false
	s.folders.seq++
	s.folders.Loading = false
	s.folderVersion++
	s.folderAutoPending = false
	s.pendingFolder = ""
	s.autoPending = false
	if s.api == nil {
		return
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	section := "download"
	if len(parts) > 1 {
		section = parts[1]
	}
	if section == "history" {
		s.loadHistory(false)
		return
	}
	if !s.loggedIn() {
		return
	}
	switch section {
	case "user":
		if len(parts) > 2 {
			uid, err := url.PathUnescape(parts[2])
			if err == nil {
				s.loadAuthor(uid)
			}
		} else if !s.follows.Loaded && !s.follows.Loading {
			s.loadFollows(false)
		}
	case "dynamic", "play-history":
		s.useSource(s.baseSource(section))
	case "favorite":
		s.pendingFolder = ""
		s.libraryTab = "favorite"
		if len(parts) >= 3 && (parts[2] == "folder" || parts[2] == "collection") {
			s.libraryTab = parts[2]
			if len(parts) > 3 {
				s.pendingFolder, _ = url.PathUnescape(parts[3])
			}
			s.useSource(source{})
			s.prepareFolders(s.libraryTab, "")
		} else {
			s.useSource(s.baseSource("favorite"))
		}
	}
}
func (s *State) loggedIn() bool {
	return s.profile.Loaded && s.profile.Error == "" && (s.profile.Data.Uid != "" || s.profile.Data.SecUid != "")
}
func duration(seconds int) string { return fmt.Sprintf("%d:%02d", seconds/60, seconds%60) }
