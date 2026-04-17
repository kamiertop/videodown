package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type FFmpeg struct {
	path string
}

func NewFFmpeg() *FFmpeg {
	return &FFmpeg{}
}

func (f *FFmpeg) ensurePath() error {
	if f.path != "" {
		return nil
	}
	return f.searchFFmpeg()
}

func (f *FFmpeg) Merge(videoPath, audioPath, outputPath string) error {
	if err := f.ensurePath(); err != nil {
		return err
	}
	if videoPath == "" || audioPath == "" || outputPath == "" {
		return errors.New("video/audio/output path is empty")
	}

	cmd := exec.Command(
		f.path,
		"-y",
		"-loglevel", "error",
		"-nostdin",
		"-i", videoPath,
		"-i", audioPath,
		"-c", "copy",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg merge failed: %w: %s", err, string(out))
	}

	return nil
}

func (f *FFmpeg) Remux(inputPath, outputPath string) error {
	if err := f.ensurePath(); err != nil {
		return err
	}
	if inputPath == "" || outputPath == "" {
		return errors.New("input/output path is empty")
	}

	cmd := exec.Command(
		f.path,
		"-y",
		"-loglevel", "error",
		"-nostdin",
		"-i", inputPath,
		"-c", "copy",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg remux failed: %w: %s", err, string(out))
	}

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
