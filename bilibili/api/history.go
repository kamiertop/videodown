package api

import (
	"bytes"
	"errors"
	"sort"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/dgraph-io/badger/v4"
	"github.com/kamiertop/videodown/bilibili/model"
)

// SearchDownloadHistory 获取历史搜索记录
// Deprecated: 该接口已废弃，搜索历史记录功能已迁移到前端实现，后端不再维护搜索历史记录
func (b *BiliBili) SearchDownloadHistory(upperNameOrTitle string) ([]model.DownloadHistoryItem, error) {
	var results []model.DownloadHistoryItem
	prefix := []byte(downloadedVideoCachePrefix)

	err := b.settings.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := it.Item().Value(func(val []byte) error {
				var history model.DownloadHistoryItem
				if err := sonic.Unmarshal(bytes.Clone(val), &history); err != nil {
					return err
				}
				kw := strings.ToLower(upperNameOrTitle)
				if strings.Contains(strings.ToLower(history.UpperName), kw) ||
					strings.Contains(strings.ToLower(history.Title), kw) {
					results = append(results, history)
				}
				return nil
			}); err != nil {
				b.logger.Errorf("read download history item failed: %v", err)
				continue
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return parseDownloadHistoryTime(results[i].Downloaded).After(parseDownloadHistoryTime(results[j].Downloaded))
	})

	return results, nil
}

// DownloadHistory 返回后端下载缓存记录；只读历史页使用，下载接口本身不暴露缓存命中细节。
func (b *BiliBili) DownloadHistory() ([]model.DownloadHistoryItem, error) {
	var items []model.DownloadHistoryItem
	prefix := []byte(downloadedVideoCachePrefix)

	err := b.settings.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var history model.DownloadHistoryItem
				if err := sonic.Unmarshal(bytes.Clone(val), &history); err != nil {
					return err
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

// ClearDownloadHistory 清空 B 站下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
func (b *BiliBili) ClearDownloadHistory() error {
	return b.settings.ClearDownloadHistory(downloadedVideoCachePrefix)
}

// DeleteDownloadHistory 删除单条下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
func (b *BiliBili) DeleteDownloadHistory(cid int64) error {
	key := downloadCacheKey(cid)
	if key == "" {
		return errors.New("视频CID为空")
	}

	return b.settings.DeleteKey(key)
}
