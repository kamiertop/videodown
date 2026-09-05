package download

import (
	"path/filepath"
	"strings"

	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/utils"
)

// resolveTargetDir 根据用户设置的分组规则，返回最终下载目录路径
// none： 不分组，直接返回 storagePath
// author：按作者分组，返回 storagePath/upperName
// author+source：按作者和来源分组，返回 storagePath/upperName/sourceName
func (s *Service) resolveTargetDir(storagePath string, task Task) (string, error) {
	allowGroup, err := s.store.SavePreference()
	if err != nil {
		return "", err
	}
	if !allowGroup {
		return storagePath, nil
	}
	rule, _ := s.store.Get(constant.GroupingRuleKey)
	if rule == "none" {
		return storagePath, nil
	}

	upperName := utils.FileName(task.UpperName)
	if upperName == "" {
		upperName = "未知作者"
	}

	authorDir := filepath.Join(storagePath, upperName)
	if rule == "author" {
		return authorDir, nil
	}
	switch strings.TrimSpace(task.SourceKind) {
	case "收藏夹", "合集", "系列":
		if sourceName := utils.FileName(task.SourceName); sourceName != "" {
			return filepath.Join(authorDir, sourceName), nil
		}
	}

	return authorDir, nil
}
