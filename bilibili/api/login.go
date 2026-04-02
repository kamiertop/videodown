package api

import "github.com/kamiertop/videodown/bilibili/model"

func (b *BiliBili) QRCode() (model.QRCodeData, error) {
	resp, err := b.client.
		R().
		SetQueryParams(map[string]string{
			"source":             "main-fe-header",
			"go_url":             "https://www.bilibili.com/",
			"web_location":       "333.1007",
			"x-bili-locale-json": `{"c_locale":{"language":"zh","region":"CN"},"always_translate":true}`,
		}).
		SetHeaders(map[string]string{
			"accept":             "*/*",
			"accept-language":    "zh-CN,zh;q=0.9",
			"priority":           "u=1, i",
			"sec-ch-ua":          `"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": `Windows"`,
			"sec-fetch-dest":     "empty",
			"sec-fetch-mode":     "cors",
			"sec-fetch-site":     "same-site",
			"referer":            "https://www.bilibili.com/",
			"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36",
		}).
		Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")

	if err != nil {

		return model.QRCodeData{}, err
	}
	var res model.QRCodeResponse
	if err := resp.Into(&res); err != nil {
		return model.QRCodeData{}, err
	}
	if res.Code != model.SuccessCode {
		return model.QRCodeData{}, err
	}

	return res.Data, nil
}
