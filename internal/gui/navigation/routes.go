// Package navigation declares the native UI routes independently of rendering.
package navigation

import (
	"net/url"
	"strconv"
	"strings"
)

// Page is one destination of the application.
type Page struct {
	ID    string
	Label string
	Hint  string
}

// Path is the page's route in the shell's router.
func (p Page) Path() string {
	if p.ID == "home" {
		return "/"
	}
	return "/" + p.ID
}

// catalog declares the application destinations.
var catalog = []Page{
	{ID: "bilibili", Label: "BiliBili", Hint: "收藏夹与视频下载"},
	{ID: "douyin", Label: "抖音", Hint: "作品解析与下载"},
	{ID: "youtube", Label: "YouTube", Hint: "视频解析与下载"},
	{ID: "settings", Label: "设置", Hint: "下载目录与偏好"},
	{ID: "home", Label: "首页", Hint: "从喜欢的平台开始，整理与下载你的视频。"},
}

// ByPath resolves only declared routes, including each page's subroutes.
func ByPath(path string) (Page, bool) {
	for _, p := range catalog {
		if p.ID == "douyin" && validDouyinDetail(path) {
			return p, true
		}
		if p.ID == "bilibili" {
			if _, ok := BilibiliAuthorID(path); ok {
				return p, true
			}
		}
		if p.Path() == path {
			return p, true
		}
		for _, section := range p.Sections() {
			if p.Path()+"/"+section.ID == path {
				return p, true
			}
		}
	}
	return Page{}, false
}

type Section struct{ ID, Label string }

func (p Page) Sections() []Section {
	switch p.ID {
	case "bilibili", "douyin":
		author := Section{"up", "UP 主"}
		if p.ID == "douyin" {
			author = Section{"user", "用户"}
		}
		sections := []Section{{"download", "视频下载"}, author, {"dynamic", "关注动态"}, {"favorite", "收藏夹"}, {"play-history", "播放历史"}, {"history", "下载历史"}, {"profile", "账号"}}
		if p.ID == "bilibili" {
			sections = append(sections, Section{"collection", "合集"})
		}
		return sections
	case "youtube":
		return []Section{{"download", "视频下载"}, {"history", "下载历史"}, {"profile", "账号"}}
	case "settings":
		return []Section{{"", "通用"}, {"bilibili", "哔哩哔哩"}, {"download", "下载"}, {"filename", "命名模板"}, {"ffmpeg", "FFmpeg"}, {"about", "关于"}}
	}
	return nil
}

func (p Page) CurrentSection(path string) string {
	if p.ID == "douyin" && validDouyinDetail(path) {
		return strings.Split(strings.Trim(path, "/"), "/")[1]
	}
	if p.ID == "bilibili" {
		if _, ok := BilibiliAuthorID(path); ok {
			return "up"
		}
	}
	id := strings.TrimPrefix(path, p.Path())
	id = strings.TrimPrefix(id, "/")
	if id == "" && (p.ID == "bilibili" || p.ID == "douyin" || p.ID == "youtube") {
		return "download"
	}
	return id
}

// All returns a copy of the declared top-level routes.
func All() []Page { return append([]Page(nil), catalog...) }

// Platforms returns the supported content providers in display order.
func Platforms() []Page {
	var result []Page
	for _, p := range catalog {
		if p.ID == "bilibili" || p.ID == "douyin" || p.ID == "youtube" {
			result = append(result, p)
		}
	}
	return result
}

// BilibiliAuthorPath gives author details a real history entry.
func BilibiliAuthorPath(mid uint64) string { return "/bilibili/up/" + strconv.FormatUint(mid, 10) }
func BilibiliAuthorID(path string) (uint64, bool) {
	const prefix = "/bilibili/up/"
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	mid, err := strconv.ParseUint(strings.TrimPrefix(path, prefix), 10, 64)
	return mid, err == nil && mid > 0
}

func validDouyinDetail(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 || parts[0] != "douyin" {
		return false
	}
	if parts[1] == "user" && len(parts) == 3 {
		uid, err := url.PathUnescape(parts[2])
		return err == nil && uid != "" && !strings.Contains(uid, "/")
	}
	if parts[1] == "favorite" && (len(parts) == 3 || len(parts) == 4) {
		if parts[2] != "folder" && parts[2] != "collection" {
			return false
		}
		return len(parts) == 3 || parts[3] != ""
	}
	return false
}
