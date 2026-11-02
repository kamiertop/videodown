package api

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
)

func TestNativeSharedVideoAndProfileLinks(t *testing.T) {
	d := &Douyin{client: req.C()}
	calls := 0
	d.client.Transport.WrapRoundTripFunc(func(http.RoundTripper) req.HttpRoundTripFunc {
		return func(r *http.Request) (*http.Response, error) {
			calls++
			location := "https://www.douyin.com/share/note/12345/?from=share"
			switch r.URL.Path {
			case "/user-short/":
				location = "https://www.douyin.com/share/user/MS4wLjAB_fixture/?from=share"
			case "/relative/":
				location = "/video/12345/"
			case "/bad/":
				location = "https://example.com/video/12345"
			case "/loop/":
				location = "/loop/"
			}
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {location}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		}
	})
	for _, raw := range []string{"12345", "https://www.douyin.com/video/12345", "https://www.douyin.com/share/slides/12345/", "https://www.douyin.com/note/12345/?x=1", "https://www.douyin.com/jingxuan?modal_id=12345", "复制打开抖音 https://v.douyin.com/short/ 查看作品", "https://v.douyin.com/relative/"} {
		id, err := d.ParseVideo(raw)
		if err != nil || id != "12345" {
			t.Fatalf("%q => %q %v", raw, id, err)
		}
	}
	for _, raw := range []string{"MS4wLjAB_fixture", "https://www.douyin.com/user/MS4wLjAB_fixture/?x=1", "复制 https://v.douyin.com/user-short/ 打开"} {
		id, err := d.ParseSecUserID(raw)
		if err != nil || id != "MS4wLjAB_fixture" {
			t.Fatalf("%q => %q %v", raw, id, err)
		}
	}
	for _, raw := range []string{"", "123abc", "https://douyin.com.evil.test/video/12345", "https://v.douyin.com/bad/", "https://v.douyin.com/loop/"} {
		if id, err := d.ParseVideo(raw); err == nil {
			t.Fatalf("accepted invalid link %q => %s", raw, id)
		}
	}
	if calls != 10 {
		t.Fatalf("unexpected redirect requests: %d", calls)
	}
}

func TestNativeCookieResetsCachedIdentity(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	d := New(logger.New(), store)
	d.secUserID, d.userID = "old-sec", "old-uid"
	d.msToken.value, d.webId.value = "old-token", "old-web"
	if err := d.SetCookie(" fixture=test "); err != nil {
		t.Fatal(err)
	}
	if store.Config().Douyin.Cookie != "fixture=test" || d.secUserID != "" || d.userID != "" || d.msToken.value != "" || d.webId.value != "" {
		t.Fatal("cookie persisted without resetting identity")
	}
	if err := d.SetCookie(""); err != nil {
		t.Fatal(err)
	}
	if _, err := d.GetCookie(); err == nil {
		t.Fatal("logged out cookie still available")
	}
}
