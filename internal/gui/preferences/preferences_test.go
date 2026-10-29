package preferences

import (
	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/internal/storage"
	"path/filepath"
	"testing"
)

func TestSavedSettingsSurviveRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "videodown.db")
	store, err := storage.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitDefaults("/downloads"); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{constant.ThemeKey: "cupcake", constant.ConcurrencyNumKey: "4", constant.SleepTimeKey: "120", constant.ParsePlayURLNumKey: "2", constant.ParsePlayURLSleepKey: "10", constant.SkipDownloadedKey: "false", constant.AutoUpdateKey: "false", constant.FilenameTemplateKey: " {author}_{title} ", constant.GroupingRuleKey: "date", constant.AllowGroupOnSaveKey: "true"}
	if err := Save(store, values); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = storage.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	data, err := Load(store)
	if err != nil {
		t.Fatal(err)
	}
	if data.Theme != "cupcake" || data.Concurrency != 4 || data.Sleep != 120 || data.ParseSleep != 10 || data.ParseConcurrency != 2 || data.SkipDownloaded || data.AutoUpdate || data.Filename != "{author}_{title}" || data.GroupingRule != "date" || !data.Grouping || data.Directory != "/downloads" {
		t.Fatalf("saved data not restored: %+v", data)
	}
	if _, configured, err := store.CloseToTray(); err != nil || configured {
		t.Fatal("default settings must not create a close preference")
	}
}
func TestInvalidSettingsAreAtomic(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.InitDefaults("/downloads"); err != nil {
		t.Fatal(err)
	}
	for _, values := range []map[string]string{
		{constant.ThemeKey: "dark", constant.ConcurrencyNumKey: "0"},
		{constant.FilenameTemplateKey: "", constant.GroupingRuleKey: "author"},
		{constant.ParsePlayURLNumKey: "6"}, {constant.SleepTimeKey: "3601"},
	} {
		if err := Save(store, values); err == nil {
			t.Fatalf("invalid values accepted: %v", values)
		}
	}
	data, err := Load(store)
	if err != nil {
		t.Fatal(err)
	}
	if data.Theme != "light" || data.Filename != "{title}" || data.GroupingRule != "author_source" {
		t.Fatal("failed validation partially wrote other settings")
	}
}
