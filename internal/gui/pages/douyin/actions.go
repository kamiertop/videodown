package douyin

import (
	"fmt"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/feedback"
	"github.com/egoist/mygo/ui"
	dyDownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/douyin/model"
)

func progressView(c *ui.Context, p dyDownload.Progress) {
	if p.Phase == "sleep" {
		core.Spinner(c, core.SpinnerOptions{Size: 14, Label: fmt.Sprintf("休眠中 · %d 秒", max(0, p.SleepRemaining)), ShowLabel: true})
		return
	}
	label := map[string]string{"video": "下载视频", "audio": "下载音频", "image": "保存图片", "merge": "合并文件", "done": "已完成", "error": "下载失败"}[p.Phase]
	if label == "" {
		label = "下载中"
	}
	feedback.Progress(c, max(0, min(100, p.Percent))/100, feedback.ProgressOptions{Label: label, ShowValue: true})
}
func (s *State) saveCover(item *queuedVideo) {
	if s.backend == nil || s.runtime.Background == nil || item.CoverBusy {
		return
	}
	task, err := taskFor(item.Data, item.Track, "")
	if err != nil && task.Cover == "" {
		s.actionError = err.Error()
		return
	}
	covers := []model.Cover{item.Data.Video.Cover, item.Data.Video.OriginCover}
	if task.Cover != "" {
		covers = append(covers, model.Cover{UrlList: []string{task.Cover}})
	}
	backend, epoch := s.backend, s.epoch
	s.actionError = ""
	s.notice = ""
	item.CoverBusy = true
	s.runtime.Background(func() error { _, err := backend.DownloadCover(covers, task); return err }, func(err error) {
		if !s.stopped && s.epoch == epoch {
			item.CoverBusy = false
			if err != nil {
				s.actionError = err.Error()
			} else {
				s.notice = "封面已保存"
			}
		}
	})
}
func (s *State) loadHistory(more bool) {
	if s.backend == nil || s.historyBusy {
		return
	}
	backend := s.backend
	query := strings.TrimSpace(s.historySearch)
	if more {
		if s.history.Loading || s.historyMore.Loading || !s.history.Data.HasMore {
			return
		}
		offset := len(s.history.Data.Items)
		load(s, &s.historyMore, func() (dyDownload.HistoryPage, error) { return backend.DownloadHistoryPage(offset, 50, query) }, func(page dyDownload.HistoryPage) {
			seen := map[string]bool{}
			for _, v := range s.history.Data.Items {
				seen[v.AwemeID] = true
			}
			for _, v := range page.Items {
				if !seen[v.AwemeID] {
					s.history.Data.Items = append(s.history.Data.Items, v)
					seen[v.AwemeID] = true
				}
			}
			s.history.Data.Total, s.history.Data.HasMore = page.Total, page.HasMore && len(page.Items) > 0
		})
		return
	}
	s.historyQuery = s.historySearch
	s.historyMore.seq++
	s.historyMore.Loading = false
	load(s, &s.history, func() (dyDownload.HistoryPage, error) { return backend.DownloadHistoryPage(0, 50, query) }, nil)
}
func (s *State) changeHistory(work func() error) {
	if s.runtime.Run == nil || s.historyBusy {
		return
	}
	s.historyBusy = true
	s.actionError = ""
	s.history.seq++
	s.historyMore.seq++
	s.history.Loading = false
	s.historyMore.Loading = false
	epoch := s.epoch
	s.runtime.Run(work, func(err error) {
		if s.stopped || s.epoch != epoch {
			return
		}
		s.historyBusy = false
		if err != nil {
			s.actionError = err.Error()
			return
		}
		s.clearOpen = false
		s.loadHistory(false)
	})
}
func (s *State) openFile(path string, directory bool) {
	action := s.runtime.OpenFile
	if directory {
		action = s.runtime.OpenLocation
	}
	if action == nil || s.runtime.Run == nil {
		return
	}
	if path == "" {
		s.actionError = "文件路径为空"
		return
	}
	epoch := s.epoch
	s.runtime.Run(func() error { return action(path) }, func(err error) {
		if !s.stopped && s.epoch == epoch && err != nil {
			s.actionError = err.Error()
		}
	})
}
