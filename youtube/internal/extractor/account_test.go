package extractor

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kamiertop/videodown/internal/netclient"
)

const accountMenuFixture = `{"actions":[{"openPopupAction":{"popup":{"multiPageMenuRenderer":{"sections":[
  {"multiPageMenuSectionRenderer":{"items":[
    {"accountItemRenderer":{"accountName":{"simpleText":"备用账号"},"email":{"simpleText":"a@example.com"},"accountPhoto":{"thumbnails":[{"url":"https://yt3.ggpht.com/a-s88"}]},"isSelected":false}},
    {"accountItemRenderer":{"accountName":{"runs":[{"text":"当前"},{"text":"账号"}]},"email":{"simpleText":"b@example.com"},"accountPhoto":{"thumbnails":[{"url":"https://yt3.ggpht.com/b-s88"},{"url":"https://yt3.ggpht.com/b-s176"}]},"isSelected":true}}
  ]}},
  {"multiPageMenuSectionRenderer":{"items":[{"compactLinkRenderer":{"title":{"simpleText":"退出账号"}}}]}}
]}}}}]}`

func accountClient(t *testing.T, menuBody, homeBody string, menuStatus int) (*Client, *[]*http.Request) {
	t.Helper()
	requests := &[]*http.Request{}
	c := &Client{Cookie: "SAPISID=fixture; LOGIN_INFO=fixture", HTTPClient: netclient.New("", 0, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		*requests = append(*requests, r)
		data, status := homeBody, http.StatusOK
		if strings.HasSuffix(r.URL.Path, "/account/account_menu") {
			data, status = menuBody, menuStatus
		}
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data)), Request: r}, nil
	}))}
	return c, requests
}

func TestGetAccountContextParsesCurrentAccount(t *testing.T) {
	c, requests := accountClient(t, accountMenuFixture, `ytcfg.set({"LOGGED_IN":true,"VISITOR_DATA":"visitor-fixture","USER_SESSION_ID":"user-fixture","INNERTUBE_CLIENT_VERSION":"2.20261029.00.00"});`, http.StatusOK)
	account, err := c.GetAccountContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !account.LoggedIn || account.Name != "当前账号" || account.Email != "b@example.com" || account.Avatar != "https://yt3.ggpht.com/b-s176" {
		t.Fatalf("unexpected account %+v", account)
	}
	var menu *http.Request
	for _, request := range *requests {
		if strings.HasSuffix(request.URL.Path, "/account/account_menu") {
			menu = request
		}
	}
	if menu == nil {
		t.Fatal("no account menu request")
	}
	for _, check := range []struct{ key, want string }{
		{"Authorization", "SAPISIDHASH "},
		{"Cookie", c.Cookie},
		{"X-Youtube-Client-Name", "1"},
		{"X-Youtube-Client-Version", "2.20261029.00.00"},
		{"X-Origin", "https://www.youtube.com"},
		{"X-Goog-AuthUser", "0"},
	} {
		got := menu.Header.Get(check.key)
		ok := got == check.want
		if check.key == "Authorization" {
			ok = strings.HasPrefix(got, check.want)
		} else if check.key == "Cookie" {
			// httpDo 会追加 CONSENT cookie。
			ok = strings.HasPrefix(got, check.want+"; ")
		}
		if !ok {
			t.Errorf("%s = %q, want %q", check.key, got, check.want)
		}
	}
	var request innertubeRequest
	if err := json.NewDecoder(menu.Body).Decode(&request); err != nil {
		t.Fatal(err)
	}
	if request.Context.Client.ClientName != "WEB" || request.Context.Client.ClientVersion != "2.20261029.00.00" {
		t.Errorf("invalid account menu context %+v", request.Context.Client)
	}
}

func TestGetAccountContextLoggedOutCookie(t *testing.T) {
	c, _ := accountClient(t, `{}`, `ytcfg.set({"LOGGED_IN":false});`, http.StatusOK)
	account, err := c.GetAccountContext(context.Background())
	if !errors.Is(err, ErrCookieNotLoggedIn) {
		t.Fatalf("want ErrCookieNotLoggedIn, got %v", err)
	}
	if account != nil {
		t.Fatalf("unexpected account %+v", account)
	}
}

// Cookie 缺少 SAPISID 时账号菜单拿不到条目，但首页仍可判定登录。
func TestGetAccountContextFallsBackToSessionState(t *testing.T) {
	c, _ := accountClient(t, `{}`, `ytcfg.set({"LOGGED_IN":true,"VISITOR_DATA":"visitor-fixture"});`, http.StatusOK)
	account, err := c.GetAccountContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !account.LoggedIn || account.Name != "" || account.Avatar != "" {
		t.Fatalf("unexpected account %+v", account)
	}
}

func TestGetAccountContextUnverified(t *testing.T) {
	for _, tc := range []struct {
		name       string
		menu       string
		menuStatus int
	}{
		{name: "unrecognized page", menu: `{}`, menuStatus: http.StatusOK},
		{name: "empty account renderer", menu: `{"activeAccountHeaderRenderer":{}}`, menuStatus: http.StatusOK},
		{name: "HTTP forbidden", menu: `{}`, menuStatus: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := accountClient(t, tc.menu, `<html></html>`, tc.menuStatus)
			account, err := c.GetAccountContext(context.Background())
			if account != nil || !errors.Is(err, ErrAccountUnverified) {
				t.Errorf("GetAccountContext() = %+v, %v; want an unverified account", account, err)
			}
		})
	}
}

func TestGetAccountContextUsesVerifiedHomepageWhenMenuIsForbidden(t *testing.T) {
	c, _ := accountClient(t, `{}`, `ytcfg.set({"LOGGED_IN":true});`, http.StatusForbidden)
	account, err := c.GetAccountContext(context.Background())
	if err != nil || account == nil || !account.LoggedIn {
		t.Fatalf("GetAccountContext() = %+v, %v; want logged in", account, err)
	}
}

func TestParseAccountMenuFormats(t *testing.T) {
	// 旧版菜单把账号条目包在 accountSectionRenderer 里。
	legacy := `{"actions":[{"openPopupAction":{"popup":{"multiPageMenuRenderer":{"sections":[{"accountSectionRenderer":{"items":[{"accountItemRenderer":{"accountName":{"simpleText":"旧版账号"},"isSelected":true}}]}}]}}}}]}`
	account, ok := parseAccountMenu([]byte(legacy))
	if !ok || account.Name != "旧版账号" || !account.LoggedIn {
		t.Fatalf("legacy menu not parsed: %+v %v", account, ok)
	}
	// 未选中任何账号时取第一个条目。
	single := `{"actions":[{"openPopupAction":{"popup":{"multiPageMenuRenderer":{"sections":[{"multiPageMenuSectionRenderer":{"items":[{"accountItemRenderer":{"accountName":{"simpleText":"唯一账号"}}}]}}]}}}}]}`
	account, ok = parseAccountMenu([]byte(single))
	if !ok || account.Name != "唯一账号" {
		t.Fatalf("unselected menu not parsed: %+v %v", account, ok)
	}
	// 新版菜单不再返回账号切换列表，当前账号放在 activeAccountHeaderRenderer，
	// 字段与 accountItemRenderer 同名（accountName、email、accountPhoto）。
	modern := `{"actions":[{"openPopupAction":{"popup":{"multiPageMenuRenderer":{"sections":[{"multiPageMenuSectionRenderer":{"items":[{"activeAccountHeaderRenderer":{"accountName":{"simpleText":"当前账号"},"email":{"simpleText":"me@example.com"},"accountPhoto":{"thumbnails":[{"url":"https://yt3.ggpht.com/avatar-s88"}]}}},{"compactLinkRenderer":{"title":{"simpleText":"退出账号"}}}]}}]}}}}]}`
	account, ok = parseAccountMenu([]byte(modern))
	if !ok || account.Name != "当前账号" || account.Email != "me@example.com" || account.Avatar != "https://yt3.ggpht.com/avatar-s88" || !account.LoggedIn {
		t.Fatalf("active account header not parsed: %+v %v", account, ok)
	}
	// 新旧格式同时出现时优先 isSelected 的条目。
	mixed := `{"actions":[{"openPopupAction":{"popup":{"multiPageMenuRenderer":{"sections":[{"multiPageMenuSectionRenderer":{"items":[{"activeAccountHeaderRenderer":{"accountName":{"simpleText":"头部账号"}}}]}},{"multiPageMenuSectionRenderer":{"items":[{"accountItemRenderer":{"accountName":{"simpleText":"选中账号"},"isSelected":true}},{"accountItemRenderer":{"accountName":{"simpleText":"首个账号"}}}]}}]}}}}]}`
	account, ok = parseAccountMenu([]byte(mixed))
	if !ok || account.Name != "选中账号" {
		t.Fatalf("selected item should beat active header: %+v %v", account, ok)
	}
	if _, ok := parseAccountMenu([]byte(`{"responseContext":{}}`)); ok {
		t.Fatal("empty menu accepted")
	}
}

// 网络层失败返回 ErrAccountNetwork，并把超时/中断类错误翻译成可读文案。
func TestGetAccountContextNetworkError(t *testing.T) {
	for _, tc := range []struct{ name, raw, want string }{
		{"timeout", `Get "https://www.youtube.com/": net/http: request canceled`, "连接超时或被中断"},
		{"timeout-headers", `Get "https://www.youtube.com/": net/http: request canceled (Client.Timeout exceeded while awaiting headers)`, "连接超时或被中断"},
		{"deadline", `Post "https://www.youtube.com/youtubei/v1/account/account_menu": context deadline exceeded`, "连接超时或被中断"},
		{"reset", `Get "https://www.youtube.com/": read: connection reset by peer`, "连接被重置或不可达"},
		{"eof", `Get "https://www.youtube.com/": unexpected EOF`, "连接被重置或不可达"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.raw
			c := &Client{Cookie: "SAPISID=fixture; LOGIN_INFO=fixture", HTTPClient: netclient.New("", 0, roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return nil, errors.New(raw)
			}))}
			account, err := c.GetAccountContext(context.Background())
			if !errors.Is(err, ErrAccountNetwork) {
				t.Fatalf("want ErrAccountNetwork, got %v", err)
			}
			if account != nil {
				t.Fatalf("unexpected account %+v", account)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("message %q lacks %q", err, tc.want)
			}
		})
	}
}
