package storage

import (
	"errors"
	"strconv"

	"github.com/dgraph-io/badger/v4"
)

const (
	storageKey           = "storage"
	sleepTimeKey         = "sleepTime"
	allowGroupOnSaveKey  = "allowGroupOnSave"
	concurrencyNumKey    = "concurrencyNum"
	parsePlayURLNumKey   = "parse_play_url_num"
	parsePlayURLSleepKey = "parse_play_url_sleep"
	autoUpdateKey        = "auto_update"
)

func (s *Store) InitPreferenceDefaults(defaultStoragePath string) error {
	defaults := map[string]string{
		storageKey:           defaultStoragePath,
		sleepTimeKey:         "60",
		allowGroupOnSaveKey:  "true",
		concurrencyNumKey:    "1",
		parsePlayURLNumKey:   "3",
		parsePlayURLSleepKey: "5",
		autoUpdateKey:        "true",
	}
	return s.Update(func(txn *badger.Txn) error {
		for key, value := range defaults {
			if _, err := txn.Get([]byte(key)); errors.Is(err, badger.ErrKeyNotFound) {
				if err := txn.Set([]byte(key), []byte(value)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *Store) AutoUpdate() (bool, error) {
	value, err := s.Get(autoUpdateKey)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

func (s *Store) SetAutoUpdate(enabled bool) error {
	return s.Set(autoUpdateKey, strconv.FormatBool(enabled))
}

func (s *Store) StoragePath() (string, error) {
	return s.Get(storageKey)
}

func (s *Store) SetStoragePath(path string) error {
	return s.Set(storageKey, path)
}

func (s *Store) SavePreference() (bool, error) {
	value, err := s.Get(allowGroupOnSaveKey)
	if err != nil {
		return true, err
	}
	return value == "true", nil
}

func (s *Store) SetSavePreference(allowGroup bool) error {
	return s.Set(allowGroupOnSaveKey, strconv.FormatBool(allowGroup))
}

func (s *Store) SleepTime() (int64, error) {
	value, err := s.Get(sleepTimeKey)
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 60, nil
	}
	return seconds, nil
}

func (s *Store) SetSleepTime(seconds int64) error {
	return s.Set(sleepTimeKey, strconv.FormatInt(seconds, 10))
}

func (s *Store) ConcurrencyNum() (int, error) {
	value, err := s.Get(concurrencyNumKey)
	if err != nil {
		return 1, err
	}
	return strconv.Atoi(value)
}

func (s *Store) SetConcurrencyNum(num int) error {
	return s.Set(concurrencyNumKey, strconv.Itoa(num))
}

func (s *Store) ParsePlayURLNum() (int, error) {
	value, err := s.Get(parsePlayURLNumKey)
	if err != nil {
		return 3, err
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return 3, err
	}
	return num, nil
}

func (s *Store) SetParsePlayURLNum(num int) error {
	return s.Set(parsePlayURLNumKey, strconv.Itoa(num))
}

func (s *Store) ParsePlayURLSleep() (int, error) {
	value, err := s.Get(parsePlayURLSleepKey)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func (s *Store) SetParsePlayURLSleep(seconds int) error {
	return s.Set(parsePlayURLSleepKey, strconv.Itoa(seconds))
}
