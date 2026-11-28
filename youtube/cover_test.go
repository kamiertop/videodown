package youtube

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/internal/storage"
)

func TestDownloadCoverSavesOriginalAndRecordsHistory(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	dir := t.TempDir()
	if err := store.UpdateConfig(func(c *storage.Config) error {
		c.Download.Directory = dir
		c.Download.GroupingRule = "author_source"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	service := New(store)
	if err := service.SetCookie("SID=fixture"); err != nil {
		t.Fatal(err)
	}

	var fixture bytes.Buffer
	if err := png.Encode(&fixture, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	const id = "jNQXAC9IVRw"
	const fallbackPath = "/vi/" + id + "/hqdefault.jpg"
	var requests int
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.URL.Hostname() != "i.ytimg.com" || r.URL.Path != fallbackPath || r.Header.Get("Cookie") != "" {
			t.Errorf("unexpected cover request %s, cookie present=%t", r.URL, r.Header.Get("Cookie") != "")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(fixture.Bytes())), Header: make(http.Header), Request: r}, nil
	})
	video := Video{ID: id, Title: "测试视频", Author: "测试作者", Cover: "https://example.com/invalid"}
	for _, name := range []string{"测试视频.png", "测试视频(1).png"} {
		want := filepath.Join(dir, "测试作者", "YouTube", name)
		path, err := service.DownloadCover(t.Context(), video)
		if err != nil || path != want {
			t.Fatalf("DownloadCover() = %q, %v; want %q", path, err, want)
		}
		data, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(data, fixture.Bytes()) {
			t.Fatalf("saved cover differs from original: %v", err)
		}
	}
	history, err := service.History()
	if err != nil || len(history) != 1 || history[0].Kind != "cover" || history[0].ID != id || filepath.Base(history[0].Path) != "测试视频(1).png" {
		t.Fatalf("History() = %+v, %v; want latest cover record", history, err)
	}
	if requests != 2 {
		t.Errorf("cover requests = %d, want 2", requests)
	}
}
