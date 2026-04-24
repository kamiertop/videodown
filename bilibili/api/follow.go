package api

import (
	"fmt"

	"github.com/kamiertop/videodown/bilibili/model"
)

func (b *BiliBili) FollowList(pn, ps int) (model.FollowData, error) {
	var resp struct {
		model.ApiResponse
		Data model.FollowData `json:"data"`
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
		SetHeader(Origin, spaceOrigin).
		SetHeader(Referer, fmt.Sprintf("https://space.bilibili.com/%s/relation/follow", myMid)).
		SetHeader(Cookie, cookies).
		Do().
		Into(&resp)
	if err != nil {
		b.logger.Errorf("request follow list api error: %v", err)
		return resp.Data, err
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("request follow list error, code: %d, message: %s", resp.Code, resp.Message)
	}

	return resp.Data, nil
}
