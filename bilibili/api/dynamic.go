package api

import (
	"errors"

	"github.com/kamiertop/videodown/bilibili/model"
)

const dynamicFeedURL = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all"

// dynamic 获取当前登录账号关注动态中的视频动态列表页。
// offset 首次请求传空字符串，后续传上次响应里的 offset。
func (b *BiliBili) dynamic(offset string) (model.DynamicData, error) {
	var resp struct {
		model.ApiResponse
		Data model.DynamicData `json:"data"`
	}

	cookies, err := b.getCookies()
	if err != nil {
		return resp.Data, err
	}

	params := map[string]string{
		"type":      "video",
		"platform":  "web",
		webLocation: "333.1365",
		"features":  "itemOpusStyle,listOnlyfans,opusBigCover,onlyfansVote,decorationCard,onlyfansAssetsV2,forwardListHidden,ugcDelete",
	}
	if offset != "" {
		params["offset"] = offset
	}

	if err = b.client.
		Get(dynamicFeedURL).
		SetQueryParams(params).
		SetHeaders(publicHeaders()).
		SetHeader(Cookie, cookies).
		SetHeader(Origin, biliBiliUrl).
		SetHeader(Referer, biliBiliUrl).
		Do().
		Into(&resp); err != nil {
		b.logger.Errorf("request dynamic video feed error: %v", err)
		return resp.Data, errors.New("获取关注动态视频失败")
	}

	switch resp.Code {
	case model.SuccessCode:
		return resp.Data, nil
	case -101:
		return resp.Data, errors.New("未登录，请先登录")
	default:
		b.logger.Errorf("request dynamic video feed failed: code=%d message=%s", resp.Code, resp.Message)
		return resp.Data, errors.New("获取关注动态视频失败: " + resp.Message)
	}
}

// DynamicVideos 获取当前登录账号关注动态中的全部视频动态。
func (b *BiliBili) DynamicVideos() ([]model.DynamicArchiveItem, error) {
	var videos []model.DynamicArchiveItem
	offset := ""

	for {
		page, err := b.DynamicVideosPage(offset)
		if err != nil {
			return videos, err
		}

		videos = append(videos, page.Items...)
		if !page.HasMore || page.Offset == "" || page.Offset == offset {
			break
		}
		offset = page.Offset
	}

	return videos, nil
}

// DynamicVideosPage 获取当前登录账号关注动态中的单页视频动态。
func (b *BiliBili) DynamicVideosPage(offset string) (model.DynamicArchivePage, error) {
	data, err := b.dynamic(offset)
	if err != nil {
		return model.DynamicArchivePage{}, err
	}

	return model.DynamicArchivePage{
		Items:   data.GetArchiveItems(),
		Offset:  data.Offset,
		HasMore: data.HasMore,
	}, nil
}
