package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dgraph-io/badger/v4"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

// HasFFmpeg 检查系统中是否可用 ffmpeg
func (s *Settings) HasFFmpeg() bool {
	ffmpeg := NewFFmpeg()

	return ffmpeg.ensurePath() == nil
}

func (s *Settings) FFmpegPath() (string, error) {
	ffmpeg := NewFFmpeg()
	// 优先使用设置中保存的路径
	if path, err := s.GetKey(ffmpegPathKey); err == nil && path != "" {
		ffmpeg.path = path
		return path, nil
	}
	path, err := s.GetKey(ffmpegPathKey)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			s.logger.Error("ffmpeg path not set in settings, will try to search")
		}
	}
	// 如果key存在，也要判断是否有效，避免用户设置了错误路径后无法使用
	_, err = exec.Command(path, "-version").CombinedOutput()
	if err != nil {
		s.logger.Errorf("ffmpeg path from settings is invalid: %v, will try to search", err)
	} else {
		return path, nil
	}

	// 设置中没有或路径无效时，尝试搜索
	if err := ffmpeg.ensurePath(); err != nil {
		return "", err
	}

	return ffmpeg.path, nil
}

const ffmpegPathKey = "ffmpeg_path"

func (s *Settings) SetFFmpegPath(path string) error {
	if path == "" {
		return errors.New("ffmpeg path cannot be empty")
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("invalid ffmpeg path: %w", err)
	}

	// 直接验证是否可执行
	cmd := exec.Command(path, "-version")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg is not executable: %w: %s", err, string(out))
	}

	// 如果验证通过，保存路径
	ffmpeg := NewFFmpeg()
	ffmpeg.path = path

	return s.SetKey(ffmpegPathKey, path)
}

func (s *Settings) SelectFFmpegPath() (string, error) {
	path, err := wailsRuntime.OpenFileDialog(s.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择 FFmpeg 可执行文件",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "可执行文件",
				Pattern:     "ffmpeg;ffmpeg.exe;*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("打开文件选择对话框失败: %w", err)
	}
	if path == "" {
		return "", errors.New("未选择文件")
	}

	if err := s.SetFFmpegPath(path); err != nil {
		return "", err
	}

	return path, nil
}
