package model

// Follow 关注列表
type Follow struct {
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
	CoverURL      Avatar `json:"cover_url"`
	FollowerCount int    `json:"follower_count"` // 粉丝数量
}

type Avatar struct {
	Height  int      `json:"height"`
	Uri     string   `json:"uri"`
	Width   int      `json:"width"`
	UrlList []string `json:"url_list"`
}
