package api

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
)

const (
	douyinCookieKey = "douyin_cookie"
)

type Douyin struct {
	logger         *logger.Logger
	client         *req.Client
	downloadClient *req.Client
	settings       *utils.Settings
	progressMu     sync.Mutex
	progressByID   map[string]float64
	webId          struct {
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

func New(log *logger.Logger, settings *utils.Settings) *Douyin {
	log = log.WithName("Douyin")
	var client = req.C().SetLogger(log).EnableAutoDecompress().SetJsonMarshal(sonic.Marshal).SetJsonUnmarshal(sonic.Unmarshal)
	if logger.IsDevMode() {
		client = client.EnableDebugLog()
	}
	return &Douyin{
		logger:         log.WithName("Douyin").WithCaller(3),
		client:         client,
		downloadClient: client.Clone().SetTimeout(0), // 下载流单独走 downloadClient，避免长视频下载受超时影响
		settings:       settings,
		progressByID:   make(map[string]float64),
	}
}

func (d *Douyin) context() context.Context {
	return d.settings.Context()
}

type cookieParams struct {
	verifyFp string
	fp       string
	uifid    string
}

// cookieQuery 从cookie中提取公共查询参数，主要是verifyFp和uifid
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
		tryKeys := [2]string{"FOLLOW_LIVE_POINT_INFO", "FOLLOW_NUMBER_YELLOW_POINT_INFO"}
		for _, key := range tryKeys {
			if rawSecUserId, ok := cookieToMap[key]; ok {
				d.logger.Infof("found %s: %s", key, rawSecUserId)
				unescape, err := url.QueryUnescape(rawSecUserId)
				if err != nil {
					d.logger.Errorf("解析 %s 失败: %v", key, err)
				} else {
					d.secUserID = strings.Split(strings.Trim(unescape, `"`), "/")[0]
				}
				d.logger.Infof("secUserID: %s", d.secUserID)
			} else {
				d.logger.Warningf("not found %s in cookie", key)
			}
		}
	}

	//if d.userID == "" {
	//	rawUserId, ok := cookieToMap["PhoneResumeUidCacheV1"]
	//	if !ok {
	//		d.logger.Warning("not found user id in cookie, key: PhoneResumeUidCacheV1")
	//	} else {
	//		unescape, err := url.QueryUnescape(rawUserId)
	//		if err != nil {
	//			d.logger.Errorf("解析 FOLLOW_LIVE_POINT_INFO 失败: %v", err)
	//		} else {
	//			// 这个字段是个json字符串, 解析后取第一个key就是userID了
	//			var userId map[string]any
	//			err = sonic.Unmarshal([]byte(unescape), &userId)
	//			if err == nil {
	//				for value := range userId {
	//					d.userID = value
	//				}
	//			}
	//		}
	//	}
	//}

	return cookieParams{
		verifyFp: cookieToMap["s_v_web_id"],
		fp:       cookieToMap["s_v_web_id"],
		uifid:    cookieToMap["UIFID"],
	}, nil
}
