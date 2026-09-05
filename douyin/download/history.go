package download

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/kamiertop/videodown/utils"
)

type HistoryItem struct {
	AwemeID      string `json:"awemeId"`
	Title        string `json:"title"`
	Cover        string `json:"cover"`
	Duration     int    `json:"duration"`
	AuthorName   string `json:"authorName"`
	PublishTime  int    `json:"publishTime"`
	DiggCount    int    `json:"diggCount"`
	CollectCount int    `json:"collectCount"`
	SourceName   string `json:"sourceName"`
	Path         string `json:"path"`
	IsImageAlbum bool   `json:"isImageAlbum"`
	ImageCount   int    `json:"imageCount"`
	DownloadKind string `json:"downloadKind"`
	// Wails 绑定生成不支持直接暴露 time.Time，保存为 RFC3339 字符串给前端解析。
	Downloaded string `json:"downloaded"`
}

// DeleteDownloadHistory 删除单条历史记录；不会删除已经下载到本地的文件。
func (d *Service) DeleteDownloadHistory(awemeID string) error {
	key := cacheKey(awemeID)
	if key == "" {
		return errors.New("视频ID为空")
	}

	return d.store.Delete(key)
}

// ClearDownloadHistory 清空抖音下载历史；不会删除已经下载到本地的文件。
func (d *Service) ClearDownloadHistory() error {
	return d.store.DeletePrefix(cachePrefix)
}

// DownloadHistory 返回抖音下载历史，用于前端历史页展示。
func (d *Service) DownloadHistory() ([]HistoryItem, error) {
	items := make([]HistoryItem, 0)
	prefix := []byte(cachePrefix)

	err := d.store.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var history HistoryItem
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
		return utils.ParseDownloadHistoryTime(items[i].Downloaded).After(utils.ParseDownloadHistoryTime(items[j].Downloaded))
	})

	return items, nil
}

// markDownloaded 写入下载成功历史；缓存失败不影响已经落盘的文件。
func (d *Service) markDownloaded(task Task, path string, isImageAlbum bool, imageCount int, downloadKind string) {
	key := cacheKey(task.AwemeID)
	if key == "" {
		return
	}
	if downloadKind == "" {
		if isImageAlbum {
			downloadKind = kindAlbum
		} else {
			downloadKind = kindVideo
		}
	}

	payload, err := json.Marshal(HistoryItem{
		AwemeID:      strings.TrimSpace(task.AwemeID),
		Title:        strings.TrimSpace(task.Title),
		Cover:        strings.TrimSpace(task.Cover),
		Duration:     task.Duration,
		AuthorName:   strings.TrimSpace(task.AuthorName),
		PublishTime:  task.PublishTime,
		DiggCount:    task.DiggCount,
		CollectCount: task.CollectCount,
		SourceName:   strings.TrimSpace(task.SourceName),
		Path:         path,
		IsImageAlbum: isImageAlbum,
		ImageCount:   imageCount,
		DownloadKind: downloadKind,
		Downloaded:   time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		d.logger.Errorf("marshal douyin downloaded cache failed: %v", err)
		return
	}
	if err = d.store.Set(key, string(payload)); err != nil {
		d.logger.Errorf("save douyin downloaded cache failed: %v", err)
	}
}
