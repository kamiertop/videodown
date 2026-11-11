package preferences

import (
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/internal/storage"
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
	values := Changes{Theme: new("cupcake"), Concurrency: new(4), Sleep: new(int64(120)), ParseConcurrency: new(2), ParseSleep: new(10), SkipDownloaded: new(false), AutoUpdate: new(false), Filename: new(" {author}_{title} "), GroupingRule: new("date"), Grouping: new(true)}
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
	data := Load(store)
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
	for _, values := range []Changes{
		{Theme: new("dark"), Concurrency: new(0)},
		{Filename: new(""), GroupingRule: new("author")},
		{Filename: new("literal without variables")},
		{ParseConcurrency: new(6)}, {Sleep: new(int64(3601))},
	} {
		if err := Save(store, values); err == nil {
			t.Fatalf("invalid values accepted: %v", values)
		}
	}
	data := Load(store)
	if data.Theme != "light" || data.Filename != "{title}" || data.GroupingRule != "author_source" {
		t.Fatal("failed validation partially wrote other settings")
	}
}

func TestPartialChangesPreserveCredentialsAndMergeLatestValue(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.UpdateConfig(func(c *storage.Config) error {
		c.Douyin.Cookie = "fixture"
		c.Bilibili.Auth.Cookie = "fixture"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	first := Changes{Sleep: new(int64(10))}
	merged := first.Merge(Changes{Sleep: new(int64(30)), AutoUpdate: new(false)})
	if err := Save(store, merged); err != nil {
		t.Fatal(err)
	}
	c := store.Config()
	if c.Download.SleepSeconds != 30 || c.App.AutoUpdate || c.Douyin.Cookie != "fixture" || c.Bilibili.Auth.Cookie != "fixture" {
		t.Fatal("partial edit replaced unrelated data")
	}
	if *first.Sleep != 10 {
		t.Fatal("merge mutated in-flight change")
	}
}

func TestRestoreDoesNotDiscardPendingOrUnsavedEdits(t *testing.T) {
	before := Defaults()
	current := before
	current.Filename = ""
	current.Sleep = 30
	current.Theme = "dark"
	failed := Changes{Sleep: new(int64(10)), Theme: new("dark")}
	restore := failed.Restore(before, Changes{Sleep: new(int64(30))})
	if err := current.Apply(restore); err != nil {
		t.Fatal(err)
	}
	if current.Sleep != 30 || current.Theme != before.Theme || current.Filename != "" {
		t.Fatal("rollback discarded pending edit or unsaved template")
	}
}
