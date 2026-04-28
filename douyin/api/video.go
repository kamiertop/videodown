package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/douyin/model"
)

func (d *Douyin) ParseVideo(link string) (string, error) {
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

func (d *Douyin) VideoDetail(awemeID string) (model.VideoDetailResponse, error) {
	var resp model.VideoDetailResponse
	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}

	headers, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/aweme/detail/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParamsAnyType(map[string]any{
			"aweme_id":       awemeID,
			"origin_type":    "video_page",
			"request_source": 600,
		}).
		SetHeaders(headers).
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Errorf("request video detail failed: %v", err)
		return resp, fmt.Errorf("请求视频详情失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request video detail failed, status code: %d", resp.StatusCode)
		return resp, fmt.Errorf("请求视频详情失败: %d", resp.StatusCode)
	}

	return resp, nil

}
