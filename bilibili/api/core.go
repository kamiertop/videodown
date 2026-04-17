package api

import (
	"context"
	"errors"
	"fmt"
	"runtime"
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

func publicHeaders() map[string]string {
	return map[string]string{
		AcceptEncoding:  "gzip, deflate, br, zstd",
		AcceptLanguage:  "zh-CN,zh;q=0.9",
		Accept:          "*/*",
		SecCHUAMobile:   "?0",
		Priority:        "u=1, i",
		SecCHUA:         `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`,
		SecCHUAPlatform: fmt.Sprintf(`"%s"`, runtime.GOOS),
		SecFetchMode:    "cors",
		SecFetchDest:    "empty",
		SecFetchSite:    "same-site",
		UserAgent:       userAgent(),
	}
}

func userAgent() string {
	if runtime.GOOS == "windows" {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	}

	return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
}
