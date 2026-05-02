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

// FavoriteVideo 收藏的视频列表，视频列表包含收藏夹中的
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

// FavoritesVideoList 获取指定收藏夹中的视频列表。
// 收藏夹 ID 可能超过 JS 安全整数范围，前端应传 collects_id_str，避免 Wails number 精度丢失。
func (d *Douyin) FavoritesVideoList(collectIDStr string, cursor, count int) (model.FavoriteVideoResponse, error) {
	var resp model.FavoriteVideoResponse
	if collectIDStr == "" {
		return resp, errors.New("收藏夹ID为空")
	}
	params, err := d.publicQueryParams()
	if err != nil {
		d.logger.Errorf("request public query params failed: %v", err)
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	headers, err := d.publicHeaders()
	if err != nil {
		d.logger.Errorf("request public header failed: %v", err)
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/collects/video/list/").
		SetQueryParamsAnyType(params).
		SetQueryParamsAnyType(map[string]any{
			"collects_id": collectIDStr, // 收藏夹ID
			"cursor":      cursor,       // 偏移量，默认是0，即从第0条数据开始返回
			"count":       count,        // 每次请求返回的数据条数，默认是10
		}).
		SetHeaders(headers).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Errorf("request favorites video list failed: %v", err)
		return resp, err
	}

	if resp.StatusCode != 0 {
		d.logger.Errorf("request favorites video list failed, status code: %d", resp.StatusCode)
		return resp, fmt.Errorf("请求收藏夹视频列表失败: %d", resp.StatusCode)
	}

	return resp, nil
}
