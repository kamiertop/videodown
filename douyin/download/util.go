package download

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/utils"
)

const (
	kindVideo = "video"
	kindAlbum = "album"
	kindCover = "cover"
)

const cachePrefix = "douyin:downloaded:"

func cacheKey(awemeID string) string {
	id := strings.TrimSpace(awemeID)
	if id == "" {
		return ""
	}

	return cachePrefix + id
}

// normalizeDouyinHTTPURL 规范化抖音视频的 HTTP URL，确保以 https:// 开头
func normalizeDouyinHTTPURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}

	return rawURL
}

func uniqueDownloadTasks(tasks []Task) []Task {
	seen := make(map[string]struct{}, len(tasks))
	unique := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		key := strings.TrimSpace(task.AwemeID)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		unique = append(unique, task)
	}

	return unique
}

func douyinAssetExt(asset Asset) string {
	ext := strings.TrimSpace(asset.Ext)
	if ext == "" {
		// 素材未显式携带扩展名时，从 URL path（而非 query）推断原始格式。
		if parsed, err := url.Parse(strings.TrimSpace(asset.URL)); err == nil {
			name := path.Base(parsed.Path)
			if dot := strings.LastIndex(name, "."); dot >= 0 {
				ext = name[dot:]
			}
		}
	}
	if ext == "" {
		if asset.Kind == "video" {
			return ".mp4"
		}
		return ".jpg"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".mp4":
		return ext
	default:
		if asset.Kind == "video" {
			return ".mp4"
		}
		return ".jpg"
	}
}

func (d *Service) resolveDownloadDir(storagePath string, task Task) (string, error) {
	allowGroup, err := d.store.SavePreference()
	if err != nil {
		return "", err
	}
	if !allowGroup {
		return storagePath, nil
	}
	rule, _ := d.store.Get(constant.GroupingRuleKey)

	sourceName := utils.FileName(task.SourceName)
	author := utils.FileName(task.AuthorName)

	switch rule {
	case "source":
		return filepath.Join(storagePath, sourceName), nil
	case "author":
		return filepath.Join(storagePath, author), nil
	case "author_source":
		return filepath.Join(storagePath, author, sourceName), nil
	default:
		return filepath.Join(storagePath), nil
	}
}
