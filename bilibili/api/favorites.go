package api

import (
	"errors"

	"github.com/kamiertop/videodown/bilibili/model"
)

// FavoritesList 获取自己的收藏列表，这里不处理他人的收藏列表了
func (b *BiliBili) FavoritesList() (model.FavoritesData, error) {
	var resp struct {
		model.ApiResponse
		Data model.FavoritesData `json:"data"`
	}
	cookie, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}

	mid, err := b.getMid()
	if err != nil {
		b.logger.Errorf("get mid error: %v", err)
		return resp.Data, err
	}
	if err = b.client.
		Get("https://api.bilibili.com/x/v3/fav/folder/created/list-all").
		SetQueryParam("up_mid", mid).SetQueryParam(webLocation, "333.1387").
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookie).
		Do().
		Into(&resp); err != nil {
		b.logger.Errorf("get favorites list error: %v", err)
		return resp.Data, errors.New("获取收藏列表失败")
	}
	switch resp.Code {
	case 11010:
		return resp.Data, errors.New("内容不存在")
	case -403:
		return resp.Data, errors.New("访问权限不足")
	case -400:
		return resp.Data, errors.New("请求错误")
	case model.SuccessCode:
		return resp.Data, nil
	default:
		b.logger.Errorf("get favorites list failed: code=%d message=%s", resp.Code, resp.Message)
		return resp.Data, errors.New("获取收藏列表失败")
	}
}

// Favorites 收藏夹详情，包含元数据和视频列表
// mediaId: model.FavoriteItem 中的id字段; pn: 页码, 默认是1;  ps: 每页数量, 默认40
func (b *BiliBili) Favorites(mediaId, pn, ps int) (model.FavoriteData, error) {
	var resp struct {
		model.ApiResponse
		Data model.FavoriteData `json:"data"`
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}
	if err = b.client.
		Get("https://api.bilibili.com/x/v3/fav/resource/list").
		SetQueryParamsAnyType(map[string]any{
			"media_id":  mediaId,
			webLocation: "333.1387",
			"platform":  "web",
			"tid":       0,
			"keyword":   "",
			"order":     "mtime",
			"type":      0,
			"pn":        pn,
			"ps":        ps,
		}).
		SetHeader("Cookie", cookies).
		SetHeaders(publicHeaders()).
		Do().
		Into(&resp); err != nil {
		b.logger.Errorf("request favorites error: %v", err)
		return resp.Data, errors.New("获取收藏夹详情失败")
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("get favorites failed: code=%d message=%s", resp.Code, resp.Message)
		return resp.Data, errors.New("获取收藏夹详情失败")
	}

	return resp.Data, nil
}
