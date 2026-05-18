package api

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/imroc/req/v3"

	"github.com/kamiertop/videodown/douyin/model"
)

// User 获取用户信息
func (d *Douyin) User(secUserId string) (model.UserResponse, error) {
	var resp model.UserResponse
	queryParams, err := d.publicQueryParams()
	if err != nil {
		d.logger.Errorf("request user info query params error: %v", err)
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		d.logger.Errorf("request user info public headers error: %v", err)
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	publicHeaders[Referer] = fmt.Sprintf("https://www.douyin.com/user/%s", secUserId)

	values := url.Values{}
	for key, value := range queryParams {
		values.Set(key, fmt.Sprint(value))
	}
	values.Set("sec_user_id", secUserId)
	values.Set("personal_center_strategy", "1")
	values.Set("profile_other_record_enable", "1")
	values.Set("land_to", "1")
	params := values.Encode()
	aBogus := GenerateABogus(params)

	apiResp := d.client.
		Get(fmt.Sprintf("https://www-hj.douyin.com/aweme/v1/web/user/profile/other/?%s&a_bogus=%s", params, url.QueryEscape(aBogus))).
		SetHeaders(publicHeaders).
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do()
	if apiResp.Err != nil {
		d.logger.Errorf("request info api error: %v", apiResp.Err)
		return resp, apiResp.Err
	}
	if apiResp.IsErrorState() {
		d.logger.Errorf("request user info http error, status code: %d", apiResp.StatusCode)
		return resp, fmt.Errorf("请求用户信息失败, status_code=%d", apiResp.StatusCode)
	}
	body, err := apiResp.ToBytes()
	if err != nil {
		d.logger.Errorf("read user info response failed: %v", err)
		return resp, fmt.Errorf("读取用户信息响应失败: %w", err)
	}
	if strings.TrimSpace(string(body)) == "" {
		d.logger.Errorf("request user info api returned empty body, status code: %d", apiResp.StatusCode)
		return resp, errors.New("请求用户信息失败：接口返回空响应")
	}
	if err = sonic.Unmarshal(body, &resp); err != nil {
		d.logger.Errorf("unmarshal user info response failed: %v", err)
		return resp, fmt.Errorf("解析用户信息失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request user info api error, status code: %d", resp.StatusCode)
		return resp, errors.New("请求用户信息失败")
	}

	return resp, nil
}

// UserVideoList 用户空间的视频列表
func (d *Douyin) UserVideoList(secUserId string, count, maxCursor int) (model.UserVideoListResponse, error) {
	var resp model.UserVideoListResponse
	queryParams, err := d.publicQueryParams()
	if err != nil {
		d.logger.Errorf("request user video list query params error: %v", err)
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		d.logger.Errorf("request user video list public headers error: %v", err)
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/aweme/post/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParamsAnyType(map[string]any{
			"sec_user_id":                 secUserId,
			"max_cursor":                  maxCursor,
			"count":                       count,
			"from_user_page":              1,
			"cut_version":                 1,
			"whale_cut_token":             "",
			"need_time_list":              1,
			"time_list_query":             0,
			"locate_query":                false,
			"show_live_replay_strategy":   1,
			"publish_video_strategy_type": 2,
		}).
		SetHeaders(publicHeaders).
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Errorf("request user video list api error: %v", err)
		return resp, errors.New("请求用户视频列表失败")
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request user video list api error, status code: %d", resp.StatusCode)
		return resp, errors.New("请求用户视频列表失败")
	}

	return resp, nil
}

// UserSeries 用户空间的合集列表
func (d *Douyin) UserSeries(secUserID string, cursor, count int) (model.UserSeriesResponse, error) {
	var resp model.UserSeriesResponse

	params, err := d.publicQueryParams()
	if err != nil {
		d.logger.Errorf("request user series query params error: %v", err)
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	headers, err := d.publicHeaders()
	if err != nil {
		d.logger.Errorf("request user series public headers error: %v", err)
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	headers["Uifid"] = params["uifid"].(string)
	headers["Referer"] = fmt.Sprintf("https://www.douyin.com/user/%s?from_tab_name=main&showSubTab=compilation", secUserID)
	delete(headers, Origin)

	err = d.client.
		Get("https://www.douyin.com/aweme/v1/web/series/list/").
		SetQueryParamsAnyType(params).
		SetQueryParamsAnyType(map[string]any{
			"sec_user_id":  secUserID,
			"cursor":       cursor,
			"count":        count,
			"read_new_mix": true,
			"req_from":     "channel_pc_web",
		}).
		SetHeaders(headers).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Errorf("request user series api error: %v", err)
		return resp, fmt.Errorf("请求用户合集列表失败: %w", err)
	}

	return resp, nil
}

// ParseSecUserID 从用户空间URL或分享链接中提取sec_user_id
func (d *Douyin) ParseSecUserID(URLOrSecUserID string) (string, error) {
	const secUserIDPrefix = "MS4wLjAB"
	URLOrSecUserID = strings.TrimSpace(URLOrSecUserID)
	if strings.Contains(URLOrSecUserID, "v.douyin.com") {
		parts := strings.Split(URLOrSecUserID, "https://v.douyin.com/")
		if len(parts) < 2 {
			return "", errors.New("无效的抖音用户URL")
		}
		shortURL := "https://v.douyin.com/" + strings.TrimSpace(parts[1])
		// 必须clone，以免影响全局client的重定向策略
		resp, err := d.client.
			Clone().
			SetRedirectPolicy(req.NoRedirectPolicy()).
			R().
			Get(shortURL)
		if err != nil {
			return "", fmt.Errorf("请求短链接失败: %w", err)
		}
		redirectURL := resp.Header.Get("Location")
		if redirectURL == "" {
			return "", errors.New("未能获取短链接重定向地址")
		}

		parsed, err := url.Parse(redirectURL)
		if err != nil {
			return "", fmt.Errorf("解析重定向URL失败: %w", err)
		}
		// 从路径 /share/user/{sec_uid} 中提取
		if pathParts := strings.Split(parsed.Path, "/share/user/"); len(pathParts) == 2 {
			secUID := strings.Split(pathParts[1], "?")[0]
			if strings.Contains(secUID, secUserIDPrefix) {
				return secUID, nil
			}
		}
		// 从查询参数 sec_uid 中提取
		if secUID := parsed.Query().Get("sec_uid"); secUID != "" && strings.Contains(secUID, secUserIDPrefix) {
			return secUID, nil
		}

		return "", errors.New("未能从短链接重定向地址中提取用户ID")
	}

	if !strings.Contains(URLOrSecUserID, secUserIDPrefix) {
		return "", errors.New("无效的抖音用户ID或URL")
	}

	if strings.Contains(URLOrSecUserID, "www.douyin.com/user/") {
		parts := strings.Split(URLOrSecUserID, "www.douyin.com/user/")
		if len(parts) < 2 {
			return "", errors.New("无效的抖音用户URL")
		}
		return strings.Split(parts[1], "?")[0], nil
	}

	return URLOrSecUserID, nil
}
