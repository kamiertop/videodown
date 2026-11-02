package extractor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type playlistTransport func(*http.Request) (*http.Response, error)

func (f playlistTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPlaylistPaginationKeepsOrderAndSkipsUnavailable(t *testing.T) {
	initial := `{"metadata":{"playlistMetadataRenderer":{"title":"课程列表"}},"contents":{"twoColumnBrowseResultsRenderer":{"tabs":[{"tabRenderer":{"content":{"sectionListRenderer":{"contents":[{"itemSectionRenderer":{"contents":[{"playlistVideoListRenderer":{"contents":[{"playlistVideoRenderer":{"videoId":"jNQXAC9IVRw","title":{"runs":[{"text":"第一课"}]},"lengthSeconds":"90","shortBylineText":{"runs":[{"text":"作者"}]}}},{"playlistVideoRenderer":{"videoId":"private0001","title":{"simpleText":"[Private video]"},"isPlayable":false}},{"continuationItemRenderer":{"continuationEndpoint":{"continuationCommand":{"token":"page2"}}}}]}}]}}]}}}}]}}}`
	next := `{"onResponseReceivedActions":[{"appendContinuationItemsAction":{"continuationItems":[{"playlistVideoRenderer":{"videoId":"KWGhxyu5osA","title":{"simpleText":"第二课"},"isPlayable":true}}]}}]}`
	posts := 0
	client := &Client{Cookie: "SID=fixture", HTTPClient: &http.Client{Transport: playlistTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Hostname() != "www.youtube.com" {
			t.Fatal("unexpected host")
		}
		body := `<script>ytcfg.set({"VISITOR_DATA":"visitor","INNERTUBE_CLIENT_VERSION":"2.fixture"}); var ytInitialData = ` + initial + `;</script>`
		if r.Method == http.MethodPost {
			posts++
			var request innertubeRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request.Continuation != "page2" || request.Context.Client.ClientName != "WEB" || r.Header.Get("X-Youtube-Client-Name") != "1" || r.Header.Get("X-Youtube-Client-Version") != "2.fixture" || !strings.Contains(r.Header.Get("Cookie"), "SID=fixture") {
				t.Fatal("bad browse continuation request")
			}
			body = next
			if posts == 2 {
				body = `{"responseContext":{}}`
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})}}
	input := "https://www.youtube.com/playlist?list=PLfixture123456"
	page, err := client.GetPlaylistPage(context.Background(), input, "")
	if err != nil || page.Title != "课程列表" || len(page.Videos) != 1 || page.Videos[0].Title != "第一课" || page.Continuation != "page2" {
		t.Fatalf("initial page: %+v %v", page, err)
	}
	page, err = client.GetPlaylistPage(context.Background(), input, page.Continuation)
	if err != nil || len(page.Videos) != 1 || page.Videos[0].Title != "第二课" || page.Continuation != "" || posts != 1 {
		t.Fatalf("next page: %+v %v", page, err)
	}
	page, err = client.GetPlaylistPage(context.Background(), input, "page2")
	if err != nil || len(page.Videos) != 0 || page.Continuation != "" {
		t.Fatalf("terminal metadata-only page: %+v %v", page, err)
	}
}

func TestPlaylistModernLockupAndContinuation(t *testing.T) {
	entry := `{"lockupViewModel":{"contentId":"jNQXAC9IVRw","contentType":"LOCKUP_CONTENT_TYPE_VIDEO","metadata":{"lockupMetadataViewModel":{"title":{"content":"现代列表视频"},"metadata":{"contentMetadataViewModel":{"metadataRows":[{"metadataParts":[{"text":{"content":"示例作者"}}]}]}}}},"contentImage":{"thumbnailViewModel":{"overlays":[{"thumbnailOverlayBadgeViewModel":{"thumbnailBadges":[{"thumbnailBadgeViewModel":{"text":"1:02:03"}}]}}]}}}}`
	for _, body := range []string{
		`{"contents":{"twoColumnBrowseResultsRenderer":{"tabs":[{"tabRenderer":{"content":{"sectionListRenderer":{"contents":[{"itemSectionRenderer":{"contents":[` + entry + `,{"lockupViewModel":{"contentId":"PLother","contentType":"LOCKUP_CONTENT_TYPE_PLAYLIST"}},{"continuationItemRenderer":{"continuationEndpoint":{"continuationCommand":{"token":"next"}}}}]}}]}}}}]}}}`,
		`{"onResponseReceivedActions":[{"appendContinuationItemsAction":{"continuationItems":[` + entry + `,{"continuationItemViewModel":{"continuationCommand":{"innertubeCommand":{"continuationCommand":{"token":"next"}}}}}]}}]}`,
	} {
		var root map[string]any
		if err := json.Unmarshal([]byte(body), &root); err != nil {
			t.Fatal(err)
		}
		page, err := parsePlaylistPage(root)
		if err != nil || len(page.Videos) != 1 || page.Videos[0].Title != "现代列表视频" || page.Videos[0].Author != "示例作者" || page.Videos[0].Duration != 3723*time.Second || page.Continuation != "next" {
			t.Fatalf("bad modern page: %+v %v", page, err)
		}
	}
}

func TestPlaylistInputAndErrors(t *testing.T) {
	for _, input := range []string{"https://evil.test/playlist?list=PLfixture123456", "https://youtube.com.evil.test/playlist?list=PLfixture123456", "https://www.youtube.com/playlist?list=bad%2Fid", "https://www.youtube.com/watch?v=jNQXAC9IVRw"} {
		if _, err := PlaylistID(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	for _, body := range []string{`{"alerts":[{"alertRenderer":{"type":"ERROR","text":{"simpleText":"Private playlist"}}}]}`, `{}`} {
		var root map[string]any
		_ = json.Unmarshal([]byte(body), &root)
		if _, err := parsePlaylistPage(root); err == nil {
			t.Fatal("accepted inaccessible playlist")
		}
	}
	var root map[string]any
	_ = json.Unmarshal([]byte(`{"contents":{"playlistVideoListRenderer":{"contents":[]}}}`), &root)
	if page, err := parsePlaylistPage(root); err != nil || len(page.Videos) != 0 {
		t.Fatal("empty playlist should be valid")
	}
}
