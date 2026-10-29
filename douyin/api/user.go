package api

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/douyin/websign"
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
	if err = json.Unmarshal(body, &resp); err != nil {
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
	queryParams["sec_user_id"] = secUserId
	queryParams["max_cursor"] = maxCursor
	queryParams["count"] = count
	queryParams["from_user_page"] = 1
	queryParams["cut_version"] = 1
	queryParams["whale_cut_token"] = ""
	queryParams["need_time_list"] = 1
	queryParams["time_list_query"] = 0
	queryParams["locate_query"] = false
	queryParams["show_live_replay_strategy"] = 1
	queryParams["publish_video_strategy_type"] = 2
	params := url.Values{}
	for key, value := range queryParams {
		params.Set(key, fmt.Sprint(value))
	}
	aBogus := GenerateABogus(params.Encode())
	params.Set("a_bogus", aBogus)
	signature, err := websign.XSecSdkWebSignature(
		"https://www-hj.douyin.com/aweme/v1/web/aweme/post/?"+params.Encode(),
		queryParams["uifid"].(string))
	if err != nil {
		return resp, fmt.Errorf("生成 SecSDK 签名失败: %w", err)
	}
	err = d.client.
		Get(signature).
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

// ParseSecUserID accepts a sec_uid, a profile URL or shared profile text.
func (d *Douyin) ParseSecUserID(text string) (string, error) {
	text = strings.TrimSpace(text)
	valid := func(s string) bool {
		if !strings.HasPrefix(s, "MS4wLjAB") {
			return false
		}
		for _, r := range s {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
				return false
			}
		}
		return true
	}
	if valid(text) {
		return text, nil
	}
	return d.resolveShare(text, func(u *url.URL) string {
		if id := u.Query().Get("sec_uid"); valid(id) {
			return id
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		for i := 0; i+1 < len(parts); i++ {
			if parts[i] == "user" && valid(parts[i+1]) {
				return parts[i+1]
			}
		}
		return ""
	})
}
