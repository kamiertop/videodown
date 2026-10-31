package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/internal/storage"
)

func TestFirstLaunchCreatesFilesAndPreservesExistingData(t *testing.T) {
	dir := t.TempDir()
	store, err := openAppData(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"config.toml", "history.sqlite"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("first launch missing %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "appicon.png")); !os.IsNotExist(err) {
		t.Fatalf("startup should use the embedded icon without exporting it: %v", err)
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

func TestDownloadDirectoryCanBeChangedOutsideApplication(t *testing.T) {
	root := t.TempDir()
	store, err := openAppData(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	outside := t.TempDir()
	if err := store.UpdateConfig(func(c *storage.Config) error { c.Download.Directory = outside; return nil }); err != nil {
		t.Fatal(err)
	}
	if store.DownloadDirectory() != outside {
		t.Fatal("custom download directory ignored")
	}
}

func TestPortableRuntimeIgnoresLaunchingDirectory(t *testing.T) {
	root := t.TempDir()
	t.Chdir(t.TempDir())
	for _, key := range []string{"TMPDIR", "TMP", "TEMP", "XDG_CACHE_HOME", "LOCALAPPDATA", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "APPDATA"} {
		t.Setenv(key, "")
	}
	if err := configurePortableRuntime(root); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	canonical, _ := filepath.EvalSymlinks(root)
	if err != nil || cwd != canonical {
		t.Fatalf("working directory: %s %v", cwd, err)
	}
	if os.TempDir() != filepath.Join(root, "tmp") {
		t.Fatal("temporary files escape portable directory")
	}
	if os.Getenv("XDG_CACHE_HOME") != filepath.Join(root, "cache") {
		t.Fatal("library cache escapes portable directory")
	}
	store, err := openAppData(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if store.DownloadDirectory() != filepath.Join(root, "download") {
		t.Fatal("default download depends on launching directory")
	}
}
