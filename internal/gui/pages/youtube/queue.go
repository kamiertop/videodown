package youtube

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	yt "github.com/kamiertop/videodown/youtube"
)

type queueItem struct {
	Group               *queueGroup
	ID                  uint64
	Video               yt.Video
	Format              int
	AudioOnly           bool
	Status, Error, Path string
	Progress            yt.Progress
	Attempt             uint64
	DetailsLoading      bool
	DetailsError        string
	DetailsAttempt      uint64
	DetailsCancel       context.CancelFunc
}

func (s *State) enqueue(video yt.Video, format int, audio bool) bool {
	return s.enqueueGrouped(video, format, audio, "manual", "手动添加", true)
}

func (s *State) enqueueGrouped(video yt.Video, format int, audio bool, groupKey, title string, expanded bool) bool {
	if s.stopped || s.api == nil || s.runtime.Background == nil || s.dispatch == nil || video.ID == "" {
		return false
	}
	for _, item := range s.queue {
		sameFormat := item.Format == format || item.Format <= 0 || format <= 0
		if item.Video.ID == video.ID && sameFormat && item.AudioOnly == audio && (item.Status == "ready" || item.Status == "queued" || item.Status == "active") {
			s.message = "该任务已在下载队列中"
			return false
		}
	}
	s.nextQueue++
	var group *queueGroup
	for _, existing := range s.queueGroups {
		if existing.Key == groupKey {
			group = existing
			break
		}
	}
	if group == nil {
		group = &queueGroup{Key: groupKey, Title: title, Expanded: expanded}
		s.queueGroups = append(s.queueGroups, group)
	}
	item := &queueItem{Group: group, ID: s.nextQueue, Video: video, Format: format, AudioOnly: audio, Status: "ready"}
	s.queue = append(s.queue, item)
	if s.playlistTasks == nil {
		s.playlistTasks = map[string]*queueItem{}
	}
	s.playlistTasks[video.ID] = item
	if video.ID == s.video.ID && s.cover != nil {
		s.thumbnailEntry(video).Bitmap = s.cover
	}
	s.message = "已加入下载队列"
	return true
}

// Queued jobs share the configured download limit across all YouTube groups.
// Workers capture snapshots and publish all state changes on the UI thread.
func (s *State) pumpQueue() {
	if s.stopped || s.queuePaused || s.api == nil || s.runtime.Background == nil {
		return
	}
	limit := max(1, min(10, s.api.DownloadConcurrency()))
	for _, item := range s.queue {
		if len(s.active) >= limit {
			return
		}
		if item.Status != "queued" {
			continue
		}
		item.Status, item.Error = "active", ""
		item.Progress = yt.Progress{Phase: "正在解析"}
		item.Attempt++
		attempt := item.Attempt
		ctx, cancel := context.WithCancel(s.ctx)
		if s.active == nil {
			s.active = map[*queueItem]context.CancelFunc{}
		}
		s.active[item] = cancel
		api, video, format, audio, dispatch := s.api, item.Video, item.Format, item.AudioOnly, s.dispatch
		var path string
		var history []yt.History
		var resolved yt.Video
		s.runtime.Background(func() error {
			defer cancel()
			if len(video.Formats)+len(video.AudioFormats) == 0 {
				var err error
				parseCtx, stop := context.WithTimeout(ctx, 60*time.Second)
				resolved, err = api.Resolve(parseCtx, video.ID)
				stop()
				if err != nil {
					return err
				}
				video = resolved
			}
			var formatErr error
			format, formatErr = chooseQueueFormat(video, format, audio)
			if formatErr != nil {
				return formatErr
			}
			var last time.Time
			var err error
			path, err = api.Download(ctx, video, format, func(p yt.Progress) {
				if p.Percent < 100 && time.Since(last) < 150*time.Millisecond {
					return
				}
				last = time.Now()
				dispatch(func() {
					if s.stopped || item.Attempt != attempt || item.Status != "active" {
						return
					}
					item.Progress = p
				})
			})
			history, _ = api.History()
			return err
		}, func(err error) {
			if item.Attempt != attempt {
				return
			}
			delete(s.active, item)
			if s.stopped {
				return
			}
			item.Path = path
			if format > 0 {
				item.Format = format
			}
			if resolved.ID != "" {
				item.Video = resolved
			}
			if history != nil {
				s.mergeHistory(history)
			}
			switch {
			case errors.Is(err, context.Canceled):
				item.Status = "canceled"
			case err != nil:
				item.Status, item.Error = "error", err.Error()
			default:
				item.Status = "done"
				item.Progress = yt.Progress{Phase: "已完成", Percent: 100}
			}
			s.pumpQueue()
		})
	}
}

func (s *State) cancelItem(item *queueItem) {
	if cancel := s.active[item]; cancel != nil {
		cancel()
		return
	}
	if item.Status == "ready" || item.Status == "queued" {
		item.Status = "canceled"
	}
}

func (s *State) retryItem(item *queueItem) {
	if item.Status != "error" && item.Status != "canceled" {
		return
	}
	s.startItem(item)
}

func (s *State) removeItem(item *queueItem) {
	s.cancelItem(item)
	s.cancelDetails(item)
	if item.Status == "active" {
		item.Status = "canceling"
	}
	s.queue = slices.DeleteFunc(s.queue, func(q *queueItem) bool { return q == item })
	s.pruneQueueGroups()
}

func queueStatus(item *queueItem) string {
	switch item.Status {
	case "ready":
		return "待下载"
	case "queued":
		return "等待下载"
	case "active":
		return transferStatus(item.Progress)
	case "done":
		return "已完成"
	case "canceled":
		return "已取消"
	case "error":
		return "下载失败：" + item.Error
	}
	return item.Status
}

func queueFormat(item *queueItem) string {
	if item.AudioOnly {
		for _, f := range item.Video.AudioFormats {
			if f.ID == item.Format {
				return "仅音频 · " + f.Summary
			}
		}
		return "仅音频 · 最佳音质"
	}
	for _, f := range item.Video.Formats {
		if f.ID == item.Format {
			if f.Summary != "" {
				return f.Summary
			}
			return f.Label
		}
	}
	if item.Format < -1 {
		return fmt.Sprintf("视频 · 最高 %dp", -item.Format)
	}
	return "视频 · 最高画质"
}
