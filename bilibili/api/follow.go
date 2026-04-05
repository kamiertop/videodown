package api

import (
	"github.com/kamiertop/videodown/bilibili/model"
)

func (b *BiliBili) FollowList(pn, ps int) (model.FollowData, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.FollowData{}, err
	}
	myMid, err := b.getMid()
	if err != nil {
		return model.FollowData{}, err
	}
	var response model.FollowResponse
	err = b.client.
		Get("https://api.bilibili.com/x/relation/followings").
		SetQueryParamsAnyType(map[string]any{
			"pn":          pn,
			"ps":          ps,
			"order":       "desc",
			"order_type":  "",
			"vmid":        myMid,
			"gaia_source": "main_web",
			webLocation:   "333.1387",
		}).
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookies).
		Do().
		Into(&response)
	if err != nil {
		b.logger.Errorf("request follow list error: %v", err)
		return model.FollowData{}, err
	}

	return response.Data, nil
}
