package model

// ImageVideo 动图Video结构
type ImageVideo struct {
	Cover        Cover         `json:"cover"`          // 视频封面
	OriginCover  Cover         `json:"origin_cover"`   // 视频原始封面
	BitRate      []BitRateItem `json:"bit_rate"`       // 视频不同质量的播放地址信息列表, 和play_addr中的一样
	Duration     int64         `json:"duration"`       // 视频时长, 单位为秒
	Height       uint32        `json:"height"`         // 视频高度
	Width        uint32        `json:"width"`          // 视频宽度
	PlayAddrH264 PlayInfo      `json:"play_addr_h264"` // 视频播放地址信息
	PlayAddr     PlayInfo      `json:"play_addr"`      // 视频播放地址信息
	HasWatermark bool          `json:"has_watermark"`  // 是否有水印
}

// AwemeItem 视频信息
type AwemeItem struct {
	AwemeId                  string          `json:"aweme_id"`    // 视频ID
	Desc                     string          `json:"desc"`        // 视频标题+#标签
	CreateTime               int             `json:"create_time"` // 视频创建时间
	Author                   Author          `json:"author"`      // 视频作者信息
	Music                    Music           `json:"music"`
	Video                    Video           `json:"video"` // 视频信息
	Statistics               VideoStatistics `json:"statistics"`
	Duration                 int             `json:"duration"` // 视频时长, 单位为毫秒
	AwemeType                int             `json:"aweme_type"`
	AuthorUserId             int64           `json:"author_user_id"` // 视频作者的用户ID
	Region                   string          `json:"region"`         // 视频发布的地区:CN等
	GroupId                  string          `json:"group_id"`
	PreventDownload          bool            `json:"prevent_download"`
	IsMomentHistory          int             `json:"is_moment_history"`
	IsMomentStory            int             `json:"is_moment_story"`
	SecItemId                string          `json:"sec_item_id"`
	ItemAigcFollowShot       int             `json:"item_aigc_follow_shot"`
	Images                   []ImageItem     `json:"images"` // 图文类型视频使用这个字段中的图片列表
	OriginDuetResourceUri    string          `json:"origin_duet_resource_uri"`
	IsImageBeat              bool            `json:"is_image_beat"`
	IsTop                    int8            `json:"is_top"` // 视频置顶状态，1是置顶，0不是
	IsLifeItem               bool            `json:"is_life_item"`
	AuthorMaskTag            int             `json:"author_mask_tag"`
	UserRecommendStatus      int             `json:"user_recommend_status"`
	CollectionCornerMark     int             `json:"collection_corner_mark"`
	IsSharePost              bool            `json:"is_share_post"`
	AuthenticationToken      string          `json:"authentication_token"` // 视频的认证token
	MediaType                int             `json:"media_type"`           // 视频类型，2图文，4视频
	ActivityVideoType        int             `json:"activity_video_type"`
	BoostStatus              int             `json:"boost_status"`
	Caption                  string          `json:"caption"`    // 视频标题，带标签，可能为空
	ItemTitle                string          `json:"item_title"` // 子标题，可能为空
	Original                 int             `json:"original"`
	LunaVideoCandidateStatus string          `json:"luna_video_candidate_status,omitempty"`
	IsMultiContent           int             `json:"is_multi_content,omitempty"`
	ImageItemQualityLevel    int             `json:"image_item_quality_level,omitempty"`
	IsLivePhoto              int             `json:"is_live_photo"` // 是否是动图，1是，0不是
	IsSliedes                bool            `json:"is_slides"`     // 有这个字段，一定是动图，并且是多张动图，没有这个的也不一定不是动图
}
