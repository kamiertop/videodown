package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/bilibili/util"
	"github.com/kamiertop/videodown/internal/constant"
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
)

// DownloadHistoryPage returns a filtered SQLite page, newest first, without requiring login.
func (b *BiliBili) DownloadHistoryPage(offset int, limit int, keyword string) (HistoryPage, error) {
	offset = max(0, offset)
	if limit <= 0 {
		limit = historyPageSizeDefault
	}
	limit = min(limit, historyPageSizeMax)
	rows, total, err := b.store.QueryHistory("bilibili", offset, limit, keyword)
	if err != nil {
		return HistoryPage{}, err
	}
	page := HistoryPage{Items: []model.DownloadHistoryItem{}, Total: total, HasMore: offset+len(rows) < total}
	for _, row := range rows {
		var item model.DownloadHistoryItem
		if err := json.Unmarshal([]byte(row), &item); err != nil {
			return HistoryPage{}, fmt.Errorf("读取下载记录失败: %w", err)
		}
		page.Items = append(page.Items, item)
	}
	return page, nil
}

// ClearDownloadHistory 清空 B 站下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
func (b *BiliBili) ClearDownloadHistory() error {
	if err := b.store.DeleteHistoryPrefix(util.CachePrefix); err != nil {
		return err
	}
	return nil
}

// DeleteDownloadHistory 删除单条下载历史；只清理缓存记录，不删除已经保存到本地的视频文件。
// 音频记录使用独立的 audio: 缓存键，需按下载类型反查；封面沿用负数 cid 编码。
func (b *BiliBili) DeleteDownloadHistory(cid int64, downloadKind string) error {
	key := util.DownloadCacheKey(cid)
	if downloadKind == util.DownloadKindAudio {
		key = util.AudioDownloadCacheKey(cid)
	}
	if key == "" {
		return errors.New("视频CID为空")
	}
	if err := b.store.DeleteHistory(key); err != nil {
		return err
	}
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
		return resp.Data, fmt.Errorf("获取播放历史失败: %s", resp.Message)
	}
	b.logger.Infof("cursor: %v", resp.Data.Cursor)

	return resp.Data, nil
}
