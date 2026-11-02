// Package apppaths keeps portable application files inside the executable directory.
package apppaths

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var customDir = flag.String("data", "", "Data directory for storing preferences and history")

func Directory() (string, error) {
	// 使用指定的目录，用作测试或后续扩展
	if *customDir != "" {
		return filepath.Abs(*customDir)
	}

	// 使用可执行文件所在目录作为数据目录
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	executable, err = filepath.EvalSymlinks(executable)

	return filepath.Dir(executable), err
}

func Inside(root, name string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(name) {
		name = filepath.Join(root, name)
	}
	name = filepath.Clean(name)
	rel, err := filepath.Rel(root, name)
	if err != nil || (rel != "." && !filepath.IsLocal(rel)) {
		return "", fmt.Errorf("路径必须位于程序目录内：%s", name)
	}
	ancestor := name
	for {
		resolved, err := filepath.EvalSymlinks(ancestor)
		if err == nil {
			rel, err = filepath.Rel(root, resolved)
			if err != nil || (rel != "." && !filepath.IsLocal(rel)) {
				return "", fmt.Errorf("路径链接指向程序目录之外：%s", name)
			}
			return name, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if _, linkErr := os.Lstat(ancestor); linkErr == nil {
			return "", err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", err
		}
		ancestor = parent
	}
}
