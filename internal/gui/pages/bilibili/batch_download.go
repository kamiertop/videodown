package bilibili

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
)

type batchPage struct {
	Items           []bilidownload.Task
	Total, Received int
	HasMore         bool
}
type batchSource struct {
	Key, Title string
	Total      int
	// Captures source values at click time; independent of the visible page.
	Load func(int) (batchPage, error)
}
type batchItem struct {
	Task          bilidownload.Task
	Status, Error string
	Progress      bilidownload.Progress
}
type batchSession struct {
	ID                                 uint64
	Order, FinishVersion               uint64
	Source                             batchSource
	Expanded                           bool
	Items                              []*batchItem
	Seen                               map[string]bool
	Page, Total, Done, Failed, Skipped int
	HasMore, Active, StopRequested     bool
	Status, Error                      string
	InFlight                           int
	Loading                            bool
}

func (s *State) batchesRunning() bool {
	for _, b := range s.batches {
		if b.Active && b.Status != "queued" {
			return true
		}
	}
	return false
}

func (s *State) PendingDownloadCount() int {
	count := 0
	for _, b := range s.batches {
		if b.Active {
			count++
		}
	}
	if s.queueDownloading {
		count++
	}
	return count + len(s.pendingManualDownloads) + len(s.queueExtraBusy)
}

func (s *State) enqueueBatch(source batchSource, navigate ...func(string)) {
	if s.downloadBackend == nil || s.runtime.Background == nil || s.runtime.Run == nil {
		s.stageError = "下载服务尚未就绪"
		return
	}
	for _, b := range s.batches {
		if b.Active && b.Source.Key == source.Key {
			if len(navigate) > 0 {
				navigate[0]("/bilibili/download")
			}
			return
		}
	}
	s.nextBatchID++
	s.nextDownloadOrder++
	b := &batchSession{ID: s.nextBatchID, Order: s.nextDownloadOrder, Source: source, Expanded: true, Seen: map[string]bool{}, Page: 1, Total: source.Total, HasMore: true, Active: true, Status: "queued"}
	s.batches = append(s.batches, b)
	s.stageError = ""
	if len(navigate) > 0 {
		navigate[0]("/bilibili/download")
	}
	s.pumpDownloads()
}

// Batch sessions and manually requested videos share FIFO ordering and one download lock.
func (s *State) pumpDownloads() {
	if s.stopped || !s.loggedIn || len(s.queueExtraBusy) > 0 || s.batchesRunning() {
		return
	}
	if s.queueDownloading {
		s.advanceManualDownloads()
		return
	}
	var next *batchSession
	for _, b := range s.batches {
		if b.Active && b.Status == "queued" && (next == nil || b.Order < next.Order) {
			next = b
		}
	}
	if len(s.pendingManualDownloads) > 0 && (next == nil || s.pendingManualDownloads[0].order < next.Order) {
		request := s.pendingManualDownloads[0]
		s.pendingManualDownloads = s.pendingManualDownloads[1:]
		s.queueDownloadTotal, s.queueDownloadDone, s.queueDownloadFailed = len(request.tasks), 0, 0
		s.queueDownloadStop, s.queueDownloading = false, true
		s.queueDownloadError = ""
		s.activeManualDownloads = &manualWork{tasks: request.tasks, epoch: s.epoch}
		s.advanceManualDownloads()
		return
	}
	if next != nil {
		s.advanceBatch(next)
	}
}

func (s *State) finishBatch(b *batchSession, status string) {
	b.Active, b.Status = false, status
	b.FinishVersion++
	if status == "done" && b.Failed == 0 && b.Error == "" {
		if s.runtime.After == nil {
			s.dismissBatch(b)
		} else {
			epoch, version := s.epoch, b.FinishVersion
			s.runtime.After(3*time.Second, func() {
				if !s.stopped && s.epoch == epoch && b.FinishVersion == version && !b.Active && b.Status == "done" && b.Failed == 0 {
					s.dismissBatch(b)
				}
			})
		}
	}
	s.pumpDownloads()
}

func (s *State) advanceBatch(b *batchSession) {
	if s.stopped {
		return
	}
	if b.Loading {
		if b.InFlight == 0 {
			b.Status = "loading"
		}
		return
	}
	if b.StopRequested {
		if b.InFlight == 0 {
			s.finishBatch(b, "stopped")
		}
		return
	}
	limit := s.downloadConcurrency()
	for _, item := range b.Items {
		if b.InFlight >= limit {
			return
		}
		if item.Status == "pending" {
			s.runBatchItem(b, item)
		}
	}
	if b.InFlight >= limit {
		return
	}
	if !b.HasMore {
		if b.InFlight == 0 {
			s.finishBatch(b, "done")
		}
		return
	}
	b.Loading = true
	if b.InFlight == 0 {
		b.Status = "loading"
	}

	epoch, page := s.epoch, b.Page
	loader := b.Source.Load
	var result batchPage
	s.runtime.Run(func() (err error) { result, err = loader(page); return err }, func(err error) {
		if s.stopped || epoch != s.epoch {
			return
		}
		b.Loading = false
		if err != nil {
			b.Error = err.Error()
			b.StopRequested = true
			s.advanceBatch(b)
			return
		}
		fresh := 0
		for _, task := range result.Items {
			if !strings.HasPrefix(task.Bvid, "BV") {
				continue
			}
			key := queueProgressKey(task.Bvid, task.Cid)
			if b.Seen[key] {
				continue
			}
			b.Seen[key] = true
			fresh++
			b.Items = append(b.Items, &batchItem{Task: task, Status: "pending"})
		}
		b.Total = max(b.Total, result.Total, len(b.Items))
		b.HasMore = result.HasMore && !(result.Total > 0 && len(b.Seen) >= result.Total)
		if b.HasMore && (result.Received == 0 || len(result.Items) > 0 && fresh == 0 || page >= 10000) {
			b.Error = "分页没有返回新视频，请刷新来源后重试"
			b.StopRequested = true
			s.advanceBatch(b)
			return
		}
		b.Page++
		s.advanceBatch(b)
	})
}

func (s *State) runBatchItem(b *batchSession, item *batchItem) {
	b.Status = "downloading"
	b.InFlight++
	item.Progress = bilidownload.Progress{}
	item.Status, item.Error = "downloading", ""
	task, api, backend, epoch := item.Task, s.api, s.downloadBackend, s.epoch
	skipped := false
	s.runtime.Background(func() error {
		if incremental, ok := backend.(interface {
			SkipDownloaded() bool
			DownloadedVideoKeys([]bilidownload.DownloadKey) ([]int, error)
		}); ok && incremental.SkipDownloaded() {
			keys, err := incremental.DownloadedVideoKeys([]bilidownload.DownloadKey{{Bvid: task.Bvid, Cid: task.Cid}})
			if err != nil {
				return err
			}
			if len(keys) > 0 {
				skipped = true
				return nil
			}
		}
		resolved := api.ResolvePlayUrl(biliapi.PlayUrlRequest{Bvid: task.Bvid, Cid: task.Cid}, 1)
		if resolved.Error != "" {
			return errors.New(resolved.Error)
		}
		if resolved.Detail == nil || resolved.PlayURL == nil {
			return errors.New("播放地址解析结果不完整")
		}
		view := resolved.Detail.View
		task.Cid = resolved.Cid
		if task.Cid == 0 {
			task.Cid = view.Cid
		}
		if task.Title == "" {
			task.Title = view.Title
		}
		if task.Cover == "" {
			task.Cover = view.Pic
		}
		if task.UpperName == "" {
			task.UpperName = view.Owner.Name
		}
		if task.Duration == 0 {
			task.Duration = view.Duration
		}
		task.VideoURL, task.AudioURL = bestAuthorStreams(resolved.PlayURL)
		if task.VideoURL == "" {
			return errors.New("没有可下载的视频流")
		}
		result, err := backend.DownloadVideosByDash([]bilidownload.Task{task})
		return downloadResultError(result, err)
	}, func(err error) {
		if s.stopped || epoch != s.epoch {
			return
		}
		item.Task = task
		if err != nil {
			item.Status, item.Error = "error", err.Error()
			b.Failed++
			b.Error = fmt.Sprintf("%s：%v", task.Title, err)
		} else if skipped {
			item.Status = "skipped"
			b.Skipped++
		} else {
			item.Status = "done"
			b.Done++
		}
		b.InFlight--
		s.advanceBatch(b)
	})
}

func downloadResultError(result bilidownload.BatchResult, err error) error {
	if err != nil {
		return err
	}
	if result.Failed > 0 {
		for _, item := range result.Results {
			if item.Error != "" {
				return errors.New(item.Error)
			}
		}
		return errors.New("下载失败")
	}
	if result.Success == 0 {
		return errors.New("下载服务未返回成功结果")
	}
	return nil
}

func (s *State) stopBatch(b *batchSession) {
	if !b.Active {
		return
	}
	b.StopRequested = true
	if b.Status == "queued" {
		s.finishBatch(b, "stopped")
	}
}
func (s *State) resumeBatch(b *batchSession) {
	if b.Active {
		return
	}
	for _, item := range b.Items {
		if item.Status == "error" {
			item.Status, item.Error = "pending", ""
		}
	}
	b.Failed, b.Error, b.StopRequested = 0, "", false
	b.FinishVersion++
	s.nextDownloadOrder++
	b.Order = s.nextDownloadOrder
	b.Active, b.Status = true, "queued"
	s.pumpDownloads()
}
func (s *State) dismissBatch(b *batchSession) {
	if !b.Active {
		s.batches = slices.DeleteFunc(s.batches, func(v *batchSession) bool { return v.ID == b.ID })
	}
}

func (s *State) downloadConcurrency() int {
	if settings, ok := s.downloadBackend.(interface{ DownloadConcurrency() int }); ok {
		return max(1, min(10, settings.DownloadConcurrency()))
	}
	return 1
}
