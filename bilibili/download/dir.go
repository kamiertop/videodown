package download

import (
	"path/filepath"
	"time"

	"github.com/kamiertop/videodown/utils"
)

// resolveTargetDir 根据用户设置的分组规则，返回最终下载目录路径。
// 收藏夹、合集和系列任务通常没有作者信息，此时直接使用来源目录。
func (s *Service) resolveTargetDir(storagePath string, task Task) (string, error) {
	cfg := s.store.Config().Download
	allowGroup := cfg.Grouping
	if !allowGroup {
		return storagePath, nil
	}

	sourceName := utils.FileName(task.SourceName)
	authorName := utils.FileName(task.UpperName)

	if authorName == "" {
		authorName = "未知作者"
	}

	rule := cfg.GroupingRule

	switch rule {
	case "source":
		return filepath.Join(storagePath, sourceName), nil
	case "author":
		return filepath.Join(storagePath, authorName), nil
	case "author_source":
		return filepath.Join(storagePath, authorName, sourceName), nil
	case "date":
		// 按下载当天的日期分目录；跨午夜的批量会自然落到两天的目录。
		return filepath.Join(storagePath, time.Now().Format(time.DateOnly)), nil
	default:
		// "none"
		return filepath.Join(storagePath), nil
	}
}
