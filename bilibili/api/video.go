package api

import (
	"errors"
	"strconv"

	"github.com/kamiertop/videodown/bilibili/model"
)

func (b *BiliBili) VideoList(mid uint, ps, pn int) (model.VideoListData, error) {
	cookies, err := b.getCookies()
	if err != nil {
		return model.VideoListData{}, err
	}
	params := map[string]string{
		"mid":              strconv.FormatUint(uint64(mid), 10),
		"ps":               strconv.Itoa(ps),
		"pn":               strconv.Itoa(pn),
		"order_avoided":    "true",
		"platform":         "web",
		"web_location":     "333.1387",
		"dm_img_list":      "[]",
		"dm_img_str":       "V2ViR0wgMS4wIChPcGVuR0wgRVMgMi4wIENocm9taXVtKQ",
		"dm_cover_img_str": "QU5HTEUgKE1pY3Jvc29mdCwgTWljcm9zb2Z0IEJhc2ljIFJlbmRlciBEcml2ZXIgKDB4MDAwMDAwOEMpIERpcmVjdDNEMTEgdnNfNV8wIHBzXzVfMCwgRDNEMTEpR29vZ2xlIEluYy4gKE1pY3Jvc29mdC",
		"dm_img_inter":     "%7B%22ds%22:[],%22wh%22:[3412,1439,78],%22of%22:[124,248,124]%7D",
	}

	params, err = b.encWbiParams(params)
	if err != nil {
		return model.VideoListData{}, err
	}
	var resp struct {
		Code    int                 `json:"code"`
		Message string              `json:"message"`
		TTL     int                 `json:"ttl"`
		Data    model.VideoListData `json:"data"`
	}
	err = b.client.
		Get("https://api.bilibili.com/x/space/wbi/arc/search").
		SetQueryParams(params).
		SetHeader(Cookie, cookies).
		SetHeaders(publicHeaders()).
		Do().
		Into(&resp)
	if err != nil {
		return model.VideoListData{}, err
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("request video list error, code: %d, message: %s", resp.Code, resp.Message)
		return model.VideoListData{}, errors.New("请求视频列表接口错误: " + resp.Message)
	}

	return resp.Data, nil
}
