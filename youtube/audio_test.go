package youtube

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamiertop/videodown/internal/storage"
)

func TestAudioDownloadKeepsOriginalStreamAndSeparateHistory(t *testing.T) {
	store, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.UpdateConfig(func(c *storage.Config) error {
		c.Download.Directory = t.TempDir()
		c.Download.FFmpegPath = "/missing/ffmpeg"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	s := New(store)
	var mediaPaths []string
	s.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body := "{}"
		if strings.Contains(r.URL.Path, "/youtubei/") {
			body = `{"playabilityStatus":{"status":"OK"},"videoDetails":{"videoId":"jNQXAC9IVRw","title":"fixture","author":"author"},"streamingData":{"formats":[{"itag":18,"url":"https://fixture.googlevideo.com/video","mimeType":"video/mp4","height":360,"audioChannels":2,"contentLength":"4"}],"adaptiveFormats":[{"itag":140,"url":"https://fixture.googlevideo.com/audio","mimeType":"audio/mp4; codecs=\"mp4a.40.2\"","bitrate":128000,"audioChannels":2,"contentLength":"4"},{"itag":251,"url":"https://fixture.googlevideo.com/opus","mimeType":"audio/webm; codecs=\"opus\"","bitrate":160000,"audioChannels":2,"contentLength":"4"}]}}`
		} else if r.URL.Hostname() == "fixture.googlevideo.com" {
			mediaPaths = append(mediaPaths, r.URL.Path)
			body = "data"
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), ContentLength: int64(len(body)), Request: r}, nil
	})
	v, err := s.Resolve(context.Background(), "jNQXAC9IVRw")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Formats) != 1 || len(v.AudioFormats) != 2 || v.AudioFormats[0].ID != 251 || !strings.Contains(v.AudioFormats[1].Label, "AAC") {
		t.Fatalf("bad format separation: %+v", v)
	}
	for _, tc := range []struct {
		format    int
		ext, path string
	}{{140, ".m4a", "/audio"}, {-1, ".webm", "/opus"}, {18, ".mp4", "/video"}} {
		path, err := s.Download(context.Background(), v, tc.format, nil)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Ext(path) != tc.ext || mediaPaths[len(mediaPaths)-1] != tc.path {
			t.Fatalf("wrong stream or extension: %s %v", path, mediaPaths)
		}
		if data, err := os.ReadFile(path); err != nil || string(data) != "data" {
			t.Fatal("audio was not saved intact")
		}
	}
	history, err := s.History()
	if err != nil || len(history) != 3 {
		t.Fatalf("separate history missing: %+v %v", history, err)
	}
	audio := 0
	for _, entry := range history {
		if entry.Kind == "audio" {
			audio++
		}
	}
	if audio != 2 {
		t.Fatal("audio history kind missing")
	}
}

func TestAudioDuplicatesAndDRCChooseMatchingMedia(t *testing.T) {
	store, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.UpdateConfig(func(c *storage.Config) error {
		c.Download.Directory = t.TempDir()
		c.Download.Grouping = false
		c.Download.SkipDownloaded = false
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	stream := func(id, rate int, path string, drc bool) map[string]any {
		mime := `audio/webm; codecs="opus"`
		if id == 140 {
			mime = `audio/mp4; codecs="mp4a.40.2"`
		}
		return map[string]any{"itag": id, "mimeType": mime, "bitrate": rate, "averageBitrate": rate, "audioChannels": 2, "contentLength": "4", "url": "https://fixture.googlevideo.com" + path, "isDrc": drc}
	}
	entries := []map[string]any{stream(140, 129000, "/aac", false), stream(140, 129000, "/aac-copy", false), stream(251, 117000, "/opus117", false), stream(251, 118000, "/opus118", false), stream(251, 120000, "/opus-drc", true), stream(251, 120000, "/opus-drc-copy", true), stream(250, 65000, "/low", false), stream(250, 65000, "/low-copy", false)}
	nonDefault := stream(251, 200000, "/dub", false)
	nonDefault["audioTrack"] = map[string]any{"id": "dub", "displayName": "配音", "audioIsDefault": false}
	entries = append(entries, nonDefault)
	player, _ := json.Marshal(map[string]any{"playabilityStatus": map[string]string{"status": "OK"}, "videoDetails": map[string]string{"title": "fixture", "author": "author"}, "streamingData": map[string]any{"adaptiveFormats": entries}})
	s := New(store)
	var mediaPath string
	s.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body := "{}"
		if strings.Contains(r.URL.Path, "/youtubei/") {
			body = string(player)
		}
		if r.URL.Hostname() == "fixture.googlevideo.com" {
			mediaPath = r.URL.Path
			body = "data"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), ContentLength: int64(len(body)), Request: r}, nil
	})
	v, err := s.Resolve(context.Background(), "jNQXAC9IVRw")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.AudioFormats) != 4 {
		t.Fatalf("duplicate formats remain: %+v", v.AudioFormats)
	}
	seen := map[int]bool{}
	for _, f := range v.AudioFormats {
		if seen[f.ID] {
			t.Fatal("multiple audio options share a selection ID")
		}
		seen[f.ID] = true
	}
	if v.AudioFormats[0].ID != 140 || v.AudioFormats[1].ID != 251 || !strings.Contains(v.AudioFormats[1].Summary, "118 kbps") || !strings.Contains(v.AudioFormats[3].Label, "动态压缩(DRC)") {
		t.Fatalf("wrong rendition description/order: %+v", v.AudioFormats)
	}
	for _, tc := range []struct {
		id   int
		path string
	}{{251, "/opus118"}, {drcAudioIDOffset + 251, "/opus-drc"}, {140, "/aac"}, {250, "/low"}} {
		if _, err := s.Download(context.Background(), v, tc.id, nil); err != nil {
			t.Fatal(err)
		}
		if mediaPath != tc.path {
			t.Fatalf("selected %d downloaded %s, want %s", tc.id, mediaPath, tc.path)
		}
	}
	history, err := s.History()
	if err != nil || len(history) != 4 {
		t.Fatal("rendition history entries collided")
	}
}
