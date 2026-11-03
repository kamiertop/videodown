package api

import (
	"math/rand"
	"sync"
	"time"

	"github.com/kamiertop/videodown/bilibili/model"
)

type PlayUrlRequest struct {
	Bvid string `json:"bvid"`
	Cid  int64  `json:"cid"`
	Qn   int    `json:"qn"`
}

type PlayUrlResult struct {
	Bvid       string                        `json:"bvid"`
	Cid        int64                         `json:"cid"`
	RequestCid int64                         `json:"requestCid"`
	Error      string                        `json:"error,omitempty"`
	Detail     *model.VideoDetailConciseData `json:"detail,omitempty"`
	PlayURL    *model.VideoURLData           `json:"play_url,omitempty"`
}

// resolveLimiter 限制同时解析播放地址的请求数。相比固定容量的 channel
// 信号量，它可以在请求进行中调整容量，让设置改动立即生效。
type resolveLimiter struct {
	mu       sync.Mutex
	cond     *sync.Cond
	inFlight int
	limit    int
}

func newResolveLimiter(limit int) *resolveLimiter {
	l := &resolveLimiter{limit: max(1, limit)}
	l.cond = sync.NewCond(&l.mu)
	return l
}

// enter 获取一个槽位；waited 表示本次曾在其他请求之后排队。
func (l *resolveLimiter) enter() (waited bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for l.inFlight >= l.limit {
		waited = true
		l.cond.Wait()
	}
	l.inFlight++
	return waited
}

func (l *resolveLimiter) leave() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.inFlight--
	l.cond.Signal()
}

// resize 应用新的并发上限；正在排队的请求立即按新上限唤醒。
func (l *resolveLimiter) resize(limit int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if limit <= 0 || limit == l.limit {
		return
	}
	l.limit = limit
	l.cond.Broadcast()
}

// ResolvePlayUrl 解析单个视频的播放地址，内部由限流器控制并发：
// 前若干个请求（并发上限以内）立即发出，排过队的请求在获得槽位后
// 随机休眠再发起，拉开批量请求的时间间隔，防止风控连坐。
func (b *BiliBili) ResolvePlayUrl(req PlayUrlRequest) PlayUrlResult {
	b.resolve.resize(b.getParsePlayURLNumSafe())
	waited := b.resolve.enter()
	defer b.resolve.leave()
	if waited {
		if maxSleep := b.getParsePlayURLSleepSafe(); maxSleep > 0 {
			d := rand.Float64() * float64(maxSleep)
			b.logger.Infof("Sleeping for %.1fs before resolving %s", d, req.Bvid)
			time.Sleep(time.Duration(d * float64(time.Second)))
		}
	}

	r := PlayUrlResult{Bvid: req.Bvid, Cid: req.Cid, RequestCid: req.Cid}

	detail, err := b.VideoDetailConciseBvid(req.Bvid)
	if err != nil {
		r.Error = err.Error()
		b.logger.Errorf("Failed to resolve video detail for BVID %s: %v", req.Bvid, err)
		return r
	}

	view := detail.View
	aid := view.Aid
	cid := req.Cid
	if cid <= 0 {
		cid = view.Cid
	}
	r.Cid = cid
	qn := req.Qn
	if qn <= 0 {
		qn = 80
	}

	play, err := b.VideoPlayURL(aid, req.Bvid, cid, qn)
	if err != nil {
		r.Error = err.Error()
		b.logger.Errorf("Failed to resolve play URL for BVID %s: %v", req.Bvid, err)
		return r
	}
	b.logger.Infof("Successfully resolved play URL for BVID %s", req.Bvid)
	r.Detail = &detail
	r.PlayURL = &play

	return r
}
