package api

import (
	"errors"
	"fmt"

	"github.com/kamiertop/videodown/bilibili/model"
)

// SeasonsSeriesList 获取合集|系列列表, 这个使用的接口是：https://api.bilibili.com/x/polymer/web-space/seasons_series_list
// https://api.bilibili.com/x/polymer/web-space/home/seasons_series返回的不全
// 投稿总量：视频+图文+音频
func (b *BiliBili) SeasonsSeriesList(upMid string, pageSize, pageNum int) (model.SeasonsSeriesData, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.SeasonsSeriesData{}, err
	}
	var response model.SeasonsSeriesResponse
	if err := b.client.
		Get("https://api.bilibili.com/x/polymer/web-space/seasons_series_list").
		SetQueryParamsAnyType(map[string]any{
			"mid":       upMid,
			"page_size": pageSize,
			"page_num":  pageNum,
			webLocation: "333.1387",
		}).
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookies).
		SetHeader(Referer, fmt.Sprintf("https://space.bilibili.com/%s/lists", upMid)).
		Do().
		Into(&response); err != nil {
		b.logger.Errorf("request seasons series list error: %v", err)
		return model.SeasonsSeriesData{}, errors.New("获取合集|系列列表失败")
	}
	if response.Code != model.SuccessCode {
		b.logger.Errorf("get seasons series list error, code: %d, message: %s", response.Code, response.Message)
		return model.SeasonsSeriesData{}, errors.New("获取合集|系列列表失败")
	}

	return response.Data, nil
}
