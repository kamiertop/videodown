// Package preferences adapts the shared application keys for native settings.
package preferences

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/utils"
)

type Store interface {
	Config() storage.Config
	UpdateConfig(func(*storage.Config) error) error
}
type Data struct {
	Directory, Theme, Filename, GroupingRule, FFmpeg string
	AutoUpdate, SkipDownloaded, Grouping             bool
	Concurrency, Sleep, ParseConcurrency, ParseSleep float64
}

func Defaults() Data { return FromConfig(storage.DefaultConfig("")) }
func FromConfig(c storage.Config) Data {
	return Data{Directory: c.Download.Directory, Theme: c.App.Theme, Filename: c.Download.FilenameTemplate, GroupingRule: c.Download.GroupingRule, FFmpeg: c.Download.FFmpegPath,
		AutoUpdate: c.App.AutoUpdate, SkipDownloaded: c.Download.SkipDownloaded, Grouping: c.Download.Grouping, Concurrency: float64(c.Download.Concurrency), Sleep: float64(c.Download.SleepSeconds), ParseConcurrency: float64(c.Bilibili.ParseConcurrency), ParseSleep: float64(c.Bilibili.ParseSleepSeconds)}
}
func (d Data) writeConfig(c *storage.Config) {
	c.App.Theme, c.App.AutoUpdate = d.Theme, d.AutoUpdate
	c.Download.Directory, c.Download.FFmpegPath, c.Download.FilenameTemplate, c.Download.GroupingRule = d.Directory, d.FFmpeg, d.Filename, d.GroupingRule
	c.Download.Grouping, c.Download.SkipDownloaded = d.Grouping, d.SkipDownloaded
	c.Download.Concurrency, c.Download.SleepSeconds = int(d.Concurrency), int64(d.Sleep)
	c.Bilibili.ParseConcurrency, c.Bilibili.ParseSleepSeconds = int(d.ParseConcurrency), int(d.ParseSleep)
}
func Load(store Store) (Data, error) {
	if store == nil {
		return Defaults(), nil
	}
	return FromConfig(store.Config()), nil
}
func (d Data) Values() map[string]string {
	return map[string]string{
		constant.StorageKey: d.Directory, constant.ThemeKey: d.Theme, constant.FilenameTemplateKey: d.Filename, constant.GroupingRuleKey: d.GroupingRule, constant.FFmpegPathKey: d.FFmpeg,
		constant.AutoUpdateKey: strconv.FormatBool(d.AutoUpdate), constant.SkipDownloadedKey: strconv.FormatBool(d.SkipDownloaded), constant.AllowGroupOnSaveKey: strconv.FormatBool(d.Grouping),
		constant.ConcurrencyNumKey: strconv.Itoa(int(d.Concurrency)), constant.SleepTimeKey: strconv.Itoa(int(d.Sleep)), constant.ParsePlayURLNumKey: strconv.Itoa(int(d.ParseConcurrency)), constant.ParsePlayURLSleepKey: strconv.Itoa(int(d.ParseSleep)),
	}
}
func (d *Data) Apply(values map[string]string) error {
	next := *d
	for key, value := range values {
		switch key {
		case constant.StorageKey:
			next.Directory = value
		case constant.FFmpegPathKey:
			next.FFmpeg = value
		case constant.ThemeKey:
			if value != "system" && value != "light" && value != "dark" && value != "cupcake" && value != "caramellatte" {
				return fmt.Errorf("不支持的主题：%s", value)
			}
			next.Theme = value
		case constant.FilenameTemplateKey:
			value = strings.TrimSpace(value)
			if value == "" || !utils.SupportedFilenameTemplate(value) {
				return errors.New("命名模板必须包含受支持的变量")
			}
			next.Filename = value
		case constant.GroupingRuleKey:
			if value != "none" && value != "author" && value != "source" && value != "author_source" && value != "date" {
				return errors.New("无效的目录分组规则")
			}
			next.GroupingRule = value
		case constant.AutoUpdateKey, constant.SkipDownloadedKey, constant.AllowGroupOnSaveKey:
			enabled, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			switch key {
			case constant.AutoUpdateKey:
				next.AutoUpdate = enabled
			case constant.SkipDownloadedKey:
				next.SkipDownloaded = enabled
			default:
				next.Grouping = enabled
			}
		case constant.ConcurrencyNumKey, constant.SleepTimeKey, constant.ParsePlayURLNumKey, constant.ParsePlayURLSleepKey:
			n, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			lo, hi := 0, 3600
			switch key {
			case constant.ConcurrencyNumKey:
				lo, hi = 1, 10
			case constant.ParsePlayURLNumKey:
				lo, hi = 1, 5
			case constant.ParsePlayURLSleepKey:
				hi = 120
			}
			if n < lo || n > hi {
				return fmt.Errorf("%s 应在 %d 到 %d 之间", key, lo, hi)
			}
			switch key {
			case constant.ConcurrencyNumKey:
				next.Concurrency = float64(n)
			case constant.SleepTimeKey:
				next.Sleep = float64(n)
			case constant.ParsePlayURLNumKey:
				next.ParseConcurrency = float64(n)
			default:
				next.ParseSleep = float64(n)
			}
		default:
			return fmt.Errorf("不支持的配置项：%s", key)
		}
	}
	*d = next
	return nil
}
func Save(store Store, values map[string]string) error {
	if store == nil {
		return errors.New("配置存储尚未初始化")
	}
	return store.UpdateConfig(func(c *storage.Config) error {
		data := FromConfig(*c)
		if err := data.Apply(values); err != nil {
			return err
		}
		data.writeConfig(c)
		return nil
	})
}

// Ensure the existing shared database implements the native settings adapter.
var _ Store = (*storage.Store)(nil)
