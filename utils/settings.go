package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/kamiertop/videodown/logger"
)

type Settings struct {
	DB     *badger.DB
	logger *logger.Logger
}

func (s *Settings) init() error {
	return s.DB.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(themeKey); errors.Is(err, badger.ErrKeyNotFound) {
			s.logger.Info("No theme found, setting to default: light")
			return txn.Set([]byte("theme"), []byte("light"))
		}
		if _, err := txn.Get(storageKey); errors.Is(err, badger.ErrKeyNotFound) {
			s.logger.Info("No storage path found, setting to default: ./downloads")
			if err := os.Mkdir("./downloads", 0755); err != nil {
				s.logger.Errorf("failed to create default storage path [./downloads]: %v", err)
				return errors.New("创建默认存储目录失败")
			}

			return txn.Set(storageKey, []byte("./downloads"))
		}
		return nil
	})
}

var themeKey = []byte("theme")
var storageKey = []byte("storage")

func (s *Settings) getValue(key []byte) (string, error) {
	var result string
	err := s.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
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
		s.logger.Infof("Get %s: %s", string(key), result)
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
	theme, err := s.getValue(themeKey)
	if err != nil {
		s.logger.Errorf("failed to get theme: %v", err)
		return "", errors.New("获取主题设置失败")
	}

	return theme, nil
}

func (s *Settings) SetTheme(theme string) error {
	if err := s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(themeKey, []byte(theme))
	}); err != nil {
		s.logger.Errorf("Failed to set new theme [%s], err: %v", theme, err)
		return errors.New("设置主题失败")
	}
	s.logger.Infof("Theme set to: %s", theme)

	return nil
}

func (s *Settings) GetStorage() (string, error) {
	path, err := s.getValue(storageKey)
	if err != nil {
		s.logger.Errorf("failed to get storage path: %v", err)
		return "", errors.New("获取存储目录失败")
	}

	return path, nil
}

func (s *Settings) SetStorage(path string) error {
	info, err := os.Stat(path)

	if err != nil {
		s.logger.Errorf("Failed to set new storage path [%s], err: %v", path, err)

		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("存储目录 [%s] 不存在", path)
		}
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("存储目录 [%s] 无法访问", path)
		}

		return errors.New("访问存储目录失败")
	}

	if !info.IsDir() {
		return fmt.Errorf("存储目录 [%s] 不是文件夹", path)
	}

	if err = s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(storageKey, []byte(path))
	}); err != nil {
		s.logger.Errorf("Failed to set new storage path [%s], err: %v", path, err)
		return errors.New("设置存储目录失败")
	}
	s.logger.Infof("Storage path set to: %s", path)

	return nil
}
