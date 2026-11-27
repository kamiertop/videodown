package douyin

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/egoist/mygo/ui"

	dyapi "github.com/kamiertop/videodown/douyin/api"
	dyDownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
)

type resource[T any] struct {
	Data            T
	Loading, Loaded bool
	Error           string
	seq             uint64
}

type videoPage struct {
	Items []model.AwemeItem
	Next  int
	More  bool
}

type folder struct {
	Key, Title, Author string
	Count              int
}

type source struct {
	Key, Title string
	Load       func(int) (videoPage, error)
}

// folderCatalog 是一类收藏目录（收藏夹 / 合集 / 某用户的系列）的独立状态，
// 与哔哩哔哩的收藏夹目录、合集目录各自保留一致：页签切换、来回导航都
// 不再重新请求已加载的目录。
type folderCatalog struct {
	kind, uid   string // uid 仅系列目录使用
	folders     resource[[]folder]
	cursor      int
	more        bool
	next        func(int) ([]folder, int, bool, error)
	version     uint64
	autoPending bool
	scrollKey   string
}

// sourceKind 把目录映射到其条目的视频来源类型；系列条目走合集接口。
func (f *folderCatalog) sourceKind() string {
	if f.kind == "series" {
		return "collection"
	}
	return f.kind
}

// listSection 是一个板块独立的视频列表：数据、分页、自动加载与滚动位置
// 各归各的，切走再回来原样恢复，与哔哩哔哩每板块独立资源一致。
type listSection struct {
	videos, more resource[videoPage]
	// src 描述当前展示的数据；零值表示目录视图（右侧空列表）。
	src         source
	autoPending bool
	autoAt      float32
	autoFill    int
	scrollKey   string
}

type State struct {
	api                                    *dyapi.Douyin
	backend                                *dyDownload.Service
	runtime                                guiruntime.Runtime
	apiMu                                  sync.Mutex
	ctx                                    context.Context
	cancel                                 context.CancelFunc
	started, stopped                       bool
	epoch                                  uint64
	activePath                             string
	profile                                resource[model.User]
	cookie, link, authorInput              string
	notice, actionError                    string
	authBusy, parsing                      bool
	authSeq, parseSeq, authorSeq           uint64
	editCookie                             bool
	follows                                resource[model.FollowResponse]
	followSearchInput, followQuery         string
	followResults                          resource[[]model.SearchFollowResponse]
	followSearchPending, followAutoPending bool
	followGeneration                       uint64
	cancelFollowSearch                     func()
	author                                 resource[model.User]
	folderCatalogs                         map[string]*folderCatalog
	libraryTab                             string
	authorTab                              string
	pendingFolder                          string
	// 每板块独立的视频列表，替代旧的单个共享列表 + 来源切换。
	dynamicFeed            listSection // 关注动态
	favoriteFeed           listSection // 收藏视频（基础列表）
	playHistoryFeed        listSection // 播放历史
	folderFeed             listSection // 当前打开的收藏夹 / 合集 / 系列
	authorFeed             listSection // 当前打开的用户作品
	scrolls                map[string]*ui.ScrollState
	images                 imagecache.Cache
	selected               map[string]bool
	downloadList           ui.ListState
	staged                 []*queuedVideo
	groups                 []*downloadGroup
	nextGroup              uint64
	history                resource[dyDownload.HistoryPage]
	historyMore            resource[dyDownload.HistoryPage]
	historyQuery           string
	historySearch          string
	historyBusy, clearOpen bool
}

func NewState() *State {
	s := &State{selected: map[string]bool{}, scrolls: map[string]*ui.ScrollState{}, images: imagecache.Cache{}, libraryTab: "favorite"}
	s.dynamicFeed.scrollKey = "videos/dynamic"
	s.favoriteFeed.scrollKey = "videos/favorite"
	s.playHistoryFeed.scrollKey = "videos/play-history"
	s.folderFeed.scrollKey = "videos/folder"
	s.authorFeed.scrollKey = "videos/user"
	return s
}

func (s *State) Configure(api *dyapi.Douyin, backend *dyDownload.Service, runtime guiruntime.Runtime) {
	s.api, s.backend, s.runtime = api, backend, runtime
}

func (s *State) Link() *string { return &s.link }

func (s *State) scroll(key string) *ui.ScrollState {
	if s.scrolls[key] == nil {
		s.scrolls[key] = &ui.ScrollState{}
	}
	return s.scrolls[key]
}

func (s *State) Start() {
	if s.started || s.api == nil {
		return
	}
	s.started = true
	s.stopped = false
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.loadProfile()
}

func (s *State) Stop() {
	s.stopped = true
	if s.cancelFollowSearch != nil {
		s.cancelFollowSearch()
	}
	s.epoch++
	if s.cancel != nil {
		s.cancel()
	}
}

func load[T any](s *State, r *resource[T], work func() (T, error), done func(T)) {
	if s.stopped || s.runtime.Run == nil {
		return
	}
	r.seq++
	seq, epoch := r.seq, s.epoch
	r.Loading = true
	r.Error = ""
	var data T
	s.runtime.Run(func() (err error) { s.apiMu.Lock(); defer s.apiMu.Unlock(); data, err = work(); return err }, func(err error) {
		if s.stopped || s.epoch != epoch || r.seq != seq {
			return
		}
		r.Loading = false
		if err != nil {
			r.Error = err.Error()
			return
		}
		r.Data, r.Loaded = data, true
		if done != nil {
			done(data)
		}
	})
}
func status(code int) error {
	if code != 0 {
		return fmt.Errorf("抖音接口返回错误：%d，请检查 Cookie 或稍后重试", code)
	}
	return nil
}

func (s *State) loadProfile() {
	if s.api == nil {
		return
	}
	api := s.api
	load(s, &s.profile, func() (model.User, error) {
		cookie, err := api.GetCookie()
		if err != nil || cookie == "" {
			return model.User{}, err
		}
		result, err := api.Profile()
		if err == nil {
			err = status(result.StatusCode)
		}
		if err == nil && result.User.Uid == "" && result.User.SecUid == "" {
			err = errors.New("Cookie 未登录或已失效")
		}
		return result.User, err
	}, func(model.User) {
		if path := s.activePath; path != "" {
			s.activePath = ""
			s.enter(path)
		}
	})
}

// cancelEditCookie 收起更新 Cookie 的编辑框并丢弃未保存的草稿；
// 已保存的登录态不受影响。仅在已登录时提供（登录表单没有可取消的对象）。
func (s *State) cancelEditCookie() {
	if s.authBusy {
		return
	}
	s.editCookie = false
	s.cookie = ""
}

func (s *State) saveCookie(clear bool) {
	if s.authBusy || s.api == nil || s.runtime.Run == nil || s.PendingDownloadCount() > 0 {
		return
	}
	cookie := s.cookie
	if clear {
		cookie = ""
	}
	s.notice = ""
	s.authBusy = true
	s.actionError = ""
	s.authSeq++
	seq, epoch := s.authSeq, s.epoch
	api := s.api
	s.runtime.Run(func() error { s.apiMu.Lock(); defer s.apiMu.Unlock(); return api.SetCookie(cookie) }, func(err error) {
		if s.stopped || s.epoch != epoch || s.authSeq != seq {
			return
		}
		s.authBusy = false
		if err != nil {
			s.reportActionError(err.Error())
			return
		}
		s.epoch++
		s.profile = resource[model.User]{}
		s.author.seq++
		s.author = resource[model.User]{seq: s.author.seq}
		s.images = imagecache.Cache{}
		s.follows = resource[model.FollowResponse]{}
		if s.cancelFollowSearch != nil {
			s.cancelFollowSearch()
		}
		s.followSearchInput, s.followQuery = "", ""
		s.followResults = resource[[]model.SearchFollowResponse]{}
		s.followSearchPending, s.followAutoPending = false, false
		s.followGeneration++
		s.folderCatalogs = nil
		s.pendingFolder = ""
		s.clearLists()
		s.groups = nil
		s.staged = nil
		s.cookie = ""
		s.editCookie = false
		s.loadProfile()
	})
}
