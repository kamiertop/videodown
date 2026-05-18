package model

type UserResponse struct {
	User       User `json:"user"`
	StatusCode int  `json:"status_code"`
}

type User struct {
	AvatarLarger   Avatar `json:"avatar_larger"`
	AvatarMedium   Avatar `json:"avatar_medium"`
	AvatarThumb    Avatar `json:"avatar_thumb"`
	City           string `json:"city"`
	AwemeCount     int    `json:"aweme_count"` // 作品数量
	Country        string `json:"country"`
	FollowerCount  int    `json:"follower_count"`  // 粉丝数量
	FollowingCount int    `json:"following_count"` // 关注数量
	IpLocation     string `json:"ip_location"`     // IP所在地
	NickName       string `json:"nickname"`
	Province       string `json:"province"`  // 省份
	Signature      string `json:"signature"` // 签名
	Uid            string `json:"uid"`
	UniqueId       string `json:"unique_id"` // 抖音号
	SecUid         string `json:"sec_uid"`   // 安全uid，某些接口需要传这个
}

// UserVideoListResponse 用户空间的视频列表响应
type UserVideoListResponse struct {
	AwemeList          []AwemeItem `json:"aweme_list"`
	HasMore            int         `json:"has_more"`    // 是否还有更多数据，0没有，1有
	StatusCode         int         `json:"status_code"` // 成功是0
	MinCursor          int         `json:"min_cursor"`  // 上一页的请求参数
	MaxCursor          int64       `json:"max_cursor"`  // 作为下一页的请求参数
	PostSerial         int         `json:"post_serial"`
	RequestItemCursor  int         `json:"request_item_cursor"`
	ReplaceSeriesCover int         `json:"replace_series_cover"`
}

type ImageItem struct {
	UrlList       []string   `json:"url_list"`        /// 不带水印的图片下载地址列表
	ClipType      int        `json:"clip_type"`       // 剪辑类型 5：动图; 2: 普通图片
	LivePhotoType int        `json:"live_photo_type"` // 1: 动图
	Video         ImageVideo `json:"video"`           // 动图视频信息，clip_type=5时存在
}

type PlayUrl struct {
	Uri     string   `json:"uri"`
	UrlList []string `json:"url_list"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	UrlKey  string   `json:"url_key"`
}

type Music struct {
	Id                             int64           `json:"id"`
	IdStr                          string          `json:"id_str"`
	Title                          string          `json:"title"`
	Author                         string          `json:"author"`
	Album                          string          `json:"album"`
	CoverHd                        Cover           `json:"cover_hd,omitempty"`
	CoverLarge                     Cover           `json:"cover_large,omitempty"`
	CoverMedium                    Cover           `json:"cover_medium,omitempty"`
	CoverThumb                     Cover           `json:"cover_thumb"`
	PlayUrl                        PlayUrl         `json:"play_url"`
	SchemaUrl                      string          `json:"schema_url"`
	SourcePlatform                 int             `json:"source_platform"`
	StartTime                      int             `json:"start_time"`
	EndTime                        int             `json:"end_time"`
	Duration                       int             `json:"duration"`
	UserCount                      int             `json:"user_count"`
	CollectStat                    int             `json:"collect_stat"`
	Status                         int             `json:"status"`
	OfflineDesc                    string          `json:"offline_desc"`
	OwnerId                        string          `json:"owner_id"`
	OwnerNickname                  string          `json:"owner_nickname"`
	IsOriginal                     bool            `json:"is_original"`
	Mid                            string          `json:"mid"`
	BindedChallengeId              int             `json:"binded_challenge_id"`
	Redirect                       bool            `json:"redirect"`
	IsRestricted                   bool            `json:"is_restricted"`
	AuthorDeleted                  bool            `json:"author_deleted"`
	IsDelVideo                     bool            `json:"is_del_video"`
	IsVideoSelfSee                 bool            `json:"is_video_self_see"`
	OwnerHandle                    string          `json:"owner_handle"`
	PreventDownload                bool            `json:"prevent_download"`
	PreventItemDownloadStatus      int             `json:"prevent_item_download_status"`
	SecUid                         string          `json:"sec_uid"`
	AvatarThumb                    Avatar          `json:"avatar_thumb"`
	AvatarMedium                   Avatar          `json:"avatar_medium"`
	AvatarLarge                    Avatar          `json:"avatar_large"`
	IsCommerceMusic                bool            `json:"is_commerce_music"`
	IsOriginalSound                bool            `json:"is_original_sound"`
	AuditionDuration               int             `json:"audition_duration"`
	ShootDuration                  int             `json:"shoot_duration"`
	ReasonType                     int             `json:"reason_type"`
	DmvAutoShow                    bool            `json:"dmv_auto_show"`
	IsPgc                          bool            `json:"is_pgc"`
	IsMatchedMetadata              bool            `json:"is_matched_metadata"`
	IsAudioUrlWithCookie           bool            `json:"is_audio_url_with_cookie"`
	CanBackgroundPlay              bool            `json:"can_background_play"`
	MusicStatus                    int             `json:"music_status"`
	VideoDuration                  int             `json:"video_duration"`
	PgcMusicType                   int             `json:"pgc_music_type"`
	AuthorStatus                   int             `json:"author_status"`
	DspStatus                      int             `json:"dsp_status"`
	MusicCollectCount              int             `json:"music_collect_count"`
	MusicCoverAtmosphereColorValue string          `json:"music_cover_atmosphere_color_value"`
	Song                           Song            `json:"song,omitempty"`
	MusicImageBeats                MusicImageBeats `json:"music_image_beats,omitempty"`
	MatchedPgcSound                MatchedPgcSound `json:"matched_pgc_sound,omitempty"`
}

type MusicImageBeats struct {
	MusicImageBeatsUrl Cover `json:"music_image_beats_url"`
}

type Song struct {
	Id     int64  `json:"id"`
	IdStr  string `json:"id_str"`
	Title  string `json:"title,omitempty"`
	Chorus Chorus `json:"chorus,omitempty"`
}

type Chorus struct {
	StartMs    int `json:"start_ms"`
	DurationMs int `json:"duration_ms"`
}

type MatchedPgcSound struct {
	Author      string `json:"author"`
	Title       string `json:"title"`
	MixedTitle  string `json:"mixed_title"`
	MixedAuthor string `json:"mixed_author"`
	CoverMedium Cover  `json:"cover_medium"`
}

// UserSeriesResponse 用户空间的合集列表响应
type UserSeriesResponse struct {
	Cursor      int              `json:"cursor"`
	HasMore     int              `json:"has_more"`
	SeriesInfos []SeriesInfoItem `json:"series_infos"`
	StatusCode  int              `json:"status_code"`
	StatusMsg   string           `json:"status_msg"`
	Total       int              `json:"total"`
}

type SeriesInfoItem struct {
	Author     Author     `json:"author"`
	CoverUrl   Cover      `json:"cover_url"`
	CreateTime int        `json:"create_time"`
	Desc       string     `json:"desc"`
	RealName   string     `json:"real_name"`
	SeriesName string     `json:"series_name"`
	SeriesID   string     `json:"series_id"`
	Stats      SeriesStat `json:"stats"`
}

type SeriesStat struct {
	UpdatedToEpisode int `json:"updated_to_episode"` // 合集数量
	TotalEpisode     int `json:"total_episode"`      // 合集数量
}
