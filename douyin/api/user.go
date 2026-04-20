package api

import (
	"fmt"
	"runtime"

	"github.com/kamiertop/videodown/douyin/model"
)

func (d *Douyin) User(secUserId string) (model.UserData, error) {
	var resp model.UserData
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
	q := map[string]any{
		"device_platform":             "webapp",
		"aid":                         "6383",
		"channel":                     "channel_pc_web",
		"publish_video_strategy_type": 2,
		"source":                      "channel_pc_web",
		"sec_user_id":                 secUserId,
		"personal_center_strategy":    1,
		"profile_other_record_enable": 1,
		"land_to":                     1,
		"update_version_code":         "170400",
		"pc_client_type":              1,
		"pc_libra_divert":             "Linux",
		"support_h265":                1,
		"support_dash":                1,
		"cpu_core_num":                runtime.NumCPU(),
		"version_code":                "170400",
		"version_name":                "17.4.0",
		"cookie_enabled":              true,
		"screen_width":                2560,
		"screen_height":               1440,
		"browser_language":            "zh-CN",
		"browser_platform":            "Linux x86_64",
		"browser_name":                "Chrome",
		"browser_version":             "146.0.0.0",
		"browser_online":              true,
		"engine_name":                 "Blink",
		"engine_version":              "146.0.0.0",
		"os_name":                     "Linux",
		"os_version":                  "x86_64",
		"device_memory":               runtime.NumCPU(),
		"platform":                    "PC",
		"downlink":                    0.4,
		"effective_type":              "4g",
		"round_trip_time":             50,
		"webid":                       webId,
		"uifid":                       p.uifid,
		"verifyFp":                    p.verifyFp,
		"fp":                          p.fp,
		"msToken":                     msToken,
	}

	err = d.client.Get("https://www-hj.douyin.com/aweme/v1/web/user/profile/other/").
		SetQueryParamsAnyType(q).
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

//func (d *Douyin) UserVideoList(secUserId string, count int) (model.UserVideoList, error) {
//	var resp model.UserVideoList
//	p, err := d.cookieQuery()
//	if err != nil {
//		return resp, err
//	}
//	cookie, err := d.getCookie()
//	if err != nil {
//		return resp, err
//	}
//	webId, err := d.getWebId()
//	if err != nil {
//		return resp, err
//	}
//	msToken, err := d.getMsToken()
//	if err != nil {
//		return resp, err
//	}
//	err = d.client.
//		Get("https://www-hj.douyin.com/aweme/v1/web/aweme/post/").
//		SetQueryParamsAnyType(map[string]any{
//			"device_platform":             "webapp",
//			"aid":                         "6383",
//			"channel":                     "channel_pc_web",
//			"sec_user_id":                 secUserId,
//			"max_cursor":                  0,
//			"locate_query":                false,
//			"show_live_replay_strategy":   1,
//			"need_time_list":              1,
//			"time_list_query":             0,
//			"whale_cut_token":             "",
//			"cut_version":                 1,
//			"count":                       count,
//			"publish_video_strategy_type": 2,
//			"update_version_code":         "170400",
//			"source_type":                 4,
//			"gps_access":                  0,
//			"is_top":                      0,
//			"pc_client_type":              1,
//			"pc_libra_divert":             "Linux",
//			"support_h265":                1,
//			"support_dash":                1,
//			"webcast_sdk_version":         "170400",
//			"webcast_version_code":        "170400",
//			"version_code":                "290100",
//			"version_name":                "29.1.0",
//			"screen_width":                2560,
//			"screen_height":               1440,
//			"cookie_enabled":              true,
//			"browser_language":            "zh-CN",
//			"browser_platform":            "Linux x86_64",
//			"browser_name":                "Chrome",
//			"browser_version":             "146.0.0.0",
//			"browser_online":              true,
//			"engine_name":                 "Blink",
//			"engine_version":              "146.0.0.0",
//			"os_name":                     "Linux",
//			"os_version":                  "x86_64",
//			"device_memory":               runtime.NumCPU(),
//			"platform":                    "PC",
//			"downlink":                    0.35,
//			"effective_type":              "4g",
//			"round_trip_time":             50,
//			"webid":                       webId,
//			"uifid":                       p.uifid,
//			"verifyFp":                    p.verifyFp,
//			"fp":                          p.fp,
//			"msToken":                     msToken,
//		}).
//		SetHeaders(map[string]string{
//			"Accept":             "application/json, text/plain, */*",
//			"Accept-Language":    "zh-CN,zh;q=0.9",
//			"Accept-Encoding":    "gzip, deflate, br, zstd",
//			"Origin":             "https://www.douyin.com/",
//			"Referer":            "https://www.douyin.com/",
//			"Priority":           "u=1, i",
//			"Sec-CH-UA":          `"Google Chrome";v="146", "Not:A-Brand";v="24", "Chromium";v="146"`,
//			"Sec-CH-UA-Mobile":   "?0",
//			"Sec-CH-UA-Platform": `"Linux"`,
//			"Sec-Fetch-Dest":     "empty",
//			"Sec-Fetch-Mode":     "cors",
//			"Sec-Fetch-Site":     "same-site",
//			"User-Agent":         userAgent(),
//			"Uifid":              p.uifid,
//			"Cookie":             cookie,
//		}).Do().
//		Into(&resp)
//	if err != nil {
//		d.logger.Errorf("request user video list api error: %v", err)
//		return resp, errors.New("请求用户视频列表失败")
//	}
//	if resp.StatusCode != 0 {
//		d.logger.Errorf("request user video list api error, status code: %d", resp.StatusCode)
//		return resp, errors.New("请求用户视频列表失败")
//	}
//
//	return resp, nil
//}
