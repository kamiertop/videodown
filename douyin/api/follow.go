package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/kamiertop/videodown/douyin/model"
)

// FollowList 关注列表，需要分页获取，前端默认20条。这里不再传参了，就默认20条了
func (d *Douyin) FollowList(offset int) (model.FollowResponse, error) {
	var resp model.FollowResponse

	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}

	publicHeaders["Accept"] = "*/*"
	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/user/following/list/").
		SetQueryParamsAnyType(map[string]any{
			"user_id":              d.userID,
			"sec_user_id":          d.secUserID,
			"offset":               offset, // 偏移量，初始为0，后续每次请求加上返回的count
			"min_time":             0,
			"max_time":             0,
			"count":                20,
			"source_type":          4,
			"gps_access":           0,
			"is_top":               1,
			"address_book_access":  0,
			"webcast_sdk_version":  "170400",
			"webcast_version_code": "170400",
		}).
		SetQueryParamsAnyType(queryParams).
		SetHeaders(publicHeaders).
		Do().
		Into(&resp)
	if err != nil {
		return resp, fmt.Errorf("请求关注列表失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request follow list failed, status_code=%d, offset=%d", resp.StatusCode, offset)
		return resp, errors.New("请求关注列表失败")
	}

	return resp, nil
}

// SearchFollow 通过昵称或抖音号搜索关注列表，返回的结果不一定是关注列表中的用户，需要二次过滤
func (d *Douyin) SearchFollow(keyword string) ([]model.SearchFollowResponse, error) {
	var resp struct {
		Msg  string `json:"msg"` // 请求成功时返回："success"
		Data []struct {
			Words []struct {
				ID     string `json:"id"`
				Word   string `json:"word"` // 通过搜索返回的昵称或抖音号
				Params struct {
					Info string `json:"info"` // 需要UnMarshal
				} `json:"params"`
			} `json:"words"`
		} `json:"data"`
	}
	followList := make([]model.SearchFollowResponse, 0)
	queryParams, err := d.publicQueryParams()
	if err != nil {
		return followList, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return followList, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	publicHeaders["Referer"] = "https://www.douyin.com/user/self?from_tab_name=main&showTab=like"
	publicHeaders["Sec-Fetch-Site"] = "same-origin"
	publicHeaders["Uifid"] = queryParams["uifid"].(string)
	values := make(url.Values)
	for key, value := range queryParams {
		values.Set(key, fmt.Sprint(value))
	}
	values.Set("count", "100")
	values.Set("query", keyword) // 搜索关键词
	values.Set("business_id", "90062")
	values.Set("pd", "aweme_at_user")
	values.Set("words_source", "aweme_at_user")
	values.Set("category_name", "aweme_at_user")
	params := values.Encode()
	aBogus := GenerateABogus(params)
	err = d.client.
		Get(fmt.Sprintf("https://www.douyin.com/aweme/v1/web/api/suggest_words/?%s&a_bogus=%s", params, url.QueryEscape(aBogus))).
		SetHeaders(publicHeaders).
		Do().
		Into(&resp)
	if err != nil {
		return followList, fmt.Errorf("请求搜索接口失败: %w", err)
	}
	if resp.Msg != "success" {
		d.logger.Errorf("request search follow failed, msg=%s", resp.Msg)
		return followList, errors.New("搜索失败")
	}
	if len(resp.Data) != 1 {
		return followList, nil
	}
	// 当前只返回一条数据，words是一个数组，里面包含了搜索结果
	words := resp.Data[0]
	for _, word := range words.Words {
		var infoMap model.SearchFollowResponse
		if err = json.Unmarshal([]byte(word.Params.Info), &infoMap); err != nil {
			d.logger.Errorf("unmarshal search follow info failed: %v", err)
			continue
		}
		if infoMap.FollowStatus == "follow" || infoMap.RelationType == "关注" {
			followList = append(followList, infoMap)
		}
	}

	return followList, nil
}
