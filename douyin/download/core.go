package download

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/internal/downloadprogress"
	"github.com/kamiertop/videodown/internal/mediafile"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
)

type Service struct {
	client *req.Client
	logger *logger.Logger
	store  *storage.Store

	progress         downloadprogress.Tracker
	progressListener atomic.Pointer[func(Progress)]
	// active 记录正在执行（含休眠间隔）的任务数，供关闭程序前判断是否有任务未完成。
	active        atomic.Int64
	publicHeaders func() (map[string]string, error)
	// 下载历史的内存缓存：按下载时间倒序，写入/删除/清空时失效，
	// 历史页翻页只切内存，不必每次查询 SQLite。
	historyMu    sync.RWMutex
	historyItems []HistoryItem
	historyValid bool
}

// ActiveCount 返回正在执行的任务数；0 表示当前没有任务在进行。
func (d *Service) ActiveCount() int {
	return int(d.active.Load())
}

func (d *Service) DownloadConcurrency() int { return d.store.Config().Download.Concurrency }

func (d *Service) SetProgressListener(fn func(Progress)) {
	d.progressListener.Store(&fn)
}

func New(logger *logger.Logger, store *storage.Store, publicHeaders func() (map[string]string, error)) *Service {
	return &Service{
		logger: logger,
		store:  store,

		publicHeaders: publicHeaders,
		client: req.C().
			EnableAutoDecompress().
			SetTimeout(0).
			SetCommonRetryCount(3).
			SetJsonMarshal(func(v any) ([]byte, error) {
				return json.Marshal(v)
			}).
			SetJsonUnmarshal(func(data []byte, v any) error {
				return json.Unmarshal(data, v)
			}),
	}
}

// DownloadTask runs one task, including the configured delay after a successful transfer.
func (d *Service) DownloadTask(task Task) (string, bool, error) {
	d.active.Add(1)
	defer d.active.Add(-1)

	path, skipped, err := d.downloadTask(task)
	if err == nil && !skipped {
		d.sleepAfterTask(task)
	}
	return path, skipped, err
}

// DownloadVideos 批量下载抖音任务；单个任务失败不会中断整批，逐条返回任务结果。
func (d *Service) DownloadVideos(tasks []Task) (BatchResult, error) {
	tasks = uniqueDownloadTasks(tasks)
	result := BatchResult{Results: make([]Result, 0, len(tasks))}
	if len(tasks) == 0 {
		return result, errors.New("下载列表为空")
	}

	workerCount := max(1, min(d.store.Config().Download.Concurrency, len(tasks)))

	jobs := make(chan Task, len(tasks))
	results := make(chan Result, len(tasks))
	var wg sync.WaitGroup
	for range workerCount {
		wg.Go(func() {
			for task := range jobs {
				targetPath, _, err := d.DownloadTask(task)
				item := Result{AwemeID: task.AwemeID, Title: task.Title, Path: targetPath}
				if err != nil {
					d.logger.Errorf("download failed for aweme %s: %v", task.AwemeID, err)
					item.Error = err.Error()
				}
				results <- item
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

func (d *Service) Download(videoURL string) error {
	_, _, err := d.DownloadTask(Task{
		AwemeID:  fmt.Sprintf("manual-%d", time.Now().UnixNano()),
		Title:    "test",
		VideoURL: videoURL,
	})
	return err
}

func (d *Service) emitProgress(p Progress) {
	if p.AwemeID != "" {
		p.Percent = downloadprogress.ClampPercent(p.Percent)
		if p.Phase == "error" {
			d.progress.Store(p.AwemeID, p.Percent)
		} else if _, ok := d.progress.Advance(p.AwemeID, p.Percent); !ok {
			return
		}
	}
	if listener := d.progressListener.Load(); listener != nil && *listener != nil {
		(*listener)(p)
	}
}

// downloadURLToFile 手动流式读取响应体，这样才能持续把字节进度推给界面。
func (d *Service) downloadURLToFile(rawURL, targetPath string, task Task, phase string, start, weight float64) (err error) {
	headers, err := d.publicHeaders()
	if err != nil {
		return err
	}
	resp, err := d.client.
		R().
		DisableAutoReadResponse().
		SetHeaders(headers).
		Get(rawURL)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	total := resp.ContentLength
	var downloaded int64
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("关闭下载文件: %w", closeErr))
		}
		if err == nil {
			d.emitProgress(Progress{
				AwemeID:    task.AwemeID,
				Title:      task.Title,
				Phase:      phase,
				Downloaded: downloaded,
				Total:      total,
				Percent:    start + weight,
			})
		}
	}()

	buf := make([]byte, 1024*1024)
	lastEmit := time.Time{}
	d.emitProgress(Progress{
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
				d.emitProgress(Progress{
					AwemeID:    task.AwemeID,
					Title:      task.Title,
					Phase:      phase,
					Downloaded: downloaded,
					Total:      total,
					Percent:    downloadprogress.WeightedPercent(start, weight, downloaded, total),
				})
				lastEmit = now
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				if total >= 0 && downloaded < total {
					return io.ErrUnexpectedEOF
				}
				break
			}
			return readErr
		}
	}

	return nil
}

// downloadTask 同时处理普通视频和图片合集；当前版本不在下载服务重新解析 bit_rate。
// 第二个返回值表示任务因已下载而被跳过，跳过的任务没有发起网络请求。
func (d *Service) downloadTask(task Task) (string, bool, error) {
	task.AwemeID = strings.TrimSpace(task.AwemeID)
	task.Title = cleanDouyinTitle(task.Title)
	if task.AwemeID == "" {
		return "", false, errors.New("视频ID为空")
	}
	if task.Title == "" {
		task.Title = task.AwemeID
	}

	// 与 B 站一致：增量下载开启时跳过已下载的作品；关闭后即使缓存命中也重新下载（设置读取失败按开启兜底）。
	skipDownloaded := d.store.Config().Download.SkipDownloaded
	if skipDownloaded {
		// 下载前先检查是否已下载，避免重复下载同一作品；批量会话在界面已过滤，这里是手动路径和并发窗口的兜底。
		path, err := d.downloadedPath(task.AwemeID)
		if err != nil {
			return "", false, err
		}
		if path != "" {
			d.emitProgress(Progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
			return path, true, nil
		}
	}

	storagePath := d.store.DownloadDirectory()
	targetDir := d.resolveDownloadDir(storagePath, task)
	isAlbum := len(task.Assets) > 0 || len(task.ImageURLs) > 0
	var targetPath string
	if !isAlbum {
		targetPath = filepath.Join(targetDir, d.videoFileName(task))
		if skipDownloaded && mediafile.ExistingVideoFile(targetPath) {
			d.emitProgress(Progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
			return targetPath, true, nil
		}
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", false, fmt.Errorf("创建下载目录失败: %w", err)
	}

	d.progress.Reset(task.AwemeID)
	p := Progress{
		AwemeID: task.AwemeID,
		Title:   task.Title,
	}
	if isAlbum {
		// 图文/动图保存为一个目录，素材按 001.<原始扩展名> 顺序落盘，配乐单独保存为 music.mp3。
		dirName := mediafile.FileName(task.Title)
		if dirName == "" {
			dirName = "douyin"
		}
		dirName = safeFileName(dirName, task.AwemeID)
		dir := mediafile.UniqueFilePath(filepath.Join(targetDir, dirName))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", false, fmt.Errorf("创建素材目录失败: %w", err)
		}
		assets := task.Assets
		if len(assets) == 0 {
			assets = make([]Asset, 0, len(task.ImageURLs))
			for _, imageURL := range task.ImageURLs {
				assets = append(assets, Asset{URL: imageURL, Kind: "image"})
			}
		}
		total := len(assets)
		if strings.TrimSpace(task.MusicURL) != "" {
			total++
		}
		if total == 0 {
			return "", false, errors.New("素材下载地址为空")
		}

		for index, asset := range assets {
			ext := douyinAssetExt(asset)
			start := float64(index) / float64(total) * 100
			weight := 100 / float64(total)
			join := filepath.Join(dir, fmt.Sprintf("%03d%s", index+1, ext))
			if err := d.downloadURLToFile(asset.URL, join, task, asset.Kind, start, weight); err != nil {
				d.emitProgress(p.withPhase("error"))
				return "", false, err
			}
		}
		if strings.TrimSpace(task.MusicURL) != "" {
			start := float64(len(assets)) / float64(total) * 100
			weight := 100 / float64(total)
			join := filepath.Join(dir, "music.mp3")
			if err := d.downloadURLToFile(task.MusicURL, join, task, "music", start, weight); err != nil {
				d.emitProgress(p.withPhase("error"))
				return "", false, err
			}
		}
		d.emitProgress(p.withPhase("done").withPercent(100))
		d.markDownloaded(task, dir, true, len(assets), kindAlbum)
		return dir, false, nil
	}

	if strings.TrimSpace(task.VideoURL) == "" {
		return "", false, errors.New("视频下载地址为空")
	}
	tmpDir, err := os.MkdirTemp(targetDir, ".douyin-")
	if err != nil {
		return "", false, err
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			d.logger.Warningf("remove download temporary directory: %v", err)
		}
	}()
	tmpPath := filepath.Join(tmpDir, "video.mp4")

	if err := d.downloadURLToFile(task.VideoURL, tmpPath, task, "video", 0, 100); err != nil {
		d.emitProgress(p.withPhase("error"))
		return "", false, err
	}
	outPath, err := mediafile.PublishVideoFile(tmpPath, targetPath)
	if err != nil {
		d.emitProgress(p.withPhase("error"))
		return "", false, err
	}
	d.emitProgress(p.withPhase("done").withPercent(100))
	d.markDownloaded(task, outPath, false, 0, kindVideo)

	return outPath, false, nil
}

func (d *Service) videoFileName(task Task) string {
	template := d.store.Config().Download.FilenameTemplate
	now := time.Now()
	values := map[string]string{"title": task.Title,
		"id": task.AwemeID, "author": task.AuthorName,
		"author_id": "", "source": task.SourceName, "folder": "",
		"collection": task.SourceName, "publish_date": "",
		"date": now.Format(time.DateOnly), "time": now.Format(time.TimeOnly)}
	if task.PublishTime > 0 {
		values["publish_date"] = time.Unix(int64(task.PublishTime), 0).Format(time.DateOnly)
	}
	fileName := mediafile.ApplyFilenameTemplate(template, values)
	if fileName == "" {
		fileName = "douyin"
	}
	fileName = safeFileName(fileName, task.AwemeID)
	return fileName + ".mp4"
}

func (d *Service) sleepAfterTask(task Task) {
	// 和 B 站下载保持一致：每个 worker 完成一条任务后按用户设置随机休眠，降低连续请求频率。
	sleepTime := d.store.Config().Download.SleepSeconds
	if sleepTime <= 0 {
		return
	}
	sleepTime = rand.Int64N(sleepTime)
	if sleepTime <= 0 {
		return
	}
	for remaining := sleepTime; remaining > 0; remaining-- {
		d.emitProgress(Progress{
			AwemeID:        task.AwemeID,
			Title:          task.Title,
			Phase:          "sleep",
			Percent:        100,
			SleepRemaining: remaining,
			SleepTotal:     sleepTime,
		})
		time.Sleep(time.Second)
	}
	d.emitProgress(Progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
}
