package api

import (
	"fmt"

	"github.com/kamiertop/videodown/douyin/model"
)

// Collection 收藏的合集
func (d *Douyin) Collection(count, cursor int) (model.CollectionResponse, error) {
	var resp model.CollectionResponse

	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}

	publicHeaders, err := d.publicHeaders()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	err = d.client.
		Get("https://www-hj.douyin.com/aweme/v1/web/mix/listcollection/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParamsAnyType(map[string]any{
			"count":  count,  // 每次请求返回的数据条数，默认12
			"cursor": cursor, // 偏移量，默认是0，即从第0条数据开始返回
		}).
		SetHeaders(publicHeaders).
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Errorf("request collection list failed: %v", err)
		return resp, fmt.Errorf("请求合集列表失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request collection list failed, status code: %d", resp.StatusCode)
		return resp, fmt.Errorf("请求合集列表失败: %d", resp.StatusCode)
	}

	return resp, nil
}

func (d *Douyin) CollectionList(secUserID, seriesID string, cursor, count int) (model.CollectionListResponse, error) {
	var resp model.CollectionListResponse

	queryParams, err := d.publicQueryParams()
	if err != nil {
		return resp, fmt.Errorf("获取公共查询参数失败: %w", err)
	}
	cookie, err := d.GetCookie()
	if err != nil {
		return resp, fmt.Errorf("获取公共请求头失败: %w", err)
	}
	err = d.client.
		Get("https://www.douyin.com/aweme/v1/web/series/aweme/").
		SetQueryParamsAnyType(queryParams).
		SetQueryParamsAnyType(map[string]any{
			"count":     count,    // 每次请求返回的数据条数，默认12
			"cursor":    cursor,   // 偏移量，默认是0，即从第0条数据开始返回
			"series_id": seriesID, // 合集ID
			"pull_type": 2,
		}).
		SetHeaders(map[string]string{
			"Pragma":        "no-cache",
			Priority:        "u=1, i",
			Referer:         fmt.Sprintf("https://www.douyin.com/user/%s?from_tab_name=main&showSubTab=compilation", secUserID),
			"Uifid":         queryParams["uifid"].(string),
			"User-Agent":    userAgent(),
			SecChFetchSite:  "same-origin",
			SecChFetchMode:  "cors",
			SecChFetchDest:  "empty",
			"Sec-Ch-Ua":     `"Google Chrome";v="147", "Not:A-Brand";v="8", "Chromium";v="147"`,
			SecCHUAMobile:   "?0",
			SecCHUAPlatform: fmt.Sprintf(`"%s"`, osName()),
			"Cookie":        cookie,
			"Cache-Control": "no-cache",
			Accept:          "application/json, text/plain, */*",
			AcceptEncoding:  "gzip, deflate, br, zstd",
			AcceptLanguage:  "zh-CN,zh;q=0.9",
		}).
		SetHeader(Referer, "https://www.douyin.com/user/self").
		SetHeader("Uifid", queryParams["uifid"].(string)).
		Do().
		Into(&resp)
	if err != nil {
		d.logger.Errorf("request collection list failed: %v", err)
		return resp, fmt.Errorf("请求合集视频列表失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request collection list failed, status code: %d", resp.StatusCode)
		return resp, fmt.Errorf("请求合集视频列表失败: %d, msg: %s", resp.StatusCode, resp.StatusMsg)
	}

	return resp, nil
}
