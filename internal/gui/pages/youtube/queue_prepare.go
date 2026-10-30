package youtube

import (
	"context"
	"errors"
	"time"

	yt "github.com/kamiertop/videodown/youtube"
)

func editableTask(item *queueItem) bool {
	return item.Status == "ready" || item.Status == "error" || item.Status == "canceled"
}

func (s *State) readyCount() int {
	count := 0
	for _, item := range s.queue {
		if item.Status == "ready" {
			count++
		}
	}
	return count
}

func (s *State) startItem(item *queueItem) {
	if s.stopped || !editableTask(item) {
		return
	}
	s.cancelDetails(item)
	item.Status, item.Error = "queued", ""
	item.Progress = yt.Progress{}
	s.pumpQueue()
}

func (s *State) startGroup(group *queueGroup) {
	for _, item := range s.groupItems(group) {
		if item.Status != "ready" {
			continue
		}
		s.cancelDetails(item)
		item.Status = "queued"
	}
	s.pumpQueue()
}

func (s *State) startAll() {
	for _, item := range s.queue {
		if item.Status != "ready" {
			continue
		}
		s.cancelDetails(item)
		item.Status = "queued"
	}
	s.pumpQueue()
}

func (s *State) cancelDetails(item *queueItem) {
	if item.DetailsCancel != nil {
		item.DetailsCancel()
		item.DetailsCancel = nil
	}
	item.DetailsAttempt++
	item.DetailsLoading = false
}

// Only visible staged rows request stream metadata. Resolving formats never
// starts a media transfer, and at most two metadata jobs are outstanding.
func (s *State) prepareItem(item *queueItem) {
	if s.stopped || !editableTask(item) || item.DetailsLoading || item.DetailsError != "" || s.detailsInFlight >= 2 || len(item.Video.Formats)+len(item.Video.AudioFormats) > 0 {
		return
	}
	s.detailsInFlight++
	item.DetailsLoading = true
	item.DetailsAttempt++
	attempt := item.DetailsAttempt
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	item.DetailsCancel = cancel
	api, id := s.api, item.Video.ID
	var video yt.Video
	s.runtime.Background(func() error { defer cancel(); var err error; video, err = api.Resolve(ctx, id); return err }, func(err error) {
		s.detailsInFlight--
		if s.stopped || item.DetailsAttempt != attempt {
			return
		}
		item.DetailsLoading = false
		item.DetailsCancel = nil
		if err != nil {
			item.DetailsError = err.Error()
			return
		}
		item.Video = video
		if item.Format < 0 {
			if format, err := chooseQueueFormat(video, item.Format, item.AudioOnly); err == nil {
				item.Format = format
			}
		}
	})
}

func chooseQueueFormat(video yt.Video, format int, audio bool) (int, error) {
	formats := video.Formats
	if audio {
		formats = video.AudioFormats
	}
	if len(formats) == 0 {
		return 0, errors.New("没有可下载的对应格式")
	}
	if format > 0 {
		for _, f := range formats {
			if f.ID == format {
				return format, nil
			}
		}
		return 0, errors.New("所选格式已不可用，请重新选择画质或音质")
	}
	if audio || format >= -1 {
		return formats[0].ID, nil
	}
	height := -format
	for _, f := range formats {
		if f.Height <= height {
			return f.ID, nil
		}
	}
	return formats[len(formats)-1].ID, nil
}
