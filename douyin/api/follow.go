package api

import (
	"fmt"
	"runtime"

	"github.com/kamiertop/videodown/douyin/model"
)

// FollowList 关注列表，需要分页获取，前端默认20条。这里不再传参了，就默认20条了。
// https://www-hj.douyin.com/aweme/v1/web/im/spotlight/relation/ 也返回关注列表，但是最后一个元素是自身信息
func (d *Douyin) FollowList() (model.Follow, error) {
	var resp model.Follow
	webId, err := d.getWebId()
	if err != nil {
		return resp, fmt.Errorf("获取 webId 失败: %w", err)
	}
	cookie, err := d.getCookie()
	if err != nil {
		return resp, fmt.Errorf("获取 cookie 失败: %w", err)
	}
	msToken, err := d.getMsToken()
	if err != nil {
		return resp, fmt.Errorf("获取 msToken 失败: %w", err)
	}
	p, err := d.cookieQuery()
	if err != nil {
		return resp, fmt.Errorf("获取 cookie 参数失败: %w", err)
	}
	// 目前不需要user_id和sec_user_id也能返回关注列表
	err = d.client.Get("https://www-hj.douyin.com/aweme/v1/web/im/spotlight/relation/").
		SetQueryParamsAnyType(map[string]any{
			"device_platform": "webapp",
			"aid":             "6383",
			"channel":         "channel_pc_web",
			//"user_id":              0,
			//"sec_user_id":          "",
			"offset":               20,
			"min_time":             0,
			"max_time":             0,
			"count":                20,
			"source_type":          4,
			"gps_access":           0,
			"is_top":               0,
			"pc_client_type":       1,
			"pc_libra_divert":      "Linux",
			"support_h265":         1,
			"support_dash":         1,
			"webcast_sdk_version":  "170400",
			"webcast_version_code": "170400",
			"version_code":         "170400",
			"version_name":         "17.4.0",
			"cookie_enabled":       true,
			"screen_width":         2560,
			"screen_height":        1440,
			"browser_language":     "zh-CN",
			"browser_platform":     "Linux x86_64",
			"browser_name":         "Chrome",
			"browser_version":      "146.0.0.0",
			"browser_online":       true,
			"engine_name":          "Blink",
			"engine_version":       "146.0.0.0",
			"os_name":              "Linux",
			"os_version":           "x86_64",
			"device_memory":        runtime.NumCPU(),
			"platform":             "PC",
			"downlink":             0.35,
			"effective_type":       "4g",
			"round_trip_time":      50,
			"webid":                webId,
			"uifid":                p.uifid,
			"verifyFp":             p.verifyFp,
			"fp":                   p.fp,
			"msToken":              msToken,
		}).
		SetHeaders(map[string]string{
			"Accept":             "*/*",
			"Accept-Language":    "zh-CN,zh;q=0.9",
			"Accept-Encoding":    "gzip, deflate, br, zstd",
			"Origin":             "https://www.douyin.com/",
			"Referer":            "https://www.douyin.com/",
			"Priority":           "u=1, i",
			"Sec-CH-UA":          `"Google Chrome";v="146", "Not:A-Brand";v="24", "Chromium";v="146"`,
			"Sec-CH-UA-Mobile":   "?0",
			"Sec-CH-UA-Platform": `"Linux"`,
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-site",
			"User-Agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
			"Cookie":             cookie,
		}).
		Do().
		Into(&resp)

	return resp, nil
}
