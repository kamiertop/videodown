package api

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/bilibili/util"
	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/utils"

	"github.com/dgraph-io/badger/v4"
)

// HistoryPage 下载历史的一页：keyword 为空时 Total 是全部记录数，非空时是
// 匹配记录数；HasMore 表示过滤后仍有下一页。
type HistoryPage struct {
	Items   []model.DownloadHistoryItem `json:"items"`
	Total   int                         `json:"total"`
	HasMore bool                        `json:"hasMore"`
}

const (
	historyPageSizeDefault = 50
	historyPageSizeMax     = 200
	// 写入方在 download 包，无法直接通知本包失效；缓存最多保鲜这么久，
	// 下载完成后稍等片刻进历史页即可看到新记录。
	historyFreshness = 3 * time.Second
)

// DownloadHistoryPage 分页返回 B 站下载历史：keyword 模糊匹配标题/UP 主名
// （小写包含），offset 是过滤后列表的偏移。首次调用全量扫描并缓存排序结果，
// 翻页与搜索只走内存，避免每次进历史页都全量扫描、全量跨桥、全量渲染。
func (b *BiliBili) DownloadHistoryPage(offset int, limit int, keyword string) (HistoryPage, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = historyPageSizeDefault
	}
	if limit > historyPageSizeMax {
		limit = historyPageSizeMax
	}

	snapshot, err := b.historySnapshot()
	if err != nil {
		return HistoryPage{}, err
	}

	filtered := snapshot
	if kw := strings.ToLower(strings.TrimSpace(keyword)); kw != "" {
		matched := make([]model.DownloadHistoryItem, 0, len(snapshot))
		for _, item := range snapshot {
			if strings.Contains(strings.ToLower(item.Title), kw) ||
				strings.Contains(strings.ToLower(item.UpperName), kw) {
				matched = append(matched, item)
			}
		}
		filtered = matched
	}

	page := HistoryPage{Items: []model.DownloadHistoryItem{}, Total: len(filtered)}
	if offset < len(filtered) {
		end := min(offset+limit, len(filtered))
		page.Items = filtered[offset:end]
		page.HasMore = end < len(filtered)
	}
	return page, nil
}

// historySnapshot 返回按下载时间倒序的历史快照；缓存失效或过期时重建一次。
func (b *BiliBili) historySnapshot() ([]model.DownloadHistoryItem, error) {
	b.historyMu.RLock()
	if b.historyValid && time.Since(b.historyStamp) < historyFreshness {
		items := b.historyItems
		b.historyMu.RUnlock()
		return items, nil
	}
	b.historyMu.RUnlock()

	items, err := b.scanDownloadHistory()
	if err != nil {
		return nil, err
	}

	b.historyMu.Lock()
	b.historyItems, b.historyValid, b.historyStamp = items, true, time.Now()
	b.historyMu.Unlock()
	return items, nil
}

// invalidateHistory 历史记录变化后调用，下次查询时重建快照。
func (b *BiliBili) invalidateHistory() {
	b.historyMu.Lock()
	b.historyItems, b.historyValid = nil, false
	b.historyMu.Unlock()
}

// scanDownloadHistory 全量扫描下载缓存并按下载时间倒序排序。
func (b *BiliBili) scanDownloadHistory() ([]model.DownloadHistoryItem, error) {
	var items []model.DownloadHistoryItem
	prefix := []byte(util.CachePrefix)

	err := b.store.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: prefix, PrefetchValues: false})
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var history model.DownloadHistoryItem
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

// ClearDownloadHistory 清空 B 站下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
func (b *BiliBili) ClearDownloadHistory() error {
	if err := b.store.DeletePrefix(util.CachePrefix); err != nil {
		return err
	}
	b.invalidateHistory()
	return nil
}

// DeleteDownloadHistory 删除单条下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
func (b *BiliBili) DeleteDownloadHistory(cid int64) error {
	key := util.DownloadCacheKey(cid)
	if key == "" {
		return errors.New("视频CID为空")
	}
	if err := b.store.Delete(key); err != nil {
		return err
	}
	b.invalidateHistory()
	return nil
}

// PlayHistory 返回播放历史记录
func (b *BiliBili) PlayHistory(cursor int, viewAt int) (model.PlayHistoryData, error) {
	var resp struct {
		model.ApiResponse
		Data model.PlayHistoryData `json:"data"`
	}

	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}

	err = b.client.
		Get("https://api.bilibili.com/x/web-interface/history/cursor").
		SetQueryParamsAnyType(map[string]any{
			"cursor":      cursor, //  初始为0，后续使用返回的data.cursor.max
			"view_at":     viewAt, // 初始为0，后续使用返回的data.cursor.view_at
			"business":    "",
			"search_type": "archive",
			"ps":          20,
			webLocation:   "333.1387",
		}).
		SetHeaders(publicHeaders()).
		SetHeader(constant.Origin, biliBiliUrl).
		SetHeader(constant.Referer, biliBiliUrl).
		SetHeader(constant.Cookie, cookies).
		Do().
		Into(&resp)
	if err != nil {
		b.logger.Errorf("request play history api error: %v", err)
		return resp.Data, err
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("request play history error, code: %d, message: %s", resp.Code, resp.Message)
	}
	b.logger.Infof("cursor: %v", resp.Data.Cursor)

	return resp.Data, nil
}
