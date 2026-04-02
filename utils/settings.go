package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/kamiertop/videodown/logger"
)

type Settings struct {
	db     *badger.DB
	logger *logger.Logger
}

func (s *Settings) init() error {
	return s.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(themeKey); errors.Is(err, badger.ErrKeyNotFound) {
			s.logger.Info("No theme found, setting to default: light")
			return txn.Set([]byte("theme"), []byte("light"))
		}
		if _, err := txn.Get(storageKey); errors.Is(err, badger.ErrKeyNotFound) {
			s.logger.Info("No storage path found, setting to default: ./downloads")
			if err := os.Mkdir("./downloads", 0755); err != nil {
				return fmt.Errorf("failed to create storage path [%s], err: %v", "./downloads", err)
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
	err := s.db.View(func(txn *badger.Txn) error {
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
	s := &Settings{db: db, logger: logger.WithName("Settings")}

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

	s := &Settings{db: db, logger: logger.WithName("Settings")}

	if err := s.init(); err != nil {
		panic(err)
	}

	return s
}

func (s *Settings) GetTheme() (string, error) {
	return s.getValue(themeKey)
}

func (s *Settings) SetTheme(theme string) error {
	if err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(themeKey, []byte(theme))
	}); err != nil {
		s.logger.Errorf("Failed to set new theme [%s], err: %v", theme, err)
		return err
	}
	s.logger.Infof("Theme set to: %s", theme)

	return nil
}

func (s *Settings) GetStorage() (string, error) {
	return s.getValue(storageKey)
}

func (s *Settings) SetStorage(path string) error {
	info, err := os.Stat(path)

	if err != nil {
		s.logger.Errorf("Failed to set new storage path [%s], err: %v", path, err)

		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("storage path [%s] does not exist", path)
		}
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("storage path [%s] is not accessible", path)
		}
		if !info.IsDir() {
			return fmt.Errorf("storage path [%s] is not a directory", path)
		}

		return fmt.Errorf("failed to access storage path [%s], err: %v", path, err)
	}

	if err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(storageKey, []byte(path))
	}); err != nil {
		s.logger.Infof("Failed to set new storage path [%s], err: %v", path, err)
		return err
	}
	s.logger.Infof("Storage path set to: %s", path)

	return nil
}
