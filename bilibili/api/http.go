package api

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/kamiertop/videodown/bilibili/util"
	"github.com/kamiertop/videodown/internal/constant"
)

const (
	biliBiliUrl = "https://www.bilibili.com/"
	spaceOrigin = "https://space.bilibili.com"
)

const (
	webLocation = "web_location"
)

func secPlatform() string {
	goOS := runtime.GOOS

	return fmt.Sprintf(`"%s"`, strings.ToUpper(goOS[:1])+goOS[1:])
}

func publicHeaders() map[string]string {
	return map[string]string{
		constant.AcceptEncoding:  "gzip, deflate, br, zstd",
		constant.AcceptLanguage:  "zh-CN,zh;q=0.9",
		constant.Accept:          "*/*",
		constant.SecCHUAMobile:   "?0",
		constant.Priority:        "u=1, i",
		constant.SecCHUA:         `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`,
		constant.SecCHUAPlatform: secPlatform(),
		constant.SecFetchMode:    "cors",
		constant.SecFetchDest:    "empty",
		constant.SecFetchSite:    "same-site",
		constant.UserAgent:       util.UserAgent(),
	}
}
