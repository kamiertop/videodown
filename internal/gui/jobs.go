package gui

import (
	"context"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/utils"
)

type backgroundJob struct {
	win  *mygo.Window
	work func() error
	done func(error)
}

// Preserve submission order for API calls, especially QR polls and logout.
// Images use a separate bounded pool so covers do not delay auth polling.
type backgroundJobs struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wait    sync.WaitGroup
	mu      sync.Mutex
	queue   []backgroundJob
	running bool
	images  chan struct{}
}

func (j *backgroundJobs) enqueue(job backgroundJob) {
	j.wait.Add(1)
	j.mu.Lock()
	j.queue = append(j.queue, job)
	start := !j.running
	j.running = true
	j.mu.Unlock()
	if start {
		go j.drain()
	}
}
func (j *backgroundJobs) drain() {
	for {
		j.mu.Lock()
		if len(j.queue) == 0 {
			j.running = false
			j.mu.Unlock()
			return
		}
		job := j.queue[0]
		j.queue[0] = backgroundJob{}
		j.queue = j.queue[1:]
		j.mu.Unlock()
		j.execute(job)
	}
}
func (j *backgroundJobs) execute(job backgroundJob) {
	defer j.wait.Done()
	if j.ctx.Err() != nil || job.win.IsDestroyed() {
		return
	}
	err := job.work()
	if j.ctx.Err() == nil && !job.win.IsDestroyed() {
		job.win.Update(func() { job.done(err) })
	}
}
func (s *Shell) pageRuntime() guiruntime.Runtime {
	return guiruntime.Runtime{
		OpenFile:     utils.OpenLocalFile,
		OpenLocation: utils.OpenDownloadLocation,
		Run: func(work func() error, done func(error)) {
			win := s.win
			if win == nil || win.IsDestroyed() || s.jobs.ctx.Err() != nil {
				return
			}
			s.jobs.enqueue(backgroundJob{win, work, done})
		},
		Background: func(work func() error, done func(error)) {
			win := s.win
			if win == nil || win.IsDestroyed() || s.jobs.ctx.Err() != nil {
				return
			}
			s.jobs.wait.Add(1)
			go s.jobs.execute(backgroundJob{win, work, done})
		},
		Image: func(work func() error, done func(error)) {
			win := s.win
			if win == nil || win.IsDestroyed() || s.jobs.ctx.Err() != nil {
				return
			}
			s.jobs.wait.Add(1)
			go func() {
				select {
				case s.jobs.images <- struct{}{}:
				case <-s.jobs.ctx.Done():
					s.jobs.wait.Done()
					return
				}
				defer func() { <-s.jobs.images }()
				s.jobs.execute(backgroundJob{win, work, done})
			}()
		},
		After: func(delay time.Duration, fn func()) func() {
			win := s.win
			ctx := s.jobs.ctx
			timer := time.AfterFunc(delay, func() {
				if ctx.Err() == nil && win != nil && !win.IsDestroyed() {
					win.Update(fn)
				}
			})
			return func() { timer.Stop() }
		},
	}
}
func (s *Shell) stopJobs() {
	s.pages.Bilibili.Stop()
	s.pages.Douyin.Stop()
	s.pages.YouTube.Stop()
	s.jobs.cancel()
}

// Wait drains native jobs after App.Run, before the shared store is closed.
func (s *Shell) Wait() { s.stopJobs(); s.jobs.wait.Wait() }
