package api

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	biliBiliUrl = "https://www.bilibili.com/"
	spaceOrigin = "https://space.bilibili.com"
)

const (
	Origin          = "Origin"
	Referer         = "Referer"
	UserAgent       = "User-Agent"
	Accept          = "Accept"
	AcceptEncoding  = "Accept-Encoding"
	AcceptLanguage  = "Accept-Language"
	Priority        = "Priority"
	SecCHUA         = "Sec-CH-UA"
	SecCHUAMobile   = "Sec-CH-UA-Mobile"
	SecCHUAPlatform = "Sec-CH-UA-Platform"
	SecFetchDest    = "Sec-Fetch-Dest"
	SecFetchMode    = "Sec-Fetch-Mode"
	SecFetchSite    = "Sec-Fetch-Site"
	Cookie          = "Cookie"
)

const (
	webLocation = "web_location"
)

func userAgent() string {
	switch runtime.GOOS {
	case "darwin":
		return "Mozilla/5.0 (Macintosh; Intel Mac  OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	case "linux":
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
	case "windows":
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	default:
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	}
}

func secPlatform() string {
	goOS := runtime.GOOS
	return fmt.Sprintf(`"%s"`, strings.ToUpper(goOS[:1])+goOS[1:])
}

func publicHeaders() map[string]string {
	return map[string]string{
		AcceptEncoding:  "gzip, deflate, br, zstd",
		AcceptLanguage:  "zh-CN,zh;q=0.9",
		Accept:          "*/*",
		SecCHUAMobile:   "?0",
		Priority:        "u=1, i",
		SecCHUA:         `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`,
		SecCHUAPlatform: secPlatform(),
		SecFetchMode:    "cors",
		SecFetchDest:    "empty",
		SecFetchSite:    "same-site",
		UserAgent:       userAgent(),
	}
}
