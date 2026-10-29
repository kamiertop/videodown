// Package storage keeps user preferences in TOML and application data in SQLite.
package storage

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/ncruces/go-sqlite3/driver"
	"github.com/pelletier/go-toml/v2"
)

var ErrNotFound = errors.New("存储项不存在")

type Store struct {
	db         *sqlx.DB
	mu         sync.RWMutex
	configPath string
	config     Config
}

// Open uses a directory containing config.toml and history.sqlite.
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("path cannot be empty")
	}
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(path, "history.sqlite")
	f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	dsn := (&url.URL{Scheme: "file", Path: filepath.ToSlash(dbPath)}).String()
	s, err := openDB(dsn, filepath.Join(path, "config.toml"))
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(s.configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		s.Close()
		return nil, err
	}
	if errors.Is(err, os.ErrNotExist) {
		if err = s.writeConfig(s.config); err != nil {
			_ = s.Close()
			return nil, fmt.Errorf("创建配置: %w", err)
		}
	}
	if len(content) != 0 {
		if err := toml.NewDecoder(bytes.NewReader(content)).DisallowUnknownFields().Decode(&s.config); err != nil {
			s.Close()
			var unknown *toml.StrictMissingError
			if errors.As(err, &unknown) {
				fields := make([]string, 0, min(4, len(unknown.Errors)))
				for _, field := range unknown.Errors[:min(4, len(unknown.Errors))] {
					fields = append(fields, strings.Join(field.Key(), "."))
				}
				return nil, fmt.Errorf("配置包含不支持的字段：%s；请参考 config.example.toml 使用 app、download、bilibili、douyin 分组", strings.Join(fields, "、"))
			}
			return nil, fmt.Errorf("读取配置: %w", err)
		}
		if err := s.config.Validate(); err != nil {
			s.Close()
			return nil, fmt.Errorf("配置无效: %w", err)
		}
	}
	return s, nil
}

func OpenMemory() (*Store, error) { return openDB("file::memory:", "") }

func openDB(dsn, configPath string) (*Store, error) {
	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, query := range []string{
		`PRAGMA busy_timeout=5000`,
		`CREATE TABLE IF NOT EXISTS download_history (key TEXT PRIMARY KEY, platform TEXT NOT NULL, bvid TEXT NOT NULL DEFAULT '', cid INTEGER NOT NULL DEFAULT 0, title TEXT NOT NULL DEFAULT '', author TEXT NOT NULL DEFAULT '', downloaded TEXT NOT NULL DEFAULT '', payload TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS download_history_platform_time ON download_history(platform, downloaded DESC)`,
		`CREATE INDEX IF NOT EXISTS download_history_bvid ON download_history(platform, bvid)`,
	} {
		if _, err := db.Exec(query); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &Store{db: db, configPath: configPath, config: DefaultConfig("")}, nil
}

func historyPlatform(key string) string {
	if strings.HasPrefix(key, "bilibili:downloaded:") {
		return "bilibili"
	}
	if strings.HasPrefix(key, "douyin:downloaded:") {
		return "douyin"
	}
	return ""
}

func (s *Store) History(key string) (string, error) {
	var value string
	err := s.db.Get(&value, `SELECT payload FROM download_history WHERE key=?`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}

	return value, err
}

func (s *Store) SaveHistory(key, value string) error {
	if platform := historyPlatform(key); platform != "" {
		return s.setHistory(key, platform, value)
	}
	return errors.New("无效的下载历史键")
}

func (s *Store) DeleteHistory(key string) error {
	_, err := s.db.Exec(`DELETE FROM download_history WHERE key=?`, key)
	return err
}

func (s *Store) DeleteHistoryPrefix(prefix string) error {
	if historyPlatform(prefix) == "" {
		return errors.New("unsupported prefix")
	}
	_, err := s.db.Exec(`DELETE FROM download_history WHERE substr(key,1,length(?))=?`, prefix, prefix)
	return err
}

// HistoryRows returns JSON models stored for one platform, newest first.
func (s *Store) HistoryRows(platform string) ([]string, error) {
	var rows []string

	return rows, s.db.Select(&rows, `SELECT payload FROM download_history WHERE platform=? ORDER BY downloaded DESC`, platform)
}

// BilibiliVideoRows excludes audio and cover records for downloaded-video matching.
func (s *Store) BilibiliVideoRows() ([]string, error) {
	var rows []string
	err := s.db.Select(&rows, `SELECT payload FROM download_history WHERE platform='bilibili' AND cid>0 AND key NOT LIKE 'bilibili:downloaded:audio:%' AND key NOT LIKE 'bilibili:downloaded:cover:%'`)
	return rows, err
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) writeConfig(config Config) error {
	if s.configPath == "" {
		return nil
	}
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.configPath)
	f, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.configPath)
}

// DownloadDirectory resolves portable relative paths against config.toml, never the caller's working directory.
func (s *Store) DownloadDirectory() string {
	directory := s.Config().Download.Directory
	if directory != "" && !filepath.IsAbs(directory) && s.configPath != "" {
		return filepath.Join(filepath.Dir(s.configPath), directory)
	}
	return directory
}
