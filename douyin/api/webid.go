package api

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

func userAgent() string {
	if runtime.GOOS == "windows" {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	}

	return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
}

func (d *Douyin) getWebId() (string, error) {
	// TODO: 校验是否需要刷新
	var resp struct {
		E     int    `json:"e"`
		WebId string `json:"web_id"`
	}
	err := d.client.
		Post("https://mcs.zijieapi.com/webid").
		SetBodyJsonMarshal(map[string]any{
			"app_id":         6383,
			"referer":        "https://www.douyin.com/",
			"url":            "https://so-landing.douyin.com/search_ai_mobile/pc?aid=6383&theme=dark&enter_from=ai_search",
			"user_agent":     userAgent(),
			"user_unique_id": "",
		}).
		SetHeaders(map[string]string{
			"Accept":             "*/*",
			"Accept-Language":    "zh-CN,zh;q=0.9",
			"Content-Type":       "application/json; charset=UTF-8",
			"Origin":             "https://so-landing.douyin.com",
			"Referer":            "https://so-landing.douyin.com",
			"Priority":           "u=1, i",
			"Sec-CH-UA":          `"Google Chrome";v="146", "Not:A-Brand";v="24", "Chromium";v="146"`,
			"Sec-CH-UA-Mobile":   "?0",
			"Sec-CH-UA-Platform": fmt.Sprintf(`"%s"`, runtime.GOOS),
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-site",
			"User-Agent":         userAgent(),
		}).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Error(fmt.Sprintf("request web_id api error: %v", err))
		return "", errors.New("获取web_id失败")
	}
	if resp.WebId == "" {
		return "", errors.New("获取web_id失败")
	}
	d.webId.value = resp.WebId
	d.webId.lastUpdated = time.Now()

	return resp.WebId, nil
}
