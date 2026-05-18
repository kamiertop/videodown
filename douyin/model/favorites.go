package model

// CollectListResponse 收藏夹列表响应
type CollectListResponse struct {
	Cursor       int            `json:"cursor"`        // 偏移量
	TotalNumber  int            `json:"total_number"`  // 收藏夹总数
	StatusCode   int            `json:"status_code"`   // 成功是0
	HasMore      bool           `json:"has_more"`      // 是否还有更多数据
	CollectsList []CollectsList `json:"collects_list"` // 收藏夹列表
}

type CollectsList struct {
	AppID                   int      `json:"app_id"`
	CollectsCover           Cover    `json:"collects_cover"`  // 收藏夹封面
	CollectsID              uint     `json:"collects_id"`     // 收藏夹ID
	CollectsIDStr           string   `json:"collects_id_str"` // 收藏夹ID的字符串形式
	CollectsName            string   `json:"collects_name"`   // 收藏夹名称
	CreateTime              int64    `json:"create_time"`     // 创建时间
	FollowStatus            int      `json:"follow_status"`
	FollowedCount           int      `json:"followed_count"`
	IsDefaultCover          bool     `json:"is_default_cover"`
	IsNormalStatus          bool     `json:"is_normal_status"`
	ItemType                int      `json:"item_type"`
	LastCollectTime         int64    `json:"last_collect_time"`
	PlayCount               int      `json:"play_count"`
	SicilyCollectsCoverList any      `json:"sicily_collects_cover_list"`
	States                  int      `json:"states"`
	Status                  int      `json:"status"`
	SystemType              int      `json:"system_type"`
	TotalNumber             int      `json:"total_number"` // 收藏夹内视频总数
	UserID                  uint     `json:"user_id"`
	UserIDStr               string   `json:"user_id_str"` // 用户ID的字符串形式
	UserInfo                UserInfo `json:"user_info"`
}

type UserInfo struct {
	AvatarLarger Avatar `json:"avatar_larger"`
	AvatarMedium Avatar `json:"avatar_medium"`
	AvatarThumb  Avatar `json:"avatar_thumb"`
	Nickname     string `json:"nickname"`
	Uid          string `json:"uid"`
}

type Cover struct {
	URI     string   `json:"uri"`
	UrlList []string `json:"url_list"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
}

// FavoriteVideoResponse 收藏视频列表响应
type FavoriteVideoResponse struct {
	Cursor     uint        `json:"cursor"`      // 下一页的偏移量
	HasMore    int         `json:"has_more"`    // 是否还有更多数据
	StatusCode int         `json:"status_code"` // 成功是0
	Uid        string      `json:"uid"`         // 用户ID
	SecUid     string      `json:"sec_uid"`     // 用户sec_uid, 接口返回的是空字符串
	AwemeList  []AwemeItem `json:"aweme_list"`  // 收藏视频列表
}

// Video 视频信息，图文类型时，play_addr是mp3格式的音频地址
type Video struct {
	Cover         Cover              `json:"cover"`          // 视频封面
	DynamicCover  Cover              `json:"dynamic_cover"`  // 动态封面
	RawCover      Cover              `json:"raw_cover"`      // 原始封面
	BitRate       []BitRateItem      `json:"bit_rate"`       // 视频不同质量的播放地址信息列表
	BitRateAudio  []BitRateAudioItem `json:"bit_rate_audio"` // 音频信息
	BigThumbs     []BigThumbItem     `json:"big_thumbs"`     // 大图信息列表
	Duration      int64              `json:"duration"`       // 视频时长, 单位为秒
	Format        string             `json:"format"`         // 视频格式, 例如 "dash"
	Height        uint32             `json:"height"`         // 视频高度
	Width         uint32             `json:"width"`          // 视频宽度
	PlayAddrH264  PlayInfo           `json:"play_addr_h264"` // 视频播放地址信息
	PlayAddr265   PlayInfo           `json:"play_addr_265"`  // 视频播放地址信息
	PlayAddr      PlayInfo           `json:"play_addr"`      // 视频播放地址信息
	OriginCover   Cover              `json:"origin_cover"`   // 视频原始封面
	GaussianCover Cover              `json:"gaussian_cover"` // 视频高斯模糊封面
}

type BitRateItem struct {
	FPS         int      `json:"FPS"`
	HDRBit      string   `json:"HDR_bit"`
	HDRType     string   `json:"HDR_type"`
	BitRate     int      `json:"bit_rate"`
	Format      string   `json:"format"` // 视频格式, 例如 "dash","mp4"
	GearName    string   `json:"gear_name"`
	IsBytevc1   int      `json:"is_bytevc1"`
	IsH265      int      `json:"is_h265"`
	PlayAddr    PlayInfo `json:"play_addr"`
	QualityType int      `json:"quality_type"`
	VideoExtra  string   `json:"video_extra"`
}
type BitRateAudioItem struct {
	AudioExtra   string    `json:"audio_extra"` // 需要json unmarshal
	AudioMeta    AudioMeta `json:"audio_meta"`
	AudioQuality int       `json:"audio_quality"` // 音频质量
}

type BigThumbItem struct {
	Duration float64  `json:"duration"` // 视频时长, 单位为秒
	Fext     string   `json:"fext"`     // 视频封面图片格式, 例如 "jpg"
	ImgNum   uint     `json:"img_num"`
	ImgUrl   string   `json:"img_url"`
	ImgUrlS  []string `json:"img_urls"`
	ImgXLen  uint     `json:"img_x_len"`
	ImgYLen  uint     `json:"img_y_len"`
	ImgXSize uint     `json:"img_x_size"`
	ImgYSize uint     `json:"img_y_size"`
	Interval float64  `json:"interval"`
	Uri      string   `json:"uri"`
	UriS     []string `json:"uris"`
}

type AudioMeta struct {
	BitRate     int              `json:"bitrate"`      // 音频比特率
	CodecType   string           `json:"codec_type"`   // 音频编码类型，比如：bytevc1
	EncodedType string           `json:"encoded_type"` // normal等
	FileHash    string           `json:"file_hash"`    // 音频文件的Hash值, 用于校验文件完整性
	Format      string           `json:"format"`       // 音频格式, 例如 "dash"
	FileID      string           `json:"file_id"`      // 音频文件ID
	FPS         int              `json:"fps"`          // 音频帧率
	LogoType    string           `json:"logo_type"`    // 音频logo类型, 例如 "normal"
	MediaType   string           `json:"media_type"`   // audio
	Size        uint             `json:"size"`         // 音频数据大小, 单位为字节
	Quality     string           `json:"quality"`      // 音频质量描述，比如：normal
	QualityDesc string           `json:"quality_desc"` // 音频质量描述，比如：普通
	SubInfo     string           `json:"sub_info"`     // 音频额外信息
	UrlList     AudioMetaUrlList `json:"url_list"`     // 音频URL列表
}

type AudioMetaUrlList struct {
	BackupUrl   string `json:"backup_url"`
	FallbackUrl string `json:"fallback_url"`
	MainUrl     string `json:"main_url"`
}

type PlayInfo struct {
	DataSize uint64   `json:"data_size"` // 视频数据大小, 单位为字节
	FileCS   string   `json:"file_cs"`   // 视频文件的CS值, 用于校验文件完整性
	FileHash string   `json:"file_hash"` // 视频文件的Hash值, 用于校验文件完整性
	Height   uint32   `json:"height"`    // 视频高度
	Width    uint32   `json:"width"`     // 视频宽度
	Uri      string   `json:"uri"`
	UrlKey   string   `json:"url_key"`  // 视频URL的Key值
	UrlList  []string `json:"url_list"` // 视频URL列表
}

type VideoStatistics struct {
	CollectCount   uint `json:"collect_count"`   // 收藏数
	CommentCount   uint `json:"comment_count"`   // 评论数
	DiggCount      uint `json:"digg_count"`      // 点赞数
	RecommendCount uint `json:"recommend_count"` // 推荐数
	ShareCount     uint `json:"share_count"`     // 分享数
}

type Author struct {
	CoverURL       []Avatar `json:"cover_url"`       // 作者头像列表
	AvatarThumb    Avatar   `json:"avatar_thumb"`    // 作者头像
	AvatarLarger   Avatar   `json:"avatar_larger"`   // 作者头像
	FollowStatus   int      `json:"follow_status"`   // 关注状态
	FollowerStatus int      `json:"follower_status"` // 粉丝状态
	Nickname       string   `json:"nickname"`        // 作者昵称
	Uid            string   `json:"uid"`             // 作者用户ID
	SecUid         string   `json:"sec_uid"`         // 作者sec_uid
}
