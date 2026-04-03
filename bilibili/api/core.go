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
		logger: logger,
		client: req.C().SetLogger(logger).EnableDebugLog().EnableAutoDecompress(),
		db:     db,
	}
}
func (b *BiliBili) getCSRF() (string, error) {
	var csrf string
	return csrf, b.db.View(func(txn *badger.Txn) error {
		value, err := txn.Get([]byte(bilibiliCSRFKey))
		if err != nil {
			return err
		}
		return value.Value(func(val []byte) error {
			csrf = string(val)
			return nil
		})
	})
}

func (b *BiliBili) saveMid(mid uint64) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(bilibiliMidKey), []byte(strconv.FormatUint(mid, 10)))
	})
}

func (b *BiliBili) getMid() (string, error) {
	var value string
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(bilibiliMidKey))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			value = string(val)
			return nil
		})
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return "", errors.New("未找到用户ID")
		}
		return "", errors.New("获取用户ID失败")
	}

	return value, nil
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
