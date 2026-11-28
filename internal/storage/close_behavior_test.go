package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
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
	snapshot.App.CloseAction = CloseActionTray
	if enabled, configured, err := store.CloseToTray(); err != nil || !configured || enabled {
		t.Fatal("snapshot mutation changed persisted preference")
	}
}

func TestClosePreferenceAcceptsExistingConfig(t *testing.T) {
	for _, tc := range []struct {
		name       string
		line       string
		enabled    bool
		configured bool
	}{
		{name: "unset", enabled: true},
		{name: "exit", line: "close_to_tray = false\n", configured: true},
		{name: "tray", line: "close_to_tray = true\n", enabled: true, configured: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.toml")
			if err := os.WriteFile(path, []byte("[app]\ntheme = \"dark\"\n"+tc.line), 0o600); err != nil {
				t.Fatal(err)
			}
			store, err := Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			if enabled, configured, err := store.CloseToTray(); err != nil || enabled != tc.enabled || configured != tc.configured {
				t.Fatalf("CloseToTray() = %t, %t, %v; want %t, %t, nil", enabled, configured, err, tc.enabled, tc.configured)
			}
			if err := store.UpdateConfig(func(c *Config) error {
				c.App.Theme = "light"
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var legacy struct {
				App struct {
					Theme       string `toml:"theme"`
					CloseToTray *bool  `toml:"close_to_tray"`
				} `toml:"app"`
			}
			if err := toml.Unmarshal(raw, &legacy); err != nil {
				t.Fatalf("read updated config with legacy fields: %v", err)
			}
			choice := legacy.App.CloseToTray
			if legacy.App.Theme != "light" || (choice != nil) != tc.configured || choice != nil && *choice != tc.enabled {
				t.Fatalf("unexpected [app] after saving %s: %s", tc.name, raw)
			}
		})
	}
}
