package api

import (
	"errors"
	"net/url"
	"strings"

	"github.com/kamiertop/videodown/bilibili/model"
)

// 用户空间

func (b *BiliBili) Info(spaceURLOrMid string) (model.UserInfoData, error) {
	var resp model.UserInfoResponse
	userMid, err := parseMid(spaceURLOrMid)
	if err != nil {
		return resp.Data, err
	}

	// 获取 WBI 密钥
	wbiKeys, err := b.getWbiKeys()
	if err != nil {
		return resp.Data, err
	}
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
	params = encWbiParams(params, wbiKeys.ImgKey, wbiKeys.SubKey)

	// 发起请求
	err = b.client.
		Get("https://api.bilibili.com/x/space/wbi/acc/info").
		SetQueryParams(params).
		SetHeader(Cookie, cookies).
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

func parseMid(spaceURLOrMid string) (string, error) {
	spaceURLOrMid = strings.TrimSpace(spaceURLOrMid)
	if spaceURLOrMid == "" {
		return "", errors.New("解析到mid为空，请输入正确的空间链接或mid")
	}

	if strings.Contains(spaceURLOrMid, "space.bilibili.com") {
		// 兼容不带协议的输入：space.bilibili.com/xxxxx
		if !strings.HasPrefix(spaceURLOrMid, "http://") && !strings.HasPrefix(spaceURLOrMid, "https://") {
			spaceURLOrMid = "https://" + spaceURLOrMid
		}
		parseUrl, err := url.Parse(spaceURLOrMid)
		if err != nil {
			return "", err
		}

		spaceURLOrMid = parseUrl.Path[1:]
	}

	if len(spaceURLOrMid) == 0 {
		return "", errors.New("解析到mid为空，请输入正确的空间链接或mid")
	}

	for _, c := range spaceURLOrMid {
		if c < '0' || c > '9' {
			return "", errors.New("解析到mid包含非数字字符，请输入正确的空间链接或mid")
		}
	}

	return spaceURLOrMid, nil
}
