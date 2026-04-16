package utils

import (
	"context"
	"errors"
	"os"
	"time"

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
		if _, err := txn.Get([]byte(themeKey)); errors.Is(err, badger.ErrKeyNotFound) {
			s.logger.Info("No theme found, setting to default: light")
			return txn.Set([]byte("theme"), []byte("light"))
		}
		if _, err := txn.Get([]byte(storageKey)); errors.Is(err, badger.ErrKeyNotFound) {
			s.logger.Info("No storage path found, setting to default: ./downloads")
			if err := os.Mkdir("./downloads", 0755); err != nil {
				s.logger.Errorf("failed to create default storage path [./downloads]: %v", err)
				return errors.New("创建默认存储目录失败")
			}

			return txn.Set([]byte(storageKey), []byte("./downloads"))
		}
		return nil
	})
}

const (
	themeKey   = "theme"
	storageKey = "storage"
	sleepTime  = time.Minute
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

func (s *Settings) GetSleepTime() time.Duration {
	return sleepTime
}
