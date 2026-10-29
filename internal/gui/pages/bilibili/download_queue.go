package bilibili

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/input"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/bilibili/model"
)

type queueResolve struct {
	data       resource[biliapi.PlayUrlResult]
	videoID    int
	audioID    int
	resolvedAt time.Time
}

func queueProgressKey(bvid string, cid int64) string {
	if cid > 0 {
		return fmt.Sprintf("%s:%d", bvid, cid)
	}
	return bvid
}

// UpdateDownloadProgress runs on the UI thread via Window.Update.
func (s *State) UpdateDownloadProgress(progress bilidownload.Progress) {
	if s.stopped {
		return
	}
	for _, b := range s.batches {
		if !b.Active {
			continue
		}
		for _, item := range b.Items {
			if item.Status == "downloading" && item.Task.Bvid == progress.Bvid && (item.Task.Cid == 0 || item.Task.Cid == progress.Cid) {
				item.Progress = progress
				return
			}
		}
	}
	key := queueProgressKey(progress.Bvid, progress.Cid)
	if !s.stagedKeys[key] && !s.stagedKeys[progress.Bvid] {
		return
	}
	if s.queueProgress == nil {
		s.queueProgress = make(map[string]bilidownload.Progress)
	}
	s.queueProgress[key] = progress
}

func (s *State) progressForKey(key string) (bilidownload.Progress, bool) {
	if manual, ok := s.manualItem(key); ok {
		progress, found := s.queueProgress[queueProgressKey(manual.Bvid, manual.Cid)]
		return progress, found
	}
	if entry := s.queueResolves[key]; entry != nil && entry.data.Loaded {
		progress, found := s.queueProgress[queueProgressKey(key, entry.data.Data.Cid)]
		return progress, found
	}
	return bilidownload.Progress{}, false
}

// stagedOrder keeps queue order stable and excludes records already removed.
func (s *State) stagedOrder() []string {
	keys := make([]string, 0, len(s.stagedKeys))
	seen := make(map[string]bool, len(s.stagedKeys))
	appendKey := func(key string) {
		if s.stagedKeys[key] && !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	for _, v := range s.stagedUploads {
		appendKey(v.Bvid)
	}
	for _, v := range s.stagedFavorites {
		appendKey(favoriteBvid(v.Bvid, v.BvID))
	}
	for _, v := range s.stagedCollections {
		appendKey(v.Bvid)
	}
	for _, v := range s.stagedSeasons {
		appendKey(v.Bvid)
	}
	for _, v := range s.stagedSeries {
		appendKey(v.Bvid)
	}
	for _, v := range s.stagedDynamics {
		appendKey(v.Bvid)
	}
	for _, v := range s.stagedPlayHistory {
		appendKey(v.History.Bvid)
	}
	for _, v := range s.stagedManual {
		appendKey(v.Key)
	}
	return keys
}

func (s *State) ensureQueueResolves() {
	if s.api == nil || !s.loggedIn {
		return
	}
	for _, key := range s.stagedOrder() {
		s.resolveQueueItem(key)
	}
}

func (s *State) resolveQueueItem(key string) {
	if s.api == nil || !s.stagedKeys[key] {
		return
	}
	if s.queueResolves == nil {
		s.queueResolves = make(map[string]*queueResolve)
	}
	entry := s.queueResolves[key]
	if entry == nil {
		entry = &queueResolve{}
		s.queueResolves[key] = entry
	}
	if entry.data.Loading || entry.data.Loaded || entry.data.Error != "" {
		return
	}
	api, count := s.api, len(s.stagedKeys)
	bvid, cid := key, int64(0)
	if item, ok := s.manualItem(key); ok {
		bvid, cid = item.Bvid, item.Cid
	}
	load(s, &entry.data, func() (biliapi.PlayUrlResult, error) {
		result := api.ResolvePlayUrl(biliapi.PlayUrlRequest{Bvid: bvid, Cid: cid}, count)
		if result.Error != "" {
			return result, errors.New(result.Error)
		}
		if result.PlayURL == nil || result.Detail == nil {
			return result, errors.New("播放地址解析结果不完整")
		}
		videoURL, _ := bestAuthorStreams(result.PlayURL)
		if videoURL == "" {
			return result, errors.New("没有可下载的视频流")
		}
		return result, nil
	}, func(result biliapi.PlayUrlResult) {
		if !s.stagedKeys[key] {
			return
		}
		entry.resolvedAt = time.Now()
		entry.videoID = bestQueueVideoID(result.PlayURL)
		entry.audioID = bestQueueAudioID(result.PlayURL)
	})
}

func (s *State) retryQueueItem(key string) {
	if entry := s.queueResolves[key]; entry != nil {
		entry.data.Error = ""
		entry.data.Loaded = false
	}
	s.resolveQueueItem(key)
}

func bestQueueVideoID(play *model.VideoURLData) int {
	best, bandwidth, pixels := 0, -1, -1
	for _, item := range play.Dash.Video {
		if queueStreamURL(item.BaseUrl, item.BaseUrl1) == "" {
			continue
		}
		if item.Bandwidth > bandwidth || item.Bandwidth == bandwidth && item.Width*item.Height > pixels {
			best, bandwidth, pixels = item.Id, item.Bandwidth, item.Width*item.Height
		}
	}
	return best
}

func bestQueueAudioID(play *model.VideoURLData) int {
	best, bandwidth := 0, -1
	for _, item := range play.Dash.Audio {
		if queueStreamURL(item.BaseUrl, item.BaseUrl1) != "" && item.Bandwidth > bandwidth {
			best, bandwidth = item.Id, item.Bandwidth
		}
	}
	return best
}

func queueStreamURL(primary, alternate string) string {
	if url := strings.TrimSpace(primary); url != "" {
		return url
	}
	return strings.TrimSpace(alternate)
}

func queueVideoOptions(play *model.VideoURLData) []input.SelectOption[int] {
	byID := make(map[int]model.VideoItem)
	for _, item := range play.Dash.Video {
		if queueStreamURL(item.BaseUrl, item.BaseUrl1) == "" {
			continue
		}
		old, found := byID[item.Id]
		if !found || item.Bandwidth > old.Bandwidth {
			byID[item.Id] = item
		}
	}
	if len(byID) == 0 && len(play.Durl) > 0 {
		return []input.SelectOption[int]{{Value: 0, Label: "默认视频流"}}
	}
	ids := make([]int, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))
	options := make([]input.SelectOption[int], 0, len(ids))
	for _, id := range ids {
		label := fmt.Sprintf("画质 %d", id)
		for _, format := range play.SupportFormats {
			if format.Quality == id {
				if format.DisplayDesc != "" {
					label = format.DisplayDesc
				} else if format.NewDescription != "" {
					label = format.NewDescription
				}
				break
			}
		}
		if id == 0 && byID[id].Height > 0 {
			label = fmt.Sprintf("%dP", byID[id].Height)
		}
		options = append(options, input.SelectOption[int]{Value: id, Label: label})
	}
	return options
}

func queueAudioOptions(play *model.VideoURLData) []input.SelectOption[int] {
	byID := make(map[int]model.AudioItem)
	for _, item := range play.Dash.Audio {
		if queueStreamURL(item.BaseUrl, item.BaseUrl1) == "" {
			continue
		}
		old, found := byID[item.Id]
		if !found || item.Bandwidth > old.Bandwidth {
			byID[item.Id] = item
		}
	}
	ids := make([]int, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return byID[ids[i]].Bandwidth > byID[ids[j]].Bandwidth })
	options := make([]input.SelectOption[int], 0, len(ids))
	for _, id := range ids {
		item := byID[id]
		label := fmt.Sprintf("%d kbps", item.Bandwidth/1000)
		if item.Codecs != "" {
			label += " · " + item.Codecs
		}
		options = append(options, input.SelectOption[int]{Value: id, Label: label})
	}
	return options
}

func queueVideoURL(play *model.VideoURLData, id int) string {
	url, bandwidth := "", -1
	for _, item := range play.Dash.Video {
		candidate := queueStreamURL(item.BaseUrl, item.BaseUrl1)
		if item.Id == id && candidate != "" && item.Bandwidth > bandwidth {
			url, bandwidth = candidate, item.Bandwidth
		}
	}
	if url == "" && len(play.Durl) > 0 {
		return strings.TrimSpace(play.Durl[0].Url)
	}
	return url
}

func queueAudioURL(play *model.VideoURLData, id int) string {
	url, bandwidth := "", -1
	for _, item := range play.Dash.Audio {
		candidate := queueStreamURL(item.BaseUrl, item.BaseUrl1)
		if item.Id == id && candidate != "" && item.Bandwidth > bandwidth {
			url, bandwidth = candidate, item.Bandwidth
		}
	}
	return url
}

func (s *State) queuedTask(key string) (bilidownload.Task, bool) {
	entry := s.queueResolves[key]
	if entry == nil || !entry.data.Loaded || entry.data.Data.PlayURL == nil || entry.data.Data.Detail == nil {
		return bilidownload.Task{}, false
	}
	result := entry.data.Data
	view := result.Detail.View
	bvid := key
	if item, ok := s.manualItem(key); ok {
		bvid = item.Bvid
	}
	task := bilidownload.Task{
		Bvid: bvid, Cid: result.Cid, Title: view.Title, Cover: view.Pic, SourceName: s.stagedSourceNames[key],
		UpperName: view.Owner.Name, Duration: view.Duration, Pubtime: view.Pubdate,
		Play: view.Stat.View, Danmaku: view.Stat.Danmaku,
		VideoURL: queueVideoURL(result.PlayURL, entry.videoID),
		AudioURL: queueAudioURL(result.PlayURL, entry.audioID),
	}
	if task.Cid == 0 {
		task.Cid = view.Cid
	}
	if item, ok := s.manualItem(key); ok {
		if item.Cid > 0 {
			task.Cid = item.Cid
		}
		if item.Title != "" {
			task.Title = item.Title
		}
		if item.Cover != "" {
			task.Cover = item.Cover
		}
		if item.Author != "" {
			task.UpperName = item.Author
		}
		if item.Duration > 0 {
			task.Duration = item.Duration
		}
		if item.Play > 0 {
			task.Play = item.Play
		}
		if item.Danmaku > 0 {
			task.Danmaku = item.Danmaku
		}
		if item.Pubtime > 0 {
			task.Pubtime = item.Pubtime
		}
	}
	for _, item := range s.stagedUploads {
		if item.Bvid == key {
			task.Title, task.Cover, task.UpperName = item.Title, item.Pic, item.Author
			task.Play, task.Danmaku, task.Pubtime = item.Play, item.VideoReview, item.Created
			break
		}
	}
	for _, item := range s.stagedSeasons {
		if item.Bvid == key {
			task.Title, task.Cover, task.Duration = item.Title, item.Pic, item.Duration
			task.Play, task.Danmaku, task.Pubtime = item.Stat.View, item.Stat.Danmaku, item.Pubdate
			break
		}
	}
	for _, item := range s.stagedSeries {
		if item.Bvid == key {
			task.Title, task.Cover, task.Duration = item.Title, item.Pic, item.Duration
			task.Play, task.Danmaku, task.Pubtime = item.Stat.View, item.Stat.Danmaku, item.PubDate
			break
		}
	}
	if task.UpperName == "" {
		task.UpperName = view.Owner.Name
	}
	return task, task.VideoURL != ""
}

type queuedDownload struct {
	key              string
	task             bilidownload.Task
	videoID, audioID int
	validUntil       time.Time
}
type manualDownloadRequest struct {
	order uint64
	tasks []queuedDownload
}

func (s *State) startQueuedDownloads()                { s.startQueuedDownloadKeys(s.stagedOrder()) }
func (s *State) startSingleQueuedDownload(key string) { s.startQueuedDownloadKeys([]string{key}) }

func (s *State) startQueuedDownloadKeys(keys []string) {
	if s.downloadBackend == nil || s.runtime.Background == nil || s.api == nil {
		return
	}
	tasks := make([]queuedDownload, 0, len(keys))
	for _, key := range keys {
		if s.queueRequested[key] {
			continue
		}
		task, ok := s.queuedTask(key)
		if !ok {
			s.queueDownloadError = "请等待视频解析完成后再下载"
			return
		}
		entry := s.queueResolves[key]
		// Reuse the preview streams while valid; refresh only after expiry.
		tasks = append(tasks, queuedDownload{key: key, task: task, videoID: entry.videoID, audioID: entry.audioID, validUntil: streamValidity(entry.resolvedAt, task.VideoURL, task.AudioURL)})
	}
	if len(tasks) == 0 {
		return
	}
	if s.queueRequested == nil {
		s.queueRequested = make(map[string]bool)
	}
	for _, task := range tasks {
		s.queueRequested[task.key] = true
	}
	s.nextDownloadOrder++
	s.pendingManualDownloads = append(s.pendingManualDownloads, manualDownloadRequest{order: s.nextDownloadOrder, tasks: tasks})
	s.queueDownloadError = ""
	s.pumpDownloads()
}

func (s *State) cancelPendingManualKey(key string) {
	for i := range s.pendingManualDownloads {
		s.pendingManualDownloads[i].tasks = slices.DeleteFunc(s.pendingManualDownloads[i].tasks, func(item queuedDownload) bool { return item.key == key })
	}
	s.pendingManualDownloads = slices.DeleteFunc(s.pendingManualDownloads, func(request manualDownloadRequest) bool { return len(request.tasks) == 0 })
	delete(s.queueRequested, key)
}

func (s *State) stopQueuedDownloads() {
	s.queueDownloadStop = true
	for _, request := range s.pendingManualDownloads {
		for _, item := range request.tasks {
			delete(s.queueRequested, item.key)
		}
	}
	s.pendingManualDownloads = nil
}

func (s *State) startQueueExtra(key, kind string) {
	backend, ok := s.downloadBackend.(DownloadExtras)
	if !ok || s.runtime.Background == nil || s.queueDownloading || s.batchesRunning() || s.queueExtraBusy[key] != "" || s.queueRequested[key] {
		return
	}
	task, ready := s.queuedTask(key)
	if !ready || kind == "audio" && task.AudioURL == "" || kind == "cover" && task.Cover == "" {
		return
	}
	s.queueExtraBusy[key] = kind
	s.queueDownloadError = ""
	epoch := s.epoch
	var path string
	s.runtime.Background(func() (err error) {
		if kind == "cover" {
			path, err = backend.DownloadCover(task.Cover, task)
		} else {
			path, err = backend.DownloadAudio(task)
		}
		return err
	}, func(err error) {
		if s.stopped || s.epoch != epoch {
			return
		}
		delete(s.queueExtraBusy, key)
		defer s.pumpDownloads()
		if err != nil {
			s.queueDownloadError = fmt.Sprintf("%s：%v", task.Title, err)
			return
		}
		label := "封面"
		if kind == "audio" {
			label = "音频"
		}
		s.notice = fmt.Sprintf("%s已保存：%s", label, path)
	})
}

type manualWork struct {
	tasks          []queuedDownload
	next, inFlight int
	epoch          uint64
}

func (s *State) advanceManualDownloads() {
	work := s.activeManualDownloads
	if work == nil {
		return
	}
	if !s.queueDownloadStop {
		// Consecutive manual requests can share slots, without overtaking an earlier batch.
		barrier := ^uint64(0)
		for _, b := range s.batches {
			if b.Active && b.Status == "queued" && b.Order < barrier {
				barrier = b.Order
			}
		}
		for len(s.pendingManualDownloads) > 0 && s.pendingManualDownloads[0].order < barrier {
			request := s.pendingManualDownloads[0]
			s.pendingManualDownloads = s.pendingManualDownloads[1:]
			work.tasks = append(work.tasks, request.tasks...)
			s.queueDownloadTotal += len(request.tasks)
		}
		for work.inFlight < s.downloadConcurrency() && work.next < len(work.tasks) {
			item := work.tasks[work.next]
			work.next++
			work.inFlight++
			s.runQueuedTask(work, item)
		}
	}
	if work.inFlight == 0 && (s.queueDownloadStop || work.next == len(work.tasks)) {
		for _, item := range work.tasks[work.next:] {
			delete(s.queueRequested, item.key)
		}
		s.activeManualDownloads = nil
		s.queueDownloading = false
		s.pumpDownloads()
	}
}

func (s *State) runQueuedTask(work *manualWork, item queuedDownload) {
	task := item.task
	delete(s.queueProgress, queueProgressKey(task.Bvid, task.Cid))
	resolveFailed := false
	api, backend := s.api, s.downloadBackend
	s.runtime.Background(func() error {
		if task.VideoURL == "" || !time.Now().Before(item.validUntil) {
			resolveFailed = true
			resolved := api.ResolvePlayUrl(biliapi.PlayUrlRequest{Bvid: task.Bvid, Cid: task.Cid}, 1)
			if resolved.Error != "" {
				return errors.New(resolved.Error)
			}
			if resolved.PlayURL == nil {
				return errors.New("播放地址解析结果不完整")
			}
			validVideo := false
			for _, option := range queueVideoOptions(resolved.PlayURL) {
				if option.Value == item.videoID {
					validVideo = true
					break
				}
			}
			if !validVideo {
				return errors.New("所选画质已不可用，请重新解析并选择画质")
			}
			task.VideoURL = queueVideoURL(resolved.PlayURL, item.videoID)
			task.AudioURL = queueAudioURL(resolved.PlayURL, item.audioID)
			if task.VideoURL == "" || item.audioID != 0 && task.AudioURL == "" {
				return errors.New("所选视频或音频流已不可用，请重新解析")
			}
		}
		resolveFailed = false
		result, err := backend.DownloadVideosByDash([]bilidownload.Task{task})
		return downloadResultError(result, err)
	}, func(err error) {
		if s.stopped || work.epoch != s.epoch {
			return
		}
		work.inFlight--
		s.queueDownloadDone++
		delete(s.queueRequested, item.key)
		if err != nil {
			s.queueDownloadFailed++
			s.queueDownloadError = fmt.Sprintf("%s：%v", task.Title, err)
			if entry := s.queueResolves[item.key]; entry != nil {
				entry.resolvedAt = time.Time{}
				if resolveFailed {
					entry.data.Loaded = false
					entry.data.Error = err.Error()
				}
			}
		} else {
			s.removeStaged(item.key)
		}
		s.advanceManualDownloads()
	})
}
