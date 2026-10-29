package bilibili

import (
	"context"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
)

// API is implemented directly by *api.BiliBili. Views retain its Go models.
type API interface {
	DownloadHistoryPage(int, int, string) (biliapi.HistoryPage, error)
	ClearDownloadHistory() error
	DeleteDownloadHistory(int64, string) error
	IsLoggedIn() bool
	RefreshCookie() (model.CookieRefreshData, error)
	MyInfo() (model.MyInfoProfile, error)
	QRCode() (model.QRCodeData, error)
	PollQRCode(string) (model.PollQRCodeData, error)
	LogOut() (model.LogOut, error)
	FollowList(int, int) (model.FollowData, error)
	Info(string) (model.UserInfoData, error)
	VideoList(uint, int, int) (model.VideoListData, error)
	SeasonsSeriesList(string, int, int) (model.SeasonsSeriesData, error)
	SeasonsArchivesList(string, int, int, int) (model.SeasonsArchivesData, error)
	SeriesList(string, int, int, int) (model.SeriesArchivesData, error)
	FavoritesList() (model.FavoritesData, error)
	Favorites(int, int, int) (model.FavoriteData, error)
	Collection(int, int) (model.CollectionData, error)
	CollectionItem(string, int, int) (model.CollectionItemData, error)
	DynamicVideosPage(string) (model.DynamicArchivePage, error)
	PlayHistory(int, int) (model.PlayHistoryData, error)
	ResolvePlayUrl(biliapi.PlayUrlRequest, int) biliapi.PlayUrlResult
	VideoDetailConcise(int) (model.VideoDetailConciseData, error)
	VideoDetailConciseBvid(string) (model.VideoDetailConciseData, error)
}

type DownloadBackend interface {
	DownloadVideosByDash([]bilidownload.Task) (bilidownload.BatchResult, error)
}

type DownloadExtras interface {
	DownloadCover(string, bilidownload.Task) (string, error)
	DownloadAudio(bilidownload.Task) (string, error)
}

type resource[T any] struct {
	Data            T
	Loading, Loaded bool
	Error           string
	seq             uint64
}

func load[T any](s *State, r *resource[T], work func() (T, error), after func(T)) {
	if s.api == nil || s.runtime.Run == nil || s.stopped {
		return
	}
	r.seq++
	seq, epoch := r.seq, s.epoch
	r.Loading = true
	r.Error = ""
	var data T
	s.runtime.Run(func() error { var err error; data, err = work(); return err }, func(err error) {
		if s.stopped || s.epoch != epoch || r.seq != seq {
			return
		}
		r.Loading = false
		if err != nil {
			r.Error = err.Error()
			return
		}
		r.Data = data
		r.Loaded = true
		if after != nil {
			after(data)
		}
	})
}

func (s *State) Configure(api API, runtime guiruntime.Runtime) { s.api = api; s.runtime = runtime }
func (s *State) ConfigureDownloads(backend DownloadBackend)    { s.downloadBackend = backend }
func (s *State) Start() {
	if s.api == nil || s.started {
		return
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.started = true
	s.stopped = false
	s.checkAccount()
}
func (s *State) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.stopped = true
	s.epoch++
	s.stopPoll()
}

func (s *State) enter(section string) {
	changed := s.section != section
	if changed && s.section == "profile" {
		s.stopPoll()
	}
	s.section = section
	if !s.started || s.stopped {
		return
	}
	if section == "profile" && changed && !s.loggedIn && !s.checking {
		if s.qr.Loaded && !s.qrExpired {
			s.schedulePoll()
		} else if !s.qr.Loading && s.qr.Error == "" {
			s.newQRCode()
		}
	}
	if section == "history" && (changed || !s.history.Loaded && !s.history.Loading && s.history.Error == "") {
		s.loadDownloadHistory(false)
	}
	if !s.loggedIn {
		return
	}
	switch section {
	case "download":
		s.ensureQueueResolves()
	case "up":
		if !s.follows.Loaded && !s.follows.Loading && s.follows.Error == "" {
			s.loadFollows(1)
		}
	case "dynamic":
		if !s.dynamics.Loaded && !s.dynamics.Loading && s.dynamics.Error == "" {
			s.loadDynamics(false)
		}
	case "play-history":
		if !s.playHistory.Loaded && !s.playHistory.Loading && s.playHistory.Error == "" {
			s.loadPlayHistory(false)
		}
	case "favorite", "collection":
		s.ensureLibrary()
	}
}

// Leave stops account polling when another platform or settings becomes active.
func (s *State) Leave() { s.routeSeq++; s.activePath = ""; s.enter("") }
