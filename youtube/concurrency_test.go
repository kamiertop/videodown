package youtube

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kamiertop/videodown/internal/storage"
)

func TestConcurrentSameTitleDownloadsKeepBothFiles(t *testing.T) {
	store, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	dir := t.TempDir()
	if err := store.UpdateConfig(func(c *storage.Config) error {
		c.Download.Directory = dir
		c.Download.Grouping = false
		c.Download.SkipDownloaded = false
		c.Download.Concurrency = 2
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	s := New(store)
	if s.DownloadConcurrency() != 2 {
		t.Fatal("service ignored saved concurrency")
	}
	if err := store.UpdateConfig(func(c *storage.Config) error { c.Download.Concurrency = 3; return nil }); err != nil {
		t.Fatal(err)
	}
	if s.DownloadConcurrency() != 3 {
		t.Fatal("service cached an obsolete limit")
	}
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	s.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body := "{}"
		if strings.Contains(r.URL.Path, "/youtubei/") {
			body = `{"playabilityStatus":{"status":"OK"},"videoDetails":{"title":"同一个标题","author":"作者"},"streamingData":{"formats":[{"itag":18,"url":"https://fixture.googlevideo.com/video","height":360,"mimeType":"video/mp4","audioChannels":2,"contentLength":"4"}]}}`
		}
		if r.URL.Hostname() == "fixture.googlevideo.com" {
			started <- struct{}{}
			select {
			case <-release:
			case <-r.Context().Done():
				return nil, r.Context().Err()
			}
			body = "data"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), ContentLength: int64(len(body)), Request: r}, nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type result struct {
		path string
		err  error
	}
	results := make(chan result, 2)
	for _, id := range []string{"jNQXAC9IVRw", "KWGhxyu5osA"} {
		go func() {
			v, err := s.Resolve(ctx, id)
			if err != nil {
				results <- result{err: err}
				return
			}
			path, err := s.Download(ctx, v, 18, nil)
			results <- result{path, err}
		}()
	}
	for range 2 {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("parallel transfer did not start")
		}
	}
	close(release)
	a, b := <-results, <-results
	if a.err != nil || b.err != nil || a.path == b.path {
		t.Fatalf("parallel files collided: %+v %+v", a, b)
	}
	for _, path := range []string{a.path, b.path} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "data" {
			t.Fatal("completed media missing")
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("partial files leaked: %v %v", entries, err)
	}
	history, err := s.History()
	if err != nil || len(history) != 2 {
		t.Fatalf("parallel history missing: %+v %v", history, err)
	}
}
