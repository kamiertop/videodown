package douyin

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	dyapi "github.com/kamiertop/videodown/douyin/api"
	"github.com/kamiertop/videodown/douyin/model"
)

const pageSize = 20

// openList 让 ls 展示 src：同来源已加载（或加载中）则原样保留，
// 切走再回来零请求；来源变化或上次失败时重置重截。
func (s *State) openList(ls *listSection, src source) {
	if ls.src.Key == src.Key && ls.videos.Error == "" && (ls.videos.Loaded || ls.videos.Loading) {
		return
	}
	s.resetList(ls, src)
}

// resetList 清空 ls 并请求第一页；零值 src 仅清空（目录视图）。
func (s *State) resetList(ls *listSection, src source) {
	ls.src = src
	ls.videos.seq++
	ls.more.seq++
	ls.videos = resource[videoPage]{seq: ls.videos.seq}
	ls.more = resource[videoPage]{seq: ls.more.seq}
	ls.autoPending = false
	ls.autoAt, ls.autoFill = 0, 0
	s.selected = map[string]bool{}
	s.scroll(ls.scrollKey).Y = 0
	s.loadList(ls, false)
}

func (s *State) loadList(ls *listSection, more bool) {
	if ls.src.Load == nil || ls.videos.Loading || ls.more.Loading {
		return
	}
	loader := ls.src.Load
	if !more {
		load(s, &ls.videos, func() (videoPage, error) { return loader(0) }, nil)
		return
	}
	if !ls.videos.Data.More {
		return
	}
	cursor, seq := ls.videos.Data.Next, ls.videos.seq
	load(s, &ls.more, func() (videoPage, error) { return loader(cursor) }, func(page videoPage) {
		// 请求期间列表被重置（换了来源）：丢弃迟到页。
		if ls.videos.seq != seq {
			return
		}
		seen := map[string]bool{}
		for _, v := range ls.videos.Data.Items {
			seen[v.AwemeId] = true
		}
		for _, v := range page.Items {
			if v.AwemeId != "" && !seen[v.AwemeId] {
				seen[v.AwemeId] = true
				ls.videos.Data.Items = append(ls.videos.Data.Items, v)
			}
		}
		ls.videos.Data.Next, ls.videos.Data.More = page.Next, page.More && page.Next != cursor && len(page.Items) > 0
	})
}

// clearLists 清空所有板块列表（保留滚动 key），用于账号切换。
func (s *State) clearLists() {
	for _, ls := range s.allLists() {
		ls.src = source{}
		ls.videos = resource[videoPage]{}
		ls.more = resource[videoPage]{}
		ls.autoPending = false
		ls.autoAt, ls.autoFill = 0, 0
		s.scroll(ls.scrollKey).Y = 0
	}
	s.selected = map[string]bool{}
}

func (s *State) allLists() []*listSection {
	return []*listSection{&s.dynamicFeed, &s.favoriteFeed, &s.playHistoryFeed, &s.folderFeed, &s.authorFeed}
}

// activeList 返回路由当前展示的视频列表；无列表的页面返回 nil。
func (s *State) activeList(path string) *listSection {
	switch section(path) {
	case "dynamic":
		return &s.dynamicFeed
	case "play-history":
		return &s.playHistoryFeed
	case "favorite":
		if s.libraryTab != "favorite" {
			return &s.folderFeed
		}
		return &s.favoriteFeed
	case "user":
		if len(strings.Split(strings.Trim(path, "/"), "/")) > 2 {
			if s.authorTab == "series" {
				return &s.folderFeed
			}
			return &s.authorFeed
		}
	}
	return nil
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
			s.reportActionError(err.Error())
			return
		}
		navigate("/douyin/user/" + url.PathEscape(uid))
	})
}
func (s *State) loadAuthor(uid string) {
	s.author.Data = model.User{}
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
		// authorFeed 是用户页自己的列表，即使导航已离开也安全。
		s.openList(&s.authorFeed, s.authorSource(uid, user.NickName))
	})
}

// folderCatalogKey 目录的身份：收藏夹与合集各自一份，系列按用户区分。
func folderCatalogKey(kind, uid string) string {
	if kind == "series" {
		return "series:" + uid
	}
	return kind
}

// catalog 返回（并按需创建）一类目录的独立状态。
func (s *State) catalog(kind, uid string) *folderCatalog {
	key := folderCatalogKey(kind, uid)
	if s.folderCatalogs == nil {
		s.folderCatalogs = map[string]*folderCatalog{}
	}
	if c, ok := s.folderCatalogs[key]; ok {
		return c
	}
	c := &folderCatalog{kind: kind, uid: uid, scrollKey: "folders/" + key}
	s.folderCatalogs[key] = c
	return c
}

// currentFolderCatalog 返回当前页面显示的目录侧栏；没有侧栏时为 nil。
func (s *State) currentFolderCatalog(path string) *folderCatalog {
	switch section(path) {
	case "user":
		if s.authorTab == "series" {
			return s.catalog("series", s.author.Data.SecUid)
		}
	case "favorite":
		if s.libraryTab != "favorite" {
			return s.catalog(s.libraryTab, "")
		}
	}
	return nil
}

// ensureFolders 目录未加载（且未在加载）时才请求一次。
func (s *State) ensureFolders(kind, uid string) *folderCatalog {
	cat := s.catalog(kind, uid)
	if cat.folders.Loaded || cat.folders.Loading {
		return cat
	}
	s.reloadFolders(cat)
	return cat
}

// reloadFolders 重置目录并请求第一页（刷新、首次进入、深层链接兜底）。
func (s *State) reloadFolders(cat *folderCatalog) {
	cat.version++
	cat.autoPending = false
	api := s.api
	cat.folders.seq++
	cat.folders = resource[[]folder]{seq: cat.folders.seq}
	cat.cursor, cat.more = 0, false
	cat.next = folderLoader(api, cat.kind, cat.uid)
	s.loadFolders(cat, false)
}

// folderLoader builds one catalog kind's page loader.
func folderLoader(api *dyapi.Douyin, kind, uid string) func(int) ([]folder, int, bool, error) {
	switch kind {
	case "folder":
		return func(cursor int) ([]folder, int, bool, error) {
			items := []folder{}
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
		}
	case "series":
		return func(cursor int) ([]folder, int, bool, error) {
			items := []folder{}
			d, err := api.UserSeries(uid, cursor, pageSize)
			if err == nil {
				err = status(d.StatusCode)
			}
			for _, v := range d.SeriesInfos {
				items = append(items, folder{Key: v.SeriesID, Title: v.SeriesName, Author: uid, Count: max(v.Stats.TotalEpisode, v.Stats.UpdatedToEpisode)})
			}
			return items, d.Cursor, d.HasMore != 0, err
		}
	default:
		return func(cursor int) ([]folder, int, bool, error) {
			items := []folder{}
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
}

func (s *State) loadFolders(cat *folderCatalog, more bool) {
	if cat.next == nil || cat.folders.Loading {
		return
	}
	cursor := 0
	if !more {
		cat.version++
		cat.autoPending = false
		s.scroll(cat.scrollKey).Y = 0
	}
	if more && !cat.more {
		return
	}
	if more {
		cursor = cat.cursor
	}
	old := append([]folder(nil), cat.folders.Data...)
	oldCount := len(old)
	loader, version := cat.next, cat.version
	var next int
	var hasMore bool
	load(s, &cat.folders, func() ([]folder, error) {
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
		if version != cat.version {
			return
		}
		cat.cursor, cat.more = next, hasMore && next != cursor
		if s.pendingFolder != "" {
			for _, f := range items {
				if f.Key == s.pendingFolder {
					s.pendingFolder = ""
					s.openList(&s.folderFeed, s.folderSource(cat.sourceKind(), f))
					break
				}
			}
			if s.pendingFolder != "" && cat.more {
				s.loadFolders(cat, true)
			}
		}
	})
}

// Request the next page roughly one card row before the end.
const videoPreloadDistance float32 = 600

func (s *State) maybeMoreList(ls *listSection) {
	scroll := s.scroll(ls.scrollKey)
	ready := func() bool {
		if !ls.videos.Loaded || !ls.videos.Data.More || ls.videos.Loading || ls.more.Loading || ls.videos.Error != "" || ls.more.Error != "" {
			return false
		}
		if scroll.MaxY <= 0 {
			return ls.autoFill < 6
		}
		return scroll.Y > ls.autoAt+1 && scroll.Y+videoPreloadDistance >= scroll.MaxY
	}
	if s.runtime.After == nil || ls.autoPending || !ready() {
		return
	}
	ls.autoPending = true
	// 序号捕获列表重置（换来源），路径捕获板块切换，任一发生都不续页。
	epoch, seq, path := s.epoch, ls.videos.seq, s.activePath
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || ls.videos.seq != seq || s.activePath != path {
			return
		}
		ls.autoPending = false
		// The build can still see the previous layout's zero scroll range.
		// Recheck after layout commits before requesting another page.
		if !ready() {
			return
		}
		if scroll.MaxY <= 0 {
			ls.autoFill++
		} else {
			ls.autoFill = 0
		}
		ls.autoAt = scroll.Y
		s.loadList(ls, true)
	})
}

func (s *State) enter(path string) {
	if path == s.activePath {
		return
	}
	// 每板块列表独立保存，导航只做“未加载才请求”，与哔哩哔哩一致。
	s.activePath = path
	if s.api == nil {
		return
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	sec := "download"
	if len(parts) > 1 {
		sec = parts[1]
	}
	if sec == "history" {
		if !s.history.Loaded && !s.history.Loading && !s.historyBusy {
			s.loadHistory(false)
		}
		return
	}
	if !s.loggedIn() {
		return
	}
	switch sec {
	case "user":
		if len(parts) > 2 {
			if uid, err := url.PathUnescape(parts[2]); err == nil {
				s.enterAuthor(uid)
			}
		} else if !s.follows.Loaded && !s.follows.Loading {
			s.loadFollows(false)
		}
	case "dynamic":
		s.openList(&s.dynamicFeed, s.baseSource("dynamic"))
	case "play-history":
		s.openList(&s.playHistoryFeed, s.baseSource("play-history"))
	case "favorite":
		s.enterFavorite(parts)
	}
}

// enterFavorite 处理收藏路由：基础列表、收藏夹/合集目录与深层条目。
// 同类型目录已加载时直接恢复目标条目，后退零请求。
func (s *State) enterFavorite(parts []string) {
	if len(parts) >= 3 && (parts[2] == "folder" || parts[2] == "collection") {
		kind := parts[2]
		s.libraryTab = kind
		s.pendingFolder = ""
		if len(parts) > 3 {
			s.pendingFolder, _ = url.PathUnescape(parts[3])
		}
		if s.restoreFolder(kind) {
			return
		}
		s.openList(&s.folderFeed, source{})
		s.reloadFolders(s.catalog(kind, ""))
		return
	}
	s.libraryTab = "favorite"
	s.pendingFolder = ""
	s.openList(&s.favoriteFeed, s.baseSource("favorite"))
}

// restoreFolder 在该类型目录已加载时恢复深层条目或目录视图。
func (s *State) restoreFolder(kind string) bool {
	cat := s.catalog(kind, "")
	if !cat.folders.Loaded {
		return false
	}
	if s.pendingFolder == "" {
		s.openList(&s.folderFeed, source{})
		return true
	}
	for _, f := range cat.folders.Data {
		if f.Key == s.pendingFolder {
			s.pendingFolder = ""
			s.openList(&s.folderFeed, s.folderSource(cat.sourceKind(), f))
			return true
		}
	}
	return false
}

// enterAuthor 恢复访问过的用户主页：资料在且列表同源则零请求。
func (s *State) enterAuthor(uid string) {
	if s.author.Loaded && s.author.Error == "" && s.author.Data.SecUid == uid {
		s.openList(&s.authorFeed, s.authorSource(uid, s.author.Data.NickName))
		return
	}
	s.loadAuthor(uid)
}

func (s *State) loggedIn() bool {
	return s.profile.Loaded && s.profile.Error == "" && (s.profile.Data.Uid != "" || s.profile.Data.SecUid != "")
}

func duration(seconds int) string { return fmt.Sprintf("%d:%02d", seconds/60, seconds%60) }
