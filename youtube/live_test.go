package youtube

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/youtube/internal/extractor"
	"github.com/pelletier/go-toml/v2"
)

// Live checks run only when explicitly enabled through VIDEODOWN_YOUTUBE_*.
// Saved credentials are read only when a configuration path is supplied.
func TestLivePlaylist(t *testing.T) {
	input := os.Getenv("VIDEODOWN_YOUTUBE_PLAYLIST")
	if input == "" {
		t.Skip("requires explicit playlist URL")
	}
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	page, err := service.Playlist(ctx, input, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Videos) == 0 {
		t.Fatal("empty public playlist")
	}
	t.Logf("playlist=%s entries=%d continuation=%v", page.Title, len(page.Videos), page.Continuation != "")
	if os.Getenv("VIDEODOWN_YOUTUBE_PLAYLIST_COVER") == "1" {
		bitmap, err := imagecache.Load(ctx, page.Videos[0].Cover, "https://www.youtube.com/")
		if err != nil || bitmap == nil {
			t.Fatalf("playlist thumbnail: %v", err)
		}
		t.Log("playlist thumbnail decoded")
	}
	if page.Continuation != "" {
		next, err := service.Playlist(ctx, input, page.Continuation)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("next entries=%d continuation=%v", len(next.Videos), next.Continuation != "")
	}
}

func TestLiveYouTube(t *testing.T) {
	if os.Getenv("VIDEODOWN_YOUTUBE_LIVE") != "1" {
		t.Skip("requires network")
	}
	store, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	video, err := service.Resolve(ctx, "https://www.youtube.com/watch?v=jNQXAC9IVRw")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("title=%s formats=%d", video.Title, len(video.Formats))
	dir := t.TempDir()
	if err := store.UpdateConfig(func(c *storage.Config) error { c.Download.Directory = dir; c.Download.Grouping = false; return nil }); err != nil {
		t.Fatal(err)
	}
	format := video.Formats[0]
	for _, f := range video.Formats {
		if f.ID > 0 && f.Height < format.Height {
			format = f
		}
	}

	path, err := service.Download(ctx, video, format.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(path)
	if err != nil || stat.Size() == 0 {
		t.Fatal("empty media output")
	}
	t.Logf("downloaded %s: %d bytes", format.Label, stat.Size())

}

func TestLiveYouTubeResolveURL(t *testing.T) {
	input := os.Getenv("VIDEODOWN_YOUTUBE_URL")
	if input == "" {
		t.Skip("requires explicit network URL")
	}
	store, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if path := os.Getenv("VIDEODOWN_YOUTUBE_CONFIG"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var config storage.Config
		if err := toml.Unmarshal(data, &config); err != nil {
			t.Fatal(err)
		}
		if err := store.UpdateConfig(func(c *storage.Config) error { c.YouTube.Cookie = config.YouTube.Cookie; return nil }); err != nil {
			t.Fatal(err)
		}
	}
	limit := 60 * time.Second
	if os.Getenv("VIDEODOWN_YOUTUBE_BENCH") == "1" {
		limit = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	service := New(store)
	if os.Getenv("VIDEODOWN_YOUTUBE_PERF") == "1" {
		start := time.Now()
		video, err := service.Resolve(ctx, input)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("cold parse: %s", time.Since(start))
		start = time.Now()
		if _, err := service.Resolve(ctx, input); err != nil {
			t.Fatal(err)
		}
		t.Logf("cached parse: %s", time.Since(start))
		start = time.Now()
		bitmap, err := imagecache.Load(ctx, video.Cover, "https://www.youtube.com/")
		if err != nil {
			t.Fatal(err)
		}
		if bitmap == nil {
			t.Fatal("cover not decoded")
		}
		t.Logf("cover decoded: %s", time.Since(start))
		return
	}
	id, idErr := VideoID(input)
	if idErr != nil {
		t.Fatal(idErr)
	}
	client, raw, err := service.fetchVideo(ctx, id)
	loggedIn, known := false, false
	if client != nil {
		loggedIn, known = client.AccountState()
	}
	t.Logf("YouTube account state known=%v loggedIn=%v", known, loggedIn)
	if err != nil {
		t.Fatal(service.parseError(err))
	}
	video := describe(raw)
	t.Logf("title=%s formats=%d", video.Title, len(video.Formats))
	if os.Getenv("VIDEODOWN_YOUTUBE_PROBE") == "1" || os.Getenv("VIDEODOWN_YOUTUBE_BENCH") == "1" {
		if len(video.Formats) == 0 {
			t.Fatal("no usable formats")
		}
		choice := video.Formats[0]
		var stream *extractor.Format
		for i := range raw.Formats {
			if raw.Formats[i].ItagNo == choice.ID {
				stream = &raw.Formats[i]
				break
			}
		}
		media, err := client.GetStreamURLContext(ctx, raw, stream)
		if err != nil {
			t.Fatal(err)
		}
		if os.Getenv("VIDEODOWN_YOUTUBE_BENCH") == "1" {
			base, err := url.Parse(media)
			if err != nil {
				t.Fatal(err)
			}
			total := min(stream.ContentLength, 16<<20)
			if total < 8<<20 {
				t.Skip("video too small for parallel benchmark")
			}
			file, err := os.CreateTemp(t.TempDir(), "bench")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			transport := service.transport
			if transport == nil {
				transport = mediaTransport
			}
			httpClient := &http.Client{Timeout: 0, Transport: transport}
			single := *base
			q := single.Query()
			q.Set("range", fmt.Sprintf("0-%d", total-1))
			single.RawQuery = q.Encode()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, single.String(), nil)
			req.Header.Set("User-Agent", client.UserAgent())
			start := time.Now()
			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 && resp.StatusCode != 206 {
				resp.Body.Close()
				t.Fatalf("baseline HTTP %d", resp.StatusCode)
			}
			_, err = io.CopyN(io.Discard, resp.Body, total)
			resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			elapsed := time.Since(start)
			t.Logf("single stream: %.2f MB/s (%s)", float64(total)/(1<<20)/elapsed.Seconds(), elapsed)
			start = time.Now()
			if err := downloadRanges(ctx, httpClient, base, client.UserAgent(), file, total, func(int64, int64) {}); err != nil {
				t.Fatal(err)
			}
			elapsed = time.Since(start)
			t.Logf("parallel ranges: %.2f MB/s (%s), sample=%d MB", float64(total)/(1<<20)/elapsed.Seconds(), elapsed, total/(1<<20))
			return
		}
		u, err := url.Parse(media)
		if err != nil {
			t.Fatal(err)
		}
		q := u.Query()
		q.Set("range", "0-1023")
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("User-Agent", client.UserAgent())
		resp, err := client.HTTPClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 && resp.StatusCode != 206 {
			t.Fatalf("media HTTP %d", resp.StatusCode)
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if err != nil || len(data) == 0 {
			t.Fatal("empty media")
		}
		t.Logf("media readable: %d bytes, %s", len(data), choice.Label)
	}

}
