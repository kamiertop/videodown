package api

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
)

const (
	bilibiliCookieKey = "bilibili_cookies"
	bilibiliCSRFKey   = "bili_jct"
	bilibiliMidKey    = "bilibili_mid"
)

type BiliBili struct {
	ctxProvider    func() context.Context
	logger         *logger.Logger
	client         *req.Client
	settings       *utils.Settings
	wbiKey         *wbiKeys // lazy init
	progressMu     sync.Mutex
	progressByBvid map[string]float64
}

func New(logger *logger.Logger, db *utils.Settings, ctxProvider ...func() context.Context) *BiliBili {
	logger = logger.WithName("BiliBili")
	var provider func() context.Context
	if len(ctxProvider) > 0 {
		provider = ctxProvider[0]
	}
	return &BiliBili{
		ctxProvider:    provider,
		logger:         logger.WithCaller(2),
		client:         req.C().SetLogger(logger).EnableDebugLog().EnableAutoDecompress(),
		settings:       db,
		progressByBvid: make(map[string]float64),
	}
}

func (b *BiliBili) context() context.Context {
	if b.ctxProvider == nil {
		return nil
	}
	return b.ctxProvider()
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
	keys := []string{bilibiliCookieKey, bilibiliCSRFKey, bilibiliMidKey}

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
