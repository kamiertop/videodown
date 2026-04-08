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
		model.ApiResponse
		Data model.VideoListData `json:"data"`
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

func (b *BiliBili) VideoDetailConcise(aid int) (model.VideoDetailConciseData, error) {
	var resp struct {
		model.ApiResponse
		Data model.VideoDetailConciseData `json:"data"`
	}
	params := map[string]string{
		"aid":           strconv.Itoa(aid),
		"need_view":     "1",
		"isGaiaAvoided": "false",
		webLocation:     "1315873",
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}
	params, err = b.encWbiParams(params)
	if err != nil {
		return resp.Data, err
	}

	err = b.client.
		Get("https://api.bilibili.com/x/web-interface/wbi/view/detail").
		SetQueryParams(params).
		SetHeader(Cookie, cookies).
		SetHeaders(publicHeaders()).
		Do().
		Into(&resp)
	if err != nil {
		return resp.Data, err
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("request video detail error, code: %d, message: %s", resp.Code, resp.Message)
		return resp.Data, errors.New("请求视频详情接口错误: " + resp.Message)
	}

	return resp.Data, nil
}

// VideoDetailConciseBvid 通过 BV 号获取精简详情（便于粘贴链接）
func (b *BiliBili) VideoDetailConciseBvid(bvid string) (model.VideoDetailConciseData, error) {
	var resp struct {
		model.ApiResponse
		Data model.VideoDetailConciseData `json:"data"`
	}
	params := map[string]string{
		"bvid":          bvid,
		"need_view":     "1",
		"isGaiaAvoided": "false",
		webLocation:     "1315873",
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}
	params, err = b.encWbiParams(params)
	if err != nil {
		return resp.Data, err
	}

	err = b.client.
		Get("https://api.bilibili.com/x/web-interface/wbi/view/detail").
		SetQueryParams(params).
		SetHeader(Cookie, cookies).
		SetHeaders(publicHeaders()).
		Do().
		Into(&resp)
	if err != nil {
		return resp.Data, err
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("request video detail error, code: %d, message: %s", resp.Code, resp.Message)
		return resp.Data, errors.New("请求视频详情接口错误: " + resp.Message)
	}

	return resp.Data, nil
}

// VideoPlayURL 获取播放地址。qn 为清晰度代码（如 80=1080P）；fnval 为流格式掩码（如 4048=DASH，0=FLV，80=MP4 等）。qn≤0 时默认 80，fnval<0 时默认 4048（0 表示 FLV 合法，不可再用负数表示默认）。
func (b *BiliBili) VideoPlayURL(avid int64, bvid string, cid int64, qn int, fnval int) (model.VideoURLData, error) {
	var resp struct {
		model.ApiResponse
		Data model.VideoURLData `json:"data"`
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}
	if qn <= 0 {
		qn = 80
	}
	if fnval < 0 {
		fnval = 4048
	}
	params := map[string]string{
		"avid":          strconv.FormatInt(avid, 10),
		"bvid":          bvid,
		"cid":           strconv.FormatInt(cid, 10),
		"qn":            strconv.Itoa(qn),
		"fnval":         strconv.Itoa(fnval),
		"fnver":         "0",
		"fourk":         "1",
		"voice_balance": "1",
		"gaia_source":   "pre-load",
		"isGaiaAvoided": "true",
		webLocation:     "1315873",
	}
	params, err = b.encWbiParams(params)
	if err != nil {
		return resp.Data, err
	}
	err = b.client.
		Get("https://api.bilibili.com/x/player/wbi/playurl").
		SetQueryParams(params).
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookies).
		Do().
		Into(&resp)
	if err != nil {
		b.logger.Errorf("request video url error, err: %v", err)
		return resp.Data, err
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("request video url error, code: %d, message: %s", resp.Code, resp.Message)
		return resp.Data, errors.New("请求视频播放链接接口错误: " + resp.Message)
	}

	return resp.Data, nil
}
