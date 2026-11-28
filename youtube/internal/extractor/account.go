// Package extractor Account cookie verification. The account/account_menu endpoint powers
// YouTube's web account switcher; yt-dlp reads the same menu for account
// selection (see THIRD_PARTY_NOTICES.md).
package extractor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Account 汇总当前 Cookie 会话对应的账号资料，用于界面确认 Cookie 生效。
type Account struct {
	LoggedIn bool
	Name     string
	Email    string
	Avatar   string
}

// ErrCookieNotLoggedIn 表示 Cookie 已送达 YouTube，但会话处于未登录状态。
var ErrCookieNotLoggedIn = errors.New("cookie 未生效：当前未登录，请粘贴已登录浏览器的 Cookie")

// ErrAccountNetwork 表示验证请求未能到达 YouTube 或响应被中断（超时、
// 连接重置等），与 Cookie 本身是否有效无关。
var ErrAccountNetwork = errors.New("无法连接 YouTube")

// networkErrorText 将超时与中断类错误翻译成可读提示，其余保留原文。
// GFW 环境下直连 YouTube 常见“建连后读体中途被重置”，此时 net/http
// 只报裸的 net/http: request canceled，对用户毫无提示性。
func networkErrorText(err error) string {
	text := err.Error()
	switch {
	case strings.Contains(text, "request canceled"),
		strings.Contains(text, "context deadline exceeded"),
		strings.Contains(text, "Timeout exceeded"),
		strings.Contains(text, "i/o timeout"),
		strings.Contains(text, "timeout"):
		return "连接超时或被中断"
	case strings.Contains(text, "connection reset"),
		strings.Contains(text, "unexpected EOF"),
		strings.Contains(text, "broken pipe"),
		strings.Contains(text, "connection refused"),
		strings.Contains(text, "network is unreachable"),
		strings.Contains(text, "no route to host"):
		return "连接被重置或不可达"
	default:
		return text
	}
}

// GetAccountContext 验证 Cookie 并返回当前登录账号的昵称、邮箱与头像。
// 与播放列表流程一致：先请求首页解析会话（LOGGED_IN、visitor、真实 WEB
// 客户端版本），再以网页客户端身份调用账号菜单接口；菜单拿不到账号条目
// 时退回首页的 LOGGED_IN 状态，可覆盖缺少 SAPISID 但仍在登录的 Cookie。
func (c *Client) GetAccountContext(ctx context.Context) (*Account, error) {
	if strings.TrimSpace(c.Cookie) == "" {
		return nil, errors.New("未设置 Cookie")
	}
	web := &Client{Cookie: c.Cookie, HTTPClient: c.HTTPClient, client: &ClientInfo{Name: "WEB", Version: "2.20261029.00.00", UserAgent: VisionOSClient.UserAgent}}
	page, err := web.httpGetBodyBytes(ctx, "https://www.youtube.com/")
	if err != nil {
		return nil, fmt.Errorf("%w：%s", ErrAccountNetwork, networkErrorText(err))
	}
	web.parseSessionConfig(string(page))
	web.visitorId.updated = time.Now()
	if match := innertubeVersionPattern.FindSubmatch(page); len(match) > 1 {
		web.client.Version = string(match[1])
	}
	data := innertubeRequest{Context: prepareInnertubeContext(*web.client)}
	data.Context.Client.VisitorData = web.visitorId.value
	body, err := web.httpPostBodyBytes(ctx, "https://www.youtube.com/youtubei/v1/account/account_menu", data)
	if err != nil {
		return nil, fmt.Errorf("%w：%s", ErrAccountNetwork, networkErrorText(err))
	}
	if account, ok := parseAccountMenu(body); ok {
		return account, nil
	}
	if loggedIn, known := web.AccountState(); known && loggedIn {
		return &Account{LoggedIn: true}, nil
	}
	return nil, ErrCookieNotLoggedIn
}

// parseAccountMenu 解析账号切换菜单。旧版菜单在 accountItemRenderer 列表里
// 给出各账号，当前账号带 isSelected 标记；新版菜单不再返回账号列表，
// 改用 activeAccountHeaderRenderer 承载当前账号（字段同名：accountName、
// email、accountPhoto），按渲染器名扫描两种包装都能命中。
func parseAccountMenu(body []byte) (*Account, bool) {
	var root map[string]any
	if json.Unmarshal(body, &root) != nil {
		return nil, false
	}
	var current, first, active *Account
	visitRenderer(root, "accountItemRenderer", func(m map[string]any) {
		item := parseAccountItem(m)
		if first == nil {
			first = item
		}
		if selected, _ := m["isSelected"].(bool); selected {
			current = item
		}
	})
	visitRenderer(root, "activeAccountHeaderRenderer", func(m map[string]any) {
		if item := parseAccountItem(m); item != nil && active == nil {
			active = item
		}
	})
	switch {
	case current != nil:
		return current, true
	case active != nil:
		return active, true
	default:
		return first, first != nil
	}
}

func parseAccountItem(m map[string]any) *Account {
	account := &Account{LoggedIn: true}
	account.Name = rendererText(m["accountName"])
	account.Email = rendererText(m["email"])
	if photo, ok := m["accountPhoto"].(map[string]any); ok {
		if thumbnails, ok := photo["thumbnails"].([]any); ok {
			for _, entry := range thumbnails {
				thumbnail, _ := entry.(map[string]any)
				// 缩略图列表末尾通常是最大尺寸。
				if value, _ := thumbnail["url"].(string); value != "" {
					account.Avatar = value
				}
			}
		}
	}
	return account
}
