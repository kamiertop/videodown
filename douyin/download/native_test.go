package download

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kamiertop/videodown/internal/storage"
)

type nativeTransport func(*http.Request) (*http.Response, error)

func (f nativeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNativeDownloadsUseConfiguredWorkersAndPublishProgress(t *testing.T) {
	s := newTestService(t)
	if err := s.store.UpdateConfig(func(c *storage.Config) error {
		c.Download.Directory = t.TempDir()
		c.Download.GroupingRule = "none"
		c.Download.Concurrency = 3
		c.Download.SleepSeconds = 0
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	started := make(chan string, 5)
	release := make(chan struct{})
	finished := make(chan BatchResult, 1)
	failures := make(chan error, 1)
	var mu sync.Mutex
	ids := map[string]bool{}
	s.SetProgressListener(func(p Progress) {
		if p.Phase == "done" {
			mu.Lock()
			ids[p.AwemeID] = true
			mu.Unlock()
		}
	})
	s.client.GetClient().Transport = nativeTransport(func(r *http.Request) (*http.Response, error) {
		started <- r.URL.Path
		<-release
		return &http.Response{StatusCode: 200, Status: "200 OK", Header: http.Header{"Content-Type": {"video/mp4"}}, Body: io.NopCloser(strings.NewReader("fixture-video")), ContentLength: 13, Request: r}, nil
	})
	tasks := []Task{}
	for i := range 5 {
		tasks = append(tasks, Task{AwemeID: fmt.Sprint(i), Title: fmt.Sprint(i), VideoURL: fmt.Sprintf("https://fixture.test/%d.mp4", i)})
	}
	go func() { result, err := s.DownloadVideos(tasks); failures <- err; finished <- result }()
	// Release blocked requests even when an assertion fails.
	released := false
	defer func() {
		if !released {
			close(release)
			<-finished
		}
	}()
	for range 3 {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("three workers did not start")
		}
	}
	select {
	case <-started:
		t.Fatal("exceeded concurrency limit")
	default:
	}
	if s.ActiveCount() != 3 {
		t.Fatalf("active count: %d", s.ActiveCount())
	}
	close(release)
	released = true
	select {
	case result := <-finished:
		if err := <-failures; err != nil || result.Success != 5 {
			t.Fatalf("download result: %+v %v", result, err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("downloads failed to finish")
	}
	mu.Lock()
	n := len(ids)
	mu.Unlock()
	if n != 5 || s.ActiveCount() != 0 {
		t.Fatalf("missing final progress or active task leak: %d / %d", n, s.ActiveCount())
	}
	page, err := s.DownloadHistoryPage(0, 20, "")
	if err != nil || page.Total != 5 {
		t.Fatalf("download history: %+v %v", page, err)
	}
}
