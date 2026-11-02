package updater

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
)

func TestUpdaterCallbacksWorkWithoutDesktopRuntime(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.UpdateConfig(func(c *storage.Config) error { c.App.AutoUpdate = true; return nil }); err != nil {
		t.Fatal(err)
	}
	previous := Version
	Version = "v1.0.0"
	defer func() { Version = previous }()
	u := New(logger.New(), store)
	calls := 0
	u.client.Transport.WrapRoundTripFunc(func(http.RoundTripper) req.HttpRoundTripFunc {
		return func(r *http.Request) (*http.Response, error) {
			calls++
			body := `{"tag_name":"v2.0.0","body":"native release"}`
			contentType := "application/json"
			if r.URL.Path == "/fixture.tar.gz" {
				body = "update archive fixture"
				contentType = "application/octet-stream"
			}
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: int64(len(body)), Request: r}, nil
		}
	})
	var available Result
	u.OnAvailable = func(r Result) { available = r }
	u.checkOnStartup(context.Background())
	if !available.Available || available.LatestVersion != "v2.0.0" || calls != 1 {
		t.Fatalf("missing update callback: %+v calls=%d", available, calls)
	}
	downloaded, total := int64(0), int64(0)
	u.OnProgress = func(n, size int64) { downloaded, total = n, size }
	path, err := u.DownloadUpdate(context.Background(), "https://fixture.test/fixture.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	body, err := os.ReadFile(path)
	if err != nil || string(body) != "update archive fixture" || downloaded != int64(len(body)) || total != int64(len(body)) {
		t.Fatalf("streamed update or progress failed: %q %d/%d %v", body, downloaded, total, err)
	}
	if err := store.UpdateConfig(func(c *storage.Config) error { c.App.AutoUpdate = false; return nil }); err != nil {
		t.Fatal(err)
	}
	u.checkOnStartup(context.Background())
	if calls != 2 {
		t.Fatal("disabled update check made a request")
	}
	if err := u.InstallUpdate(path); err == nil {
		t.Fatal("unconfigured host quit must not start installer")
	}
}

func TestCurrentVersionAndReleaseComparison(t *testing.T) {
	previous := Version
	defer func() { Version = previous }()
	for _, test := range []struct {
		version, normalized string
		newer               bool
	}{{"v1.0.0", "v1.0.0", true}, {"1.0.0", "v1.0.0", true}, {"v2.0.0", "v2.0.0", false}, {"v3.0.0", "v3.0.0", false}, {"", "dev", true}, {"dev", "dev", true}} {
		Version = test.version
		if got := CurrentVersion(); got != test.normalized || newerRelease("v2.0.0") != test.newer {
			t.Fatalf("version normalization or comparison failed for %q", test.version)
		}
	}
}
func TestCancelledNativeCheckDoesNotRequest(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	u := New(logger.New(), store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := u.Check(ctx); err != context.Canceled {
		t.Fatalf("cancelled check: %v", err)
	}
}
