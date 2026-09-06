package download

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kamiertop/videodown/bilibili/util"
	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"

	"github.com/imroc/req/v3"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Service struct {
	store     *storage.Store
	logger    *logger.Logger
	events    *application.EventManager
	client    *req.Client
	mu        sync.Mutex
	progress  map[string]float64
	getCookie func() (string, error)
}

func NewService(logger *logger.Logger, store *storage.Store, events *application.EventManager, getCookie func() (string, error)) *Service {
	return &Service{
		store:     store,
		logger:    logger,
		events:    events,
		progress:  make(map[string]float64),
		mu:        sync.Mutex{},
		getCookie: getCookie,
		client: req.
			C().
			EnableAutoDecompress().
			SetJsonMarshal(func(v any) ([]byte, error) {
				return json.Marshal(v)
			}).
			SetJsonUnmarshal(func(data []byte, v any) error {
				return json.Unmarshal(data, v)
			}).
			SetCommonRetryCount(3).
			SetTimeout(0), // 不设置超时，因为长视频下载可能需要很长时间，超时会导致下载中断
	}
}

// resetDownloadProgress 下载开始前重置进度缓存，避免同一 BV 多次下载时进度事件被旧任务覆盖
// 单 P 直接按 BV 区分，多 P 按 BV:CID 区分，确保不同 P 之间的进度事件互不干扰。
func (s *Service) resetDownloadProgress(bvid string, cid int64) {
	s.mu.Lock()
	delete(s.progress, progressKey(bvid, cid))
	s.mu.Unlock()
}

// emitProgress 通过 Wails 事件把下载进度推给前端，前端按 bvid 更新对应卡片。
func (s *Service) emitProgress(p progress) {
	key := progressKey(p.Bvid, p.Cid)
	if key != "" {
		p.Percent = utils.ClampPercent(p.Percent)

		s.mu.Lock()
		prev, ok := s.progress[key]
		switch p.Phase {
		case "done":
			p.Percent = 100
			s.progress[key] = 100
		case "error":
			// 失败状态需要传给前端，但不能把进度条从已下载位置拉回 0。
			if ok && p.Percent < prev {
				p.Percent = prev
			}
			s.progress[key] = p.Percent
		default:
			if ok && p.Percent < prev {
				s.mu.Unlock()
				return
			}
			s.progress[key] = p.Percent
		}
		s.mu.Unlock()
	}
	if s.events != nil {
		s.events.Emit("bilibili-download-progress", p)
	}
}

// emitDownloadCompleted 下载+休眠全部完成后推送给前端，前端收到后立即从列表移除该视频卡片。
func (s *Service) emitDownloadCompleted(item Result) {
	if s.events != nil {
		s.events.Emit("bilibili-download-completed", item)
	}
}

// downloadToFile 使用无总时长限制的 HTTP client 流式读取响应体；长视频下载不能复用接口请求的整体超时。
func (s *Service) downloadToFile(rawURL, targetPath, bvid, title string, cid int64, phase, cookies string, start, weight float64) (err error) {
	var resp *req.Response
	resp, err = s.client.
		R().
		DisableAutoReadResponse().
		SetHeader(constant.UserAgent, util.UserAgent()).
		SetHeader(constant.Referer, fmt.Sprintf("https://www.bilibili.com/video/%s", strings.TrimSpace(bvid))).
		SetHeader(constant.Origin, util.BiliBiliURL).
		SetHeader(constant.Cookie, cookies).
		Get(rawURL)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 1024*1024)
	lastEmit := time.Time{}
	// 下载开始时先发一次事件，确保前端能及时更新状态；后续按时间间隔或下载完成时发事件
	s.emitProgress(progress{
		Bvid:    bvid,
		Cid:     cid,
		Title:   title,
		Phase:   phase,
		Total:   total,
		Percent: start,
	})

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err = f.Write(buf[:n]); err != nil {
				return err
			}
			downloaded += int64(n)

			now := time.Now()
			if now.Sub(lastEmit) >= 200*time.Millisecond || (total > 0 && downloaded >= total) {
				s.emitProgress(progress{
					Bvid:       bvid,
					Cid:        cid,
					Title:      title,
					Phase:      phase,
					Downloaded: downloaded,
					Total:      total,
					Percent:    utils.WeightedPercent(start, weight, downloaded, total),
				})
				lastEmit = now
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return readErr
		}
	}

	s.emitProgress(progress{
		Bvid:       bvid,
		Cid:        cid,
		Title:      title,
		Phase:      phase,
		Downloaded: downloaded,
		Total:      total,
		Percent:    start + weight,
	})
	return nil
}

// downloadDashTask 下载一个 DASH 任务，供单个下载和批量下载复用。
func (s *Service) downloadDashTask(task Task) (string, error) {
	s.resetDownloadProgress(task.Bvid, task.Cid)
	// 下载前先检查是否已下载，避免重复下载同一 CID 的视频
	if path, ok := s.isDownloaded(task.Cid); ok {
		s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
		return path, nil
	}

	if task.VideoURL == "" {
		return "", errors.New("视频流地址为空")
	}

	cookies, err := s.getCookie()
	if err != nil {
		return "", err
	}

	storagePath, err := s.store.StoragePath()
	if err != nil {
		return "", err
	}

	targetDir, err := s.resolveTargetDir(storagePath, task)
	if err != nil {
		return "", err
	}

	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		return "", errors.New("创建下载目录失败")
	}

	template, _ := s.store.FilenameTemplate()
	now := time.Now()
	values := map[string]string{"title": task.Title, "id": task.Bvid, "author": task.UpperName, "author_id": "", "source": task.SourceName, "folder": "", "collection": task.SourceName, "publish_date": "", "date": now.Format("2006-01-02"), "time": now.Format("15-04-05")}
	if task.Pubtime > 0 {
		values["publish_date"] = time.Unix(int64(task.Pubtime), 0).Format("2006-01-02")
	}
	fileName := utils.ApplyFilenameTemplate(template, values)
	if fileName == "" {
		fileName = sanitizeFilename(task.Title)
	}
	if strings.TrimSpace(task.Bvid) == "" {
		task.Bvid = "BV_UNKNOWN"
	}

	outPath := utils.UniqueFilePath(filepath.Join(targetDir, fileName+".mp4"))

	tmpDir := filepath.Join(storagePath, ".tmp", fmt.Sprintf("%s-%d", task.Bvid, time.Now().UnixNano()))
	if err = os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", errors.New("创建临时目录失败")
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	videoTmp := filepath.Join(tmpDir, "video.m4s")
	videoWeight := 60.0
	if task.AudioURL == "" {
		videoWeight = 90.0
	}
	if err = s.downloadToFile(task.VideoURL, videoTmp, task.Bvid, task.Title, task.Cid, "video", cookies, 0, videoWeight); err != nil {
		s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
		return "", err
	}

	ff := utils.NewFFmpeg()
	if task.AudioURL == "" {
		s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "merge", Percent: 95})
		if err = ff.Remux(videoTmp, outPath); err != nil {
			s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
			return "", err
		}
		s.markDownloaded(task, outPath, kindVideo)
		s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
		return outPath, nil
	}

	audioTmp := filepath.Join(tmpDir, "audio.m4s")
	if err = s.downloadToFile(task.AudioURL, audioTmp, task.Bvid, task.Title, task.Cid, "audio", cookies, 60, 30); err != nil {
		s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
		return "", err
	}
	s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "merge", Percent: 95})
	if err = ff.Merge(videoTmp, audioTmp, outPath); err != nil {
		s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
		return "", err
	}

	s.markDownloaded(task, outPath, kindVideo)
	s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
	return outPath, nil
}

// DownloadVideoByDash 下载单个视频，保留旧入口以兼容前端或其他调用方。
func (s *Service) DownloadVideoByDash(sourceName, bvid, title, videoURL, audioURL string) (string, error) {
	return s.downloadDashTask(Task{
		SourceName: sourceName,
		UpperName:  "",
		Bvid:       bvid,
		Title:      title,
		VideoURL:   videoURL,
		AudioURL:   audioURL,
	})
}

// sleepAfterTask 按设置项在同一个 worker 中休眠，避免连续请求过快；并发 worker 互不阻塞。
// 根据设置的时间上下浮动
func (s *Service) sleepAfterTask(task Task) {
	sleepTime, err := s.store.SleepTime()
	if err != nil || sleepTime <= 0 {
		return
	}

	sleepTime = rand.Int64N(sleepTime)
	if sleepTime <= 0 {
		return
	}
	s.logger.Infof("download success, sleep %d second", sleepTime)

	for remaining := sleepTime; remaining > 0; remaining-- {
		s.emitProgress(progress{
			Bvid:           task.Bvid,
			Cid:            task.Cid,
			Title:          task.Title,
			Phase:          "sleep",
			Percent:        100,
			SleepRemaining: remaining,
			SleepTotal:     sleepTime,
		})
		time.Sleep(time.Second)
	}
	s.emitProgress(progress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
}

// dashDownloadDedupKey 生成下载去重键；同一 BV 单 P 只保留一个任务，多 P 按 cid 区分，避免进度事件互相覆盖
func dashDownloadDedupKey(task Task) string {
	bv := strings.ToUpper(strings.TrimSpace(task.Bvid))
	if bv == "" {
		return ""
	}
	if task.Cid > 0 {
		return bv + ":" + strconv.FormatInt(task.Cid, 10)
	}
	return bv
}

// uniqueDashDownloadTasks 对下载任务列表去重；同一 BV 单 P 只保留一个任务，多 P 按 cid 区分，避免进度事件互相覆盖
func uniqueDashDownloadTasks(tasks []Task) []Task {
	seen := make(map[string]struct{}, len(tasks))
	unique := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		key := dashDownloadDedupKey(task)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		// 同一 BV 在单 P 时只保留一个任务；多 P 时按 cid 区分，避免进度事件互相覆盖。
		unique = append(unique, task)
	}

	return unique
}

// DownloadVideosByDash 批量下载 DASH 任务；后端负责并发、失败隔离和每个任务后的休眠控制。
func (s *Service) DownloadVideosByDash(tasks []Task) (BatchResult, error) {
	tasks = uniqueDashDownloadTasks(tasks)
	result := BatchResult{
		Results: make([]Result, 0, len(tasks)),
	}
	if len(tasks) == 0 {
		return result, errors.New("下载列表为空")
	}

	workerCount, err := s.store.ConcurrencyNum()
	if err != nil {
		return result, err
	}

	jobs := make(chan Task, len(tasks))
	results := make(chan Result, len(tasks))
	var wg sync.WaitGroup
	// 启动固定数量的 worker 并发下载，worker 数量由设置项控制；每个 worker 从 jobs 通道接收任务，完成后把结果发送到 results 通道
	for range workerCount {
		wg.Go(func() {
			for task := range jobs {
				// 每个任务独立下载，失败不影响其他任务；下载完成后根据设置项休眠，避免连续请求过快；并发 worker 互不阻塞
				path, err := s.downloadDashTask(task)
				item := Result{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Path: path}
				if err != nil {
					s.logger.Errorf("download failed for BV %s CID %d: %v", task.Bvid, task.Cid, err)
					item.Error = err.Error()
				}
				// 下载完成后把下载结果发送到 results 通道，供主协程统计成功失败数量和返回给前端
				results <- item
				if err == nil {
					// 下载成功才休眠，下载失败立即开始下一个任务，避免连续下载失败时长时间无响应
					s.sleepAfterTask(task)
					// 休眠完成后逐条推送完成事件，前端收到后立即从列表移除该视频
					s.emitDownloadCompleted(item)
				}
			}
		})
	}

	for _, task := range tasks {
		jobs <- task
	}
	close(jobs)
	wg.Wait()
	close(results)

	for item := range results {
		if item.Error == "" {
			result.Success += 1
		} else {
			result.Failed += 1
		}
		result.Results = append(result.Results, item)
	}

	return result, nil
}

// FilterIncrementalTasks 将任务分为待下载和已下载两组；已下载项附带当时保存的文件路径。
func (s *Service) FilterIncrementalTasks(tasks []Task) (toDownload []Task, alreadyDone []Result) {
	tasks = uniqueDashDownloadTasks(tasks)
	for _, task := range tasks {
		if path, ok := s.isDownloaded(task.Cid); ok {
			alreadyDone = append(alreadyDone, Result{
				Bvid:  task.Bvid,
				Cid:   task.Cid,
				Title: task.Title,
				Path:  path,
			})
		} else {
			toDownload = append(toDownload, task)
		}
	}
	return
}

// DownloadVideosByDashIncremental 增量下载：跳过已下载的视频，只下载新增部分。
func (s *Service) DownloadVideosByDashIncremental(tasks []Task) (BatchResult, error) {
	toDownload, alreadyDone := s.FilterIncrementalTasks(tasks)

	result := BatchResult{
		Results: make([]Result, 0, len(tasks)),
	}

	for _, item := range alreadyDone {
		result.Results = append(result.Results, item)
		result.Success += 1
	}

	if len(toDownload) == 0 {
		return result, nil
	}

	batchResult, err := s.DownloadVideosByDash(toDownload)
	if err != nil {
		return result, err
	}

	result.Results = append(result.Results, batchResult.Results...)
	result.Success += batchResult.Success
	result.Failed += batchResult.Failed

	return result, nil
}

// DownloadCover 下载视频封面到当前下载目录，返回保存后的文件路径。
func (s *Service) DownloadCover(cover string, task Task) (string, error) {
	cover = normalizeHTTPURL(cover)
	if !strings.HasPrefix(cover, "http://") && !strings.HasPrefix(cover, "https://") {
		return "", errors.New("封面地址无效")
	}

	storagePath, err := s.store.StoragePath()
	if err != nil {
		return "", err
	}
	targetDir, err := s.resolveTargetDir(storagePath, task)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		return "", errors.New("创建下载目录失败")
	}

	resp, err := s.client.
		R().
		SetHeader(constant.UserAgent, util.UserAgent()).
		SetHeader(constant.Referer, util.BiliBiliURL).
		SetHeader(constant.Accept, "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8").
		Get(cover)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("封面下载失败: %s", resp.Status)
	}

	ext := utils.ImageExtFromResponse(cover, resp.Response)
	fileName := sanitizeFilename(task.Title)
	if fileName == "video" {
		fileName = "cover"
	}
	outPath := utils.UniqueFilePath(filepath.Join(targetDir, fileName+ext))
	if err = os.WriteFile(outPath, resp.Bytes(), 0o644); err != nil {
		return "", err
	}
	if task.Cid > 0 {
		task.Cid = -task.Cid
		s.markDownloaded(task, outPath, kindCover)
	}

	return outPath, nil
}
