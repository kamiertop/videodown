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

// emitPlayUrlResolved 通过 Wails 事件把单条解析结果推送给前端，
// 前端按 bvid+cid 更新对应卡片的解析状态，无需等待整批解析完成。
func (b *BiliBili) emitPlayUrlResolved(r PlayUrlResult) {
	if b.events != nil {
		b.events.Emit("bilibili-playurl-resolved", r)
	}
}

// BatchResolvePlayUrl 批量解析播放地址。单条失败只写入该条结果，不中断整批解析。
// 每条结果解析完成后会通过 bilibili-playurl-resolved 事件推送给前端，实现渐进式 UI 更新。
func (b *BiliBili) BatchResolvePlayUrl(reqs []PlayUrlRequest) []PlayUrlResult {
	results := make([]PlayUrlResult, len(reqs))
	var wg sync.WaitGroup
	for i, req := range reqs {
		wg.Go(func() {
			results[i] = b.ResolvePlayUrl(req, len(reqs))
			// 每条解析完成后立即推送事件，前端逐条更新卡片状态
			b.emitPlayUrlResolved(results[i])
		})
	}
	wg.Wait()

	return results
}

// ResolvePlayUrl 解析单个视频的播放地址，内部由模块级信号量控制并发。
// totalCount 为本次批量解析的总视频数，当数量不超过并发上限时跳过休眠以提升体验。
func (b *BiliBili) ResolvePlayUrl(req PlayUrlRequest, totalCount int) PlayUrlResult {
	b.acquireSem()
	defer b.releaseSem()
	// 仅当批量视频数超过并发上限时才休眠，拉开不同请求的时间间隔，防止风控连坐。
	// 少量视频（≤ 并发数）全部在同一窗口内发出，睡不睡效果相同，直接跳过。
	if totalCount > b.settings.GetParsePlayURLNumSafe() {
		if maxSleep := b.settings.GetParsePlayURLSleepSafe(); maxSleep > 0 {
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
