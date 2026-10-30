package youtube

import (
	"context"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	yt "github.com/kamiertop/videodown/youtube"
)

type API interface {
	Cookie() string
	SetCookie(string) error
	Resolve(context.Context, string) (yt.Video, error)
	Playlist(context.Context, string, string) (yt.Playlist, error)
	Download(context.Context, yt.Video, int, func(yt.Progress)) (string, error)
	History() ([]yt.History, error)
	DownloadConcurrency() int
}
type State struct {
	started               bool
	detailsInFlight       int
	audioOnly             bool
	queue                 []*queueItem
	queueGroups           []*queueGroup
	playlistTasks         map[string]*queueItem
	thumbnails            map[string]*thumbnail
	thumbnailUse          uint64
	thumbnailLoader       func(context.Context, []string) (*ui.Bitmap, error)
	nextQueue             uint64
	active                map[*queueItem]context.CancelFunc
	queuePaused           bool
	parseCancel           context.CancelFunc
	playlistLink          string
	playlist              yt.Playlist
	playlistRows          []*playlistRow
	playlistRemoved       map[string]bool
	playlistLoading       bool
	playlistError         string
	playlistSource        string
	playlistGeneration    uint64
	playlistCancel        context.CancelFunc
	Link, cookie, message string
	authMessage           string
	cookieSaved           bool
	busy, stopped         bool
	video                 yt.Video
	format                int
	history               []yt.History
	historySearch         string
	cover                 *ui.Bitmap
	api                   API
	runtime               guiruntime.Runtime
	dispatch              func(func())
	ctx                   context.Context
	stop                  context.CancelFunc
	generation            uint64
	coverGeneration       uint64
	coverError            string
	coverLoading          bool
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
			s.mergeHistory(history)
		}
	})
}
func (s *State) PendingDownloadCount() int { return len(s.active) + s.waitingCount() }
func (s *State) waitingCount() int {
	count := 0
	for _, item := range s.queue {
		if item.Status == "queued" {
			count++
		}
	}
	return count
}
func (s *State) Stop() {
	s.stopped = true
	if s.stop != nil {
		s.stop()
	}
	s.generation++
	if s.parseCancel != nil {
		s.parseCancel()
	}
	if s.playlistCancel != nil {
		s.playlistCancel()
	}
	for _, cancel := range s.active {
		cancel()
	}
}
func (s *State) resolve() {
	if s.busy || s.api == nil || s.runtime.Background == nil {
		return
	}
	s.busy = true
	s.audioOnly = false
	s.message = ""
	s.video = yt.Video{}
	s.cover = nil
	s.coverGeneration++
	s.coverError = ""
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	s.parseCancel = cancel
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
		s.selectDefaultFormat()
		s.loadCover()
	})
}
func (s *State) saveCookie() {
	if s.busy || s.readyCount() > 0 || s.PendingDownloadCount() > 0 || s.api == nil || s.runtime.Background == nil {
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
		if s.playlistCancel != nil {
			s.playlistCancel()
		}
		s.playlistGeneration++
		s.playlistRows = nil
		s.playlist = yt.Playlist{}
		s.playlistLoading = false
		s.playlistError = ""
		s.video = yt.Video{}
		s.cover = nil
		s.coverGeneration++
		s.authMessage = "Cookie 已保存"
	})
}
func (s *State) stagePreview() {
	if s.busy || s.video.ID == "" {
		return
	}
	if s.enqueue(s.video, s.format, s.audioOnly) {
		s.video = yt.Video{}
		s.cover = nil
		s.coverGeneration++
	}
}

func (s *State) formats() []yt.Format {
	if s.audioOnly {
		return s.video.AudioFormats
	}
	return s.video.Formats
}
func (s *State) selectDefaultFormat() {
	s.format = 0
}
