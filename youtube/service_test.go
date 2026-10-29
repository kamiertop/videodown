package youtube

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/youtube/internal/extractor"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestVideoID(t *testing.T) {
	for _, input := range []string{"jNQXAC9IVRw", "https://www.youtube.com/watch?v=jNQXAC9IVRw&list=x", "https://youtu.be/jNQXAC9IVRw", "https://www.youtube.com/shorts/jNQXAC9IVRw"} {
		id, err := VideoID(input)
		if err != nil || id != "jNQXAC9IVRw" {
			t.Fatalf("%s: %s %v", input, id, err)
		}
	}
	for _, input := range []string{"https://evil.test/watch?v=jNQXAC9IVRw", "https://youtube.com.evil.test/watch?v=jNQXAC9IVRw", "https://youtube.com/playlist?list=x", "bad"} {
		if _, err := VideoID(input); err == nil {
			t.Fatalf("accepted %s", input)
		}
	}
}
func TestDownloadCookieHistoryAndFailureCleanup(t *testing.T) {
	store, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	dir := t.TempDir()
	if err := store.UpdateConfig(func(c *storage.Config) error { c.Download.Directory = dir; c.Download.Grouping = false; return nil }); err != nil {
		t.Fatal(err)
	}
	service := New(store)
	if err := service.SetCookie("Cookie: SID=fixture"); err != nil {
		t.Fatal(err)
	}
	requests := 0
	metadataRequests := 0
	fail := false
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if err := r.Context().Err(); err != nil {
			return nil, err
		}
		data := ""
		status := 200
		switch r.URL.Hostname() {
		case "youtube.com":
			data = `<script src="/s/player/test/player_ias/base.js"></script>`
		case "www.youtube.com":
			if strings.Contains(r.URL.Path, "/youtubei/") {
				metadataRequests++
			}
			if strings.Contains(r.URL.Path, "base.js") {
				data = "signatureTimestamp:12345"
				break
			}
			if !strings.Contains(r.Header.Get("Cookie"), "SID=fixture") {
				t.Error("missing cookie")
			}
			data = `{"playabilityStatus":{"status":"OK"},"videoDetails":{"videoId":"jNQXAC9IVRw","title":"fixture","author":"author","lengthSeconds":"5"},"streamingData":{"formats":[{"itag":18,"url":"https://fixture.googlevideo.com/video","mimeType":"video/mp4","height":360,"audioChannels":2,"contentLength":"4"}]}}`
		case "fixture.googlevideo.com":
			if _, deadline := r.Context().Deadline(); deadline {
				t.Error("media request has an overall download timeout")
			}
			requests++
			if r.Header.Get("Cookie") != "" {
				t.Error("cookie leaked to media host")
			}
			data = "data"
			if fail {
				status = 403
				data = "denied"
			}
		default:
			t.Fatalf("unexpected request %s", r.URL)
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(data)), Header: make(http.Header), ContentLength: int64(len(data)), Request: r}, nil
	})
	video, err := service.Resolve(context.Background(), "jNQXAC9IVRw")
	if err != nil || len(video.Formats) != 1 {
		t.Fatalf("parse: %+v %v", video, err)
	}
	path, err := service.Download(context.Background(), video, 18, nil)
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "data" {
		t.Fatal("bad output")
	}
	if metadataRequests != 1 {
		t.Fatalf("download repeated video parsing: %d", metadataRequests)
	}
	history, err := service.History()
	if err != nil || len(history) != 1 || history[0].Path != path {
		t.Fatalf("history: %+v %v", history, err)
	}
	if _, err := service.Download(context.Background(), video, 18, nil); err != nil || requests != 1 {
		t.Fatal("incremental download failed")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	fail = true
	if _, err := service.Download(context.Background(), video, 18, nil); err == nil {
		t.Fatal("403 accepted")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("partial files leaked: %v", entries)
	}
	if err := service.SetCookie("SID=fixture\nBad: header"); err == nil {
		t.Fatal("newline accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Download(ctx, video, 18, nil); err == nil {
		t.Fatal("canceled download succeeded")
	}
	entries, _ = os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatal("canceled download left partial file")
	}
	if service.Cookie() != "SID=fixture" {
		t.Fatal("failed save changed cookie")
	}
}
func TestFormatLabelsExplainVariantsWithoutMerging(t *testing.T) {
	v := describe(&extractor.Video{Formats: extractor.FormatList{
		{ItagNo: 137, Height: 1080, FPS: 30, Bitrate: 2500000, MimeType: `video/mp4; codecs="avc1.640028"`, URL: "fixture"},
		{ItagNo: 399, Height: 1080, FPS: 60, Bitrate: 1800000, MimeType: `video/mp4; codecs="av01.0.08M.08"`, URL: "fixture"},
		{ItagNo: 248, Height: 1080, FPS: 30, Bitrate: 2100000, MimeType: `video/webm; codecs="vp09.00.40.08"`, URL: "fixture"},
	}})
	if len(v.Formats) != 3 {
		t.Fatal("formats were merged")
	}
	for i, codec := range []string{"H.264", "AV1", "VP9"} {
		label := v.Formats[i].Label
		for _, detail := range []string{"1080p", codec, "Mbps", "帧", "需合并音频"} {
			if !strings.Contains(label, detail) {
				t.Fatalf("missing %s in %s", detail, label)
			}
		}
	}
	if v.Formats[0].Label == v.Formats[1].Label || v.Formats[1].Label == v.Formats[2].Label {
		t.Fatal("indistinguishable formats")
	}
	if label := formatLabel(extractor.Format{ItagNo: 18, Height: 360, AudioChannels: 2, MimeType: `video/mp4; codecs="avc1.42001E, mp4a.40.2"`}); !strings.Contains(label, "含音频") {
		t.Fatal("missing audio distinction")
	}
}

func TestSavedCookieLoginFailureRetriesAnonymous(t *testing.T) {
	store, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store)
	if err := service.SetCookie("SID=fixture"); err != nil {
		t.Fatal(err)
	}
	var cookieRequests, anonymousRequests int
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		data := "{}"
		if strings.Contains(r.URL.Path, "/youtubei/") {
			if strings.Contains(r.Header.Get("Cookie"), "SID=fixture") {
				cookieRequests++
				data = `{"playabilityStatus":{"status":"LOGIN_REQUIRED","reason":"Sign in to confirm you are not a bot"}}`
			} else {
				anonymousRequests++
				data = `{"playabilityStatus":{"status":"OK"},"videoDetails":{"title":"public"},"streamingData":{"formats":[{"itag":18,"url":"https://fixture.googlevideo.com/video","height":360,"mimeType":"video/mp4"}]}}`
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data)), Header: make(http.Header), Request: r}, nil
	})
	if _, err := service.Resolve(context.Background(), "KWGhxyu5osA"); err != nil {
		t.Fatal(err)
	}
	if cookieRequests != 1 || anonymousRequests != 1 || service.Cookie() != "SID=fixture" {
		t.Fatal("fallback did not preserve stored cookie")
	}
	verified := service.parseError(fmt.Errorf("%w: reload", errAuthenticatedPlayback)).Error()
	if !strings.Contains(verified, "Cookie 有效") {
		t.Fatal("valid login was confused with missing cookie")
	}
	reload := service.parseError(&extractor.ErrPlayabiltyStatus{Status: "UNPLAYABLE", Reason: "The page needs to be reloaded."}).Error()
	if !strings.Contains(reload, "已读取保存的 Cookie") || !strings.Contains(reload, "尚未适配") {
		t.Fatal("reload error confused with missing cookie")
	}
	message := service.parseError(fmt.Errorf("%w: bot challenge", extractor.ErrLoginRequired)).Error()
	if strings.Contains(message, "可填写 Cookie") || !strings.Contains(message, "已读取保存") {
		t.Fatalf("misleading saved-cookie message: %s", message)
	}
}
func TestParseCacheExpiresAndCookieChangesInvalidate(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store)
	requests := 0
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		data := "{}"
		if strings.Contains(r.URL.Path, "/youtubei/") {
			requests++
			data = `{"playabilityStatus":{"status":"OK"},"videoDetails":{"title":"fixture"},"streamingData":{"formats":[{"itag":18,"url":"https://fixture.googlevideo.com/video","mimeType":"video/mp4","height":360}]}}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data)), Header: make(http.Header), Request: r}, nil
	})
	for range 2 {
		if _, err := service.Resolve(context.Background(), "KWGhxyu5osA"); err != nil {
			t.Fatal(err)
		}
	}
	if requests != 1 {
		t.Fatal("parse cache missed")
	}
	for key, value := range service.cache {
		value.at = time.Now().Add(-6 * time.Minute)
		service.cache[key] = value
	}
	if _, err := service.Resolve(context.Background(), "KWGhxyu5osA"); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatal("expired cache reused")
	}
	if err := service.SetCookie("SID=new-fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Resolve(context.Background(), "KWGhxyu5osA"); err != nil {
		t.Fatal(err)
	}
	if requests != 3 {
		t.Fatal("cookie change reused old cache")
	}
}
