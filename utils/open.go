package utils

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	stdRuntime "runtime"
	"strings"
)

// OpenDownloadLocation opens the containing directory in the system file manager.
func OpenDownloadLocation(path string) error {
	target := strings.TrimSpace(path)
	if target == "" {
		return errors.New("文件路径为空")
	}

	info, err := os.Stat(target)
	if err == nil {
		if !info.IsDir() {
			target = filepath.Dir(target)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		target = filepath.Dir(target)
		if _, statErr := os.Stat(target); statErr != nil {
			return errors.New("文件所在目录不存在")
		}
	} else {
		return err
	}

	var cmd *exec.Cmd
	switch stdRuntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}

	return cmd.Start()
}

func OpenLocalFile(path string) error {
	target := strings.TrimSpace(path)
	if target == "" {
		return errors.New("文件路径为空")
	}
	if _, err := os.Stat(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("文件不存在")
		}
		return err
	}

	var cmd *exec.Cmd
	switch stdRuntime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}

	return cmd.Start()
}
