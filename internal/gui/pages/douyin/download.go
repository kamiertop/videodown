package douyin

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	dyDownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/douyin/model"
)

type queuedVideo struct {
	Data          model.AwemeItem
	CoverBusy     bool
	Track         int
	Status, Error string
	Progress      dyDownload.Progress
	ResolvedAt    time.Time
}
type downloadGroup struct {
	ID, Order, Version             uint64
	Title, Status, Error           string
	Open, Stop, Loading, More      bool
	Items                          []*queuedVideo
	Source                         source
	Cursor, InFlight, Done, Failed int
	Seen                           map[string]bool
}

func first(urls []string) string {
	for _, u := range urls {
		if strings.HasPrefix(u, "//") {
			u = "https:" + u
		}
		if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
			return u
		}
	}
	return ""
}
func title(v model.AwemeItem) string {
	for _, s := range []string{v.ItemTitle, v.Caption, v.Desc, v.AwemeId} {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return "未命名作品"
}
func cover(v model.AwemeItem) string {
	for _, c := range []model.Cover{v.Video.Cover, v.Video.OriginCover, v.Video.DynamicCover} {
		if u := first(c.UrlList); u != "" {
			return u
		}
	}
	if len(v.Images) > 0 {
		return first(v.Images[0].UrlList)
	}
	return ""
}
func bestTrack(v model.AwemeItem) int {
	index, rate := -1, -1
	for i, track := range v.Video.BitRate {
		if first(track.PlayAddr.UrlList) != "" && track.BitRate > rate {
			index, rate = i, track.BitRate
		}
	}
	return index
}
func taskFor(v model.AwemeItem, track int, sourceName string) (dyDownload.Task, error) {
	task := dyDownload.Task{AwemeID: v.AwemeId, SourceName: sourceName, Title: title(v), Cover: cover(v), AuthorName: v.Author.Nickname, Duration: videoDuration(v), PublishTime: v.CreateTime, DiggCount: int(v.Statistics.DiggCount), CollectCount: int(v.Statistics.CollectCount), MusicURL: first(v.Music.PlayUrl.UrlList)}
	if v.PreventDownload {
		return task, errors.New("该作品不允许下载")
	}
	for _, image := range v.Images {
		if first(image.UrlList) == "" {
			return task, errors.New("图集图片地址缺失，需要重新获取详情")
		}
		if u := first(image.UrlList); u != "" {
			task.ImageURLs = append(task.ImageURLs, u)
			task.Assets = append(task.Assets, dyDownload.Asset{URL: u, Kind: "image"})
		}
		if u := first(image.Video.PlayAddr.UrlList); u != "" {
			task.Assets = append(task.Assets, dyDownload.Asset{URL: u, Kind: "video", Ext: ".mp4"})
		}
	}
	if len(task.Assets) > 0 {
		return task, nil
	}
	if track >= 0 && track < len(v.Video.BitRate) {
		task.VideoURL = first(v.Video.BitRate[track].PlayAddr.UrlList)
	}
	if track < 0 {
		task.VideoURL = first(v.Video.PlayAddrH264.UrlList)
		if task.VideoURL == "" {
			task.VideoURL = first(v.Video.PlayAddr.UrlList)
		}
	}
	if task.VideoURL == "" {
		return task, errors.New("没有可下载的视频或图集地址")
	}
	return task, nil
}
func (s *State) stage(v model.AwemeItem) {
	if v.AwemeId == "" {
		return
	}
	for _, item := range s.staged {
		if item.Data.AwemeId == v.AwemeId {
			return
		}
	}
	s.staged = append(s.staged, &queuedVideo{Data: v, Track: bestTrack(v), ResolvedAt: time.Now()})
}
func (s *State) stageSelected(navigate func(string)) {
	for _, v := range s.videos.Data.Items {
		if s.selected[v.AwemeId] {
			s.stage(v)
		}
	}
	s.selected = map[string]bool{}
	navigate("/douyin/download")
}
func (s *State) parse() {
	if s.parsing || s.api == nil || s.runtime.Run == nil {
		return
	}
	raw := strings.TrimSpace(s.link)
	if raw == "" {
		return
	}
	s.parsing = true
	s.parseSeq++
	seq, epoch := s.parseSeq, s.epoch
	api := s.api
	s.actionError = ""
	var video model.AwemeItem
	s.runtime.Run(func() error {
		s.apiMu.Lock()
		defer s.apiMu.Unlock()
		id, err := api.ParseVideo(raw)
		if err != nil {
			return err
		}
		video, err = api.VideoDetail(id)
		if err == nil && video.AwemeId == "" {
			return errors.New("视频详情为空，请检查链接或 Cookie")
		}
		return err
	}, func(err error) {
		if s.stopped || s.epoch != epoch || s.parseSeq != seq {
			return
		}
		s.parsing = false
		if strings.TrimSpace(s.link) != raw {
			return
		}
		if err != nil {
			s.actionError = err.Error()
			return
		}
		s.stage(video)
	})
}
func (s *State) request(items []*queuedVideo) {
	if s.backend == nil {
		return
	}
	pending := []*queuedVideo{}
	for _, item := range items {
		if !item.CoverBusy && item.Status != "active" && item.Status != "queued" {
			item.Status = "queued"
			item.Error = ""
			pending = append(pending, item)
		}
	}
	if len(pending) == 0 || s.backend == nil {
		return
	}
	s.nextGroup++
	s.groups = append(s.groups, &downloadGroup{ID: s.nextGroup, Order: s.nextGroup, Status: "queued", Items: pending, Seen: map[string]bool{}})
	s.pump()
}
func (s *State) requestAll(navigate func(string)) {
	if s.source.Load == nil || s.backend == nil {
		return
	}
	for _, g := range s.groups {
		if g.Source.Key == s.source.Key && (g.Status == "queued" || g.Status == "active") {
			navigate("/douyin/download")
			return
		}
	}
	s.nextGroup++
	s.groups = append(s.groups, &downloadGroup{ID: s.nextGroup, Order: s.nextGroup, Title: s.source.Title, Source: s.source, Open: true, More: true, Status: "queued", Seen: map[string]bool{}})
	navigate("/douyin/download")
	s.pump()
}
func (s *State) PendingDownloadCount() int {
	n := 0
	for _, item := range s.staged {
		if item.CoverBusy {
			n++
		}
	}
	for _, g := range s.groups {
		if g.Status == "queued" || g.Status == "active" {
			n++
		}
	}
	return n
}
func (s *State) pump() {
	if s.stopped || s.backend == nil || s.runtime.Background == nil {
		return
	}
	var next *downloadGroup
	for _, g := range s.groups {
		if g.Status == "active" {
			return
		}
		if g.Status == "queued" && (next == nil || g.Order < next.Order) {
			next = g
		}
	}
	if next != nil {
		next.Status = "active"
		s.advance(next)
	}
}
func (s *State) advance(g *downloadGroup) {
	if s.stopped || g.Status != "active" || g.Loading {
		return
	}
	if !g.Stop {
		limit := max(1, min(10, s.backend.DownloadConcurrency()))
		for _, item := range g.Items {
			if g.InFlight >= limit {
				return
			}
			if item.Status == "queued" {
				s.runItem(g, item)
			}
		}
		if g.InFlight >= limit {
			return
		}
		if g.More && g.Source.Load != nil {
			g.Loading = true
			cursor, epoch, loader := g.Cursor, s.epoch, g.Source.Load
			var page videoPage
			s.runtime.Run(func() (err error) { s.apiMu.Lock(); defer s.apiMu.Unlock(); page, err = loader(cursor); return err }, func(err error) {
				if s.stopped || s.epoch != epoch {
					return
				}
				g.Loading = false
				if err != nil {
					g.Error = err.Error()
					g.Stop = true
					s.advance(g)
					return
				}
				fresh := 0
				for _, v := range page.Items {
					if v.AwemeId != "" && !g.Seen[v.AwemeId] {
						g.Seen[v.AwemeId] = true
						fresh++
						g.Items = append(g.Items, &queuedVideo{Data: v, Track: bestTrack(v), Status: "queued", ResolvedAt: time.Now()})
					}
				}
				if page.More && (page.Next == cursor || len(page.Items) == 0 || fresh == 0) {
					g.Error = "分页没有返回新作品"
					g.Stop = true
					s.advance(g)
					return
				}
				g.Cursor, g.More = page.Next, page.More
				s.advance(g)
			})
			return
		}
	}
	if g.InFlight > 0 {
		return
	}
	if g.Stop {
		g.Status = "stopped"
	} else {
		g.Status = "done"
	}
	g.Version++
	if g.Title == "" {
		s.dismiss(g)
		s.pump()
		return
	}
	if g.Status == "done" && g.Failed == 0 && s.runtime.After != nil {
		epoch, version := s.epoch, g.Version
		s.runtime.After(3*time.Second, func() {
			if !s.stopped && s.epoch == epoch && g.Version == version && g.Status == "done" {
				s.dismiss(g)
			}
		})
	}
	s.pump()
}
func (s *State) runItem(g *downloadGroup, item *queuedVideo) {
	item.Status, item.Error, item.Progress = "active", "", dyDownload.Progress{}
	g.InFlight++
	video, track, epoch, backend, api := item.Data, item.Track, s.epoch, s.backend, s.api
	resolvedAt := item.ResolvedAt
	refreshed := false
	s.runtime.Background(func() error {
		if _, err := taskFor(video, track, g.Title); err != nil || time.Since(resolvedAt) > 10*time.Minute {
			if api == nil {
				return errors.New("视频详情服务不可用")
			}
			s.apiMu.Lock()
			fresh, err := api.VideoDetail(video.AwemeId)
			s.apiMu.Unlock()
			if err != nil {
				return err
			}
			if track >= 0 && track < len(video.Video.BitRate) {
				old := video.Video.BitRate[track]
				track = -2
				for i, v := range fresh.Video.BitRate {
					if v.GearName == old.GearName && v.QualityType == old.QualityType {
						track = i
						break
					}
				}
				if track == -2 {
					return errors.New("所选画质已不可用，请重新添加作品")
				}
			}
			video = fresh
			resolvedAt = time.Now()
			refreshed = true
		}
		task, err := taskFor(video, track, g.Title)
		if err != nil {
			return err
		}
		result, err := backend.DownloadVideos([]dyDownload.Task{task})
		if err != nil {
			return err
		}
		if result.Failed > 0 {
			for _, v := range result.Results {
				if v.Error != "" {
					return errors.New(v.Error)
				}
			}
			return errors.New("下载失败")
		}
		if result.Success == 0 {
			return errors.New("下载服务未返回成功结果")
		}
		return nil
	}, func(err error) {
		if s.stopped || s.epoch != epoch {
			return
		}
		g.InFlight--
		if refreshed {
			item.Data, item.Track, item.ResolvedAt = video, track, resolvedAt
		}
		if err != nil {
			item.Status, item.Error = "error", err.Error()
			g.Failed++
			g.Error = fmt.Sprintf("%s：%v", title(item.Data), err)
		} else {
			item.Status = "done"
			g.Done++
			s.staged = slices.DeleteFunc(s.staged, func(v *queuedVideo) bool { return v == item })
		}
		s.advance(g)
	})
}
func (s *State) UpdateDownloadProgress(p dyDownload.Progress) {
	for _, g := range s.groups {
		if g.Status != "active" {
			continue
		}
		for _, item := range g.Items {
			if item.Status == "active" && item.Data.AwemeId == p.AwemeID {
				item.Progress = p
				return
			}
		}
	}
}
func (s *State) stopGroup(g *downloadGroup) {
	g.Stop = true
	if g.Status == "queued" {
		g.Status = "stopped"
	}
	s.advance(g)
}
func (s *State) resume(g *downloadGroup) {
	if g.Status == "active" || g.Status == "queued" {
		return
	}
	for _, v := range g.Items {
		if v.Status == "error" {
			v.Status, v.Error = "queued", ""
		}
	}
	g.Stop = false
	g.Failed = 0
	g.Error = ""
	g.Version++
	s.nextGroup++
	g.Order = s.nextGroup
	g.Status = "queued"
	s.pump()
}
func (s *State) dismiss(g *downloadGroup) {
	if g.Status != "active" && g.Status != "queued" {
		s.groups = slices.DeleteFunc(s.groups, func(v *downloadGroup) bool { return v.ID == g.ID })
	}
}

func videoDuration(v model.AwemeItem) int {
	if v.Duration > 0 {
		return v.Duration / 1000
	}
	return int(v.Video.Duration) / 1000
}
