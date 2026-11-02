package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTOMLAndSQLitePersistIndependently(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data")
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateConfig(func(c *Config) error { c.App.Theme = "dark"; return nil }); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateConfig(func(c *Config) error {
		c.Bilibili.Auth.Cookie = "SESSDATA=secret"
		c.Bilibili.Auth.CSRF = "secret"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	row := `{"bvid":"BV1","cid":42,"title":"video","upperName":"author","downloaded":"2026-10-28T00:00:00Z"}`
	if err := s.SaveHistory("bilibili:downloaded:42", row); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveHistory("bilibili:downloaded:audio:42", row); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveHistory("bilibili:downloaded:cover:42", row); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveHistory("douyin:downloaded:99", `{"awemeId":"99","title":"other"}`); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(config), "dark") || !strings.Contains(string(config), "SESSDATA=secret") || strings.Contains(string(config), "video") {
		t.Fatal("TOML must contain settings and login data, but not video history")
	}
	if _, err := os.Stat(filepath.Join(dir, "history.sqlite")); err != nil {
		t.Fatal(err)
	}
	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.Config().App.Theme != "dark" {
		t.Fatal("theme not persisted")
	}
	if s.Config().Bilibili.Auth.Cookie != "SESSDATA=secret" {
		t.Fatal("cookie not persisted")
	}
	rows, err := s.HistoryRows("bilibili")
	if err != nil || len(rows) != 3 {
		t.Fatalf("history: %v %v", rows, err)
	}
	videos, err := s.BilibiliVideoRows()
	if err != nil || len(videos) != 1 || videos[0] != row {
		t.Fatalf("video rows: %v %v", videos, err)
	}
	if err := s.DeleteHistoryPrefix("bilibili:downloaded:"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.History("bilibili:downloaded:42"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete history: %v", err)
	}
	if v := s.Config().Bilibili.Auth.Cookie; v == "" {
		t.Fatalf("delete history touched secret: %q %v", v, err)
	}
	if _, err := s.History("douyin:downloaded:99"); err != nil {
		t.Fatalf("delete history touched douyin: %v", err)
	}
}

func TestFirstOpenCreatesConfigurationWithSpecialDirectoryName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "portable # 文件夹")
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Fatal("first open did not create config.toml")
	}
	key := "bilibili:downloaded:BVfixture"
	if err := s.SaveHistory(key, `{"bvid":"BVfixture","title":"fixture"}`); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.History(key); err != nil {
		t.Fatal("history did not persist in special-character directory")
	}
}
