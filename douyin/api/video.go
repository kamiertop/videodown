package api

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/douyin/websign"
)

// ParseVideo accepts IDs, normal video/note URLs and shared short links.
func (d *Douyin) ParseVideo(text string) (string, error) {
	text = strings.TrimSpace(text)
	if numericID(text) {
		return text, nil
	}
	return d.resolveShare(text, func(u *url.URL) string {
		if id := u.Query().Get("modal_id"); numericID(id) {
			return id
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		for i := 0; i+1 < len(parts); i++ {
			if (parts[i] == "video" || parts[i] == "note" || parts[i] == "slides") && numericID(parts[i+1]) {
				return parts[i+1]
			}
		}
		return ""
	})
}

func numericID(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// resolveShare follows only Douyin redirects, without changing the shared client.
func (d *Douyin) resolveShare(text string, extract func(*url.URL) string) (string, error) {
	raw := shareURL.FindString(text)
	if raw == "" {
		return "", errors.New("请输入有效的抖音链接或作品 ID")
	}
	raw = strings.TrimRight(raw, "，。！!；;、）)]}>")
	current, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("无效的抖音链接")
	}
	client := d.client.Clone().SetRedirectPolicy(req.NoRedirectPolicy())
	for range 6 {
		host := strings.ToLower(current.Hostname())
		if (current.Scheme != "https" && current.Scheme != "http") || (host != "douyin.com" && !strings.HasSuffix(host, ".douyin.com")) || current.User != nil {
			return "", errors.New("请使用抖音链接")
		}
		if id := extract(current); id != "" {
			return id, nil
		}
		resp, err := client.R().Get(current.String())
		if err != nil {
			return "", fmt.Errorf("请求分享链接失败: %w", err)
		}
		location := resp.Header.Get("Location")
		if resp.StatusCode < 300 || resp.StatusCode >= 400 || location == "" {
			return "", errors.New("分享链接没有返回有效的跳转地址")
		}
		next, err := url.Parse(location)
		if err != nil {
			return "", errors.New("分享链接跳转地址无效")
		}
		current = current.ResolveReference(next)
	}
	return "", errors.New("分享链接跳转次数过多")
}

var shareURL = regexp.MustCompile(`https?://[^\s<>"'，。！；（）]+`)

// VideoDetail 获取视频详情
func (d *Douyin) VideoDetail(awemeID string) (model.AwemeItem, error) {
	var resp struct {
		StatusCode int             `json:"status_code"`
		AwemeItem  model.AwemeItem `json:"aweme_detail"`
	}
	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp.AwemeItem, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp.AwemeItem, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	queryParams["aweme_id"] = awemeID

	params := url.Values{}
	for key, value := range queryParams {
		params.Set(key, fmt.Sprint(value))
	}
	aBogus := GenerateABogus(params.Encode())
	params.Set("a_bogus", aBogus)
	signed, err := websign.XSecSdkWebSignature(
		"https://www.douyin.com/aweme/v1/web/aweme/detail/?"+params.Encode(),
		queryParams["uifid"].(string))
	if err != nil {
		return resp.AwemeItem, fmt.Errorf("生成 SecSDK 签名失败: %w", err)
	}
	err = d.client.
		Get(signed).
		SetHeaders(publicHeaders).
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Errorf("request video detail failed: %v", err)
		return resp.AwemeItem, fmt.Errorf("请求视频详情失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request video detail failed, status code: %d", resp.StatusCode)
		return resp.AwemeItem, fmt.Errorf("请求视频详情失败: %d", resp.StatusCode)
	}

	return resp.AwemeItem, nil
}

/*

https://www.douyin.com/aweme/v1/web/aweme/post/ 也可以获取用户的视频列表，但只有前20条，后面还是会用上面的接口获取

*/
