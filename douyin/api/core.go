package api

import (
	"net/url"
	"strings"
	"time"

	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
)

type Douyin struct {
	logger *logger.Logger
	client *req.Client
	store  *storage.Store
	webId  struct {
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

func New(log *logger.Logger, store *storage.Store) *Douyin {
	var client = req.C().
		EnableAutoDecompress().
		SetTimeout(20*time.Second).
		SetCommonRetryCount(2).
		SetCommonRetryBackoffInterval(300*time.Millisecond, 2*time.Second)

	if logger.IsDevMode() {
		client.SetLogger(log).EnableDebugLog()
	}

	return &Douyin{
		logger: log.WithName("Douyin"),
		client: client,
		store:  store,
	}
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

	cookies := strings.SplitSeq(cookie, ";")
	for v := range cookies {
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

	return cookieParams{
		verifyFp: cookieToMap["s_v_web_id"],
		fp:       cookieToMap["s_v_web_id"],
		uifid:    cookieToMap["UIFID"],
	}, nil
}
