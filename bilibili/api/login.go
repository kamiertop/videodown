package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/header"
)

// QRCode 获取二维码, 180秒内有效
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
			header.Accept:          "*/*",
			header.AcceptLanguage:  "zh-CN,zh;q=0.9",
			header.Priority:        "u=1, i",
			header.SecCHUA:         `"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`,
			header.SecCHUAMobile:   "?0",
			header.SecCHUAPlatform: `Windows"`,
			header.SecFetchDest:    "empty",
			header.SecFetchMode:    "cors",
			header.SecFetchSite:    "same-site",
			header.Referer:         header.BiliBiliUrl,
			header.UserAgent:       header.UserAgentValue,
		}).
		Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")

	if err != nil {
		b.logger.Errorf("failed to get bilibili qrcode: %v", err)
		return model.QRCodeData{}, errors.New("获取二维码失败")
	}
	var res model.QRCodeResponse
	if err := resp.Into(&res); err != nil {
		b.logger.Errorf("failed to decode bilibili qrcode response: %v", err)
		return model.QRCodeData{}, errors.New("解析二维码响应失败")
	}
	if res.Code != model.SuccessCode {
		b.logger.Errorf("bilibili qrcode request failed: code=%d message=%s", res.Code, res.Message)
		return model.QRCodeData{}, errors.New("获取二维码失败")
	}

	return res.Data, nil
}

// PollQRCode 轮询二维码状态, 直到扫码成功或二维码过期
func (b *BiliBili) PollQRCode(qrcodeKey string) (model.PollQRCodeData, error) {
	resp, err := b.client.
		R().
		SetQueryParams(map[string]string{
			"qrcode_key":         qrcodeKey,
			"source":             "main-fe-header",
			"web_location":       "333.1007",
			"x-bili-locale-json": `{"c_locale":{"language":"zh","region":"CN"},"always_translate":true}`,
		}).
		SetHeaders(map[string]string{
			header.Referer:         header.BiliBiliUrl,
			header.SecCHUA:         `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`,
			header.SecCHUAMobile:   "?0",
			header.SecCHUAPlatform: `"Windows"`,
			header.UserAgent:       header.UserAgentValue,
		}).
		Get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll")

	if err != nil {
		b.logger.Errorf("failed to poll bilibili qrcode: %v", err)
		return model.PollQRCodeData{}, errors.New("轮询二维码状态失败")
	}

	var res model.PollQRCodeResponse
	if err := resp.Into(&res); err != nil {
		b.logger.Errorf("failed to decode bilibili qrcode poll response: %v", err)
		return model.PollQRCodeData{}, errors.New("解析二维码状态响应失败")
	}

	if res.Code != model.SuccessCode || res.Message != "OK" {
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
		return res.Data, errors.New("二维码未扫码")
	case model.PollQRCodeStatusScannedUnconfirmed:
		b.logger.Infof("bilibili qrcode scanned but unconfirmed: message=%s", res.Data.Message)
		return res.Data, errors.New("二维码已扫码但未确认")
	case model.PollQRCodeStatusExpired:
		b.logger.Infof("bilibili qrcode expired: message=%s", res.Data.Message)
		return res.Data, errors.New("二维码已失效")
	default:
		b.logger.Errorf("unexpected bilibili qrcode status: data_code=%d message=%s", res.Data.Code, res.Data.Message)
		return res.Data, errors.New("未知二维码状态")
	}
}

// IsLoggedIn 检测是否已登录, 通过检查数据库中是否存在有效的 cookies 来判断登录状态
func (b *BiliBili) IsLoggedIn() bool {
	_, err := b.GetCookies()
	return err == nil
}

// GetCookies 获取已登录的 cookies
func (b *BiliBili) GetCookies() (string, error) {
	const bilibiliCookieKey = "bilibili_cookies"

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
	const bilibiliCookieKey = "bilibili_cookies"

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

	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(bilibiliCookieKey), []byte(cookieStr))
	}); err != nil {
		b.logger.Errorf("failed to save bilibili cookies: %v", err)
		return errors.New("保存登录信息失败")
	}

	return nil
}
