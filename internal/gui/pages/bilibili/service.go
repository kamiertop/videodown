package bilibili

import (
	"context"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
)

type resource[T any] struct {
	Data            T
	Loading, Loaded bool
	Error           string
	seq             uint64
}

func load[T any](s *State, r *resource[T], work func() (T, error), after func(T)) {
	loadWith(s, r, s.runtime.Run, work, after)
}

// loadBackground is load, but each request runs on its own goroutine: used
// for play-URL resolves, whose pace the api package's limiter already keeps.
func loadBackground[T any](s *State, r *resource[T], work func() (T, error), after func(T)) {
	loadWith(s, r, s.runtime.Background, work, after)
}

// loadWith schedules work through the given runner with shared stale guards:
// seq drops late responses to the same resource, epoch drops everything from
// before the last Stop, and stopped drops post-shutdown writes.
func loadWith[T any](s *State, r *resource[T], schedule func(work func() error, done func(error)), work func() (T, error), after func(T)) {
	if s.api == nil || schedule == nil || s.stopped {
		return
	}
	r.seq++
	seq, epoch := r.seq, s.epoch
	r.Loading = true
	r.Error = ""
	var data T
	schedule(func() error { var err error; data, err = work(); return err }, func(err error) {
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

func (s *State) Configure(api *biliapi.BiliBili, runtime guiruntime.Runtime) {
	s.api = api
	s.runtime = runtime
}
func (s *State) ConfigureDownloads(backend *bilidownload.Service) { s.downloadBackend = backend }
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
