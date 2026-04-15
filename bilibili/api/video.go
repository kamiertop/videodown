package api

import (
	"errors"
	"strconv"

	"github.com/kamiertop/videodown/bilibili/model"
)

func (b *BiliBili) VideoList(mid uint, ps, pn int) (model.VideoListData, error) {
	var resp struct {
		model.ApiResponse
		Data model.VideoListData `json:"data"`
	}
	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
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
		return resp.Data, err
	}
	err = b.client.
		Get("https://api.bilibili.com/x/space/wbi/arc/search").
		SetQueryParams(params).
		SetHeader(Cookie, cookies).
		SetHeaders(publicHeaders()).
		Do().
		Into(&resp)
	if err != nil {
		b.logger.Errorf("request video list api error: %v", err)
		return resp.Data, err
	}
	if resp.Code != model.SuccessCode {
		b.logger.Errorf("request video list error, code: %d, message: %s", resp.Code, resp.Message)
		return resp.Data, errors.New("请求视频列表接口错误: " + resp.Message)
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
		"bvid":                bvid,
		"platform":            "web",
		"need_operation_card": "1",
		"web_rn_repeat":       "1",
		"need_elec":           "1",
		"out_referer":         "https://www.bilibili.com/",
		"page_no":             "1",
		"p":                   "1",
		"gaia_source":         "main_web",
		"need_view":           "1",
		"isGaiaAvoided":       "false",
		webLocation:           "1315873",
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
		"qn":            strconv.Itoa(qn),    // 视频清晰度选择
		"fnval":         strconv.Itoa(fnval), // 视频流格式标识
		"fnver":         "0",                 // 0
		"fourk":         "1",                 // 是否允许4k视频，这里默认就是允许
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

/*
视频为DASH或MP4格式
新的视频中，较高分辨率均采用dash，dash格式会取到所有分辨率的流地址，较低分辨率与老视频还保留了MP4格式
qn：清晰度标识
80: 1080P 高清
100：智能修复，人工智能增强画质（大会员）
112 1080P+高码率（大会员认证）
116：1080P高帧率 大会员认证
120：4K 超清 大会员认证,且fourk=1
125：4K HDR真彩，仅支持DASH格式，需要fnval&64=64（大会员认证）
126：杜比视界，仅支持DASH格式， 需要fnval&512=512（大会员认证）
127：8K超高清，仅支持DASH格式，需要fnval&1024=1024（大会员认证）
129: HDR Vivid 大会员认证

fnval: 视频流格式标识
0: FLV格式，已下线
1：MP4格式，仅H.264编码
16：DASH格式，与MP4格式互斥
64：是否需求HDR视频，需求DASH格式， 仅H.265编码，需要qn=125，大会员认证
128：是否需求 4K 分辨率，该值与fourk字段协同作用，需要qn=120，大会员认证
256：是否需求杜比音频， 需求DASH格式，大会员认证
512：是否需求杜比视界， 需求DASH格式，大会员认证
1024：是否需求8K分辨率， 需求DASH格式，大会员认证，需要qn=127
2048：是否需求AV1编码，需求DASH格式
4048： 所有可用DASH视频流，即一次性返回所有可用DASH格式视频流
16384：是否需要HDR Vivid, 仅APP接口可用

视频编码代码
7：AVC编码，8k视频不支持该格式
12：HEVC编码
13：AV1编码

视频伴音音质代码
30126: 64k
30232: 132k
30280: 192k
30250: 杜比全景声
30251： Hi-Res无损
*/
