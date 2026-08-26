package api

import (
	"fmt"
	"net/url"

	"github.com/kamiertop/videodown/douyin/model"
)

// History 观看历史接口，获取用户观看过的视频列表
// maxCursor参数是分页的游标
// status: 用于过滤历史记录的状态，取值范围："-1" 表示不限制，"1" 表示已看完，"0" 表示未看完
// category: 用于过滤历史记录的分类，0：不限制，1：二次元，2：音乐，3：体育，4：电影，5：游戏
// directory: 用于过滤历史记录的时长，0：不限制，1：小于1分钟，2：1-3分钟，3：3-10分钟，4：10分钟以上
func (d *Douyin) history(maxCursor int, status, category, directory string) (model.HistoryResponse, error) {
	var resp model.HistoryResponse

	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, err
	}
	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, err
	}
	publicHeaders["Uifid"] = queryParams["uifid"].(string)
	queryParams["max_cursor"] = maxCursor
	queryParams["count"] = 20
	queryParams["status"] = status
	queryParams["category"] = category
	queryParams["directory"] = directory
	params := url.Values{}
	for key, value := range queryParams {
		params.Set(key, fmt.Sprint(value))
	}
	params.Set("a_bogus", GenerateABogus(params.Encode()))

	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/history/read/?" + params.Encode()).
		SetHeaders(publicHeaders).
		Do().
		Into(&resp)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request history api error: %v", resp.StatusCode)
		return resp, err
	}

	return resp, nil
}
