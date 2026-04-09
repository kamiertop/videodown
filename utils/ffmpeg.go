package utils

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type FFmpeg struct {
	path string
}

func (f *FFmpeg) Merge() error {

	return nil
}

// searchFFmpeg 在多个位置搜索 ffmpeg
func (f *FFmpeg) searchFFmpeg() error {
	// 1. 尝试系统 PATH (最常用)
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		f.path = path
		return nil
	}

	// 2. 尝试应用同目录
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates := []string{
			filepath.Join(exeDir, "ffmpeg"),
			filepath.Join(exeDir, "bin", "ffmpeg"),
		}

		if runtime.GOOS == "windows" {
			for i, c := range candidates {
				candidates[i] = c + ".exe"
			}
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				f.path = candidate
				return nil
			}
		}
	}

	return errors.New("ffmpeg not found")
}
