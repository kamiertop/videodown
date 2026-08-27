package model

// HistoryResponse 观看历史接口的响应
type HistoryResponse struct {
	HasMore    uint8         `json:"has_more"`    // 是否有更多数据，1表示有，0表示没有
	MaxCursor  int64         `json:"max_cursor"`  // 最大游标，下一页请求时传入
	StatusCode int           `json:"status_code"` // 状态码，成功时为0
	StatusMsg  string        `json:"status_msg"`  // 状态信息，成功时为""
	AwemeList  []HistoryItem `json:"aweme_list"`  // 观看历史列表
}
type HistoryItem struct {
	Author      Author          `json:"author"`
	AwemeId     string          `json:"aweme_id"`      // 视频ID
	AwemeType   int             `json:"aweme_type"`    // 抖音业务类型码，辅助排查类型差异
	Caption     string          `json:"caption"`       // 标题，可能为空
	CreateTime  int             `json:"create_time"`   // 发布时间
	Desc        string          `json:"desc"`          // 标题+#标签
	Images      []ImageItem     `json:"images"`        // 图文/动图素材列表
	IsLivePhoto int             `json:"is_live_photo"` // 1 表示动图
	IsSliedes   bool            `json:"is_slides"`     // 多张动图信号
	IsTop       int             `json:"is_top"`        // 置顶状态，1 是置顶
	ItemTitle   string          `json:"item_title"`    // 子标题，可能为空
	MediaType   int             `json:"media_type"`    // 2 图文，4 普通视频，42 常见于动图/Live Photo
	Music       Music           `json:"music"`
	Statistics  VideoStatistics `json:"statistics"` // 视频统计信息
	Video       Video           `json:"video"`
}

// SearchHistoryResponse 搜索观看历史接口的响应
type SearchHistoryResponse struct {
	StatusCode int           `json:"status_code"`
	Cursor     int           `json:"cursor"`
	HasMore    int           `json:"has_more"`
	AwemeList  []HistoryItem `json:"aweme_list"`
}
