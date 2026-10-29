package youtube

import (
	"context"
	"errors"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	yt "github.com/kamiertop/videodown/youtube"
)

type API interface {
	Cookie() string
	SetCookie(string) error
	Resolve(context.Context, string) (yt.Video, error)
	Download(context.Context, yt.Video, int, func(yt.Progress)) (string, error)
	History() ([]yt.History, error)
}
type State struct {
	started                       bool
	Link, cookie, message, output string
	authMessage                   string
	cookieSaved                   bool
	busy, downloading, stopped    bool
	video                         yt.Video
	format                        int
	progress                      yt.Progress
	history                       []yt.History
	historySearch                 string
	cover                         *ui.Bitmap
	api                           API
	runtime                       guiruntime.Runtime
	dispatch                      func(func())
	cancel                        context.CancelFunc
	ctx                           context.Context
	stop                          context.CancelFunc
	generation                    uint64
	coverGeneration               uint64
	coverError                    string
	coverLoading                  bool
}

func NewState() *State { return &State{} }
func (s *State) Configure(api API, runtime guiruntime.Runtime, dispatch func(func())) {
	s.ctx, s.stop = context.WithCancel(context.Background())
	s.api = api
	s.runtime = runtime
	s.dispatch = dispatch
	if api != nil {
		s.cookie = api.Cookie()
		s.cookieSaved = s.cookie != ""
	}
}
func (s *State) Start() {
	if s.started || s.stopped || s.api == nil || s.runtime.Background == nil {
		return
	}
	s.started = true
	api := s.api
	var history []yt.History
	s.runtime.Background(func() error { var err error; history, err = api.History(); return err }, func(err error) {
		if !s.stopped && err == nil {
			s.history = history
		}
	})
}
func (s *State) PendingDownloadCount() int {
	if s.downloading {
		return 1
	}
	return 0
}
func (s *State) Stop() {
	s.stopped = true
	if s.stop != nil {
		s.stop()
	}
	s.generation++
	if s.cancel != nil {
		s.cancel()
	}
}
func (s *State) resolve() {
	if s.busy || s.api == nil || s.runtime.Background == nil {
		return
	}
	s.busy = true
	s.message = ""
	s.output = ""
	s.video = yt.Video{}
	s.cover = nil
	s.coverGeneration++
	s.coverError = ""
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	s.cancel = cancel
	api, input := s.api, s.Link
	var video yt.Video

	s.runtime.Background(func() error {
		defer cancel()
		var err error
		video, err = api.Resolve(ctx, input)

		return err
	}, func(err error) {
		if s.stopped || generation != s.generation {
			return
		}
		s.busy = false
		if err != nil {
			s.message = err.Error()
			return
		}
		s.video = video
		s.format = video.Formats[0].ID
		s.loadCover()
	})
}
func (s *State) saveCookie() {
	if s.busy || s.api == nil || s.runtime.Background == nil {
		return
	}
	s.busy = true
	s.authMessage = ""
	api, value := s.api, s.cookie
	s.runtime.Background(func() error { return api.SetCookie(value) }, func(err error) {
		if s.stopped {
			return
		}
		s.busy = false
		if err != nil {
			s.authMessage = err.Error()
			return
		}
		s.cookie = api.Cookie()
		s.cookieSaved = s.cookie != ""
		s.video = yt.Video{}
		s.progress = yt.Progress{}
		s.output = ""
		s.cover = nil
		s.coverGeneration++
		s.authMessage = "Cookie 已保存"
	})
}
func (s *State) download() {
	if s.busy || s.api == nil || s.runtime.Background == nil || s.dispatch == nil || s.video.ID == "" {
		return
	}
	s.busy = true
	s.downloading = true
	s.message = ""
	s.output = ""
	s.progress = yt.Progress{}
	ctx, cancel := context.WithCancel(s.ctx)
	s.cancel = cancel
	s.generation++
	generation := s.generation
	api, video, format, dispatch := s.api, s.video, s.format, s.dispatch
	var path string
	var history []yt.History
	s.runtime.Background(func() error {
		defer cancel()
		var last time.Time
		var err error
		path, err = api.Download(ctx, video, format, func(p yt.Progress) {
			if p.Percent < 100 && time.Since(last) < 150*time.Millisecond {
				return
			}
			last = time.Now()
			dispatch(func() {
				if !s.stopped && s.generation == generation {
					s.progress = p
				}
			})
		})
		history, _ = api.History()
		return err
	}, func(err error) {
		if s.stopped || s.generation != generation {
			return
		}
		s.busy = false
		s.downloading = false
		s.output = path
		s.history = history
		if errors.Is(err, context.Canceled) {
			s.message = "已取消下载"
		} else if err != nil {
			s.message = err.Error()
		} else {
			s.message = "下载完成"
			s.progress.Percent = 100
		}
	})
}
