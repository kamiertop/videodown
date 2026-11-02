package youtube

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
	"github.com/kamiertop/videodown/internal/netclient"
)

type proxyTestAPI struct{ fakeAPI }

func (a *proxyTestAPI) SetProxy(raw string, enabled bool) error {
	value, err := netclient.NormalizeProxy(raw)
	if err == nil {
		if value == "" {
			value = netclient.DefaultProxy
		}
		a.proxy = value
		a.proxyEnabled = enabled
	}
	return err
}

func TestAccountProxySaveValidateAndClear(t *testing.T) {
	s := NewState()
	api := &proxyTestAPI{fakeAPI: fakeAPI{proxy: netclient.DefaultProxy}}
	s.Configure(api, guiruntime.Runtime{Background: func(work func() error, done func(error)) { done(work()) }}, func(fn func()) { fn() })
	defer s.Stop()
	scroll := &ui.ScrollState{}
	view := ui.NewTester(func(c *ui.Context) { core.Use(c, core.Settings{}); s.accountPage(c, scroll) }, 760, 1000)
	view.Frame()
	if err := view.Click("YouTube 代理服务器"); err != nil {
		t.Fatal(err)
	}
	// Default address is already filled in.
	if s.proxy != netclient.DefaultProxy || s.proxyEnabled {
		t.Fatal("unexpected proxy defaults")
	}
	view.Frame()
	oldImage := &thumbnail{Loading: true, Generation: 1}
	s.thumbnails = map[string]*thumbnail{"pending": oldImage}
	if err := view.Click("启用 YouTube 代理"); err != nil {
		t.Fatal(err)
	}
	view.Frame()
	if api.Proxy() != "http://127.0.0.1:7890" || s.imageClient == nil || !api.ProxyEnabled() {
		t.Fatal("proxy input did not save")
	}
	if len(s.thumbnails) != 0 || oldImage.Generation != 2 {
		t.Fatal("old proxy image callbacks were not invalidated")
	}

	if view.HasText("保存代理") {
		t.Fatal("manual save button should be removed")
	}
	s.proxy = ""
	view.Frame()
	if err := view.Click("YouTube 代理服务器"); err != nil {
		t.Fatal(err)
	}
	view.Type("http://127.0.0.1:7891")
	view.Frame()
	if api.Proxy() != "http://127.0.0.1:7891" {
		t.Fatal("editing the input did not auto-save")
	}
	s.proxy = netclient.DefaultProxy
	s.saveProxy()
	view.Advance(300 * time.Millisecond)
	f, err := os.Create(filepath.Join(os.TempDir(), "videodown-youtube-proxy.png"))
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, view.Image())
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	s.proxy = "ftp://invalid.test"
	s.saveProxy()
	if api.Proxy() != "http://127.0.0.1:7890" || s.proxyMessage == "代理设置已保存" {
		t.Fatal("invalid proxy changed active settings")
	}
	s.proxy = netclient.DefaultProxy
	s.proxyEnabled = false
	s.saveProxy()
	view.Frame()
	if api.ProxyEnabled() || s.busy || s.proxyBusy {
		t.Fatal("clearing proxy failed")
	}
	s.busy = true
	s.proxy = "http://127.0.0.1:7891"
	s.saveProxy()
	if api.ProxyEnabled() || api.Proxy() != netclient.DefaultProxy {
		t.Fatal("proxy changed during active work")
	}
}

func TestProxyAutoSaveDebouncesAndKeepsCookieIdle(t *testing.T) {
	s := NewState()
	api := &proxyTestAPI{fakeAPI: fakeAPI{proxy: netclient.DefaultProxy}}
	var timers []func()
	type pendingJob struct {
		work func() error
		done func(error)
	}
	var jobs []pendingJob
	s.Configure(api, guiruntime.Runtime{
		Background: func(work func() error, done func(error)) { jobs = append(jobs, pendingJob{work, done}) },
		After: func(delay time.Duration, fn func()) func() {
			if delay != proxySaveDelay {
				t.Fatal("wrong debounce delay")
			}
			timers = append(timers, fn)
			return func() {}
		},
	}, func(fn func()) { fn() })
	defer s.Stop()
	s.proxy = "http://127.0.0.1:7891"
	s.scheduleProxySave()
	s.proxy = "http://127.0.0.1:7892"
	s.scheduleProxySave()
	timers[0]()
	if len(jobs) != 0 {
		t.Fatal("obsolete timer saved a partial address")
	}
	timers[1]()
	if len(jobs) != 1 || !s.proxyBusy || s.busy {
		t.Fatal("proxy save affected Cookie loading state")
	}
	// A newer edit must survive the previous save's completion.
	s.proxy = "http://127.0.0.1:7893"
	s.scheduleProxySave()
	timers[2]()
	if !s.proxySavePending {
		t.Fatal("edit during a save was dropped")
	}
	jobs[0].done(jobs[0].work())
	if s.proxy != "http://127.0.0.1:7893" || len(jobs) != 2 {
		t.Fatal("completion replaced a newer edit")
	}
	jobs[1].done(jobs[1].work())
	if api.Proxy() != s.proxy || s.proxyMessage != "" || s.busy || s.proxyBusy {
		t.Fatal("latest address was not saved quietly")
	}
}
