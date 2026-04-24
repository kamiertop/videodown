package api

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	originURL = "https://www.douyin.com"
	referURL  = "https://www.douyin.com/"
)

const (
	Accept          = "Accept"
	AcceptEncoding  = "Accept-Encoding"
	AcceptLanguage  = "Accept-Language"
	Origin          = "Origin"
	Cookie          = "Cookie"
	Referer         = "Referer"
	Priority        = "Priority"
	SecCHUA         = "Sec-CH-UA"
	SecCHUAMobile   = "Sec-CH-UA-Mobile"
	SecCHUAPlatform = "Sec-CH-UA-Platform"
	SecChFetchDest  = "Sec-Fetch-Dest"
	SecChFetchMode  = "Sec-Fetch-Mode"
	SecChFetchSite  = "Sec-Fetch-Site"
	UserAgent       = "User-Agent"
	CacheControl    = "Cache-Control"
)

// 默认返回Windows的
func userAgent() string {
	switch runtime.GOOS {
	case "darwin":
		return "Mozilla/5.0 (Macintosh; Intel Mac  OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	case "linux":
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	case "windows":
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	}
}

func osName() string {
	goOS := runtime.GOOS

	return strings.ToUpper(goOS[:1]) + goOS[1:]
}

func (d *Douyin) publicQueryParams() (map[string]any, error) {
	webId, err := d.getWebId()
	if err != nil {
		return nil, fmt.Errorf("获取 webId 失败: %w", err)
	}
	msToken, err := d.getMsToken()
	if err != nil {
		return nil, fmt.Errorf("获取 msToken 失败: %w", err)
	}
	p, err := d.cookieQuery()
	if err != nil {
		return nil, fmt.Errorf("获取 cookie 参数失败: %w", err)
	}
	return map[string]any{
		"device_platform":     "webapp",
		"aid":                 6383,
		"channel":             "channel_pc_web",
		"pc_client_type":      1,
		"pc_libra_divert":     osName(),
		"update_version_code": 170400,
		"support_h265":        1,
		"support_dash":        1,
		"version_code":        "170400",
		"version_name":        "17.4.0",
		"cookie_enabled":      true,
		"screen_width":        2560,
		"screen_height":       1440,
		"browser_language":    "zh-CN",
		"browser_platform":    "Linux x86_64", // TODO
		"browser_name":        "Chrome",
		"browser_version":     "147.0.0.0",
		"os_name":             osName(),
		"os_version":          "x86_64",
		"platform":            "PC",
		"downlink":            1.4,
		"device_memory":       32,
		"cpu_core_num":        runtime.NumCPU(),
		"browser_online":      true,
		"engine_name":         "Blink",
		"engine_version":      "147.0.0.0",
		"effective_type":      "4g",
		"round_trip_time":     100,
		"webid":               webId,
		"uifid":               p.uifid,
		"verifyFp":            p.verifyFp,
		"fp":                  p.fp,
		"msToken":             msToken,
	}, nil
}

func (d *Douyin) publicHeaders() (map[string]string, error) {
	cookie, err := d.GetCookie()
	if err != nil {
		return nil, fmt.Errorf("获取 cookie 失败: %w", err)
	}
	//p, err := d.CookieQuery()
	//if err != nil {
	//	return nil, fmt.Errorf("获取 cookie 参数失败: %w", err)
	//}
	return map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "zh-CN,zh;q=0.9",
		"Accept-Encoding":    "gzip, deflate, br, zstd",
		"Origin":             originURL,
		"Pragma":             "no-cache",
		"Priority":           "u=1, i",
		"Cache-Control":      "no-cache",
		"Referer":            referURL,
		"Sec-CH-UA":          `"Google Chrome";v="147", "Not:A-Brand";v="8", "Chromium";v="147"`,
		"Sec-CH-UA-Mobile":   "?0",
		"Sec-CH-UA-Platform": fmt.Sprintf(`"%s"`, osName()),
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"User-Agent":         userAgent(),
		//"Uifid":              p.uifid,
		Cookie: cookie,
	}, nil

}
