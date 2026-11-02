package douyin

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/egoist/mygo/ui"
	dyDownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
)

type API interface {
	GetCookie() (string, error)
	SetCookie(string) error
	Profile() (model.MyInfoResponse, error)
	ParseVideo(string) (string, error)
	VideoDetail(string) (model.AwemeItem, error)
	FollowList(int) (model.FollowResponse, error)
	SearchFollow(string) ([]model.SearchFollowResponse, error)
	ParseSecUserID(string) (string, error)
	User(string) (model.UserResponse, error)
	UserVideoList(string, int, int) (model.UserVideoListResponse, error)
	UserSeries(string, int, int) (model.UserSeriesResponse, error)
	CollectList(int) (model.CollectListResponse, error)
	FavoriteVideo(uint, uint) (model.FavoriteVideoResponse, error)
	FavoritesVideoList(string, int, int) (model.FavoriteVideoResponse, error)
	Collection(int, int) (model.CollectionResponse, error)
	CollectionList(string, string, int, int) (model.CollectionListResponse, error)
	Dynamic(int) (model.DynamicResponse, error)
	History(int, int, int, int) (model.HistoryResponse, error)
}
type DownloadBackend interface {
	DownloadVideos([]dyDownload.Task) (dyDownload.BatchResult, error)
	DownloadCover([]model.Cover, dyDownload.Task) (string, error)
	DownloadHistoryPage(int, int, string) (dyDownload.HistoryPage, error)
	DeleteDownloadHistory(string) error
	ClearDownloadHistory() error
	DownloadConcurrency() int
}

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

type State struct {
	api                                    API
	backend                                DownloadBackend
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
	folders                                resource[[]folder]
	libraryTab                             string
	authorTab                              string
	folderCursor                           int
	folderMore                             bool
	folderAutoPending                      bool
	folderNext                             func(int) ([]folder, int, bool, error)
	folderVersion                          uint64
	pendingFolder                          string
	videos                                 resource[videoPage]
	more                                   resource[videoPage]
	source                                 source
	sourceVersion                          uint64
	selected                               map[string]bool
	scrolls                                map[string]*ui.ScrollState
	images                                 imagecache.Cache
	autoPending                            bool
	autoAt                                 float32
	autoFill                               int
	staged                                 []*queuedVideo
	groups                                 []*downloadGroup
	nextGroup                              uint64
	history                                resource[dyDownload.HistoryPage]
	historyMore                            resource[dyDownload.HistoryPage]
	historyQuery                           string
	historySearch                          string
	historyBusy, clearOpen                 bool
}

func NewState() *State {
	return &State{selected: map[string]bool{}, scrolls: map[string]*ui.ScrollState{}, images: imagecache.Cache{}, libraryTab: "favorite"}
}
func (s *State) Configure(api API, backend DownloadBackend, runtime guiruntime.Runtime) {
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
			s.actionError = err.Error()
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
		s.folders = resource[[]folder]{}
		s.videos = resource[videoPage]{}
		s.more = resource[videoPage]{}
		s.source = source{}
		s.selected = map[string]bool{}
		s.groups = nil
		s.staged = nil
		s.folderVersion++
		s.sourceVersion++
		s.autoPending = false
		s.cookie = ""
		s.editCookie = false
		s.loadProfile()
	})
}
