package api

import (
	"errors"

	"github.com/kamiertop/videodown/bilibili/model"
)

// MyInfo 个人资料
func (b *BiliBili) MyInfo() (model.MyInfoProfile, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.MyInfoProfile{}, err
	}
	var response model.MyInfoResponse
	if err = b.client.
		Get("https://api.bilibili.com/x/space/v2/myinfo").
		SetQueryParam("web_location", "333.1007").
		SetHeaders(publicHeaders()).
		SetHeader("Cookie", cookies).
		Do().
		Into(&response); err != nil {
		b.logger.Errorf("failed to get my_info: %v", err)
		return model.MyInfoProfile{}, errors.New("获取个人资料失败")
	}
	if response.Code != model.SuccessCode {
		b.logger.Errorf("my_info request failed: code=%d message=%s", response.Code, response.Message)
		return model.MyInfoProfile{}, errors.New("获取个人资料失败")
	}

	if response.Data.Profile.Mid != 0 {
		if err := b.saveMid(response.Data.Profile.Mid); err != nil {
			b.logger.Errorf("failed to save mid from my_info: %v", err)
			return model.MyInfoProfile{}, errors.New("保存用户ID失败")
		}
	}

	return response.Data.Profile, nil
}
