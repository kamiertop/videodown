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
	settings       *utils.Settings
	wbiKey         *wbiKeys // lazy init
	progressMu     sync.Mutex
	progressByBvid map[string]float64
}

func New(logger *logger.Logger, settings *utils.Settings) *BiliBili {
	logger = logger.WithName("BiliBili")

	return &BiliBili{
		logger: logger.WithName("BiliBili").WithCaller(3),
		client: req.
			C().
			SetLogger(logger).
			EnableDebugLog().
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
