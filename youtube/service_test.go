package youtube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/kamiertop/videodown/internal/netclient"
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
func TestSetCookieNormalization(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store)
	for _, input := range []string{
		"Cookie: SID=x",
		"cookie: SID=x",
		"  Cookie: SID=x  ",
		"SID=x",
	} {
		if err := service.SetCookie(input); err != nil {
			t.Fatalf("%q: %v", input, err)
		}
		if service.Cookie() != "SID=x" {
			t.Fatalf("%q normalized to %q", input, service.Cookie())
		}
	}
	if err := service.SetCookie("cookie: SID=x\nBad: header"); err == nil {
		t.Fatal("newline accepted")
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
			if strings.Contains(r.Header.Get("Cookie"), "SID=fixture") {
				t.Error("account cookie sent for public video")
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
	invalid := Video{ID: video.ID, Formats: []Format{{ID: 0}}}
	if _, err := service.Download(t.Context(), invalid, 0, nil); !errors.Is(err, errSelectedFormatUnavailable) {
		t.Fatalf("invalid default format: %v", err)
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

func TestResolvePrefersGuestAndUsesCookieForRestrictedContent(t *testing.T) {
	const public = `{"playabilityStatus":{"status":"OK"},"videoDetails":{"title":"fixture"},"streamingData":{"formats":[{"itag":313,"url":"https://fixture.googlevideo.com/video","height":2160,"mimeType":"video/webm"}]}}`
	for _, tc := range []struct {
		name       string
		guest      string
		status     int
		wantCalls  []bool
		wantHeight int
	}{
		{name: "public 4K", guest: public, wantCalls: []bool{false}, wantHeight: 2160},
		{name: "private", guest: `{"playabilityStatus":{"status":"LOGIN_REQUIRED","reason":"This video is private"}}`, wantCalls: []bool{false, true}, wantHeight: 2160},
		{name: "bot check", guest: `{"playabilityStatus":{"status":"LOGIN_REQUIRED","reason":"Sign in to confirm you are not a bot"}}`, wantCalls: []bool{false}},
		{name: "rate limited", status: http.StatusTooManyRequests, wantCalls: []bool{false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, err := storage.OpenMemory()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			service := New(store)
			if err := service.SetCookie("SID=fixture"); err != nil {
				t.Fatal(err)
			}
			var calls []bool
			service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				body, status := "{}", http.StatusOK
				if strings.Contains(r.URL.Path, "/youtubei/") {
					authenticated := strings.Contains(r.Header.Get("Cookie"), "SID=fixture")
					calls = append(calls, authenticated)
					body = public
					if !authenticated {
						body = tc.guest
						if tc.status != 0 {
							status = tc.status
						}
					}
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
			})
			video, err := service.Resolve(context.Background(), "KWGhxyu5osA")
			if tc.wantHeight == 0 {
				if err == nil {
					t.Fatal("Resolve succeeded without a usable guest response")
				}
			} else if err != nil || len(video.Formats) != 1 || video.Formats[0].Height != tc.wantHeight {
				t.Fatalf("Resolve() = %+v, %v; want %dp", video, err, tc.wantHeight)
			}
			if !slices.Equal(calls, tc.wantCalls) {
				t.Errorf("account-cookie use per request = %v; want %v", calls, tc.wantCalls)
			}
		})
	}
}

func TestSelectedFormatUsesMatchingSessionWithoutReplacingGuestCache(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	service := New(store)
	if err := service.SetCookie("SID=fixture"); err != nil {
		t.Fatal(err)
	}
	const id = "KWGhxyu5osA"
	var calls []bool
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body := "{}"
		if strings.Contains(r.URL.Path, "/youtubei/") {
			authenticated := strings.Contains(r.Header.Get("Cookie"), "SID=fixture")
			calls = append(calls, authenticated)
			itag, height := 18, 360
			if authenticated {
				itag, height = 313, 2160
			}
			body = fmt.Sprintf(`{"playabilityStatus":{"status":"OK"},"videoDetails":{"title":"fixture"},"streamingData":{"formats":[{"itag":%d,"height":%d,"url":"https://fixture.googlevideo.com/video","mimeType":"video/webm"}]}}`, itag, height)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})
	ctx := context.Background()
	for range 2 {
		client, raw, err := service.fetchVideo(ctx, id, 313)
		if err != nil || client.Cookie == "" || len(raw.Formats) != 1 || raw.Formats[0].ItagNo != 313 {
			t.Fatalf("could not resolve selected account format: %v", err)
		}
	}
	video, err := service.Resolve(ctx, id)
	if err != nil || len(video.Formats) != 1 || video.Formats[0].ID != 18 {
		t.Fatalf("ordinary Resolve() = %+v, %v; want guest formats", video, err)
	}
	if !slices.Equal(calls, []bool{false, true}) {
		t.Errorf("account-cookie use per request = %v; want [false true]", calls)
	}
}

func TestPlaybackErrorsDistinguishAccountAndNetwork(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	service := New(store)
	if err := service.SetCookie("SID=fixture"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		err  error
		want string
	}{
		{fmt.Errorf("%w: reload", errAuthenticatedPlayback), "Cookie 有效"},
		{&extractor.ErrPlayabiltyStatus{Status: "UNPLAYABLE", Reason: "The page needs to be reloaded."}, "播放校验"},
		{&extractor.ErrPlayabiltyStatus{Status: "LOGIN_REQUIRED", Reason: "Sign in to confirm you’re not a bot"}, "验证当前网络"},
		{extractor.ErrVideoPrivate, "账号权限"},
	} {
		if got := service.parseError(tc.err).Error(); !strings.Contains(got, tc.want) {
			t.Errorf("parseError(%v) = %q; want %q", tc.err, got, tc.want)
		}
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
	var requestCookie string
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		data := "{}"
		if strings.Contains(r.URL.Path, "/youtubei/") {
			requests++
			requestCookie = r.Header.Get("Cookie")
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
	if strings.Contains(requestCookie, "SID=new-fixture") {
		t.Fatal("first parse after saving sent the cookie for a public video")
	}
	previousClient := service.requestClient()
	if err := service.SetProxy(" http://127.0.0.1:7890/ ", true); err != nil {
		t.Fatal(err)
	}
	if service.Proxy() != "http://127.0.0.1:7890" {
		t.Fatal("proxy was not normalized and persisted")
	}
	if service.requestClient() == previousClient {
		t.Fatal("proxy change reused old network client")
	}
	if _, err := service.Resolve(context.Background(), "KWGhxyu5osA"); err != nil {
		t.Fatal(err)
	}
	if requests != 4 {
		t.Fatal("proxy change reused a signed URL from the old network")
	}
	if err := service.SetProxy("ftp://invalid.test", true); err == nil {
		t.Fatal("accepted unsupported proxy")
	}
	if service.Proxy() != "http://127.0.0.1:7890" {
		t.Fatal("failed save changed the proxy")
	}
	if err := service.SetProxy("", false); err != nil {
		t.Fatal(err)
	}
	if service.ProxyEnabled() || service.Proxy() != netclient.DefaultProxy {
		t.Fatal("proxy was not disabled")
	}

}

func TestProxySwitchDefaultsAndPersistence(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	service := New(store)
	if service.Proxy() != netclient.DefaultProxy || service.ProxyEnabled() {
		t.Fatal("proxy should default to the local address, disabled")
	}
	if err := service.SetProxy("socks5://127.0.0.1:1080", true); err != nil {
		t.Fatal(err)
	}
	enabledClient := service.requestClient()
	if err := service.SetProxy(service.Proxy(), false); err != nil {
		t.Fatal(err)
	}
	if service.requestClient() == enabledClient {
		t.Fatal("disabling proxy reused the old client")
	}
	if service.Proxy() != "socks5://127.0.0.1:1080" || service.ProxyEnabled() {
		t.Fatal("disabling should preserve the address")
	}
	store.Close()
	store, err = storage.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service = New(store)
	if service.Proxy() != "socks5://127.0.0.1:1080" || service.ProxyEnabled() {
		t.Fatal("proxy settings were not persisted")
	}
	if err := service.SetProxy(service.Proxy(), true); err != nil {
		t.Fatal(err)
	}
	if !service.ProxyEnabled() {
		t.Fatal("could not re-enable saved proxy")
	}
}

func TestAccountVerifiesCookie(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store)
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		data := `ytcfg.set({"LOGGED_IN":true,"VISITOR_DATA":"visitor-fixture","INNERTUBE_CLIENT_VERSION":"2.20261029.00.00"});`
		if strings.HasSuffix(r.URL.Path, "/account/account_menu") {
			data = `{"actions":[{"openPopupAction":{"popup":{"multiPageMenuRenderer":{"sections":[{"multiPageMenuSectionRenderer":{"items":[{"accountItemRenderer":{"accountName":{"simpleText":"账号_fixture"},"email":{"simpleText":"fixture@example.com"},"accountPhoto":{"thumbnails":[{"url":"https://yt3.ggpht.com/fixture"}]},"isSelected":true}}]}}]}}}}]}`
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data)), Request: r}, nil
	})
	if _, err := service.Account(context.Background()); err == nil {
		t.Fatal("accepted missing cookie")
	}
	if err := service.SetCookie("SAPISID=fixture; LOGIN_INFO=fixture"); err != nil {
		t.Fatal(err)
	}
	info, err := service.Account(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !info.LoggedIn || info.Name != "账号_fixture" || info.Email != "fixture@example.com" || info.Avatar != "https://yt3.ggpht.com/fixture" {
		t.Fatalf("unexpected account info %+v", info)
	}
}

func TestAccountNetworkErrorClassification(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store)
	if err := service.SetCookie("SAPISID=fixture; LOGIN_INFO=fixture"); err != nil {
		t.Fatal(err)
	}
	service.transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("get \"https://www.youtube.com/\": net/http: request canceled")
	})
	_, err = service.Account(context.Background())
	if !errors.Is(err, ErrAccountNetwork) {
		t.Fatalf("want ErrAccountNetwork, got %v", err)
	}
	if !errors.Is(err, extractor.ErrAccountNetwork) {
		t.Fatal("service did not re-export the network sentinel")
	}
	if !strings.Contains(err.Error(), "连接超时或被中断") {
		t.Fatalf("raw error surfaced untranslated: %v", err)
	}
}
