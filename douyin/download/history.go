package download

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"os"
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

// isDownloaded 判断 awemeID 是否已下载：缓存存在且文件（或图文目录）仍在磁盘上。
// 与 B 站 bilibili/download/util.go 的 isDownloaded 语义一致，文件被删后自动视为未下载。
func (d *Service) isDownloaded(awemeID string) bool {
	key := cacheKey(awemeID)
	if key == "" {
		return false
	}
	val, err := d.store.Get(key)
	if err != nil {
		return false
	}
	var history HistoryItem
	if err := json.Unmarshal([]byte(val), &history); err != nil {
		return false
	}
	if history.Path == "" {
		return false
	}
	if _, err := os.Stat(history.Path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// DownloadedAwemeIDs 返回给定 awemeID 中已下载（且文件仍在）的子集，
// 供前端批量会话做增量过滤。
func (d *Service) DownloadedAwemeIDs(awemeIDs []string) ([]string, error) {
	seen := make(map[string]struct{}, len(awemeIDs))
	downloaded := make([]string, 0, len(awemeIDs))
	for _, awemeID := range awemeIDs {
		id := strings.TrimSpace(awemeID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if d.isDownloaded(id) {
			downloaded = append(downloaded, id)
		}
	}
	return downloaded, nil
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
