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
	"time"

	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Service struct {
	client        *req.Client
	logger        *logger.Logger
	store         *storage.Store
	events        *application.EventManager
	mu            sync.Mutex
	progress      map[string]float64
	publicHeaders func() (map[string]string, error)
}

func New(logger *logger.Logger, store *storage.Store, events *application.EventManager, publicHeaders func() (map[string]string, error)) *Service {
	return &Service{
		logger:        logger,
		store:         store,
		events:        events,
		mu:            sync.Mutex{},
		publicHeaders: publicHeaders,
		progress:      make(map[string]float64),
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

// DownloadVideos 批量下载抖音任务；单个任务失败不会中断整批，结果逐条返回给前端。
func (d *Service) DownloadVideos(tasks []Task) (BatchResult, error) {
	tasks = uniqueDownloadTasks(tasks)
	result := BatchResult{Results: make([]Result, 0, len(tasks))}
	if len(tasks) == 0 {
		return result, errors.New("下载列表为空")
	}

	workerCount, err := d.store.ConcurrencyNum()
	if err != nil {
		return result, err
	}

	jobs := make(chan Task, len(tasks))
	results := make(chan Result, len(tasks))
	var wg sync.WaitGroup
	for range workerCount {
		wg.Go(func() {
			for task := range jobs {
				targetPath, err := d.downloadTask(task)
				item := Result{AwemeID: task.AwemeID, Title: task.Title, Path: targetPath}
				if err != nil {
					d.logger.Errorf("download task failed, task: %v, err: %v", task, err)
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

func (d *Service) Download(videoURL string) error {
	result, err := d.DownloadVideos([]Task{{
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

func (d *Service) emitProgress(p progress) {
	if p.AwemeID != "" {
		p.Percent = utils.ClampPercent(p.Percent)
		d.mu.Lock()
		prev, ok := d.progress[p.AwemeID]
		// 并发下载时避免旧事件把同一任务的进度条回退；错误事件仍允许透传给前端。
		if ok && p.Percent < prev && p.Phase != "error" {
			d.mu.Unlock()
			return
		}
		d.progress[p.AwemeID] = p.Percent
		d.mu.Unlock()
	}
	if d.events != nil {
		d.events.Emit("douyin-download-progress", p)
	}
}

// downloadURLToFile 手动流式读取响应体，这样才能持续把字节进度推给前端。
func (d *Service) downloadURLToFile(rawURL, targetPath string, task Task, phase string, start, weight float64) error {
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
	d.emitProgress(progress{
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
				d.emitProgress(progress{
					AwemeID:    task.AwemeID,
					Title:      task.Title,
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

	d.emitProgress(progress{
		AwemeID:    task.AwemeID,
		Title:      task.Title,
		Phase:      phase,
		Downloaded: downloaded,
		Total:      total,
		Percent:    start + weight,
	})

	return nil
}

// downloadTask 同时处理普通视频和图片合集；当前版本不在后端重新解析 bit_rate。
func (d *Service) downloadTask(task Task) (string, error) {
	task.AwemeID = strings.TrimSpace(task.AwemeID)
	task.Title = strings.TrimSpace(task.Title)
	if task.AwemeID == "" {
		return "", errors.New("视频ID为空")
	}
	if task.Title == "" {
		task.Title = task.AwemeID
	}

	storagePath, err := d.store.StoragePath()
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

	d.mu.Lock()
	delete(d.progress, task.AwemeID)
	d.mu.Unlock()

	if len(task.Assets) > 0 || len(task.ImageURLs) > 0 {
		// 图文/动图保存为一个目录，素材按 001.<原始扩展名> 顺序落盘，配乐单独保存为 music.mp3。
		dirName := utils.FileName(task.Title)
		if dirName == "" {
			dirName = "douyin"
		}
		dir := utils.UniqueFilePath(filepath.Join(targetDir, dirName))
		if err = os.MkdirAll(dir, 0o755); err != nil {
			return "", errors.New("创建素材目录失败")
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
			return "", errors.New("素材下载地址为空")
		}
		for index, asset := range assets {
			ext := douyinAssetExt(asset)
			start := float64(index) / float64(total) * 100
			weight := 100 / float64(total)
			join := filepath.Join(dir, fmt.Sprintf("%03d%s", index+1, ext))
			if err = d.downloadURLToFile(asset.URL, join, task, asset.Kind, start, weight); err != nil {
				d.emitProgress(progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "error"})
				return "", err
			}
		}
		if strings.TrimSpace(task.MusicURL) != "" {
			start := float64(len(assets)) / float64(total) * 100
			weight := 100 / float64(total)
			join := filepath.Join(dir, "music.mp3")
			if err = d.downloadURLToFile(task.MusicURL, join, task, "music", start, weight); err != nil {
				d.emitProgress(progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "error"})
				return "", err
			}
		}
		d.emitProgress(progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
		d.markDownloaded(task, dir, true, len(assets), kindAlbum)
		return dir, nil
	}

	if strings.TrimSpace(task.VideoURL) == "" {
		return "", errors.New("视频下载地址为空")
	}
	template, _ := d.store.FilenameTemplate()
	now := time.Now()
	values := map[string]string{"title": task.Title, "id": task.AwemeID, "author": task.AuthorName, "author_id": "", "source": task.SourceName, "folder": "", "collection": task.SourceName, "publish_date": "", "date": now.Format("2006-01-02"), "time": now.Format("15-04-05")}
	if task.PublishTime > 0 {
		values["publish_date"] = time.Unix(int64(task.PublishTime), 0).Format("2006-01-02")
	}
	fileName := utils.ApplyFilenameTemplate(template, values)
	if fileName == "" {
		fileName = "douyin"
	}
	outPath := utils.UniqueFilePath(filepath.Join(targetDir, fileName+".mp4"))
	if err = d.downloadURLToFile(task.VideoURL, outPath, task, "video", 0, 100); err != nil {
		d.emitProgress(progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "error"})
		return "", err
	}
	d.emitProgress(progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
	d.markDownloaded(task, outPath, false, 0, kindVideo)

	return outPath, nil
}

func (d *Service) sleepAfterTask(task Task) {
	// 和 B 站下载保持一致：每个 worker 完成一条任务后按用户设置随机休眠，降低连续请求频率。
	sleepTime, err := d.store.SleepTime()
	if err != nil || sleepTime <= 0 {
		return
	}
	sleepTime = rand.Int64N(sleepTime)
	if sleepTime <= 0 {
		return
	}
	for remaining := sleepTime; remaining > 0; remaining-- {
		d.emitProgress(progress{
			AwemeID:        task.AwemeID,
			Title:          task.Title,
			Phase:          "sleep",
			Percent:        100,
			SleepRemaining: remaining,
			SleepTotal:     sleepTime,
		})
		// TODO
	}
	d.emitProgress(progress{AwemeID: task.AwemeID, Title: task.Title, Phase: "done", Percent: 100})
}
