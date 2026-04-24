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
	var resp struct {
		model.ApiResponse
		Data model.SeasonsSeriesData `json:"data"`
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}
	if err := b.client.
		Get("https://api.bilibili.com/x/polymer/web-space/seasons_series_list").
		SetQueryParamsAnyType(map[string]any{
			"mid":       upMid,
			"page_size": pageSize,
			"page_num":  pageNum,
			webLocation: "333.1387",
		}).
		SetHeaders(publicHeaders()).
		SetHeader(Origin, spaceOrigin).
		SetHeader(Cookie, cookies).
		SetHeader(Referer, fmt.Sprintf("https://space.bilibili.com/%s/lists", upMid)).
		Do().
		Into(&resp); err != nil {
		b.logger.Errorf("request seasons series list error: %v", err)
		return resp.Data, errors.New("获取合集|系列列表失败")
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("get seasons series list error, code: %d, message: %s", resp.Code, resp.Message)
		return resp.Data, errors.New("获取合集|系列列表失败")
	}

	return resp.Data, nil
}

// SeasonsArchivesList 获取一个合集的视频列表
func (b *BiliBili) SeasonsArchivesList(upMid string, pageSize, pageNum, seasonId int) (model.SeasonsArchivesData, error) {
	var resp struct {
		model.ApiResponse
		Data model.SeasonsArchivesData `json:"data"`
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}

	err = b.client.Get("https://api.bilibili.com/x/polymer/web-space/seasons_archives_list").
		SetQueryParamsAnyType(map[string]any{
			"mid":           upMid,
			"season_id":     seasonId,
			"page_size":     pageSize,
			"page_num":      pageNum,
			"sort_reversed": false,
			webLocation:     "333.1387",
		}).
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookies).
		SetHeader(Referer, fmt.Sprintf("https://space.bilibili.com/%s/lists/%d?type=season", upMid, seasonId)).
		Do().
		Into(&resp)
	if err != nil {
		b.logger.Errorf("request seasons archives list api error: %v", err)
		return resp.Data, errors.New("获取合集视频列表失败")
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("get seasons archives list error, code: %d, message: %s", resp.Code, resp.Message)
		return resp.Data, errors.New("获取合集视频列表失败")
	}

	return resp.Data, nil
}

// SeriesList 获取一个系列的视频列表, 这个接口和合集视频列表的接口不一样
func (b *BiliBili) SeriesList(upMid string, ps, pn, seriesId int) (model.SeriesArchivesData, error) {
	var resp struct {
		model.ApiResponse
		Data model.SeriesArchivesData `json:"data"`
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}
	myMid, err := b.getMid()
	if err != nil {
		return resp.Data, err
	}

	err = b.client.
		Get("https://api.bilibili.com/x/series/archives").
		SetQueryParamsAnyType(map[string]any{
			"mid":         upMid,
			"current_mid": myMid,
			"series_id":   seriesId,
			"ps":          ps,
			"pn":          pn,
			"sort":        "desc",
			webLocation:   "333.1387",
			"only_normal": true,
		}).SetHeader(Cookie, cookies).
		SetHeaders(publicHeaders()).
		SetHeader(Referer, fmt.Sprintf("https://space.bilibili.com/%s/lists/%d?type=series", upMid, seriesId)).
		Do().
		Into(&resp)
	if err != nil {
		b.logger.Errorf("request series list api error: %v", err)
		return resp.Data, errors.New("获取系列视频列表失败")
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("get series list error, code: %d, message: %s", resp.Code, resp.Message)
		return resp.Data, errors.New("获取系列视频列表失败")
	}

	return resp.Data, nil
}
