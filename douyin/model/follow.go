package model

// FollowResponse 关注列表
type FollowResponse struct {
	Followings   []FollowItem `json:"followings"`
	HasMore      bool         `json:"has_more"`    // 是否还有更多数据
	StatusCode   int          `json:"status_code"` // 成功是0
	Total        int          `json:"total"`       // 总关注人数，和mix_count相等
	MixCount     int          `json:"mix_count"`
	MyselfUserId string       `json:"myself_user_id"` // 自己的用户ID
	Offset       int          `json:"offset"`         // 当前页的偏移量，默认是20，即每次请求20条数据
}

type FollowItem struct {
	Signature     string `json:"signature"`   // 个性签名
	AwemeCount    int    `json:"aweme_count"` // 作品数量
	Nickname      string `json:"nickname"`    // 昵称
	SecUid        string `json:"sec_uid"`
	Uid           string `json:"uid"`
	UniqueId      string `json:"unique_id"`
	ShortId       string `json:"short_id"` // 和unique_id一样
	AvatarLarger  Avatar `json:"avatar_larger"`
	AvatarMedium  Avatar `json:"avatar_medium"`
	AvatarThumb   Avatar `json:"avatar_thumb"`
	FollowerCount int    `json:"follower_count"` // 粉丝数量
}

type Avatar struct {
	Height  int      `json:"height"`
	Width   int      `json:"width"`
	Uri     string   `json:"uri"`
	UrlList []string `json:"url_list"`
}

type SearchFollowResponse struct {
	NickName     string `json:"rich_sug_nickname"`      // 昵称
	UserID       string `json:"rich_sug_user_id"`       // 用户ID
	SecUid       string `json:"rich_sug_sec_uid"`       // 用户sec_uid
	FollowStatus string `json:"rich_sug_follow_status"` // 关注状态，follow表示已关注
	RelationType string `json:"rich_sug_relation_type"` // 关系类型，关注
	AvatarURI    string `json:"rich_sug_avatar_uri"`    // 头像URI
	ShortID      string `json:"rich_sug_short_id"`      // 抖音号
}

// DynamicResponse 关注动态列表响应
type DynamicResponse struct {
	Cursor     int64         `json:"cursor"`
	HasMore    int           `json:"has_more"`
	StatusCode int           `json:"status_code"`
	Data       []DynamicItem `json:"data"`
}

type DynamicItem struct {
	Aweme        AwemeItem `json:"aweme"`
	CommentCount int       `json:"comment_count"`
}
