package api

import (
	"bytes"
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

	"github.com/bytedance/sonic"
	"github.com/dgraph-io/badger/v4"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var douyinInvalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

const douyinDownloadedCachePrefix = "douyin:downloaded:"

// DouyinDownloadTask 是前端提交给后端的最小任务协议。
// 前端已经完成清晰度选择，所以后端只消费最终 videoURL；图片合集则使用 ImageURLs。
type DouyinDownloadTask struct {
	AwemeID      string                `json:"awemeId"`
	SourceKind   string                `json:"sourceKind"`
	SourceName   string                `json:"sourceName"`
	Title        string                `json:"title"`
	Cover        string                `json:"cover"`
	Duration     int                   `json:"duration"`
	AuthorName   string                `json:"authorName"`
	PublishTime  int                   `json:"publishTime"`
	DiggCount    int                   `json:"diggCount"`
	CollectCount int                   `json:"collectCount"`
	VideoURL     string                `json:"videoURL"`
	ImageURLs    []string              `json:"imageURLs"`
	Assets       []DouyinDownloadAsset `json:"assets"`
	MusicURL     string                `json:"musicURL"`
}

type DouyinDownloadAsset struct {
	URL  string `json:"url"`
	Kind string `json:"kind"`
	Ext  string `json:"ext"`
}

type DouyinDownloadResult struct {
	AwemeID string `json:"awemeId"`
	Title   string `json:"title"`
	Path    string `json:"path"`
	Error   string `json:"error"`
}

type DouyinDownloadBatchResult struct {
	Results []DouyinDownloadResult `json:"results"`
	Success int                    `json:"success"`
	Failed  int                    `json:"failed"`
}

type DouyinDownloadHistoryItem struct {
	AwemeID      string `json:"awemeId"`
	Title        string `json:"title"`
	Cover        string `json:"cover"`
	Duration     int    `json:"duration"`
	AuthorName   string `json:"authorName"`
	PublishTime  int    `json:"publishTime"`
	DiggCount    int    `json:"diggCount"`
	CollectCount int    `json:"collectCount"`
	SourceName   string `json:"sourceName"`
	SourceKind   string `json:"sourceKind"`
	Path         string `json:"path"`
	IsImageAlbum bool   `json:"isImageAlbum"`
	ImageCount   int    `json:"imageCount"`
	// Wails 绑定生成不支持直接暴露 time.Time，保存为 RFC3339 字符串给前端解析。
	Downloaded string `json:"downloaded"`
}

// douyinDownloadProgress 通过 Wails 事件推给前端，前端按 awemeId 更新对应卡片进度。
type douyinDownloadProgress struct {
	AwemeID        string  `json:"awemeId"`
	Title          string  `json:"title"`
	Phase          string  `json:"phase"`
	Downloaded     int64   `json:"downloaded"`
	Total          int64   `json:"total"`
	Percent        float64 `json:"percent"`
	SleepRemaining int64   `json:"sleepRemaining"`
	SleepTotal     int64   `json:"sleepTotal"`
}

func douyinDownloadCacheKey(awemeID string) string {
	id := strings.TrimSpace(awemeID)
	if id == "" {
		return ""
	}
	return douyinDownloadedCachePrefix + id
}

func sanitizeDouyinFilename(name string) string {
	t := strings.TrimSpace(name)
	if t == "" {
		return "douyin"
	}
	t = douyinInvalidFilenameChars.ReplaceAllString(t, "_")
	t = strings.Trim(t, " .")
	if t == "" {
		return "douyin"
	}
	return t
}

func uniqueDouyinFilePath(path string) string {
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

func clampDouyinPercent(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func (d *Douyin) emitDownloadProgress(p douyinDownloadProgress) {
	ctx := d.context()
	if ctx == nil {
		return
	}
	if p.AwemeID != "" {
		p.Percent = clampDouyinPercent(p.Percent)
		d.progressMu.Lock()
		prev, ok := d.progressByID[p.AwemeID]
		// 并发下载时避免旧事件把同一任务的进度条回退；错误事件仍允许透传给前端。
		if ok && p.Percent < prev && p.Phase != "error" {
			d.progressMu.Unlock()
			return
		}
		d.progressByID[p.AwemeID] = p.Percent
		d.progressMu.Unlock()
	}
	wailsRuntime.EventsEmit(ctx, "douyin-download-progress", p)
}

// douyinWeightedPercent 用于图片合集：每张图片占一段权重，拼成整条 0-100 进度。
func douyinWeightedPercent(start, weight float64, downloaded, total int64) float64 {
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

// markDownloaded 写入下载成功历史；缓存失败不影响已经落盘的文件。
func (d *Douyin) markDownloaded(task DouyinDownloadTask, path string, isImageAlbum bool, imageCount int) {
	key := douyinDownloadCacheKey(task.AwemeID)
	if key == "" {
		return
	}

	payload, err := sonic.Marshal(DouyinDownloadHistoryItem{
		AwemeID:      strings.TrimSpace(task.AwemeID),
		Title:        strings.TrimSpace(task.Title),
		Cover:        strings.TrimSpace(task.Cover),
		Duration:     task.Duration,
		AuthorName:   strings.TrimSpace(task.AuthorName),
		PublishTime:  task.PublishTime,
		DiggCount:    task.DiggCount,
		CollectCount: task.CollectCount,
		SourceName:   strings.TrimSpace(task.SourceName),
		SourceKind:   strings.TrimSpace(task.SourceKind),
		Path:         path,
		IsImageAlbum: isImageAlbum,
		ImageCount:   imageCount,
		Downloaded:   time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		d.logger.Errorf("marshal douyin downloaded cache failed: %v", err)
		return
	}
	if err = d.settings.SetKey(key, string(payload)); err != nil {
		d.logger.Errorf("save douyin downloaded cache failed: %v", err)
	}
}

// DownloadHistory 返回抖音下载历史，用于前端历史页展示。
func (d *Douyin) DownloadHistory() ([]DouyinDownloadHistoryItem, error) {
	items := make([]DouyinDownloadHistoryItem, 0)
	prefix := []byte(douyinDownloadedCachePrefix)

	err := d.settings.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var history DouyinDownloadHistoryItem
				if err := sonic.Unmarshal(bytes.Clone(val), &history); err != nil {
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
		return parseDouyinDownloadHistoryTime(items[i].Downloaded).After(parseDouyinDownloadHistoryTime(items[j].Downloaded))
	})
	return items, nil
}

func parseDouyinDownloadHistoryTime(value string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}

// DeleteDownloadHistory 删除单条历史记录；不会删除已经下载到本地的文件。
func (d *Douyin) DeleteDownloadHistory(awemeID string) error {
	key := douyinDownloadCacheKey(awemeID)
	if key == "" {
		return errors.New("视频ID为空")
	}
	return d.settings.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})
}

// ClearDownloadHistory 清空抖音下载历史；不会删除已经下载到本地的文件。
func (d *Douyin) ClearDownloadHistory() error {
	prefix := []byte(douyinDownloadedCachePrefix)
	return d.settings.Update(func(txn *badger.Txn) error {
		keys := make([][]byte, 0)
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			keys = append(keys, it.Item().KeyCopy(nil))
		}
		for _, key := range keys {
			if err := txn.Delete(key); err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
		}
		return nil
	})
}

// downloadURLToFile 手动流式读取响应体，这样才能持续把字节进度推给前端。
func (d *Douyin) downloadURLToFile(rawURL, targetPath string, task DouyinDownloadTask, phase string, start, weight float64) error {
	rawURL = strings.TrimSpace(rawURL)
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return errors.New("下载地址无效")
	}

	headers, err := d.publicHeaders()
	if err != nil {
		return err
	}
	resp, err := d.client.R().
		DisableAutoReadResponse().
		SetHeaders(headers).
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

	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 256*1024)
	lastEmit := time.Time{}
	d.emitDownloadProgress(douyinDownloadProgress{
		AwemeID: task.AwemeID,
		Title:   task.Title,
		Phase:   phase,
		Total:   total,
		Percent: start,
	})

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err = file.Write(buf[:n]); err != nil {
				return err
			}
			downloaded += int64(n)
			now := time.Now()
			if now.Sub(lastEmit) >= 200*time.Millisecond || (total > 0 && downloaded >= total) {
				d.emitDownloadProgress(douyinDownloadProgress{
					AwemeID:    task.AwemeID,
					Title:      task.Title,
					Phase:      phase,
					Downloaded: downloaded,
					Total:      total,
					Percent:    douyinWeightedPercent(start, weight, downloaded, total),
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

	d.emitDownloadProgress(douyinDownloadProgress{
		AwemeID:    task.AwemeID,
		Title:      task.Title,
		Phase:      phase,
		Downloaded: downloaded,
		Total:      total,
		Percent:    start + weight,
	})
	return nil
}

func (d *Douyin) resolveDownloadDir(storagePath string, task DouyinDownloadTask) (string, error) {
	allowGroup, err := d.settings.GetSavePreference()
	if err != nil {
		return "", err
	}
	if !allowGroup {
		return storagePath, nil
	}
	sourceName := sanitizeDouyinFilename(task.SourceName)
	switch strings.TrimSpace(task.SourceKind) {
	case "收藏合集", "用户合集":
		if author := sanitizeDouyinFilename(task.AuthorName); author != "" && sourceName != "" {
			return filepath.Join(storagePath, author, sourceName), nil
		}
		if sourceName != "" {
			return filepath.Join(storagePath, sourceName), nil
		}
	case "收藏视频":
		return filepath.Join(storagePath, "收藏视频"), nil
	case "用户作品":
		if sourceName != "" {
			return filepath.Join(storagePath, sourceName), nil
		}
	}
	author := sanitizeDouyinFilename(task.AuthorName)
	if author == "" {
		return storagePath, nil
	}
	return filepath.Join(storagePath, author), nil
}

func douyinAssetExt(asset DouyinDownloadAsset) string {
	ext := strings.TrimSpace(asset.Ext)
	if ext == "" {
		if asset.Kind == "video" {
			return ".mp4"
		}
		return ".jpg"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".mp4":
		return ext
	default:
		if asset.Kind == "video" {
			return ".mp4"
		}
		return ".jpg"
	}
}

// downloadTask 同时处理普通视频和图片合集；当前版本不在后端重新解析 bit_rate。
func (d *Douyin) downloadTask(task DouyinDownloadTask) (string, error) {
	task.AwemeID = strings.TrimSpace(task.AwemeID)
	task.Title = strings.TrimSpace(task.Title)
	if task.AwemeID == "" {
		return "", errors.New("视频ID为空")
	}
	if task.Title == "" {
		task.Title = task.AwemeID
	}

	storagePath, err := d.settings.GetStorage()
	if err != nil {
		return "", err
	}
	targetDir, err := d.resolveDownloadDir(storagePath, task)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		return "", errors.New("创建下载目录失败")
	}

	d.progressMu.Lock()
	delete(d.progressByID, task.AwemeID)
	d.progressMu.Unlock()

	if len(task.Assets) > 0 || len(task.ImageURLs) > 0 {
		// 图文/动图保存为一个目录，素材按 001.jpg、002.mp4 顺序落盘，配乐单独保存为 music.mp3。
		dir := uniqueDouyinFilePath(filepath.Join(targetDir, sanitizeDouyinFilename(task.Title)))
		if err = os.MkdirAll(dir, 0o755); err != nil {
			return "", errors.New("创建素材目录失败")
		}
		assets := task.Assets
		if len(assets) == 0 {
			assets = make([]DouyinDownloadAsset, 0, len(task.ImageURLs))
			for _, imageURL := range task.ImageURLs {
				assets = append(assets, DouyinDownloadAsset{URL: imageURL, Kind: "image", Ext: ".jpg"})
			}
		}
		total := len(assets)
		if strings.TrimSpace(task.MusicURL) != "" {
			total++
		}
		if total == 0 {
			return "", errors.New("素材下载地址为空")
		}
		for index, asset := range assets {
			ext := douyinAssetExt(asset)
			start := float64(index) / float64(total) * 100
			weight := 100 / float64(total)
			path := filepath.Join(dir, fmt.Sprintf("%03d%s", index+1, ext))
			if err = d.downloadURLToFile(asset.URL, path, task, asset.Kind, start, weight); err != nil {
				d.emitDownloadProgress(douyinDownloadProgress{AwemeID: task.AwemeID, Title: task.Title, Phase: "error"})
				return "", err
			}
		}
		if strings.TrimSpace(task.MusicURL) != "" {
			start := float64(len(assets)) / float64(total) * 100
			weight := 100 / float64(total)
			path := filepath.Join(dir, "music.mp3")
			if err = d.downloadURLToFile(task.MusicURL, path, task, "music", start, weight); err != nil {
				d.emitDownloadProgress(douyinDownloadProgress{AwemeID: task.AwemeID, Title: task.Title, Phase: "error"})
				return "", err
			}
		}
		d.emitDownloadProgress(douyinDownloadProgress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
		d.markDownloaded(task, dir, true, len(assets))
		return dir, nil
	}

	if strings.TrimSpace(task.VideoURL) == "" {
		return "", errors.New("视频下载地址为空")
	}
	outPath := uniqueDouyinFilePath(filepath.Join(targetDir, sanitizeDouyinFilename(task.Title)+".mp4"))
	if err = d.downloadURLToFile(task.VideoURL, outPath, task, "video", 0, 100); err != nil {
		d.emitDownloadProgress(douyinDownloadProgress{AwemeID: task.AwemeID, Title: task.Title, Phase: "error"})
		return "", err
	}
	d.emitDownloadProgress(douyinDownloadProgress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
	d.markDownloaded(task, outPath, false, 0)
	return outPath, nil
}

func uniqueDouyinDownloadTasks(tasks []DouyinDownloadTask) []DouyinDownloadTask {
	seen := make(map[string]struct{}, len(tasks))
	unique := make([]DouyinDownloadTask, 0, len(tasks))
	for _, task := range tasks {
		key := strings.TrimSpace(task.AwemeID)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		unique = append(unique, task)
	}
	return unique
}

func (d *Douyin) sleepAfterTask(task DouyinDownloadTask) {
	// 和 B 站下载保持一致：每个 worker 完成一条任务后按用户设置随机休眠，降低连续请求频率。
	sleepTime, err := d.settings.GetSleepTime()
	if err != nil || sleepTime <= 0 {
		return
	}
	sleepTime = rand.Int63n(sleepTime)
	if sleepTime <= 0 {
		return
	}
	for remaining := sleepTime; remaining > 0; remaining-- {
		d.emitDownloadProgress(douyinDownloadProgress{
			AwemeID:        task.AwemeID,
			Title:          task.Title,
			Phase:          "sleep",
			Percent:        100,
			SleepRemaining: remaining,
			SleepTotal:     sleepTime,
		})
		time.Sleep(time.Second)
	}
	d.emitDownloadProgress(douyinDownloadProgress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
}

// DownloadVideos 批量下载抖音任务；单个任务失败不会中断整批，结果逐条返回给前端。
func (d *Douyin) DownloadVideos(tasks []DouyinDownloadTask) (DouyinDownloadBatchResult, error) {
	tasks = uniqueDouyinDownloadTasks(tasks)
	result := DouyinDownloadBatchResult{Results: make([]DouyinDownloadResult, 0, len(tasks))}
	if len(tasks) == 0 {
		return result, errors.New("下载列表为空")
	}

	workerCount, err := d.settings.GetConcurrencyNum()
	if err != nil {
		return result, err
	}
	if workerCount < 1 {
		workerCount = 1
	}
	if len(tasks) < workerCount {
		workerCount = len(tasks)
	}

	jobs := make(chan DouyinDownloadTask)
	results := make(chan DouyinDownloadResult, len(tasks))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Go(func() {
			for task := range jobs {
				path, err := d.downloadTask(task)
				item := DouyinDownloadResult{AwemeID: task.AwemeID, Title: task.Title, Path: path}
				if err != nil {
					item.Error = err.Error()
				}
				results <- item
				if err == nil {
					d.sleepAfterTask(task)
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
			result.Success++
		} else {
			result.Failed++
		}
		result.Results = append(result.Results, item)
	}
	return result, nil
}

func (d *Douyin) Download(videoURL string) error {
	result, err := d.DownloadVideos([]DouyinDownloadTask{{
		AwemeID:  fmt.Sprintf("manual-%d", time.Now().UnixNano()),
		Title:    "test",
		VideoURL: videoURL,
	}})
	if err != nil {
		return err
	}
	if result.Failed > 0 && len(result.Results) > 0 {
		return errors.New(result.Results[0].Error)
	}
	return nil
}
