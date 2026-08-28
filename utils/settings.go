package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	stdRuntime "runtime"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
)

const (
	// themeKey 主题设置，默认为 "light"
	themeKey = "theme"
	// closeToTrayKey 关闭按钮行为，无默认值。用户首次点击关闭时弹窗选择后才写入。
	closeToTrayKey = "closeToTray"
)

type Settings struct {
	store  *storage.Store
	logger *logger.Logger
}

func (s *Settings) init() error {
	executable, err := os.Executable()
	if err != nil {
		s.logger.Errorf("Get VideoDown Executable Path Error: %v", err)
		return err
	}

	defaultStoragePath := filepath.Join(filepath.Dir(executable), "download")
	if err := s.store.InitPreferenceDefaults(defaultStoragePath); err != nil {
		return err
	}
	s.logger.Infof("set default storage path: %s", defaultStoragePath)

	return s.store.Update(func(txn *badger.Txn) error {
		defaultValue := map[string]string{
			themeKey: "light",
			// 其他设置项的默认值
		}
		var errList error
		for key, value := range defaultValue {
			if _, err := txn.Get([]byte(key)); errors.Is(err, badger.ErrKeyNotFound) {
				s.logger.Infof("No %s found, setting to default: %s", key, value)
				// 只有在 key 不存在时才设置默认值，避免覆盖用户已修改的设置。
				if err := txn.Set([]byte(key), []byte(value)); err != nil {
					errList = errors.Join(errList, fmt.Errorf("failed to set key: [%s], value: [%s], err: %w", key, value, err))
				}
			}
		}

		return errList
	})
}

func NewSettingsWithMemory(logger *logger.Logger) *Settings {
	store, err := storage.OpenMemory()
	if err != nil {
		panic(err)
	}
	s := &Settings{store: store, logger: logger.WithName("Settings")}

	if err = s.init(); err != nil {
		_ = store.Close()
		panic(err)
	}

	return s
}

func NewSettings(logger *logger.Logger, store *storage.Store) (*Settings, error) {
	s := &Settings{store: store, logger: logger.WithName("Settings")}
	if err := s.init(); err != nil {
		_ = store.Close()
		return nil, err
	}
	return s, nil
}

// GetTheme 获取主题设置
func (s *Settings) GetTheme() (string, error) {
	theme, err := s.store.Get(themeKey)
	if err != nil {
		s.logger.Errorf("failed to get theme: %v", err)
		return "", errors.New("获取主题设置失败")
	}

	return theme, nil
}

// SetTheme 设置主题，前端调用时会传入 "light"、"dark"等，落库保存供下次启动时加载使用
func (s *Settings) SetTheme(theme string) error {
	if err := s.store.Set(themeKey, theme); err != nil {
		s.logger.Errorf("Failed to set new theme [%s], err: %v", theme, err)
		return errors.New("设置主题失败")
	}
	s.logger.Infof("Theme set to: %s", theme)

	return nil
}

// GetStorage 获取存储目录设置
func (s *Settings) GetStorage() (string, error) {
	path, err := s.store.StoragePath()
	if err != nil {
		s.logger.Errorf("failed to get storage path: %v", err)
		return "", errors.New("获取存储目录失败")
	}

	return path, nil
}

// SetStoragePath 保存存储目录。
func (s *Settings) SetStoragePath(dir string) error {
	if err := s.store.SetStoragePath(dir); err != nil {
		s.logger.Errorf("Failed to set new storage path [%s], err: %v", dir, err)
		return errors.New("设置存储目录失败")
	}
	s.logger.Infof("Storage path set to: %s", dir)

	return nil
}

// GetSleepTime 下载完一个视频之后的休眠时间；配置值按“秒”保存，避免把默认值 60 误解释成 60 纳秒。
func (s *Settings) GetSleepTime() (int64, error) {
	value, err := s.store.SleepTime()
	if err != nil {
		s.logger.Errorf("failed to get sleep time: %v", err)
		return 0, errors.New("获取休眠时间失败")
	}
	return value, nil
}

// SetSleepTime 保存休眠秒数；前端传入 time.Duration 时统一落库为秒，便于用户理解和配置。
func (s *Settings) SetSleepTime(d int64) error {
	s.logger.Infof("setting sleep time: %d", d)

	return s.store.SetSleepTime(d)
}

func (s *Settings) GetSavePreference() (bool, error) {
	return s.store.SavePreference()
}

// SetSavePreference 保存时是否自动分组
func (s *Settings) SetSavePreference(allowGroup bool) error {
	s.logger.Infof("Setting save preference to: %t", allowGroup)

	return s.store.SetSavePreference(allowGroup)
}

// GetConcurrencyNum 获取同时下载的视频数量
func (s *Settings) GetConcurrencyNum() (int, error) {
	return s.store.ConcurrencyNum()
}

// SetConcurrencyNum 保存同时下载的视频数量
func (s *Settings) SetConcurrencyNum(num int) error {
	s.logger.Infof("Setting concurrency num to %d", num)
	return s.store.SetConcurrencyNum(num)
}

// OpenDownloadLocation 打开下载历史中的文件位置.
func (s *Settings) OpenDownloadLocation(path string) error {
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

// OpenLocalFile 使用系统默认程序打开下载历史中的本地文件或目录.
func (s *Settings) OpenLocalFile(path string) error {
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

// HasCloseToTrayChoice 检查用户是否已对关闭行为做出选择（key 是否存在）。
func (s *Settings) HasCloseToTrayChoice() bool {
	_, err := s.store.Get(closeToTrayKey)
	return err == nil
}

// IsCloseToTray 返回用户选择：true=缩小到托盘，false=退出程序。未选择时默认返回 true。
func (s *Settings) IsCloseToTray() bool {
	v, err := s.GetCloseToTray()
	if err != nil {
		return true // 未选择时默认缩小到托盘
	}
	return v
}

// GetCloseToTray 获取关闭按钮行为设置。
func (s *Settings) GetCloseToTray() (bool, error) {
	val, err := s.store.Get(closeToTrayKey)
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

// SetCloseToTray 设置关闭按钮行为：false=退出程序，true=缩小到托盘。
func (s *Settings) SetCloseToTray(v bool) error {
	if v {
		return s.store.Set(closeToTrayKey, "true")
	}
	return s.store.Set(closeToTrayKey, "false")
}

// ServiceShutdown closes resources when the Wails application shuts down.
func (s *Settings) ServiceShutdown() error {
	if err := s.store.Close(); err != nil {
		s.logger.Errorf("Failed to close settings DB: %v", err)
		return err
	}
	s.logger.Info("Close BadgerDB and shutdown application")
	return nil
}
