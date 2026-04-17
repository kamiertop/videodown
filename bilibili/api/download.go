package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/kamiertop/videodown/utils"
)

var invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

const (
	defaultBatchDownloadConcurrency = 2
	downloadedVideoCachePrefix      = "bilibili:downloaded:"
)

// DashDownloadTask 是前端提交给后端的最小下载任务；流地址已经由前端按用户选择的画质/音质确定。
type DashDownloadTask struct {
	SourceName string `json:"sourceName"`
	SourceKind string `json:"sourceKind"`
	UpperName  string `json:"upperName"`
	Bvid       string `json:"bvid"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Duration   int    `json:"duration"`
	Play       int    `json:"play"`
	Danmaku    int    `json:"danmaku"`
	Pubtime    int    `json:"pubtime"`
	VideoURL   string `json:"videoURL"`
	AudioURL   string `json:"audioURL"`
}

// DashDownloadResult 记录单个任务的结果，前端据此移除已完成的视频并保留失败项。
type DashDownloadResult struct {
	Bvid  string `json:"bvid"`
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

type DownloadHistoryItem struct {
	Bvid       string `json:"bvid"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Duration   int    `json:"duration"`
	UpperName  string `json:"upperName"`
	Play       int    `json:"play"`
	Danmaku    int    `json:"danmaku"`
	Pubtime    int    `json:"pubtime"`
	SourceName string `json:"sourceName"`
	SourceKind string `json:"sourceKind"`
	Path       string `json:"path"`
	// Wails 绑定生成不支持直接暴露 time.Time，保存为 RFC3339 字符串给前端解析。
	Downloaded string `json:"downloaded"`
}

func streamBaseURL(v string) string {
	return strings.TrimSpace(v)
}

func downloadCacheKey(bvid string) string {
	t := strings.ToUpper(strings.TrimSpace(bvid))
	if t == "" {
		return ""
	}
	return downloadedVideoCachePrefix + t
}

func sanitizeFilename(name string) string {
	t := strings.TrimSpace(name)
	if t == "" {
		return "video"
	}
	t = invalidFilenameChars.ReplaceAllString(t, "_")
	t = strings.Trim(t, " .")
	if t == "" {
		return "video"
	}
	return t
}

func sanitizeDirName(name string) string {
	t := strings.TrimSpace(name)
	if t == "" {
		return ""
	}
	t = invalidFilenameChars.ReplaceAllString(t, "_")
	t = strings.Trim(t, " .")
	return t
}

func uniqueFilePath(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s(%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

// downloadedCachePath 返回已下载缓存中的文件路径；缓存只在后端使用，不增加前端协议字段。
func (b *BiliBili) downloadedCachePath(bvid string) (string, bool) {
	key := downloadCacheKey(bvid)
	if key == "" {
		return "", false
	}

	raw, err := b.settings.GetKey(key)
	if err != nil {
		return "", false
	}

	var cached DownloadHistoryItem
	if err = json.Unmarshal([]byte(raw), &cached); err != nil {
		return "", false
	}
	return cached.Path, true
}

// markDownloaded 写入下载成功缓存；写缓存失败不影响已经完成的文件保存。
func (b *BiliBili) markDownloaded(task DashDownloadTask, path string) {
	key := downloadCacheKey(task.Bvid)
	if key == "" {
		return
	}

	payload, err := json.Marshal(DownloadHistoryItem{
		Bvid:       strings.TrimSpace(task.Bvid),
		Title:      strings.TrimSpace(task.Title),
		Cover:      strings.TrimSpace(task.Cover),
		Duration:   task.Duration,
		UpperName:  strings.TrimSpace(task.UpperName),
		Play:       task.Play,
		Danmaku:    task.Danmaku,
		Pubtime:    task.Pubtime,
		SourceName: strings.TrimSpace(task.SourceName),
		SourceKind: strings.TrimSpace(task.SourceKind),
		Path:       path,
		Downloaded: time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		b.logger.Errorf("marshal downloaded cache failed: %v", err)
		return
	}
	if err = b.settings.SetKey(key, string(payload)); err != nil {
		b.logger.Errorf("save downloaded cache failed: %v", err)
	}
}

// DownloadHistory 返回后端下载缓存记录；只读历史页使用，下载接口本身不暴露缓存命中细节。
func (b *BiliBili) DownloadHistory() ([]DownloadHistoryItem, error) {
	items := make([]DownloadHistoryItem, 0)
	prefix := []byte(downloadedVideoCachePrefix)

	err := b.settings.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var history DownloadHistoryItem
				if err := json.Unmarshal(bytes.Clone(val), &history); err != nil {
					return nil
				}
				items = append(items, history)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		return parseDownloadHistoryTime(items[i].Downloaded).After(parseDownloadHistoryTime(items[j].Downloaded))
	})
	return items, nil
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

	sourceName := sanitizeDirName(task.SourceName)
	if sourceName == "" {
		return storagePath, nil
	}

	// 合集需要更深一层分组：存储目录 / UP 主名 / 合集名。
	if task.SourceKind == "合集" {
		if upperName := sanitizeDirName(task.UpperName); upperName != "" {
			return filepath.Join(storagePath, upperName, sourceName), nil
		}
	}

	return filepath.Join(storagePath, sourceName), nil
}

type downloadProgress struct {
	Bvid       string  `json:"bvid"`
	Title      string  `json:"title"`
	Phase      string  `json:"phase"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percent    float64 `json:"percent"`
}

func progressKey(bvid string) string {
	return strings.ToUpper(strings.TrimSpace(bvid))
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

func (b *BiliBili) resetDownloadProgress(bvid string) {
	key := progressKey(bvid)
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
		return
	}
	key := progressKey(p.Bvid)
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

// downloadToFile 使用项目封装的 req client 流式读取响应体；手动 Read/Write 才能拿到实时字节进度。
func (b *BiliBili) downloadToFile(rawURL, targetPath, bvid, title, phase, cookies string, start, weight float64) error {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return errors.New("流地址无效")
	}

	resp, err := b.client.R().
		DisableAutoReadResponse().
		SetHeader(UserAgent, userAgent()).
		SetHeader(Referer, fmt.Sprintf("https://www.bilibili.com/video/%s", strings.TrimSpace(bvid))).
		SetHeader(Origin, BiliBiliUrl).
		SetHeader(Cookie, cookies).
		Get(rawURL)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 256*1024)
	lastEmit := time.Time{}
	b.emitDownloadProgress(downloadProgress{
		Bvid:    bvid,
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
	b.resetDownloadProgress(task.Bvid)
	if path, ok := b.downloadedCachePath(task.Bvid); ok {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "done", Percent: 100})
		return path, nil
	}

	task.VideoURL = streamBaseURL(task.VideoURL)
	task.AudioURL = streamBaseURL(task.AudioURL)
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

	outPath := uniqueFilePath(filepath.Join(targetDir, fileName+".mp4"))

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
	if err = b.downloadToFile(task.VideoURL, videoTmp, task.Bvid, task.Title, "video", cookies, 0, videoWeight); err != nil {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "error"})
		return "", err
	}

	ff := utils.NewFFmpeg()
	if task.AudioURL == "" {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "merge", Percent: 95})
		if err = ff.Remux(videoTmp, outPath); err != nil {
			b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "error"})
			return "", err
		}
		b.markDownloaded(task, outPath)
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "done", Percent: 100})
		return outPath, nil
	}

	audioTmp := filepath.Join(tmpDir, "audio.m4s")
	if err = b.downloadToFile(task.AudioURL, audioTmp, task.Bvid, task.Title, "audio", cookies, 60, 30); err != nil {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "error"})
		return "", err
	}
	b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "merge", Percent: 95})
	if err = ff.Merge(videoTmp, audioTmp, outPath); err != nil {
		b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "error"})
		return "", err
	}

	b.markDownloaded(task, outPath)
	b.emitDownloadProgress(downloadProgress{Bvid: task.Bvid, Title: task.Title, Phase: "done", Percent: 100})
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
func (b *BiliBili) sleepAfterTask() {
	sleepTime, err := b.settings.GetSleepTime()
	if err != nil || sleepTime <= 0 {
		return
	}

	sleepTime = rand.Int63n(sleepTime)
	b.logger.Infof("download success, sleep %d second", sleepTime)

	time.Sleep(time.Duration(sleepTime) * time.Second)
}

func uniqueDashDownloadTasks(tasks []DashDownloadTask) []DashDownloadTask {
	seen := make(map[string]struct{}, len(tasks))
	unique := make([]DashDownloadTask, 0, len(tasks))
	for _, task := range tasks {
		key := strings.ToUpper(strings.TrimSpace(task.Bvid))
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		// 同一个 BV 只允许进入一个下载任务，避免并发任务向同一个前端进度条交替推送。
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

	workerCount := defaultBatchDownloadConcurrency
	if len(tasks) < workerCount {
		workerCount = len(tasks)
	}

	jobs := make(chan DashDownloadTask)
	results := make(chan DashDownloadResult, len(tasks))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				path, err := b.downloadDashTask(task)
				item := DashDownloadResult{Bvid: task.Bvid, Title: task.Title, Path: path}
				if err != nil {
					item.Error = err.Error()
				}
				results <- item
				b.sleepAfterTask()
			}
		}()
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
