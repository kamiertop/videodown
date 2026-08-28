package api

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
	"github.com/wailsapp/wails/v3/pkg/application"
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
	store          *storage.Store
	events         *application.EventManager
	wbiKey         *wbiKeys // lazy init
	progressMu     sync.Mutex
	progressByBvid map[string]float64
}

func New(log *logger.Logger, store *storage.Store, events *application.EventManager) *BiliBili {
	var client = req.C().EnableAutoDecompress().
		SetCommonRetryCount(2).
		SetCommonRetryBackoffInterval(300*time.Millisecond, 2*time.Second)
	if logger.IsDevMode() {
		client.SetLogger(log).EnableDebugLog()
	}
	return &BiliBili{
		logger:         log.WithName("BiliBili"),
		downloadClient: client.Clone().SetTimeout(0), // 下载流单独走 downloadClient，避免长视频下载受超时影响
		client:         client,
		store:          store,
		events:         events,
		progressByBvid: make(map[string]float64),
	}
}

func (b *BiliBili) getParsePlayURLNumSafe() int {
	value, err := b.store.ParsePlayURLNum()
	if err != nil || value <= 0 {
		return 3
	}
	return value
}

func (b *BiliBili) getParsePlayURLSleepSafe() int {
	value, err := b.store.ParsePlayURLSleep()
	if err != nil || value < 0 {
		return 5
	}
	return value
}

func (b *BiliBili) getCSRF() (string, error) {
	return b.store.Get(bilibiliCSRFKey)
}

func (b *BiliBili) saveMid(mid uint64) error {
	return b.store.Set(bilibiliMidKey, strconv.FormatUint(mid, 10))
}

func (b *BiliBili) getMid() (string, error) {
	return b.store.Get(bilibiliMidKey)
}

func (b *BiliBili) clearAuthState() error {
	keys := []string{bilibiliCookieKey, bilibiliCSRFKey, bilibiliMidKey, bilibiliRefreshTokenKey}

	return b.store.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			err := txn.Delete([]byte(key))
			if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
		}

		return nil
	})
}
