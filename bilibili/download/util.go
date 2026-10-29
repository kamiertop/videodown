package download

import (
	"encoding/json/v2"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/bilibili/util"
	"github.com/kamiertop/videodown/utils"
)

const (
	kindVideo = "video"
	kindCover = "cover"
	kindAudio = util.DownloadKindAudio
)

func normalizeHTTPURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}
	return rawURL
}

func sanitizeFilename(name string) string {
	t := utils.FileName(name)
	if t == "" {
		return "video"
	}
	return t
}

// downloadedCachePath 返回已下载缓存中的文件路径；缓存只在后端使用，不增加前端协议字段。
func (s *Service) downloadedCachePath(cid int64) (string, bool) {
	key := util.DownloadCacheKey(cid)
	if key == "" {
		return "", false
	}

	raw, err := s.store.History(key)
	if err != nil {
		return "", false
	}

	var cached model.DownloadHistoryItem
	if err = json.Unmarshal([]byte(raw), &cached); err != nil {
		return "", false
	}
	return cached.Path, true
}

// normalizeBvid 统一 BV 号大小写与首尾空白；历史记录和查询键都先归一再比较。
func normalizeBvid(bvid string) string {
	return strings.ToUpper(strings.TrimSpace(bvid))
}

// pathExists 路径为空或文件已删除视为不存在，其他 stat 错误不拦（与 isDownloaded 语义一致）。
func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// isDownloaded 检查缓存记录且确认文件仍存在于磁盘；缓存可能因手动删文件而过期。
func (s *Service) isDownloaded(cid int64) (string, bool) {
	path, ok := s.downloadedCachePath(cid)
	if !ok {
		return "", false
	}
	return path, pathExists(path)
}

// DownloadedVideoKeys 返回 keys 中已下载（缓存存在且文件仍在）元素的下标，
// 供前端批量会话在解析播放地址前做增量过滤，省掉对整页已下载视频的解析请求。
// cid > 0 时精确匹配该分 P；未提供 cid 时该 bvid 任意分 P 已下载即视为已下载。
func (s *Service) DownloadedVideoKeys(keys []DownloadKey) ([]int, error) {
	// 全部带 cid 时直接按缓存 key 点查，不必扫描下载历史（cid 全局唯一，命中即该视频）。
	allHaveCid := true
	for _, key := range keys {
		if key.Cid <= 0 {
			allHaveCid = false
			break
		}
	}
	matched := make([]int, 0)
	if allHaveCid {
		for i, key := range keys {
			if _, ok := s.isDownloaded(key.Cid); ok {
				matched = append(matched, i)
			}
		}
		return matched, nil
	}

	// 只收集查询涉及的 bvid，不为整个下载历史建立索引。
	wanted := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if bv := normalizeBvid(key.Bvid); bv != "" {
			wanted[bv] = struct{}{}
		}
	}
	if len(wanted) == 0 {
		return matched, nil
	}

	type downloadedRecord struct {
		cid  int64
		path string
	}
	candidates := make(map[string][]downloadedRecord, len(wanted))
	rows, err := s.store.BilibiliVideoRows()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		var history model.DownloadHistoryItem
		if json.Unmarshal([]byte(row), &history) != nil {
			continue
		}
		bv := normalizeBvid(history.Bvid)
		if _, ok := wanted[bv]; ok {
			candidates[bv] = append(candidates[bv], downloadedRecord{cid: history.Cid, path: history.Path})
		}
	}

	for i, key := range keys {
		records, ok := candidates[normalizeBvid(key.Bvid)]
		if !ok {
			continue
		}
		for _, record := range records {
			if key.Cid > 0 && record.cid != key.Cid {
				continue
			}
			if pathExists(record.path) {
				matched = append(matched, i)
				break
			}
		}
	}
	return matched, nil
}

// markDownloaded 写入下载成功缓存；写缓存失败不影响已经完成的文件保存。
// 封面以负数 cid 存进历史（历史页删除接口按同一编码反查）；音频用独立前缀，
// 避免音频记录被视频已下载判断命中。
func (s *Service) markDownloaded(task Task, path string, downloadKind string) {
	if downloadKind == "" {
		downloadKind = kindVideo
	}

	cid := task.Cid
	key := util.DownloadCacheKey(cid)
	switch downloadKind {
	case kindCover:
		if cid > 0 {
			cid = -cid
			key = util.DownloadCacheKey(cid)
		}
	case kindAudio:
		key = util.AudioDownloadCacheKey(cid)
	}
	if key == "" {
		return
	}

	payload, err := json.Marshal(model.DownloadHistoryItem{
		Bvid:         strings.TrimSpace(task.Bvid),
		Cid:          cid,
		Title:        strings.TrimSpace(task.Title),
		Cover:        strings.TrimSpace(task.Cover),
		Duration:     task.Duration,
		UpperName:    strings.TrimSpace(task.UpperName),
		Play:         task.Play,
		Danmaku:      task.Danmaku,
		Pubtime:      task.Pubtime,
		SourceName:   strings.TrimSpace(task.SourceName),
		Path:         path,
		DownloadKind: downloadKind,
		Downloaded:   time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		s.logger.Errorf("marshal downloaded cache failed: %v", err)
		return
	}
	if err = s.store.SaveHistory(key, string(payload)); err != nil {
		s.logger.Errorf("save downloaded cache failed: %v", err)
	}
}

func progressKey(bvid string, cid int64) string {
	bv := strings.ToUpper(strings.TrimSpace(bvid))
	if bv == "" {
		return ""
	}
	if cid > 0 {
		return bv + ":" + strconv.FormatInt(cid, 10)
	}

	return bv
}
