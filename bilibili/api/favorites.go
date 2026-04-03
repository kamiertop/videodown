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
