package api

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kamiertop/videodown/bilibili/model"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/kamiertop/videodown/utils"
)

const (
	downloadedVideoCachePrefix = "bilibili:downloaded:"
)

const (
	bilibiliDownloadKindVideo = "video"
	bilibiliDownloadKindCover = "cover"
)

// DashDownloadTask 是前端提交给后端的最小下载任务；流地址已经由前端按用户选择的画质/音质确定。
type DashDownloadTask struct {
	SourceName string `json:"sourceName"`
	SourceKind string `json:"sourceKind"`
	UpperName  string `json:"upperName"`
	Bvid       string `json:"bvid"`
	// Cid 为分 P 稿件某一 P 的 cid；单 P 或未填时为 0，行为与旧版一致。
	Cid      int64  `json:"cid"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Duration int    `json:"duration"`
	Play     int    `json:"play"`
	Danmaku  int    `json:"danmaku"`
	Pubtime  int    `json:"pubtime"`
	VideoURL string `json:"videoURL"`
	AudioURL string `json:"audioURL"`
}

// DashDownloadResult 记录单个任务的结果，前端据此移除已完成的视频并保留失败项。
type DashDownloadResult struct {
	Bvid  string `json:"bvid"`
	Cid   int64  `json:"cid"`
	Title string `json:"title"`
	Path  string `json:"path"`
	Error string `json:"error"`
}

// DashDownloadBatchResult 汇总批量下载结果；失败不会中断整批任务。
type DashDownloadBatchResult struct {
	Results []DashDownloadResult `json:"results"`
	Success int                  `json:"success"`
	Failed  int                  `json:"failed"`
}

func normalizeHTTPURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}
	return rawURL
}

func downloadCacheKey(cid int64) string {
	if cid == 0 {
		return ""
	}
	if cid < 0 {
		return downloadedVideoCachePrefix + "cover:" + strconv.FormatInt(-cid, 10)
	}
	return downloadedVideoCachePrefix + strconv.FormatInt(cid, 10)
}

func sanitizeFilename(name string) string {
	t := utils.FileName(name)
	if t == "" {
		return "video"
	}
	return t
}

// downloadedCachePath 返回已下载缓存中的文件路径；缓存只在后端使用，不增加前端协议字段。
func (b *BiliBili) downloadedCachePath(cid int64) (string, bool) {
	key := downloadCacheKey(cid)
	if key == "" {
		return "", false
	}

	raw, err := b.settings.GetKey(key)
	if err != nil {
		return "", false
	}

	var cached model.DownloadHistoryItem
	if err = sonic.Unmarshal([]byte(raw), &cached); err != nil {
		return "", false
	}
	return cached.Path, true
}

// isDownloaded 检查缓存记录且确认文件仍存在于磁盘；缓存可能因手动删文件而过期。
func (b *BiliBili) isDownloaded(cid int64) (string, bool) {
	path, ok := b.downloadedCachePath(cid)
	if !ok {
		return "", false
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", false
	}
	return path, true
}

// markDownloaded 写入下载成功缓存；写缓存失败不影响已经完成的文件保存。
func (b *BiliBili) markDownloaded(task DashDownloadTask, path string, downloadKind string) {
	key := downloadCacheKey(task.Cid)
	if key == "" {
		return
	}
	if downloadKind == "" {
		downloadKind = bilibiliDownloadKindVideo
	}

	payload, err := sonic.Marshal(model.DownloadHistoryItem{
		Bvid:         strings.TrimSpace(task.Bvid),
		Cid:          task.Cid,
		Title:        strings.TrimSpace(task.Title),
		Cover:        strings.TrimSpace(task.Cover),
		Duration:     task.Duration,
		UpperName:    strings.TrimSpace(task.UpperName),
		Play:         task.Play,
		Danmaku:      task.Danmaku,
		Pubtime:      task.Pubtime,
		SourceName:   strings.TrimSpace(task.SourceName),
		SourceKind:   strings.TrimSpace(task.SourceKind),
		Path:         path,
		DownloadKind: downloadKind,
		Downloaded:   time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		b.logger.Errorf("marshal downloaded cache failed: %v", err)
		return
	}
	if err = b.settings.SetKey(key, string(payload)); err != nil {
		b.logger.Errorf("save downloaded cache failed: %v", err)
	}
}

func parseDownloadHistoryTime(value string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}

// resolveTargetDir 只在后端决定落盘目录，避免前端把目录规则写死。
func (b *BiliBili) resolveTargetDir(storagePath string, task DashDownloadTask) (string, error) {
	allowGroup, err := b.settings.GetSavePreference()
	if err != nil {
		return "", err
	}
	if !allowGroup {
		return storagePath, nil
	}

	kind := strings.TrimSpace(task.SourceKind)
	sourceName := utils.FileName(task.SourceName)
	upperName := utils.FileName(task.UpperName)

	switch kind {
	case "解析结果":
		return storagePath, nil
	case "全部投稿":
		return filepath.Join(storagePath, upperName), nil
	case "合集", "系列", "分P":
		return filepath.Join(storagePath, upperName, sourceName), nil
	case "收藏夹":
		return filepath.Join(storagePath, sourceName), nil
	default:
		return filepath.Join(storagePath, sourceName), nil
	}
}

type downloadProgress struct {
	Bvid           string  `json:"bvid"`
	Cid            int64   `json:"cid"`
	Title          string  `json:"title"`
	Phase          string  `json:"phase"`
	Downloaded     int64   `json:"downloaded"`
	Total          int64   `json:"total"`
	Percent        float64 `json:"percent"`
	SleepRemaining int64   `json:"sleepRemaining"`
	SleepTotal     int64   `json:"sleepTotal"`
}

func progressKey(bvid string, cid int64) string {
	bv := strings.ToUpper(strings.TrimSpace(bvid))
	if bv == "" {
		return ""
	}
	if cid > 0 {
		return bv + ":" + strconv.FormatInt(cid, 10)
	}
	return bv
}

func clampPercent(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

// resetDownloadProgress 下载开始前重置进度缓存，避免同一 BV 多次下载时进度事件被旧任务覆盖
// 单 P 直接按 BV 区分，多 P 按 BV:CID 区分，确保不同 P 之间的进度事件互不干扰。
func (b *BiliBili) resetDownloadProgress(bvid string, cid int64) {
	key := progressKey(bvid, cid)
	if key == "" {
		return
	}
	b.progressMu.Lock()
	delete(b.progressByBvid, key)
	b.progressMu.Unlock()
}

// emitDownloadProgress 通过 Wails 事件把下载进度推给前端，前端按 bvid 更新对应卡片。
func (b *BiliBili) emitDownloadProgress(p downloadProgress) {
	ctx := b.context()
	if ctx == nil {
		b.logger.Errorf("emitDownloadProgress failed: context is nil")
		return
	}
	key := progressKey(p.Bvid, p.Cid)
	if key != "" {
		p.Percent = clampPercent(p.Percent)

		b.progressMu.Lock()
		prev, ok := b.progressByBvid[key]
		switch p.Phase {
		case "done":
			p.Percent = 100
			b.progressByBvid[key] = 100
		case "error":
			// 失败状态需要传给前端，但不能把进度条从已下载位置拉回 0。
			if ok && p.Percent < prev {
				p.Percent = prev
			}
			b.progressByBvid[key] = p.Percent
		default:
			if ok && p.Percent < prev {
				b.progressMu.Unlock()
				return
			}
			b.progressByBvid[key] = p.Percent
		}
		b.progressMu.Unlock()
	}
	wailsRuntime.EventsEmit(ctx, "bilibili-download-progress", p)
}

// weightedPercent 把单个阶段的字节进度映射到整条任务进度，便于前端展示统一进度条。
func weightedPercent(start, weight float64, downloaded, total int64) float64 {
	if total <= 0 {
		return start
	}
	ratio := float64(downloaded) / float64(total)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	return start + ratio*weight
}

// downloadToFile 使用无总时长限制的 HTTP client 流式读取响应体；长视频下载不能复用接口请求的整体超时。
func (b *BiliBili) downloadToFile(rawURL, targetPath, bvid, title string, cid int64, phase, cookies string, start, weight float64) (err error) {
	resp, err := b.downloadClient.R().
		DisableAutoReadResponse().
		SetRetryCount(2).
		SetHeader(UserAgent, userAgent()).
		SetHeader(Referer, fmt.Sprintf("https://www.bilibili.com/video/%s", strings.TrimSpace(bvid))).
		SetHeader(Origin, biliBiliUrl).
		SetHeader(Cookie, cookies).
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
	buf := make([]byte, 256*1024)
	lastEmit := time.Time{}
	b.emitDownloadProgress(downloadProgress{
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
				b.emitDownloadProgress(downloadProgress{
					Bvid:       bvid,
					Cid:        cid,
					Title:      title,
					Phase:      phase,
					Downloaded: downloaded,
					Total:      total,
					Percent:    weightedPercent(start, weight, downloaded, total),
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

	b.emitDownloadProgress(downloadProgress{
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
func (b *BiliBili) downloadDashTask(task DashDownloadTask) (string, error) {
	b.resetDownloadProgress(task.Bvid, task.Cid)
	// 下载前先检查是否已下载，避免重复下载同一 CID 的视频
	if path, ok := b.isDownloaded(task.Cid); ok {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
		return path, nil
	}

	if task.VideoURL == "" {
		return "", errors.New("视频流地址为空")
	}

	cookies, err := b.getCookies()
	if err != nil {
		return "", err
	}

	storagePath, err := b.settings.GetStorage()
	if err != nil {
		return "", err
	}

	targetDir, err := b.resolveTargetDir(storagePath, task)
	if err != nil {
		return "", err
	}

	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		return "", errors.New("创建下载目录失败")
	}

	fileName := sanitizeFilename(task.Title)
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
	if err = b.downloadToFile(task.VideoURL, videoTmp, task.Bvid, task.Title, task.Cid, "video", cookies, 0, videoWeight); err != nil {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
		return "", err
	}

	ff := utils.NewFFmpeg()
	if task.AudioURL == "" {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "merge", Percent: 95})
		if err = ff.Remux(videoTmp, outPath); err != nil {
			b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
			return "", err
		}
		b.markDownloaded(task, outPath, bilibiliDownloadKindVideo)
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
		return outPath, nil
	}

	audioTmp := filepath.Join(tmpDir, "audio.m4s")
	if err = b.downloadToFile(task.AudioURL, audioTmp, task.Bvid, task.Title, task.Cid, "audio", cookies, 60, 30); err != nil {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
		return "", err
	}
	b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "merge", Percent: 95})
	if err = ff.Merge(videoTmp, audioTmp, outPath); err != nil {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "error"})
		return "", err
	}

	b.markDownloaded(task, outPath, bilibiliDownloadKindVideo)
	b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
	return outPath, nil
}

// DownloadVideoByDash 下载单个视频，保留旧入口以兼容前端或其他调用方。
func (b *BiliBili) DownloadVideoByDash(sourceName, bvid, title, videoURL, audioURL string) (string, error) {
	return b.downloadDashTask(DashDownloadTask{
		SourceName: sourceName,
		SourceKind: "",
		UpperName:  "",
		Bvid:       bvid,
		Title:      title,
		VideoURL:   videoURL,
		AudioURL:   audioURL,
	})
}

// sleepAfterTask 按设置项在同一个 worker 中休眠，避免连续请求过快；并发 worker 互不阻塞。
// 根据设置的时间上下浮动
func (b *BiliBili) sleepAfterTask(task DashDownloadTask) {
	sleepTime, err := b.settings.GetSleepTime()
	if err != nil || sleepTime <= 0 {
		return
	}

	sleepTime = rand.Int63n(sleepTime)
	if sleepTime <= 0 {
		return
	}
	b.logger.Infof("download success, sleep %d second", sleepTime)

	for remaining := sleepTime; remaining > 0; remaining-- {
		b.emitDownloadProgress(downloadProgress{
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
	b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Phase: "done", Percent: 100})
}

// dashDownloadDedupKey 生成下载去重键；同一 BV 单 P 只保留一个任务，多 P 按 cid 区分，避免进度事件互相覆盖
func dashDownloadDedupKey(task DashDownloadTask) string {
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
func uniqueDashDownloadTasks(tasks []DashDownloadTask) []DashDownloadTask {
	seen := make(map[string]struct{}, len(tasks))
	unique := make([]DashDownloadTask, 0, len(tasks))
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
func (b *BiliBili) DownloadVideosByDash(tasks []DashDownloadTask) (DashDownloadBatchResult, error) {
	tasks = uniqueDashDownloadTasks(tasks)
	result := DashDownloadBatchResult{
		Results: make([]DashDownloadResult, 0, len(tasks)),
	}
	if len(tasks) == 0 {
		return result, errors.New("下载列表为空")
	}

	workerCount, err := b.settings.GetConcurrencyNum()
	if err != nil {
		return result, err
	}

	jobs := make(chan DashDownloadTask, len(tasks))
	results := make(chan DashDownloadResult, len(tasks))
	var wg sync.WaitGroup
	// 启动固定数量的 worker 并发下载，worker 数量由设置项控制；每个 worker 从 jobs 通道接收任务，完成后把结果发送到 results 通道
	for range workerCount {
		wg.Go(func() {
			for task := range jobs {
				// 每个任务独立下载，失败不影响其他任务；下载完成后根据设置项休眠，避免连续请求过快；并发 worker 互不阻塞
				path, err := b.downloadDashTask(task)
				item := DashDownloadResult{Bvid: task.Bvid, Cid: task.Cid, Title: task.Title, Path: path}
				if err != nil {
					b.logger.Errorf("download failed for BV %s CID %d: %v", task.Bvid, task.Cid, err)
					item.Error = err.Error()
				}
				// 下载完成后把下载结果发送到 results 通道，供主协程统计成功失败数量和返回给前端
				results <- item
				if err == nil {
					// 下载成功才休眠，下载失败立即开始下一个任务，避免连续下载失败时长时间无响应
					b.sleepAfterTask(task)
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
func (b *BiliBili) FilterIncrementalTasks(tasks []DashDownloadTask) (toDownload []DashDownloadTask, alreadyDone []DashDownloadResult) {
	tasks = uniqueDashDownloadTasks(tasks)
	for _, task := range tasks {
		if path, ok := b.isDownloaded(task.Cid); ok {
			alreadyDone = append(alreadyDone, DashDownloadResult{
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
func (b *BiliBili) DownloadVideosByDashIncremental(tasks []DashDownloadTask) (DashDownloadBatchResult, error) {
	toDownload, alreadyDone := b.FilterIncrementalTasks(tasks)

	result := DashDownloadBatchResult{
		Results: make([]DashDownloadResult, 0, len(tasks)),
	}

	for _, item := range alreadyDone {
		result.Results = append(result.Results, item)
		result.Success += 1
	}

	if len(toDownload) == 0 {
		return result, nil
	}

	batchResult, err := b.DownloadVideosByDash(toDownload)
	if err != nil {
		return result, err
	}

	result.Results = append(result.Results, batchResult.Results...)
	result.Success += batchResult.Success
	result.Failed += batchResult.Failed

	return result, nil
}

// DownloadCover 下载视频封面到当前下载目录，返回保存后的文件路径。
func (b *BiliBili) DownloadCover(cover string, task DashDownloadTask) (string, error) {
	cover = normalizeHTTPURL(cover)
	if !strings.HasPrefix(cover, "http://") && !strings.HasPrefix(cover, "https://") {
		return "", errors.New("封面地址无效")
	}

	storagePath, err := b.settings.GetStorage()
	if err != nil {
		return "", err
	}
	targetDir := storagePath
	if upperName := utils.FileName(task.UpperName); upperName != "" {
		targetDir = filepath.Join(storagePath, upperName)
	}
	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		return "", errors.New("创建下载目录失败")
	}

	resp, err := b.client.R().
		SetRetryCount(2).
		SetHeader(UserAgent, userAgent()).
		SetHeader(Referer, biliBiliUrl).
		SetHeader(Accept, "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8").
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
		b.markDownloaded(task, outPath, bilibiliDownloadKindCover)
	}

	return outPath, nil
}
