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

	raw, err := s.store.Get(key)
	if err != nil {
		return "", false
	}

	var cached model.DownloadHistoryItem
	if err = json.Unmarshal([]byte(raw), &cached); err != nil {
		return "", false
	}
	return cached.Path, true
}

// isDownloaded 检查缓存记录且确认文件仍存在于磁盘；缓存可能因手动删文件而过期。
func (s *Service) isDownloaded(cid int64) (string, bool) {
	path, ok := s.downloadedCachePath(cid)
	if !ok {
		return "", false
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", false
	}
	return path, true
}

// markDownloaded 写入下载成功缓存；写缓存失败不影响已经完成的文件保存。
func (s *Service) markDownloaded(task Task, path string, downloadKind string) {
	key := util.DownloadCacheKey(task.Cid)
	if key == "" {
		return
	}
	if downloadKind == "" {
		downloadKind = kindVideo
	}

	payload, err := json.Marshal(model.DownloadHistoryItem{
		Bvid:         strings.TrimSpace(task.Bvid),
		Cid:          task.Cid,
		Title:        strings.TrimSpace(task.Title),
		Cover:        strings.TrimSpace(task.Cover),
		Duration:     task.Duration,
		UpperName:    strings.TrimSpace(task.UpperName),
		Play:         task.Play,
		Danmaku:      task.Danmaku,
		Pubtime:      task.Pubtime,
		SourceName:   strings.TrimSpace(task.SourceName),
		SourceKind:   strings.TrimSpace(task.SourceKind),
		Path:         path,
		DownloadKind: downloadKind,
		Downloaded:   time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		s.logger.Errorf("marshal downloaded cache failed: %v", err)
		return
	}
	if err = s.store.Set(key, string(payload)); err != nil {
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
