package api

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/douyin/model"
)

// ParseVideo 解析抖音视频链接，返回awemeID，为保证整体工作流，精选链接只返回awemeID，然后前端再次根据awemeID请求视频详情接口获取视频详情
func (d *Douyin) ParseVideo(link string) (string, error) {
	if strings.Contains(link, "douyin.com/jingxuan") && strings.Contains(link, "modal_id=") {
		return d.parseJingXuanAwemeID(link)
	}
	link = strings.TrimSpace(link)
	split := strings.Split(link, " ")
	for _, item := range split {
		if strings.Contains(item, "https") {
			link = item
			break
		}
	}
	if link == "" {
		return "", errors.New("链接不能为空")
	}
	resp, err := d.client.
		Clone().
		SetRedirectPolicy(req.NoRedirectPolicy()).
		R().
		Get(link)
	if err != nil {
		return "", fmt.Errorf("请求链接失败: %w", err)
	}
	redirectURL := resp.Header.Get("Location")
	if redirectURL == "" {
		return "", errors.New("未能获取重定向地址")
	}
	parts := strings.Split(redirectURL, "share/video/")
	if len(parts) < 2 {
		return "", errors.New("未能从重定向地址中提取视频ID")
	}
	videoID := strings.Split(parts[1], "?")[0]

	return strings.TrimSuffix(videoID, "/"), nil
}

// VideoDetail 获取视频详情
func (d *Douyin) VideoDetail(awemeID string) (model.AwemeItem, error) {
	var resp struct {
		StatusCode int             `json:"status_code"`
		AwemeItem  model.AwemeItem `json:"aweme_detail"`
	}
	headers, err := d.publicHeaders()
	if err != nil {
		return resp.AwemeItem, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	params := detailParams(awemeID)
	aBogus := GenerateABogus(params)
	err = d.client.
		Get(fmt.Sprintf("https://www-hj.douyin.com/aweme/v1/web/aweme/detail/?%s&a_bogus=%s", params, url.QueryEscape(aBogus))).
		SetHeaders(headers).
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
