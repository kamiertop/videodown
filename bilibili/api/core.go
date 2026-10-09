package api

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/bilibili/model"
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
	logger *logger.Logger
	client *req.Client
	store  *storage.Store
	events *application.EventManager
	wbiKey *wbiKeys // lazy init
	// 下载历史的内存缓存：按下载时间倒序。删除/清空在本包可直接失效；
	// 写入发生在 download 包（无法跨包通知），用短 TTL 兜底新记录的可见性。
	historyMu    sync.RWMutex
	historyItems []model.DownloadHistoryItem
	historyValid bool
	historyStamp time.Time
}

func New(log *logger.Logger, store *storage.Store, events *application.EventManager) *BiliBili {
	var client = req.C().
		EnableAutoDecompress().
		SetCommonRetryCount(2).
		SetCommonRetryBackoffInterval(300*time.Millisecond, 2*time.Second)
	if logger.IsDevMode() {
		client.SetLogger(log).EnableDebugLog()
	}
	return &BiliBili{
		logger: log.WithName("BiliBili"),
		client: client,
		store:  store,
		events: events,
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

func (b *BiliBili) CookieFunc() func() (string, error) {
	return b.getCookies
}
