package gui

import (
	"context"
	"github.com/egoist/mygo"
	"github.com/kamiertop/videodown/internal/gui/pages/bilibili"
	"github.com/kamiertop/videodown/utils"
	"sync"
	"time"
)

type nativeBilibiliJob struct {
	win  *mygo.Window
	work func() error
	done func(error)
}

// Preserve submission order for API calls, especially QR polls and logout.
// Images use a separate bounded pool so covers do not delay auth polling.
type bilibiliJobs struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wait    sync.WaitGroup
	mu      sync.Mutex
	queue   []nativeBilibiliJob
	running bool
	images  chan struct{}
}

func (j *bilibiliJobs) enqueue(job nativeBilibiliJob) {
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
func (j *bilibiliJobs) drain() {
	for {
		j.mu.Lock()
		if len(j.queue) == 0 {
			j.running = false
			j.mu.Unlock()
			return
		}
		job := j.queue[0]
		j.queue[0] = nativeBilibiliJob{}
		j.queue = j.queue[1:]
		j.mu.Unlock()
		j.execute(job)
	}
}
func (j *bilibiliJobs) execute(job nativeBilibiliJob) {
	defer j.wait.Done()
	if j.ctx.Err() != nil || job.win.IsDestroyed() {
		return
	}
	err := job.work()
	if j.ctx.Err() == nil && !job.win.IsDestroyed() {
		job.win.Update(func() { job.done(err) })
	}
}
func (s *Shell) bilibiliRuntime() bilibili.Runtime {
	return bilibili.Runtime{
		OpenFile:     utils.OpenLocalFile,
		OpenLocation: utils.OpenDownloadLocation,
		Run: func(work func() error, done func(error)) {
			win := s.win
			if win == nil || win.IsDestroyed() || s.biliJobs.ctx.Err() != nil {
				return
			}
			s.biliJobs.enqueue(nativeBilibiliJob{win, work, done})
		},
		Background: func(work func() error, done func(error)) {
			win := s.win
			if win == nil || win.IsDestroyed() || s.biliJobs.ctx.Err() != nil {
				return
			}
			s.biliJobs.wait.Add(1)
			go s.biliJobs.execute(nativeBilibiliJob{win, work, done})
		},
		Image: func(work func() error, done func(error)) {
			win := s.win
			if win == nil || win.IsDestroyed() || s.biliJobs.ctx.Err() != nil {
				return
			}
			s.biliJobs.wait.Add(1)
			go func() {
				select {
				case s.biliJobs.images <- struct{}{}:
				case <-s.biliJobs.ctx.Done():
					s.biliJobs.wait.Done()
					return
				}
				defer func() { <-s.biliJobs.images }()
				s.biliJobs.execute(nativeBilibiliJob{win, work, done})
			}()
		},
		After: func(delay time.Duration, fn func()) func() {
			win := s.win
			ctx := s.biliJobs.ctx
			timer := time.AfterFunc(delay, func() {
				if ctx.Err() == nil && win != nil && !win.IsDestroyed() {
					win.Update(fn)
				}
			})
			return func() { timer.Stop() }
		},
	}
}
func (s *Shell) stopBilibili() { s.pages.Bilibili.Stop(); s.pages.Douyin.Stop(); s.biliJobs.cancel() }

// Wait drains native jobs after App.Run, before the shared store is closed.
func (s *Shell) Wait() { s.stopBilibili(); s.biliJobs.wait.Wait() }
