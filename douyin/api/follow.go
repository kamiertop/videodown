package api

import (
	"errors"
	"fmt"

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
