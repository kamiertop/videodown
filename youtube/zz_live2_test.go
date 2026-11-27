package youtube

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kamiertop/videodown/internal/netclient"
	"github.com/kamiertop/videodown/youtube/internal/extractor"
)

// 分别用直连与程序配置的代理复现 GetAccountContext，定位差异。
func TestLiveAccountPaths(t *testing.T) {
	raw, err := os.ReadFile("/tmp/ytcookie.txt")
	if err != nil {
		t.Skip("no cookie fixture")
	}
	cookie := string(raw)
	for _, tc := range []struct {
		name    string
		proxy   string
		enabled bool
	}{
		{"direct", "", false},
		{"proxy", "http://127.0.0.1:7890", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &extractor.Client{Cookie: cookie, HTTPClient: netclient.NewWithProxy(tc.proxy, tc.enabled, 30*time.Second, nil)}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			start := time.Now()
			account, err := c.GetAccountContext(ctx)
			t.Logf("%s: elapsed=%v account=%+v err=%v", tc.name, time.Since(start).Round(time.Millisecond), account, err)
		})
	}
}
