package api

import (
	"fmt"
	"time"

	"github.com/kamiertop/videodown/douyin/model"
)

// Profile 个人信息接口
// 没找到单独返回个人信息的接口,这个接口返回的第一个是个人信息(没想明白为啥这么设计)
func (d *Douyin) Profile() (model.MyInfo, error) {
	var resp model.MyInfo
	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	err = d.client.Get("https://www-hj.douyin.com/aweme/v1/web/im/spotlight/relation/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParamsAnyType(map[string]any{
			"count":                   10,
			"source":                  "coldup",
			"max_time":                time.Now().Unix(),
			"min_time":                0,
			"need_remove_share_panel": true,
			"need_sorted_info":        true,
			"with_fstatus":            1,
		}).
		SetHeaders(publicHeaders).
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do().
		Into(&resp)

	if err != nil {
		d.logger.Errorf("request info api error: %v", err)
		return resp, err
	}

	return resp, nil
}
