package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dgraph-io/badger/v4"

	"github.com/kamiertop/videodown/bilibili/model"
)

// QRCode 获取二维码, 180秒内有效
func (b *BiliBili) QRCode() (model.QRCodeData, error) {
	var resp struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		TTL     int              `json:"ttl"`
		Data    model.QRCodeData `json:"data"`
	}

	if err := b.client.
		Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate").
		SetQueryParams(map[string]string{
			"source":             "main-fe-header",
			"go_url":             "https://www.bilibili.com/",
			"web_location":       "333.1007",
			"x-bili-locale-json": `{"c_locale":{"language":"zh","region":"CN"},"always_translate":true}`,
		}).
		SetHeaders(map[string]string{
			Accept:          "*/*",
			AcceptLanguage:  "zh-CN,zh;q=0.9",
			Priority:        "u=1, i",
			SecCHUA:         `"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`,
			SecCHUAMobile:   "?0",
			SecCHUAPlatform: `Windows"`,
			SecFetchDest:    "empty",
			SecFetchMode:    "cors",
			SecFetchSite:    "same-site",
			Referer:         BiliBiliUrl,
			UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36",
		}).
		Do().
		Into(&resp); err != nil {
		b.logger.Errorf("failed to get bilibili qrcode: %v", err)
		return model.QRCodeData{}, errors.New("获取二维码失败")
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("bilibili qrcode request failed: code=%d message=%s", resp.Code, resp.Message)
		return model.QRCodeData{}, errors.New("获取二维码失败")
	}

	return resp.Data, nil
}

// PollQRCode 轮询二维码状态, 直到扫码成功或二维码过期
func (b *BiliBili) PollQRCode(qrcodeKey string) (model.PollQRCodeData, error) {
	resp, err := b.client.
		R().
		SetQueryParams(map[string]string{
			"qrcode_key":         qrcodeKey,
			"source":             "main-fe-header",
			webLocation:          "333.1007",
			"x-bili-locale-json": `{"c_locale":{"language":"zh","region":"CN"},"always_translate":true}`,
		}).
		SetHeaders(map[string]string{
			Referer:         BiliBiliUrl,
			SecCHUA:         `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`,
			SecCHUAMobile:   "?0",
			SecCHUAPlatform: `"Windows"`,
			UserAgent:       userAgent(),
		}).
		Get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll")

	if err != nil {
		b.logger.Errorf("failed to poll bilibili qrcode: %v", err)
		return model.PollQRCodeData{}, errors.New("轮询二维码状态失败")
	}

	var res struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    model.PollQRCodeData `json:"data"`
		TTL     int                  `json:"ttl"`
	}
	if err = resp.Into(&res); err != nil {
		b.logger.Errorf("failed to decode bilibili qrcode poll response: %v", err)
		return model.PollQRCodeData{}, errors.New("解析二维码状态响应失败")
	}

	if res.Code != model.SuccessCode {
		b.logger.Errorf("bilibili qrcode poll request failed: code=%d message=%s", res.Code, res.Message)
		return model.PollQRCodeData{}, errors.New("轮询二维码状态失败")
	}

	switch res.Data.Code {
	case model.PollQRCodeStatusSuccess:
		if err := b.saveCookies(resp.Cookies()); err != nil {
			return model.PollQRCodeData{}, err
		}
		return res.Data, nil
	case model.PollQRCodeStatusNotScanned:
		b.logger.Infof("bilibili qrcode not scanned yet: message=%s", res.Data.Message)
		return res.Data, nil
	case model.PollQRCodeStatusScannedUnconfirmed:
		b.logger.Infof("bilibili qrcode scanned but unconfirmed: message=%s", res.Data.Message)
		return res.Data, nil
	case model.PollQRCodeStatusExpired:
		b.logger.Infof("bilibili qrcode expired: message=%s", res.Data.Message)
		return res.Data, nil
	default:
		b.logger.Errorf("unexpected bilibili qrcode status: data_code=%d message=%s", res.Data.Code, res.Data.Message)
		return res.Data, errors.New("未知二维码状态")
	}
}

// IsLoggedIn 检测是否已登录, 通过检查数据库中是否存在有效的 cookies 来判断登录状态
func (b *BiliBili) IsLoggedIn() bool {
	cookies, err := b.getCookies()
	if err != nil {
		b.logger.Errorf("failed to get bilibili cookies: %v", err)
		return false
	}
	b.logger.Info(cookies)

	if _, err = b.getCSRF(); err != nil {
		b.logger.Errorf("failed to get bilibili csrf: %v", err)
		return false
	}

	b.logger.Info("bilibili is logged in")

	return true
}

// getCookies 获取已登录的 cookies
func (b *BiliBili) getCookies() (string, error) {
	var cookie string
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(bilibiliCookieKey))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			cookie = string(val)
			return nil
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return "", errors.New("未登录")
	}
	if err != nil {
		b.logger.Errorf("failed to get bilibili cookies: %v", err)
		return "", errors.New("获取登录信息失败")
	}
	if cookie == "" {
		b.logger.Error("bilibili cookies in db are empty")
		return "", errors.New("未登录")
	}

	return cookie, nil
}

func (b *BiliBili) saveCookies(cookies []*http.Cookie) error {
	cookieMap := make(map[string]string, len(cookies))
	for _, cookie := range cookies {
		cookieMap[cookie.Name] = cookie.Value
	}

	requiredCookies := []string{"SESSDATA", "bili_jct", "DedeUserID", "DedeUserID__ckMd5", "sid"}
	parts := make([]string, 0, len(requiredCookies))
	for _, key := range requiredCookies {
		if value, ok := cookieMap[key]; ok {
			parts = append(parts, key+"="+value)
		}
	}

	if len(parts) == 0 {
		b.logger.Error("no required bilibili cookies found in response")
		return errors.New("保存登录信息失败")
	}

	cookieStr := strings.Join(parts, ";")
	b.logger.Infof("bilibili cookies assembled successfully: %s", cookieStr)

	csrf := cookieMap[bilibiliCSRFKey]
	if csrf == "" {
		b.logger.Error("missing bili_jct in login cookies")
		return errors.New("保存登录信息失败")
	}
	if err := b.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte(bilibiliCSRFKey), []byte(csrf)); err != nil {
			b.logger.Errorf("failed to save bilibili cookies with [bili_jct], cookies: %v,err: %e", cookieMap, err)
			return err
		}
		return txn.Set([]byte(bilibiliCookieKey), []byte(cookieStr))
	}); err != nil {
		b.logger.Errorf("failed to save bilibili cookies: %v", err)
		return errors.New("保存登录信息失败")
	}

	return nil
}

func (b *BiliBili) IsRefresh() (model.RefreshData, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.RefreshData{}, err
	}
	csrf, err := b.getCSRF()
	if err != nil {
		return model.RefreshData{}, err
	}
	var response model.RefreshResponse
	if err = b.client.
		Get("https://passport.bilibili.com/x/passport-login/web/cookie/info").
		SetQueryParam(webLocation, "333.1387").
		SetQueryParam("csrf", csrf).
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookies).
		Do().
		Into(&response); err != nil {
		b.logger.Errorf("check if cookies need to be refreshed error: %v", err)
		return model.RefreshData{}, errors.New("检查Cookie是否需要刷新失败")
	}
	if response.Code != model.SuccessCode {
		b.logger.Errorf("refresh request failed: code=%d message=%s", response.Code, response.Message)
		return model.RefreshData{}, errors.New("检查Cookie是否需要刷新失败")
	}

	return response.Data, nil
}

func (b *BiliBili) LogOut() (model.LogOut, error) {
	defer func() {
		if err := b.clearAuthState(); err != nil {
			b.logger.Errorf("failed to clear auth state: %v", err)
		}
	}()
	cookies, err := b.getCookies()
	if err != nil {
		return model.LogOut{}, err
	}
	csrf, err := b.getCSRF()
	if err != nil {
		return model.LogOut{}, err
	}
	var response model.LogOut
	if err = b.client.
		Post("https://passport.bilibili.com/login/exit/v2").
		SetQueryParam("biliCSRF", csrf).
		SetQueryParam("gourl", BiliBiliUrl).
		SetHeader(Cookie, cookies).
		SetHeader(Referer, BiliBiliOrigin).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Pragma", "no-cache").
		SetHeaders(publicHeaders()).
		Do().
		Into(&response); err != nil {
		b.logger.Errorf("logout request failed: %v", err)
		return model.LogOut{}, errors.New("退出登录失败")
	}
	if response.Code == 2202 {
		return model.LogOut{}, errors.New("CSRF请求非法，可能是因为登录状态无效或已过期")
	}
	if response.Code != model.SuccessCode {
		b.logger.Errorf("logout request failed: code=%d", response.Code)
		return model.LogOut{}, errors.New("退出登录失败")
	}

	return response, nil
}
