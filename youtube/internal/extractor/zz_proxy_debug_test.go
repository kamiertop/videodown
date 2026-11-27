package extractor

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kamiertop/videodown/internal/netclient"
)

func TestLiveProxyDebug(t *testing.T) {
	raw, err := os.ReadFile("/tmp/ytcookie.txt")
	if err != nil {
		t.Skip("no cookie fixture")
	}
	cookie := string(raw)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 1) 裸 netclient（带 netclient 全部设置）GET 首页
	c1 := netclient.NewWithProxy("http://127.0.0.1:7890", true, 30*time.Second, nil)
	start := time.Now()
	resp, err := c1.R().SetContext(ctx).Get("https://www.youtube.com/")
	if err != nil {
		t.Logf("bare netclient GET: err=%v elapsed=%v", err, time.Since(start).Round(time.Millisecond))
	} else {
		t.Logf("bare netclient GET: status=%d elapsed=%v", resp.StatusCode, time.Since(start).Round(time.Millisecond))
		resp.Body.Close()
	}

	// 2) extractor 的 httpDo 路径（VisionOS UA、Cookie 头等）
	c2 := &Client{Cookie: cookie, HTTPClient: netclient.NewWithProxy("http://127.0.0.1:7890", true, 30*time.Second, nil), client: &ClientInfo{Name: "WEB", Version: "2.20261029.00.00", UserAgent: VisionOSClient.UserAgent}}
	start = time.Now()
	page, err := c2.httpGetBodyBytes(ctx, "https://www.youtube.com/")
	t.Logf("extractor homepage: bytes=%d err=%v elapsed=%v", len(page), err, time.Since(start).Round(time.Millisecond))
}
