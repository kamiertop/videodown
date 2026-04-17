package api

import (
	"fmt"
	"runtime"
	"time"

	"github.com/kamiertop/videodown/douyin/model"
)

// Profile 个人信息接口
// 没找到单独返回个人信息的接口,这个接口返回的第一个是个人信息(没想明白为啥这么设计)
func (d *Douyin) Profile() (model.MyInfo, error) {
	var resp model.MyInfo
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
	q := NewQueryBuilder([]queryKV{
		{"device_platform", "webapp"},
		{"aid", "6383"},
		{"channel", "channel_pc_web"},
		{"count", 10},
		{"source", "coldup"},
		{"max_time", time.Now().Unix()},
		{"min_time", 0},
		{"need_remove_share_panel", true},
		{"need_sorted_info", true},
		{"with_fstatus", 1},
		{"update_version_code", "170400"},
		{"pc_client_type", 1},
		{"pc_libra_divert", "Linux"},
		{"support_h265", 1},
		{"support_dash", 1},
		{"cpu_core_num", runtime.NumCPU()},
		{"version_code", "170400"},
		{"version_name", "17.4.0"},
		{"cookie_enabled", true},
		{"screen_width", 2560},
		{"screen_height", 1440},
		{"browser_language", "zh-CN"},
		{"browser_platform", "Linux x86_64"},
		{"browser_name", "Chrome"},
		{"browser_version", "146.0.0.0"},
		{"browser_online", true},
		{"engine_name", "Blink"},
		{"engine_version", "146.0.0.0"},
		{"os_name", "Linux"},
		{"os_version", "x86_64"},
		{"device_memory", runtime.NumCPU()},
		{"platform", "PC"},
		{"downlink", 0.35},
		{"effective_type", "4g"},
		{"round_trip_time", 50},
		{"webid", webId},
		{"uifid", p.uifid},
		{"verifyFp", p.verifyFp},
		{"fp", p.fp},
		{"msToken", msToken},
	})

	err = d.client.Get("https://www-hj.douyin.com/aweme/v1/web/im/spotlight/relation/").
		SetQueryString(q.BuildEscaped()).
		SetHeaders(map[string]string{
			"Accept":             "application/json, text/plain, */*",
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
			"Uifid":              p.uifid,
			"Cookie":             cookie,
		}).
		Do().
		Into(&resp)

	if err != nil {
		d.logger.Errorf("request info api error: %v", err)
		return resp, err
	}

	return resp, nil
}
