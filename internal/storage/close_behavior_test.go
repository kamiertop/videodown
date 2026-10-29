package storage

import (
	"path/filepath"
	"testing"
)

func TestClosePreferencePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "videodown.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	enabled, configured, err := store.CloseToTray()
	if err != nil || configured || !enabled {
		t.Fatalf("missing choice: %v %v %v", enabled, configured, err)
	}
	if err := store.SetCloseToTray(false); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	enabled, configured, err = store.CloseToTray()
	if err != nil || !configured || enabled {
		t.Fatalf("saved false lost after reopening: %v %v %v", enabled, configured, err)
	}
	snapshot := store.Config()
	*snapshot.App.CloseToTray = true
	if enabled, configured, err := store.CloseToTray(); err != nil || !configured || enabled {
		t.Fatal("snapshot mutation changed persisted preference")
	}
}
