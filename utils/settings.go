package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/kamiertop/videodown/logger"
)

type Settings struct {
	*badger.DB
	logger *logger.Logger
}

func (s *Settings) init() error {
	return s.DB.Update(func(txn *badger.Txn) error {
		defaultValue := map[string]string{
			themeKey:            "light",
			storageKey:          "./downloads",
			allowGroupOnSaveKey: "true",
			sleepTimeKey:        "60",
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

const (
	themeKey            = "theme"
	storageKey          = "storage"
	sleepTimeKey        = "sleepTime"
	allowGroupOnSaveKey = "allowGroupOnSave"
)

func (s *Settings) GetKey(key string) (string, error) {
	var result string
	err := s.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			result = string(val)
			return nil
		})
		if err != nil {
			return err
		}
		s.logger.Infof("Get %s: %s", key, result)
		return nil
	})

	return result, err
}

func NewSettingsWithMemory(logger *logger.Logger) *Settings {
	db, err := badger.Open(badger.
		DefaultOptions("").
		WithInMemory(true).
		WithLoggingLevel(badger.ERROR))
	if err != nil {
		panic(err)
	}
	s := &Settings{DB: db, logger: logger.WithName("Settings")}

	if err := s.init(); err != nil {
		panic(err)
	}

	return s
}

func NewSettings(logger *logger.Logger) *Settings {
	db, err := badger.Open(badger.DefaultOptions("videodown.db").WithLogger(logger).WithLoggingLevel(badger.ERROR))
	if err != nil {
		panic(err)
	}

	s := &Settings{DB: db, logger: logger.WithName("Settings")}

	if err := s.init(); err != nil {
		panic(err)
	}

	return s
}

func (s *Settings) GetTheme() (string, error) {
	theme, err := s.GetKey(themeKey)
	if err != nil {
		s.logger.Errorf("failed to get theme: %v", err)
		return "", errors.New("获取主题设置失败")
	}

	return theme, nil
}

func (s *Settings) SetTheme(theme string) error {
	if err := s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(themeKey), []byte(theme))
	}); err != nil {
		s.logger.Errorf("Failed to set new theme [%s], err: %v", theme, err)
		return errors.New("设置主题失败")
	}
	s.logger.Infof("Theme set to: %s", theme)

	return nil
}

func (s *Settings) GetStorage() (string, error) {
	path, err := s.GetKey(storageKey)
	if err != nil {
		s.logger.Errorf("failed to get storage path: %v", err)
		return "", errors.New("获取存储目录失败")
	}

	return path, nil
}

// SetStorage 前端不能使用这个，依赖App的ctx
func (s *Settings) SetStorage(ctx context.Context) (string, error) {
	dir, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title: "选择下载目录",
	})

	if err = s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(storageKey), []byte(dir))
	}); err != nil {
		s.logger.Errorf("Failed to set new storage path [%s], err: %v", dir, err)
		return "", errors.New("设置存储目录失败")
	}
	s.logger.Infof("Storage path set to: %s", dir)

	return dir, nil
}

func (s *Settings) SetKey(key, value string) error {
	return s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

// GetSleepTime 下载完一个视频之后的休眠时间；配置值按“秒”保存，避免把默认值 60 误解释成 60 纳秒。
func (s *Settings) GetSleepTime() (int64, error) {
	value, err := s.GetKey(sleepTimeKey)
	if err != nil {
		s.logger.Errorf("failed to get sleep time: %v", err)
		return 0, errors.New("获取休眠时间失败")
	}
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		s.logger.Errorf("failed to parse sleep time: %v", err)
		return 60, nil
	}

	return val, nil
}

// SetSleepTime 保存休眠秒数；前端传入 time.Duration 时统一落库为秒，便于用户理解和配置。
func (s *Settings) SetSleepTime(d int64) error {
	return s.SetKey(sleepTimeKey, strconv.FormatInt(d, 10))
}

func (s *Settings) GetSavePreference() (bool, error) {
	key, err := s.GetKey(allowGroupOnSaveKey)
	if err != nil {
		return true, err
	}
	if key == "true" {
		return true, nil
	}

	return false, nil
}

// SetSavePreference 保存时是否自动分组
func (s *Settings) SetSavePreference(allowGroup bool) error {
	var b string
	if allowGroup {
		b = "true"
	} else {
		b = "false"
	}

	return s.SetKey(allowGroupOnSaveKey, b)
}
