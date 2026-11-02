// Playlist renderer paths and continuation protocol adapted from kkdai/youtube,
// commit 3876742b95fd1cd96f836f396c2d54be8912a297 (MIT; see LICENSE).
// This implementation uses encoding/json, bounded pages and explicit pagination.
package extractor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PlaylistPage struct {
	ID, Title, Continuation string
	Videos                  []PlaylistEntry
}

type PlaylistEntry struct {
	ID, Title, Author string
	Duration          time.Duration
}

var playlistIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{2,80}$`)

func PlaylistID(input string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(input))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return "", errors.New("请输入 YouTube 播放列表链接")
	}
	switch strings.ToLower(u.Hostname()) {
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com", "youtu.be":
	default:
		return "", errors.New("请输入 YouTube 播放列表链接")
	}
	id := u.Query().Get("list")
	if !playlistIDPattern.MatchString(id) {
		return "", errors.New("链接中没有有效的播放列表 ID")
	}
	return id, nil
}

// GetPlaylistPage fetches metadata only; stream resolution happens when queued.
// A fresh WEB profile keeps browse requests separate from the playback profile.
func (c *Client) GetPlaylistPage(ctx context.Context, input, continuation string) (PlaylistPage, error) {
	id, err := PlaylistID(input)
	if err != nil {
		return PlaylistPage{}, err
	}
	web := &Client{Cookie: c.Cookie, HTTPClient: c.HTTPClient, client: &ClientInfo{Name: "WEB", Version: "2.20261029.00.00", UserAgent: VisionOSClient.UserAgent}}
	page, err := web.httpGetBodyBytes(ctx, "https://www.youtube.com/playlist?list="+url.QueryEscape(id)+"&hl=en")
	if err != nil {
		return PlaylistPage{}, err
	}
	web.parseSessionConfig(string(page))
	web.visitorId.updated = time.Now()
	// Use the version supplied by the page rather than depending on a pinned version.
	if match := regexp.MustCompile(`"INNERTUBE_CLIENT_VERSION"\s*:\s*"([^"]+)"`).FindSubmatch(page); len(match) > 1 {
		web.client.Version = string(match[1])
	}
	var root map[string]any
	found := false
	for _, prefix := range []string{"var ytInitialData =", "window[\"ytInitialData\"] =", "ytInitialData =", "ytInitialData="} {
		if _, tail, ok := strings.Cut(string(page), prefix); ok {
			if json.NewDecoder(strings.NewReader(tail)).Decode(&root) == nil {
				found = true
				break
			}
		}
	}
	if !found {
		return PlaylistPage{}, errors.New("无法读取播放列表，可能需要登录或验证网络")
	}
	result, err := parsePlaylistPage(root)
	if err != nil {
		return PlaylistPage{}, err
	}
	result.ID = id
	if result.Title == "" {
		result.Title = "YouTube 播放列表"
	}
	if continuation == "" {
		return result, nil
	}
	data := innertubeRequest{Continuation: continuation, Context: prepareInnertubeContext(*web.client)}
	data.Context.Client.VisitorData = web.visitorId.value
	body, err := web.httpPostBodyBytes(ctx, "https://www.youtube.com/youtubei/v1/browse", data)
	if err != nil {
		return PlaylistPage{}, err
	}
	root = nil
	if err := json.Unmarshal(body, &root); err != nil {
		return PlaylistPage{}, err
	}
	// WEB may advertise one final continuation and answer with session metadata
	// only. This is the end of the list, not an inaccessible initial playlist.
	if _, ok := root["responseContext"]; ok && (len(root) == 1 || (len(root) == 2 && root["trackingParams"] != nil)) {
		return PlaylistPage{ID: id, Title: result.Title}, nil
	}
	next, err := parsePlaylistPage(root)
	if err != nil {
		return PlaylistPage{}, err
	}
	next.ID, next.Title = id, result.Title
	if next.Continuation != "" && next.Continuation == continuation {
		return PlaylistPage{}, errors.New("播放列表分页未前进，请重试")
	}
	return next, nil
}

func playlistText(v any) string {
	m, _ := v.(map[string]any)
	if text, _ := m["simpleText"].(string); text != "" {
		return text
	}
	var text strings.Builder
	runs, _ := m["runs"].([]any)
	for _, run := range runs {
		r, _ := run.(map[string]any)
		t, _ := r["text"].(string)
		text.WriteString(t)
	}
	return text.String()
}

// visitPlaylistKey follows renderer names while retaining array order.
func visitPlaylistKey(node any, key string, fn func(map[string]any)) {
	switch n := node.(type) {
	case map[string]any:
		if match, ok := n[key].(map[string]any); ok {
			fn(match)
			return
		}
		for _, child := range n {
			visitPlaylistKey(child, key, fn)
		}
	case []any:
		for _, child := range n {
			visitPlaylistKey(child, key, fn)
		}
	}
}

func parsePlaylistPage(root map[string]any) (PlaylistPage, error) {
	var result PlaylistPage
	alerts, _ := root["alerts"].([]any)
	for _, alert := range alerts {
		var message string
		visitPlaylistKey(alert, "alertRenderer", func(m map[string]any) {
			if m["type"] == "ERROR" {
				message = playlistText(m["text"])
			}
		})
		if message != "" {
			return result, fmt.Errorf("播放列表不可用：%s", message)
		}
	}
	visitPlaylistKey(root["metadata"], "playlistMetadataRenderer", func(m map[string]any) { result.Title, _ = m["title"].(string) })
	if result.Title == "" {
		visitPlaylistKey(root["header"], "playlistHeaderRenderer", func(m map[string]any) { result.Title = playlistText(m["title"]) })
	}
	var contents []any
	foundContents := false
	collect := func(m map[string]any) {
		if items, ok := m["contents"].([]any); ok {
			foundContents = true
			contents = append(contents, items...)
		}
	}
	visitPlaylistKey(root["contents"], "playlistVideoListRenderer", collect)
	// Current WEB pages place lockupViewModel entries directly in item sections.
	if !foundContents {
		visitPlaylistKey(root["contents"], "itemSectionRenderer", collect)
	}
	visitPlaylistKey(root["continuationContents"], "playlistVideoListContinuation", collect)
	for _, key := range []string{"onResponseReceivedActions", "onResponseReceivedEndpoints"} {
		visitPlaylistKey(root[key], "appendContinuationItemsAction", func(m map[string]any) {
			if items, ok := m["continuationItems"].([]any); ok {
				foundContents = true
				contents = append(contents, items...)
			}
		})
	}
	if !foundContents {
		return result, errors.New("未找到播放列表内容，列表可能为私密或需要登录")
	}
	for _, value := range contents {
		m, _ := value.(map[string]any)
		if v, ok := m["lockupViewModel"].(map[string]any); ok {
			if entry, ok := playlistLockup(v); ok {
				result.Videos = append(result.Videos, entry)
			}
		}
		if v, ok := m["playlistVideoRenderer"].(map[string]any); ok {
			id, _ := v["videoId"].(string)
			if playable, exists := v["isPlayable"].(bool); exists && !playable {
				continue
			}
			if len(id) != 11 {
				continue
			}
			seconds, _ := v["lengthSeconds"].(string)
			n, _ := strconv.Atoi(seconds)
			title := playlistText(v["title"])
			if title == "[Deleted video]" || title == "[Private video]" {
				continue
			}
			if title == "" {
				title = id
			}
			result.Videos = append(result.Videos, PlaylistEntry{ID: id, Title: title, Author: playlistText(v["shortBylineText"]), Duration: time.Duration(n) * time.Second})
		}
		if token := playlistContinuation(value); token != "" {
			result.Continuation = token
		}
	}
	if result.Continuation == "" {
		result.Continuation = playlistContinuation(root["contents"])
	}
	return result, nil
}

func playlistContinuation(node any) string {
	var token string
	visitPlaylistKey(node, "continuationCommand", func(m map[string]any) {
		if token != "" {
			return
		}
		token, _ = m["token"].(string)
		if token == "" {
			token = playlistContinuation(m["innertubeCommand"])
		}
	})
	return token
}

func playlistPath(node any, keys ...string) any {
	for _, key := range keys {
		m, _ := node.(map[string]any)
		node = m[key]
	}
	return node
}

func playlistLockup(v map[string]any) (PlaylistEntry, bool) {
	id, _ := v["contentId"].(string)
	if v["contentType"] != "LOCKUP_CONTENT_TYPE_VIDEO" || len(id) != 11 {
		return PlaylistEntry{}, false
	}
	metadata := playlistPath(v, "metadata", "lockupMetadataViewModel")
	title, _ := playlistPath(metadata, "title", "content").(string)
	if title == "" {
		title = id
	}
	entry := PlaylistEntry{ID: id, Title: title}
	rows, _ := playlistPath(metadata, "metadata", "contentMetadataViewModel", "metadataRows").([]any)
	if len(rows) > 0 {
		parts, _ := playlistPath(rows[0], "metadataParts").([]any)
		if len(parts) > 0 {
			entry.Author, _ = playlistPath(parts[0], "text", "content").(string)
		}
	}
	visitPlaylistKey(v["contentImage"], "thumbnailBadgeViewModel", func(badge map[string]any) {
		text, _ := badge["text"].(string)
		if text == "" {
			text, _ = playlistPath(badge, "text", "content").(string)
		}
		var seconds int
		for _, part := range strings.Split(text, ":") {
			n, err := strconv.Atoi(part)
			if err != nil {
				return
			}
			seconds = seconds*60 + n
		}
		entry.Duration = time.Duration(seconds) * time.Second
	})
	return entry, true
}
