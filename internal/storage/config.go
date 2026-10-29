package storage

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kamiertop/videodown/utils"
)

// Config is the application's TOML model. Snapshot it with Store.Config;
// commit edits through Store.UpdateConfig to preserve concurrent updates.
type Config struct {
	YouTube  YouTubeConfig  `toml:"youtube"`
	App      AppConfig      `toml:"app"`
	Download DownloadConfig `toml:"download"`
	Bilibili BilibiliConfig `toml:"bilibili"`
	Douyin   DouyinConfig   `toml:"douyin"`
}
type AppConfig struct {
	Theme       string `toml:"theme"`
	AutoUpdate  bool   `toml:"auto_update"`
	CloseToTray *bool  `toml:"close_to_tray,omitempty"`
}
type DownloadConfig struct {
	Directory        string `toml:"directory"`
	FFmpegPath       string `toml:"ffmpeg_path"`
	FilenameTemplate string `toml:"filename_template"`
	Grouping         bool   `toml:"grouping"`
	GroupingRule     string `toml:"grouping_rule"`
	Concurrency      int    `toml:"concurrency"`
	SleepSeconds     int64  `toml:"sleep_seconds"`
	SkipDownloaded   bool   `toml:"skip_downloaded"`
}
type BilibiliConfig struct {
	DefaultQuality    int          `toml:"default_quality"`
	ParseConcurrency  int          `toml:"parse_concurrency"`
	ParseSleepSeconds int          `toml:"parse_sleep_seconds"`
	Auth              BilibiliAuth `toml:"auth"`
}
type BilibiliAuth struct {
	Cookie       string `toml:"cookie"`
	CSRF         string `toml:"csrf"`
	UID          uint64 `toml:"uid"`
	RefreshToken string `toml:"refresh_token"`
}
type YouTubeConfig struct {
	Cookie string `toml:"cookie"`
}

type DouyinConfig struct {
	Cookie string `toml:"cookie"`
}

func DefaultConfig(directory string) Config {
	return Config{
		App: AppConfig{Theme: "light", AutoUpdate: true},
		Download: DownloadConfig{
			Directory: directory, FilenameTemplate: "{title}", Grouping: true, GroupingRule: "author_source", Concurrency: 1, SleepSeconds: 60, SkipDownloaded: true,
		},
		Bilibili: BilibiliConfig{ParseConcurrency: 3, ParseSleepSeconds: 5},
	}
}

func (c Config) Validate() error {
	if !slices.Contains([]string{"system", "light", "dark", "cupcake", "caramellatte"}, c.App.Theme) {
		return fmt.Errorf("不支持的主题：%s", c.App.Theme)
	}
	if strings.TrimSpace(c.Download.FilenameTemplate) == "" || !utils.SupportedFilenameTemplate(c.Download.FilenameTemplate) {
		return errors.New("命名模板必须包含受支持的变量")
	}
	if !slices.Contains([]string{"none", "author", "source", "author_source", "date"}, c.Download.GroupingRule) {
		return errors.New("无效的目录分组规则")
	}
	if c.Download.Concurrency < 1 || c.Download.Concurrency > 10 {
		return errors.New("下载并发数应在 1 到 10 之间")
	}
	if c.Download.SleepSeconds < 0 || c.Download.SleepSeconds > 3600 {
		return errors.New("休眠秒数应在 0 到 3600 之间")
	}
	if c.Bilibili.ParseConcurrency < 1 || c.Bilibili.ParseConcurrency > 5 {
		return errors.New("解析并发数应在 1 到 5 之间")
	}
	if c.Bilibili.ParseSleepSeconds < 0 || c.Bilibili.ParseSleepSeconds > 120 {
		return errors.New("解析休眠秒数应在 0 到 120 之间")
	}
	if c.Bilibili.DefaultQuality < 0 {
		return errors.New("默认画质不能小于 0")
	}

	return nil
}

func cloneConfig(c Config) Config {
	if c.App.CloseToTray != nil {
		value := *c.App.CloseToTray
		c.App.CloseToTray = &value
	}

	return c
}

// Config returns an independent typed snapshot; mutating it does not change the store.
func (s *Store) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneConfig(s.config)
}

// UpdateConfig applies an edit to the latest snapshot, validates it, and atomically
// saves it. The callback must not call Store methods or retain its argument.
func (s *Store) UpdateConfig(edit func(*Config) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := cloneConfig(s.config)
	if err := edit(&next); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if err := s.writeConfig(next); err != nil {
		return err
	}
	s.config = cloneConfig(next)

	return nil
}

// InitDefaults supplies an installation-specific directory on first launch.
func (s *Store) InitDefaults(directory string) error {
	if s.Config().Download.Directory != "" {
		return nil
	}
	return s.UpdateConfig(func(c *Config) error {
		if c.Download.Directory == "" {
			c.Download.Directory = directory
		}
		return nil
	})
}
