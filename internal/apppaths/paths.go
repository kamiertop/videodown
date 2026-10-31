// Package apppaths keeps portable application files inside the executable directory.
package apppaths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Directory() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return "", err
	}
	return filepath.Dir(executable), nil
}

// Inside also checks existing parent symlinks before a new path is created.
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
