package api

import (
	"fmt"
	"net/url"

	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/douyin/websign"
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
	publicHeaders["Uifid"] = queryParams["uifid"].(string)
	queryParams["count"] = count
	queryParams["cursor"] = cursor

	params := url.Values{}
	for key, value := range queryParams {
		params.Set(key, fmt.Sprint(value))
	}
	aBogus := GenerateABogus(params.Encode())
	params.Set("a_bogus", aBogus)

	signed, err := websign.XSecSdkWebSignature(
		"https://www-hj.douyin.com/aweme/v1/web/mix/listcollection/?"+params.Encode(),
		queryParams["uifid"].(string),
	)
	if err != nil {
		return resp, fmt.Errorf("生成 SecSDK 签名失败: %w", err)
	}

	if err = d.client.Get(signed).SetHeaders(publicHeaders).Do().Into(&resp); err != nil {
		d.logger.Errorf("request collection list failed: %v", err)
		return resp, fmt.Errorf("请求合集列表失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request collection list failed, status code: %d", resp.StatusCode)
		return resp, fmt.Errorf("请求合集列表失败: %d", resp.StatusCode)
	}

	return resp, nil
}

// CollectionList 合集视频列表
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
	queryParams["count"] = count
	queryParams["cursor"] = cursor
	queryParams["series_id"] = seriesID
	queryParams["pull_type"] = 2

	params := url.Values{}
	for key, value := range queryParams {
		params.Set(key, fmt.Sprint(value))
	}
	aBogus := GenerateABogus(params.Encode())
	params.Set("a_bogus", aBogus)

	signed, err := websign.XSecSdkWebSignature(
		"https://www.douyin.com/aweme/v1/web/series/aweme/?"+params.Encode(),
		queryParams["uifid"].(string),
	)
	if err != nil {
		return resp, fmt.Errorf("生成 SecSDK 签名失败: %w", err)
	}
	headers := map[string]string{
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
	}

	if err = d.client.Get(signed).SetHeaders(headers).Do().Into(&resp); err != nil {
		d.logger.Errorf("request collection list failed: %v", err)
		return resp, fmt.Errorf("请求合集视频列表失败: %w", err)
	}
	if resp.StatusCode != 0 {
		d.logger.Errorf("request collection list failed, status code: %d", resp.StatusCode)
		return resp, fmt.Errorf("请求合集视频列表失败: %d, msg: %s", resp.StatusCode, resp.StatusMsg)
	}

	return resp, nil
}
