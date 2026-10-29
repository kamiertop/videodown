package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type FFmpeg struct {
	path string
}

// NewFFmpeg uses the configured executable, or discovers one when custom is empty.
func NewFFmpeg(custom string) *FFmpeg { return &FFmpeg{path: custom} }

func (f *FFmpeg) ensurePath() error {
	path, _, err := CheckFFmpeg(f.path)
	if err == nil {
		f.path = path
	}
	return err
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

func CheckFFmpeg(custom string) (path, version string, err error) {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	var candidates []string
	if custom != "" {
		candidates = append(candidates, custom)
	} else {
		if executable, e := os.Executable(); e == nil {
			dir := filepath.Dir(executable)
			candidates = append(candidates, filepath.Join(dir, name), filepath.Join(dir, "bin", name))
		}
		if found, e := exec.LookPath(name); e == nil {
			candidates = append(candidates, found)
		}
	}
	for _, candidate := range candidates {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		output, e := exec.CommandContext(ctx, candidate, "-version").CombinedOutput()
		cancel()
		if e == nil {
			return candidate, strings.SplitN(string(output), "\n", 2)[0], nil
		}
	}
	return "", "", fmt.Errorf("未检测到可用的 FFmpeg，请安装或手动指定路径")
}
