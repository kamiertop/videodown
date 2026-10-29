package api

import (
	"strconv"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
)

const bilibiliCSRFKey = "bili_jct"

type BiliBili struct {
	logger *logger.Logger
	client *req.Client
	store  *storage.Store

	wbiKey *wbiKeys // lazy init
	wbiMu  sync.Mutex
}

func New(log *logger.Logger, store *storage.Store) *BiliBili {
	var client = req.C().
		EnableAutoDecompress().
		SetTimeout(20*time.Second).
		SetCommonRetryCount(2).
		SetCommonRetryBackoffInterval(300*time.Millisecond, 2*time.Second)
	if logger.IsDevMode() {
		client.SetLogger(log)
	}
	return &BiliBili{
		logger: log.WithName("BiliBili"),
		client: client,
		store:  store,
	}
}

func (b *BiliBili) getParsePlayURLNumSafe() int   { return b.store.Config().Bilibili.ParseConcurrency }
func (b *BiliBili) getParsePlayURLSleepSafe() int { return b.store.Config().Bilibili.ParseSleepSeconds }

func (b *BiliBili) getCSRF() (string, error) {
	value := b.store.Config().Bilibili.Auth.CSRF
	if value == "" {
		return "", storage.ErrNotFound
	}
	return value, nil
}

func (b *BiliBili) saveMid(mid uint64) error {
	return b.store.UpdateConfig(func(c *storage.Config) error { c.Bilibili.Auth.UID = mid; return nil })
}

func (b *BiliBili) getMid() (string, error) {
	uid := b.store.Config().Bilibili.Auth.UID
	if uid == 0 {
		return "", storage.ErrNotFound
	}
	return strconv.FormatUint(uid, 10), nil
}

func (b *BiliBili) clearAuthState() error {
	return b.store.UpdateConfig(func(c *storage.Config) error { c.Bilibili.Auth = storage.BilibiliAuth{}; return nil })
}

func (b *BiliBili) CookieFunc() func() (string, error) {
	return b.getCookies
}
