package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestStructuredConfigRoundTripAndDefaults(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.InitDefaults("/downloads"); err != nil {
		t.Fatal(err)
	}
	if err = s.UpdateConfig(func(c *Config) error {
		c.App.Theme = "dark"
		c.App.AutoUpdate = false
		disabled := false
		c.App.CloseToTray = &disabled
		c.Download.Concurrency = 4
		c.Download.SleepSeconds = 0
		c.Download.Grouping = false
		c.Download.SkipDownloaded = false
		c.Bilibili.Auth = BilibiliAuth{Cookie: "test-cookie", CSRF: "test-csrf", UID: 42, RefreshToken: "test-token"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	before := s.Config()
	raw, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[app]", "[download]", "[bilibili.auth]", "concurrency = 4", "skip_downloaded = false", "sleep_seconds = 0"} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("missing typed TOML field %s", want)
		}
	}
	if strings.Contains(string(raw), "[settings]") {
		t.Fatal("configuration still uses the string map format")
	}
	var decoded Config
	if err = toml.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Download.Concurrency != 4 || decoded.Bilibili.Auth.UID != 42 || decoded.Download.Directory != "/downloads" || decoded.App.CloseToTray == nil || *decoded.App.CloseToTray {
		t.Fatal("typed TOML round trip failed")
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	after := s.Config()
	if after.Bilibili.Auth != before.Bilibili.Auth || after.Download != before.Download || after.App.AutoUpdate {
		t.Fatal("restart lost config fields or explicit false/zero values")
	}
}

func TestTypedConfigUpdatesAreAtomicAndIndependent(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	before := s.Config()
	if err = s.UpdateConfig(func(c *Config) error { c.App.Theme = "dark"; c.Download.Concurrency = 0; return nil }); err == nil {
		t.Fatal("invalid config accepted")
	}
	if s.Config().App.Theme != before.App.Theme {
		t.Fatal("failed validation partially committed")
	}
	if err = s.UpdateConfig(func(c *Config) error { c.App.Theme = "dark"; return errors.New("cancel edit") }); err == nil || s.Config().App.Theme != before.App.Theme {
		t.Fatal("aborted update committed")
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := s.UpdateConfig(func(c *Config) error { c.Bilibili.DefaultQuality++; return nil }); err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := s.UpdateConfig(func(c *Config) error { c.Bilibili.Auth.UID++; return nil }); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if c := s.Config(); c.Bilibili.DefaultQuality != 20 || c.Bilibili.Auth.UID != 20 {
		t.Fatal("concurrent updates lost fields")
	}
	s.configPath = filepath.Join(t.TempDir(), "missing", "config.toml")
	if err = s.UpdateConfig(func(c *Config) error { c.App.Theme = "dark"; return nil }); err == nil || s.Config().App.Theme != before.App.Theme {
		t.Fatal("failed disk write changed the live config")
	}
}

func TestStructuredConfigOverlayAndRejectsLegacyMap(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(file, []byte("[app]\ntheme='dark'\n[download]\nsleep_seconds=0\nskip_downloaded=false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	c := s.Config()
	s.Close()
	if !c.App.AutoUpdate || c.Download.Concurrency != 1 || c.Download.SleepSeconds != 0 || c.Download.SkipDownloaded {
		t.Fatal("omitted fields and explicit zero/false were confused")
	}
	legacy := []byte("[settings]\ntheme='dark'\nbilibili_cookies='do-not-log-cookie'\n")
	if err = os.WriteFile(file, legacy, 0600); err != nil {
		t.Fatal(err)
	}
	if s, err = Open(dir); err == nil {
		s.Close()
		t.Fatal("legacy map was silently accepted")
	}
	if !strings.Contains(err.Error(), "settings") || !strings.Contains(err.Error(), "config.example.toml") || strings.Contains(err.Error(), "do-not-log-cookie") {
		t.Fatal("configuration error must identify unknown fields and the example without exposing values")
	}
	raw, err := os.ReadFile(file)
	if err != nil || string(raw) != string(legacy) {
		t.Fatal("invalid config was overwritten")
	}
}
