package youtube

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/internal/storage"
	yt "github.com/kamiertop/videodown/youtube"
)

func syncRuntime() guiruntime.Runtime {
	return guiruntime.Runtime{
		Background: func(work func() error, done func(error)) { done(work()) },
		Dispatch:   func(fn func()) { fn() },
	}
}

// 保存 Cookie 后自动验证：成功时账号页能看到新账号，失败时给出原因。
// 网络类失败与 Cookie 无效分开标记，提示文案不同。
func TestSaveCookieVerifiesAccount(t *testing.T) {
	for _, tc := range []struct {
		name    string
		account yt.AccountInfo
		err     error
		network bool
	}{
		{name: "login", account: yt.AccountInfo{LoggedIn: true, Name: "账号_fixture", Email: "fixture@example.com"}},
		{name: "failure", err: yt.ErrCookieNotLoggedIn},
		{name: "network", err: fmt.Errorf("%w：%s", yt.ErrAccountNetwork, "连接超时或被中断"), network: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, err := storage.OpenMemory()
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			service := yt.New(store)
			s := NewState()
			s.Configure(service, syncRuntime())
			s.accountLoader = func(context.Context) (yt.AccountInfo, error) {
				return tc.account, tc.err
			}
			s.cookie = "SAPISID=fixture"
			s.saveCookie()
			if !s.cookieSaved || s.busy || s.accountBusy {
				t.Fatal("cookie not saved or check unfinished")
			}
			if tc.err == nil {
				if s.accountError != nil || s.account != tc.account {
					t.Fatalf("account not verified: %+v %q", s.account, s.accountError)
				}
			} else {
				if !errors.Is(s.accountError, tc.err) || s.account.LoggedIn || errors.Is(s.accountError, yt.ErrAccountNetwork) != tc.network {
					t.Fatalf("failure not classified: %q network=%v", s.accountError, errors.Is(s.accountError, yt.ErrAccountNetwork))
				}
			}
		})
	}
}

// 留空保存即清除 Cookie 并清空账号信息。
func TestSaveCookieClearsAccount(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := yt.New(store)
	s := NewState()
	s.Configure(service, syncRuntime())
	s.accountLoader = func(context.Context) (yt.AccountInfo, error) {
		return yt.AccountInfo{LoggedIn: true, Name: "账号_fixture"}, nil
	}
	s.cookie = "SAPISID=fixture"
	s.saveCookie()
	if !s.account.LoggedIn {
		t.Fatal("account not verified")
	}
	s.cookie = ""
	s.saveCookie()
	if s.cookieSaved || s.authMessage != "Cookie 已清除" || s.account.LoggedIn || s.accountError != nil || s.accountAvatar != nil {
		t.Fatalf("account not cleared: %+v %q", s.account, s.authMessage)
	}
}

// 验证在途时再次保存：旧验证被取消、迟到结果丢弃，最终显示新 Cookie
// 的账号，不会残留旧结果。
func TestCheckAccountSupersedesWhileBusy(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := yt.New(store)
	s := NewState()
	// 工作立即执行，完成回调先进先出地延后释放，才能复现保存与旧验证交错的时序。
	var pending []func()
	s.Configure(service, guiruntime.Runtime{
		Background: func(work func() error, done func(error)) {
			err := work()
			pending = append(pending, func() { done(err) })
		},
		Dispatch: func(fn func()) { fn() },
	})
	release := func() {
		if len(pending) == 0 {
			t.Fatal("nothing to release")
		}
		fn := pending[0]
		pending = pending[1:]
		fn()
	}
	var calls int
	s.accountLoader = func(context.Context) (yt.AccountInfo, error) {
		calls++
		return yt.AccountInfo{LoggedIn: true, Name: "账号_" + service.Cookie()}, nil
	}
	s.cookie = "SAPISID=first"
	s.saveCookie()
	release() // 保存完成，验证 #1 开始（saved=first）。
	s.cookie = "SAPISID=second"
	s.saveCookie()
	release() // 验证 #1 迟到完成：发现 Cookie 已变，自动重查（saved=second）。
	release() // 第二次保存完成：取消在途验证并重启验证 #3（saved=second）。
	release() // 被取消的验证 #2 迟到完成：generation 已变，结果丢弃。
	release() // 验证 #3 完成：新账号生效。
	if calls != 3 || s.account.Name != "账号_SAPISID=second" || s.accountBusy || s.accountCheckedAt.IsZero() {
		t.Fatalf("stale result kept: calls=%d account=%+v busy=%v at=%v", calls, s.account, s.accountBusy, s.accountCheckedAt)
	}
	if service.Cookie() != "SAPISID=second" {
		t.Fatal("cookie not replaced")
	}
}

// 账号页布局：登录后显示头像、昵称、邮箱与生效提示，失败时显示原因；
// 网络类失败提示检查代理，Cookie 无效提示重新复制。
func TestAccountPageShowsVerifiedAccount(t *testing.T) {
	for _, tc := range []struct {
		name    string
		account yt.AccountInfo
		error   error
		network bool
	}{
		{name: "登录成功", account: yt.AccountInfo{LoggedIn: true, Name: "账号_fixture", Email: "fixture@example.com"}},
		{name: "验证失败", error: yt.ErrCookieNotLoggedIn},
		{name: "状态未确认", error: yt.ErrAccountUnverified},
		{name: "网络失败", error: fmt.Errorf("%w：%s", yt.ErrAccountNetwork, "连接超时或被中断"), network: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewState()
			s.cookieSaved = true
			s.account = tc.account
			s.accountCheckedAt = time.Date(2026, 11, 27, 15, 4, 5, 0, time.Local)
			s.accountError = tc.error
			scroll := &ui.ScrollState{}
			test := ui.NewTester(func(c *ui.Context) {
				style.UseTheme(c)
				s.Form(c, "/youtube/profile", scroll)
			}, 900, 700)
			test.Frame()
			for _, label := range []string{"当前登录账号", "YouTube 账号"} {
				if _, ok := test.Find(label); ok {
					t.Fatalf("redundant account heading %q is visible", label)
				}
			}
			if tc.error == nil {
				for _, want := range []string{"账号_fixture", "fixture@example.com", "Cookie 有效，已登录（15:04:05 验证）", "保存并验证", "刷新账号"} {
					if _, ok := test.Find(want); !ok {
						t.Fatalf("missing %q", want)
					}
				}
			} else {
				if _, ok := test.Find("Cookie 验证失败：" + tc.error.Error()); !ok {
					t.Fatal("missing verification failure message")
				}
				if _, ok := test.Find("Cookie 有效"); ok {
					t.Fatal("success message shown after failure")
				}
				want := "请从已登录的浏览器重新复制 Cookie 后保存"
				if errors.Is(tc.error, yt.ErrAccountUnverified) {
					if _, ok := test.Find(want); ok {
						t.Fatal("unverified account shown as requiring a new Cookie")
					}
					return
				}
				if tc.network {
					want = "程序当前直连 YouTube；可开启代理软件的 TUN（虚拟网卡）模式，或在此启用与浏览器相同的代理后重试"
				}
				if _, ok := test.Find(want); !ok {
					t.Fatalf("missing hint %q", want)
				}
			}
		})
	}
}

// 上次验证因网络失败时，保存代理后自动用新出口重验。
func TestSaveProxyRetriesNetworkFailedCheck(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := yt.New(store)
	s := NewState()
	s.Configure(service, syncRuntime())
	var calls int
	s.accountLoader = func(context.Context) (yt.AccountInfo, error) {
		calls++
		if calls == 1 {
			return yt.AccountInfo{}, fmt.Errorf("%w：%s", yt.ErrAccountNetwork, "连接超时或被中断")
		}
		return yt.AccountInfo{LoggedIn: true, Name: "账号_fixture"}, nil
	}
	s.cookie = "SAPISID=fixture"
	s.saveCookie()
	if !errors.Is(s.accountError, yt.ErrAccountNetwork) || s.account.LoggedIn {
		t.Fatal("network failure not recorded")
	}
	s.proxyEnabled = true
	s.saveProxy()
	if calls != 2 || !s.account.LoggedIn || errors.Is(s.accountError, yt.ErrAccountNetwork) {
		t.Fatalf("proxy save did not re-verify: calls=%d account=%+v network=%v", calls, s.account, errors.Is(s.accountError, yt.ErrAccountNetwork))
	}
}

// 回归：网络中断后再次点击“保存并验证”，旧的“已登录”结果必须被清除，
// 显示网络失败而不是残留上一次的成功状态。
func TestSaveCookieClearsStaleSuccessWhenNetworkFails(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := yt.New(store)
	s := NewState()
	s.Configure(service, syncRuntime())
	var calls int
	s.accountLoader = func(context.Context) (yt.AccountInfo, error) {
		calls++
		if calls == 1 {
			return yt.AccountInfo{LoggedIn: true, Name: "账号_fixture"}, nil
		}
		return yt.AccountInfo{}, fmt.Errorf("%w：%s", yt.ErrAccountNetwork, "连接超时或被中断")
	}
	s.cookie = "SAPISID=fixture"
	s.saveCookie()
	if !s.account.LoggedIn || s.accountCheckedAt.IsZero() {
		t.Fatal("first verification should succeed")
	}
	// 模拟网络断开后的再次点击：保存同样的 Cookie 也必须重新验证。
	s.saveCookie()
	if s.account.LoggedIn || s.accountCheckedAt != (time.Time{}) || s.accountError == nil || !errors.Is(s.accountError, yt.ErrAccountNetwork) {
		t.Fatalf("stale success kept: account=%+v error=%q network=%v", s.account, s.accountError, errors.Is(s.accountError, yt.ErrAccountNetwork))
	}
}
