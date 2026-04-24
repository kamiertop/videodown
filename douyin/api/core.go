package api

import (
	"net/url"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/wudi/jsonx"

	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
)

const (
	douyinCookieKey = "douyin_cookie"
)

type Douyin struct {
	logger   *logger.Logger
	client   *req.Client
	settings *utils.Settings
	webId    struct {
		value       string
		lastUpdated time.Time
	}
	msToken struct {
		value       string    // msToken的值
		lastUpdated time.Time // 最后更新时间
	}
	secUserID string
	userID    string
}

func New(logger *logger.Logger, settings *utils.Settings) *Douyin {
	logger = logger.WithName("Douyin")
	return &Douyin{
		logger: logger.WithName("Douyin"),
		client: req.
			C().
			SetLogger(logger).
			EnableDebugLog().
			EnableAutoDecompress().
			SetJsonUnmarshal(jsonx.Unmarshal).
			SetJsonMarshal(jsonx.Marshal),
		settings: settings,
	}
}

type cookieParams struct {
	verifyFp string
	fp       string
	uifid    string
}

func (d *Douyin) cookieQuery() (cookieParams, error) {
	cookieToMap := map[string]string{}
	cookie, err := d.GetCookie()
	if err != nil {
		return cookieParams{}, err
	}

	cookies := strings.Split(cookie, ";")
	for _, v := range cookies {
		v = strings.TrimSpace(v)
		if strings.Contains(v, "=") {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) == 2 {
				cookieToMap[parts[0]] = parts[1]
			}
		}
	}
	if d.secUserID == "" {
		rawSecUserId, ok := cookieToMap["FOLLOW_LIVE_POINT_INFO"]
		if ok {
			unescape, err := url.QueryUnescape(rawSecUserId)
			if err != nil {
				d.logger.Errorf("解析 FOLLOW_LIVE_POINT_INFO 失败: %v", err)
			} else {
				d.secUserID = strings.Split(strings.Trim(unescape, `"`), "/")[0]
			}
		} else {
			d.logger.Warning("not found sec user id in cookie, key: FOLLOW_LIVE_POINT_INFO")
		}
	}
	if d.userID == "" {
		rawUserId, ok := cookieToMap["PhoneResumeUidCacheV1"]
		if !ok {
			d.logger.Warning("not found user id in cookie, key: PhoneResumeUidCacheV1")
		} else {
			unescape, err := url.QueryUnescape(rawUserId)
			if err != nil {
				d.logger.Errorf("解析 FOLLOW_LIVE_POINT_INFO 失败: %v", err)
			} else {
				// 这个字段是个json字符串, 解析后取第一个key就是userID了
				var userId map[string]any
				err = jsonx.Unmarshal([]byte(unescape), &userId)
				if err == nil {
					for value := range userId {
						d.userID = value
					}
				}
			}
		}
	}

	return cookieParams{
		verifyFp: cookieToMap["s_v_web_id"],
		fp:       cookieToMap["s_v_web_id"],
		uifid:    cookieToMap["UIFID"],
	}, nil
}
