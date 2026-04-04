package api

import (
	"errors"

	"github.com/kamiertop/videodown/bilibili/model"
)

// FavoritesList 获取自己的收藏列表，这里不处理他人的收藏列表了
func (b *BiliBili) FavoritesList() (model.FavoritesData, error) {
	cookie, err := b.getCookies()
	if err != nil {
		return model.FavoritesData{}, err
	}

	mid, err := b.getMid()
	if err != nil {
		b.logger.Errorf("get mid error: %v", err)
		return model.FavoritesData{}, err
	}

	resp, err := b.client.
		R().
		SetQueryParam("up_mid", mid).SetQueryParam(webLocation, "333.1387").
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookie).
		Get("https://api.bilibili.com/x/v3/fav/folder/created/list-all")
	if err != nil {
		b.logger.Errorf("get favorites list error: %v", err)
		return model.FavoritesData{}, err
	}
	var response model.FavoritesResponse
	if err := resp.Into(&response); err != nil {
		b.logger.Errorf("unmarshal favorites list response error: %v", err)
		return model.FavoritesData{}, errors.New("解析收藏列表响应失败")
	}
	switch response.Code {
	case 11010:
		return model.FavoritesData{}, errors.New("内容不存在")
	case -403:
		return model.FavoritesData{}, errors.New("访问权限不足")
	case -400:
		return model.FavoritesData{}, errors.New("请求错误")
	case model.SuccessCode:
		return response.Data, nil
	default:
		b.logger.Errorf("get favorites list failed: code=%d message=%s", response.Code, response.Message)
		return model.FavoritesData{}, errors.New("获取收藏列表失败")
	}
}

// Favorites 收藏夹详情，包含元数据和视频列表
// mediaId: model.FavoriteItem 中的id字段; pn: 页码, 默认是1;  ps: 每页数量, 默认40
func (b *BiliBili) Favorites(mediaId, pn, ps int) (model.FavoriteData, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.FavoriteData{}, err
	}
	resp, err := b.client.
		R().
		SetQueryParamsAnyType(map[string]any{
			"media_id":  mediaId, // 收藏夹的列表中的id model.FavoriteItem 中的id字段
			webLocation: "333.1387",
			"platform":  "web",
			"tid":       0, // 分区id，0表示全部分区
			"keyword":   "",
			"order":     "mtime", // 排序方式，默认按收藏时间排序,播放量： view，投稿时间：pubtime
			"type":      0,       // 查询范围.0：当前收藏夹，对应media_id，1：全部收藏夹
			"pn":        pn,      // 页码
			"ps":        ps,      // 每页数量
		}).
		SetHeader("Cookie", cookies).
		SetHeaders(publicHeaders()).
		Get("https://api.bilibili.com/x/v3/fav/resource/list")
	if err != nil {
		b.logger.Errorf("request favorites error: %v", err)
		return model.FavoriteData{}, errors.New("请求收藏夹详情接口失败")
	}
	var response model.FavoriteResponse
	if err := resp.Into(&response); err != nil {
		b.logger.Errorf("unmarshal favorites response error: %v", err)
		return model.FavoriteData{}, errors.New("解析收藏夹详情响应失败")
	}
	if response.Code != model.SuccessCode {
		b.logger.Errorf("get favorites failed: code=%d message=%s", response.Code, response.Message)
		return model.FavoriteData{}, errors.New("获取收藏夹详情失败")
	}

	return response.Data, nil
}
