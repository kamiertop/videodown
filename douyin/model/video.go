package model

type VideoDetailResponse struct {
	StatusCode  int       `json:"status_code"`
	AwemeDetail AwemeItem `json:"aweme_detail"`
}

// ImageVideo 动图Video结构
type ImageVideo struct {
	Cover        Cover          `json:"cover"`          // 视频封面
	OriginCover  Cover          `json:"origin_cover"`   // 视频原始封面
	BitRate      []BitRateItem  `json:"bit_rate"`       // 视频不同质量的播放地址信息列表, 和play_addr中的一样
	BigThumbs    []BigThumbItem `json:"big_thumbs"`     // 看了两个都是空的
	Duration     int64          `json:"duration"`       // 视频时长, 单位为秒
	Height       uint32         `json:"height"`         // 视频高度
	Width        uint32         `json:"width"`          // 视频宽度
	PlayAddrH264 PlayInfo       `json:"play_addr_h264"` // 视频播放地址信息
	PlayAddr     PlayInfo       `json:"play_addr"`      // 视频播放地址信息
	HasWatermark bool           `json:"has_watermark"`  // 是否有水印
}
