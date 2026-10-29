package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/storage"
)

func TestPortableDefaultDataDirectory(t *testing.T) {
	for _, tc := range []struct{ exe, cwd, temp, want string }{
		{"/opt/videodown/videodown", "/unrelated", "/tmp", "/opt/videodown"},
		{"/tmp/go-build123/b001/exe/videodown", "/work/project", "/tmp", "/work/project"},
		{"/tmp/portable/videodown", "/elsewhere", "/tmp", "/tmp/portable"},
	} {
		if got := defaultDataDir(filepath.FromSlash(tc.exe), filepath.FromSlash(tc.cwd), filepath.FromSlash(tc.temp)); got != filepath.FromSlash(tc.want) {
			t.Fatalf("data directory: got %s want %s", got, tc.want)
		}
	}
}
func TestFirstLaunchCreatesFilesAndPreservesExistingData(t *testing.T) {
	dir := t.TempDir()
	store, err := openAppData(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"config.toml", "history.sqlite", "appicon.png"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("first launch missing %s: %v", name, err)
		}
	}
	icon, err := os.ReadFile(filepath.Join(dir, "appicon.png"))
	if err != nil || !bytes.Equal(icon, assets.AppIconPNG) {
		t.Fatal("bundled icon not exported")
	}
	if got := store.DownloadDirectory(); got != filepath.Join(dir, "download") {
		t.Fatalf("default download escaped application folder: %s", got)
	}
	if err := store.UpdateConfig(func(c *storage.Config) error { c.App.Theme = "dark"; c.Douyin.Cookie = "fixture=test"; return nil }); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.toml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte("\n# retain hand-written comments\n")...)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	store, err = openAppData(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	reopened, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, reopened) {
		t.Fatal("startup rewrote existing config")
	}
	if store.Config().App.Theme != "dark" || store.Config().Douyin.Cookie != "fixture=test" {
		t.Fatal("startup lost preferences or login")
	}
}
func TestPortableFolderCanMove(t *testing.T) {
	base := t.TempDir()
	before := filepath.Join(base, "before")
	after := filepath.Join(base, "after")
	store, err := openAppData(before)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(before, after); err != nil {
		t.Fatal(err)
	}
	store, err = openAppData(after)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if store.DownloadDirectory() != filepath.Join(after, "download") || store.Config().Download.Directory != "./download" {
		t.Fatal("moving folder retained old download path")
	}
}
