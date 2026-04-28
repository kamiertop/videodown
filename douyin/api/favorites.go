package api

import (
	"errors"
	"fmt"

	"github.com/kamiertop/videodown/douyin/model"
)

// CollectList 收藏夹列表接口，需要分页获取，前端默认10条。cursor参数是偏移量，默认0，即从第0条数据开始返回，每次请求返回10条数据。
func (d *Douyin) CollectList(cursor int) (model.CollectListResponse, error) {
	var resp model.CollectListResponse

	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}

	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	publicHeaders["Uifid"] = queryParams["uifid"].(string)
	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/collects/list/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParamsAnyType(map[string]any{
			"cursor": cursor, // 偏移量，默认是0，即从第0条数据开始返回
			"count":  10,     // 每次请求返回的数据条数，默认是10
		}).
		SetHeaders(publicHeaders).
		Do().
		Into(&resp)
	if err != nil || resp.StatusCode != 0 {
		d.logger.Errorf("request collect list failed: %v", err)
		return resp, errors.New("请求收藏列表失败")
	}

	return resp, nil
}

// FavoriteVideo 收藏的视频列表
func (d *Douyin) FavoriteVideo(count, cursor uint) (model.FavoriteVideoResponse, error) {
	var resp model.FavoriteVideoResponse

	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}

	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}

	publicHeaders["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
	publicHeaders["Uifid"] = queryParams["uifid"].(string)

	err = d.client.
		Post("https://www-hj.douyin.com/aweme/v1/web/aweme/listcollection/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParam("publish_video_strategy_type", "2").
		SetFormDataAnyType(map[string]any{
			"count":  count,
			"cursor": cursor,
		}).
		SetHeaders(publicHeaders).
		Do().
		Into(&resp)
	if err != nil || resp.StatusCode != 0 {
		d.logger.Errorf("request favorite video failed: %v", err)
		return resp, errors.New("请求收藏视频列表失败")
	}

	return resp, nil
}
