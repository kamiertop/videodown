package bilibili

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/kamiertop/videodown/bilibili/model"
)

var (
	videoBVPattern = regexp.MustCompile(`(?i)BV[0-9A-Za-z]{10}`)
	videoAVPattern = regexp.MustCompile(`(?i)av([0-9]+)`)
)

type manualVideo struct {
	Key, Bvid, Title, Cover, Author, SourceName string
	Cid                                         int64
	Duration, Play, Danmaku, Pubtime            int
}

type parsedVideoGroup struct {
	Kind, Title  string
	Current      manualVideo
	Items        []manualVideo
	Selected     map[string]bool
	OpenSections []string
}

func parseVideoReference(raw string) (bvid string, avid, part int, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, 0, errors.New("请输入视频链接或 BV、AV 号")
	}
	part = 1
	if parsed, parseErr := url.Parse(raw); parseErr == nil {
		if v := parsed.Query().Get("p"); v != "" {
			if n, convErr := strconv.Atoi(v); convErr == nil && n > 0 {
				part = n
			}
		} else if v := parsed.Query().Get("page"); v != "" {
			if n, convErr := strconv.Atoi(v); convErr == nil && n > 0 {
				part = n
			}
		}
	}
	if match := videoBVPattern.FindString(raw); match != "" {
		return "BV" + match[2:], 0, part, nil
	}
	if match := videoAVPattern.FindStringSubmatch(raw); len(match) == 2 {
		n, convErr := strconv.Atoi(match[1])
		if convErr == nil && n > 0 {
			return "", n, part, nil
		}
	}
	return "", 0, part, errors.New("请输入有效的 B 站视频链接或 BV、AV 号")
}

func manualCover(raw string) string {
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	return raw
}

func manualFromView(view model.VideoDetailView) manualVideo {
	return manualVideo{Key: view.Bvid, Bvid: view.Bvid, Title: view.Title, Cover: manualCover(view.Pic), Author: view.Owner.Name,
		Cid: view.Cid, Duration: view.Duration, Play: view.Stat.View, Danmaku: view.Stat.Danmaku, Pubtime: view.Pubdate}
}

func groupFromView(view model.VideoDetailView, part int) *parsedVideoGroup {
	current := manualFromView(view)
	if len(view.UgcSeason.Sections) > 0 {
		items := []manualVideo{}
		seen := make(map[string]bool)
		for _, section := range view.UgcSeason.Sections {
			for _, episode := range section.Episodes {
				bvid := strings.TrimSpace(episode.Bvid)
				if bvid == "" || seen[bvid] {
					continue
				}
				seen[bvid] = true
				title := episode.Arc.Title
				if title == "" {
					title = episode.Title
				}
				cover := episode.Arc.Pic
				if cover == "" {
					cover = view.UgcSeason.Cover
				}
				if cover == "" {
					cover = view.Pic
				}
				author := episode.Arc.Author.Name
				if author == "" {
					author = view.Owner.Name
				}
				duration := episode.Arc.Duration
				if duration == 0 {
					duration = episode.Page.Duration
				}
				items = append(items, manualVideo{Key: bvid, Bvid: bvid, Title: title, Cover: manualCover(cover), Author: author,
					SourceName: view.UgcSeason.Title, Duration: duration, Play: episode.Arc.Stat.View,
					Danmaku: episode.Arc.Stat.Danmaku, Pubtime: episode.Arc.Pubdate})
				if bvid == view.Bvid {
					current = items[len(items)-1]
				}
			}
		}
		if len(items) > 1 {
			return newParsedGroup("合集", view.UgcSeason.Title, current, items)
		}
	}
	if len(view.Pages) > 1 {
		items := make([]manualVideo, 0, len(view.Pages))
		for _, page := range view.Pages {
			if page.Cid <= 0 {
				continue
			}
			item := manualFromView(view)
			item.Key = fmt.Sprintf("%s:%d", view.Bvid, page.Cid)
			item.Cid = page.Cid
			item.Title = fmt.Sprintf("%s · P%d", view.Title, page.Page)
			if page.Part != "" {
				item.Title += " " + page.Part
			}
			item.SourceName = view.Title
			item.Duration = page.Duration
			if page.FirstFrame != "" {
				item.Cover = manualCover(page.FirstFrame)
			}
			items = append(items, item)
			if page.Page == part {
				current = item
			}
		}
		if len(items) > 1 {
			if current.Key == view.Bvid {
				current = items[0]
			}
			return newParsedGroup("分P", view.Title, current, items)
		}
	}
	return nil
}

func newParsedGroup(kind, title string, current manualVideo, items []manualVideo) *parsedVideoGroup {
	selected := make(map[string]bool, len(items))
	for _, item := range items {
		selected[item.Key] = true
	}
	return &parsedVideoGroup{Kind: kind, Title: title, Current: current, Items: items, Selected: selected, OpenSections: []string{"videos"}}
}

func (s *State) parseDownloadInput() {
	if s.manualParse.Loading || s.api == nil || s.runtime.Run == nil {
		return
	}
	raw := strings.TrimSpace(s.link)
	bvid, avid, part, err := parseVideoReference(raw)
	if err != nil {
		s.manualParse.Error = err.Error()
		return
	}
	if !s.loggedIn {
		s.manualParse.Error = "请先登录 B 站账号"
		return
	}
	s.manualGroup = nil
	s.manualParseInput = raw
	api := s.api
	load(s, &s.manualParse, func() (model.VideoDetailConciseData, error) {
		var detail model.VideoDetailConciseData
		var err error
		if bvid != "" {
			detail, err = api.VideoDetailConciseBvid(bvid)
		} else {
			detail, err = api.VideoDetailConcise(avid)
		}
		if err != nil {
			return detail, err
		}
		if detail.View.Bvid == "" {
			return detail, errors.New("解析成功，但详情中没有 BV 号")
		}
		return detail, nil
	}, func(detail model.VideoDetailConciseData) {
		if s.manualParseInput != raw || strings.TrimSpace(s.link) != raw {
			return
		}
		if group := groupFromView(detail.View, part); group != nil {
			s.manualGroup = group
			s.notice = ""
			return
		}
		item := manualFromView(detail.View)
		s.stageManual(item)
	})
}

func (s *State) stageManual(item manualVideo) bool {
	if !s.stageKey(item.Key) {
		return false
	}
	s.stagedManual = append(s.stagedManual, item)
	s.stagedSourceNames[item.Key] = item.SourceName
	return true
}

func (s *State) manualItem(key string) (manualVideo, bool) {
	for _, item := range s.stagedManual {
		if item.Key == key {
			return item, true
		}
	}
	return manualVideo{}, false
}
