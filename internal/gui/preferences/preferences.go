// Package preferences adapts typed application configuration to native controls.
package preferences

import (
	"errors"
	"strings"

	"github.com/kamiertop/videodown/internal/storage"
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

func Defaults() Data {
	return FromConfig(storage.DefaultConfig(""))
}

func FromConfig(c storage.Config) Data {
	return Data{Directory: c.Download.Directory, Theme: c.App.Theme, Filename: c.Download.FilenameTemplate, GroupingRule: c.Download.GroupingRule, FFmpeg: c.Download.FFmpegPath,
		AutoUpdate: c.App.AutoUpdate, SkipDownloaded: c.Download.SkipDownloaded, Grouping: c.Download.Grouping, Concurrency: float64(c.Download.Concurrency), Sleep: float64(c.Download.SleepSeconds), ParseConcurrency: float64(c.Bilibili.ParseConcurrency), ParseSleep: float64(c.Bilibili.ParseSleepSeconds)}
}

// Changes holds only edited fields, preserving unrelated settings and credentials.
type Changes struct {
	Directory        *string
	Theme            *string
	Filename         *string
	GroupingRule     *string
	FFmpeg           *string
	AutoUpdate       *bool
	SkipDownloaded   *bool
	Grouping         *bool
	Concurrency      *int
	Sleep            *int64
	ParseConcurrency *int
	ParseSleep       *int
}

func (p Changes) Merge(next Changes) Changes {
	if next.Directory != nil {
		p.Directory = new(*next.Directory)
	}
	if next.Theme != nil {
		p.Theme = new(*next.Theme)
	}
	if next.Filename != nil {
		p.Filename = new(*next.Filename)
	}
	if next.GroupingRule != nil {
		p.GroupingRule = new(*next.GroupingRule)
	}
	if next.FFmpeg != nil {
		p.FFmpeg = new(*next.FFmpeg)
	}
	if next.AutoUpdate != nil {
		p.AutoUpdate = new(*next.AutoUpdate)
	}
	if next.SkipDownloaded != nil {
		p.SkipDownloaded = new(*next.SkipDownloaded)
	}
	if next.Grouping != nil {
		p.Grouping = new(*next.Grouping)
	}
	if next.Concurrency != nil {
		p.Concurrency = new(*next.Concurrency)
	}
	if next.Sleep != nil {
		p.Sleep = new(*next.Sleep)
	}
	if next.ParseConcurrency != nil {
		p.ParseConcurrency = new(*next.ParseConcurrency)
	}
	if next.ParseSleep != nil {
		p.ParseSleep = new(*next.ParseSleep)
	}

	return p
}
func (p Changes) Empty() bool {
	return p.Directory == nil && p.Theme == nil && p.Filename == nil && p.GroupingRule == nil && p.FFmpeg == nil && p.AutoUpdate == nil && p.SkipDownloaded == nil && p.Grouping == nil && p.Concurrency == nil && p.Sleep == nil && p.ParseConcurrency == nil && p.ParseSleep == nil
}
func (p Changes) apply(c *storage.Config) {
	if p.Directory != nil {
		c.Download.Directory = *p.Directory
	}
	if p.Theme != nil {
		c.App.Theme = *p.Theme
	}
	if p.Filename != nil {
		c.Download.FilenameTemplate = *p.Filename
	}
	if p.GroupingRule != nil {
		c.Download.GroupingRule = *p.GroupingRule
	}
	if p.FFmpeg != nil {
		c.Download.FFmpegPath = *p.FFmpeg
	}
	if p.AutoUpdate != nil {
		c.App.AutoUpdate = *p.AutoUpdate
	}
	if p.SkipDownloaded != nil {
		c.Download.SkipDownloaded = *p.SkipDownloaded
	}
	if p.Grouping != nil {
		c.Download.Grouping = *p.Grouping
	}
	if p.Concurrency != nil {
		c.Download.Concurrency = *p.Concurrency
	}
	if p.Sleep != nil {
		c.Download.SleepSeconds = *p.Sleep
	}
	if p.ParseConcurrency != nil {
		c.Bilibili.ParseConcurrency = *p.ParseConcurrency
	}
	if p.ParseSleep != nil {
		c.Bilibili.ParseSleepSeconds = *p.ParseSleep
	}
	if p.Filename != nil {
		c.Download.FilenameTemplate = strings.TrimSpace(c.Download.FilenameTemplate)
	}
}
func Load(store Store) Data {
	if store == nil {
		return Defaults()
	}
	return FromConfig(store.Config())
}
func Save(store Store, p Changes) error {
	if store == nil {
		return errors.New("配置存储尚未初始化")
	}
	return store.UpdateConfig(func(c *storage.Config) error { p.apply(c); return nil })
}

func (d *Data) Apply(p Changes) error {
	check := storage.DefaultConfig("")
	p.apply(&check)
	if err := check.Validate(); err != nil {
		return err
	}
	c := storage.DefaultConfig("")
	c.Download.Directory = d.Directory
	c.App.Theme = d.Theme
	c.Download.FilenameTemplate = d.Filename
	c.Download.GroupingRule = d.GroupingRule
	c.Download.FFmpegPath = d.FFmpeg
	c.App.AutoUpdate = d.AutoUpdate
	c.Download.SkipDownloaded = d.SkipDownloaded
	c.Download.Grouping = d.Grouping
	c.Download.Concurrency = int(d.Concurrency)
	c.Download.SleepSeconds = int64(d.Sleep)
	c.Bilibili.ParseConcurrency = int(d.ParseConcurrency)
	c.Bilibili.ParseSleepSeconds = int(d.ParseSleep)
	p.apply(&c)
	*d = FromConfig(c)
	return nil
}

// Restore returns failed immediate edits, excluding edits superseded while saving.
func (p Changes) Restore(before Data, pending Changes) Changes {
	r := Changes{}
	if p.Directory != nil && pending.Directory == nil {
		r.Directory = new(before.Directory)
	}
	if p.Theme != nil && pending.Theme == nil {
		r.Theme = new(before.Theme)
	}
	if p.FFmpeg != nil && pending.FFmpeg == nil {
		r.FFmpeg = new(before.FFmpeg)
	}
	if p.AutoUpdate != nil && pending.AutoUpdate == nil {
		r.AutoUpdate = new(before.AutoUpdate)
	}
	if p.SkipDownloaded != nil && pending.SkipDownloaded == nil {
		r.SkipDownloaded = new(before.SkipDownloaded)
	}
	if p.Concurrency != nil && pending.Concurrency == nil {
		r.Concurrency = new(int(before.Concurrency))
	}
	if p.Sleep != nil && pending.Sleep == nil {
		r.Sleep = new(int64(before.Sleep))
	}
	if p.ParseConcurrency != nil && pending.ParseConcurrency == nil {
		r.ParseConcurrency = new(int(before.ParseConcurrency))
	}
	if p.ParseSleep != nil && pending.ParseSleep == nil {
		r.ParseSleep = new(int(before.ParseSleep))
	}

	return r
}
