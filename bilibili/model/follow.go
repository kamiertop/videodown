package model

// FollowResponse 关注列表， 只记录部分重要字段
type FollowResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	TTL     int        `json:"ttl"`
	Data    FollowData `json:"data"`
}

type FollowData struct {
	List      []FollowDataList `json:"list"`
	ReVersion int              `json:"re_version"` // 未知
	Total     int              `json:"total"`      // 自己关注了多少人
}

type FollowDataList struct {
	Mid       int64  `json:"mid"`       // 关注的用户的mid
	Attribute int    `json:"attribute"` // 对方对于自己的关系属性. 0： 未关注， 2： 已关注, 6: 互粉, 128: 已拉黑
	Mtime     int    `json:"mtime"`     // 关注时间
	Tag       any    `json:"tag"`
	Special   int    `json:"special"`
	Uname     string `json:"uname"` // 用户昵称
	Face      string `json:"face"`  // 用户头像url
	Sign      string `json:"sign"`  // 用户签名
}
