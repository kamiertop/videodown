package api

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/logger"
)

const (
	bilibiliCookieKey = "bilibili_cookies"
	bilibiliCSRFKey   = "bili_jct"
	bilibiliMidKey    = "bilibili_mid"
)

type BiliBili struct {
	logger *logger.Logger
	client *req.Client
	db     *badger.DB
}

func New(logger *logger.Logger, db *badger.DB) *BiliBili {
	logger = logger.WithName("BiliBili")
	return &BiliBili{
		logger: logger.WithCaller(2),
		client: req.C().SetLogger(logger).EnableDebugLog().EnableAutoDecompress(),
		db:     db,
	}
}
func (b *BiliBili) getCSRF() (string, error) {
	return b.getKey(bilibiliCSRFKey)
}

func (b *BiliBili) saveMid(mid uint64) error {
	return b.setKey(bilibiliMidKey, strconv.FormatUint(mid, 10))
}

func (b *BiliBili) getMid() (string, error) {
	return b.getKey(bilibiliMidKey)
}

func (b *BiliBili) clearAuthState() error {
	keys := []string{bilibiliCookieKey, bilibiliCSRFKey, bilibiliMidKey}

	return b.db.Update(func(txn *badger.Txn) error {
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
		Origin:          BiliBiliOrigin,
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

func (b *BiliBili) getKey(key string) (string, error) {
	var value string

	return value, b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			value = string(val)
			return nil
		})
	})
}

func (b *BiliBili) setKey(key, value string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}
