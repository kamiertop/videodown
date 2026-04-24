package api

import (
	"errors"
	"fmt"

	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/utils"
)

// Info 用户空间
func (b *BiliBili) Info(spaceURLOrMid string) (model.UserInfoData, error) {
	var resp struct {
		model.ApiResponse
		Data model.UserInfoData `json:"data"`
	}
	userMid, err := utils.ParseMid(spaceURLOrMid)
	if err != nil {
		return resp.Data, err
	}

	// 获取 WBI 密钥

	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}
	// 组装参数
	params := map[string]string{
		"mid":              userMid,
		"token":            "",
		"platform":         "web",
		"web_location":     "1550101",
		"dm_img_list":      "[]",
		"dm_img_str":       "V2ViR0wgMS4wIChPcGVuR0wgRVMgMi4wIENocm9taXVtKQ",
		"dm_cover_img_str": "QU5HTEUgKE1pY3Jvc29mdCwgTWljcm9zb2Z0IEJhc2ljIFJlbmRlciBEcml2ZXIgKDB4MDAwMDAwOEMpIERpcmVjdDNEMTEgdnNfNV8wIHBzXzVfMCwgRDNEMTEpR29vZ2xlIEluYy4gKE1pY3Jvc29mdC",
		"dm_img_inter":     "%7B%22ds%22:[],%22wh%22:[3412,1439,78],%22of%22:[124,248,124]%7D",
	}

	// 添加 WBI 签名
	params, err = b.encWbiParams(params)
	if err != nil {
		return resp.Data, err
	}

	// 发起请求
	err = b.client.
		Get("https://api.bilibili.com/x/space/wbi/acc/info").
		SetQueryParams(params).
		SetHeader(Cookie, cookies).
		SetHeader(Origin, spaceOrigin).
		SetHeader(Referer, fmt.Sprintf("https://space.bilibili.com/%s", userMid)).
		SetHeaders(publicHeaders()).
		Do().Into(&resp)
	if err != nil {
		return resp.Data, errors.New("请求用户信息接口失败")
	}
	if resp.Code != 0 {
		return resp.Data, errors.New("请求用户信息接口错误: " + resp.Message)
	}

	return resp.Data, nil
}
