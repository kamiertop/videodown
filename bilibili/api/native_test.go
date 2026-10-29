package api

import (
	"fmt"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(r *http.Request, body string) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}
}
func testAPI(t *testing.T) (*BiliBili, *storage.Store) {
	t.Helper()
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return New(logger.New(), store), store
}
func TestNativeQRLoginPersistsAndLogoutClearsSharedStore(t *testing.T) {
	b, store := testAPI(t)
	b.client.GetClient().Transport = testTransport(func(r *http.Request) (*http.Response, error) {
		if strings.Contains(r.URL.Path, "qrcode/poll") {
			resp := response(r, `{"code":0,"data":{"code":0,"refresh_token":"test-refresh"}}`)
			resp.Header.Add("Set-Cookie", "SESSDATA=test-session; Path=/")
			resp.Header.Add("Set-Cookie", "bili_jct=test-csrf; Path=/")
			return resp, nil
		}
		if strings.Contains(r.URL.Path, "myinfo") {
			return response(r, `{"code":0,"data":{"profile":{"mid":42,"name":"原生账号"}}}`), nil
		}
		if strings.Contains(r.URL.Path, "exit") {
			return response(r, `{"code":2202}`), nil
		}
		return nil, fmt.Errorf("unexpected request: %s", r.URL.Path)
	})
	if _, err := b.PollQRCode("test-key"); err != nil {
		t.Fatal(err)
	}
	profile, err := b.MyInfo()
	if err != nil || profile.Mid != 42 {
		t.Fatalf("profile: %+v %v", profile, err)
	}
	other := New(logger.New(), store)
	if !other.IsLoggedIn() {
		t.Fatal("new API instance cannot read persisted login")
	}
	mid, err := other.getMid()
	if err != nil || mid != "42" {
		t.Fatal("account UID not saved for follow/favorite API")
	}
	if _, err := b.LogOut(); err == nil {
		t.Fatal("remote logout error should be visible")
	}
	if other.IsLoggedIn() {
		t.Fatal("remote error left local login active")
	}
	if store.Config().Bilibili.Auth != (storage.BilibiliAuth{}) {
		t.Fatal("auth survived logout")
	}
}
func TestFollowErrorIsNotAnEmptySuccessfulPage(t *testing.T) {
	b, store := testAPI(t)
	store.UpdateConfig(func(c *storage.Config) error { c.Bilibili.Auth.Cookie = "SESSDATA=test"; return nil })
	store.UpdateConfig(func(c *storage.Config) error { c.Bilibili.Auth.UID = 42; return nil })
	b.client.GetClient().Transport = testTransport(func(r *http.Request) (*http.Response, error) {
		return response(r, `{"code":-101,"message":"未登录"}`), nil
	})
	if _, err := b.FollowList(1, 20); err == nil {
		t.Fatal("nonzero response code must fail")
	}
}
func TestPlayHistoryUsesCursorAndReportsAPIErrors(t *testing.T) {
	b, store := testAPI(t)
	if err := store.UpdateConfig(func(c *storage.Config) error { c.Bilibili.Auth.Cookie = "SESSDATA=test"; return nil }); err != nil {
		t.Fatal(err)
	}
	b.client.GetClient().Transport = testTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("cursor") != "42" || r.URL.Query().Get("view_at") != "100" {
			t.Fatalf("wrong history cursor: %s", r.URL.RawQuery)
		}
		return response(r, `{"code":0,"data":{"cursor":{"max":21,"view_at":90},"list":[{"title":"历史视频","history":{"bvid":"BVhistory"}}]}}`), nil
	})
	page, err := b.PlayHistory(42, 100)
	if err != nil || len(page.List) != 1 || page.List[0].History.Bvid != "BVhistory" || page.Cursor.Max != 21 {
		t.Fatalf("history response was not decoded into the Go model: %+v %v", page, err)
	}
	b.client.GetClient().Transport = testTransport(func(r *http.Request) (*http.Response, error) {
		return response(r, `{"code":-101,"message":"未登录"}`), nil
	})
	if _, err := b.PlayHistory(0, 0); err == nil {
		t.Fatal("nonzero history response code must be shown as an error")
	}
}
func TestConcurrentWBISigningUsesOneKeyFetch(t *testing.T) {
	b, store := testAPI(t)
	if err := store.UpdateConfig(func(c *storage.Config) error { c.Bilibili.Auth.Cookie = "SESSDATA=test"; return nil }); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	b.client.GetClient().Transport = testTransport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return response(r, `{"code":0,"data":{"wbi_img":{"img_url":"https://i.example/abcdefghijklmnopqrstuvwxyz012345.png","sub_url":"https://i.example/012345abcdefghijklmnopqrstuvwxyz.png"}}}`), nil
	})
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Go(func() {
			if _, err := b.encWbiParams(map[string]string{"mid": "42"}); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("key fetches: %d", calls.Load())
	}
}
