package download

import (
	"path/filepath"

	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/utils"
)

// resolveTargetDir 根据用户设置的分组规则，返回最终下载目录路径。
// 收藏夹、合集和系列任务通常没有作者信息，此时直接使用来源目录。
func (s *Service) resolveTargetDir(storagePath string, task Task) (string, error) {
	allowGroup, err := s.store.SavePreference()
	if err != nil {
		return "", err
	}
	if !allowGroup {
		return storagePath, nil
	}

	sourceName := utils.FileName(task.SourceName)
	authorName := utils.FileName(task.UpperName)

	rule, _ := s.store.Get(constant.GroupingRuleKey)
	if authorName == "" {
		authorName = "未知作者"
	}

	switch rule {
	case "source":
		return filepath.Join(storagePath, sourceName), nil
	case "author":
		return filepath.Join(storagePath, authorName), nil
	case "author_source":
		return filepath.Join(storagePath, authorName, sourceName), nil
	default:
		// "none"
		return filepath.Join(storagePath), nil
	}
}
