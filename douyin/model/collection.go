package model

type CollectionResponse struct {
	Cursor     int              `json:"cursor"`      // 下一页的偏移量
	HasMore    int              `json:"has_more"`    // 是否还有更多数据
	StatusCode int              `json:"status_code"` // 成功是0
	MixInfos   []CollectionItem `json:"mix_infos"`   // 合集列表
}

type CollectionItem struct {
	Author      Author              `json:"author"`      // 合集作者信息
	CoverUrl    Cover               `json:"cover_url"`   // 合集封面URL
	CreateTime  int                 `json:"create_time"` // 合集创建时间
	Desc        string              `json:"desc"`        // 合集描述
	IsSerialMix int                 `json:"is_serial_mix"`
	MixId       string              `json:"mix_id"`   // 合集ID
	MixName     string              `json:"mix_name"` // 合集名称
	MixType     int                 `json:"mix_type"` // 合集类型，暂未探明
	UpdateTime  int                 `json:"update_time"`
	Statis      CollectionStatistic `json:"statis"` // 合集统计信息
}

// CollectionStatistic 合集统计信息，下面两个字段一样
type CollectionStatistic struct {
	HasUpdatedEpisode int `json:"has_updated_episode"` // 合集数量
	UpdatedToEpisode  int `json:"updated_to_episode"`  // 合集数量
}

// CollectionListResponse 获取合集下的视频列表响应
type CollectionListResponse struct {
	StatusCode int         `json:"status_code"`
	StatusMsg  string      `json:"status_msg"`
	HasMore    int         `json:"has_more"`
	MaxCursor  int         `json:"max_cursor"`
	MinCursor  int         `json:"min_cursor"`
	AwemeList  []AwemeItem `json:"aweme_list"`
}
