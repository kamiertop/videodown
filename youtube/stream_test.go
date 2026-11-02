package youtube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallelRangesWriteInOrderAndReportMonotonicProgress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var active, peak atomic.Int32
	var release sync.Once
	ready := make(chan struct{})
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for {
			old := peak.Load()
			if old >= n || peak.CompareAndSwap(old, n) {
				break
			}
		}
		if n == 4 {
			release.Do(func() { close(ready) })
		}
		select {
		case <-ready:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		var start, end int64
		if _, err := fmt.Sscanf(r.URL.Query().Get("range"), "%d-%d", &start, &end); err != nil {
			return nil, err
		}
		data := bytes.Repeat([]byte{byte(start / streamChunkSize)}, int(end-start+1))
		return &http.Response{StatusCode: 200, ContentLength: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
	})}
	file, err := os.CreateTemp(t.TempDir(), "parts")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	base, _ := url.Parse("https://fixture.googlevideo.com/media")
	total := 4 * streamChunkSize
	var last int64
	if err := downloadRanges(ctx, client, base, "fixture", file, total, func(done, size int64) {
		if done < last || done > total || size != total {
			t.Error("invalid progress")
		}
		last = done
	}); err != nil {
		t.Fatal(err)
	}
	if peak.Load() < 2 || last != total {
		t.Fatal("download remained serial or incomplete")
	}
	data, err := os.ReadFile(file.Name())
	if err != nil || int64(len(data)) != total {
		t.Fatal("bad file size")
	}
	for i, value := range data {
		if value != byte(int64(i)/streamChunkSize) {
			t.Fatalf("out-of-order chunk at %d", i)
		}
	}
}
func TestRangeRetryDoesNotDoubleCountOrCorruptFile(t *testing.T) {
	requests := 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		data := "abcd"
		if requests == 1 {
			data = "ab"
		}
		return &http.Response{StatusCode: 200, ContentLength: 4, Body: io.NopCloser(bytes.NewBufferString(data)), Header: make(http.Header)}, nil
	})}
	file, err := os.CreateTemp(t.TempDir(), "retry")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	base, _ := url.Parse("https://fixture.googlevideo.com/media")
	var done int64
	if err := downloadRanges(context.Background(), client, base, "fixture", file, 4, func(n, _ int64) {
		done = n
		if n > 4 {
			t.Error("retry double-counted bytes")
		}
	}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(file.Name())
	if string(data) != "abcd" || requests != 2 || done != 4 {
		t.Fatalf("bad retry: %q requests=%d progress=%d", data, requests, done)
	}
}
