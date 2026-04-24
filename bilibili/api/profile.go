package api

import (
	"errors"

	"github.com/kamiertop/videodown/bilibili/model"
)

// MyInfo 个人资料
func (b *BiliBili) MyInfo() (model.MyInfoProfile, error) {
	var resp struct {
		model.ApiResponse
		Data model.MyInfoData `json:"data"`
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data.Profile, err
	}
	if err = b.client.
		Get("https://api.bilibili.com/x/space/v2/myinfo").
		SetQueryParam("web_location", "333.1007").
		SetHeaders(publicHeaders()).
		SetHeader(Origin, spaceOrigin).
		SetHeader("Cookie", cookies).
		Do().
		Into(&resp); err != nil {
		b.logger.Errorf("failed to get my_info: %v", err)
		return resp.Data.Profile, errors.New("获取个人资料失败")
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("my_info request failed: code=%d message=%s", resp.Code, resp.Message)
		return resp.Data.Profile, errors.New("获取个人资料失败")
	}

	if resp.Data.Profile.Mid != 0 {
		if err := b.saveMid(resp.Data.Profile.Mid); err != nil {
			b.logger.Errorf("failed to save mid from my_info: %v", err)
			return resp.Data.Profile, errors.New("保存用户ID失败")
		}
	}

	return resp.Data.Profile, nil
}
