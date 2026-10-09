package download

import (
	"net/url"
	"path"
	"path/filepath"
	"regexp"
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

// trailingHashtags 匹配抖音简介尾部的话题标签串（如 " #猫咪 #养猫日常"）。
// 每个标签前必须是行首或空白分隔（\s 不含全角空格，补 \p{Zs}），只剥
// “独立的”标签串：纯标签标题因此剥到空、触发保留原文的守卫，而正文里
// 内嵌的井号（如 “含#号的正文”）不会被误伤。
var trailingHashtags = regexp.MustCompile(`(?:(?:^|[\s\p{Zs}]+)#[^\s#\p{Zs}]+)+[\s\p{Zs}]*$`)

// cleanDouyinTitle 清洗用作文件名/目录名的抖音标题：去掉尾部的话题标签串。
// 标签是检索用的元信息而非内容本身，却常常占据超长简介的大头；剥离后为空
// 的纯标签标题保留原文，避免标题意外变空。
func cleanDouyinTitle(title string) string {
	trimmed := trailingHashtags.ReplaceAllString(strings.TrimSpace(title), "")
	if strings.TrimSpace(trimmed) == "" {
		return strings.TrimSpace(title)
	}
	return trimmed
}

// safeFileName 名字超过安全字节上限时退回 fallback（视频 ID）：
// 抖音简介动辄数百字，超限名字会让 os.Stat/os.Create 返回 ENAMETOOLONG。
func safeFileName(name, fallback string) string {
	if len(name) > utils.MaxFileNameBytes {
		return fallback
	}
	return name
}

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
	rule, err := d.store.Get(constant.GroupingRuleKey)
	if err != nil {
		return "", err
	}
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
