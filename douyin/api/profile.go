package api

import (
	"fmt"

	"github.com/kamiertop/videodown/douyin/model"
)

// Profile 个人信息接口
func (d *Douyin) Profile() (model.MyInfoResponse, error) {
	var resp model.MyInfoResponse
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	err = d.client.Get("https://www.douyin.com/aweme/v1/web/user/profile/self/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParamsAnyType(map[string]any{
			"publish_video_strategy_type": 2,
			"source":                      "channel_pc_web",
			"personal_center_strategy":    1,
		}).
		SetHeaders(publicHeaders).
		SetHeader(Referer, "https://www.douyin.com/follow/").
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do().
		Into(&resp)

	if err != nil {
		d.logger.Errorf("request info api error: %v", err)
		return resp, err
	}

	return resp, nil
}
