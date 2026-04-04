package api

import (
	"errors"

	"github.com/kamiertop/videodown/bilibili/model"
)

// Collection 获取：我追的合集列表, 默认pn：1, ps: 50
func (b *BiliBili) Collection(pn, ps int) (model.CollectionData, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.CollectionData{}, err
	}
	upMid, err := b.getMid()
	if err != nil {
		b.logger.Errorf("get mid error: %v", err)
		return model.CollectionData{}, err
	}
	var collection model.CollectionResponse
	if err := b.client.
		Get("https://api.bilibili.com/x/v3/fav/folder/collected/list").
		SetQueryParamsAnyType(map[string]any{
			"pn":        pn,
			"ps":        ps,
			"up_mid":    upMid,
			webLocation: "333.1387",
			"platform":  "web",
		}).
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookies).
		Do().
		Into(&collection); err != nil {
		b.logger.Errorf("request collection list error: %v", err)
		return model.CollectionData{}, errors.New("获取合集列表失败")
	}
	if collection.Code != model.SuccessCode {
		b.logger.Errorf("collection list error, code: %d, : %v", collection.Code, collection.Message)
		return model.CollectionData{}, errors.New("获取合集列表失败")
	}

	return collection.Data, nil
}

// CollectionItem 获取合集下的视频, seasonId: model.CollectionDataList.ID
// 这个接口没有分页，可能是认为获取的合集数量比较少，不需要分页
func (b *BiliBili) CollectionItem(seasonId string, pn, ps int) (model.CollectionItemData, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.CollectionItemData{}, err
	}
	var item model.CollectionItemResponse
	if err = b.client.
		Get("https://api.bilibili.com/x/space/fav/season/list").
		SetQueryParamsAnyType(map[string]any{
			"season_id": seasonId,
			webLocation: "333.1387",
			"pn":        pn,
			"ps":        ps,
		}).
		SetHeader(Cookie, cookies).
		SetHeaders(publicHeaders()).
		Do().
		Into(&item); err != nil {
		b.logger.Errorf("request collection item error: %v", err)
		return model.CollectionItemData{}, errors.New("获取合集视频列表失败")
	}
	if item.Code != model.SuccessCode {
		b.logger.Errorf("collection item error, code: %d, : %v", item.Code, item.Message)
		return model.CollectionItemData{}, errors.New("获取合集视频列表失败")
	}

	return item.Data, nil
}
