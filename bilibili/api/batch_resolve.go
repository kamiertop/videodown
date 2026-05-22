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
	Bvid    string                        `json:"bvid"`
	Cid     int64                         `json:"cid"`
	Error   string                        `json:"error,omitempty"`
	Detail  *model.VideoDetailConciseData `json:"detail,omitempty"`
	PlayURL *model.VideoURLData           `json:"play_url,omitempty"`
}

// playUrlSem 是模块级的并发控制通道，由 ensureSem 懒初始化
var playUrlSem chan struct{}
var playUrlSemOnce sync.Once

func (b *BiliBili) ensureSem() {
	playUrlSemOnce.Do(func() {
		playUrlSem = make(chan struct{}, b.settings.GetParsePlayURLNumSafe())
	})
}

func (b *BiliBili) acquireSem() {
	b.ensureSem()
	playUrlSem <- struct{}{}
}

func (b *BiliBili) releaseSem() {
	<-playUrlSem
}

// BatchResolvePlayUrl 批量解析播放地址。单条失败只写入该条结果，不中断整批解析。
func (b *BiliBili) BatchResolvePlayUrl(reqs []PlayUrlRequest) []PlayUrlResult {
	results := make([]PlayUrlResult, len(reqs))
	var wg sync.WaitGroup
	for i, req := range reqs {
		i, req := i, req
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = b.ResolvePlayUrl(req)
		}()
	}
	wg.Wait()

	return results
}

// ResolvePlayUrl 解析单个视频的播放地址，内部由模块级信号量控制并发
func (b *BiliBili) ResolvePlayUrl(req PlayUrlRequest) PlayUrlResult {
	b.acquireSem()
	defer b.releaseSem()
	// 休眠放在 API 调用之前，拉开不同请求的时间间隔，防止风控连坐
	if maxSleep := b.settings.GetParsePlayURLSleepSafe(); maxSleep > 0 {
		d := rand.Float64() * float64(maxSleep)
		b.logger.Infof("Sleeping for %vs before resolving %s", d, req.Bvid)
		time.Sleep(time.Duration(d * float64(time.Second)))
	}

	r := PlayUrlResult{Bvid: req.Bvid, Cid: req.Cid}

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
