package api

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/dgraph-io/badger/v4"
	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
)

const (
	bilibiliCookieKey       = "bilibili_cookies"
	bilibiliCSRFKey         = "bili_jct"
	bilibiliMidKey          = "bilibili_mid"
	bilibiliRefreshTokenKey = "bilibili_refresh_token"
)

type BiliBili struct {
	logger         *logger.Logger
	client         *req.Client
	downloadClient *req.Client
	settings       *utils.Settings
	wbiKey         *wbiKeys // lazy init
	progressMu     sync.Mutex
	progressByBvid map[string]float64
}

func New(log *logger.Logger, settings *utils.Settings) *BiliBili {
	log = log.WithName("BiliBili")
	var client = req.C()
	if logger.IsDevMode() {
		client = client.EnableDebugLog()
	}
	return &BiliBili{
		logger: log.WithName("BiliBili").WithCaller(3),
		downloadClient: req.
			C().
			SetTimeout(0).
			SetLogger(log).
			EnableAutoDecompress().
			SetJsonMarshal(sonic.Marshal).
			SetJsonUnmarshal(sonic.Unmarshal),
		// 普通接口请求使用有默认超时的 client；下载流单独走 downloadClient，避免长视频下载受接口超时影响。
		client: client.
			SetLogger(log).
			EnableAutoDecompress().
			SetJsonMarshal(sonic.Marshal).
			SetJsonUnmarshal(sonic.Unmarshal),
		settings:       settings,
		progressByBvid: make(map[string]float64),
	}
}

func (b *BiliBili) context() context.Context {
	return b.settings.Context()
}

func (b *BiliBili) getCSRF() (string, error) {
	return b.settings.GetKey(bilibiliCSRFKey)
}

func (b *BiliBili) saveMid(mid uint64) error {
	return b.settings.SetKey(bilibiliMidKey, strconv.FormatUint(mid, 10))
}

func (b *BiliBili) getMid() (string, error) {
	return b.settings.GetKey(bilibiliMidKey)
}

func (b *BiliBili) clearAuthState() error {
	keys := []string{bilibiliCookieKey, bilibiliCSRFKey, bilibiliMidKey, bilibiliRefreshTokenKey}

	return b.settings.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			err := txn.Delete([]byte(key))
			if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
		}

		return nil
	})
}
