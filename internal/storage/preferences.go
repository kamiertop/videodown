package storage

import (
	"errors"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/kamiertop/videodown/internal/constant"
)

func (s *Store) InitPreferenceDefaults(defaultSettings map[string]string) error {
	return s.Update(func(txn *badger.Txn) error {
		for key, value := range defaultSettings {
			if _, err := txn.Get([]byte(key)); errors.Is(err, badger.ErrKeyNotFound) {
				// 只有key不存在时才设置默认值，避免覆盖用户已设置的值
				if err = txn.Set([]byte(key), []byte(value)); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (s *Store) FilenameTemplate() (string, error) {
	return s.Get(constant.FilenameTemplateKey)
}

func (s *Store) AutoUpdate() (bool, error) {
	value, err := s.Get(constant.AutoUpdateKey)
	if err != nil {
		return false, err
	}

	return value == "true", nil
}

func (s *Store) StoragePath() (string, error) {
	return s.Get(constant.StorageKey)
}

func (s *Store) SavePreference() (bool, error) {
	value, err := s.Get(constant.AllowGroupOnSaveKey)
	if err != nil {
		return true, err
	}
	return value == "true", nil
}

func (s *Store) SetSavePreference(allowGroup bool) error {
	return s.Set(constant.AllowGroupOnSaveKey, strconv.FormatBool(allowGroup))
}

// SkipDownloaded 返回增量下载开关；开启时下载任务跳过历史中已存在且文件仍在的内容。
// 读取出错时按开启兜底（与设置服务 GetSkipDownloaded 一致），避免异常情况下意外全量重下。
func (s *Store) SkipDownloaded() bool {
	value, err := s.Get(constant.SkipDownloadedKey)
	if err != nil {
		return true
	}
	return value == "true"
}

func (s *Store) SleepTime() (int64, error) {
	value, err := s.Get(constant.SleepTimeKey)
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 60, nil
	}

	return seconds, nil
}

func (s *Store) ConcurrencyNum() (int, error) {
	value, err := s.Get(constant.ConcurrencyNumKey)
	if err != nil {
		return 1, err
	}

	return strconv.Atoi(value)
}

// ParsePlayURLNum 并发解析哔哩哔哩视频下载地址的数量
func (s *Store) ParsePlayURLNum() (int, error) {
	value, err := s.Get(constant.ParsePlayURLNumKey)
	if err != nil {
		return 3, err
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return 3, err
	}

	return num, nil
}

func (s *Store) ParsePlayURLSleep() (int, error) {
	value, err := s.Get(constant.ParsePlayURLSleepKey)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(value)
}
