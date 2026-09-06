package api

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"sort"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/bilibili/util"
	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/utils"

	"github.com/dgraph-io/badger/v4"
)

// DownloadHistory 返回后端下载缓存记录；只读历史页使用，下载接口本身不暴露缓存命中细节。
func (b *BiliBili) DownloadHistory() ([]model.DownloadHistoryItem, error) {
	var items []model.DownloadHistoryItem
	prefix := []byte(util.CachePrefix)

	err := b.store.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var history model.DownloadHistoryItem
				if err := json.Unmarshal(bytes.Clone(val), &history); err != nil {
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
		return utils.ParseDownloadHistoryTime(items[i].Downloaded).After(utils.ParseDownloadHistoryTime(items[j].Downloaded))
	})

	return items, nil
}

// ClearDownloadHistory 清空 B 站下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
func (b *BiliBili) ClearDownloadHistory() error {
	return b.store.DeletePrefix(util.CachePrefix)
}

// DeleteDownloadHistory 删除单条下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
func (b *BiliBili) DeleteDownloadHistory(cid int64) error {
	key := util.DownloadCacheKey(cid)
	if key == "" {
		return errors.New("视频CID为空")
	}

	return b.store.Delete(key)
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
